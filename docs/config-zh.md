# 配置参考

所有运行时配置从单一 YAML 文件读取（默认为工作目录下的 `config.yaml`）。

通过 `--config` 标志指定自定义路径：

```bash
go run ./cmd/main.go --config /etc/claw/config.yaml
```

CLI 标志的优先级始终高于配置文件中的值。

---

## `server`

HTTP 管理 API 设置。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `port` | int | `8088` | 管理 API 监听端口。 |
| `jwt_secret` | string | `""` | 用于签发 JWT Token 的密钥。若为空，启动时自动生成 32 字节随机密钥，并持久化到 `<data_dir>/.jwt_secret`，保证重启后 Token 仍然有效。生产环境建议显式设置。 |
| `data_dir` | string | `.claw-swarm` | 持久化数据目录（SQLite 数据库文件、JWT 密钥文件）。目录不存在时自动创建。 |

**对应 CLI 标志：** `--api-port`

---

## `adapter`

选择并配置用于管理 Claw 实例的后端适配器。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `type` | string | `k8s` | 后端类型。目前仅支持 `k8s`。 |

### `adapter.k8s`

Kubernetes 专用配置（`adapter.type = "k8s"` 时生效）。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `kubeconfig` | string | `""` | kubeconfig 文件路径。解析顺序：(1) 此字段，(2) `KUBECONFIG` 环境变量或 `~/.kube/config`，(3) 集群内 Service Account。 |
| `namespace` | string | `""` | 管理 Claw StatefulSet 所在的 Kubernetes 命名空间。依次回退到 `POD_NAMESPACE` 环境变量，再到 `"default"`。 |

---

## `init`

初始化 Claw 实例时应用的参数。当前包含 LiteLLM 集成配置块。

### `init.litellm` — LiteLLM 集成

LiteLLM 是兼容 OpenAI 接口的 AI 模型代理。配置后，每次分配 Claw 实例时，系统将自动完成以下操作：

1. 为请求方的 `user_id` 在 LiteLLM 中创建（或获取已有）用户
2. 生成一个仅限于配置模型层级的 API Key
3. 通过 `openclaw config set` 将 Key 和模型端点注入 Claw Pod
4. 将 Pod 的默认模型设置为 `tabtab-litellm/<model_id>`

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `baseurl` | string | `http://localhost:4000` | LiteLLM 代理服务的 Base URL。 |
| `master_key` | string | `""` | LiteLLM 管理员主密钥（`Authorization: Bearer <master_key>`）。**必填项**，用于启用集成。若为空，分配时回退到随机 32 字符 API Key。 |
| `default_team` | string | `""` | 分配给所有新用户的 LiteLLM 团队 UUID。留空则跳过团队分配。 |
| `default_max_budget` | float | `0` | 每个生成 Key 的最大消费额度（美元）。`0` 表示不限额。 |
| `default_budget_duration` | string | `""` | 预算重置周期（如 `"30d"`、`"1h"`）。留空表示不自动重置。 |

#### `init.litellm.models` — 模型层级映射

将分配层级映射到 LiteLLM 模型 ID。调用方通过 `POST /claw/alloc` 的 `model_type` 字段选择层级；系统在创建 API Key 和配置 Pod 时使用对应的模型 ID。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `models.lite.modelID` | string | `tabtab-lite` | **lite** 层级（默认层级）对应的 LiteLLM 模型 ID。 |
| `models.pro.modelID` | string | `tabtab-pro` | **pro** 层级对应的 LiteLLM 模型 ID。 |

#### 模型配置注入方式

Claw Pod 进入 `running` 状态后，Operator 通过 Kubernetes exec API 在容器内执行两条命令：

```
# 1. 注册 LiteLLM Provider（携带生成的 API Key）
openclaw config set models.providers.tabtab-litellm \
  '{"baseUrl":"<litellm_baseurl>","models":[{"id":"<model_id>","name":"<model_id>","input":["text"],"contextWindow":200000}],"apiKey":"<generated_key>","auth":"api-key","api":"openai-completions"}'

# 2. 设置 Pod 默认模型
openclaw config set agents.defaults.model.primary "tabtab-litellm/<model_id>"
```

这意味着每个实例完全自包含：它持有自己的 LiteLLM 凭证，并在无需额外配置的情况下知道默认使用哪个模型。

#### 分配流程（含 LiteLLM）

```
POST /claw/alloc  { user_id, model_type }
        │
        ├─ model_type 默认为 "lite"（若未传入）
        │
        ├─ resolveLiteLLMKey()
        │     ├─ 已配置 master_key?
        │     │     是 → POST {baseurl}/user/new  （新用户）
        │     │          或 POST {baseurl}/key/generate  （已有用户）
        │     │          → 返回有范围限制的 API Key
        │     │     否 → 生成随机 32 字符回退 Key
        │     └─ 返回 apiKey
        │
        ├─ 占用空闲 StatefulSet，设置 replicas=1，标签=allocating
        ├─ 立即返回响应（alloc_status: "allocating"）
        │
        └─ [异步] configurePodModels(apiKey, model_type)
              ├─ Exec: openclaw config set models.providers.tabtab-litellm {...}
              ├─ Exec: openclaw config set agents.defaults.model.primary tabtab-litellm/<modelID>
              └─ 更新标签 → allocated
```

请轮询 `GET /claw/instances`，待 `alloc_status` 变为 `"allocated"` 后，再向实例发送模型请求。

#### 优雅降级

若 LiteLLM 不可达或 `master_key` 为空，分配流程**不会被阻塞**：系统使用随机 32 字符字符串作为 API Key。Pod 仍会完成配置，但 AI 模型请求可能失败（除非实例预先配置了其他模型 Provider）。

#### 最小化 LiteLLM 配置示例

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

#### 完整 LiteLLM 配置示例

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

SQLite 数据库配置。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `file` | string | `""` | SQLite 数据库文件路径。为空时默认使用 `<data_dir>/claw.db`。 |

---

## 完整配置示例

```yaml
# ── API 服务器 ────────────────────────────────────────────────────────────────
server:
  port: 8088
  jwt_secret: "生产环境请替换此值"
  data_dir: ".claw-swarm"

# ── 适配器后端 ────────────────────────────────────────────────────────────────
adapter:
  type: "k8s"
  k8s:
    kubeconfig: ""        # 空 → 集群内或 ~/.kube/config
    namespace: "default"

# ── 实例初始化 ────────────────────────────────────────────────────────────────
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

# ── 数据库 ────────────────────────────────────────────────────────────────────
db:
  file: ""                # 空 → <data_dir>/claw.db
```
