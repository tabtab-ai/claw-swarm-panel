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

// Package adapterfactory creates ClawAdapter instances based on runtime config.
package adapterfactory

import (
	"fmt"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
	k8sadapter "gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter/k8s"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/config"
)

// Compile-time interface conformance checks.
var (
	_ adapter.ClawAdapter = (*k8sadapter.K8sAdapter)(nil)
)

// New creates a ClawAdapter based on cfg.Adapter.Type.
// Currently only "k8s" is supported.
func New(cfg config.Config) (adapter.ClawAdapter, error) {
	t := cfg.Adapter.Type
	if t == "" {
		t = string(adapter.RuntimeK8s)
	}

	switch adapter.RuntimeType(t) {
	case adapter.RuntimeK8s:
		return k8sadapter.New(cfg)
	default:
		return nil, fmt.Errorf("unknown runtime type: %q (supported: k8s)", t)
	}
}
