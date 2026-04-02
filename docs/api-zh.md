# Claw Swarm Operator — API 参考

Operator 暴露 HTTP 管理 API（默认端口 **8088**）。
所有请求和响应体均为 **JSON**（`Content-Type: application/json`）。

---

## Base URL

```
http://<operator-service>:<port>
```

---

## 认证

除 `POST /auth/login` 外，所有端点均需认证。支持两种方式：

### Bearer JWT（交互式 / WebUI）

通过 `POST /auth/login` 获取 Token，每次请求携带：

```
Authorization: Bearer <token>
```

Token 有效期 **24 小时**。

### API Secret（自动化 / 脚本）

每个用户拥有永久不过期的 API Secret，通过 `X-API-Key` Header 传递：

```
X-API-Key: claw_<40位十六进制字符>
```

通过 `GET /auth/api-secret` 或 WebUI 用户下拉菜单 → **API Secret** 查看。
通过 `POST /auth/api-secret/regenerate` 重新生成（旧 Secret 立即失效）。

若两个 Header 同时存在，优先验证 Bearer JWT。

### 未认证请求

```json
HTTP 401  { "msg": "unauthorized" }
```

### 无权限（非管理员访问写操作端点）

```json
HTTP 403  { "msg": "forbidden: write operations require admin role" }
```

---

## 认证端点

### POST `/auth/login` _（公开）_

```json
{ "username": "admin", "password": "happyclaw" }
```

**响应 `200 OK`**

```json
{
  "token": "eyJ...",
  "username": "admin",
  "role": "admin",
  "force_change_password": true
}
```

---

### GET `/auth/me` _（已认证）_

返回当前用户记录。

**响应 `200 OK`**

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

### POST `/auth/change-password` _（已认证）_

```json
{ "old_password": "happyclaw", "new_password": "mynewpass" }
```

**响应 `200 OK`**

```json
{ "msg": "password updated", "token": "eyJ..." }
```

返回新 JWT，`force_change_password` 标记已清除。

---

### GET `/auth/api-secret` _（已认证）_

返回当前用户的 API Secret。

**响应 `200 OK`**

```json
{ "api_secret": "claw_a3f2e1d4b5c6..." }
```

---

### POST `/auth/api-secret/regenerate` _（已认证）_

生成新 API Secret，旧值**立即失效**。

**响应 `200 OK`**

```json
{ "api_secret": "claw_<新40位十六进制字符>" }
```

---

## 用户管理端点 _（仅限管理员）_

### GET `/users`

列出所有用户。

**响应 `200 OK`**

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

创建新用户。

**请求**

```json
{
  "username": "alice",
  "password": "secret123",
  "role": "user"       // "admin" 或 "user"（默认："user"）
}
```

**响应 `200 OK`** — 返回创建的用户对象。

---

### DELETE `/users/:id`

按数字 ID 删除用户，不能删除自己。

**响应 `200 OK`**

```json
{ "msg": "deleted" }
```

---

## Claw 实例对象

所有返回实例数据的端点均使用此通用结构：

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

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | StatefulSet 名称，每个实例唯一 |
| `user_id` | string | 分配给此实例的用户；未分配时为空字符串 |
| `claw_webui_url` | string | 实例 OpenClaw 网关的 HTTPS URL |
| `claw_wss_url` | string | 实例的 WebSocket URL（由 Operator 通过 Annotation 设置） |
| `occupied` | bool | `user_id` 非空时为 `true`（即实例已被分配） |
| `state` | string | 由 StatefulSet 推导的运行状态，见下表 |
| `alloc_status` | string | StatefulSet 上的分配阶段标签，见下表 |
| `token` | string | 来自 `claw_tokens` 表的网关认证 Token；未配置时为空 |
| `resources` | object | CPU / 内存 Request 和 Limit（如 `"250m"`、`"2Gi"`） |

---

### `state` — 运行状态

| 值 | 条件 | 含义 |
|----|------|------|
| `"pending"` | `replicas > 0` 且 `availableReplicas < replicas` | Pod 已调度但未就绪（启动中或崩溃） |
| `"running"` | `replicas > 0` 且 `availableReplicas == replicas` | 所有副本运行并就绪 |
| `"paused"` | `replicas == 0` | 已缩容至零；实例已暂停 |
| `"paused"` | Annotation `tabtab.app.scheduled.deletion.time` 为有效 RFC3339 时间 | 已设置暂停计时器（Pod 可能仍在运行） |

---

### `alloc_status` — 分配阶段

设置为 StatefulSet 上的 Kubernetes 标签（`tabtabai.com/tabclaw-alloc-status`）。

| 值 | 标签是否存在 | 含义 |
|----|------------|------|
| `""` （空字符串） | 标签缺失 | 实例空闲，未分配给任何用户 |
| `"allocating"` | `tabclaw-alloc-status=allocating` | 分配已启动；实例已认领，副本数已置为 1，但模型配置（`configurePodModels`）仍在后台异步运行 |
| `"allocated"` | `tabclaw-alloc-status=allocated` | 分配完全完成；模型配置已成功结束 |

