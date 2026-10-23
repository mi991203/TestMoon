#!/usr/bin/env bash
# Sourced after ROOT is set. Existing credentials and encryption keys are never replaced.
ENVFILE="${TESTMOON_ENV_FILE:-$ROOT/deploy/local/.env}"
ensure_local_env() {
  if [[ -f "$ENVFILE" ]]; then return; fi
  for volume in mysql-data redis-data kafka-data; do
    if docker volume inspect "${COMPOSE_PROJECT_NAME:-testmoon}_$volume" >/dev/null 2>&1; then
      echo "Existing data volumes found; restore the original private environment file: $ENVFILE" >&2
      return 1
    fi
  done
  (umask 077; python3 - "$ENVFILE" <<'PYENV'
import sys,secrets,base64,uuid
from pathlib import Path
values={'MYSQL_ROOT_PASSWORD':secrets.token_hex(24),'MYSQL_PASSWORD':secrets.token_hex(24),'REDIS_PASSWORD':secrets.token_hex(24),'MYSQL_PORT':'13306','REDIS_PORT':'16379','KAFKA_PORT':'19092','GATEWAY_PORT':'18080','KAFKA_CLUSTER_ID':base64.urlsafe_b64encode(uuid.uuid4().bytes).decode().rstrip('='),'TESTMOON_SERVICE_KEY':secrets.token_hex(32),'TESTMOON_ENCRYPTION_KEY':secrets.token_hex(32),'TESTMOON_BOOTSTRAP_USERNAME':'admin','TESTMOON_BOOTSTRAP_PASSWORD':secrets.token_urlsafe(20)}
with Path(sys.argv[1]).open('x') as f:
    f.write(''.join(f'{k}={v}\n' for k,v in values.items()))
PYENV
  )
  echo "Created private environment file: $ENVFILE; read bootstrap credentials there."
}
