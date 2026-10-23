#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
command -v docker >/dev/null || { echo "Docker is required" >&2; exit 1; }
docker compose version >/dev/null
docker info >/dev/null 2>&1 || { echo "Docker daemon unavailable" >&2; exit 1; }
source "$ROOT/scripts/lib/local-env.sh"
ensure_local_env
set -a; source "$ENVFILE"; set +a
compose=(docker compose --env-file "$ENVFILE" -f "$ROOT/deploy/local/compose.yaml")
action="${1:-status}"; if [[ $# -gt 0 ]]; then shift; fi
case "$action" in
up)
  if [[ $# -eq 0 ]]; then set -- mysql redis kafka; fi
  mysql_selected=false
  for service in "$@"; do case "$service" in mysql) mysql_selected=true;;redis|kafka);;*) echo "Unknown service: $service" >&2;exit 2;;esac;done
  "${compose[@]}" up -d --wait --wait-timeout 240 "$@"
  if $mysql_selected; then
    while IFS= read -r database; do
      [[ "$database" =~ ^[a-z_]+$ ]] || { echo "Invalid database name" >&2; exit 2; }
      "${compose[@]}" exec -T -e "MYSQL_PWD=$MYSQL_ROOT_PASSWORD" mysql mysql -uroot -e "CREATE DATABASE IF NOT EXISTS $database CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; GRANT ALL ON $database.* TO 'testmoon'@'%';" < /dev/null
    done < "$ROOT/deploy/local/mysql/databases.txt"
  fi;;
down) "${compose[@]}" down;;
logs) "${compose[@]}" logs --tail 100 "$@";;
status) "${compose[@]}" ps;;
*) echo "Usage: infra.sh up [mysql redis kafka] | down | logs | status" >&2;exit 2;;
esac
