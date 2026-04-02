# Configuration Reference

All runtime settings are read from a single YAML file (default: `config.yaml` in the working directory).

Pass a custom path with the `--config` flag:

```bash
go run ./cmd/main.go --config /etc/claw/config.yaml
```

CLI flags always take precedence over file values.

---

## `server`

HTTP management API settings.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | int | `8088` | Port the management API listens on. |
| `jwt_secret` | string | `""` | Secret used to sign JWT tokens. If empty, a random 32-byte secret is generated on startup and persisted to `<data_dir>/.jwt_secret` so tokens survive restarts. Set this explicitly in production for predictable behaviour. |
| `data_dir` | string | `.claw-swarm` | Directory for persistent data (SQLite database file, JWT secret file). Created automatically if absent. |

**CLI flag equivalents:** `--api-port`

---

## `adapter`

Selects and configures the backend adapter used to manage claw instances.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | `k8s` | Backend type. Currently only `k8s` is supported. |

### `adapter.k8s`

Kubernetes-specific settings (used when `adapter.type = "k8s"`).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kubeconfig` | string | `""` | Path to a kubeconfig file. Resolution order: (1) this field, (2) `KUBECONFIG` env var or `~/.kube/config`, (3) in-cluster service account. |
| `namespace` | string | `""` | Kubernetes namespace where claw StatefulSets are managed. Falls back to the `POD_NAMESPACE` environment variable, then `"default"`. |

---

## `init`

Parameters applied when initialising claw instances. Currently contains the LiteLLM integration block.

### `init.litellm` — LiteLLM Integration

LiteLLM is an OpenAI-compatible AI model proxy. When configured, every claw instance allocation automatically:

1. Creates (or retrieves) a LiteLLM user for the requesting `user_id`
2. Generates a scoped API key limited to the configured model tier
3. Injects the key and model endpoint into the claw pod via `openclaw config set`
4. Sets the pod's default model to `tabtab-litellm/<model_id>`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `baseurl` | string | `http://localhost:4000` | Base URL of the LiteLLM proxy service. |
| `master_key` | string | `""` | LiteLLM admin master key (`Authorization: Bearer <master_key>`). **Required** to enable the integration. If empty, allocation falls back to a random 32-character API key. |
| `default_team` | string | `""` | LiteLLM team UUID assigned to all new users. Leave empty to skip team assignment. |
| `default_max_budget` | float | `0` | Maximum spend (USD) applied to each generated key. `0` means unlimited. |
| `default_budget_duration` | string | `""` | Budget reset period (e.g. `"30d"`, `"1h"`). Empty means no automatic reset. |

#### `init.litellm.models`

Maps allocation tiers to LiteLLM model IDs. The caller selects a tier via the `model_type` field in `POST /claw/alloc`; the system uses the corresponding model ID when creating the API key and configuring the pod.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `models.lite.modelID` | string | `tabtab-lite` | LiteLLM model ID for the **lite** tier (default tier). |
| `models.pro.modelID` | string | `tabtab-pro` | LiteLLM model ID for the **pro** tier. |

#### How model configuration is injected

After a claw pod becomes `running`, the operator executes two commands inside the container via the Kubernetes exec API:

```
# 1. Register the LiteLLM provider with the generated API key
openclaw config set models.providers.tabtab-litellm \
  '{"baseUrl":"<litellm_baseurl>","models":[{"id":"<model_id>","name":"<model_id>","input":["text"],"contextWindow":200000}],"apiKey":"<generated_key>","auth":"api-key","api":"openai-completions"}'

# 2. Set the pod's default model
openclaw config set agents.defaults.model.primary "tabtab-litellm/<model_id>"
```

This means the instance is fully self-contained: it holds its own LiteLLM credentials and knows which model to use by default without any further configuration.

#### Allocation flow with LiteLLM

```
POST /claw/alloc  { user_id, model_type }
        │
        ├─ model_type defaults to "lite" if omitted
        │
        ├─ resolveLiteLLMKey()
        │     ├─ master_key configured?
        │     │     YES → POST {baseurl}/user/new  (new user)
        │     │            or POST {baseurl}/key/generate  (existing user)
        │     │            → returns scoped API key
        │     │     NO  → generate random 32-char fallback key
        │     └─ return apiKey
        │
        ├─ Claim idle StatefulSet, set replicas=1, label=allocating
        ├─ Return response immediately  (alloc_status: "allocating")
        │
        └─ [async] configurePodModels(apiKey, model_type)
              ├─ Exec: openclaw config set models.providers.tabtab-litellm {...}
              ├─ Exec: openclaw config set agents.defaults.model.primary tabtab-litellm/<modelID>
              └─ Update label → allocated
```

Poll `GET /claw/instances` until `alloc_status` becomes `"allocated"` before sending model requests to the instance.

#### Graceful degradation

If LiteLLM is unreachable or `master_key` is empty, allocation is **never blocked**: a random 32-character string is used as the API key. The pod will still be configured, but AI model requests will likely fail unless the instance has an alternative model provider pre-configured.

#### Minimal LiteLLM configuration example

```yaml
init:
  litellm:
    baseurl: "http://litellm.example.com"
    master_key: "sk-my-master-key"
    models:
      lite:
        modelID: "my-lite-model"
      pro:
        modelID: "my-pro-model"
```

#### Full LiteLLM configuration example

```yaml
init:
  litellm:
    baseurl: "http://litellm.example.com"
    master_key: "sk-my-master-key"
    default_team: "team-uuid-1234"
    default_max_budget: 10.0
    default_budget_duration: "30d"
    models:
      lite:
        modelID: "my-lite-model"
      pro:
        modelID: "my-pro-model"
```

---

## `db`

SQLite database settings.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `file` | string | `""` | Path to the SQLite database file. If empty, defaults to `<data_dir>/claw.db`. |

---

## Complete example

```yaml
# ── API Server ────────────────────────────────────────────────────────────────
server:
  port: 8088
  jwt_secret: "change-me-in-production"
  data_dir: ".claw-swarm"

# ── Adapter backend ───────────────────────────────────────────────────────────
adapter:
  type: "k8s"
  k8s:
    kubeconfig: ""      # empty → in-cluster or ~/.kube/config
    namespace: "default"

# ── Instance initialisation ───────────────────────────────────────────────────
init:
  litellm:
    baseurl: "http://litellm.example.com"
    master_key: "sk-my-master-key"
    default_team: "team-uuid-1234"
    default_max_budget: 10.0
    default_budget_duration: "30d"
    models:
      lite:
        modelID: "my-lite-model"
      pro:
        modelID: "my-pro-model"

# ── Database ──────────────────────────────────────────────────────────────────
db:
  file: ""              # empty → <data_dir>/claw.db
```
