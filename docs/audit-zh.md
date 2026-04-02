# Audit Log — 审计日志

## 概述

所有对 OpenClaw 实例的关键操作（分配、暂停、恢复、释放）均会写入审计日志，持久化存储在 SQLite 数据库的 `audit_logs` 表中。

审计日志记录以下信息：

| 字段 | 说明 |
|------|------|
| `id` | 日志自增主键 |
| `timestamp` | 操作发生的本地时间（SQLite `datetime('now','localtime')`） |
| `operator_id` | 操作者的系统用户 ID |
| `operator_username` | 操作者的用户名 |
| `action` | 操作类型：`alloc` / `free` / `pause` / `resume` |
| `target_user_id` | 被操作实例所属的目标用户 ID |
| `instance_name` | 被操作的实例名（StatefulSet name） |
| `status` | 操作结果：`success` / `failed` |
| `detail` | 附加信息，例如 `model_type=lite`、`delay_minutes=10` |

---

## 触发时机

| 操作 | 触发条件 |
|------|---------|
| `alloc` | `POST /claw/alloc` 成功分配实例后写入，`detail` 包含 `model_type` |
| `free` | `POST /claw/free` 成功删除实例后写入 |
| `pause` | `POST /claw/pause` 成功设置暂停时间戳后写入，`detail` 包含 `delay_minutes` |
| `resume` | `POST /claw/resume` 成功恢复实例后写入 |

> 当前版本仅记录成功操作。若操作在参数校验或 K8s 查询阶段失败，不会写入审计日志。

---

## API

### GET `/audit/logs` _(admin only)_

分页查询审计日志，支持按操作类型过滤。

**权限：** 仅 admin 角色可访问。

**Query 参数**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码（从 1 开始） |
| `page_size` | int | 20 | 每页条数 |
| `action` | string | — | 按操作类型过滤：`alloc` / `free` / `pause` / `resume`；不传则返回全部 |

**Request 示例**

```http
GET /audit/logs?page=1&page_size=20&action=alloc
X-API-Key: claw_<your-api-secret>
```

**Response `200 OK`**

```json
{
  "data": [
    {
      "id": 42,
      "timestamp": "2026-03-20T14:32:10Z",
      "operator_id": 1,
      "operator_username": "admin",
      "action": "alloc",
      "target_user_id": "user-12345",
      "instance_name": "claw-abc123",
      "status": "success",
      "detail": "model_type=lite"
    }
  ],
  "total": 128,
  "page": 1,
  "page_size": 20
}
```

**Error 响应**

| Status | `msg` | 原因 |
|--------|-------|------|
| `403` | `admin only` | 非 admin 用户访问 |
| `500` | `failed to list audit logs` | 数据库查询失败 |

---

## curl 示例

**查询最新 20 条日志**

```bash
curl "http://<host>/audit/logs" \
  -H "X-API-Key: claw_<your-api-secret>"
```

**查询第 2 页，每页 10 条，仅看 free 操作**

```bash
curl "http://<host>/audit/logs?page=2&page_size=10&action=free" \
  -H "X-API-Key: claw_<your-api-secret>"
```

---

## 数据库表结构

```sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id                INTEGER  PRIMARY KEY AUTOINCREMENT,
    timestamp         DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
    operator_id       INTEGER  NOT NULL,
    operator_username TEXT     NOT NULL,
    action            TEXT     NOT NULL,
    target_user_id    TEXT     NOT NULL DEFAULT '',
    instance_name     TEXT     NOT NULL DEFAULT '',
    status            TEXT     NOT NULL,
    detail            TEXT     NOT NULL DEFAULT ''
);
```

---

## WebUI

管理员登录后，导航栏会显示 **审计日志** 入口（`/audit`）。

页面功能：

- 按操作类型筛选（全部 / 分配 / 释放 / 暂停 / 恢复）
- 分页浏览，每页 20 条
- 显示操作时间、操作者、操作类型、目标用户、实例、状态和详情
- 右上角刷新按钮手动重新加载

> 审计日志页面仅对 admin 角色可见，普通用户无法访问 `/audit` 路由。
