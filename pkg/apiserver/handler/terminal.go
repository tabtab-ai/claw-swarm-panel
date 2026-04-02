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

package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// wsWriter serialises concurrent WebSocket writes.
type wsWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.conn.WriteMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// TerminalInstance handles GET /claw/instances/{name}/terminal.
// It upgrades the connection to WebSocket and bridges stdin/stdout to a K8s
// exec PTY session.
//
// Auth: JWT must be supplied as ?token=<jwt> because browsers cannot set
// custom headers on WebSocket connections.
//
// Client→Server messages:
//   - Binary frames: raw stdin bytes
//   - Text frames:   JSON {"type":"resize","cols":N,"rows":N}
//
// Server→Client messages:
//   - Binary frames: stdout/stderr (TTY merges them)
func (h *ClawHandler) TerminalInstance(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "instance name is required")
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		klog.Errorf("terminal: websocket upgrade error for %s: %v", name, err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Stdin pipe: WebSocket binary frames → K8s stdin
	pr, pw := io.Pipe()

	// Resize queue: text JSON frames → K8s TerminalSizeQueue
	resizeCh := make(chan adapter.TerminalSize, 4)

	// Read WebSocket messages and dispatch to stdin pipe or resize channel.
	go func() {
		defer pw.Close()
		defer close(resizeCh)
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				cancel()
				return
			}
			switch msgType {
			case websocket.BinaryMessage:
				if _, err := pw.Write(msg); err != nil {
					cancel()
					return
				}
			case websocket.TextMessage:
				var resize struct {
					Type string `json:"type"`
					Cols uint16 `json:"cols"`
					Rows uint16 `json:"rows"`
				}
				if json.Unmarshal(msg, &resize) == nil && resize.Type == "resize" {
					select {
					case resizeCh <- adapter.TerminalSize{Width: resize.Cols, Height: resize.Rows}:
					default:
					}
				}
			}
		}
	}()

	if err := h.svc.Terminal(ctx, name, pr, &wsWriter{conn: conn}, resizeCh); err != nil {
		klog.V(4).Infof("terminal: session ended for %s: %v", name, err)
	}
}
