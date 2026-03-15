#!/usr/bin/env bash
set -euo pipefail

PREFIX="${VPC_INSTALL_PREFIX:-/opt/virtualpc}"
ETC_DIR="${VPC_ETC_DIR:-/etc/virtualpc}"
PURGE_STATE="${VPC_PURGE_STATE:-0}"

systemctl stop virtualpcd 2>/dev/null || true
systemctl disable virtualpcd 2>/dev/null || true
rm -f /etc/systemd/system/virtualpcd.service
systemctl daemon-reload

pkill -f 'virtualpcd|firecracker|vpc-agent' 2>/dev/null || true
rm -f /tmp/virtualpcd.sock /run/virtualpc/*.sock 2>/dev/null || true

rm -rf "$PREFIX/bin"
rm -f "$ETC_DIR/virtualpcd.env"

if [[ "$PURGE_STATE" == "1" ]]; then
  rm -rf /var/lib/virtualpc
fi

echo "VirtualPC uninstalled (purge_state=$PURGE_STATE)"
