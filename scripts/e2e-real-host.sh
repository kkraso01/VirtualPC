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
  local name="$1"; shift
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
  echo "[WARN] $1: $2"
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

[[ -x "$VPC_BIN" && -x "$VPCD_BIN" ]] || { echo "missing binaries in ./bin" >&2; exit 1; }

"$VPCD_BIN" > "$LOG_DIR/daemon.log" 2>&1 &
DAEMON_PID=$!
for _ in $(seq 1 60); do
  "$VPC_BIN" daemon status > /dev/null 2>&1 && break
  sleep 0.2
done

probe_host() {
  local machine="$1"
  local host="$2"
  "$VPC_BIN" machine exec "$machine" -- /bin/sh -lc "ping -c1 -W2 $host >/dev/null"
}

create_start_machine() {
  local profile="$1"
  local out id
  out="$("$VPC_BIN" machine create --profile "$profile")"
  id="$(json_get "$out" id)"
  [[ -n "$id" ]] || return 1
  "$VPC_BIN" machine start "$id" >/dev/null
  echo "$id"
}

NAT_ID="$(create_start_machine minimal-shell)"
OFFLINE_ID="$(create_start_machine minimal-shell-offline)"
ALLOW_ID="$(create_start_machine minimal-shell-allowlist)"
run_step "machines started for nat/offline/allowlist" test -n "$NAT_ID"

run_step "network nat allows egress" probe_host "$NAT_ID" "1.1.1.1"
run_step "network offline blocks egress" bash -lc "! $VPC_BIN machine exec $OFFLINE_ID -- /bin/sh -lc "ping -c1 -W2 1.1.1.1 >/dev/null""
run_step "network allowlist allows listed host" probe_host "$ALLOW_ID" "1.1.1.1"
run_step "network allowlist blocks non-listed host" bash -lc "! $VPC_BIN machine exec $ALLOW_ID -- /bin/sh -lc "ping -c1 -W2 9.9.9.9 >/dev/null""

PARENT_ID="$NAT_ID"
run_step "parent write pre-snapshot marker" bash -lc "$VPC_BIN machine exec $PARENT_ID -- /bin/sh -lc 'echo parent-v1 > /tmp/marker.txt' >/dev/null"
SNAP_JSON="$("$VPC_BIN" snapshot create "$PARENT_ID")"
SNAP_ID="$(json_get "$SNAP_JSON" id)"
run_step "snapshot create" test -n "$SNAP_ID"
FORK_JSON="$("$VPC_BIN" machine fork "$SNAP_ID")"
FORK_ID="$(json_get "$FORK_JSON" id)"
run_step "fork create" test -n "$FORK_ID"
run_step "fork start" bash -lc "$VPC_BIN machine start $FORK_ID >/dev/null"
run_step "fork sees parent snapshot state" bash -lc "$VPC_BIN machine exec $FORK_ID -- /bin/cat /tmp/marker.txt | grep -q parent-v1"
run_step "fork mutates state" bash -lc "$VPC_BIN machine exec $FORK_ID -- /bin/sh -lc 'echo fork-v2 > /tmp/marker.txt' >/dev/null"
run_step "parent remains isolated" bash -lc "$VPC_BIN machine exec $PARENT_ID -- /bin/cat /tmp/marker.txt | grep -q parent-v1"

FC_PID="$("$VPC_BIN" machine inspect "$PARENT_ID" | sed -n 's/.*"runtime_id": "\([^"]*\)".*/\1/p' | head -n1)"
pkill -f "firecracker.*$PARENT_ID" 2>/dev/null || true
run_step "recover after firecracker kill" bash -lc "$VPC_BIN machine stop $PARENT_ID >/dev/null || true; $VPC_BIN machine start $PARENT_ID >/dev/null"

pkill -f "vpc-agent.*$PARENT_ID" 2>/dev/null || true
sleep 1
run_step "recover after guest-agent kill" bash -lc "$VPC_BIN machine exec $PARENT_ID -- /bin/echo agent-recovered >/dev/null"

kill "$DAEMON_PID" && wait "$DAEMON_PID" || true
"$VPCD_BIN" > "$LOG_DIR/daemon-restart.log" 2>&1 &
DAEMON_PID=$!
for _ in $(seq 1 60); do
  "$VPC_BIN" daemon status > /dev/null 2>&1 && break
  sleep 0.2
done
run_step "daemon restart reconciles machine state" bash -lc "$VPC_BIN machine inspect $PARENT_ID >/dev/null && $VPC_BIN machine inspect $FORK_ID >/dev/null"

for m in "$NAT_ID" "$OFFLINE_ID" "$ALLOW_ID" "$FORK_ID"; do
  "$VPC_BIN" machine stop "$m" >/dev/null || true
  "$VPC_BIN" machine destroy "$m" >/dev/null || true
done

orph_dirs=$(find "$VPCD_FIRECRACKER_DIR" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l || true)
orph_socks=$(find "$VPCD_FIRECRACKER_DIR" -type s 2>/dev/null | wc -l || true)
orph_taps=$(ip -o link show 2>/dev/null | awk -F': ' '{print $2}' | grep -E '^tap-' | wc -l || true)
run_step "no orphan runtime dirs" test "$orph_dirs" -eq 0
run_step "no orphan runtime sockets" test "$orph_socks" -eq 0
run_step "no orphan tap interfaces" test "$orph_taps" -eq 0

echo "--- e2e summary ---"
echo "pass=$pass_count fail=$fail_count warn=$warn_count"
(( fail_count == 0 ))
