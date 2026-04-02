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

	"k8s.io/klog/v2"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
)

func (a *K8sAdapter) Pause(ctx context.Context, req adapter.PauseRequest) (*adapter.Instance, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	sts, err := a.findStatefulSet(ctx, req.Name, req.UserID)
	if err != nil {
		return nil, adapter.ErrNotFound
	}

	if sts.Annotations == nil {
		sts.Annotations = make(map[string]string)
	}
	sts.Annotations[ScheduledDeletionTime] = time.Now().
		Add(time.Duration(req.DelayMinutes) * time.Minute).
		Format(time.RFC3339)

	if err := a.client.Update(ctx, &sts); err != nil {
		return nil, fmt.Errorf("update statefulset %s: %w", sts.Name, err)
	}
	klog.Infof("[pause] scheduled %s (user=%s, delay=%dm)", sts.Name, req.UserID, req.DelayMinutes)

	inst := a.toInstance(sts)
	return inst, nil
}
