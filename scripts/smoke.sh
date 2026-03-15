#!/usr/bin/env bash
set -euo pipefail
SOCK=${VPC_UNIX_SOCKET:-/tmp/virtualpcd.sock}
export VPC_UNIX_SOCKET="$SOCK"
./bin/vpc daemon status >/dev/null
MID=$(./bin/vpc machine create --profile minimal-shell | sed -n 's/.*"id": "\([^"]*\)".*/\1/p' | head -n1)
./bin/vpc machine start "$MID" >/dev/null
./bin/vpc machine exec "$MID" -- echo hello >/dev/null
SID=$(./bin/vpc snapshot create "$MID" | sed -n 's/.*"id": "\([^"]*\)".*/\1/p' | head -n1)
./bin/vpc machine fork "$SID" >/dev/null
echo "smoke ok"
