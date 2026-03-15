package unit

import (
	"testing"

	"virtualpc/internal/runtime/firecracker"
)

func TestFirecrackerManagerLifecycle(t *testing.T) {
	d := t.TempDir()
	fc := writeExecutable(t, d, "fake-firecracker.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	agent := writeExecutable(t, d, "fake-agent.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	m := firecracker.NewManager(t.TempDir(), fc, agent)
	id, err := m.Start("abc1234567890def")
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("expected runtime id")
	}
	if err := m.Stop("abc1234567890def"); err != nil {
		t.Fatal(err)
	}
	if err := m.Destroy("abc1234567890def"); err != nil {
		t.Fatal(err)
	}
}
