package unit

import (
	"testing"

	"virtualpc/internal/runtime/firecracker"
)

func TestGuestCommandFlow(t *testing.T) {
	m := firecracker.NewManager(t.TempDir())
	_, _ = m.Start("abc1234567890def")
	out, err := m.Exec("abc1234567890def", []string{"echo", "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("expected output")
	}
}
