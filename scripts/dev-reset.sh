#!/usr/bin/env bash
set -euo pipefail
if [[ -f /tmp/virtualpcd.pid ]]; then kill "$(cat /tmp/virtualpcd.pid)" || true; rm -f /tmp/virtualpcd.pid; fi
rm -f /tmp/virtualpcd.sock
rm -rf ./data
docker compose -f packaging/docker/compose.dev.yml down -v
