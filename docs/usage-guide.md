# Claw Swarm Operation Console — Usage Guide

## Table of Contents

1. [Product Overview](#1-product-overview)
2. [Login & Authentication](#2-login--authentication)
3. [Dashboard](#3-dashboard)
4. [Instance Management](#4-instance-management)
5. [Allocating an Instance](#5-allocating-an-instance)
6. [User Management (Admin only)](#6-user-management-admin-only)
7. [API Access](#7-api-access)
8. [Typical Workflows](#8-typical-workflows)

---

## 1. Product Overview

**Claw Swarm Operation Console** is a multi-cluster Kubernetes container instance management platform for centrally scheduling and operating **OpenClaw** session containers.

### Core Capabilities

| Capability | Description |
|-----------|-------------|
| **Instance pool management** | Maintains a pool of pre-warmed OpenClaw containers, allocated to user sessions on demand |
| **One-click allocate / free** | Binds an idle instance to a specified user by user ID; releases it back to the pool when done |
| **Instance lifecycle** | Supports pause / resume / delete with real-time CPU & memory monitoring |
| **Multi-tenant user management** | Admins can create accounts and assign roles (`user` / `admin`) |
| **REST API** | Full HTTP API for programmatic integration |

---

## 2. Login & Authentication

### Login Steps

1. Navigate to the console URL. Unauthenticated users are automatically redirected to `/login`.
2. Enter your **username** and **password**, then click **Login**.
3. If the account has **Force Change Password** set, you will be prompted to change your password on first login.

### Changing Your Password

Top-right username → dropdown menu → **Change Password**. Enter your current password and a new password (minimum 6 characters).

### API Secret

Top-right username → dropdown menu → **API Secret**. You can view, copy, or regenerate your API key here.

> **Tip:** The JWT token is stored in browser `localStorage`. You remain logged in after closing the tab. Click **Logout** in the top-right menu to sign out explicitly.

---

## 3. Dashboard

After login you land on the Dashboard, which provides a global cluster overview.

### Summary Metric Cards

| Metric | Description |
|--------|-------------|
| Total Instances | Total number of OpenClaw instances |
| Avg CPU Usage | Average CPU across all running instances |
| Avg Memory | Average memory across all running instances |
| Allocated Instances | Number of instances currently bound to a user |

### Instance State Reference

| State | Meaning |
|-------|---------|
| 🟢 Active | Running |
| 🟡 Idle | Free (not allocated) |
| 🔴 Error | Abnormal |
| ⚫ Stopped | Paused / stopped |

### Instance List

The lower section of the Dashboard shows the 5 most recent instances with state filtering (All / Running / Paused). Click an instance name to open its detail page.

> **Tip:** The page auto-refreshes every 10 seconds. You can also click the refresh button to reload manually.

---

## 4. Instance Management

Click **Instances** in the top navigation to open the instance list.

### Filtering & Search

- **Tab switching:** All / Allocated / Free — quickly switch between allocated and idle views
- **Search box:** Fuzzy match by instance name or user ID

### Instance Card Actions

Click the **⋮** menu in the top-right corner of any instance card:

| Action | Description |
|--------|-------------|
| **Pause Instance** | Suspend the instance (stops resource usage, retains the user binding) |
| **Resume Instance** | Resume a paused instance |
| **Allocate to User** | Bind a free instance to a specified user ID |
| **Free Instance** | Release the binding and return the instance to the pool |
| **Delete Instance** | Permanently delete the instance (⚠️ irreversible) |

### Instance Detail Page

Click an instance name to open its detail page:

- **Basic info:** name, state, namespace
- **Resource quota:** CPU / memory Request & Limit
- **Claw Web UI:** link to the instance's built-in frontend; can be opened or copied directly
- **Gateway Token:** gateway access credential — click **Reveal Token** to view and copy

> ⚠️ **Warning:** Delete is irreversible. Confirm the instance is no longer needed before deleting.

---

## 5. Allocating an Instance

Click the **Allocate Instance** button in the top-right area to begin the allocation flow.

### Steps

1. **Enter user ID:** The unique identifier for the target user (e.g. session ID, username). This ID is bound to the instance.
2. **Click Allocate:** The system picks an idle instance from the pool and binds it automatically.
3. **View allocation result:** On success the page shows:
   - Instance name and current state
   - Bound user ID (copyable)
   - Gateway Token (copyable)
   - Claw Web UI link (openable directly)

> **Tip:** If there are no idle instances, allocation will fail. Ensure the pool has spare capacity or ask an admin to scale up.

---

## 6. User Management (Admin only)

Admin accounts see the **Users** entry in the navigation bar.

### Creating a User

1. Click **New User** in the top-right corner to open the creation dialog.
2. Fill in:
   - **Username:** login username
   - **Password:** initial password (minimum 6 characters)
   - **Role:** `user` (standard) or `admin` (administrator)
3. Click **Create**. The user is active immediately.

### Deleting a User

Click the delete button at the end of a user's row (the currently logged-in account cannot be self-deleted).

### User Field Reference

| Field | Description |
|-------|-------------|
| ID | System-assigned user number |
| Username | Login username |
| Role | `admin` has access to all features; `user` can only view instances |
| Force Change Pwd | Flag that prompts a password change on first login |
| Created At | Account creation timestamp |

---

## 7. API Access

All UI functionality has a corresponding REST API for programmatic access by external systems (CI/CD, agents, etc.).

### Authentication

Include the API key in every request header:

```http
X-API-Key: <your-api-secret>
```

Retrieve your API Secret from the top-right user menu → **API Secret**.

### Endpoint Overview

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/claw/instances` | List all instances |
| `GET` | `/claw/instances/{name}` | Get a single instance by name |
| `POST` | `/claw/alloc` | Allocate an instance (body: `{"user_id": "..."}`) |
| `POST` | `/claw/free` | Free an instance (body: `{"name": "..."}`) |
| `POST` | `/claw/pause` | Pause an instance (body: `{"name": "..."}`) |
| `POST` | `/claw/resume` | Resume an instance (body: `{"name": "..."}`) |
| `GET` | `/claw/token?name={name}` | Get the gateway token for an instance |
| `GET` | `/auth/me` | Get current user info |
| `POST` | `/auth/login` | Login (body: `{"username": "...", "password": "..."}`) |
| `GET` | `/users` | List users (admin only) |
| `POST` | `/users` | Create a user (admin only) |
| `DELETE` | `/users/{id}` | Delete a user (admin only) |

### Example: Allocate an Instance

```bash
curl -X POST http://<host>/claw/alloc \
  -H "X-API-Key: <your-api-secret>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user-12345"}'
```

**Response example:**

```json
{
  "name": "claw-abc123",
  "user_id": "user-12345",
  "claw_webui_url": "https://claw.example.com/claw-abc123/overview",
  "occupied": true,
  "state": "pending",
  "alloc_status": "allocating",
  "token": "my-gateway-token"
}
```

### Example: Free an Instance

```bash
curl -X POST http://<host>/claw/free \
  -H "X-API-Key: <your-api-secret>" \
  -H "Content-Type: application/json" \
  -d '{"name": "claw-abc123"}'
```

### Example: List Idle Instances

```bash
curl "http://<host>/claw/instances?occupied=false" \
  -H "X-API-Key: <your-api-secret>"
```

---

## 8. Typical Workflows

### Scenario 1: Ops manually allocates a session container for a user

```
1. Log in to the console (admin or ops account)
2. Click [Allocate Instance] in the top-right corner
3. Enter the user ID (e.g. session ID), click Allocate
4. Copy the Gateway Token and Claw Web UI link, send to the user
5. After the user session ends → Instances list → Free Instance
```

### Scenario 2: External agent system auto-allocates

```
1. Log in to the console → user menu → API Secret, copy the key
2. Call POST /claw/alloc (with X-API-Key and user_id)
3. Extract instance name, Token, and Web UI link from the response
4. Call POST /claw/free when the session ends to release resources
```

**Code example (Python):**

```python
import requests

BASE_URL = "http://<host>"
API_KEY = "<your-api-secret>"
HEADERS = {"X-API-Key": API_KEY, "Content-Type": "application/json"}

# Allocate an instance
resp = requests.post(f"{BASE_URL}/claw/alloc",
                     json={"user_id": "session-001"},
                     headers=HEADERS)
instance = resp.json()
print(f"Instance: {instance['name']}")
print(f"Web UI:   {instance['claw_webui_url']}")
print(f"Token:    {instance['token']}")

# Free the instance after use
requests.post(f"{BASE_URL}/claw/free",
              json={"name": instance["name"]},
              headers=HEADERS)
```

### Scenario 3: Temporarily suspend resource usage

```
1. Search for the instance by name or user ID in the Instances list
2. Click ⋮ menu → Pause Instance
3. The instance enters paused state — no CPU consumed, binding is retained
4. Click Resume Instance to bring it back when needed
```

> **Recommendation:** In production, automate allocation and release via API to reduce manual work and improve resource utilisation.

---

*For issues, contact your system administrator.*
