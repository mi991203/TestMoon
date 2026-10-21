FROM golang:1.27.3-bookworm@sha256:5c3e9b43cfd3bb95c7937462e29d87ada57b481a96fa23ad0eb61778d7a00110 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY libs/go libs/go
COPY protocols protocols
COPY backend/testing/runner backend/testing/runner
RUN CGO_ENABLED=0 go build -trimpath -o /runner ./backend/testing/runner/cmd/runner
FROM node:24.22.0-bookworm-slim@sha256:a5f4366e834695d5f363b920c104953a535d3a9864d1bdf74566a0ebe01daee5
COPY --from=build /runner /app/runner
ENTRYPOINT ["/app/runner"]
