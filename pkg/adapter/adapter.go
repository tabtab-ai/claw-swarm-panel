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

package adapter

import (
	"context"
	"errors"
	"io"
	"time"
)

// Sentinel errors
var (
	ErrNotFound               = errors.New("instance not found")
	ErrUserAlreadyHasInstance = errors.New("user already has an instance")
	ErrNoAvailableInstance    = errors.New("no available instance")
)

type ResourceSpec struct {
	CPURequest    string // e.g. "250m"
	CPULimit      string // e.g. "1"
	MemoryRequest string // e.g. "512Mi"
	MemoryLimit   string // e.g. "2Gi"
}

// RuntimeType 标识当前 Adapter 类型
type RuntimeType string

const (
	RuntimeK8s RuntimeType = "k8s"
)

// Instance 代表一个 Claw 实例（与运行时无关）
type Instance struct {
	Name        string
	UserID      string
	State       InstanceState // running | paused | pending | deleting
	AllocStatus AllocStatus   // idle | allocating | allocated
	AccessURL   string        // WebUI URL
	WssURL      string        // WebSocket URL
	Token       string
	Resources   ResourceSpec
	CreatedAt   time.Time
	RuntimeType RuntimeType
}

type InstanceState string

const (
	StateRunning  InstanceState = "running"
	StatePaused   InstanceState = "paused"
	StatePending  InstanceState = "pending"
	StateDeleting InstanceState = "deleting"
)

type AllocStatus string

const (
	AllocIdle       AllocStatus = "idle"
	AllocAllocating AllocStatus = "allocating"
	AllocAllocated  AllocStatus = "allocated"
)

// AllocRequest 分配请求
type AllocRequest struct {
	UserID    string
	ModelType string // "lite" | "pro"
}

// PauseRequest pause 请求（支持延迟）
type PauseRequest struct {
	Name         string // required: instance name
	UserID       string // optional: for audit logging
	DelayMinutes int
}

// ExecRequest is the input for executing a command inside a running instance.
type ExecRequest struct {
	Name      string   // instance name (required)
	Command   []string // command + args, e.g. ["bash", "-c", "ls -la"]
	Container string   // container name; empty = default container
}

// ExecResult holds the captured output of a remote command execution.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// TerminalSize carries a terminal resize event.
type TerminalSize struct {
	Width  uint16
	Height uint16
}

// TerminalRequest is the input for opening an interactive PTY session.
type TerminalRequest struct {
	Name      string // instance name (required)
	Container string // container name; empty = default container
}

// ClawAdapter 是所有运行时后端的统一接口
type ClawAdapter interface {
	// 核心生命周期
	Alloc(ctx context.Context, req AllocRequest) (*Instance, error)
	Free(ctx context.Context, name, userID string) (*Instance, error)
	Pause(ctx context.Context, req PauseRequest) (*Instance, error)
	Resume(ctx context.Context, name, userID string) (*Instance, error)

	// 查询
	List(ctx context.Context) ([]*Instance, error)
	GetByUser(ctx context.Context, userID string) (*Instance, error)
	GetByName(ctx context.Context, name string) (*Instance, error)

	// 命令执行
	Exec(ctx context.Context, req ExecRequest) (*ExecResult, error)

	// Terminal opens an interactive PTY session (WebSocket exec).
	// stdin/stdout are wired to the caller; resize receives terminal size events.
	Terminal(ctx context.Context, req TerminalRequest, stdin io.Reader, stdout io.Writer, resize <-chan TerminalSize) error

	// 运行时信息
	Type() RuntimeType
	Healthy(ctx context.Context) error
}
