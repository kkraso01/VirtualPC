#!/usr/bin/env bash
set -euo pipefail

PREFIX="${VPC_INSTALL_PREFIX:-/opt/virtualpc}"
ETC_DIR="${VPC_ETC_DIR:-/etc/virtualpc}"
RUN_DIR="${VPC_RUN_DIR:-/run/virtualpc}"
DATA_DIR="${VPC_DATA_DIR:-/var/lib/virtualpc}"

require_bin(){ command -v "$1" >/dev/null || { echo "missing required command: $1" >&2; exit 1; }; }
SYSTEMCTL_BIN="${SYSTEMCTL_BIN:-systemctl}"

[[ "$(uname -s)" == "Linux" ]] || { echo "Linux host required" >&2; exit 1; }
require_bin install
require_bin "$SYSTEMCTL_BIN"

if [[ "${VPC_INSTALL_SKIP_KVM_CHECK:-0}" != "1" && ! -e /dev/kvm ]]; then
  echo "missing /dev/kvm; hardware virtualization is required" >&2
  exit 1
fi

if ! command -v "${FIRECRACKER_BIN:-firecracker}" >/dev/null; then
  echo "firecracker not found in PATH; install Firecracker before proceeding" >&2
  exit 1
fi

mkdir -p "$PREFIX/bin" "$ETC_DIR" "$RUN_DIR" "$DATA_DIR"
install -m 0755 bin/virtualpcd "$PREFIX/bin/virtualpcd"
install -m 0755 bin/vpc "$PREFIX/bin/vpc"
install -m 0755 bin/vpc-agent "$PREFIX/bin/vpc-agent"
install -m 0644 packaging/releases/example-config.env "$ETC_DIR/virtualpcd.env"
install -m 0644 packaging/systemd/virtualpcd.service /etc/systemd/system/virtualpcd.service

if [[ ! -f data/firecracker/rootfs.ext4 || ! -f data/firecracker/vmlinux ]]; then
  echo "missing guest assets under data/firecracker; run scripts/build-guest-image.sh" >&2
  exit 1
fi

"$SYSTEMCTL_BIN" daemon-reload

echo "Install complete. Next commands:"
echo "  sudo systemctl enable --now virtualpcd"
echo "  $PREFIX/bin/vpc doctor"
echo "  $PREFIX/bin/vpc daemon status"
