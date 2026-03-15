#!/usr/bin/env bash
set -euo pipefail
SOCK=${VPC_UNIX_SOCKET:-/tmp/virtualpcd.sock}
export VPC_UNIX_SOCKET="$SOCK"

./bin/vpc daemon status >/dev/null
MID=$(./bin/vpc machine create --profile minimal-shell | sed -n 's/.*"id": "\([^"]*\)".*/\1/p' | head -n1)
./bin/vpc machine start "$MID" >/dev/null
./bin/vpc machine exec "$MID" -- echo hello >/dev/null

echo "demo payload" > /tmp/vpc-smoke.txt
./bin/vpc machine cp-to "$MID" /tmp/vpc-smoke.txt /tmp/vpc-smoke.txt >/dev/null
./bin/vpc machine cp-from "$MID" /tmp/vpc-smoke.txt /tmp/vpc-smoke-back.txt >/dev/null

SVC=$(./bin/vpc service create --machine "$MID" --name db --image postgres:16 | sed -n 's/.*"id": "\([^"]*\)".*/\1/p' | head -n1)
./bin/vpc service list --machine "$MID" >/dev/null
./bin/vpc service logs "$SVC" >/dev/null

SID=$(./bin/vpc snapshot create "$MID" | sed -n 's/.*"id": "\([^"]*\)".*/\1/p' | head -n1)
./bin/vpc machine fork "$SID" >/dev/null

TID=$(./bin/vpc task create --machine "$MID" --goal "echo task ok" | sed -n 's/.*"id": "\([^"]*\)".*/\1/p' | head -n1)
./bin/vpc task run "$TID" >/dev/null

echo "smoke ok"
