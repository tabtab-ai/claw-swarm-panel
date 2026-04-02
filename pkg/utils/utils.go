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

package utils

import (
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

const (
	letters      = "abcdefghijklmnopqrstuvwxyz"
	alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func RandomName(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[n.Int64()]
	}
	return string(b), nil
}

// RandomString generates a random alphanumeric string of the specified length
func RandomString(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumeric))))
		if err != nil {
			return "", err
		}
		b[i] = alphanumeric[n.Int64()]
	}
	return string(b), nil
}

// EnsureDataDir creates the directory (and parents) if it does not exist.
func EnsureDataDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}

// LoadOrCreateJWTSecret reads the JWT secret from <dir>/jwt_secret.
// If the file does not exist or is empty, a new 32-char random secret is
// generated, written to the file (mode 0600), and returned.
// The secret persists across restarts so existing tokens remain valid.
func LoadOrCreateJWTSecret(dir string) (string, error) {
	path := filepath.Join(dir, ".jwt_secret")
	data, err := os.ReadFile(path)
	if err == nil {
		if s := strings.TrimSpace(string(data)); s != "" {
			return s, nil
		}
	}
	secret, err := RandomString(32)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(secret), 0600); err != nil {
		return "", err
	}
	return secret, nil
}
