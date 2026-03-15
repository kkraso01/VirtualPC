#!/usr/bin/env bash
set -euo pipefail
docker compose -f packaging/docker/compose.dev.yml up -d --build
./bin/virtualpcd &
echo $! > /tmp/virtualpcd.pid
