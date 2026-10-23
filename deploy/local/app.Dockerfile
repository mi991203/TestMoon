FROM golang:1.27.3-bookworm@sha256:5c3e9b43cfd3bb95c7937462e29d87ada57b481a96fa23ad0eb61778d7a00110 AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY libs/go libs/go
COPY protocols protocols
COPY backend backend
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -o /out/identity ./backend/platform/identity/cmd/server && \
    CGO_ENABLED=0 go build -trimpath -o /out/core ./backend/testing/core/cmd/server && \
    CGO_ENABLED=0 go build -trimpath -o /out/execution ./backend/testing/execution/cmd/server && \
    CGO_ENABLED=0 go build -trimpath -o /out/mock ./backend/testing/mock/cmd/server && \
    CGO_ENABLED=0 go build -trimpath -o /out/gateway ./backend/platform/gateway/cmd/server

FROM node:24.22.0-bookworm-slim@sha256:a5f4366e834695d5f363b920c104953a535d3a9864d1bdf74566a0ebe01daee5 AS web-build
WORKDIR /src
RUN npm install -g pnpm@11.28.6
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY frontend/portal/package.json frontend/portal/package.json
COPY frontend/testing/package.json frontend/testing/package.json
COPY libs/frontend/api/package.json libs/frontend/api/package.json
COPY libs/frontend/web/package.json libs/frontend/web/package.json
RUN --mount=type=cache,target=/pnpm/store pnpm install --frozen-lockfile --store-dir /pnpm/store
COPY frontend frontend
COPY libs/frontend libs/frontend
RUN pnpm --filter @testmoon/portal --filter @testmoon/testing build

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=go-build /out /app/bin
COPY --from=web-build /src/frontend/portal/dist /app/frontend/portal/dist
COPY --from=web-build /src/frontend/testing/dist /app/frontend/testing/dist
ENV TESTMOON_ROOT=/app
USER 10001:10001
CMD ["/app/bin/gateway"]
