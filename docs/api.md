# Claw Swarm Operator — API Reference

The operator exposes an HTTP management API (default port **8088**).
All request and response bodies are **JSON** (`Content-Type: application/json`).

---

## Base URL

```
http://<operator-service>:<port>
```

---

## Authentication

All endpoints except `POST /auth/login` require authentication. Two methods are accepted:

### Bearer JWT (interactive / WebUI)

Obtain a token via `POST /auth/login`, then pass it on every request:

```
Authorization: Bearer <token>
```

Tokens expire after **24 hours**.

### API Secret (automation / scripts)

Each user has a permanent, non-expiring API secret. Pass it via the `X-API-Key` header:

```
X-API-Key: claw_<40-hex-chars>
```

View your secret via `GET /auth/api-secret` or the WebUI user dropdown → **API Secret**.
Regenerate it via `POST /auth/api-secret/regenerate` (old secret is invalidated immediately).

If both headers are present, the Bearer JWT is checked first.

### Unauthenticated requests

```json
HTTP 401  { "msg": "unauthorized" }
```

### Forbidden (non-admin on write endpoint)

```json
HTTP 403  { "msg": "forbidden: write operations require admin role" }
```

---

## Auth Endpoints

### POST `/auth/login` _(public)_

```json
{ "username": "admin", "password": "happyclaw" }
```

**Response `200 OK`**

```json
{
  "token": "eyJ...",
  "username": "admin",
  "role": "admin",
  "force_change_password": true
}
```

---

### GET `/auth/me` _(authenticated)_

Returns the current user record.

**Response `200 OK`**

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

---

### POST `/auth/change-password` _(authenticated)_

```json
{ "old_password": "happyclaw", "new_password": "mynewpass" }
```

**Response `200 OK`**

```json
{ "msg": "password updated", "token": "eyJ..." }
```

A fresh JWT is issued with `force_change_password` cleared.

---

### GET `/auth/api-secret` _(authenticated)_

Returns the caller's current API secret.

**Response `200 OK`**

```json
{ "api_secret": "claw_a3f2e1d4b5c6..." }
```

---

### POST `/auth/api-secret/regenerate` _(authenticated)_

Generates a new API secret. The previous value is **immediately invalidated**.

**Response `200 OK`**

```json
{ "api_secret": "claw_<new-40-hex-chars>" }
```

---

## User Management Endpoints _(admin only)_

### GET `/users`

List all users.

**Response `200 OK`**

