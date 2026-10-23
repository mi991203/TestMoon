#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
for tool in docker go python3; do
  command -v "$tool" >/dev/null || { echo "$tool is required" >&2; exit 1; }
done
docker compose version >/dev/null
docker info >/dev/null 2>&1 || { echo 'Docker daemon unavailable' >&2; exit 1; }
source "$ROOT/scripts/lib/local-env.sh"
ensure_local_env
docker compose --env-file "$ENVFILE" -f "$ROOT/deploy/local/compose.yaml" config --quiet
LOCAL_DIR="${TESTMOON_LOCAL_DIR:-$ROOT/.local}"
mkdir -p "$LOCAL_DIR/bin"
go build -o "$LOCAL_DIR/bin/runner" ./backend/testing/runner/cmd/runner
docker build -f deploy/runner/worker.Dockerfile -t "${TESTMOON_WORKER_IMAGE:-testmoon/worker:0.1.0}" .
echo 'Ready. Start the platform with docker compose, then scripts/runner.sh start.'
