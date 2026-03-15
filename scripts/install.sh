#!/usr/bin/env bash
set -euo pipefail

command -v uname >/dev/null
if [[ "$(uname -s)" != "Linux" ]]; then
  echo "VirtualPC installer supports Linux only" >&2
  exit 1
fi

if [[ ! -e /dev/kvm ]]; then
  echo "warning: /dev/kvm not found (falling back to mock firecracker mode)" >&2
fi

mkdir -p bin data

go build -o bin/virtualpcd ./cmd/virtualpcd
go build -o bin/vpc ./cmd/vpc
go build -o bin/vpc-agent ./cmd/vpc-agent

if ! command -v firecracker >/dev/null; then
  echo "firecracker not found in PATH. Set VPCD_FIRECRACKER_BIN or install Firecracker manually." >&2
fi

scripts/build-guest-image.sh

echo "Install complete. Run: ./bin/virtualpcd"
