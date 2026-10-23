#!/usr/bin/env bash
set -euo pipefail
umask 077
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOCAL_DIR="${TESTMOON_LOCAL_DIR:-$ROOT/.local}"
mkdir -p "$LOCAL_DIR/pids" "$LOCAL_DIR/logs"
LOCAL_DIR="$(cd "$LOCAL_DIR" && pwd)"
binary="$LOCAL_DIR/bin/runner"
pidfile="$LOCAL_DIR/pids/runner"
logfile="$LOCAL_DIR/logs/runner.log"
action="${1:-status}"
case "$action" in start|stop|status);; *) echo 'Usage: runner.sh start|stop|status' >&2; exit 2;; esac

pid_matches() {
  [[ "${1:-}" =~ ^[1-9][0-9]*$ ]] || return 1
  kill -0 "$1" 2>/dev/null || return 1
  local actual
  actual="$(ps -p "$1" -o command= 2>/dev/null)" || return 1
  [[ "$actual" == "$binary" || "$actual" == "$binary "* ]]
}
pid=""
if [[ -f "$pidfile" ]]; then read -r pid < "$pidfile" || true; fi
if [[ "$action" == status ]]; then
  if pid_matches "$pid"; then echo "Runner running (PID $pid)"; else echo 'Runner stopped'; exit 1; fi
  exit 0
fi

# Serialize start/stop so two starts cannot launch the same Runner state.
lock="$LOCAL_DIR/pids/runner.lock"
mkdir "$lock" 2>/dev/null || { echo "Another Runner command is active; if it crashed, remove $lock" >&2; exit 1; }
trap 'rmdir "$lock"' EXIT
# Refresh after acquiring the lock.
pid=""
if [[ -f "$pidfile" ]]; then read -r pid < "$pidfile" || true; fi
if [[ "$action" == stop ]]; then
  if ! pid_matches "$pid"; then rm -f "$pidfile"; echo 'Runner already stopped'; exit 0; fi
  kill -TERM "$pid"
  for ((attempt=0; attempt<30; attempt++)); do
    if ! pid_matches "$pid"; then rm -f "$pidfile"; echo 'Runner stopped'; exit 0; fi
    sleep 1
  done
  echo "Runner still stopping (PID $pid); see $logfile" >&2
  exit 1
fi
if pid_matches "$pid"; then echo "Runner already running (PID $pid)"; exit 0; fi
processes="$(ps -axo pid=,command=)" || { echo 'Cannot check existing Runner processes' >&2; exit 1; }
while read -r other_pid command; do
  if [[ "$command" == "$binary" || "$command" == "$binary "* ]]; then
    echo "Runner already running outside this script (PID $other_pid); stop that process first." >&2
    exit 1
  fi
done <<< "$processes"
[[ -x "$binary" ]] || { echo 'Run scripts/local-prepare.sh first' >&2; exit 1; }
source "$ROOT/scripts/lib/local-env.sh"
[[ -f "$ENVFILE" ]] || { echo 'Run scripts/local-prepare.sh first' >&2; exit 1; }
# Read the published port in a subshell; platform/database secrets stay out of
# the Runner process environment. An explicit shell port takes precedence.
file_port="$(source "$ENVFILE"; printf '%s' "${GATEWAY_PORT:-18080}")"
export TESTMOON_URL="${TESTMOON_URL:-http://localhost:${GATEWAY_PORT:-$file_port}}"
TESTMOON_URL="${TESTMOON_URL%/}"
export TESTMOON_RUNNER_STATE="${TESTMOON_RUNNER_STATE:-$LOCAL_DIR/runner.json}"
export TESTMOON_WORKER_IMAGE="${TESTMOON_WORKER_IMAGE:-testmoon/worker:0.1.0}"
case "$TESTMOON_URL" in http://localhost:*|http://127.0.0.1:*) export TESTMOON_ALLOW_HTTP=1;; esac
if [[ -f "$TESTMOON_RUNNER_STATE" ]]; then
  python3 - "$TESTMOON_RUNNER_STATE" "$TESTMOON_URL" <<'PY'
import json, sys
try:
    state = json.load(open(sys.argv[1]))
    if not state.get('credential') or state.get('url') != sys.argv[2]:
        raise ValueError()
except (OSError, ValueError, TypeError, AttributeError):
    sys.exit('Invalid Runner state or origin mismatch; keep the original TESTMOON_URL or use another state file.')
PY
else
  [[ -n "${TESTMOON_ENROLLMENT_TOKEN:-}" ]] || { echo 'First start requires TESTMOON_ENROLLMENT_TOKEN from Runner management.' >&2; exit 1; }
fi
if [[ -z "${DOCKER_HOST:-}" ]]; then
  export DOCKER_HOST="$(docker context inspect --format '{{.Endpoints.docker.Host}}')"
fi
docker info >/dev/null 2>&1 || { echo 'Docker daemon unavailable' >&2; exit 1; }
docker image inspect "$TESTMOON_WORKER_IMAGE" >/dev/null 2>&1 || { echo 'Worker image missing; run scripts/local-prepare.sh' >&2; exit 1; }
nohup "$binary" >> "$logfile" 2>&1 < /dev/null &
pid=$!
printf '%s\n' "$pid" > "$pidfile"
sleep 2
if ! pid_matches "$pid"; then rm -f "$pidfile"; echo "Runner failed to start; see $logfile" >&2; exit 1; fi
echo "Runner started (PID $pid); log: $logfile"
