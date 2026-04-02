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

package claw

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	createTable = `	CREATE TABLE IF NOT EXISTS claw_tokens (
			name       TEXT    PRIMARY KEY,
			token      TEXT    NOT NULL,
			created_at DATETIME DEFAULT (datetime('now','localtime'))
		)`

	insert = `INSERT INTO claw_tokens (name, token) VALUES (?, ?)`
	query  = `SELECT token FROM claw_tokens WHERE name = ?`
	remove = `DELETE FROM claw_tokens WHERE name = ?`
)

type ClawTokenStorage struct {
	db *sql.DB
}

func NewClawTokenStorage(db *sql.DB) ClawTokenStorage {
	if _, err := db.Exec(createTable); err != nil {
		panic(err)
	}
	return ClawTokenStorage{db: db}
}

func (c ClawTokenStorage) Insert(ctx context.Context, name, token string) error {
	_, err := c.db.ExecContext(ctx, insert, name, token)
	if err != nil {
		return fmt.Errorf("insert new token with name %s error %w", name, err)
	}

	return nil
}

func (c ClawTokenStorage) Query(ctx context.Context, name string) (string, error) {
	var token string
	err := c.db.QueryRowContext(ctx, query, name).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("query token by %s error %w", name, err)
	}
	return token, nil
}

func (c ClawTokenStorage) Remove(ctx context.Context, name string) error {
	_, err := c.db.ExecContext(ctx, remove, name)
	if err != nil {
		return fmt.Errorf("remove token by %s error %w", name, err)
	}
	return nil
}
