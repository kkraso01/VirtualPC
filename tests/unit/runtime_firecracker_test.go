package unit

import (
	"testing"

	"virtualpc/internal/runtime/firecracker"
)

func TestFirecrackerManagerLifecycle(t *testing.T) {
	m := firecracker.NewManager(t.TempDir())
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
