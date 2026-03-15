#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VPC_BIN="${VPC_BIN:-$ROOT_DIR/bin/vpc}"
LOOPS="${SOAK_LOOPS:-10}"
RUN_DIR="${VPC_SOAK_RUN_DIR:-$ROOT_DIR/.tmp/soak}"
mkdir -p "$RUN_DIR"
REPORT="$RUN_DIR/report.txt"
: > "$REPORT"

pass=0
fail=0
json_get(){
  local json_input="$1"
  local expr="$2"
  python3 - "$expr" "$json_input" <<'PY'
import json,sys
expr=sys.argv[1]
obj=json.loads(sys.argv[2])
for part in expr.split('.'):
    if not part: continue
    obj = obj[int(part)] if isinstance(obj,list) else obj.get(part)
print(obj if obj is not None else "")
PY
}

record(){
  local name="$1"; shift
  if "$@"; then
    pass=$((pass+1)); echo "PASS $name" >> "$REPORT"
  else
    fail=$((fail+1)); echo "FAIL $name" >> "$REPORT"
  fi
}

for i in $(seq 1 "$LOOPS"); do
  m_json="$($VPC_BIN machine create --profile minimal-shell)"
  m_id="$(json_get "$m_json" id)"
  record "create-$i" test -n "$m_id"
  record "start-$i" bash -lc "$VPC_BIN machine start $m_id >/dev/null"
  record "exec-$i" bash -lc "$VPC_BIN machine exec $m_id -- /bin/echo loop-$i >/dev/null"
  record "pty-shell-$i" bash -lc "$VPC_BIN machine shell $m_id >/dev/null"

  snap_json="$($VPC_BIN snapshot create $m_id)"
  snap_id="$(json_get "$snap_json" id)"
  record "snapshot-$i" test -n "$snap_id"
  fork_json="$($VPC_BIN machine fork $snap_id)"
  fork_id="$(json_get "$fork_json" id)"
  record "fork-$i" test -n "$fork_id"

  svc_json="$($VPC_BIN service create --machine $m_id --name soak-$i --image alpine:latest)"
  svc_id="$(json_get "$svc_json" id)"
  record "service-create-$i" test -n "$svc_id"
  record "service-list-$i" bash -lc "$VPC_BIN service list --machine $m_id >/dev/null"
  record "service-logs-$i" bash -lc "$VPC_BIN service logs $svc_id >/dev/null"
  record "service-destroy-$i" bash -lc "$VPC_BIN service destroy $svc_id >/dev/null"

  large="$RUN_DIR/large-$i.bin"
  dd if=/dev/urandom of="$large" bs=1M count=2 status=none
  record "cp-to-large-$i" bash -lc "$VPC_BIN machine cp-to $m_id $large /tmp/large-$i.bin >/dev/null"
  record "cp-from-large-$i" bash -lc "$VPC_BIN machine cp-from $m_id /tmp/large-$i.bin $RUN_DIR/down-$i.bin >/dev/null"

  pkill -f "firecracker.*$m_id" 2>/dev/null || true
  record "recover-after-firecracker-kill-$i" bash -lc "$VPC_BIN machine stop $m_id >/dev/null || true; $VPC_BIN machine start $m_id >/dev/null"

  pkill -f "vpc-agent.*$m_id" 2>/dev/null || true
  sleep 0.2
  record "guest-agent-reconnect-$i" bash -lc "$VPC_BIN machine exec $m_id -- /bin/echo reconnect-$i >/dev/null"

  record "stop-$i" bash -lc "$VPC_BIN machine stop $m_id >/dev/null"
  record "destroy-$i" bash -lc "$VPC_BIN machine destroy $m_id >/dev/null"
done

orph_dirs=$(find "${VPCD_FIRECRACKER_DIR:-$ROOT_DIR/data/firecracker}" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l || true)
orph_procs=$(pgrep -f 'firecracker|vpc-agent' | wc -l || true)
orph_socks=$(find /tmp -maxdepth 1 -name 'virtualpc*.sock' 2>/dev/null | wc -l || true)
orph_taps=$(ip -o link show 2>/dev/null | awk -F': ' '{print $2}' | grep -E '^tap-' | wc -l || true)
orph_disks=$(find "${VPCD_FIRECRACKER_DIR:-$ROOT_DIR/data/firecracker}" -type f \( -name '*.ext4' -o -name '*.img' -o -name '*.snap' \) 2>/dev/null | wc -l || true)

{
  echo "pass=$pass"
  echo "fail=$fail"
  echo "orphaned_runtime_dirs=$orph_dirs"
  echo "orphaned_processes=$orph_procs"
  echo "leaked_sockets=$orph_socks"
  echo "leaked_tap_interfaces=$orph_taps"
  echo "leaked_disks_artifacts=$orph_disks"
} >> "$REPORT"

cat "$REPORT"
(( fail == 0 ))
