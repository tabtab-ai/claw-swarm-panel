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

// Package service provides the ClawService business logic layer.
// It sits between HTTP handlers and the ClawAdapter, handling audit logging
// and any cross-cutting business rules.
package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"k8s.io/klog/v2"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/audit"
)

// ClawService coordinates instance lifecycle operations.
type ClawService struct {
	adapter    adapter.ClawAdapter
	auditStore *audit.AuditStore
}

// New creates a ClawService.
func New(a adapter.ClawAdapter, auditStore *audit.AuditStore) *ClawService {
	return &ClawService{
		adapter:    a,
		auditStore: auditStore,
	}
}

// Alloc allocates a claw instance to a user.
// Returns ErrUserAlreadyHasInstance if the user already holds one.
func (s *ClawService) Alloc(ctx context.Context, operatorID int64, operatorUsername, userID, modelType string) (*adapter.Instance, error) {
	inst, err := s.adapter.Alloc(ctx, adapter.AllocRequest{
		UserID:    userID,
		ModelType: modelType,
	})
	if err != nil {
		if errors.Is(err, adapter.ErrUserAlreadyHasInstance) {
			return inst, err // Return existing instance + sentinel error
		}
		return nil, fmt.Errorf("alloc: %w", err)
	}

	s.logAudit(ctx, operatorID, operatorUsername, "alloc", userID, inst.Name, "success", "model_type="+modelType)
	return inst, nil
}

// Free deallocates the instance identified by name (required) or userID (fallback).
func (s *ClawService) Free(ctx context.Context, operatorID int64, operatorUsername, name, userID string) (*adapter.Instance, error) {
	inst, err := s.adapter.Free(ctx, name, userID)
	if err != nil {
		return nil, fmt.Errorf("free: %w", err)
	}

	s.logAudit(ctx, operatorID, operatorUsername, "free", userID, inst.Name, "success", "")
	return inst, nil
}

// Pause schedules a pause for the instance identified by name (required) or userID (fallback).
func (s *ClawService) Pause(ctx context.Context, operatorID int64, operatorUsername, name, userID string, delayMinutes int) (*adapter.Instance, error) {
	inst, err := s.adapter.Pause(ctx, adapter.PauseRequest{
		Name:         name,
		UserID:       userID,
		DelayMinutes: delayMinutes,
	})
	if err != nil {
		return nil, fmt.Errorf("pause: %w", err)
	}

	s.logAudit(ctx, operatorID, operatorUsername, "pause", userID, inst.Name, "success",
		fmt.Sprintf("delay_minutes=%d", delayMinutes))
	return inst, nil
}

// Resume resumes the instance identified by name (required) or userID (fallback).
func (s *ClawService) Resume(ctx context.Context, operatorID int64, operatorUsername, name, userID string) (*adapter.Instance, error) {
	inst, err := s.adapter.Resume(ctx, name, userID)
	if err != nil {
		return nil, fmt.Errorf("resume: %w", err)
	}

	s.logAudit(ctx, operatorID, operatorUsername, "resume", userID, inst.Name, "success", "")
	return inst, nil
}

// List returns all known instances.
func (s *ClawService) List(ctx context.Context) ([]*adapter.Instance, error) {
	return s.adapter.List(ctx)
}

// GetByUser returns the instance allocated to userID.
func (s *ClawService) GetByUser(ctx context.Context, userID string) (*adapter.Instance, error) {
	return s.adapter.GetByUser(ctx, userID)
}

// GetByName returns an instance by name, optionally enriching it with a token.
func (s *ClawService) GetByName(ctx context.Context, name string) (*adapter.Instance, error) {
	return s.adapter.GetByName(ctx, name)
}

// Exec runs a command inside the named instance's primary container.
func (s *ClawService) Exec(ctx context.Context, name string, command []string, container string) (*adapter.ExecResult, error) {
	result, err := s.adapter.Exec(ctx, adapter.ExecRequest{
		Name:      name,
		Command:   command,
		Container: container,
	})
	if err != nil {
		return result, fmt.Errorf("exec: %w", err)
	}
	return result, nil
}

// Terminal opens an interactive PTY session into the named instance's container.
func (s *ClawService) Terminal(ctx context.Context, name string, stdin io.Reader, stdout io.Writer, resize <-chan adapter.TerminalSize) error {
	return s.adapter.Terminal(ctx, adapter.TerminalRequest{Name: name}, stdin, stdout, resize)
}

// Healthy checks whether the adapter backend is reachable.
func (s *ClawService) Healthy(ctx context.Context) error {
	return s.adapter.Healthy(ctx)
}

func (s *ClawService) logAudit(ctx context.Context, operatorID int64, operatorUsername, action, targetUserID, instanceName, status, detail string) {
	if s.auditStore == nil {
		return
	}
	if err := s.auditStore.Log(ctx, operatorID, operatorUsername, action, targetUserID, instanceName, status, detail); err != nil {
		klog.Warningf("[audit] write failed: %v", err)
	}
}
