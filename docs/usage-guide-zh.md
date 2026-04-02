# Claw Swarm Operation Console — 产品使用说明

## 目录

1. [产品概述](#1-产品概述)
2. [登录与认证](#2-登录与认证)
3. [控制台总览（Dashboard）](#3-控制台总览dashboard)
4. [实例管理（Instances）](#4-实例管理instances)
5. [分配实例](#5-分配实例)
6. [用户管理（仅管理员）](#6-用户管理仅管理员)
7. [API 访问](#7-api-访问)
8. [典型使用流程](#8-典型使用流程)

---

## 1. 产品概述

**Claw Swarm Operation Console** 是一套多集群 Kubernetes 容器实例管理平台，用于统一调度和管理 **OpenClaw** 会话容器。

### 核心能力

| 能力 | 说明 |
|------|------|
| **实例池管理** | 维护一批预先就绪的 OpenClaw 容器，按需分配给用户会话 |
| **一键分配/释放** | 通过用户 ID 将空闲实例分配给指定用户，用完即释放回池 |
| **实例生命周期** | 支持暂停 / 恢复 / 删除实例，实时监控 CPU & 内存 |
| **多租户用户管理** | 管理员可创建账号并分配角色（user / admin） |
| **REST API** | 提供完整 HTTP API，支持程序化接入 |

---

## 2. 登录与认证

### 登录步骤

1. 访问控制台地址，未登录时自动跳转到 `/login` 页面。
2. 填写管理员分配的**用户名**和**密码**，点击 **Login**。
3. 若账号被标记了 **Force Change Password**，首次登录后需先修改密码。

### 修改密码

右上角用户名 → 下拉菜单 → **Change Password**，填写当前密码和新密码（最少 6 位）。

### API Secret

右上角用户名 → 下拉菜单 → **API Secret**，可查看、复制或重新生成 API 密钥。

> **提示**：Token 保存在浏览器本地存储中，关闭标签页后仍保持登录，点击右上角 **Logout** 可主动退出。

---

## 3. 控制台总览（Dashboard）

登录后默认进入 Dashboard，提供集群全局概况。

### 统计指标卡片

| 指标 | 说明 |
|------|------|
| Total Instances | 全部 OpenClaw 实例数量 |
| Avg CPU Usage | 所有运行中实例的 CPU 均值 |
| Avg Memory | 所有运行中实例的内存均值 |
| Allocated Instances | 当前已绑定用户的实例数量 |

### 实例状态说明

| 状态 | 含义 |
|------|------|
| 🟢 Active | 运行中 |
| 🟡 Idle | 空闲（未分配） |
| 🔴 Error | 异常 |
| ⚫ Stopped | 已停止 |

### 实例列表

Dashboard 下方展示最近 5 条实例记录，支持按状态过滤（All / Running / Paused），点击实例名称可进入详情页。

> **提示**：页面每 10 秒自动刷新一次，也可点击刷新按钮手动触发。

---

## 4. 实例管理（Instances）

点击顶部导航 **Instances** 进入实例列表页。

### 过滤与搜索

- **Tab 切换**：All / Allocated / Free，快速切换已分配 / 空闲视图
- **搜索框**：按实例名称或用户 ID 模糊匹配

### 实例卡片操作

每张实例卡右上角 **⋮** 菜单提供以下操作：

| 操作 | 说明 |
|------|------|
| **Pause Instance** | 暂停实例（停止资源使用，保留绑定关系） |
| **Resume Instance** | 恢复已暂停的实例 |
| **Allocate to User** | 将空闲实例绑定到指定用户 ID |
| **Free Instance** | 解除绑定，将实例归还实例池 |
| **Delete Instance** | 永久删除实例（⚠️ 不可恢复） |

### 实例详情页

点击实例名称进入详情，可查看：

- **基本信息**：名称、状态、命名空间
- **资源配额**：CPU / 内存 Request & Limit
- **Claw Web UI**：实例自带的前端页面链接，可直接打开或复制
- **Gateway Token**：网关访问凭证，点击 **Reveal Token** 后可复制

> ⚠️ **警告**：Delete 操作不可撤销，请确认实例不再使用后再删除。

---

## 5. 分配实例

点击顶部右侧 **Allocate Instance** 按钮，进入分配流程。

### 操作步骤

1. **输入用户 ID**：填写目标用户的唯一标识（如会话 ID、用户名等），该 ID 将与实例绑定。
2. **点击 Allocate**：系统从空闲池中选取一个实例，自动完成绑定。
3. **查看分配结果**：成功后页面展示：
   - 实例名称与当前状态
   - 绑定的用户 ID（可复制）
   - Gateway Token（可复制）
   - Claw Web UI 链接（可直接打开）

> **提示**：若当前没有空闲实例，分配会失败。请先确保实例池中有空闲实例，或联系管理员扩充容量。

---

## 6. 用户管理（仅管理员）

管理员账号在导航栏可见 **Users** 入口。

### 创建用户

1. 点击右上角 **New User** 按钮，弹出创建对话框。
2. 填写以下信息：
   - **Username**：登录用户名
   - **Password**：初始密码（最少 6 位）
   - **Role**：`user`（普通用户）或 `admin`（管理员）
3. 点击 **Create**，用户即时生效。

### 删除用户

在用户列表行末点击删除按钮（当前登录账号不可自删）。

### 用户字段说明

| 字段 | 说明 |
|------|------|
| ID | 系统自动分配的用户编号 |
| Username | 登录用户名 |
| Role | `admin` 可访问所有功能；`user` 只能操作实例 |
| Force Change Pwd | 首次登录强制修改密码的标记 |
| Created At | 账号创建时间 |

---

## 7. API 访问

所有 UI 功能均有对应的 REST API，供外部系统（CI/CD、Agent 等）程序化调用。

### 认证方式

所有请求需在 Header 中携带 API 密钥：

```http
X-API-Key: <your-api-secret>
```

API Secret 在控制台右上角用户菜单 → **API Secret** 中获取。

### 接口列表

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/claw/instances` | 获取实例列表 |
| `GET` | `/claw/instances/{name}` | 获取单个实例详情 |
| `POST` | `/claw/alloc` | 分配实例（body: `{"user_id": "..."}` ） |
| `POST` | `/claw/free` | 释放实例（body: `{"name": "..."}` ） |
| `POST` | `/claw/pause` | 暂停实例（body: `{"name": "..."}` ） |
| `POST` | `/claw/resume` | 恢复实例（body: `{"name": "..."}` ） |
| `GET` | `/claw/token?name={name}` | 获取网关 Token |
| `GET` | `/auth/me` | 获取当前用户信息 |
| `POST` | `/auth/login` | 登录（body: `{"username": "...", "password": "..."}` ） |
| `GET` | `/users` | 获取用户列表（仅 admin） |
| `POST` | `/users` | 创建用户（仅 admin） |
| `DELETE` | `/users/{id}` | 删除用户（仅 admin） |

### 示例：分配实例

```bash
curl -X POST http://<host>/claw/alloc \
  -H "X-API-Key: <your-api-secret>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user-12345"}'
```

**响应示例：**

```json
{
  "name": "claw-abc123",
  "namespace": "default",
  "web_ui_url": "http://claw-abc123.example.com",
  "token": "eyJhbGci...",
  "status": "active"
}
```

### 示例：释放实例

```bash
curl -X POST http://<host>/claw/free \
  -H "X-API-Key: <your-api-secret>" \
  -H "Content-Type: application/json" \
  -d '{"name": "claw-abc123"}'
```

### 示例：获取实例列表

```bash
curl -X GET "http://<host>/claw/instances?occupied=false" \
  -H "X-API-Key: <your-api-secret>"
```

---

## 8. 典型使用流程

### 场景一：运维人员手动为用户分配会话容器

```
1. 登录控制台（管理员或运维账号）
2. 点击右上角 [Allocate Instance]
3. 输入用户 ID（如会话 ID），点击 Allocate
4. 复制 Gateway Token 和 Claw Web UI 链接，发给用户
5. 用户会话结束后 → Instances 列表 → Free Instance
```

### 场景二：外部 Agent 系统自动化分配

```
1. 登录控制台 → 用户菜单 → API Secret，复制密钥
2. 调用 POST /claw/alloc（携带 X-API-Key 和 user_id）
3. 从响应中获取实例名、Token、Web UI 链接
4. 会话结束后调用 POST /claw/free 释放资源
```

**代码示例（Python）：**

```python
import requests

BASE_URL = "http://<host>"
API_KEY = "<your-api-secret>"
HEADERS = {"X-API-Key": API_KEY, "Content-Type": "application/json"}

# 分配实例
resp = requests.post(f"{BASE_URL}/claw/alloc",
                     json={"user_id": "session-001"},
                     headers=HEADERS)
instance = resp.json()
print(f"实例名: {instance['name']}")
print(f"Web UI: {instance['web_ui_url']}")
print(f"Token:  {instance['token']}")

# 使用完毕后释放
requests.post(f"{BASE_URL}/claw/free",
              json={"name": instance["name"]},
              headers=HEADERS)
```

### 场景三：临时暂停资源占用

```
1. 在 Instances 列表中搜索实例名或用户 ID
2. 点击 ⋮ 菜单 → Pause Instance
3. 实例进入暂停态，不消耗 CPU，保留绑定关系
4. 需要时点击 Resume Instance 恢复
```

> **建议**：在生产环境中通过 API 自动化完成分配和释放，减少人工操作，提升资源利用率。

---

*如有问题请联系系统管理员。*
