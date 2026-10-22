FROM golang:1.27.3-bookworm@sha256:5c3e9b43cfd3bb95c7937462e29d87ada57b481a96fa23ad0eb61778d7a00110 AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY libs/go libs/go
COPY protocols protocols
COPY backend/testing/runner backend/testing/runner
RUN CGO_ENABLED=0 go build -trimpath -o /worker ./backend/testing/runner/cmd/worker
FROM node:24.22.0-bookworm-slim@sha256:a5f4366e834695d5f363b920c104953a535d3a9864d1bdf74566a0ebe01daee5 AS js-build
WORKDIR /src
RUN npm install -g pnpm@11.28.6
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY backend/testing/runner/worker/package.json backend/testing/runner/worker/package.json
RUN pnpm install --filter @testmoon/script-worker --frozen-lockfile --ignore-scripts
COPY backend/testing/runner/worker/src backend/testing/runner/worker/src
RUN pnpm --filter @testmoon/script-worker deploy --legacy --prod /out
FROM node:24.22.0-bookworm-slim@sha256:a5f4366e834695d5f363b920c104953a535d3a9864d1bdf74566a0ebe01daee5
WORKDIR /app/worker
COPY --from=go-build /worker /app/worker-go
COPY --from=js-build /out/node_modules ./node_modules
COPY --from=js-build /out/src/worker.mjs ./worker.mjs
USER 10001:10001
ENTRYPOINT ["/app/worker-go"]
