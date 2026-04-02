# VERSION defines the project version.
VERSION ?= 0.0.1

# Image URL
IMG ?= claw-swarm-panel:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes.
	$(GOLANGCI_LINT) run --fix

##@ Frontend

.PHONY: install-webui
install-webui: ## Install webui dependencies.
	cd webui && pnpm install

.PHONY: run-webui
run-webui: ## Start webui dev server (proxies /claw/* to localhost:8088).
	cd webui && pnpm dev

.PHONY: build-webui
build-webui: ## Build webui for production (outputs to webui/dist/).
	cd webui && pnpm build

##@ Build

.PHONY: build
build: fmt vet ## Build apiserver binary.
	go build -o bin/apiserver ./cmd/apiserver/

.PHONY: run
run: fmt vet ## Run apiserver from host.
	go run ./cmd/apiserver/

##@ Docker

.PHONY: docker-build
docker-build: ## Build container image (IMG).
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push container image (IMG).
	$(CONTAINER_TOOL) push ${IMG}

##@ Helm

.PHONY: helm-lint
helm-lint: ## Lint helm charts.
	./scripts/helm_lint.sh

.PHONY: helm-package
helm-package: ## Package helm chart.
	helm package charts/claw-swarm-panel


##@ Dependencies

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

GOLANGCI_LINT   = $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.1.1

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef
