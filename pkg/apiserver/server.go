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

// Package apiserver provides the combined HTTP API server for the legacy
// single-node deployment mode (operator + apiserver in one process).
package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"k8s.io/klog/v2"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/apiserver/handler"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/apiserver/service"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/audit"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/auth"
)

// Server is the combined HTTP API server.
type Server struct {
	userStore  *auth.UserStore
	auditStore *audit.AuditStore
	jwtSecret  string
	clawSvc    *service.ClawService
}

// NewServer creates a Server.
func NewServer(
	userStore *auth.UserStore,
	auditStore *audit.AuditStore,
	jwtSecret string,
	clawSvc *service.ClawService,
) *Server {
	return &Server{
		userStore:  userStore,
		auditStore: auditStore,
		jwtSecret:  jwtSecret,
		clawSvc:    clawSvc,
	}
}

// Run starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Run(ctx context.Context, port int) {
	sig, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	mux := s.buildMux()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		<-sig.Done()
		klog.Info("apiserver: shutting down...")
		_ = srv.Shutdown(context.Background())
	}()

	klog.Infof("apiserver: listening on :%d", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		klog.Errorf("apiserver: %v", err)
	}
}

// buildMux constructs the full route table.
func (s *Server) buildMux() *http.ServeMux {
	mux := http.NewServeMux()
	clawH := handler.NewClawHandler(s.clawSvc)

	// ── Public ────────────────────────────────────────────────────────────
	mux.HandleFunc("POST /auth/login", s.login)

	// ── Authenticated ─────────────────────────────────────────────────────
	mustAuth := s.authMiddleware

	mux.Handle("GET /auth/me", mustAuth(http.HandlerFunc(s.me)))
	mux.Handle("POST /auth/change-password", mustAuth(http.HandlerFunc(s.changePassword)))
	mux.Handle("GET /auth/api-secret", mustAuth(http.HandlerFunc(s.handleAPISecret)))
	mux.Handle("POST /auth/api-secret/regenerate", mustAuth(http.HandlerFunc(s.handleAPISecret)))

	// Read claw ops: any authenticated user
	mux.Handle("GET /claw/instances/{name}", mustAuth(http.HandlerFunc(clawH.GetInstance)))
	mux.Handle("GET /claw/instances", mustAuth(http.HandlerFunc(clawH.ListInstances)))
	mux.Handle("GET /claw/used", mustAuth(http.HandlerFunc(clawH.ListByUser)))

	// ── Admin-only ────────────────────────────────────────────────────────
	mustAdmin := func(h http.Handler) http.Handler { return mustAuth(s.adminMiddleware(h)) }

	mux.Handle("GET /users", mustAdmin(http.HandlerFunc(s.listUsers)))
	mux.Handle("POST /users", mustAdmin(http.HandlerFunc(s.createUser)))
	mux.Handle("DELETE /users/{id}", mustAdmin(http.HandlerFunc(s.deleteUser)))

	mux.Handle("GET /audit/logs", mustAdmin(http.HandlerFunc(s.listAuditLogs)))

	mux.Handle("GET /claw/token", mustAdmin(http.HandlerFunc(clawH.GetToken)))
	mux.Handle("POST /claw/alloc", mustAdmin(http.HandlerFunc(clawH.Alloc)))
	mux.Handle("POST /claw/free", mustAdmin(http.HandlerFunc(clawH.Free)))
	mux.Handle("POST /claw/pause", mustAdmin(http.HandlerFunc(clawH.Pause)))
	mux.Handle("POST /claw/resume", mustAdmin(http.HandlerFunc(clawH.Resume)))
	mux.Handle("POST /claw/instances/{name}/exec", mustAdmin(http.HandlerFunc(clawH.ExecInstance)))
	mux.Handle("GET /claw/terminal/{name}", mustAdmin(http.HandlerFunc(clawH.TerminalInstance)))

	return mux
}

// ── Middleware ────────────────────────────────────────────────────────────────

// authMiddleware validates a Bearer JWT or X-API-Key header and stores claims
// in the request context. Returns 401 if neither is valid.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var claims *auth.Claims

		if tok := auth.BearerToken(r); tok != "" {
			if c, err := auth.ValidateToken(s.jwtSecret, tok); err == nil {
				claims = c
			}
		}
		// Also accept ?token= query param for WebSocket connections (browsers
		// cannot set Authorization headers during the WebSocket handshake).
		if claims == nil {
			if tok := r.URL.Query().Get("token"); tok != "" {
				if c, err := auth.ValidateToken(s.jwtSecret, tok); err == nil {
					claims = c
				}
			}
		}
		if claims == nil {
			if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
				if user, err := s.userStore.GetByAPISecret(r.Context(), apiKey); err == nil {
					claims = &auth.Claims{
						UserID:   user.ID,
						Username: user.Username,
						Role:     user.Role,
					}
				}
			}
		}
		if claims == nil {
			auth.WriteJSON(w, http.StatusUnauthorized, map[string]string{"msg": "unauthorized"})
			return
		}
		next.ServeHTTP(w, auth.WithClaims(r, claims))
	})
}

