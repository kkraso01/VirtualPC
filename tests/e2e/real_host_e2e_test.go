package e2e

import (
	"os"
	"os/exec"
	"testing"
)

func TestRealHostE2E(t *testing.T) {
	if os.Getenv("VPC_E2E_REAL_HOST") != "1" {
		t.Skip("set VPC_E2E_REAL_HOST=1 on a Linux/KVM host to run release-grade e2e")
	}
	cmd := exec.Command("bash", "scripts/e2e-real-host.sh")
	cmd.Dir = "../.."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("real-host e2e failed: %v", err)
	}
}