> **轮询建议：** 调用 `POST /claw/alloc` 后，响应立即返回，`alloc_status` 为 `"allocating"`。请轮询 `GET /claw/instances` 或 `GET /claw/used`，直到 `alloc_status` 变为 `"allocated"` 后，再向实例发送模型请求。

---

## Claw 端点

### GET `/claw/instances` _（已认证）_

列出实例池中的所有实例（空闲和已分配）。可按分配状态过滤。

**查询参数**

| 参数 | 值 | 说明 |
|------|-----|------|
| `occupied` | `true` | 仅返回已分配实例（`occupied == true`） |
| `occupied` | `false` | 仅返回空闲实例（`occupied == false`） |
| _（省略）_ | — | 返回所有实例 |

**响应 `200 OK`**

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

### GET `/claw/instances/{name}` _（已认证）_

按 StatefulSet 名称获取单个实例。若 `claw_tokens` 表中存在对应 Token，也一并返回。

**路径参数**

| 参数 | 说明 |
|------|------|
| `name` | 实例的 StatefulSet 名称（必填） |

**响应 `200 OK`**

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

**错误响应**

| 状态码 | `msg` | 原因 |
|--------|-------|------|
| `400` | `instance name is required` | 缺少路径参数 |
| `404` | `instance not found` | 找不到对应实例 |

---

### POST `/claw/instances/{name}/exec` _（仅限管理员）_

在运行中的 Claw 实例的主容器内执行命令（等同于 `kubectl exec`，目标 Pod 为 `<name>-0`）。

命令同步执行，stdout 和 stderr 被捕获并在响应中返回。实例必须处于 `running` 状态。

**路径参数**

| 参数 | 说明 |
|------|------|
| `name` | 实例的 StatefulSet 名称（必填） |

**请求**

```json
{
  "command":   ["bash", "-c", "ls -la /workspace"],  // 必填 — 命令及参数
  "container": "tabclaw"                              // 可选 — 默认为主容器
}
```

**响应 `200 OK`**

```json
{
  "stdout":    "total 8\ndrwxr-xr-x 2 root root 4096 Jan 1 00:00 .\n",
  "stderr":    "",
  "exit_code": 0
}
```

> **注意：** 非零 `exit_code` 仍以 HTTP `200` 返回。HTTP 状态码反映的是 exec 请求本身是否成功，而非容器内命令是否成功。

**错误响应**

| 状态码 | `msg` | 原因 |
|--------|-------|------|
| `400` | `instance name is required` | 缺少路径参数 |
| `400` | `command is required` | `command` 数组为空 |
| `404` | `instance not found` | 找不到对应实例 |
| `500` | _（错误详情）_ | Exec 失败（Pod 未运行、容器未找到等） |

---

### GET `/claw/used` _（已认证）_

仅列出已占用（已分配）的实例。等同于 `GET /claw/instances?occupied=true`，但支持按用户过滤。

**查询参数**

| 参数 | 说明 |
|------|------|
| `user_id` | 仅返回分配给此用户的实例（可选） |

**响应 `200 OK`**

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

### GET `/claw/token` _（仅限管理员）_

从 `claw_tokens` 表中获取指定实例的网关认证 Token。此端点**只读**，不消耗或删除 Token。

**查询参数**

| 参数 | 说明 |
|------|------|
| `name` | 实例的 StatefulSet 名称（必填） |

**响应 `200 OK`**

```json
{ "token": "my-gateway-token" }
```

若此实例尚未配置 Token，返回 `{ "token": "" }`。

**错误响应**

| 状态码 | `msg` | 原因 |
|--------|-------|------|
| `400` | `name is required` | 缺少 `name` 查询参数 |

---

### POST `/claw/alloc` _（仅限管理员）_

将空闲 OpenClaw 实例分配给用户。若用户已有实例则直接返回（幂等）。若配置了 LiteLLM，还会创建 LiteLLM 用户并生成 API Key。

**请求**

```json
{
  "user_id":    "string",   // 必填 — 用户唯一标识
  "model_type": "string"    // 可选 — "lite" 或 "pro"，默认 "lite"
}
```

**响应 `200 OK`**

```json
{
  "name":           "string",   // 分配给此用户的 StatefulSet 名称
  "user_id":        "string",   // 该实例分配到的 user_id
  "claw_webui_url": "string",   // OpenClaw 网关的 HTTPS URL
  "occupied":       true,
  "state":          "pending",  // 实例初始为 pending，就绪后变为 "running"
  "alloc_status":   "allocating", // 模型配置完成后变为 "allocated"
  "token":          "string"    // 网关认证 Token（若已配置，来自 claw_tokens 表）
}
```

