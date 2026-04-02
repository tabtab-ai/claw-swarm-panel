/*
Copyright 2026 The TabTabAI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8sadapter

import (
	"context"
	"fmt"
	"os"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/claw"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/config"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/db"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/litellm"
)

// K8sAdapter implements adapter.ClawAdapter using Kubernetes StatefulSets.
type K8sAdapter struct {
	client     client.Client
	clientset  kubernetes.Interface
	restConfig *rest.Config
	namespace  string
	config     config.Config
	litellm    *litellm.Client
	tokenStore claw.ClawTokenStorage
	mu         sync.Mutex
}

// buildRestConfig resolves a *rest.Config using the following priority:
//  1. explicit kubeconfig path (if non-empty)
//  2. default kubeconfig rules: KUBECONFIG env var, then ~/.kube/config
//  3. in-cluster config
func buildRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err == nil {
		return cfg, nil
	}
	return rest.InClusterConfig()
}

// New creates a K8sAdapter. Config resolution order:
//  1. cfg.Adapter.K8s.Kubeconfig (explicit path)
//  2. Default kubeconfig rules (KUBECONFIG env var or ~/.kube/config)
//  3. In-cluster config
func New(cfg config.Config) (*K8sAdapter, error) {
	restCfg, err := buildRestConfig(cfg.Adapter.K8s.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("build k8s config: %w", err)
	}

	scheme := newScheme()
	c, err := client.New(restCfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("create k8s client: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("create k8s clientset: %w", err)
	}

	namespace := cfg.Adapter.K8s.Namespace
	if namespace == "" {
		namespace = os.Getenv("POD_NAMESPACE")
	}
	if namespace == "" {
		namespace = "default"
	}

	llmClient := litellm.NewClient(
		cfg.Init.LiteLLM.BaseURL,
		cfg.Init.LiteLLM.MasterKey,
		cfg.Init.LiteLLM.DefaultTeam,
		cfg.Init.LiteLLM.DefaultMaxBudget,
		cfg.Init.LiteLLM.DefaultBudgetDuration,
	)

	a := &K8sAdapter{
		client:     c,
		clientset:  clientset,
		restConfig: restCfg,
		namespace:  namespace,
		config:     cfg,
		litellm:    llmClient,
		tokenStore: claw.NewClawTokenStorage(db.DB()),
	}

	// Trigger operator reconciliation on startup
	if err := a.triggerReconcile(context.Background()); err != nil {
		klog.Warningf("[k8s-adapter] failed to trigger reconcile: %v", err)
	}

	return a, nil
}

// NewWithConfig creates a K8sAdapter with an explicit rest.Config (e.g. from cluster manager).
func NewWithConfig(cfg config.Config, restCfg *rest.Config) (*K8sAdapter, error) {
	scheme := newScheme()
	c, err := client.New(restCfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("create k8s client: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("create k8s clientset: %w", err)
	}

	namespace := cfg.Adapter.K8s.Namespace
	if namespace == "" {
		namespace = os.Getenv("POD_NAMESPACE")
	}
	if namespace == "" {
		namespace = "default"
	}

	llmClient := litellm.NewClient(
		cfg.Init.LiteLLM.BaseURL,
		cfg.Init.LiteLLM.MasterKey,
		cfg.Init.LiteLLM.DefaultTeam,
		cfg.Init.LiteLLM.DefaultMaxBudget,
		cfg.Init.LiteLLM.DefaultBudgetDuration,
	)

	a := &K8sAdapter{
		client:     c,
		clientset:  clientset,
		restConfig: restCfg,
		namespace:  namespace,
		config:     cfg,
		litellm:    llmClient,
		tokenStore: claw.NewClawTokenStorage(db.DB()),
	}

	if err := a.triggerReconcile(context.Background()); err != nil {
		klog.Warningf("[k8s-adapter] failed to trigger reconcile: %v", err)
	}

	return a, nil
}

func (a *K8sAdapter) Type() adapter.RuntimeType {
	return adapter.RuntimeK8s
}

func (a *K8sAdapter) Healthy(ctx context.Context) error {
	var list appsv1.StatefulSetList
	return a.client.List(ctx, &list, client.InNamespace(a.namespace))
}

// triggerReconcile creates the sentinel claw-start StatefulSet that wakes the operator.
func (a *K8sAdapter) triggerReconcile(ctx context.Context) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "claw-start",
			Namespace: a.namespace,
			Labels:    map[string]string{TAB_CLAW_INIT_TRIGGER: ""},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: new(int32),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{TAB_CLAW_INIT_TRIGGER: ""},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{TAB_CLAW_INIT_TRIGGER: ""},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "nginx", Image: "nginx"}},
				},
			},
		},
	}

	err := a.client.Create(ctx, sts)
	if err == nil {
		return nil
	}
	if !k8serror.IsAlreadyExists(err) {
		return err
	}

	// Already exists – ensure trigger label is present
	var existing appsv1.StatefulSet
	if err := a.client.Get(ctx, client.ObjectKey{Namespace: a.namespace, Name: "claw-start"}, &existing); err != nil {
		return err
	}
	if existing.Labels == nil {
		existing.Labels = make(map[string]string)
	}
	if _, ok := existing.Labels[TAB_CLAW_INIT_TRIGGER]; !ok {
		existing.Labels[TAB_CLAW_INIT_TRIGGER] = ""
		return a.client.Update(ctx, &existing)
	}
	return nil
}
