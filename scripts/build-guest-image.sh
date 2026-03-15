#!/usr/bin/env bash
set -euo pipefail
ROOT=${1:-./data/guest-image}
mkdir -p "$ROOT/bin" "$ROOT/etc" "$ROOT/var/lib/containerd" "$ROOT/opt/cni/bin"
cp ./bin/vpc-agent "$ROOT/bin/vpc-agent" 2>/dev/null || true
cat > "$ROOT/etc/vpc-agent.env" <<EOC
VPC_AGENT_SOCKET=/run/vpc-agent.sock
CONTAINER_RUNTIME=containerd
EOC
cat > "$ROOT/README" <<EOC
VirtualPC guest image layout
- includes vpc-agent
- containerd/nerdctl expected in image pipeline
- compatible with firecracker rootfs build stage
EOC
echo "guest image prepared at $ROOT"
