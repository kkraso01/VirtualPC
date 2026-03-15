package unit

import (
	"path/filepath"
	"testing"

	"virtualpc/internal/state"
)

func TestMachinePersistence(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "state.json")
	s, err := state.New(p)
	if err != nil {
		t.Fatal(err)
	}
	m, err := s.CreateMachine("minimal-shell")
	if err != nil {
		t.Fatal(err)
	}
	s2, err := state.New(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s2.GetMachine(m.ID); err != nil {
		t.Fatalf("expected machine to persist: %v", err)
	}
}
