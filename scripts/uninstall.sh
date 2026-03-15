#!/usr/bin/env bash
set -euo pipefail

PREFIX="${VPC_INSTALL_PREFIX:-/opt/virtualpc}"
ETC_DIR="${VPC_ETC_DIR:-/etc/virtualpc}"
PURGE_STATE="${VPC_PURGE_STATE:-0}"
SYSTEMCTL_BIN="${SYSTEMCTL_BIN:-systemctl}"
RUN_DIR="${VPC_RUN_DIR:-/run/virtualpc}"
DATA_DIR="${VPC_DATA_DIR:-/var/lib/virtualpc}"

"$SYSTEMCTL_BIN" stop virtualpcd 2>/dev/null || true
"$SYSTEMCTL_BIN" disable virtualpcd 2>/dev/null || true
rm -f /etc/systemd/system/virtualpcd.service
"$SYSTEMCTL_BIN" daemon-reload

pkill -f 'virtualpcd|firecracker|vpc-agent' 2>/dev/null || true
rm -f /tmp/virtualpcd.sock "$RUN_DIR"/*.sock 2>/dev/null || true

rm -rf "$PREFIX/bin"
rm -f "$ETC_DIR/virtualpcd.env"

ip -o link show 2>/dev/null | awk -F': ' '{print $2}' | grep -E '^tap-' | xargs -r -n1 ip link del >/dev/null 2>&1 || true

rm -rf "$RUN_DIR"

if [[ "$PURGE_STATE" == "1" ]]; then
  rm -rf "$DATA_DIR"
fi

echo "VirtualPC uninstalled (purge_state=$PURGE_STATE)"
