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

package audit

import (
	"context"
	"database/sql"
	"time"
)

// AuditLog records a single audit event.
type AuditLog struct {
	ID               int64     `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	OperatorID       int64     `json:"operator_id"`
	OperatorUsername string    `json:"operator_username"`
	Action           string    `json:"action"`
	TargetUserID     string    `json:"target_user_id"`
	InstanceName     string    `json:"instance_name"`
	Status           string    `json:"status"`
	Detail           string    `json:"detail"`
}

// AuditStore persists audit logs to SQLite.
type AuditStore struct {
	db *sql.DB
}

// NewAuditStore creates an AuditStore backed by the given database.
func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

// InitSchema creates the audit_logs table if it does not exist.
func (s *AuditStore) InitSchema() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
		operator_id INTEGER NOT NULL,
		operator_username TEXT NOT NULL,
		action TEXT NOT NULL,
		target_user_id TEXT NOT NULL DEFAULT '',
		instance_name TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL,
		detail TEXT NOT NULL DEFAULT ''
	)`)
	return err
}

// Log inserts an audit entry. Errors are intentionally non-fatal for callers.
func (s *AuditStore) Log(ctx context.Context, operatorID int64, operatorUsername, action, targetUserID, instanceName, status, detail string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_logs (operator_id, operator_username, action, target_user_id, instance_name, status, detail) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		operatorID, operatorUsername, action, targetUserID, instanceName, status, detail,
	)
	return err
}

// ListFilter controls what List returns.
type ListFilter struct {
	Action   string
	Page     int
	PageSize int
}

// ListResult is the paginated response from List.
type ListResult struct {
	Data     []AuditLog `json:"data"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// List returns paginated audit log entries, optionally filtered by action.
func (s *AuditStore) List(ctx context.Context, filter ListFilter) (*ListResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	var whereClause string
	var whereArgs []any
	if filter.Action != "" {
		whereClause = " WHERE action = ?"
		whereArgs = append(whereArgs, filter.Action)
	}

	var total int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM audit_logs"+whereClause,
		whereArgs...,
	).Scan(&total); err != nil {
		return nil, err
	}

	listArgs := append(append([]any{}, whereArgs...), filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, timestamp, operator_id, operator_username, action, target_user_id, instance_name, status, detail FROM audit_logs"+whereClause+
			" ORDER BY id DESC LIMIT ? OFFSET ?",
		listArgs...,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	logs := []AuditLog{}
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.Timestamp, &l.OperatorID, &l.OperatorUsername, &l.Action, &l.TargetUserID, &l.InstanceName, &l.Status, &l.Detail); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return &ListResult{
		Data:     logs,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}
