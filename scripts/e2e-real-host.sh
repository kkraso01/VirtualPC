#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VPC_BIN="${VPC_BIN:-$ROOT_DIR/bin/vpc}"
VPCD_BIN="${VPCD_BIN:-$ROOT_DIR/bin/virtualpcd}"
SOCK="${VPC_UNIX_SOCKET:-/tmp/virtualpcd.sock}"
RUN_DIR="${VPC_E2E_RUN_DIR:-$ROOT_DIR/.tmp/e2e-real-host}"
LOG_DIR="$RUN_DIR/logs"
ARTIFACT_DIR="$RUN_DIR/artifacts"
mkdir -p "$LOG_DIR" "$ARTIFACT_DIR"

pass_count=0
fail_count=0
warn_count=0

json_get() {
  local json_input="$1"
  local expr="$2"
  python3 - "$expr" "$json_input" <<'PY'
import json,sys
expr=sys.argv[1]
obj=json.loads(sys.argv[2])
for part in expr.split('.'):
    if not part:
        continue
    if isinstance(obj, list):
        obj=obj[int(part)]
    else:
        obj=obj.get(part)
print(obj if obj is not None else "")
PY
}

run_step() {
  local name="$1"
  shift
  if "$@"; then
    echo "[PASS] $name"
    pass_count=$((pass_count+1))
  else
    echo "[FAIL] $name"
    fail_count=$((fail_count+1))
    return 1
  fi
}

warn_step() {
  local name="$1"
  local msg="$2"
  echo "[WARN] $name: $msg"
  warn_count=$((warn_count+1))
}

cleanup() {
  if [[ -n "${DAEMON_PID:-}" ]] && kill -0 "$DAEMON_PID" 2>/dev/null; then
    kill "$DAEMON_PID" || true
    wait "$DAEMON_PID" || true
  fi
  rm -f "$SOCK"
}
trap cleanup EXIT

export VPC_UNIX_SOCKET="$SOCK"
export VPCD_UNIX_SOCKET="$SOCK"
export VPCD_DATA_PATH="${VPCD_DATA_PATH:-$RUN_DIR/state.json}"
export VPCD_FIRECRACKER_DIR="${VPCD_FIRECRACKER_DIR:-$RUN_DIR/firecracker}"

if [[ ! -x "$VPC_BIN" || ! -x "$VPCD_BIN" ]]; then
  echo "missing binaries in ./bin; run ./scripts/build-binaries.sh first" >&2
  exit 1
fi

doctor_json="$($VPC_BIN doctor)"
printf '%s
' "$doctor_json" > "$LOG_DIR/doctor.json"
if [[ "$(json_get "$doctor_json" 'healthy')" != "True" ]]; then
  cat "$LOG_DIR/doctor.json"
  echo "doctor failed; refusing to continue"
  exit 1
fi

echo "starting daemon"
"$VPCD_BIN" > "$LOG_DIR/daemon.log" 2>&1 &
DAEMON_PID=$!
for _ in $(seq 1 40); do
  if "$VPC_BIN" daemon status > "$LOG_DIR/status.json" 2>/dev/null; then
    break
  fi
  sleep 0.25
done

machine_json="$($VPC_BIN machine create --profile minimal-shell)"
MACHINE_ID="$(json_get "$machine_json" 'id')"

run_step "machine create" test -n "$MACHINE_ID"
run_step "machine start" bash -lc "$VPC_BIN machine start $MACHINE_ID >/dev/null"
run_step "guest-agent handshake (exec path)" bash -lc "$VPC_BIN machine exec $MACHINE_ID -- /bin/echo handshake-ok | tee $LOG_DIR/exec.json >/dev/null"
run_step "machine shell" bash -lc "$VPC_BIN machine shell $MACHINE_ID >/dev/null"

UPLOAD_FILE="$ARTIFACT_DIR/upload.txt"
DOWNLOAD_FILE="$ARTIFACT_DIR/download.txt"
echo "hello-from-host" > "$UPLOAD_FILE"
run_step "file upload" bash -lc "$VPC_BIN machine cp-to $MACHINE_ID $UPLOAD_FILE /tmp/upload.txt >/dev/null"
run_step "file download" bash -lc "$VPC_BIN machine cp-from $MACHINE_ID /tmp/upload.txt $DOWNLOAD_FILE >/dev/null"
run_step "file transfer integrity" grep -q "hello-from-host" "$DOWNLOAD_FILE"

svc_json="$($VPC_BIN service create --machine "$MACHINE_ID" --name e2e-svc --image alpine:latest)"
SVC_ID="$(json_get "$svc_json" 'id')"
run_step "service create" test -n "$SVC_ID"
run_step "service list" bash -lc "$VPC_BIN service list --machine $MACHINE_ID >/dev/null"
run_step "service logs" bash -lc "$VPC_BIN service logs $SVC_ID >/dev/null"
run_step "service stop" bash -lc "$VPC_BIN service stop $SVC_ID >/dev/null"
run_step "service destroy" bash -lc "$VPC_BIN service destroy $SVC_ID >/dev/null"

snap_json="$($VPC_BIN snapshot create "$MACHINE_ID")"
SNAP_ID="$(json_get "$snap_json" 'id')"
run_step "snapshot create" test -n "$SNAP_ID"
fork_json="$($VPC_BIN machine fork "$SNAP_ID")"
FORK_ID="$(json_get "$fork_json" 'id')"
run_step "fork from snapshot" test -n "$FORK_ID"

task_create_json="$($VPC_BIN task create --machine "$MACHINE_ID" --goal 'echo task-ok')"
TASK_ID="$(json_get "$task_create_json" 'id')"
run_step "task create/run" bash -lc "test -n '$TASK_ID' && $VPC_BIN task run $TASK_ID >/dev/null"

echo "restarting daemon for reconciliation check"
kill "$DAEMON_PID" && wait "$DAEMON_PID" || true
"$VPCD_BIN" > "$LOG_DIR/daemon-restart.log" 2>&1 &
DAEMON_PID=$!
for _ in $(seq 1 40); do
  if "$VPC_BIN" daemon status >/dev/null 2>&1; then
    break
  fi
  sleep 0.25
done
run_step "daemon restart reconciliation" bash -lc "$VPC_BIN machine inspect $MACHINE_ID >/dev/null"

if "$VPC_BIN" profile list | grep -q allowlist; then
  run_step "network policy profile visibility" true
else
  warn_step "network policy enforcement" "allowlist/offline policy endpoints are not fully exposed in CLI flow"
fi

echo "--- e2e summary ---"
echo "pass=$pass_count fail=$fail_count warn=$warn_count"
(( fail_count == 0 ))
