#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
mkdir -p .local/logs .local/pids
services=(identity core execution mock gateway)
stop_services() {
  for service in gateway mock execution core identity; do
    file=".local/pids/$service"
    if [[ -f "$file" ]]; then
      pid="$(cat "$file")"
      if [[ "$pid" =~ ^[0-9]+$ ]] && kill -0 "$pid" 2>/dev/null; then
        actual="$(ps -p "$pid" -o command=)"
        if [[ "$actual" == "$ROOT/.local/bin/$service"* ]]; then kill "$pid"; fi
      fi
      rm -f "$file"
    fi
  done
}
case "${1:-start}" in
stop) stop_services; exit 0;;
status) for service in "${services[@]}"; do file=".local/pids/$service"; if [[ -f "$file" ]] && kill -0 "$(cat "$file")" 2>/dev/null; then echo "$service running"; else echo "$service stopped"; fi; done; exit 0;;
start) ;;
*) echo 'Usage: dev.sh start|stop|status' >&2; exit 2;;
esac
[[ -f deploy/local/.env ]] || { echo 'Run make infra-up first' >&2; exit 1; }
for service in "${services[@]}"; do [[ -x ".local/bin/$service" ]] || { echo 'Run make build-all first' >&2; exit 1; }; done
set -a
source deploy/local/.env
export TESTMOON_ROOT="$ROOT" TESTMOON_GATEWAY_HTTP=127.0.0.1:18080 TESTMOON_WEB_ORIGIN=http://localhost:18080 TESTMOON_MOCK_HOST=mock.localhost
export TESTMOON_IDENTITY_HTTP=127.0.0.1:8091 TESTMOON_IDENTITY_GRPC=127.0.0.1:9091
export TESTMOON_CORE_HTTP=127.0.0.1:8092 TESTMOON_CORE_GRPC=127.0.0.1:9092
export TESTMOON_EXECUTION_HTTP=127.0.0.1:8093 TESTMOON_EXECUTION_GRPC=127.0.0.1:9093
export TESTMOON_MOCK_HTTP=127.0.0.1:8094 TESTMOON_MOCK_GRPC=127.0.0.1:9094 TESTMOON_MOCK_PUBLIC_HTTP=127.0.0.1:8095
export TESTMOON_REDIS="127.0.0.1:${REDIS_PORT:-16379}" TESTMOON_REDIS_PASSWORD="$REDIS_PASSWORD" TESTMOON_KAFKA="127.0.0.1:${KAFKA_PORT:-19092}"
export TESTMOON_IDENTITY_DSN="testmoon:$MYSQL_PASSWORD@tcp(127.0.0.1:${MYSQL_PORT:-13306})/testmoon_identity?charset=utf8mb4&parseTime=True&loc=UTC"
export TESTMOON_CORE_DSN="testmoon:$MYSQL_PASSWORD@tcp(127.0.0.1:${MYSQL_PORT:-13306})/testmoon_core?charset=utf8mb4&parseTime=True&loc=UTC"
export TESTMOON_EXECUTION_DSN="testmoon:$MYSQL_PASSWORD@tcp(127.0.0.1:${MYSQL_PORT:-13306})/testmoon_execution?charset=utf8mb4&parseTime=True&loc=UTC"
export TESTMOON_MOCK_DSN="testmoon:$MYSQL_PASSWORD@tcp(127.0.0.1:${MYSQL_PORT:-13306})/testmoon_mock?charset=utf8mb4&parseTime=True&loc=UTC"
set +a
for service in "${services[@]}"; do
  file=".local/pids/$service"
  if [[ -f "$file" ]] && kill -0 "$(cat "$file")" 2>/dev/null; then echo "$service already running"; continue; fi
  nohup "$ROOT/.local/bin/$service" > ".local/logs/$service.log" 2>&1 < /dev/null &
  pid=$!; echo "$pid" > "$file"
  ready=false
  case "$service" in identity) port=8091;;core) port=8092;;execution) port=8093;;mock) port=8094;;gateway) port=18080;;esac
  for attempt in {1..60}; do
    if curl -fsS "http://127.0.0.1:$port/healthz" >/dev/null 2>&1; then ready=true; break; fi
    if ! kill -0 "$pid" 2>/dev/null; then break; fi
    sleep 1
  done
  if ! $ready; then echo "$service failed; see .local/logs/$service.log" >&2; exit 1; fi
  echo "$service ready"
done
echo 'Web: http://localhost:18080 ; Mock: http://mock.localhost:18080'
echo 'Bootstrap credentials: deploy/local/.env (private, ignored by Git)'
