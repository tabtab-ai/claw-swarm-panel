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

package litellm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a LiteLLM API client
type Client struct {
	baseURL               string
	masterKey             string
	defaultTeam           string
	defaultMaxBudget      float64
	defaultBudgetDuration string
	client                *http.Client
}

// NewClient creates a new LiteLLM client
func NewClient(baseURL, masterKey, defaultTeam string, defaultMaxBudget float64, defaultBudgetDuration string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:4000"
	}
	return &Client{
		baseURL:               baseURL,
		masterKey:             masterKey,
		defaultTeam:           defaultTeam,
		defaultMaxBudget:      defaultMaxBudget,
		defaultBudgetDuration: defaultBudgetDuration,
		client:                &http.Client{Timeout: 30 * time.Second},
	}
}

// NewUserRequest represents a request to create a new user
type NewUserRequest struct {
	UserID         string                 `json:"user_id,omitempty"`
	UserAlias      string                 `json:"user_alias,omitempty"`
	UserEmail      string                 `json:"user_email,omitempty"`
	UserRole       string                 `json:"user_role,omitempty"`
	Teams          []string               `json:"teams,omitempty"`
	MaxBudget      float64                `json:"max_budget,omitempty"`
	BudgetDuration string                 `json:"budget_duration,omitempty"`
	Models         []string               `json:"models,omitempty"`
	AutoCreateKey  bool                   `json:"auto_create_key,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// NewUserResponse represents the response from creating a new user
type NewUserResponse struct {
	UserID    string        `json:"user_id,omitempty"`
	UserAlias string        `json:"user_alias,omitempty"`
	UserEmail string        `json:"user_email,omitempty"`
	UserRole  string        `json:"user_role,omitempty"`
	Key       string        `json:"key,omitempty"`
	Token     string        `json:"token,omitempty"`
	TokenID   string        `json:"token_id,omitempty"`
	BudgetID  string        `json:"budget_id,omitempty"`
	MaxBudget float64       `json:"max_budget,omitempty"`
	Models    []interface{} `json:"models,omitempty"`
	Teams     []Team        `json:"teams,omitempty"`
	CreatedAt string        `json:"created_at,omitempty"`
	UpdatedAt string        `json:"updated_at,omitempty"`
}

// GenerateKeyRequest represents a request to generate a new API key
type GenerateKeyRequest struct {
	UserID         string                 `json:"user_id,omitempty"`
	KeyAlias       string                 `json:"key_alias,omitempty"`
	TeamID         string                 `json:"team_id,omitempty"`
	MaxBudget      float64                `json:"max_budget,omitempty"`
	BudgetDuration string                 `json:"budget_duration,omitempty"`
	Models         []string               `json:"models,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Team represents a team in LiteLLM
type Team struct {
	TeamID         string                 `json:"team_id,omitempty"`
	TeamAlias      string                 `json:"team_alias,omitempty"`
	Organization   string                 `json:"organization_id,omitempty"`
	MaxBudget      float64                `json:"max_budget,omitempty"`
	BudgetDuration string                 `json:"budget_duration,omitempty"`
	Models         []string               `json:"models,omitempty"`
	Blocked        bool                   `json:"blocked,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      string                 `json:"created_at,omitempty"`
	UpdatedAt      string                 `json:"updated_at,omitempty"`
}

// UserInfo represents the user information from LiteLLM
type UserInfo struct {
	UserID         string          `json:"user_id,omitempty"`
	UserAlias      string          `json:"user_alias,omitempty"`
	UserEmail      string          `json:"user_email,omitempty"`
	UserRole       string          `json:"user_role,omitempty"`
	Teams          json.RawMessage `json:"teams,omitempty"` // Can be []string or []Team
	MaxBudget      float64         `json:"max_budget,omitempty"`
	BudgetDuration string          `json:"budget_duration,omitempty"`
	Models         []string        `json:"models,omitempty"`
	CreatedAt      string          `json:"created_at,omitempty"`
	UpdatedAt      string          `json:"updated_at,omitempty"`
}

// UserInfoResponse represents the response from GET /user/info
type UserInfoResponse struct {
	UserID   string    `json:"user_id,omitempty"`
	UserInfo *UserInfo `json:"user_info,omitempty"`
	Keys     []KeyInfo `json:"keys,omitempty"`
	Teams    []Team    `json:"teams,omitempty"`
}

// KeyInfo represents API key information from LiteLLM
type KeyInfo struct {
	Key       string   `json:"key,omitempty"`
	Token     string   `json:"token,omitempty"`
	TokenID   string   `json:"token_id,omitempty"`
	KeyAlias  string   `json:"key_alias,omitempty"`
	UserID    string   `json:"user_id,omitempty"`
	TeamID    string   `json:"team_id,omitempty"`
	MaxBudget float64  `json:"max_budget,omitempty"`
	Models    []string `json:"models,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
}

// GenerateKeyResponse represents the response from generating a new API key
type GenerateKeyResponse struct {
	Key       string        `json:"key,omitempty"`
	Token     string        `json:"token,omitempty"`
	TokenID   string        `json:"token_id,omitempty"`
	UserID    string        `json:"user_id,omitempty"`
	KeyAlias  string        `json:"key_alias,omitempty"`
	BudgetID  string        `json:"budget_id,omitempty"`
	MaxBudget float64       `json:"max_budget,omitempty"`
	Models    []interface{} `json:"models,omitempty"`
	CreatedAt string        `json:"created_at,omitempty"`
	UpdatedAt string        `json:"updated_at,omitempty"`
}

// APIErrorDetail represents the error details from LiteLLM API
type APIErrorDetail struct {
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// APIError represents an error response from the LiteLLM API
type APIError struct {
	Err APIErrorDetail `json:"error,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("LiteLLM API error: %s (type: %s, code: %s)", e.Err.Message, e.Err.Type, e.Err.Code)
}

// CreateUser creates a new user in LiteLLM and optionally generates an API key
func (c *Client) CreateUser(ctx context.Context, req *NewUserRequest) (*NewUserResponse, error) {
	if !req.AutoCreateKey {
		req.AutoCreateKey = true // Always auto-create key for simplicity
	}
	if len(req.Teams) == 0 && c.defaultTeam != "" {
		req.Teams = []string{c.defaultTeam}
	}

	url := fmt.Sprintf("%s/user/new", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.masterKey))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return nil, fmt.Errorf("LiteLLM API returned status %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, &apiErr
	}

	var result NewUserResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GenerateKey generates a new API key for an existing user
func (c *Client) GenerateKey(ctx context.Context, req *GenerateKeyRequest) (*GenerateKeyResponse, error) {
	url := fmt.Sprintf("%s/key/generate", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.masterKey))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return nil, fmt.Errorf("LiteLLM API returned status %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, &apiErr
	}

	var result GenerateKeyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// CreateUserWithKey creates a user and returns the user ID and API key
func (c *Client) CreateUserWithKey(ctx context.Context, userID, userAlias, userEmail string, models []string) (string, string, error) {
	req := &NewUserRequest{
		UserID:        userID,
		UserAlias:     userAlias,
		UserEmail:     userEmail,
		UserRole:      "internal_user",
		AutoCreateKey: true,
		Models:        models,
	}

	resp, err := c.CreateUser(ctx, req)
	if err != nil {
		return "", "", err
	}

	return resp.UserID, resp.Key, nil
}

// GetUser retrieves user information from LiteLLM by user ID
func (c *Client) GetUser(ctx context.Context, userID string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/user/info?user_id=%s", c.baseURL, userID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.masterKey))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return nil, fmt.Errorf("LiteLLM API returned status %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, &apiErr
	}

	var result UserInfoResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.UserInfo, nil
}

// IsUserExists checks if a user exists in LiteLLM
func (c *Client) IsUserExists(ctx context.Context, userID string) (bool, error) {
	_, err := c.GetUser(ctx, userID)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			// User not found - check multiple possible error indicators
			if apiErr.Err.Code == "not_found" ||
				apiErr.Err.Type == "not_found" ||
				strings.Contains(strings.ToLower(apiErr.Err.Message), "not found") ||
				strings.Contains(strings.ToLower(apiErr.Err.Message), "does not exist") {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

// CreateUserOrGetKey creates a user if not exists, or generates a new key for existing user
// Returns the API key and a boolean indicating if the user was newly created
func (c *Client) CreateUserOrGetKey(ctx context.Context, userID, userAlias string, models []string) (string, bool, error) {
	// First check if user already exists
	exists, err := c.IsUserExists(ctx, userID)
	if err != nil {
		return "", false, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		// User already exists, generate a new key for this user
		// Use a unique key alias to ensure we create a new key each time
		keyAlias := fmt.Sprintf("%s-%d", userID, time.Now().Unix())
		keyReq := &GenerateKeyRequest{
			UserID:         userID,
			KeyAlias:       keyAlias,
			Models:         models,
			MaxBudget:      c.defaultMaxBudget,
			BudgetDuration: c.defaultBudgetDuration,
		}

		if c.defaultTeam != "" {
			keyReq.TeamID = c.defaultTeam
		}

		keyResp, err := c.GenerateKey(ctx, keyReq)
		if err != nil {
			return "", false, fmt.Errorf("failed to generate key for existing user: %w", err)
		}

		return keyResp.Key, false, nil
	}

	// User does not exist, create a new user without budget (per-key budget approach)
	req := &NewUserRequest{
		UserID:        userID,
		UserAlias:     userAlias,
		UserRole:      "internal_user",
		AutoCreateKey: false, // Don't auto-create key, we create it separately with budget
		Models:        models,
		// Note: MaxBudget is not set (0 = unlimited) - budget is managed per-key
	}

	resp, err := c.CreateUser(ctx, req)
	if err != nil {
		return "", false, err
	}

	// Create a new key with budget settings
	keyAlias := fmt.Sprintf("%s-%d", userID, time.Now().Unix())
	keyReq := &GenerateKeyRequest{
		UserID:         resp.UserID,
		KeyAlias:       keyAlias,
		Models:         models,
		MaxBudget:      c.defaultMaxBudget,
		BudgetDuration: c.defaultBudgetDuration,
	}

	if c.defaultTeam != "" {
		keyReq.TeamID = c.defaultTeam
	}

	keyResp, err := c.GenerateKey(ctx, keyReq)
	if err != nil {
		return "", false, fmt.Errorf("failed to generate key for new user: %w", err)
	}

	return keyResp.Key, true, nil
}
