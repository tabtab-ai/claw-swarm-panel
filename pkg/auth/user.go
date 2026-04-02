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

package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"k8s.io/klog/v2"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// User represents a management panel user.
type User struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	PasswordHash   string    `json:"-"`
	Role           string    `json:"role"`
	ForceChangePwd bool      `json:"force_change_password"`
	APISecret      string    `json:"api_secret,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// UserStore provides CRUD operations for users backed by SQLite.
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a UserStore for the given database connection.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// InitSchema creates the users table if it does not exist.
func (s *UserStore) InitSchema() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		force_change_password INTEGER NOT NULL DEFAULT 0,
		api_secret TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT (datetime('now','localtime'))
	)`)
	return err
}

// MigrateSchema adds any missing columns to existing tables.
// Safe to call on both fresh and existing databases.
func (s *UserStore) MigrateSchema() error {
	rows, err := s.db.Query(`PRAGMA table_info(users)`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	hasAPISecret := false
	for rows.Next() {
		var cid int
		var name string
		var typ, notnull, dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "api_secret" {
			hasAPISecret = true
			break
		}
	}
	if !hasAPISecret {
		if _, err := s.db.Exec(`ALTER TABLE users ADD COLUMN api_secret TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("adding api_secret column: %w", err)
		}
		klog.Info("migrated users table: added api_secret column")
	}
	return nil
}

// generateAPISecret returns a new random secret with a "claw_" prefix.
func generateAPISecret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "claw_" + hex.EncodeToString(b), nil
}

// EnsureAdminUser creates the default admin account if it doesn't exist yet.
// Default credentials: admin / happyclaw (must be changed on first login).
func (s *UserStore) EnsureAdminUser() error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count); err != nil {
		return fmt.Errorf("checking admin user: %w", err)
	}
	if count > 0 {
		// Ensure existing admin has an api_secret generated.
		var secret string
		if err := s.db.QueryRow("SELECT api_secret FROM users WHERE username = 'admin'").Scan(&secret); err != nil {
			return fmt.Errorf("reading admin api_secret: %w", err)
		}
		if secret == "" {
			newSecret, err := generateAPISecret()
			if err != nil {
				return fmt.Errorf("generating admin api_secret: %w", err)
			}
			if _, err := s.db.Exec(`UPDATE users SET api_secret = ? WHERE username = 'admin'`, newSecret); err != nil {
				return fmt.Errorf("setting admin api_secret: %w", err)
			}
			klog.Infof("generated api_secret for existing admin user")
		}
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("happyclaw"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}
	secret, err := generateAPISecret()
	if err != nil {
		return fmt.Errorf("generating admin api_secret: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT INTO users (username, password_hash, role, force_change_password, api_secret) VALUES (?, ?, ?, ?, ?)`,
		"admin", string(hash), RoleAdmin, 1, secret,
	)
	if err != nil {
		return fmt.Errorf("creating admin user: %w", err)
	}
	klog.Info("created default admin user (username: admin, password: happyclaw, must change on first login)")
	return nil
}

// GetByUsername looks up a user by username.
func (s *UserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	u := &User{}
	var forceChangePwd int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, force_change_password, api_secret, created_at FROM users WHERE username = ?`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &forceChangePwd, &u.APISecret, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.ForceChangePwd = forceChangePwd == 1
	return u, nil
}

// GetByID looks up a user by ID.
func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	u := &User{}
	var forceChangePwd int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, force_change_password, api_secret, created_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &forceChangePwd, &u.APISecret, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.ForceChangePwd = forceChangePwd == 1
	return u, nil
}

// GetByAPISecret looks up a user by their API secret.
func (s *UserStore) GetByAPISecret(ctx context.Context, secret string) (*User, error) {
	if secret == "" {
		return nil, fmt.Errorf("empty secret")
	}
	u := &User{}
	var forceChangePwd int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, force_change_password, api_secret, created_at FROM users WHERE api_secret = ?`,
		secret,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &forceChangePwd, &u.APISecret, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.ForceChangePwd = forceChangePwd == 1
	return u, nil
}

// RegenerateAPISecret creates a new API secret for the given user and returns it.
func (s *UserStore) RegenerateAPISecret(ctx context.Context, id int64) (string, error) {
	secret, err := generateAPISecret()
	if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE users SET api_secret = ? WHERE id = ?`, secret, id)
	if err != nil {
		return "", err
	}
	return secret, nil
}

// List returns all users ordered by ID.
func (s *UserStore) List(ctx context.Context) ([]User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, username, role, force_change_password, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	users := []User{}
	for rows.Next() {
		var u User
		var forceChangePwd int
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &forceChangePwd, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.ForceChangePwd = forceChangePwd == 1
		users = append(users, u)
	}
	return users, nil
}

// Create adds a new user with the given role and auto-generates an API secret.
func (s *UserStore) Create(ctx context.Context, username, password, role string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	secret, err := generateAPISecret()
	if err != nil {
		return nil, err
	}
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, role, force_change_password, api_secret) VALUES (?, ?, ?, 0, ?)`,
		username, string(hash), role, secret,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return s.GetByID(ctx, id)
}

// UpdatePassword sets a new hashed password and clears force_change_password.
func (s *UserStore) UpdatePassword(ctx context.Context, id int64, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, force_change_password = 0 WHERE id = ?`,
		string(hash), id,
	)
	return err
}

// Delete removes a user by ID.
func (s *UserStore) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

// CheckPassword returns true if the plaintext password matches the user's hash.
func (s *UserStore) CheckPassword(user *User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}
