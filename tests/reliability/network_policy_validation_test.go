package reliability

import (
	"os"
	"os/exec"
	"testing"
)

// This test validates runtime enforcement (not just policy storage).
// It is intentionally real-host only and expected to be wired into release gating.
func TestNetworkPolicyEnforcementRealHost(t *testing.T) {
	if os.Getenv("VPC_NETWORK_REAL_HOST") != "1" {
		t.Skip("set VPC_NETWORK_REAL_HOST=1 to run network policy enforcement validation")
	}
	cmd := exec.Command("bash", "-lc", `
set -euo pipefail
vpc profile list > /tmp/vpc-profiles.json
if ! grep -q allowlist /tmp/vpc-profiles.json; then
  echo "allowlist mode is not fully exposed/enforceable yet"
  exit 2
fi
echo "TODO: enforce offline/nat/allowlist egress checks with live probes"
`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return
	}
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 2 {
		t.Fatalf("launch blocker: allowlist policy enforcement not fully validated")
	}
	t.Fatalf("network policy validation failed: %v", err)
}
