# AGENTS.md — Claw Swarm Operator

This file provides guidance for AI coding agents working in this repository.

---

## Project Overview

**claw-swarm-operator** is a Kubernetes operator that manages a pool of [OpenClaw](https://hub.docker.com/r/alpine/openclaw) containers as on-demand AI coding environments. Each container runs the OpenClaw gateway on port **18789** and is assigned to a conversation via a simple HTTP API.

Key facts:
- Language: Go 1.22
- Framework: [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) (Kubebuilder)
- Runtime image: `alpine/openclaw:latest`
- Module path: `gitlab.botnow.cn/agentic/claw-swarm-operator`

## LiteLLM Integration

The operator integrates with LiteLLM to automatically create users and generate API keys when allocating OpenClaw instances:

- **Endpoint**: `POST /claw/alloc` automatically calls LiteLLM's `/user/new` API
- **Configuration**: Set `init.litellm.master_key` in `config.yaml` with your LiteLLM master key
- **Default Team**: Set `init.litellm.default_team` to assign new users to a specific team (optional)
- **API Key Generation**: The generated API key is returned in the response and configured in the OpenClaw instance
- **Models**: Based on `model_type` ("lite" or "pro"), the corresponding model from config is assigned to the user

See `pkg/litellm/client.go` for the LiteLLM client implementation.

---

## Repository Layout

```
cmd/main.go                        # Binary entry point — flags, manager setup
internal/controller/
  claw_controller.go               # Main reconciler — pool management, image sync
  pvc_controller.go                # PVC finalizer + Redis volume tracking
pkg/
  claw/
    const.go                       # Labels, annotation keys, constants
    http_server.go                 # HTTP management API (RunServer + handlers)
  config/config.go                 # Config struct read from config.yaml
  utils/utils.go                   # Random name generator
templates/
  claw_statefulset.yaml            # Go template for OpenClaw StatefulSet
  claw_service.yaml                # Go template for ClusterIP Service
  claw_ingress.yaml                # Go template for Kong Ingress
  warnup.job.yaml                  # Go template for ACK warm-up Job
charts/claw-swarm-panel/        # Helm chart
docs/
  api.md                           # HTTP API reference
  design-openclaw-migration.md     # Migration design doc
```

---

## Building

```bash
go build ./...          # compile
go test ./...           # run tests
go vet ./...            # static analysis
```

The binary is `/manager`. It is built inside the Docker image — see `Dockerfile`.

---

## Key Concepts

### Instance Pool
The controller maintains a configurable pool (default 5) of idle OpenClaw StatefulSets. When a conversation requests an instance via `/claw/alloc`, one idle StatefulSet is claimed by adding the `tabtab.app.conversation` label.

### Labels
| Label | Purpose |
|---|---|
| `tabtab.statefulset.app` | Marks resources managed by this operator |
| `tabtab.app.conversation=<id>` | Links a StatefulSet to a conversation |
| `tabtab.app.deployed` | Blocks pause while a service is deployed inside |
| `tabtab.app.starter` | Dummy StatefulSet used to trigger the first reconcile |

### Annotations
| Annotation | Purpose |
|---|---|
| `tabtab.app.scheduled.deletion.time` | RFC3339 time when the instance should be paused |
| `tabtab.app.scheduled.deletion` | Trigger added to force a requeue |

### Templates
StatefulSets, Services, and Ingresses are created from Go text/template files under `./templates/`. Template variables are defined in `addFromTemplate()` in `claw_controller.go`.

### PVC Naming
PVC names follow the pattern: `<volumeClaimTemplate.name>-<statefulset-name>-0`
The VolumeClaimTemplate is named `openclaw-config`, mounted at `/home/node/.openclaw`.

---

## Configuration

The operator reads `config.yaml` at startup (path set via `--extra-config` flag). Key fields:

```yaml
resources:
  limits: { cpu: "1", memory: 2Gi }
  requests: { cpu: 250m, memory: 512Mi }
imagePullSecrets:
  - name: image-pull-secret
ack:
  enabled: false          # Alibaba ACK-specific features
redis:
  addr: "localhost:6379"  # Used for PV name tracking
```

---

## HTTP API

The operator runs an HTTP server on `--claw-port` (default **18789**).
Full reference: [`docs/api.md`](docs/api.md)

Quick summary:

| Method | Path | Description |
|--------|------|-------------|
| POST | `/claw/alloc` | Assign an OpenClaw instance to a conversation |
| POST | `/claw/free` | Delete the instance for a conversation |
| POST | `/claw/pause` | Schedule instance suspension |
| POST | `/claw/resume` | Resume a paused instance |
| POST | `/claw/deploy` | Mark instance as deployed (blocks pause) |
| POST | `/claw/undeploy` | Remove deployed mark |
| GET  | `/claw/used` | List all allocated instances |

All endpoints accept `Content-Type: application/json`.

---

## Development Guidelines for Agents

1. **Do not change Kubernetes label keys** (`tabtab.*`). They are stable infrastructure identifiers used by label selectors in production.
2. **Template files** (`templates/*.yaml`) use Go `text/template` syntax with map keys, not struct fields.
3. **Container name** in templates and controller must stay `openclaw` — it is used to identify the container for image reconciliation.
4. **PVC volume claim template name** must stay `openclaw-config` — the controller and http_server derive the PVC name from it.
5. **Do not add commands** to the StatefulSet container spec. OpenClaw uses its built-in entrypoint.
6. When adding a new API endpoint, register it in `ServeHTTP` in `pkg/claw/http_server.go` and document it in `docs/api.md`.
7. Run `go build ./...` and `go vet ./...` before committing.
