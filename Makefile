SHELL := /bin/bash
GO ?= go
PNPM ?= pnpm
BIN := .local/bin
APIMOON_BUILD_DIR ?= $(BIN)/apimoon
APIMOON_BUILD_TARGETS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64
APIMOON_BIN_DIR ?= $(HOME)/.local/bin
APIMOON_SKILLS_DIR ?= $(HOME)/.agents/skills
APIMOON_VERSION ?=
APIMOON_RELEASE_DIR ?= .local/apimoon-release/$(APIMOON_VERSION)
SERVICES := identity core execution mock gateway
.PHONY: tools generate deps build-go prepare-apimoon-build build-apimoon build-apimoon-all install-apimoon install-apimoon-skill release-apimoon build-web build-all typecheck infra-up infra-down infra-status dev stop worker-image runner-image
tools:
	mkdir -p .tools/bin
	GOBIN=$(CURDIR)/.tools/bin $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
	GOBIN=$(CURDIR)/.tools/bin $(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
	GOBIN=$(CURDIR)/.tools/bin $(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.32.0
	GOBIN=$(CURDIR)/.tools/bin $(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.32.0
generate:
	./scripts/generate-proto.sh
deps:
	$(GO) mod download
	$(PNPM) install --frozen-lockfile
prepare-apimoon-build:
	@set -eu; \
	if [ -f "$(APIMOON_BUILD_DIR)" ] && [ ! -L "$(APIMOON_BUILD_DIR)" ]; then \
		previous=$$(mktemp "$(APIMOON_BUILD_DIR).XXXXXX"); \
		mv "$(APIMOON_BUILD_DIR)" "$$previous"; \
		mkdir -p "$(APIMOON_BUILD_DIR)"; \
		mv "$$previous" "$(APIMOON_BUILD_DIR)/apimoon"; \
	else \
		mkdir -p "$(APIMOON_BUILD_DIR)"; \
	fi
build-apimoon: prepare-apimoon-build
	$(GO) build -o "$(APIMOON_BUILD_DIR)/apimoon$$($(GO) env GOEXE)" ./tools/apimoon/cmd
build-apimoon-all: prepare-apimoon-build
	@set -eu; \
	for target in $(APIMOON_BUILD_TARGETS); do \
		goos=$${target%/*}; \
		goarch=$${target#*/}; \
		extension=; \
		if [ "$$goos" = windows ]; then extension=.exe; fi; \
		echo "Building APIMoon for $$target"; \
		CGO_ENABLED=0 GOOS="$$goos" GOARCH="$$goarch" $(GO) build -trimpath \
			-o "$(APIMOON_BUILD_DIR)/apimoon_$${goos}_$${goarch}$$extension" ./tools/apimoon/cmd; \
	done
install-apimoon:
	mkdir -p "$(APIMOON_BIN_DIR)"
	$(GO) build -trimpath -o "$(APIMOON_BIN_DIR)/apimoon" ./tools/apimoon/cmd
install-apimoon-skill:
	$(GO) run ./tools/apimoon/cmd skill install --agent codex --dir "$(APIMOON_SKILLS_DIR)"
release-apimoon:
	bash scripts/package-apimoon.sh "$(APIMOON_VERSION)" "$(APIMOON_RELEASE_DIR)"
build-go: build-apimoon
	mkdir -p $(BIN)
	$(GO) build -o $(BIN)/identity ./backend/platform/identity/cmd/server
	$(GO) build -o $(BIN)/core ./backend/testing/core/cmd/server
	$(GO) build -o $(BIN)/execution ./backend/testing/execution/cmd/server
	$(GO) build -o $(BIN)/mock ./backend/testing/mock/cmd/server
	$(GO) build -o $(BIN)/gateway ./backend/platform/gateway/cmd/server
	$(GO) build -o $(BIN)/runner ./backend/testing/runner/cmd/runner
	$(GO) build -o $(BIN)/worker ./backend/testing/runner/cmd/worker
build-web:
	$(PNPM) build
typecheck:
	$(PNPM) typecheck
build-all: build-go build-web
infra-up:
	./scripts/infra.sh up
infra-down:
	./scripts/infra.sh down
infra-status:
	./scripts/infra.sh status
dev:
	./scripts/dev.sh start
stop:
	./scripts/dev.sh stop
worker-image:
	docker build -f deploy/runner/worker.Dockerfile -t testmoon/worker:0.1.0 .
runner-image:
	docker build -f deploy/runner/runner.Dockerfile -t testmoon/runner:0.1.0 .
