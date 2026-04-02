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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
)

// Exec runs a command inside the primary pod of the named instance.
// For a StatefulSet named <name>, the pod is <name>-0.
func (a *K8sAdapter) Exec(ctx context.Context, req adapter.ExecRequest) (*adapter.ExecResult, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("instance name is required")
	}
	if len(req.Command) == 0 {
		return nil, fmt.Errorf("command is required")
	}

	container := req.Container
	if container == "" {
		container = CONTAINER_NAME
	}

	podName := req.Name + "-0"

	execOpts := &corev1.PodExecOptions{
		Container: container,
		Command:   req.Command,
		Stdout:    true,
		Stderr:    true,
	}

	restReq := a.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(a.namespace).
		SubResource("exec").
		VersionedParams(execOpts, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(a.restConfig, "POST", restReq.URL())
	if err != nil {
		return nil, fmt.Errorf("create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	streamErr := executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	result := &adapter.ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if streamErr != nil {
		// Extract exit code from the error if possible.
		if exitErr, ok := streamErr.(interface{ ExitStatus() int }); ok {
			result.ExitCode = exitErr.ExitStatus()
		} else {
			return result, fmt.Errorf("exec stream: %w", streamErr)
		}
	}

	return result, nil
}
