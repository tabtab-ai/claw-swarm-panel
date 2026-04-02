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
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/utils"
)

func (a *K8sAdapter) Alloc(ctx context.Context, req adapter.AllocRequest) (*adapter.Instance, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Fast-path: check if user already has an instance
	existing, err := a.GetByUser(ctx, req.UserID)
	if err == nil && existing != nil {
		return existing, adapter.ErrUserAlreadyHasInstance
	}

	// Determine model type
	modelType := req.ModelType
	if modelType == "" {
		modelType = "lite"
	}

	// Create or retrieve LiteLLM API key
	apiKey, err := a.resolveLiteLLMKey(ctx, req.UserID, modelType)
	if err != nil {
		return nil, fmt.Errorf("resolve litellm key: %w", err)
	}

	// Allocate a StatefulSet (with retries)
	var stsName string
	_ = retry.OnError(retry.DefaultBackoff, func(e error) bool {
		return e != nil && !errors.Is(e, adapter.ErrNoAvailableInstance)
	}, func() error {
		var allocErr error
		stsName, allocErr = a.allocStatefulset(ctx, req.UserID, apiKey, modelType)
		return allocErr
	})

	if stsName == "" {
		return nil, adapter.ErrNoAvailableInstance
	}

	// Fetch the full instance so URL annotations (set by operator) are included.
	inst, err := a.GetByName(ctx, stsName)
	if err != nil {
		klog.Warningf("[alloc] fetch instance %s after alloc: %v", stsName, err)
		// Return a minimal instance rather than failing the whole alloc.
		return &adapter.Instance{
			Name:        stsName,
			UserID:      req.UserID,
			State:       adapter.StatePending,
			AllocStatus: adapter.AllocAllocating,
			RuntimeType: adapter.RuntimeK8s,
		}, nil
	}
	return inst, nil
}

// resolveLiteLLMKey creates a LiteLLM user+key or generates a random fallback.
func (a *K8sAdapter) resolveLiteLLMKey(ctx context.Context, userID, modelType string) (string, error) {
	if a.litellm != nil && a.config.Init.LiteLLM.MasterKey != "" {
		models := []string{a.config.Init.LiteLLM.Models.Lite.ModelID}
		if modelType == "pro" {
			models = []string{a.config.Init.LiteLLM.Models.Pro.ModelID}
		}
		key, isNew, err := a.litellm.CreateUserOrGetKey(ctx, userID, userID, models)
		if err != nil {
			return "", err
		}
		if isNew {
			klog.Infof("[alloc] created LiteLLM user %s key %s...", userID, key[:min(8, len(key))])
		} else {
			klog.Infof("[alloc] new key for existing LiteLLM user %s: %s...", userID, key[:min(8, len(key))])
		}
		return key, nil
	}
	// Fallback: random key
	klog.Warning("[alloc] LiteLLM not configured, using random API key")
	return utils.RandomString(32)
}

// allocStatefulset selects an idle StatefulSet, marks it as allocating, and
// asynchronously configures the pod models.
func (a *K8sAdapter) allocStatefulset(ctx context.Context, userID, apiKey, modelType string) (string, error) {
	labelSelector, _ := labels.Parse(TAB_CLAW)
	var list appsv1.StatefulSetList
	if err := a.client.List(ctx, &list, client.MatchingLabelsSelector{Selector: labelSelector}, client.InNamespace(a.namespace)); err != nil {
		return "", err
	}
	if len(list.Items) == 0 {
		return "", adapter.ErrNoAvailableInstance
	}

	// Prefer earliest-created (pods already warm)
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].CreationTimestamp.Before(&list.Items[j].CreationTimestamp)
	})

	selected := -1
	for i, item := range list.Items {
		v, ok := item.Labels[TAB_CLAW_OCCUPIED]
		if !ok {
			if selected == -1 {
				selected = i
			}
			continue
		}
		if v == userID {
			// User already has this instance
			return item.Name, adapter.ErrNoAvailableInstance
		}
	}
	if selected == -1 {
		return "", adapter.ErrNoAvailableInstance
	}

	used := list.Items[selected]
	if used.Labels == nil {
		used.Labels = make(map[string]string)
	}
	used.Labels[TAB_CLAW_OCCUPIED] = userID
	used.Labels[TAB_CLAW_ALLOC_STATUS] = AllocStatusAllocating
	if used.Spec.Replicas == nil {
		used.Spec.Replicas = new(int32)
	}
	if *used.Spec.Replicas == 0 {
		*used.Spec.Replicas = 1
	}
	if err := a.client.Update(ctx, &used); err != nil {
		return "", fmt.Errorf("update statefulset: %w", err)
	}
	klog.Infof("[alloc] marked %s as allocating for user %s", used.Name, userID)

	// Configure models asynchronously, then flip to allocated
	stsName := used.Name
	go func() {
		bgCtx := context.Background()
		if err := a.configurePodModels(bgCtx, stsName, apiKey, modelType); err != nil {
			klog.Warningf("[alloc] configure models for %s: %v", stsName, err)
		}
		var sts appsv1.StatefulSet
		if err := a.client.Get(bgCtx, client.ObjectKey{Namespace: a.namespace, Name: stsName}, &sts); err != nil {
			klog.Warningf("[alloc] fetch %s for status update: %v", stsName, err)
			return
		}
		sts.Labels[TAB_CLAW_ALLOC_STATUS] = AllocStatusAllocated
		if err := a.client.Update(bgCtx, &sts); err != nil {
			klog.Warningf("[alloc] mark %s as allocated: %v", stsName, err)
		} else {
			klog.Infof("[alloc] %s is now allocated", stsName)
		}
	}()

	return stsName, nil
}

