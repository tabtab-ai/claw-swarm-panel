# ── Stage 1: Build Go apiserver binary ────────────────────────────────────────
FROM golang:1.24.0-alpine3.21 AS go-builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN apk add --no-cache gcc musl-dev \
    && GOPROXY=https://goproxy.cn,direct go mod download

COPY cmd/ cmd/
COPY pkg/ pkg/

RUN CGO_ENABLED=1 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -a -o apiserver ./cmd/apiserver/

# ── Stage 2: Build Web UI ──────────────────────────────────────────────────────
FROM node:24-alpine AS webui-builder

WORKDIR /webui
COPY webui/package.json ./
RUN npm install --global pnpm && pnpm install

COPY webui/ .
RUN pnpm build

# ── Stage 3: Final image ───────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache nginx

WORKDIR /

COPY --from=go-builder /workspace/apiserver .
COPY nginx.conf /etc/nginx/http.d/default.conf
COPY --from=webui-builder /webui/dist /usr/share/nginx/html
COPY scripts/start.sh /start.sh
RUN chmod +x /start.sh

EXPOSE 80 8088

ENTRYPOINT ["/start.sh"]
