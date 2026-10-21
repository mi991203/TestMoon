#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
command -v docker >/dev/null || { echo "Docker is required" >&2; exit 1; }
docker compose version >/dev/null
docker info >/dev/null 2>&1 || { echo "Docker daemon unavailable" >&2; exit 1; }
ENVFILE="$ROOT/deploy/local/.env"
if [[ ! -f "$ENVFILE" ]]; then
  (umask 077; python3 - "$ENVFILE" <<'PYENV'
import sys,secrets,base64,uuid
from pathlib import Path
values={'MYSQL_ROOT_PASSWORD':secrets.token_hex(24),'MYSQL_PASSWORD':secrets.token_hex(24),'REDIS_PASSWORD':secrets.token_hex(24),'MYSQL_PORT':'13306','REDIS_PORT':'16379','KAFKA_PORT':'19092','KAFKA_CLUSTER_ID':base64.urlsafe_b64encode(uuid.uuid4().bytes).decode().rstrip('='),'TESTMOON_SERVICE_KEY':secrets.token_hex(32),'TESTMOON_ENCRYPTION_KEY':secrets.token_hex(32),'TESTMOON_BOOTSTRAP_USERNAME':'admin','TESTMOON_BOOTSTRAP_PASSWORD':secrets.token_urlsafe(20)}
with Path(sys.argv[1]).open('x') as f:f.write(''.join(f'{k}={v}\n' for k,v in values.items()))
PYENV
  )
  echo "Created private deploy/local/.env; read bootstrap credentials there."
fi
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
