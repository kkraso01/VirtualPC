#!/usr/bin/env bash
set -euo pipefail

echo "Upgrading VirtualPC in-place"
./scripts/install.sh
systemctl restart virtualpcd

echo "Upgrade complete. Validate with:"
echo "  /opt/virtualpc/bin/vpc doctor"
echo "  /opt/virtualpc/bin/vpc daemon status"
