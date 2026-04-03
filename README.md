**[English](README.md) | [中文](README_zh.md)**

---

# 🦞 Claw Swarm Panel

**Web management panel for [OpenClaw](https://github.com/openclaw/openclaw)🦞 — HTTP API server + React dashboard, built for small and medium-sized businesses.**

Give every employee a dedicated, isolated AI coding workspace — allocated in under 500 ms, with automatic resource recycling and zero ops overhead.

> **Project split notice:** This repository ships only the **management panel** (API server + Web UI).
> The Kubernetes adapter and instance lifecycle controller live in a separate project:
> **[claw-swarm-operator](https://gitlab.botnow.cn/agentic/claw-swarm-operator)** — deploy it into your cluster first.

---

## Why Claw Swarm Panel?

The biggest challenge for SMBs adopting OpenClaw is not technical capability — it is **security and operational cost**: how to guarantee data isolation between employees, how to prevent API key leakage, and how to run reliably without growing the ops team.

### The Pain of Manual Management

| Problem | Cost of Doing It Manually |
|---------|--------------------------|
| Slow onboarding | 5–15 minutes per employee; no access on day one |
| Weak isolation | Shared environments expose code and config across users |
| API key risk | One leaked shared key compromises everyone |
| Resource waste | Containers run 24/7, burning compute overnight and on weekends |
| High ops barrier | Requires a dedicated engineer; hard to delegate to IT |
| Tedious model config | Each container configured individually; batch changes are impractical |

### What Claw Swarm Panel Delivers

| Core Value | How It Works |
|-----------|--------------|
| **Complete employee data isolation** | Each person owns a dedicated workspace container — data never crosses boundaries |
| **Per-employee API key control** | Integrates with LiteLLM to auto-generate individual keys per user; keys are revoked automatically on offboarding |
| **Allocation in < 500 ms** | Pre-warmed instance pool stays ready; allocation API responds in under 500 ms — no cold-start wait |
| **Automatic suspend & resume** | Idle instances scale to zero; resume on demand without data loss |
| **IT-operable, no infra expertise needed** | Web management dashboard covers all day-to-day operations |
| **Easy system integration** | REST API + API Secret auth — connect HR systems, Slack/Teams bots, SSO platforms |

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    claw-swarm-panel                           │
│                                                               │
│  ┌─────────────────┐    ┌────────────────────────────────┐   │
│  │   Web Dashboard  │    │       Management API           │   │
│  │   (React, :80)   │◄──►│  /auth  /users  /claw  (:8088)│   │
│  └─────────────────┘    └──────────────┬───────────────--┘   │
│                                        │                      │
│                          ┌─────────────▼─────────────────┐   │
│                          │      adapter.ClawAdapter       │   │
│                          │  ┌──────────────────────────┐  │   │
│                          │  │  k8s adapter (built-in)  │  │   │
│                          │  └────────────┬─────────────┘  │   │
│                          └──────────────┼─────────────────┘   │
│  ┌─────────────────┐                    │                      │
│  │  SQLite (claw.db)│                   │                      │
│  │  - users         │                   │                      │
│  │  - audit_logs    │                   │                      │
│  │  - claw_tokens   │                   │                      │
│  └─────────────────┘                   │                      │
└───────────────────────────────────────┼─────────────────────┘
                                        │ Kubernetes API
                    ┌───────────────────▼──────────────────┐
                    │         claw-swarm-operator           │
                    │   ClawReconciler — pool management    │
                    │   PVCReconciler  — storage GC         │
                    └───────────────────┬──────────────────┘
                                        │
          ┌──────────┬──────────────────┼──────────────────┐
          ▼          ▼                  ▼                   ▼
     Employee A  Employee B        Employee C           LiteLLM
     Workspace   Workspace         Workspace       (AI model routing
      + Storage   + Storage         + Storage       & API key mgmt)
```

| Component | Repo | Role |
|-----------|------|------|
| **claw-swarm-panel** | this repo | HTTP API + Web UI — instance operations, user management |
| **claw-swarm-operator** | separate repo | K8s controller — maintains instance pool, manages lifecycle |

---

## Features

- **Full lifecycle control** — alloc / free / pause (with optional delay) / resume
- **Employee data isolation** — each person gets a dedicated container + persistent storage; data is never shared
- **LiteLLM integration** — auto-provision user + API key on allocation; async model config injection; graceful degradation when LiteLLM is unavailable
- **Dual authentication** — JWT (interactive web sessions) + API Secret (system integrations)
- **Role-based access** — `admin` (full control) and `user` (read-only)
- **Audit logging** — every alloc / free / pause / resume / user-management action is recorded
- **Gateway token management** — per-instance access tokens stored in the database, returned on allocation

---

## Quick Start

### Prerequisites

Deploy `claw-swarm-operator` into your Kubernetes cluster first — it manages the instance pool that this panel controls.

### Run locally

```bash
make run          # start API server (port 8088)
make run-webui    # start Web UI dev server (port 5173)
```

### Build

```bash
make build        # outputs: bin/apiserver
make build-webui  # outputs: webui/dist/
```

---

## Configuration

All settings are read from `config.yaml` (pass via `--config`).

```yaml
# ── API Server ─────────────────────────────────────────────────────────────────
server:
  port: 8088
  data_dir: ".claw-swarm"        # SQLite DB + JWT secret location

# ── Adapter backend ────────────────────────────────────────────────────────────
adapter:
  type: "k8s"
  k8s:
    kubeconfig: ""               # explicit path; empty = ~/.kube/config or in-cluster
    namespace: ""                # falls back to POD_NAMESPACE env var

# ── LiteLLM Integration ────────────────────────────────────────────────────────
init:
  litellm:
    baseurl: "http://litellm.example.com"
    master_key: "sk-..."
    default_team: ""
    default_max_budget: 1
    models:
      lite:
        model_id: "tabtab-lite"
      pro:
        model_id: "tabtab-pro"
```

---

## API Reference

All endpoints require `Authorization: Bearer <jwt>` or `X-API-Key: claw_...` except `/auth/login`.

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/login` | Login — returns JWT (24 h) |
| `GET`  | `/auth/me` | Current user profile |
| `POST` | `/auth/change-password` | Change password |
| `GET`  | `/auth/api-secret` | View API secret |
| `POST` | `/auth/api-secret/regenerate` | Regenerate API secret |

### Instance Lifecycle (admin only for writes)

| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/claw/instances` | List all instances (`?occupied=true\|false`) |
| `GET`  | `/claw/used` | List allocated instances |
| `GET`  | `/claw/token` | Get gateway token (`?name=<name>`) |
| `POST` | `/claw/alloc` | Allocate instance to user |
| `POST` | `/claw/free` | Free (delete) instance |
| `POST` | `/claw/pause` | Pause instance (supports `delay_minutes`) |
| `POST` | `/claw/resume` | Resume paused instance |

**Alloc:**
```json
{ "user_id": "alice", "model_type": "lite" }
```

**Free / Resume:**
```json
{ "name": "claw-a1b2c3d4" }
```

**Pause with delay:**
```json
{ "name": "claw-a1b2c3d4", "delay_minutes": 10 }
```

**Instance response:**
```json
{
  "name": "claw-a1b2c3d4",
  "user_id": "alice",
  "state": "running",
  "alloc_status": "allocated",
  "access_url": "https://claw-claw-a1b2c3d4.example.com/overview",
  "token": "sk-...",
  "resources": {
    "cpu_request": "250m", "cpu_limit": "1",
    "memory_request": "512Mi", "memory_limit": "2Gi"
  },
  "created_at": "2026-03-25T10:00:00Z"
}
```

**States:** `running` | `pending` | `paused` | `deleting`

**Alloc status:** `allocating` (model config in progress) → `allocated` (ready to use)

### User Management (admin only)

| Method | Path | Description |
|--------|------|-------------|
| `GET`    | `/users`     | List all panel users |
| `POST`   | `/users`     | Create user (username, password, role) |
| `DELETE` | `/users/:id` | Delete user |

---

## Deployment (Helm)

**Install from OCI registry (recommended):**

> The following parameters must be adjusted to match your environment:
>
> | Parameter | Description |
> |-----------|-------------|
> | `ingress.hosts[0].host` | Domain where the panel Web UI and API will be exposed |
> | `ingress.className` | Ingress class matching your cluster's ingress controller (e.g. `kong`, `nginx`) |
> | `claw.init.litellm.baseurl` | LiteLLM service URL for AI model routing |
> | `claw.init.litellm.master_key` | LiteLLM master key |

```bash
helm install claw-swarm-panel oci://registry-1.docker.io/tabtabai/claw-swarm-panel-chart \
  --namespace tabclaw --create-namespace \
  --set ingress.hosts[0].host=claw-panel.example.com \
  --set ingress.className=kong \
  --set claw.init.litellm.baseurl=http://litellm.example.com \
  --set claw.init.litellm.master_key=sk-...
```

**Install from local source:**

```bash
helm install claw-swarm-panel charts/claw-swarm-panel \
  --namespace tabclaw --create-namespace \
  -f values.yaml
```

Key `values.yaml` options:

```yaml
image:
  repository: registry.example.com/agentic/claw-swarm-panel
  tag: v1.0.0

service:
  type: ClusterIP
  port: 8088

ingress:
  enabled: true
  className: "kong"
  hosts:
    - host: claw-panel.example.com
      paths:
        - path: /
          pathType: Prefix

claw:
  server:
    domain: claw.example.com
    data_dir: /data/.claw-swarm
  init:
    litellm:
      baseurl: "http://litellm.example.com"
      master_key: "sk-..."
```

```bash
make helm-lint     # validate
make helm-package  # package to .tgz
make helm-push     # push to OCI registry
```

---

## Integration

Connect Claw Swarm Panel to external systems via `X-API-Key` authentication:

| System | Trigger | API Call |
|--------|---------|---------|
| HR / Onboarding | Employee joins | `POST /claw/alloc` |
| Offboarding | Employee leaves | `POST /claw/free` |
| Slack / Teams bot | User's first conversation | `POST /claw/alloc` (idempotent) |
| Monitoring | Periodic health check | `GET /claw/instances` |

---

## Make Targets

```
Development
  fmt / vet       Format and vet Go code
  lint            Run golangci-lint
  lint-fix        Run golangci-lint with auto-fix

Frontend
  install-webui   Install webui dependencies (pnpm)
  run-webui       Start Vite dev server (:5173)
  build-webui     Build webui for production

Build
  build           Build apiserver binary (bin/apiserver)
  run             Run apiserver from host

Docker
  docker-build    Build container image (IMG)
  docker-push     Push container image (IMG)

Helm
  helm-lint       Lint helm chart
  helm-package    Package helm chart
  helm-push       Push chart to OCI registry
```

---

## Default Credentials

| Username | Password | Note |
|----------|----------|------|
| `admin` | `happyclaw` | Must be changed on first login |

---

## Troubleshooting

**`unable to load in-cluster configuration`**
Set `adapter.k8s.kubeconfig` in `config.yaml`, or ensure `~/.kube/config` points to a reachable cluster.

**`address already in use` on port 8088**
Change `server.port` in `config.yaml` and update the proxy target in `webui/vite.config.ts`.

**Empty instance list after startup**
The instance pool is managed by `claw-swarm-operator`. Verify it is running in the cluster:
```bash
kubectl logs -l app.kubernetes.io/component=operator -n tabclaw -f
```

**Allocation stuck at `alloc_status: "allocating"`**
The async model config step (exec → `openclaw config set`) is still running or failed silently. Check `claw-swarm-operator` logs for `[alloc]` entries. If LiteLLM is unavailable, a warning is logged and a random API key is used — allocation still completes.
