#!/usr/bin/env bash
set -euo pipefail
mkdir -p /usr/local/bin
cp ./bin/virtualpcd ./bin/vpc ./bin/vpc-agent /usr/local/bin/
echo "installed virtualpc binaries to /usr/local/bin"
