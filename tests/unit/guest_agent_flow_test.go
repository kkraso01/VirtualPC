package unit

import (
	"strings"
	"testing"

	"virtualpc/internal/runtime/firecracker"
)

func TestGuestCommandFlowRequiresAgent(t *testing.T) {
	d := t.TempDir()
	fc := writeExecutable(t, d, "fake-firecracker.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	agent := writeExecutable(t, d, "fake-agent.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	m := firecracker.NewManager(t.TempDir(), fc, agent)
	_, _ = m.Start("abc1234567890def")
	_, err := m.Exec("abc1234567890def", []string{"echo", "hello"})
	if err == nil {
		t.Fatal("expected hard failure when guest agent path is unavailable")
	}
	if !strings.Contains(err.Error(), "fallback disabled") {
		t.Fatalf("unexpected error: %v", err)
	}
}
