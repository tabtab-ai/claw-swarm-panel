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
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
)

// termSizeQueue implements remotecommand.TerminalSizeQueue via a channel.
type termSizeQueue struct {
	ch <-chan adapter.TerminalSize
}

func (q *termSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.ch
	if !ok {
		return nil
	}
	return &remotecommand.TerminalSize{Width: size.Width, Height: size.Height}
}

// Terminal opens an interactive PTY exec session into the primary pod of the
// named instance.  stdin/stdout are bridged to the caller; resize receives
// TerminalSize events forwarded to the K8s TerminalSizeQueue.
func (a *K8sAdapter) Terminal(ctx context.Context, req adapter.TerminalRequest, stdin io.Reader, stdout io.Writer, resize <-chan adapter.TerminalSize) error {
	if req.Name == "" {
		return fmt.Errorf("instance name is required")
	}

	container := req.Container
	if container == "" {
		container = CONTAINER_NAME
	}

	podName := req.Name + "-0"

	restReq := a.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(a.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{"/bin/bash"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(a.restConfig, "POST", restReq.URL())
	if err != nil {
		return fmt.Errorf("create executor: %w", err)
	}

	klog.Infof("terminal: starting stream for pod %s/%s container %s", a.namespace, podName, container)
	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stdout, // TTY merges stderr into stdout
		Tty:               true,
		TerminalSizeQueue: &termSizeQueue{ch: resize},
	})
}