```json
{
  "data": [
    {
      "id": 1,
      "username": "admin",
      "role": "admin",
      "force_change_password": false,
      "api_secret": "claw_...",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

### POST `/users`

Create a new user.

**Request**

```json
{
  "username": "alice",
  "password": "secret123",
  "role": "user"       // "admin" or "user" (default: "user")
}
```

**Response `200 OK`** — the created user object.

---

### DELETE `/users/:id`

Delete a user by numeric ID. Cannot delete yourself.

**Response `200 OK`**

```json
{ "msg": "deleted" }
```

---

## Claw Instance Object

All claw endpoints that return instance data use this common structure:

```json
{
  "name":           "string",
  "user_id":        "string",
  "claw_webui_url": "string",
  "claw_wss_url":   "string",
  "occupied":       false,
  "state":          "string",
  "alloc_status":   "string",
  "token":          "string",
  "resources": {
    "cpu_request":    "string",
    "cpu_limit":      "string",
    "memory_request": "string",
    "memory_limit":   "string"
  }
}
```

### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | StatefulSet name, unique per instance |
| `user_id` | string | User assigned to this instance; empty string if free |
| `claw_webui_url` | string | HTTPS URL to the OpenClaw gateway for this instance |
| `claw_wss_url` | string | WebSocket URL for the instance (set by the operator via annotation) |
| `occupied` | bool | `true` if `user_id` is non-empty (i.e. the instance is allocated) |
| `state` | string | Runtime state derived from the StatefulSet — see table below |
| `alloc_status` | string | Allocation phase label on the StatefulSet — see table below |
| `token` | string | Gateway auth token from `claw_tokens` table; empty if not provisioned |
| `resources` | object | CPU / memory request and limit strings (e.g. `"250m"`, `"2Gi"`) |

---

### `state` — Runtime State

Derived from the StatefulSet status and annotations at query time.

| Value | Condition | Meaning |
|-------|-----------|---------|
| `"pending"` | `replicas > 0` AND `availableReplicas < replicas` | Pod is scheduled but not yet ready (starting up or crashing) |
| `"running"` | `replicas > 0` AND `availableReplicas == replicas` | All replicas running and ready |
| `"paused"` | `replicas == 0` | Scaled to zero; instance is suspended |
| `"paused"` | Annotation `tabtab.app.scheduled.deletion.time` is a valid RFC3339 time | A pause has been scheduled (timer is counting down); pod may still be running |

> **Note:** `"paused"` covers two distinct sub-states — actively scaled-down (`replicas == 0`) and pause-pending (timer set, still running). Callers who need to distinguish them can inspect `occupied` and `user_id`: a paused but still-occupied instance is awaiting scale-down.

---

### `alloc_status` — Allocation Phase

Set as a Kubernetes label (`tabtabai.com/tabclaw-alloc-status`) on the StatefulSet.

| Value | Label present | Meaning |
|-------|--------------|---------|
| `""` (empty string) | Label absent | Instance is free; not assigned to any user |
| `"allocating"` | `tabclaw-alloc-status=allocating` | Allocation has started; the instance is claimed and replicas are set to 1, but model configuration (`configurePodModels`) is still running asynchronously in the background |
| `"allocated"` | `tabclaw-alloc-status=allocated` | Allocation is fully complete; model configuration finished successfully |

> **Polling guidance:** After calling `POST /claw/alloc`, the response returns immediately with `alloc_status: "allocating"`. Poll `GET /claw/instances` or `GET /claw/used` until `alloc_status` becomes `"allocated"` before sending model requests to the instance.

---

## Claw Endpoints

### GET `/claw/instances` _(authenticated)_

List all OpenClaw instances in the pool (both free and allocated). Optionally filter by allocation state.

**Query parameters**

| Param | Values | Description |
|-------|--------|-------------|
| `occupied` | `true` | Return only allocated instances (`occupied == true`) |
| `occupied` | `false` | Return only free instances (`occupied == false`) |
| _(omit)_ | — | Return all instances |

**Response `200 OK`**

```json
{
  "data": [
    {
      "name":           "claw-abc123",
      "user_id":        "",
      "claw_webui_url": "https://claw.example.com/claw-abc123/overview",
      "occupied":       false,
      "state":          "running",
      "alloc_status":   "",
      "token":          "",
      "resources": {
        "cpu_request": "250m", "cpu_limit": "1",
        "memory_request": "512Mi", "memory_limit": "2Gi"
      },
      "created_at": "2026-03-19T10:00:00Z"
    }
  ]
}
```

---

### GET `/claw/instances/{name}` _(authenticated)_

Retrieve a single instance by its StatefulSet name. Also enriches the response with the gateway token from the `claw_tokens` table if one has been provisioned.

**Path parameters**

| Param | Description |
|-------|-------------|
| `name` | StatefulSet name of the instance (required) |

**Response `200 OK`**

```json
{
  "data": {
    "name":           "claw-abc123",
    "user_id":        "user-xyz",
    "claw_webui_url": "https://claw.example.com/claw-abc123/overview",
    "claw_wss_url":   "wss://claw.example.com/claw-abc123/ws",
    "occupied":       true,
    "state":          "running",
    "alloc_status":   "allocated",
    "token":          "my-gateway-token",
    "resources": {
      "cpu_request": "250m", "cpu_limit": "1",
      "memory_request": "512Mi", "memory_limit": "2Gi"
    }
  }
}
```

**Error responses**

| Status | Body `msg` | Reason |
|--------|-----------|--------|
| `400` | `instance name is required` | Missing path parameter |
| `404` | `instance not found` | No claw instance with this name |

---

### POST `/claw/instances/{name}/exec` _(admin only)_

Execute a command inside the primary container of a running claw instance. Uses the Kubernetes exec API (`kubectl exec` equivalent) targeting pod `<name>-0`.

The command runs synchronously; stdout and stderr are captured and returned in the response. The instance must be in `running` state (replicas > 0).

**Path parameters**

| Param | Description |
|-------|-------------|
| `name` | StatefulSet name of the instance (required) |

**Request**

```json
{
  "command":   ["bash", "-c", "ls -la /workspace"],  // required — command + args
  "container": "tabclaw"                              // optional — defaults to primary container
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `command` | `string[]` | yes | Command and its arguments as an array |
| `container` | string | no | Container name to exec into; defaults to `tabclaw` |

**Response `200 OK`**

```json
{
  "stdout":    "total 8\ndrwxr-xr-x 2 root root 4096 Jan 1 00:00 .\n",
  "stderr":    "",
  "exit_code": 0
}
```

| Field | Type | Description |
|-------|------|-------------|
| `stdout` | string | Captured standard output of the command |
| `stderr` | string | Captured standard error of the command |
| `exit_code` | int | Process exit code; `0` = success |

> **Note:** A non-zero `exit_code` is returned with HTTP `200`. The HTTP status reflects whether the exec request itself succeeded, not whether the command inside the container succeeded.

**Error responses**

| Status | Body `msg` | Reason |
|--------|-----------|--------|
| `400` | `instance name is required` | Missing path parameter |
| `400` | `command is required` | Empty `command` array |
| `404` | `instance not found` | No claw instance with this name |
| `500` | _(error detail)_ | Exec failed (pod not running, container not found, etc.) |

---

### GET `/claw/used` _(authenticated)_

List allocated (occupied) instances only. Equivalent to `GET /claw/instances?occupied=true` but supports filtering by specific user.

**Query parameters**

| Param | Description |
|-------|-------------|
| `user_id` | Return only the instance assigned to this specific user (optional) |

**Response `200 OK`**

```json
{
  "data": [
    {
      "name":           "claw-abc123",
      "user_id":        "user-xyz",
      "claw_webui_url": "https://claw.example.com/claw-abc123/overview",
      "occupied":       true,
      "state":          "running",
      "alloc_status":   "allocated",
      "token":          "my-gateway-token",
      "resources": {
        "cpu_request": "250m", "cpu_limit": "1",
        "memory_request": "512Mi", "memory_limit": "2Gi"
      },
      "created_at": "2026-03-19T10:00:00Z"
    }
  ]
}
```

---

### GET `/claw/token` _(admin only)_

Retrieve the gateway auth token for a named instance from the `claw_tokens` table. The token is **read-only** — calling this endpoint does **not** consume or remove the token.

**Query parameters**

| Param | Description |
|-------|-------------|
| `name` | StatefulSet name of the instance (required) |

**Response `200 OK`**

```json
{ "token": "my-gateway-token" }
```

Returns `{ "token": "" }` if no token has been provisioned for this instance.

**Error responses**

| Status | Body `msg` | Reason |
|--------|-----------|--------|
| `400` | `name is required` | Missing `name` query parameter |

---

### POST `/claw/alloc` _(admin only)_

Allocate an idle OpenClaw instance to a user. If the user already has an instance, the existing one is returned (idempotent). Also creates a LiteLLM user and generates an API key if LiteLLM is configured.

**Request**

```json
{
  "user_id":    "string",   // required — unique identifier for the user
  "model_type": "string"    // optional — "lite" or "pro", default "lite"
}
```

**Response `200 OK`**

```json
{
  "name":           "string",   // StatefulSet name assigned to this user
  "user_id":        "string",   // user_id this instance is assigned to
  "claw_webui_url": "string",   // HTTPS URL to the OpenClaw gateway
  "occupied":       true,
  "state":          "pending",  // instance starts as pending; becomes "running" once ready
  "alloc_status":   "allocating", // becomes "allocated" once model configuration completes
  "token":          "string"    // gateway auth token (from claw_tokens table, if present)
}
```

> **Note:** Model configuration (`configurePodModels`) runs **asynchronously** after the response is returned. The `alloc_status` label on the StatefulSet transitions from `allocating` → `allocated` once it completes. Poll `GET /claw/instances` to observe the transition.

**Error responses**

| Status | Body `msg` | Reason |
|--------|-----------|--------|
| `400` | `incorrect request body format` | Missing or wrong `Content-Type` |
| `400` | `incorrect request body` | Malformed JSON |
| `500` | `unable to find available service at the moment, try again later` | Pool exhausted |

---

### POST `/claw/free` _(admin only)_

Permanently delete an OpenClaw instance. All associated Kubernetes resources are deleted.

**Request** — identify by `user_id` **or** `name` (user_id takes precedence when both are given):

```json
{
  "user_id": "string",   // user assigned to the instance (optional if name given)
  "name":    "string"    // StatefulSet name (optional if user_id given)
}
```

**Response `200 OK`**

```json
{ "msg": "success" }
```

**Error responses**

| Status | Body `msg` | Reason |
|--------|-----------|--------|
| `404` | `not found claw instance` | No instance found for the given user/name |

---

### POST `/claw/pause` _(admin only)_

Schedule an instance for suspension. A deletion timestamp is set; the operator sets replicas to 0 after `delay_pause_minutes`.

**Request**

```json
{
  "user_id": "string",   // optional if name given
  "name":    "string"    // optional if user_id given
}
```

**Response `200 OK`**

```json
{ "msg": "success" }
```

**Error responses**

| Status | Body `msg` | Reason |
|--------|-----------|--------|
| `404` | `not found claw instance` | No instance found |

---

### POST `/claw/resume` _(admin only)_

Resume a paused instance (sets replicas back to 1) and clears any pending scheduled pause.

**Request**

```json
{
  "user_id": "string",   // optional if name given
  "name":    "string"    // optional if user_id given
}
```

**Response `200 OK`**

```json
{ "msg": "success" }
```

**Error responses**

| Status | Body `msg` | Reason |
|--------|-----------|--------|
| `404` | `not found claw instance` | No instance found |

---

## ACK-only Endpoints _(admin only)_

The following endpoints are only available when `claw.ack.enabled: true`.

---

### POST `/claw/warnup-job`

Create a Kubernetes Job that pre-warms the OpenClaw image cache on ACK nodes.

**Request**

```json
{
  "count": 1   // number of parallel warm-up pods
}
```

**Response `200 OK`**

```json
{
  "jobName":      "string",
  "jobNamespace": "string"
}
```

---

### POST `/claw/remove-warnup-job`

Delete a previously created warm-up Job and its pods.

**Request**

```json
{
  "jobName": "string"
}
```

**Response `200 OK`** — plain text `success`.

---

### POST `/claw/pv`

Retrieve the PersistentVolume name backing an instance's workspace.

**Request**

```json
{
  "user_id": "string"
}
```

**Response `200 OK`**

```json
{
  "pv": "string"   // PersistentVolume name
}
```

**Error responses**

| Status | Body `msg` | Reason |
|--------|-----------|--------|
| `500` | `can't find claw` | No instance found for this user |
| `500` | `can't find claw pvc` | PVC not yet created |

---

## Instance Lifecycle

```
             ┌──────────────────────────────┐
             │          IDLE POOL            │
             │  (pre-created, no user_id)    │
             └──────────────┬───────────────┘
                            │ POST /claw/alloc
                            ▼
             ┌──────────────────────────────┐
             │   ALLOCATING (async config)   │
             │  alloc_status = "allocating"  │
             └──────────────┬───────────────┘
                            │ configurePodModels completes
                            ▼
             ┌──────────────────────────────┐
             │          ALLOCATED            │
             │  alloc_status = "allocated"   │
             │  replicas = 1, running        │
             └────────┬─────────────────────┘
                      │ POST /claw/pause
                      ▼
        ┌─────────────────────────┐
        │   PAUSING (delayed)     │
        │  waiting delay_pause_   │
        │  minutes before scale-0 │
        └────────────┬────────────┘
                     │ timer expires
                     ▼
        ┌─────────────────────────┐
        │         PAUSED          │
        │       replicas = 0      │
        └────────────┬────────────┘
                     │ POST /claw/resume
                     ▼
        ┌─────────────────────────┐
        │         RUNNING         │
        └────────────┬────────────┘
                     │ POST /claw/free
                     ▼
                 (deleted)
```

---

## Error Response Format

All errors return JSON with a single `msg` field:

```json
{ "msg": "human-readable error description" }
```
