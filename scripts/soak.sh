#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VPC_BIN="${VPC_BIN:-$ROOT_DIR/bin/vpc}"
VPCD_BIN="${VPCD_BIN:-$ROOT_DIR/bin/virtualpcd}"
SOCK="${VPC_UNIX_SOCKET:-/tmp/virtualpcd.sock}"
LOOPS="${SOAK_LOOPS:-10}"
RUN_DIR="${VPC_SOAK_RUN_DIR:-$ROOT_DIR/.tmp/soak}"
mkdir -p "$RUN_DIR"
REPORT="$RUN_DIR/report.txt"
: > "$REPORT"

VM_BOOT_THRESHOLD="${VM_BOOT_THRESHOLD:-99}"
AGENT_THRESHOLD="${AGENT_THRESHOLD:-99}"
EXEC_THRESHOLD="${EXEC_THRESHOLD:-99}"
SNAPSHOT_FORK_THRESHOLD="${SNAPSHOT_FORK_THRESHOLD:-98}"

export VPC_UNIX_SOCKET="$SOCK"
export VPCD_UNIX_SOCKET="$SOCK"
export VPCD_DATA_PATH="${VPCD_DATA_PATH:-$RUN_DIR/state.json}"
export VPCD_FIRECRACKER_DIR="${VPCD_FIRECRACKER_DIR:-$RUN_DIR/firecracker}"

"$VPCD_BIN" > "$RUN_DIR/daemon.log" 2>&1 &
DAEMON_PID=$!
cleanup(){
  kill "$DAEMON_PID" >/dev/null 2>&1 || true
  wait "$DAEMON_PID" >/dev/null 2>&1 || true
  rm -f "$SOCK"
}
trap cleanup EXIT
for _ in $(seq 1 60); do
  "$VPC_BIN" daemon status >/dev/null 2>&1 && break
  sleep 0.2
done

pass=0; fail=0
boot_ok=0; agent_ok=0; exec_ok=0; snapfork_ok=0

json_get(){
  local json_input="$1" expr="$2"
  python3 - "$expr" "$json_input" <<'PY'
import json,sys
obj=json.loads(sys.argv[2])
for part in sys.argv[1].split('.'):
    if not part: continue
    obj=obj[int(part)] if isinstance(obj,list) else obj.get(part)
print(obj if obj is not None else "")
PY
}

record(){
  local name="$1"; shift
  if "$@"; then
    pass=$((pass+1)); echo "PASS $name" >> "$REPORT"; return 0
  fi
  fail=$((fail+1)); echo "FAIL $name" >> "$REPORT"; return 1
}

for i in $(seq 1 "$LOOPS"); do
  m_json="$($VPC_BIN machine create --profile minimal-shell)"
  m_id="$(json_get "$m_json" id)"
  record "create-$i" test -n "$m_id"

  if record "start-$i" bash -lc "$VPC_BIN machine start $m_id >/dev/null"; then boot_ok=$((boot_ok+1)); fi
  if record "agent-connect-$i" bash -lc "$VPC_BIN machine exec $m_id -- /bin/echo agent-$i >/dev/null"; then agent_ok=$((agent_ok+1)); fi
  if record "exec-$i" bash -lc "$VPC_BIN machine exec $m_id -- /bin/sh -lc 'echo exec-$i' >/dev/null"; then exec_ok=$((exec_ok+1)); fi

  record "seed-file-$i" bash -lc "$VPC_BIN machine exec $m_id -- /bin/sh -lc 'echo soak-$i > /tmp/seed.txt' >/dev/null"
  snap_json="$($VPC_BIN snapshot create $m_id)"
  snap_id="$(json_get "$snap_json" id)"
  ok_snap=0
  if record "snapshot-$i" test -n "$snap_id"; then
    fork_json="$($VPC_BIN machine fork $snap_id)"
    fork_id="$(json_get "$fork_json" id)"
    if record "fork-$i" test -n "$fork_id" && \
       record "fork-start-$i" bash -lc "$VPC_BIN machine start $fork_id >/dev/null" && \
       record "fork-integrity-$i" bash -lc "$VPC_BIN machine exec $fork_id -- /bin/cat /tmp/seed.txt | grep -q soak-$i" && \
       record "fork-isolation-$i" bash -lc "$VPC_BIN machine exec $fork_id -- /bin/sh -lc 'echo fork-$i > /tmp/seed.txt' >/dev/null; $VPC_BIN machine exec $m_id -- /bin/cat /tmp/seed.txt | grep -q soak-$i"; then
      ok_snap=1
      snapfork_ok=$((snapfork_ok+1))
    fi
    $VPC_BIN machine stop $fork_id >/dev/null || true
    $VPC_BIN machine destroy $fork_id >/dev/null || true
  fi

  pkill -f "firecracker.*$m_id" 2>/dev/null || true
  record "recover-after-firecracker-kill-$i" bash -lc "$VPC_BIN machine stop $m_id >/dev/null || true; $VPC_BIN machine start $m_id >/dev/null"

  pkill -f "vpc-agent.*$m_id" 2>/dev/null || true
  sleep 0.2
  record "recover-after-agent-kill-$i" bash -lc "$VPC_BIN machine exec $m_id -- /bin/echo reconnect-$i >/dev/null"

  record "stop-$i" bash -lc "$VPC_BIN machine stop $m_id >/dev/null"
  record "destroy-$i" bash -lc "$VPC_BIN machine destroy $m_id >/dev/null"
