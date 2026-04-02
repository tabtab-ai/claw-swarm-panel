# Audit Log

## Overview

All critical operations on OpenClaw instances (allocate, pause, resume, free) are written to the audit log and persisted in the `audit_logs` table of the SQLite database.

Each audit log entry records:

| Field | Description |
|-------|-------------|
| `id` | Auto-increment primary key |
| `timestamp` | Local time when the operation occurred (`datetime('now','localtime')`) |
| `operator_id` | System user ID of the operator |
| `operator_username` | Username of the operator |
| `action` | Operation type: `alloc` / `free` / `pause` / `resume` |
| `target_user_id` | User ID of the target instance owner |
| `instance_name` | Name of the affected instance (StatefulSet name) |
| `status` | Operation result: `success` / `failed` |
| `detail` | Additional info, e.g. `model_type=lite`, `delay_minutes=10` |

---

## When Entries Are Written

| Operation | Trigger |
|-----------|---------|
| `alloc` | Written after `POST /claw/alloc` successfully allocates an instance; `detail` includes `model_type` |
| `free` | Written after `POST /claw/free` successfully deletes an instance |
| `pause` | Written after `POST /claw/pause` successfully sets the pause timestamp; `detail` includes `delay_minutes` |
| `resume` | Written after `POST /claw/resume` successfully resumes an instance |

> The current version only records successful operations. If an operation fails during parameter validation or Kubernetes lookup, no audit entry is written.

---

## API

### GET `/audit/logs` _(admin only)_

Paginated query of audit logs with optional action-type filtering.

**Permission:** Admin role required.

**Query parameters**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | 1 | Page number (1-indexed) |
| `page_size` | int | 20 | Number of records per page |
| `action` | string | — | Filter by action type: `alloc` / `free` / `pause` / `resume`; omit to return all |

**Request example**

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

**Error responses**

| Status | `msg` | Reason |
|--------|-------|--------|
| `403` | `admin only` | Non-admin user accessing the endpoint |
| `500` | `failed to list audit logs` | Database query failure |

---

## curl Examples

**Fetch the latest 20 log entries**

```bash
curl "http://<host>/audit/logs" \
  -H "X-API-Key: claw_<your-api-secret>"
```

**Fetch page 2, 10 entries per page, filtered to `free` actions**

```bash
curl "http://<host>/audit/logs?page=2&page_size=10&action=free" \
  -H "X-API-Key: claw_<your-api-secret>"
```

---

## Database Schema

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

## Web UI

After logging in as admin, the **Audit Log** entry appears in the navigation bar (`/audit`).

Page features:

- Filter by action type (All / Alloc / Free / Pause / Resume)
- Paginated view, 20 entries per page
- Columns: timestamp, operator, action, target user, instance, status, detail
- Manual refresh button in the top-right corner

> The audit log page is only visible to the `admin` role. Regular users cannot access the `/audit` route.