// adminMiddleware rejects non-admin callers with 403. Must be composed after authMiddleware.
func (s *Server) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _ := auth.ClaimsFromContext(r.Context())
		if claims == nil || claims.Role != auth.RoleAdmin {
			auth.WriteJSON(w, http.StatusForbidden, map[string]string{"msg": "admin only"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── Auth handlers ─────────────────────────────────────────────────────────────

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"msg": "invalid request body"})
		return
	}
	user, err := s.userStore.GetByUsername(r.Context(), body.Username)
	if err != nil || !s.userStore.CheckPassword(user, body.Password) {
		auth.WriteJSON(w, http.StatusUnauthorized, map[string]string{"msg": "invalid username or password"})
		return
	}
	token, err := auth.GenerateToken(s.jwtSecret, user)
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"msg": "failed to generate token"})
		return
	}
	auth.WriteJSON(w, http.StatusOK, map[string]any{
		"token":                 token,
		"username":              user.Username,
		"role":                  user.Role,
		"force_change_password": user.ForceChangePwd,
	})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	user, err := s.userStore.GetByID(r.Context(), claims.UserID)
	if err != nil {
		auth.WriteJSON(w, http.StatusNotFound, map[string]string{"msg": "user not found"})
		return
	}
	auth.WriteJSON(w, http.StatusOK, user)
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"msg": "invalid request body"})
		return
	}
	if len(body.NewPassword) < 6 {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"msg": "new password must be at least 6 characters"})
		return
	}
	user, err := s.userStore.GetByID(r.Context(), claims.UserID)
	if err != nil {
		auth.WriteJSON(w, http.StatusNotFound, map[string]string{"msg": "user not found"})
		return
	}
	if !s.userStore.CheckPassword(user, body.OldPassword) {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"msg": "incorrect current password"})
		return
	}
	if err := s.userStore.UpdatePassword(r.Context(), claims.UserID, body.NewPassword); err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"msg": "failed to update password"})
		return
	}
	user.ForceChangePwd = false
	token, _ := auth.GenerateToken(s.jwtSecret, user)
	auth.WriteJSON(w, http.StatusOK, map[string]string{"msg": "password updated", "token": token})
}

func (s *Server) handleAPISecret(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	if strings.HasSuffix(r.URL.Path, "/regenerate") {
		newSecret, err := s.userStore.RegenerateAPISecret(r.Context(), claims.UserID)
		if err != nil {
			auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"msg": "failed to regenerate api secret"})
			return
		}
		auth.WriteJSON(w, http.StatusOK, map[string]string{"api_secret": newSecret})
		return
	}
	user, err := s.userStore.GetByID(r.Context(), claims.UserID)
	if err != nil {
		auth.WriteJSON(w, http.StatusNotFound, map[string]string{"msg": "user not found"})
		return
	}
	auth.WriteJSON(w, http.StatusOK, map[string]string{"api_secret": user.APISecret})
}

// ── User handlers (admin only) ────────────────────────────────────────────────

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.userStore.List(r.Context())
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"msg": "failed to list users"})
		return
	}
	auth.WriteJSON(w, http.StatusOK, map[string]any{"data": users})
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"msg": "invalid request body"})
		return
	}
	if body.Username == "" || body.Password == "" {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"msg": "username and password are required"})
		return
	}
	if body.Role != auth.RoleAdmin && body.Role != auth.RoleUser {
		body.Role = auth.RoleUser
	}
	user, err := s.userStore.Create(r.Context(), body.Username, body.Password, body.Role)
	if err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"msg": "failed to create user: " + err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, user)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"msg": "invalid user id"})
		return
	}
	if id == claims.UserID {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"msg": "cannot delete yourself"})
		return
	}
	if err := s.userStore.Delete(r.Context(), id); err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"msg": "failed to delete user"})
		return
	}
	auth.WriteJSON(w, http.StatusOK, map[string]string{"msg": "deleted"})
}

// ── Audit handler (admin only) ────────────────────────────────────────────────

func (s *Server) listAuditLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	result, err := s.auditStore.List(r.Context(), audit.ListFilter{
		Action:   q.Get("action"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"msg": "failed to list audit logs"})
		return
	}
	auth.WriteJSON(w, http.StatusOK, result)
}
