package reliability

import (
	"os"
	"os/exec"
	"testing"
)

func TestNetworkPolicyEnforcementRealHost(t *testing.T) {
	if os.Getenv("VPC_NETWORK_REAL_HOST") != "1" {
		t.Skip("set VPC_NETWORK_REAL_HOST=1 to run network policy enforcement validation")
	}
	cmd := exec.Command("bash", "-lc", `
set -euo pipefail
ROOT_DIR="$(pwd)"
SOCK="$ROOT_DIR/.tmp/network-policy.sock"
RUN_DIR="$ROOT_DIR/.tmp/network-policy"
mkdir -p "$RUN_DIR"
export VPC_UNIX_SOCKET="$SOCK"
export VPCD_UNIX_SOCKET="$SOCK"
export VPCD_DATA_PATH="$RUN_DIR/state.json"
export VPCD_FIRECRACKER_DIR="$RUN_DIR/firecracker"
"$ROOT_DIR/bin/virtualpcd" > "$RUN_DIR/daemon.log" 2>&1 &
DAEMON_PID=$!
trap 'kill "$DAEMON_PID" >/dev/null 2>&1 || true; rm -f "$SOCK"' EXIT
for _ in $(seq 1 60); do
  vpc daemon status >/dev/null 2>&1 && break
  sleep 0.2
done
probe() {
  local machine="$1"
  local host="$2"
  vpc machine exec "$machine" -- /bin/sh -lc "ping -c1 -W2 $host >/dev/null"
}
create_start() {
  local profile="$1"
  local id
  id="$(vpc machine create --profile "$profile" | sed -n 's/.*"id": "\([^"]*\)".*/\1/p' | head -n1)"
  test -n "$id"
  vpc machine start "$id" >/dev/null
  echo "$id"
}
NAT_ID="$(create_start minimal-shell)"
OFF_ID="$(create_start minimal-shell-offline)"
ALL_ID="$(create_start minimal-shell-allowlist)"
trap 'vpc machine stop "$NAT_ID" >/dev/null 2>&1 || true; vpc machine destroy "$NAT_ID" >/dev/null 2>&1 || true; vpc machine stop "$OFF_ID" >/dev/null 2>&1 || true; vpc machine destroy "$OFF_ID" >/dev/null 2>&1 || true; vpc machine stop "$ALL_ID" >/dev/null 2>&1 || true; vpc machine destroy "$ALL_ID" >/dev/null 2>&1 || true' EXIT
probe "$NAT_ID" 1.1.1.1
! probe "$OFF_ID" 1.1.1.1
probe "$ALL_ID" 1.1.1.1
! probe "$ALL_ID" 9.9.9.9
`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("network policy validation failed: %v", err)
	}
}
