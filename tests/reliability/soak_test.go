package reliability

import (
	"os"
	"os/exec"
	"testing"
)

func TestSoak(t *testing.T) {
	if os.Getenv("VPC_SOAK_REAL_HOST") != "1" {
		t.Skip("set VPC_SOAK_REAL_HOST=1 to run destructive soak tests on real host")
	}
	cmd := exec.Command("bash", "scripts/soak.sh")
	cmd.Dir = "../.."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("soak failed: %v", err)
	}
}
