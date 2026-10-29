SHELL := /bin/bash
GO ?= go
PNPM ?= pnpm
BIN := .local/bin
SERVICES := identity core execution mock gateway
.PHONY: tools generate deps build-go build-apimoon build-web build-all typecheck infra-up infra-down infra-status dev stop worker-image runner-image
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
build-apimoon:
	mkdir -p $(BIN)
	$(GO) build -o $(BIN)/apimoon ./tools/apimoon/cmd
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
