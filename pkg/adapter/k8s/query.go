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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
)

func (a *K8sAdapter) List(ctx context.Context) ([]*adapter.Instance, error) {
	labelSelector, _ := labels.Parse(TAB_CLAW)
	var list appsv1.StatefulSetList
	if err := a.client.List(ctx, &list, client.MatchingLabelsSelector{Selector: labelSelector}, client.InNamespace(a.namespace)); err != nil {
		if k8serror.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list statefulsets: %w", err)
	}

	result := make([]*adapter.Instance, len(list.Items))
	for i, sts := range list.Items {
		inst := a.toInstance(sts)
		result[i] = inst
	}
	return result, nil
}

func (a *K8sAdapter) GetByUser(ctx context.Context, userID string) (*adapter.Instance, error) {
	selectorStr := fmt.Sprintf("%s,%s=%s", TAB_CLAW, TAB_CLAW_OCCUPIED, userID)
	labelSelector, _ := labels.Parse(selectorStr)
	var list appsv1.StatefulSetList
	if err := a.client.List(ctx, &list, client.MatchingLabelsSelector{Selector: labelSelector}, client.InNamespace(a.namespace)); err != nil {
		return nil, fmt.Errorf("list statefulsets by user: %w", err)
	}
	if len(list.Items) == 0 {
		return nil, adapter.ErrNotFound
	}
	inst := a.toInstance(list.Items[0])
	return inst, nil
}

func (a *K8sAdapter) GetByName(ctx context.Context, name string) (*adapter.Instance, error) {
	var sts appsv1.StatefulSet
	if err := a.client.Get(ctx, client.ObjectKey{Namespace: a.namespace, Name: name}, &sts); err != nil {
		if k8serror.IsNotFound(err) {
			return nil, adapter.ErrNotFound
		}
		return nil, fmt.Errorf("get statefulset %s: %w", name, err)
	}
	if _, ok := sts.Labels[TAB_CLAW]; !ok {
		return nil, adapter.ErrNotFound
	}
	inst := a.toInstance(sts)

	var secret v1.Secret
	if err := a.client.Get(ctx, client.ObjectKey{Namespace: a.namespace, Name: name}, &secret); err != nil {
		if k8serror.IsNotFound(err) {
			return nil, adapter.ErrNotFound
		}
	}

	inst.Token = string(secret.Data["token"])
	return inst, nil
}

// toInstance converts a StatefulSet to an adapter.Instance.
func (a *K8sAdapter) toInstance(sts appsv1.StatefulSet) *adapter.Instance {
	userID := sts.Labels[TAB_CLAW_OCCUPIED]
	return &adapter.Instance{
		Name:        sts.Name,
		UserID:      userID,
		State:       stsState(sts),
		AllocStatus: stsAllocStatus(sts),
		AccessURL:   sts.Annotations[AnnotationGatewayURL],
		WssURL:      sts.Annotations[AnnotationWssURL],
		Resources:   stsResourceSpec(sts),
		CreatedAt:   sts.CreationTimestamp.Time,
		RuntimeType: adapter.RuntimeK8s,
	}
}

func stsState(sts appsv1.StatefulSet) adapter.InstanceState {
	replicas := sts.Status.Replicas
	availableReplicas := sts.Status.AvailableReplicas

	var state adapter.InstanceState
	switch replicas {
	case 0:
		state = adapter.StatePaused
	case availableReplicas:
		state = adapter.StateRunning
	default:
		state = adapter.StatePending
	}
	if sts.Annotations != nil {
		if _, err := time.Parse(time.RFC3339, sts.Annotations[ScheduledDeletionTime]); err == nil {
			state = adapter.StatePaused
		}
	}
	return state
}

func stsAllocStatus(sts appsv1.StatefulSet) adapter.AllocStatus {
	switch sts.Labels[TAB_CLAW_ALLOC_STATUS] {
	case AllocStatusAllocating:
		return adapter.AllocAllocating
	case AllocStatusAllocated:
		return adapter.AllocAllocated
	default:
		return adapter.AllocIdle
	}
}

func stsResourceSpec(sts appsv1.StatefulSet) adapter.ResourceSpec {
	if len(sts.Spec.Template.Spec.Containers) == 0 {
		return adapter.ResourceSpec{}
	}
	res := sts.Spec.Template.Spec.Containers[0].Resources
	spec := adapter.ResourceSpec{}
	if q := res.Requests.Cpu(); q != nil && !q.IsZero() {
		spec.CPURequest = q.String()
	}
	if q := res.Limits.Cpu(); q != nil && !q.IsZero() {
		spec.CPULimit = q.String()
	}
	if q := res.Requests.Memory(); q != nil && !q.IsZero() {
		spec.MemoryRequest = q.String()
	}
	if q := res.Limits.Memory(); q != nil && !q.IsZero() {
		spec.MemoryLimit = q.String()
	}
	return spec
}
