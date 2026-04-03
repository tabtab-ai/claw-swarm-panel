**[English](README.md) | [中文](README_zh.md)**

---

# 🦞 Claw Swarm Panel

**[OpenClaw](https://github.com/openclaw/openclaw)🦞 Web 管理面板 —— HTTP API 服务 + React 控制台，专为中小企业打造。**

让每位员工拥有独立隔离的 AI 编程工作空间，500 毫秒内完成分配，自动回收闲置资源，零运维负担。

> **项目拆分说明：** 本仓库只包含**管理面板**（API Server + Web UI）。
> Kubernetes 适配器与实例生命周期控制器已独立为单独项目：
> **[claw-swarm-operator](https://gitlab.botnow.cn/agentic/claw-swarm-operator)** —— 请先将其部署到集群中。

---

## 为什么选择 Claw Swarm Panel？

中小企业引入 OpenClaw 面临的最大挑战不是技术能力，而是**安全性与运营成本**：如何保证员工数据隔离、如何防止 API 密钥泄露、如何在不增加运维负担的前提下稳定运行。

### 手动管理的痛点

| 问题 | 手动方式的代价 |
|------|--------------|
| 开通速度慢 | 每人 5–15 分钟，入职当天无法使用 |
| 数据隔离难 | 多人共用环境，代码与配置相互可见 |
| API 密钥风险 | 共享密钥一旦泄露，全员受影响 |
| 资源浪费 | 容器全天运行，夜间和周末白白消耗算力 |
| 运维门槛高 | 需要专职工程师维护，难以委托给 IT |
| 模型配置繁琐 | 每个容器独立配置，批量变更无从下手 |

### Claw Swarm Panel 的解法

| 核心价值 | 实现方式 |
|---------|---------|
| **员工数据完全隔离** | 每人独占一个工作空间容器，数据永不交叉 |
| **API 密钥按人管控** | 对接 LiteLLM，自动为每位员工生成独立密钥，离职即回收 |
| **500ms 内完成分配** | 预热实例池常备就绪，分配 API 响应 < 500ms，无冷启动等待 |
| **自动暂停与恢复** | 闲置实例自动缩容至零，工作日早晨一键恢复，数据不丢失 |
| **IT 可自主运营** | Web 管理面板覆盖全部日常操作，无需掌握底层技术 |
| **系统集成友好** | REST API + API Secret，可对接 HR 系统、钉钉/飞书机器人、SSO 平台 |

---

## 架构

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
│                          │  │  k8s adapter（内置）      │  │   │
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
                    │   ClawReconciler — 实例池管理          │
                    │   PVCReconciler  — 存储 GC             │
                    └───────────────────┬──────────────────┘
                                        │
          ┌──────────┬──────────────────┼──────────────────┐
          ▼          ▼                  ▼                   ▼
      员工 A 专属  员工 B 专属       员工 C 专属          LiteLLM
       工作空间    工作空间           工作空间        （AI 模型路由
        + 存储      + 存储             + 存储           & 密钥管理）
```

| 组件 | 仓库 | 职责 |
|------|------|------|
| **claw-swarm-panel** | 本仓库 | HTTP API + Web UI —— 实例操作、用户管理 |
| **claw-swarm-operator** | 独立仓库 | K8s 控制器 —— 维护实例池、管理生命周期 |

---

## 功能列表

- **完整生命周期** —— alloc / free / pause（支持延迟）/ resume
- **员工数据隔离** —— 每人独占容器 + 持久存储，数据永不共享
- **LiteLLM 集成** —— 分配时自动创建用户 + API 密钥，异步注入模型配置；LiteLLM 不可用时优雅降级
- **双重认证** —— JWT（Web 交互会话）+ API Secret（系统集成）
- **角色权限** —— `admin`（完全控制）和 `user`（只读）
- **审计日志** —— 每次分配/释放/暂停/恢复/用户管理操作均有记录
- **网关令牌管理** —— 每个实例的访问令牌存储于数据库，分配时一并返回

---

## 快速开始

### 前提条件

先将 `claw-swarm-operator` 部署到 Kubernetes 集群 —— 它负责管理本面板所操作的实例池。

### 本地运行

```bash
make run          # 启动 API 服务（端口 8088）
make run-webui    # 启动 Web UI 开发服务器（端口 5173）
```

### 构建

```bash
make build        # 输出：bin/apiserver
make build-webui  # 输出：webui/dist/
```

---

## 配置

所有配置从 `config.yaml` 读取（通过 `--config` 指定路径）。

```yaml
# ── API Server ─────────────────────────────────────────────────────────────────
server:
  port: 8088
  data_dir: ".claw-swarm"        # SQLite 数据库 + JWT 密钥目录

# ── Adapter 后端 ───────────────────────────────────────────────────────────────
adapter:
  type: "k8s"
  k8s:
    kubeconfig: ""               # 显式路径；空 = ~/.kube/config 或 in-cluster
    namespace: ""                # 回退至 POD_NAMESPACE 环境变量

# ── LiteLLM 集成 ───────────────────────────────────────────────────────────────
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

## API 参考

除 `/auth/login` 外，所有接口均需 `Authorization: Bearer <jwt>` 或 `X-API-Key: claw_...`。

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/auth/login` | 登录，返回 JWT（有效期 24h） |
| `GET`  | `/auth/me` | 当前用户信息 |
| `POST` | `/auth/change-password` | 修改密码 |
| `GET`  | `/auth/api-secret` | 查看 API Secret |
| `POST` | `/auth/api-secret/regenerate` | 重新生成 API Secret |

### 实例生命周期（写操作仅限 admin）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET`  | `/claw/instances` | 列出所有实例（`?occupied=true\|false`） |
| `GET`  | `/claw/used` | 列出已分配实例 |
| `GET`  | `/claw/token` | 获取网关令牌（`?name=<name>`） |
| `POST` | `/claw/alloc` | 为用户分配实例 |
| `POST` | `/claw/free` | 释放（删除）实例 |
| `POST` | `/claw/pause` | 暂停实例（支持 `delay_minutes`） |
| `POST` | `/claw/resume` | 恢复已暂停实例 |

**分配：**
```json
{ "user_id": "alice", "model_type": "lite" }
```

**释放 / 恢复：**
```json
{ "name": "claw-a1b2c3d4" }
```

**延迟暂停：**
```json
{ "name": "claw-a1b2c3d4", "delay_minutes": 10 }
```

**实例响应示例：**
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

**实例状态：** `running` | `pending` | `paused` | `deleting`

**分配状态：** `allocating`（模型配置进行中）→ `allocated`（可正常使用）

### 用户管理（仅限 admin）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET`    | `/users`     | 列出所有面板用户 |
| `POST`   | `/users`     | 创建用户（用户名、密码、角色） |
| `DELETE` | `/users/:id` | 删除用户 |

---

## 部署（Helm）

**从 OCI 仓库安装（推荐）：**

> 以下参数需根据实际环境修改：
>
> | 参数 | 说明 |
> |------|------|
> | `ingress.hosts[0].host` | 面板 Web UI 和 API 对外暴露的域名 |
> | `ingress.className` | 集群 Ingress 控制器对应的 class 名称（如 `kong`、`nginx`） |
> | `claw.init.litellm.baseurl` | LiteLLM 服务地址，用于 AI 模型路由 |
> | `claw.init.litellm.master_key` | LiteLLM master key |

```bash
helm install claw-swarm-panel oci://registry-1.docker.io/tabtabai/claw-swarm-panel-chart \
  --namespace tabclaw --create-namespace \
  --set ingress.hosts[0].host=claw-panel.example.com \
  --set ingress.className=kong \
  --set claw.init.litellm.baseurl=http://litellm.example.com \
  --set claw.init.litellm.master_key=sk-...
```

**从本地源码安装：**

```bash
helm install claw-swarm-panel charts/claw-swarm-panel \
  --namespace tabclaw --create-namespace \
  -f values.yaml
```

关键 `values.yaml` 配置项：

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
make helm-lint     # 校验
make helm-package  # 打包为 .tgz
make helm-push     # 推送到 OCI 镜像仓库
```

---

## 系统集成

通过 `X-API-Key` 认证将 Claw Swarm Panel 对接外部系统：

| 系统 | 触发时机 | API 调用 |
|------|---------|---------|
| HR / 入职系统 | 员工入职 | `POST /claw/alloc` |
| 离职流程 | 员工离职 | `POST /claw/free` |
| 钉钉 / 飞书机器人 | 用户首次发起对话 | `POST /claw/alloc`（幂等） |
| 监控系统 | 定期健康检查 | `GET /claw/instances` |

---

## Make 命令

```
开发
  fmt / vet       格式化与静态检查
  lint            运行 golangci-lint
  lint-fix        运行 golangci-lint（自动修复）

前端
  install-webui   安装 webui 依赖（pnpm）
  run-webui       启动 Vite 开发服务器（:5173）
  build-webui     构建生产版本 webui

构建
  build           构建 apiserver 二进制（bin/apiserver）
  run             本地运行 apiserver

Docker
  docker-build    构建容器镜像（IMG）
  docker-push     推送容器镜像（IMG）

Helm
  helm-lint       校验 helm chart
  helm-package    打包 helm chart
  helm-push       推送 chart 到 OCI 仓库
```

---

## 默认凭据

| 用户名 | 密码 | 说明 |
|--------|------|------|
| `admin` | `happyclaw` | 首次登录必须修改密码 |

---

## 常见问题

**`unable to load in-cluster configuration`**
在 `config.yaml` 中设置 `adapter.k8s.kubeconfig`，或确保 `~/.kube/config` 指向可访问的集群。

**`address already in use` on port 8088**
修改 `config.yaml` 中的 `server.port`，并同步更新 `webui/vite.config.ts` 中的代理目标。

**启动后实例列表为空**
实例池由 `claw-swarm-operator` 管理，请确认其已在集群中正常运行：
```bash
kubectl logs -l app.kubernetes.io/component=operator -n tabclaw -f
```

**分配长时间停留在 `alloc_status: "allocating"`**
异步模型配置步骤（exec → `openclaw config set`）仍在进行或静默失败。查看 `claw-swarm-operator` 日志中的 `[alloc]` 条目。若 LiteLLM 不可用，系统会记录警告并使用随机 API 密钥继续完成分配。