done

pct(){ awk -v ok="$1" -v total="$2" 'BEGIN { if (total==0) print "0.00"; else printf "%.2f", (ok*100.0)/total }'; }
boot_rate="$(pct "$boot_ok" "$LOOPS")"
agent_rate="$(pct "$agent_ok" "$LOOPS")"
exec_rate="$(pct "$exec_ok" "$LOOPS")"
snapfork_rate="$(pct "$snapfork_ok" "$LOOPS")"

orph_dirs=$(find "${VPCD_FIRECRACKER_DIR:-$ROOT_DIR/data/firecracker}" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l || true)
orph_procs=$(pgrep -f 'firecracker|vpc-agent' | wc -l || true)
orph_socks=$(find /tmp -maxdepth 1 -name 'virtualpc*.sock' 2>/dev/null | wc -l || true)
orph_taps=$(ip -o link show 2>/dev/null | awk -F': ' '{print $2}' | grep -E '^tap-' | wc -l || true)
orph_disks=$(find "${VPCD_FIRECRACKER_DIR:-$ROOT_DIR/data/firecracker}" -type f \( -name '*.ext4' -o -name '*.img' -o -name '*.snap' \) 2>/dev/null | wc -l || true)

{
  echo "pass=$pass"
  echo "fail=$fail"
  echo "boot_success_rate=$boot_rate"
  echo "guest_agent_success_rate=$agent_rate"
  echo "exec_success_rate=$exec_rate"
  echo "snapshot_fork_success_rate=$snapfork_rate"
  echo "orphaned_runtime_dirs=$orph_dirs"
  echo "orphaned_processes=$orph_procs"
  echo "leaked_sockets=$orph_socks"
  echo "leaked_tap_interfaces=$orph_taps"
  echo "leaked_disks_artifacts=$orph_disks"
} >> "$REPORT"

cat "$REPORT"
awk -v x="$boot_rate" -v t="$VM_BOOT_THRESHOLD" 'BEGIN{exit (x+0>=t+0)?0:1}' || { echo "boot success rate gate failed" >&2; exit 1; }
awk -v x="$agent_rate" -v t="$AGENT_THRESHOLD" 'BEGIN{exit (x+0>=t+0)?0:1}' || { echo "agent success rate gate failed" >&2; exit 1; }
awk -v x="$exec_rate" -v t="$EXEC_THRESHOLD" 'BEGIN{exit (x+0>=t+0)?0:1}' || { echo "exec success rate gate failed" >&2; exit 1; }
awk -v x="$snapfork_rate" -v t="$SNAPSHOT_FORK_THRESHOLD" 'BEGIN{exit (x+0>=t+0)?0:1}' || { echo "snapshot+fork success gate failed" >&2; exit 1; }
(( fail == 0 ))
