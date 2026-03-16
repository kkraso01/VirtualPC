package safety

import (
	"testing"
	"time"
)

func TestCommandPolicyBlocksForkBomb(t *testing.T) {
	p := DefaultCommandPolicy()
	d, _ := p.Evaluate(":(){ :|:& };:", "block")
	if d != DecisionBlock {
		t.Fatalf("expected block")
	}
}

func TestFilesystemGuard(t *testing.T) {
	g := NewFilesystemGuard("/workspace", "/tmp")
	if err := g.Validate("/workspace/a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := g.Validate("/etc/passwd"); err == nil {
		t.Fatal("expected block")
	}
}

func TestResourceLimits(t *testing.T) {
	r := ResourceLimits{MaxCommandsPerSession: 1, MaxRuntime: time.Second, MaxIterations: 1, MaxFailures: 1, MaxRepeatedCommand: 1}
	if err := r.Validate(2, time.Now(), 0, 0, 0); err == nil {
		t.Fatal("expected commands limit")
	}
}
