# 产品需求文档 — Claw Swarm Panel

**版本：** 1.0
**日期：** 2026-03-19
**状态：** 持续更新中

---

## 目录

1. [产品概述](#1-产品概述)
2. [目标用户](#2-目标用户)
3. [产品目标与非目标](#3-产品目标与非目标)
4. [核心价值主张](#4-核心价值主张)
5. [系统架构](#5-系统架构)
6. [功能需求](#6-功能需求)
   - 6.1 认证与安全
   - 6.2 用户管理
   - 6.3 实例池管理
   - 6.4 实例生命周期操作
   - 6.5 网关 Token 管理
   - 6.6 LiteLLM 集成
   - 6.7 Web 管理界面
7. [数据模型](#7-数据模型)
8. [API 概览](#8-api-概览)
9. [配置说明](#9-配置说明)
10. [部署方案](#10-部署方案)
11. [集成说明](#11-集成说明)
12. [非功能性需求](#12-非功能性需求)
13. [未来规划](#13-未来规划)

---

## 1. 产品概述

### 背景

随着 AI 编程助手技术的成熟，越来越多的企业希望为员工提供专属的 AI 编程环境。**OpenClaw** 是一款高性能的 AI 编程助手容器，能够大幅提升开发人员的工作效率。然而，在企业环境中批量管理这些容器面临以下挑战：

- 手动创建和分配容器效率极低，无法满足企业规模化需求
- 缺乏统一的管控平台，运维成本高
- 员工需要独占的隔离环境，数据不互通、互不干扰
- 容器资源浪费问题突出——大量容器在员工下班后仍在运行

### 产品定义

**Claw Swarm Panel**（以下简称"管理面板"）是一个面向**中小型企业**的 Kubernetes 原生 OpenClaw 实例管理平台。它实现了以下核心能力：

- 为企业内每位员工分配一个**专属的、隔离的** OpenClaw 工作环境
- 维护一个预热实例池，员工使用时**秒级分配**，无需等待启动
- 闲置实例自动挂起，员工上线自动恢复，**按需使用、节省资源**
- 提供统一的 Web 管理界面，IT 管理员可通过可视化操作完成全部管理工作
- 通过 API 与企业内部系统（如 HR 系统、SSO、对话平台）打通，实现自动化分配

### 典型使用场景

> **场景：一家 50 人的软件开发公司**
>
> IT 管理员在公司 Kubernetes 集群中部署 Claw Swarm Panel，配置实例池大小为 60（略大于员工人数）。每名开发人员拥有自己专属的 OpenClaw 实例，工作数据持久存储在各自的 PVC 中互不干扰。员工上班登录时，系统自动为其分配实例（如已有则直接恢复）；下班后实例自动挂起节省资源。IT 管理员通过 Web 界面实时查看所有实例状态，并可一键对指定员工的实例进行操作。

---

## 2. 目标用户

### 主要目标群体：中小型企业（10～500 人）

Claw Swarm Panel 专为以下类型的中小型企业设计：

| 企业类型 | 典型规模 | 需求特点 |
|---------|---------|---------|
| 软件开发公司 | 10～200 人 | 开发人员多，人均需要独立编程环境，对 AI 工具依赖程度高 |
| 互联网/科技公司 | 20～500 人 | 研发团队为核心，需要快速为新员工开通 AI 工具 |
| 企业研发部门 | 10～100 人 | 隶属于大型企业的研发团队，独立管理内部工具 |
| 外包/咨询公司 | 10～100 人 | 项目制工作，需要灵活分配和回收实例 |

### 系统使用角色

#### IT 管理员（Admin）

企业 IT 部门或平台负责人，负责系统的日常运营与管理。

| 职责 | 操作 |
|-----|------|
| 用户管理 | 创建/删除员工账号，设置角色 |
| 实例分配 | 为员工手动分配 OpenClaw 实例 |
| 实例回收 | 释放离职员工的实例，回收资源 |
| 实例监控 | 查看所有实例状态、资源使用情况 |
| 网关 Token 管理 | 查看各实例的认证 Token |
| 系统配置 | 调整实例池大小、资源规格、模型配置等 |

#### 普通员工（User）

企业中使用 OpenClaw 的开发人员或其他技术人员。

| 职责 | 操作 |
|-----|------|
| 查看实例 | 查看自己的实例状态 |
| 访问工作台 | 通过网关 URL 访问专属 OpenClaw 工作台 |
| 账号管理 | 修改密码、管理 API Secret |

#### 自动化系统（Automation）

与企业内部平台（对话系统、HR 系统、入职流程等）集成的自动化客户端。

| 能力 | 说明 |
|-----|------|
| API 分配 | 通过 `POST /claw/alloc` 为指定用户自动分配实例 |
| API 回收 | 通过 `POST /claw/free` 在员工离职时自动回收 |
| 状态查询 | 定时轮询实例状态，用于监控告警 |

---

## 3. 产品目标与非目标

### 目标

- ✅ 为企业每位员工提供**专属独占**的 OpenClaw 实例
- ✅ 实例**秒级分配**（预热池模式，API 响应 < 500ms）
- ✅ 闲置实例**自动挂起**节省资源（可配置超时时间）
- ✅ 提供 Web 管理界面，无需命令行操作即可完成日常管理
- ✅ 支持 REST API，可与企业内部系统集成
- ✅ 支持 JWT 和 API Secret 双认证，兼顾交互使用和自动化场景
- ✅ 用户数据持久存储，重启不丢失
- ✅ 支持阿里云 ACK 等国内主流云平台

### 非目标

- ❌ **不是多租户 SaaS 平台**（面向单个企业内部部署）
- ❌ **不提供实时资源监控**（CPU/内存用量展示为占位，不采集实时数据）
- ❌ **不开发 OpenClaw 应用本身**（仅负责管理和编排容器）
- ❌ **不涉及计费功能**（企业内部使用，无需计量收费）
- ❌ **不支持多集群联邦**（单 Kubernetes 集群部署）
- ❌ **不提供高可用数据库**（使用 SQLite，适合中小规模；大规模场景可替换）

---

## 4. 核心价值主张

### 对比传统方案

| 维度 | 传统方案（手动管理） | Claw Swarm Panel |
|-----|-------------------|-----------------|
| 实例分配速度 | 手动操作，需 5～15 分钟 | API 调用，< 500ms |
| 员工隔离性 | 难以保证独立环境 | 每人专属实例，数据完全隔离 |
| 资源利用率 | 容器长期运行，资源浪费 | 自动挂起/恢复，按需使用 |
| 运维成本 | 需要熟悉 kubectl，门槛高 | Web 界面操作，运维友好 |
| 与内部系统集成 | 无标准接口 | REST API，易于集成 |
| AI 模型接入 | 需手动配置每个容器 | 分配时自动注入 LiteLLM 配置 |

### 核心优势

1. **员工专属环境**：每人一个隔离的 OpenClaw 实例，工作区数据、配置、历史记录完全独立，互不干扰
2. **秒级响应**：预热实例池保证分配请求立即返回，员工无需等待容器冷启动
3. **智能资源管理**：员工下班后实例自动挂起（0 副本），上班时一键恢复，资源利用率最大化
4. **零运维门槛**：Web 管理界面覆盖全部日常操作，IT 管理员无需了解 Kubernetes
5. **开箱即用的 AI 模型接入**：分配时自动为员工配置 LiteLLM API Key，无需手动配置
6. **企业级安全**：bcrypt 密码加密、JWT 认证、基于角色的权限控制

---

## 5. 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                       Claw Swarm Panel                           │
│                                                                  │
│  ┌────────────────┐     ┌──────────────────────────────────────┐ │
│  │   Web 管理界面  │     │          HTTP 管理 API               │ │
│  │  (React, 80)  │◄───►│   /auth  /users  /claw  (8088)       │ │
│  └────────────────┘     └──────────────────┬─────────────────--┘ │
│                                            │                     │
│                          ┌─────────────────▼──────────────────┐  │
│                          │         Kubernetes Operator         │  │
│                          │        (controller-runtime)         │  │
│                          │                                     │  │
│                          │  ClawReconciler ──► 实例池同步      │  │
│                          │  PVCReconciler  ──► 存储卷清理      │  │
│                          └─────────────────┬───────────────────┘  │
│                                            │                      │
│  ┌──────────────────┐   ┌─────────────────▼──────────────────┐   │
│  │   SQLite 数据库   │   │          Kubernetes API             │   │
│  │   (claw.db)      │   │  StatefulSets / Services /          │   │
│  │  - 用户表         │   │  Ingresses / PVCs                   │   │
│  │  - Token 表       │   └────────────────────────────────────┘   │
│  └──────────────────┘                                             │
└─────────────────────────────────────────────────────────────────┘
                               │
               ┌───────────────▼──────────────────┐
               │            LiteLLM                │
               │   AI 模型路由 & API Key 管理       │
               └──────────────────────────────────┘

                     每个员工 → 一个专属 OpenClaw 实例
                     ┌──────────┐ ┌──────────┐ ┌──────────┐
                     │ 员工 A   │ │ 员工 B   │ │ 员工 C   │
                     │ 实例-001 │ │ 实例-002 │ │ 实例-003 │
                     │  PVC-A   │ │  PVC-B   │ │  PVC-C   │
                     └──────────┘ └──────────┘ └──────────┘
```

### 组件职责

| 组件 | 职责 |
|-----|------|
| **Web 管理界面** | 实例列表/详情、用户管理、分配操作、认证流程 |
| **HTTP 管理 API** | 所有实例和用户操作的 REST 接口 |
| **K8s Operator（ClawReconciler）** | 维护实例池大小，管理实例创建/删除/生命周期 |
| **K8s Operator（PVCReconciler）** | 清理已删除实例的孤立 PVC（基于 Redis 状态追踪） |
| **SQLite** | 用户信息、bcrypt 密码、API Secret、网关 Token |
| **LiteLLM** | 分配时自动为员工生成 AI 模型 API Key |

---

## 6. 功能需求

### 6.1 认证与安全

#### 两种认证方式

**JWT Bearer Token**（交互式客户端、Web 界面）
- 通过 `POST /auth/login` 获取
- 算法：HS256，32 字节密钥签名
- 有效期：24 小时
- 请求头：`Authorization: Bearer <token>`

**API Secret**（自动化脚本、企业系统集成）
- 格式：`claw_<40位十六进制>` （20 字节随机数）
- 永久有效，可按需重新生成
- 请求头：`X-API-Key: claw_...`

#### 密码安全
- 使用 bcrypt 加密存储（代价因子 10）
- 新密码最短 6 位
- `force_change_password` 标志：首次登录强制修改密码

#### 权限控制
- 写操作（分配/回收/暂停/恢复）仅限 `admin` 角色
- 读操作（查看实例列表）所有认证用户均可访问
- 非管理员执行写操作返回 `403 Forbidden`

#### 默认管理员账号

| 字段 | 值 |
|-----|---|
| 用户名 | `admin` |
| 初始密码 | `happyclaw` |
| 首次登录 | 强制修改密码 |

---

### 6.2 用户管理

IT 管理员通过 Web 界面或 API 管理企业员工账号。

#### 创建用户
- 字段：用户名（唯一）、密码、角色（`admin` | `user`）
- 系统自动生成 API Secret
- 默认角色：`user`

#### 用户列表
- 显示所有用户：ID、用户名、角色、创建时间等

#### 删除用户
- 删除员工账号（不能删除自己的账号）
- 建议：删除用户前先释放其关联的实例

#### 员工自助服务
- **修改密码**：需提供旧密码，成功后签发新 JWT
- **查看 API Secret**：通过 Web 界面或 `GET /auth/api-secret`
- **重新生成 API Secret**：旧 Secret 立即失效

---

### 6.3 实例池管理

Kubernetes Operator 自动维护一个预热实例池，保证随时有可用实例。

#### 实例池工作原理

```
目标：始终保持 pool_size 个空闲实例

  当前空闲数 < pool_size  →  Operator 自动创建新实例补充
  当前空闲数 > pool_size  →  Operator 自动删除多余空闲实例
```

#### 每个实例包含的 Kubernetes 资源

| 资源类型 | 说明 |
|---------|-----|
| StatefulSet | 实例的核心资源，管理 Pod 生命周期 |
| Pod | OpenClaw 容器，端口 18789 |
| PersistentVolumeClaim | 员工工作区存储，默认 2Gi |
| Service | ClusterIP，端口 18789 |
| Ingress | 路由 `/<实例名>/` 到对应 Service |

#### 实例状态说明

| `state` 值 | 含义 | 触发条件 |
|-----------|-----|---------|
| `"pending"` | 启动中 | 副本数 > 0 但 Pod 尚未就绪 |
| `"running"` | 运行中 | 所有副本就绪 |
| `"paused"` | 已挂起 | 副本数 = 0，或已设置挂起计划时间注解 |

#### 分配状态说明

| `alloc_status` 值 | 含义 |
|------------------|-----|
| `""` （空）| 实例空闲，未分配给任何员工 |
| `"allocating"` | 分配中：实例已占用，模型配置正在后台异步执行 |
| `"allocated"` | 分配完成：模型配置成功，实例可正常使用 |

> **说明：** 分配 API (`POST /claw/alloc`) 会立即返回 `"allocating"` 状态的响应，模型配置（向 Pod 注入 LiteLLM 配置）在后台异步完成。客户端可轮询 `GET /claw/instances` 直到 `alloc_status` 变为 `"allocated"`。

---

### 6.4 实例生命周期操作

#### 分配实例（`POST /claw/alloc`）

**适用场景：** 员工入职、首次使用时，由 IT 管理员或自动化系统为其分配专属实例。

操作流程：
1. 加锁（防止并发竞争）
2. 从池中找到最早创建的空闲实例
3. 若该用户已有实例，直接返回（幂等）
4. 设置实例标签：`occupied=<user_id>`，`alloc_status=allocating`，副本数置为 1
5. **立即返回响应**
6. 后台异步：向 Pod 注入 LiteLLM 模型配置
7. 后台异步完成：更新 `alloc_status=allocated`

#### 释放实例（`POST /claw/free`）

**适用场景：** 员工离职、长期不使用时回收实例资源。

- 删除 StatefulSet，Kubernetes 通过 OwnerReference 级联删除 Service 和 Ingress
- PVC 保留（Operator 单独管理清理）
- Operator 检测到删除后自动补充新的空闲实例到池中

#### 挂起实例（`POST /claw/pause`）

**适用场景：** 员工临时不使用，或系统检测到闲置超时后自动触发。

- 设置注解 `tabtab.app.scheduled.deletion.time = <当前时间 + delay_pause_minutes>`
- Operator 在定时器到期后将副本数置为 0
- 挂起期间 `state` 报告为 `"paused"`

#### 恢复实例（`POST /claw/resume`）

**适用场景：** 员工重新使用时，由分配平台主动调用恢复实例。

- 清除挂起注解
- 将副本数置为 1
- Pod 重新启动，状态从 `"pending"` 变为 `"running"`

---

### 6.5 网关 Token 管理

每个 OpenClaw 实例有一个对应的网关认证 Token，存储在 SQLite 的 `claw_tokens` 表中。

- Token 由外部系统（Operator 初始化流程）写入，长期有效
- `GET /claw/token?name=<实例名>` 可查询 Token，**不会消耗或删除 Token**
- Token 为空时，表示该实例尚未被分配或尚未初始化
- 分配响应（`POST /claw/alloc`）中也会返回当前 Token

---

### 6.6 LiteLLM 集成

每次分配实例时，系统自动为员工在 LiteLLM 中创建账号并生成 API Key，注入到 OpenClaw 实例中。

#### 分配流程中的 AI 配置

1. 根据 `model_type`（`lite` 或 `pro`，默认 `lite`）确定模型
2. 调用 LiteLLM `POST /user/new` 创建用户或获取已有用户的 Key
3. 通过 `kubectl exec` 进入 Pod，执行 `openclaw config set` 注入模型配置
4. 设置 `agents.defaults.model.primary = tabtab-litellm/<model_id>`

#### 容错处理

- LiteLLM 不可用时：自动生成随机 API Key，分配操作不受影响
- 模型注入失败时：仅记录警告日志，不影响实例分配

#### 模型层级

| 层级 | `model_type` 参数 | 说明 |
|-----|-----------------|-----|
| 标准版 | `"lite"`（默认） | 适合日常编程辅助场景 |
| 专业版 | `"pro"` | 适合复杂任务、需要更强模型能力的场景 |

---

### 6.7 Web 管理界面

基于 React 18 + TypeScript + Zustand + shadcn/ui + Tailwind CSS 构建。

#### 页面清单

##### 登录页（`/login`）
- 用户名/密码输入
- 登录成功后 JWT 存入 `localStorage`，跳转至仪表盘
- 首次登录触发强制修改密码流程

##### 仪表盘（`/`）
- **统计卡片**：总实例数、运行中、空闲、已分配数量
- **实例列表**（最新 5 条）：状态指示器、快捷操作
- **筛选器**：按状态（全部/运行中/挂起/启动中）和分配状态（全部/已分配/空闲）筛选
- 异常实例提示横幅

##### 实例列表（`/instances`）
- 所有实例的网格视图
- 按实例名或员工 ID 搜索
- 每个实例显示：状态点、实例名、员工 ID（可复制）、分配状态徽章、Web UI 链接
- 快捷操作：暂停/恢复、分配、释放

##### 实例详情（`/instances/:id`）
页面包含以下信息块：

| 信息块 | 内容 |
|------|-----|
| 页头 | 实例名、状态徽章（空闲/分配中/已分配）、操作按钮 |
| 基本信息 | 名称、运行状态、分配状态（三色徽章） |
| 用户归属 | 分配的员工 ID（可复制） |
| 资源配置 | CPU/内存 request/limit |
| OpenClaw Web UI | 网关访问链接（可复制、可直接打开） |
| 网关 Token | Token 值（可复制），空时显示"未分配" |

分配状态徽章颜色：
- 🟡 黄色 `Allocating` — 分配中
- 🔵 青色 `Allocated` — 已分配完成
- ⚫ 灰色 `Free` — 空闲

##### 创建/分配实例（`/instances/new`）
- **表单页**：输入员工 ID，提交分配请求
- **结果页**：显示实例名、状态、员工 ID、网关 Token（如有）、Web UI 地址

##### 用户管理（`/users`，仅管理员）
- 员工账号列表：ID、用户名、角色徽章、创建时间
- 新建用户对话框：用户名、密码（最短 6 位）、角色选择
- 删除用户（防止误删自己）

#### 顶部导航
- Logo + 产品名称"Claw Swarm Operation Console"
- 导航链接：仪表盘、实例列表、创建实例、用户管理（仅管理员）
- 用户下拉菜单：
  - 当前用户名 + 角色徽章
  - 修改密码对话框
  - API Secret 查看/复制/重新生成
  - 退出登录

---

## 7. 数据模型

### 实例（API 响应）

```json
{
  "name":           "claw-abc123",
  "user_id":        "alice",
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

### 用户（API 响应）

```json
{
  "id": 1,
  "username": "alice",
  "role": "user",
  "force_change_password": false,
  "api_secret": "claw_a3f2e1d4b5c6...",
  "created_at": "2026-01-01T00:00:00Z"
}
```

### K8s 标签

| 标签键 | 值 | 说明 |
|-------|---|-----|
| `tabtabai.com/tabclaw` | `""` | 标记 Operator 管理的实例 |
| `tabtabai.com/tabclaw-occupied` | `<user_id>` | 已分配给该员工 |
| `tabtabai.com/tabclaw-alloc-status` | `allocating` \| `allocated` | 分配阶段 |

### SQLite 表结构

#### users 表
```sql
id                  INTEGER  PRIMARY KEY
username            TEXT     UNIQUE
password_hash       TEXT     -- bcrypt 加密
role                TEXT     -- 'admin' | 'user'
force_change_password INTEGER -- 0 | 1
api_secret          TEXT     -- 'claw_<40hex>'
created_at          DATETIME
```

#### claw_tokens 表
```sql
name       TEXT     PRIMARY KEY  -- StatefulSet 名称
token      TEXT     NOT NULL     -- 网关认证 Token
created_at DATETIME DEFAULT (datetime('now','localtime'))
```

---

## 8. API 概览

完整文档见 [`docs/api.md`](./api.md)

### 认证接口

| 方法 | 路径 | 权限 | 说明 |
|-----|------|-----|-----|
| `POST` | `/auth/login` | 公开 | 获取 JWT Token |
| `GET` | `/auth/me` | 所有人 | 当前用户信息 |
| `POST` | `/auth/change-password` | 所有人 | 修改密码 |
| `GET` | `/auth/api-secret` | 所有人 | 查看 API Secret |
| `POST` | `/auth/api-secret/regenerate` | 所有人 | 重新生成 API Secret |

### 用户管理（仅管理员）

| 方法 | 路径 | 说明 |
|-----|------|-----|
| `GET` | `/users` | 获取所有用户列表 |
| `POST` | `/users` | 创建新用户 |
| `DELETE` | `/users/:id` | 删除用户 |

### 实例操作

| 方法 | 路径 | 权限 | 说明 |
|-----|------|-----|-----|
| `GET` | `/claw/instances` | 所有人 | 查询所有实例（可过滤） |
| `GET` | `/claw/used` | 所有人 | 查询已分配实例 |
| `GET` | `/claw/token` | 管理员 | 查询实例网关 Token |
| `POST` | `/claw/alloc` | 管理员 | 分配实例给员工 |
| `POST` | `/claw/free` | 管理员 | 释放实例 |
| `POST` | `/claw/pause` | 管理员 | 挂起实例 |
| `POST` | `/claw/resume` | 管理员 | 恢复实例 |

---

## 9. 配置说明

配置文件：`config.yaml`（可通过 `--config` 指定路径）。CLI 参数优先级高于配置文件。

### `server` 配置块

| 配置项 | 默认值 | 说明 |
|-------|-------|-----|
| `port` | `8088` | HTTP API 监听端口 |
| `domain` | — | 实例网关 URL 的基础域名 |
| `tls_secret_name` | `tabtabai` | Ingress TLS 证书 Secret 名称 |
| `ingress_name` | `portal-ingress` | Ingress 类名 |
| `delay_pause_minutes` | `1` | 闲置多少分钟后自动挂起（分钟） |
| `node_selector_value` | `""` | 指定运行节点的标签值（空=不限制） |
| `jwt_secret` | `""` | JWT 签名密钥（空则自动生成并持久化） |
| `data_dir` | `.claw-swarm` | SQLite 数据库和 JWT 密钥文件目录 |

### `operator` 配置块

| 配置项 | 默认值 | 说明 |
|-------|-------|-----|
| `runtime_image` | `ghcr.io/openclaw/openclaw:latest` | OpenClaw 容器镜像 |
| `storage_class` | `standard` | PVC 使用的 StorageClass |

### `init` 配置块

| 配置项 | 默认值 | 说明 |
|-------|-------|-----|
| `pool_size` | `2` | 目标空闲实例数量（建议略大于员工数） |
| `storage_size` | `2Gi` | 每个实例的 PVC 容量 |
| `gateway.port` | `18789` | OpenClaw 网关端口 |
| `gateway.auth` | `token` | 网关认证方式 |
| `gateway.token` | — | 网关 Bearer Token |
| `litellm.baseurl` | — | LiteLLM 服务地址 |
| `litellm.master_key` | — | LiteLLM 管理员密钥 |
| `litellm.models.lite.model_id` | `tabtab-lite` | 标准版模型 ID |
| `litellm.models.pro.model_id` | `tabtab-pro` | 专业版模型 ID |

### 资源配置建议（中小型企业参考值）

| 规模 | CPU Request | CPU Limit | Memory Request | Memory Limit | pool_size |
|-----|------------|----------|----------------|-------------|----------|
| 小型（< 20 人）| `250m` | `1` | `512Mi` | `2Gi` | 人数 + 3 |
| 中型（20～100 人）| `500m` | `2` | `1Gi` | `4Gi` | 人数 + 5 |
| 中大型（100～500 人）| `500m` | `2` | `1Gi` | `4Gi` | 人数 + 10 |

---

## 10. 部署方案

### Helm 安装

```bash
helm install claw-swarm-panel ./charts/claw-swarm-panel \
  --namespace claw-system \
  --create-namespace \
  --values my-values.yaml
```

### 关键 Chart 资源

| 资源 | 说明 |
|-----|-----|
| `Deployment` | Operator 主进程 |
| `ConfigMap` | 渲染后的 `config.yaml`，挂载到 Operator Pod |
| `ServiceAccount` + RBAC | 管理 StatefulSet/Service/Ingress/PVC 的集群权限 |
| `Service` | ClusterIP，端口 8088 |
| `Ingress` | 路由 `/claw`、`/auth`、`/users` 路径到 Service |
| `PVC` | 持久化存储（SQLite 数据库、JWT 密钥） |

### 数据持久化要求

Operator 需要将 `data_dir` 目录挂载到 PVC，其中包含：
- `claw.db` — SQLite 数据库（用户、Token）
- `.jwt_secret` — JWT 签名密钥

> ⚠️ **重要：** 不配置持久化 PVC 时，Pod 重启后所有用户数据和 Token 将丢失。

### RBAC 权限要求

Operator ServiceAccount 需要以下 Kubernetes 权限：

- `StatefulSets`：get / list / watch / create / update / patch / delete
- `Services`、`Ingresses`：get / list / watch / create / update / patch / delete
- `PersistentVolumeClaims`：get / list / watch / delete
- `Pods`：get / list / watch / exec（用于模型配置注入）

---

## 11. 集成说明

### LiteLLM 集成

| 集成点 | 说明 |
|-------|-----|
| 用户创建 | 分配时调用 `POST /user/new` |
| Key 生成 | 已有用户调用 `POST /key/generate` |
| 认证方式 | Master Key 通过 `Authorization: Bearer` 传递 |
| 模型注入 | 通过 Pod Exec 执行 `openclaw config set` |
| 容错 | LiteLLM 不可用时使用随机 Key，不阻塞分配流程 |

### Kubernetes 集成

| 集成点 | 说明 |
|-------|-----|
| 客户端 | `controller-runtime` + `client-go` |
| Watch | 监听带 `tabtabai.com/tabclaw` 标签的 StatefulSet |
| Pod Exec | 使用 SPDY 协议执行模型配置命令 |
| 资源管理 | StatefulSets / Services / Ingresses / PVCs / Jobs |

### 企业系统集成（典型场景）

| 系统 | 集成方式 | 说明 |
|-----|---------|-----|
| HR/OA 系统 | REST API（API Secret 认证） | 员工入职时自动调用 `/claw/alloc` 分配实例 |
| 对话平台 | REST API | 用户首次对话时触发分配，后续通过 `user_id` 复用实例 |
| 监控系统 | 轮询 `GET /claw/instances` | 定期采集实例状态，生成告警 |
| 离职流程 | REST API | 员工离职时自动调用 `/claw/free` 回收资源 |

---

## 12. 非功能性需求

### 性能

| 指标 | 目标值 |
|-----|-------|
| 分配 API 响应时间 | < 500ms（模型配置异步处理） |
| 实例池补充延迟 | 删除实例后 < 30s 内补充新实例 |
| Web UI 首次加载 | < 3s |
| 并发分配请求 | 通过互斥锁串行处理，保证数据一致性 |

### 可靠性

| 要求 | 实现方式 |
|-----|---------|
| Pod 重启不丢用户数据 | SQLite 存储到 PVC，JWT 密钥持久化 |
| Operator 崩溃恢复 | controller-runtime 幂等 Reconcile，重启后自动追平状态 |
| 分配操作幂等 | 同一 `user_id` 多次调用返回同一实例 |

### 安全

| 要求 | 实现 |
|-----|-----|
| 密码加密 | bcrypt（代价因子 10） |
| JWT 安全 | HS256，32 字节随机密钥，24 小时过期 |
| API Secret 熵值 | 20 字节 `crypto/rand`，`claw_` 前缀 |
| 文件权限 | `data_dir` 目录 0700，密钥文件 0600 |
| HTTP/2 | 默认禁用（CVE 规避） |

### 可观测性

| 能力 | 说明 |
|-----|-----|
| 结构化日志 | `klog` 分级日志，关键操作均有日志记录 |
| 健康检查 | `/healthz`（存活）、`/readyz`（就绪），端口 8081 |
| pprof 性能分析 | `--profiling=true` 时开启，路径 `/debug/pprof/` |
| Prometheus 指标 | 通过 controller-runtime 内置指标服务暴露 |

---

## 13. 未来规划

以下功能已识别为潜在增强方向，**当前版本不在范围内**：

| 功能 | 描述 | 优先级 |
|-----|-----|-------|
| **实时资源监控** | 从 K8s metrics-server 采集实例 CPU/内存用量 | 高 |
| **批量导入用户** | 支持从 CSV 或对接 LDAP/AD 批量创建员工账号 | 高 |
| **Webhook 通知** | 分配/释放/挂起事件触发外部通知（企业微信、钉钉等） | 中 |
| **实例使用审计日志** | 记录所有分配/释放/操作事件，支持合规审计 | 中 |
| **自动池大小调整** | 根据分配需求动态调整 `pool_size` | 中 |
| **多集群管理** | 通过 kubeconfig 管理多个 K8s 集群的实例 | 低 |
| **高可用数据库** | 替换 SQLite 为 PostgreSQL，支持多副本部署 | 低 |
| **实例规格模板** | 支持多种资源规格，分配时按需选择 | 低 |
| **技能预装** | Init 容器从镜像复制预装技能到工作区 | 低 |
