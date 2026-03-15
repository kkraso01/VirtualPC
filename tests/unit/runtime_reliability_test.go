package unit

import (
	"testing"

	"virtualpc/internal/runtime/firecracker"
)

func TestRepeatedLifecycleLoops(t *testing.T) {
	d := t.TempDir()
	fc := writeExecutable(t, d, "fake-firecracker.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	agent := writeExecutable(t, d, "fake-agent.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	m := firecracker.NewManager(t.TempDir(), fc, agent)
	for i := 0; i < 5; i++ {
		id := "abc1234567890de" + string(rune('a'+i))
		if _, err := m.Start(id, firecracker.NetworkModeNAT, nil); err != nil {
			t.Fatalf("loop %d start: %v", i, err)
		}
		if err := m.Stop(id); err != nil {
			t.Fatalf("loop %d stop: %v", i, err)
		}
		if err := m.Destroy(id); err != nil {
			t.Fatalf("loop %d destroy: %v", i, err)
		}
	}
}
