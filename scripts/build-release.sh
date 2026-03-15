#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${1:-$ROOT_DIR/packaging/releases/out}"
VERSION="${VPC_RELEASE_VERSION:-$(git -C "$ROOT_DIR" rev-parse --short HEAD)}"
STAGE="$OUT_DIR/virtualpc-$VERSION"

rm -rf "$STAGE"
mkdir -p "$STAGE/bin" "$STAGE/assets" "$STAGE/systemd"

"$ROOT_DIR/scripts/build-binaries.sh"
cp "$ROOT_DIR/bin/virtualpcd" "$ROOT_DIR/bin/vpc" "$ROOT_DIR/bin/vpc-agent" "$STAGE/bin/"

if [[ -f "$ROOT_DIR/data/firecracker/vmlinux" ]]; then
  cp "$ROOT_DIR/data/firecracker/vmlinux" "$STAGE/assets/kernel"
fi
if [[ -f "$ROOT_DIR/data/firecracker/rootfs.ext4" ]]; then
  cp "$ROOT_DIR/data/firecracker/rootfs.ext4" "$STAGE/assets/guest-image.ext4"
fi

cp "$ROOT_DIR/packaging/releases/example-config.env" "$STAGE/assets/"
cp "$ROOT_DIR/scripts/install.sh" "$ROOT_DIR/scripts/uninstall.sh" "$ROOT_DIR/scripts/upgrade.sh" "$STAGE/"
cp "$ROOT_DIR/packaging/systemd/virtualpcd.service" "$STAGE/systemd/"

firecracker_version="$(firecracker --version 2>/dev/null || echo unknown)"
kernel_version="$(uname -r)"
agent_version="$VERSION"
daemon_version="$VERSION"
guest_image_version="$VERSION"

cat > "$STAGE/release-manifest.json" <<MANIFEST
{
  "release_version": "$VERSION",
  "firecracker_version": "$firecracker_version",
  "kernel_version": "$kernel_version",
  "guest_image_version": "$guest_image_version",
  "agent_version": "$agent_version",
  "daemon_version": "$daemon_version"
}
MANIFEST

(
  cd "$OUT_DIR"
  tar -czf "virtualpc-$VERSION.tar.gz" "virtualpc-$VERSION"
  sha256sum "virtualpc-$VERSION.tar.gz" > "virtualpc-$VERSION.tar.gz.sha256"
)

echo "release bundle created at $OUT_DIR/virtualpc-$VERSION.tar.gz"
