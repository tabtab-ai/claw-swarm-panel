# Product Requirements Document — Claw Swarm Panel

**Version:** 1.0
**Date:** 2026-03-19
**Status:** Living Document

---

## Table of Contents

1. [Product Overview](#1-product-overview)
2. [Goals & Non-Goals](#2-goals--non-goals)
3. [User Roles](#3-user-roles)
4. [System Architecture](#4-system-architecture)
5. [Features](#5-features)
   - 5.1 Authentication & Security
   - 5.2 User Management
   - 5.3 Instance Pool Management
   - 5.4 Instance Lifecycle Operations
   - 5.5 Gateway Token Management
   - 5.6 LiteLLM Integration
   - 5.7 Web UI
6. [Data Models](#6-data-models)
7. [API Summary](#7-api-summary)
8. [Configuration](#8-configuration)
9. [Deployment](#9-deployment)
10. [Integration Points](#10-integration-points)
11. [Non-Functional Requirements](#11-non-functional-requirements)
12. [Future Scope](#12-future-scope)

---

## 1. Product Overview

**Claw Swarm Panel** (also referred to as `claw-swarm-operator`) is a Kubernetes-native management platform for provisioning, allocating, and operating pools of **OpenClaw** AI coding environment containers.

### Problem Statement

AI coding assistants powered by OpenClaw require isolated, stateful container environments per user session. Manually provisioning and managing these environments is error-prone, slow, and does not scale. Platform operators need a centralized system to:

- Maintain a ready pool of pre-warmed OpenClaw instances
- Allocate instances to users on demand with sub-second API response time
- Recycle idle instances automatically to conserve resources
- Integrate seamlessly with AI model routing (LiteLLM)
- Provide visibility into pool health via a web dashboard

### Value Proposition

| Benefit | Description |
|---------|-------------|
| **Instant allocation** | Pre-warmed pool means instances are ready immediately on request |
| **Automatic recycling** | Idle instances are paused and returned to the pool after a configurable timeout |
| **LiteLLM-native** | Each allocation automatically provisions a LiteLLM user and API key |
| **Role-based access** | Admin / User roles with JWT and API secret authentication |
| **Kubernetes-native** | Operator pattern ensures declarative, self-healing instance management |
| **Web dashboard** | Real-time pool visibility and one-click instance management |

### Target Users

- **Platform operators / SREs**: Deploy and configure the system, manage users, monitor pool health
- **Automation systems**: Allocate and free instances via API Key, integrate with chat/agent platforms
- **End users (indirect)**: Access their allocated OpenClaw instance via the gateway URL

---

## 2. Goals & Non-Goals

### Goals

- Manage a pool of OpenClaw StatefulSets on Kubernetes
- Expose a REST API for instance lifecycle operations (alloc / free / pause / resume)
- Provide an admin web UI for pool visibility and management
- Integrate with LiteLLM for per-user API key provisioning
- Support both interactive (JWT) and automated (API secret) clients
- Persist user data and tokens across restarts via SQLite
- Support Alibaba Cloud ACK as a deployment target with image caching optimizations

### Non-Goals

- Multi-cluster orchestration (single cluster per deployment)
- Real-time CPU / memory metrics collection (display fields are placeholders)
- OpenClaw application development or runtime behaviour
- Billing or metering of instance usage
- High availability (single-replica operator; SQLite is not distributed)

---

## 3. User Roles

### Admin

Full access to all API endpoints and all web UI pages.

| Capability | Details |
|-----------|---------|
| Manage users | Create, list, delete users |
| Allocate instances | `POST /claw/alloc` |
| Free instances | `POST /claw/free` |
| Pause / resume instances | `POST /claw/pause`, `POST /claw/resume` |
| View gateway tokens | `GET /claw/token` |
| View all instances | `GET /claw/instances`, `GET /claw/used` |
| ACK operations | Warmup jobs, PV inspection |

### User (Read-only)

Limited access — can observe pool state but cannot modify instances.

| Capability | Details |
|-----------|---------|
| View all instances | `GET /claw/instances`, `GET /claw/used` |
| View own profile | `GET /auth/me` |
| Manage own credentials | Change password, view / regenerate API secret |

### Default Account

On first startup the system creates:

- **Username:** `admin`
- **Password:** `happyclaw`
- **Flag:** `force_change_password = true` (user is prompted to change on first login)

---

## 4. System Architecture

```
┌────────────────────────────────────────────────────────────┐
│                     Claw Swarm Panel                       │
│                                                            │
│  ┌──────────────┐    ┌──────────────────────────────────┐  │
│  │   React UI   │    │        HTTP Management API        │  │
│  │  (Port 80)   │◄──►│   /auth  /users  /claw  (8088)  │  │
│  └──────────────┘    └──────────────────┬───────────────┘  │
│                                         │                  │
│                       ┌─────────────────▼──────────────┐   │
│                       │       K8s Operator              │   │
│                       │   (controller-runtime)          │   │
│                       │                                 │   │
│                       │  ClawReconciler ─► Pool Sync    │   │
│                       │  PVCReconciler  ─► Volume GC    │   │
│                       └─────────────────┬───────────────┘   │
│                                         │                   │
│  ┌──────────────────┐   ┌───────────────▼─────────────┐    │
│  │  SQLite DB       │   │       Kubernetes API          │    │
│  │  (claw.db)       │   │  StatefulSets / Services /    │    │
│  │  - users         │   │  Ingresses / PVCs             │    │
│  │  - claw_tokens   │   └───────────────────────────────┘    │
│  └──────────────────┘                                        │
└────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────▼──────────────┐
              │         LiteLLM              │
              │  AI model routing & API keys  │
              └──────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility |
|-----------|---------------|
| **React Web UI** | Dashboard, instance list/detail, user management, auth flows |
| **HTTP API** | REST interface for all instance and user operations |
| **K8s Operator (ClawReconciler)** | Pool size reconciliation, instance creation/deletion, lifecycle |
| **K8s Operator (PVCReconciler)** | PVC cleanup for deleted instances (Redis-backed tracking) |
| **SQLite** | User records, bcrypt passwords, API secrets, gateway tokens |
| **LiteLLM** | AI model API key provisioning and user management |

---

## 5. Features

### 5.1 Authentication & Security

#### Two Authentication Methods

**JWT Bearer Token** (interactive clients, web UI)
- Issued on successful `POST /auth/login`
- Algorithm: HS256, signed with 32-byte secret
- Expiry: 24 hours
- Header: `Authorization: Bearer <token>`
- Secret persisted to `<data_dir>/.jwt_secret` for cross-restart validity

**API Secret** (automation, scripts, integrations)
- Format: `claw_<40 hex chars>` (20 random bytes)
- Never expires; regenerable on demand
- Header: `X-API-Key: claw_...`
- If both headers are present, JWT is checked first

#### Password Security
- Stored as bcrypt hash (cost 10)
- Minimum new password length: 6 characters
- `force_change_password` flag prompts change on first login

#### Role Enforcement
- All write endpoints enforce admin role at the HTTP handler level
- Read endpoints (`/claw/instances`, `/claw/used`) are accessible to any authenticated user
- Non-admin write attempts return `403 Forbidden`

---

### 5.2 User Management

Admins manage users via the web UI or API.

#### Create User
- Fields: username (unique), password, role (`admin` | `user`)
- System auto-generates an API secret on creation
- Default role: `user` if not specified

#### List Users
- Returns all users; API secret is included for admin self-inspection

#### Delete User
- Cannot delete own account (self-deletion protection)

#### User Self-Service
- **Change password**: old + new password required; issues a fresh JWT
- **View API secret**: accessible via web UI dropdown or `GET /auth/api-secret`
- **Regenerate API secret**: old secret is immediately invalidated

---

### 5.3 Instance Pool Management

The K8s operator maintains a pool of idle OpenClaw instances.

#### Pool Reconciliation (ClawReconciler)

Triggered on every StatefulSet event. The reconciler:

1. Lists all StatefulSets labelled `tabtabai.com/tabclaw`
2. Counts how many are **unoccupied** (free)
3. If free count < `pool_size`: creates new StatefulSets to fill the gap
4. If free count > `pool_size`: deletes surplus idle instances
5. If `RuntimeImage` changed: patches all running instances to the new image

#### Per-Instance Kubernetes Resources

Each instance consists of:

| Resource | Name Pattern | Notes |
|----------|-------------|-------|
| StatefulSet | `<generated-name>` | Owner of all other resources |
| Pod | `<sts-name>-0` | OpenClaw container |
| PersistentVolumeClaim | `<sts-name>-openclaw-config-0` | Workspace storage, 2Gi default |
| Service | `<sts-name>` | ClusterIP, port 18789 |
| Ingress | `<sts-name>` | Routes `/<sts-name>/` path to service |

#### Instance States

| K8s State | API `state` | Meaning |
|-----------|------------|---------|
| Replicas > 0, all ready | `running` | Pod is running and ready |
| Replicas > 0, not ready | `pending` | Pod starting up |
| Replicas = 0 | `paused` | Suspended / scaled to zero |
| Scheduled deletion annotation set | `paused` | Pause timer in progress |

#### Allocation Status Label

| Label Value | `alloc_status` | Meaning |
|-------------|---------------|---------|
| _(absent)_ | `""` | Instance is free |
| `allocating` | `"allocating"` | Allocation in progress (model config running) |
| `allocated` | `"allocated"` | Fully allocated and configured |

---

### 5.4 Instance Lifecycle Operations

#### Allocate (`POST /claw/alloc`)

1. Lock acquired (prevents concurrent races)
2. Find earliest-created free StatefulSet from pool
3. If user already has an instance, return it (idempotent)
4. Set labels: `tabclaw-occupied = <user_id>`, `tabclaw-alloc-status = allocating`
5. Set replicas = 1 if paused; persist to K8s
6. **Return response immediately** with `alloc_status: "allocating"`
7. Async goroutine: run `configurePodModels` (exec into pod, set LiteLLM provider)
8. Async goroutine: on completion, update label to `tabclaw-alloc-status = allocated`

> **Rationale for async model config**: `configurePodModels` may take several seconds (exec into pod + HTTP calls). The caller receives the allocation result instantly and can poll `GET /claw/instances` to observe `alloc_status` transition.

#### Free (`POST /claw/free`)

- Delete the StatefulSet; K8s owner references cascade-delete Service and Ingress
- PVC is retained by default (operator manages separate cleanup)
- Operator reconciler detects the deletion and creates a new idle instance to refill the pool

#### Pause (`POST /claw/pause`)

- Sets annotation `tabtab.app.scheduled.deletion.time = <now + delay_pause_minutes>`
- Operator reconciler detects the annotation and scales StatefulSet to 0 replicas after the delay
- While the annotation is present, `state` is reported as `paused`

#### Resume (`POST /claw/resume`)

- Removes scheduled-deletion annotations
- Sets replicas = 1
- State returns to `pending` → `running` as pod starts

---

### 5.5 Gateway Token Management

Each OpenClaw instance may have an associated gateway authentication token stored in the `claw_tokens` SQLite table. These tokens are provisioned externally (by the operator or init process) and allow users to authenticate with the OpenClaw gateway.

#### Token Storage Schema

```sql
CREATE TABLE claw_tokens (
  name       TEXT PRIMARY KEY,   -- StatefulSet name
  token      TEXT NOT NULL,
  created_at DATETIME DEFAULT (datetime('now','localtime'))
)
```

#### Token Behaviour

- Tokens are **persistent** — they remain in the table until explicitly removed
- The `GET /claw/token?name=<name>` endpoint reads the token without consuming it
- An empty token indicates the instance has not yet been assigned a gateway token
- The alloc response (`POST /claw/alloc`) also includes the token if one is present

#### UI Representation

- **Instance Detail page**: Always shows the "Gateway Token" card
  - Empty state: `未分配 — 暂无 Gateway Token` (instance not yet provisioned with a token)
  - Populated: token value with copy button
- **Alloc Result page**: Shows token if returned in allocation response

---

### 5.6 LiteLLM Integration

On each allocation the system provisions an AI model API key for the user via LiteLLM.

#### Flow

1. Determine model tier based on `model_type` param (`lite` or `pro`, default `lite`)
2. Call LiteLLM `POST /user/new` with master key to create user or retrieve existing key
3. Receive API key; inject into OpenClaw pod via `openclaw config set` exec command
4. Set `agents.defaults.model.primary` to `tabtab-litellm/<model_id>`

#### Model Configuration

```yaml
litellm:
  baseurl: "http://litellm.example.com"
  master_key: "sk-..."
  default_team: "team-uuid"
  models:
    lite:
      model_id: "tabtab-lite"
    pro:
      model_id: "tabtab-pro"
```

#### Graceful Degradation

- If LiteLLM is unreachable or not configured: a random 32-char API key is generated and allocation proceeds
- Allocation is never blocked by LiteLLM failures

---

### 5.7 Web UI

Built with React 18, TypeScript, Zustand (state), shadcn/ui + Tailwind CSS (design system).

#### Pages

##### Login (`/login`)
- Username and password fields
- On success: JWT stored in `localStorage`, redirect to Dashboard
- Force-change-password flow: prompted after first login

##### Dashboard (`/`)
- Summary cards: Total Instances, Active, Free, Allocated count
- Instance list (first 5 by recency) with status indicators
- Allocation filter: All / Allocated / Free
- Status filter: All / Running / Pending / Paused
- Error banner for instances in error state
- Quick-action buttons: allocate, pause/resume, free

##### Instance List (`/instances`)
- Grid of all instances
- Search by instance name or user ID
- Per-instance: status dot, name, user ID (with copy), alloc badge, WebUI link, resource summary
- Quick actions: pause/resume, free

##### Instance Detail (`/instances/:id`)
Sections:
1. **Header**: name, status, allocation badge (Free / Allocating / Allocated), Pause/Resume/Free buttons
2. **Basic Info**: name, status, allocation state with colour-coded badge
   - Yellow `Allocating` → Cyan `Allocated` → Grey `Free`
3. **User Assignment**: assigned user ID with copy button; "Not assigned" if free
4. **Resources**: CPU request/limit, memory request/limit
5. **Claw Web UI**: gateway URL with copy + external link
6. **Gateway Token**: token value with copy button, or "未分配" placeholder

##### Create Instance / Alloc Finish (`/instances/new`)
- Step 1 (Form): Enter user ID → Submit
- Step 2 (Result): Shows allocated instance name, state, user ID, gateway token (if present), web UI URL

##### User Management (`/users`) — admin only
- Table: ID, username, role badge, force_change_password, created_at
- Create user dialog: username, password (min 6 chars), role select
- Delete user (with confirmation; self-delete protected)

#### Navigation
- Sidebar: Dashboard, Instances, Create Instance, Users (admin only)
- User dropdown:
  - Current user name + role badge
  - Change Password dialog
  - API Secret dialog (show / copy / regenerate)
  - Logout

#### State Management (Zustand)

| Store | State | Actions |
|-------|-------|---------|
| `authStore` | `user`, `token`, `isAuthenticated` | `login`, `logout`, `changePassword` |
| `clawStore` | `instances[]`, `loading`, `error`, `occupiedFilter` | `refresh`, `createInstance`, `deleteInstance`, `updateInstanceStatus`, `occupyInstance`, `releaseInstance` |

---

## 6. Data Models

### Instance (API Response)

```json
{
  "name":           "claw-abc123",
  "user_id":        "user-xyz789",
  "claw_webui_url": "https://claw.example.com/claw-abc123/overview",
  "occupied":       true,
  "state":          "running",
  "alloc_status":   "allocated",
  "token":          "gateway-token-value",
  "resources": {
    "cpu_request":    "250m",
    "cpu_limit":      "1",
    "memory_request": "512Mi",
    "memory_limit":   "2Gi"
  },
  "created_at": "2026-03-19T10:00:00Z"
}
```

### User (API Response)

```json
{
  "id": 1,
  "username": "admin",
  "role": "admin",
  "force_change_password": false,
  "api_secret": "claw_a3f2e1d4b5c6...",
  "created_at": "2026-01-01T00:00:00Z"
}
```

### K8s Labels

| Label Key | Value | Meaning |
|-----------|-------|---------|
| `tabtabai.com/tabclaw` | `""` | Marks operator-managed instances |
| `tabtabai.com/tabclaw-occupied` | `<user_id>` | Assigned to this user |
| `tabtabai.com/tabclaw-alloc-status` | `allocating` \| `allocated` | Allocation phase |
| `tabtabai.com/tabclaw-init` | `""` | Init-trigger StatefulSet marker |

### K8s Annotations

| Annotation Key | Value | Meaning |
|----------------|-------|---------|
| `tabtab.app.scheduled.deletion.time` | RFC3339 datetime | Pause scheduled at this time |
| `tabtab.app.scheduled.deletion` | any | Pause trigger flag |

---

## 7. API Summary

Full documentation: [`docs/api.md`](./api.md)

### Auth Endpoints (public / authenticated)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/auth/login` | Public | Obtain JWT token |
| `GET` | `/auth/me` | Any | Current user profile |
| `POST` | `/auth/change-password` | Any | Change password |
| `GET` | `/auth/api-secret` | Any | View API secret |
| `POST` | `/auth/api-secret/regenerate` | Any | Regenerate API secret |

### User Management (admin only)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/users` | List all users |
| `POST` | `/users` | Create user |
| `DELETE` | `/users/:id` | Delete user |

### Instance Operations

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/claw/instances` | Any | List all instances (optional `?occupied=true\|false`) |
| `GET` | `/claw/used` | Any | List allocated instances (optional `?user_id=`) |
| `GET` | `/claw/token` | Admin | Get gateway token for instance (`?name=`) |
| `POST` | `/claw/alloc` | Admin | Allocate instance to user |
| `POST` | `/claw/free` | Admin | Delete and free instance |
| `POST` | `/claw/pause` | Admin | Schedule instance pause |
| `POST` | `/claw/resume` | Admin | Resume paused instance |

### ACK-only (admin, `ack.enabled: true`)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/claw/warnup-job` | Create image warmup job |
| `POST` | `/claw/remove-warnup-job` | Delete warmup job |
| `POST` | `/claw/pv` | Get PV name for user |

---

## 8. Configuration

Configuration file: `config.yaml` (path settable via `--config` flag). CLI flags override file values.

### `server` block

| Key | Default | Description |
|-----|---------|-------------|
| `port` | `8088` | HTTP API listen port |
| `domain` | `claw.10.20.51.23.nip.io` | Base domain for instance gateway URLs |
| `tls_secret_name` | `tabtabai` | K8s TLS secret name for ingress |
| `ingress_name` | `portal-ingress` | Ingress class / resource name |
| `delay_pause_minutes` | `1` | Minutes of inactivity before auto-pause |
| `node_selector_value` | `""` | Node label value for pod scheduling (empty = no restriction) |
| `jwt_secret` | `""` | JWT signing secret (auto-generated if empty) |
| `data_dir` | `.claw-swarm` | Directory for SQLite DB and JWT secret file |

### `operator` block

| Key | Default | Description |
|-----|---------|-------------|
| `runtime_image` | `ghcr.io/openclaw/openclaw:latest` | Container image for OpenClaw pods |
| `storage_class` | `standard` | StorageClass for PVCs |

### `init` block

| Key | Default | Description |
|-----|---------|-------------|
| `pool_size` | `2` | Target number of idle (unallocated) instances |
| `storage_size` | `2Gi` | PVC capacity per instance |
| `gateway.port` | `18789` | OpenClaw gateway port |
| `gateway.auth` | `token` | Gateway auth mode (`token` or `none`) |
| `gateway.token` | `123456` | Gateway bearer token |
| `env` | _(see below)_ | Environment variables injected into all pods |
| `litellm.baseurl` | — | LiteLLM service base URL |
| `litellm.master_key` | — | LiteLLM admin master key |
| `litellm.default_team` | — | Default LiteLLM team UUID for new users |
| `litellm.models.lite.model_id` | `tabtab-lite` | Model ID for lite tier |
| `litellm.models.pro.model_id` | `tabtab-pro` | Model ID for pro tier |

### Default Pod Environment Variables

| Variable | Default Value | Purpose |
|----------|--------------|---------|
| `OPENCLAW_GATEWAY_BIND` | `lan` | Network interface binding |
| `OPENCLAW_WORKSPACE_DIR` | `/home/node/.openclaw/workspace` | User workspace directory |
| `OPENCLAW_CONFIG_DIR` | `/home/node/.openclaw` | OpenClaw config directory |
| `OPENCLAW_SANDBOX` | `1` | Enable sandbox mode |

### `resources` block

Standard K8s `resources` object applied to every OpenClaw container.

| Key | Default |
|-----|---------|
| `requests.cpu` | `250m` |
| `requests.memory` | `512Mi` |
| `limits.cpu` | `1` |
| `limits.memory` | `2Gi` |

### `ack` block (optional, Alibaba Cloud ACK)

| Key | Default | Description |
|-----|---------|-------------|
| `enabled` | `false` | Enable ACK-specific features |
| `podAdditionalLabels` | `{}` | Extra labels applied to all pods |
| `podAdditionalAnnotations` | `{}` | Extra annotations (e.g. image cache hints) |
| `warnupImage` | `""` | Image used for warmup jobs |
| `warnupCount` | `0` | Number of warmup job pods |

---

## 9. Deployment

### Helm Chart

Chart path: `charts/claw-swarm-panel/`

```bash
helm install claw-swarm-operator ./charts/claw-swarm-panel \
  --namespace claw-system \
  --values my-values.yaml
```

#### Key Chart Resources

| Resource | Description |
|----------|-------------|
| `Deployment` (operator) | K8s operator process |
| `ConfigMap` (resource) | Rendered `config.yaml` mounted into operator pod |
| `ServiceAccount` + RBAC | Cluster-scoped permissions for StatefulSet/Service/Ingress/PVC management |
| `Service` | ClusterIP on port 8088 |
| `Ingress` | Routes `/claw`, `/auth`, `/users` to the service |
| `PVC` | Persistent storage for `data_dir` (SQLite, JWT secret) |

#### Data Persistence

The operator requires a PVC mounted at `/data` (or the configured `data_dir`) for:
- `claw.db` — SQLite database (users, tokens)
- `.jwt_secret` — JWT signing key

Without persistence, all users and tokens are lost on pod restart.

#### RBAC Requirements

The operator ServiceAccount requires the following K8s permissions:

- `StatefulSets`: get, list, watch, create, update, patch, delete
- `Services`, `Ingresses`: get, list, watch, create, update, patch, delete
- `PersistentVolumeClaims`: get, list, watch, delete
- `Pods`: get, list, watch, exec (for `configurePodModels`)
- `Jobs`, `Pods` (ACK only): create, delete

---

## 10. Integration Points

### LiteLLM

| Integration | Detail |
|-------------|--------|
| User provisioning | `POST /user/new` on allocation |
| Key retrieval | `POST /key/generate` for existing users |
| Auth | Master key via `Authorization: Bearer` header |
| Model injection | Via `openclaw config set` exec into pod |
| Failure handling | Random fallback key generated; allocation proceeds |

### Kubernetes

| Integration | Detail |
|-------------|--------|
| Client | `controller-runtime` + `client-go` |
| Watch | StatefulSets with `tabtabai.com/tabclaw` label |
| Pod exec | `remotecommand.SPDY` for model configuration |
| Resources | StatefulSets, Services, Ingresses, PVCs, Jobs |

### Redis (optional)

Used by `PVCReconciler` to track PVC-to-StatefulSet associations across reconcile cycles, enabling safe cleanup of orphaned PVCs after StatefulSet deletion.

### Kong Ingress

Each instance gets an Ingress rule routing `/<instance-name>/` (via Kong's path rewriting) to the ClusterIP service on port 18789.

---

## 11. Non-Functional Requirements

### Performance

| Requirement | Target |
|-------------|--------|
| Alloc API response time | < 500 ms (model config is async) |
| Pool reconciliation lag | < 30 s after instance deletion |
| Web UI initial load | < 3 s |

### Reliability

| Requirement | Detail |
|-------------|--------|
| Operator crash recovery | controller-runtime leader election; reconciler is idempotent |
| JWT validity across restarts | JWT secret persisted to `<data_dir>/.jwt_secret` |
| DB durability | SQLite write-ahead logging; PVC-backed in production |

### Security

| Requirement | Detail |
|-------------|--------|
| Password storage | bcrypt, cost 10 |
| JWT signing | HS256, 32-byte random secret |
| API secret entropy | 20 bytes (`crypto/rand`), hex-encoded with `claw_` prefix |
| File permissions | `data_dir` mode 0700, secret files mode 0600 |
| HTTP/2 | Disabled by default (CVE mitigation) |

### Observability

| Feature | Detail |
|---------|--------|
| Structured logging | `klog` with verbosity levels throughout operator |
| Health probes | `/healthz` (liveness), `/readyz` (readiness) on port 8081 |
| pprof profiling | Available on metrics server at `/debug/pprof/` when `--profiling=true` |
| Metrics | Prometheus-compatible via controller-runtime metrics server |

---

## 12. Future Scope

The following capabilities are identified as potential future enhancements but are **out of scope** for the current version:

| Feature | Description |
|---------|-------------|
| **Multi-cluster management** | UI and API to manage instances across multiple K8s clusters via stored kubeconfigs |
| **Usage metrics** | Per-instance CPU/memory utilisation from K8s metrics-server |
| **Billing & quotas** | Per-user instance limits, usage tracking, soft/hard quotas |
| **Audit log** | Persistent log of alloc/free/user management events |
| **HA SQLite** | Replace SQLite with PostgreSQL for high-availability deployments |
| **Webhook notifications** | Notify external systems on lifecycle events (alloc, free, pause) |
| **Instance templates** | Multiple resource profiles selectable at allocation time |
| **Auto-scale pool** | Dynamic `pool_size` based on allocation demand |
| **Skill pre-loading** | Init container to pre-populate workspace from a skills image |
