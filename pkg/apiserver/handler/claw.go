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

// Package handler provides HTTP handlers for the /claw/* routes.
// Handlers decode requests, call ClawService, and encode responses.
// No K8s or business logic lives here.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/apiserver/service"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/auth"
)

// ClawHandler handles /claw/* HTTP routes.
type ClawHandler struct {
	svc *service.ClawService
}

// NewClawHandler creates a ClawHandler backed by the given service.
func NewClawHandler(svc *service.ClawService) *ClawHandler {
	return &ClawHandler{svc: svc}
}

// RegisterRoutes registers all /claw/* routes on mux.
func (h *ClawHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /claw/alloc", h.Alloc)
	mux.HandleFunc("POST /claw/free", h.Free)
	mux.HandleFunc("POST /claw/pause", h.Pause)
	mux.HandleFunc("POST /claw/resume", h.Resume)
	mux.HandleFunc("GET /claw/instances", h.ListInstances)
	mux.HandleFunc("GET /claw/instances/{name}", h.GetInstance)
	mux.HandleFunc("POST /claw/instances/{name}/exec", h.ExecInstance)
	mux.HandleFunc("GET /claw/used", h.ListByUser)
	mux.HandleFunc("GET /claw/token", h.GetToken)
}

// ── request / response types ──────────────────────────────────────────────────

type allocRequest struct {
	UserID    string `json:"user_id"`
	ModelType string `json:"model_type,omitempty"` // "lite" | "pro"
}

type freeRequest struct {
	Name   string `json:"name"`    // required
	UserID string `json:"user_id"` // optional
}

type pauseRequest struct {
	Name         string `json:"name"`          // required
	UserID       string `json:"user_id"`       // optional
	DelayMinutes int    `json:"delay_minutes"` // optional, 0 = immediate
}

type resumeRequest struct {
	Name   string `json:"name"`    // required
	UserID string `json:"user_id"` // optional
}

type execRequest struct {
	Command   []string `json:"command"`             // required, e.g. ["bash", "-c", "ls"]
	Container string   `json:"container,omitempty"` // optional, defaults to primary container
}

type execResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// ── handlers ─────────────────────────────────────────────────────────────────

// POST /claw/alloc
func (h *ClawHandler) Alloc(w http.ResponseWriter, r *http.Request) {
	var req allocRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if req.ModelType == "" {
		req.ModelType = "lite"
	}

	claims, _ := auth.ClaimsFromContext(r.Context())

	inst, err := h.svc.Alloc(r.Context(), claims.UserID, claims.Username, req.UserID, req.ModelType)
	if errors.Is(err, adapter.ErrUserAlreadyHasInstance) {
		writeJSON(w, http.StatusConflict, map[string]string{"msg": "user already has an instance"})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, instanceResponse(inst))
}

// POST /claw/free
func (h *ClawHandler) Free(w http.ResponseWriter, r *http.Request) {
	var req freeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	claims, _ := auth.ClaimsFromContext(r.Context())

	if _, err := h.svc.Free(r.Context(), claims.UserID, claims.Username, req.Name, req.UserID); err != nil {
		if errors.Is(err, adapter.ErrNotFound) {
			writeError(w, http.StatusNotFound, "instance not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"msg": "success"})
}

// POST /claw/pause
func (h *ClawHandler) Pause(w http.ResponseWriter, r *http.Request) {
	var req pauseRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	claims, _ := auth.ClaimsFromContext(r.Context())

	if _, err := h.svc.Pause(r.Context(), claims.UserID, claims.Username, req.Name, req.UserID, req.DelayMinutes); err != nil {
		if errors.Is(err, adapter.ErrNotFound) {
			writeError(w, http.StatusNotFound, "instance not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"msg": "success"})
}

// POST /claw/resume
func (h *ClawHandler) Resume(w http.ResponseWriter, r *http.Request) {
	var req resumeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	claims, _ := auth.ClaimsFromContext(r.Context())

	if _, err := h.svc.Resume(r.Context(), claims.UserID, claims.Username, req.Name, req.UserID); err != nil {
		if errors.Is(err, adapter.ErrNotFound) {
			writeError(w, http.StatusNotFound, "instance not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"msg": "success"})
}

// GET /claw/instances
func (h *ClawHandler) ListInstances(w http.ResponseWriter, r *http.Request) {
	instances, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Optional occupied filter: ?occupied=true|false
	occupiedParam := r.URL.Query().Get("occupied")
	result := make([]instanceResp, 0, len(instances))
	for _, inst := range instances {
		occupied := inst.UserID != ""
		if occupiedParam == "true" && !occupied {
			continue
		}
		if occupiedParam == "false" && occupied {
			continue
		}
		result = append(result, instanceResponse(inst))
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

// GET /claw/instances/{name}
func (h *ClawHandler) GetInstance(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "instance name is required")
		return
	}

	inst, err := h.svc.GetByName(r.Context(), name)
	if errors.Is(err, adapter.ErrNotFound) {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": instanceResponse(inst)})
}

// GET /claw/used?user_id=<id>
func (h *ClawHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	if userID != "" {
		inst, err := h.svc.GetByUser(r.Context(), userID)
		if errors.Is(err, adapter.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"data": []instanceResp{}})
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []instanceResp{instanceResponse(inst)}})
		return
	}

	// No user_id: return all occupied
	all, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]instanceResp, 0)
	for _, inst := range all {
		if inst.UserID != "" {
			result = append(result, instanceResponse(inst))
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

// POST /claw/instances/{name}/exec
func (h *ClawHandler) ExecInstance(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "instance name is required")
		return
	}

	var req execRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.Command) == 0 {
		writeError(w, http.StatusBadRequest, "command is required")
		return
	}

	result, err := h.svc.Exec(r.Context(), name, req.Command, req.Container)
	if errors.Is(err, adapter.ErrNotFound) {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}
	if err != nil && result == nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := execResponse{ExitCode: 0}
	if result != nil {
		resp.Stdout = result.Stdout
		resp.Stderr = result.Stderr
		resp.ExitCode = result.ExitCode
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /claw/token?name=<instance-name>
func (h *ClawHandler) GetToken(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	inst, err := h.svc.GetByName(r.Context(), name)
	if errors.Is(err, adapter.ErrNotFound) {
		writeJSON(w, http.StatusOK, map[string]string{"token": ""})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": inst.Token})
}

// ── response type & helpers ───────────────────────────────────────────────────

type resourcesResp struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

type instanceResp struct {
	Name        string        `json:"name"`
	UserID      string        `json:"user_id"`
	GatewayURL  string        `json:"claw_webui_url"`
	ClawWSSURL  string        `json:"claw_wss_url"`
	Occupied    bool          `json:"occupied"`
	State       string        `json:"state"`
	AllocStatus string        `json:"alloc_status"`
	Resources   resourcesResp `json:"resources"`
	Token       string        `json:"token,omitempty"`
}

func instanceResponse(inst *adapter.Instance) instanceResp {
	return instanceResp{
		Name:        inst.Name,
		UserID:      inst.UserID,
		GatewayURL:  inst.AccessURL,
		ClawWSSURL:  inst.WssURL,
		Occupied:    inst.UserID != "",
		State:       string(inst.State),
		AllocStatus: string(inst.AllocStatus),
		Resources: resourcesResp{
			CPURequest:    inst.Resources.CPURequest,
			CPULimit:      inst.Resources.CPULimit,
			MemoryRequest: inst.Resources.MemoryRequest,
			MemoryLimit:   inst.Resources.MemoryLimit,
		},
		Token: inst.Token,
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"msg": msg})
}
