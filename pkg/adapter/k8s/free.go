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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
)

func (a *K8sAdapter) Free(ctx context.Context, name, userID string) (*adapter.Instance, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	sts, err := a.findStatefulSet(ctx, name, userID)
	if err != nil {
		return nil, adapter.ErrNotFound
	}

	inst := a.toInstance(sts)

	if err := a.client.Delete(ctx, &sts); err != nil {
		return nil, fmt.Errorf("delete statefulset %s: %w", sts.Name, err)
	}
	klog.Infof("[free] deleted instance %s (user=%s)", sts.Name, userID)

	return inst, nil
}

// findStatefulSetByUserID returns the StatefulSet occupied by userID.
func (a *K8sAdapter) findStatefulSetByUserID(ctx context.Context, userID string) (appsv1.StatefulSet, error) {
	inst, err := a.GetByUser(ctx, userID)
	if err != nil {
		return appsv1.StatefulSet{}, err
	}
	return a.getSTS(ctx, inst.Name)
}

// getSTS fetches a StatefulSet by name.
func (a *K8sAdapter) getSTS(ctx context.Context, name string) (appsv1.StatefulSet, error) {
	var sts appsv1.StatefulSet
	err := a.client.Get(ctx, client.ObjectKey{Namespace: a.namespace, Name: name}, &sts)
	return sts, err
}

// findStatefulSet resolves a StatefulSet by name (primary, required) or falls
// back to userID when name is empty.
func (a *K8sAdapter) findStatefulSet(ctx context.Context, name, userID string) (appsv1.StatefulSet, error) {
	if name != "" {
		return a.getSTS(ctx, name)
	}
	if userID != "" {
		return a.findStatefulSetByUserID(ctx, userID)
	}
	return appsv1.StatefulSet{}, fmt.Errorf("name is required")
}