// configurePodModels executes openclaw config commands inside the instance pod.
func (a *K8sAdapter) configurePodModels(ctx context.Context, stsName, apiKey, modelType string) error {
	podName := stsName + "-0"

	var pod corev1.Pod
	if err := a.client.Get(ctx, client.ObjectKey{Namespace: a.namespace, Name: podName}, &pod); err != nil {
		return fmt.Errorf("get pod %s: %w", podName, err)
	}
	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("pod %s not running (phase: %s)", podName, pod.Status.Phase)
	}

	modelID := a.config.Init.LiteLLM.Models.Lite.ModelID
	if modelType == "pro" {
		modelID = a.config.Init.LiteLLM.Models.Pro.ModelID
	}
	if modelID == "" {
		modelID = "tabtab-lite"
	}

	baseURL := a.config.Init.LiteLLM.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:4000"
	}

	var providerJSON string
	if apiKey != "" {
		providerJSON = fmt.Sprintf(
			`{"baseUrl":%q,"models":[{"id":%q,"name":%q,"input":["text"],"contextWindow":200000}],"apiKey":%q,"auth":"api-key","api":"openai-completions"}`,
			baseURL, modelID, modelID, apiKey,
		)
	} else {
		providerJSON = fmt.Sprintf(
			`{"baseUrl":%q,"models":[{"id":%q,"name":%q,"input":["text"],"contextWindow":200000}],"api":"openai-completions"}`,
			baseURL, modelID, modelID,
		)
	}

	cmd := []string{"openclaw", "config", "set", "models.providers.tabtab-litellm", providerJSON}
	stdout, stderr, err := a.execInPod(ctx, a.namespace, podName, CONTAINER_NAME, cmd)
	if err != nil {
		return fmt.Errorf("set provider config: %w (stderr: %s)", err, stderr)
	}
	klog.V(4).Infof("[alloc] set provider config for %s: stdout=%s", stsName, stdout)

	primaryModel := fmt.Sprintf("tabtab-litellm/%s", modelID)
	cmd = []string{"openclaw", "config", "set", "agents.defaults.model.primary", primaryModel}
	stdout, stderr, err = a.execInPod(ctx, a.namespace, podName, CONTAINER_NAME, cmd)
	if err != nil {
		return fmt.Errorf("set primary model: %w (stderr: %s)", err, stderr)
	}
	klog.V(4).Infof("[alloc] set primary model for %s: stdout=%s", stsName, stdout)

	return nil
}

// execInPod runs a command inside a pod container via the K8s exec API.
func (a *K8sAdapter) execInPod(ctx context.Context, namespace, podName, containerName string, cmd []string) (string, string, error) {
	req := a.clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName).
		Param("stdout", "true").
		Param("stderr", "true")

	for _, c := range cmd {
		req = req.Param("command", c)
	}

	exec, err := remotecommand.NewSPDYExecutor(a.restConfig, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	return stdout.String(), stderr.String(), err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