> **注意：** 模型配置（`configurePodModels`）在响应返回后**异步**执行。StatefulSet 上的 `alloc_status` 标签会从 `allocating` 变为 `allocated`。请轮询 `GET /claw/instances` 观察该变化。

**错误响应**

| 状态码 | `msg` | 原因 |
|--------|-------|------|
| `400` | `incorrect request body format` | 缺少或错误的 `Content-Type` |
| `400` | `incorrect request body` | JSON 格式错误 |
| `500` | `unable to find available service at the moment, try again later` | 实例池已耗尽 |

---

### POST `/claw/free` _（仅限管理员）_

永久删除 OpenClaw 实例，关联的 Kubernetes 资源将级联删除。

**请求** — 通过 `user_id` **或** `name` 指定（同时传入时 `user_id` 优先）：

```json
{
  "user_id": "string",   // 分配给实例的用户（与 name 二选一）
  "name":    "string"    // StatefulSet 名称（与 user_id 二选一）
}
```

**响应 `200 OK`**

```json
{ "msg": "success" }
```

**错误响应**

| 状态码 | `msg` | 原因 |
|--------|-------|------|
| `404` | `not found claw instance` | 找不到对应实例 |

---

### POST `/claw/pause` _（仅限管理员）_

将实例挂起。设置删除时间戳，Operator 在 `delay_pause_minutes` 分钟后将副本数置为 0。

**请求**

```json
{
  "user_id": "string",   // 与 name 二选一
  "name":    "string"    // 与 user_id 二选一
}
```

**响应 `200 OK`**

```json
{ "msg": "success" }
```

**错误响应**

| 状态码 | `msg` | 原因 |
|--------|-------|------|
| `404` | `not found claw instance` | 找不到对应实例 |

---

### POST `/claw/resume` _（仅限管理员）_

恢复已暂停的实例（副本数恢复为 1），并清除所有待执行的暂停计划。

**请求**

```json
{
  "user_id": "string",   // 与 name 二选一
  "name":    "string"    // 与 user_id 二选一
}
```

**响应 `200 OK`**

```json
{ "msg": "success" }
```

**错误响应**

| 状态码 | `msg` | 原因 |
|--------|-------|------|
| `404` | `not found claw instance` | 找不到对应实例 |

---

## 仅限 ACK 的端点 _（仅限管理员）_

以下端点仅在 `claw.ack.enabled: true` 时可用。

---

### POST `/claw/warnup-job`

创建一个 Kubernetes Job，在 ACK 节点上预热 OpenClaw 镜像缓存。

**请求**

```json
{
  "count": 1   // 并行预热 Pod 数量
}
```

**响应 `200 OK`**

```json
{
  "jobName":      "string",
  "jobNamespace": "string"
}
```

---

### POST `/claw/remove-warnup-job`

删除之前创建的预热 Job 及其 Pod。

**请求**

```json
{
  "jobName": "string"
}
```

**响应 `200 OK`** — 纯文本 `success`。

---

### POST `/claw/pv`

获取实例工作区对应的 PersistentVolume 名称。

**请求**

```json
{
  "user_id": "string"
}
```

**响应 `200 OK`**

```json
{
  "pv": "string"   // PersistentVolume 名称
}
```

**错误响应**

| 状态码 | `msg` | 原因 |
|--------|-------|------|
| `500` | `can't find claw` | 找不到该用户的实例 |
| `500` | `can't find claw pvc` | PVC 尚未创建 |

---

## 实例生命周期

```
             ┌──────────────────────────────┐
             │          空闲实例池           │
             │  (预创建，无 user_id)         │
             └──────────────┬───────────────┘
                            │ POST /claw/alloc
                            ▼
             ┌──────────────────────────────┐
             │   分配中（异步配置）          │
             │  alloc_status = "allocating" │
             └──────────────┬───────────────┘
                            │ configurePodModels 完成
                            ▼
             ┌──────────────────────────────┐
             │          已分配              │
             │  alloc_status = "allocated"  │
             │  replicas = 1，运行中        │
             └────────┬─────────────────────┘
                      │ POST /claw/pause
                      ▼
        ┌─────────────────────────┐
        │   暂停中（延迟）        │
        │  等待 delay_pause_      │
        │  minutes 后缩容至 0     │
        └────────────┬────────────┘
                     │ 计时器到期
                     ▼
        ┌─────────────────────────┐
        │         已暂停          │
        │       replicas = 0      │
        └────────────┬────────────┘
                     │ POST /claw/resume
                     ▼
        ┌─────────────────────────┐
        │         运行中          │
        └────────────┬────────────┘
                     │ POST /claw/free
                     ▼
                 （已删除）
```

---

## 错误响应格式

所有错误均返回包含单个 `msg` 字段的 JSON：

```json
{ "msg": "人类可读的错误描述" }
```
