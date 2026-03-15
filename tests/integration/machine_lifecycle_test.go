package integration

import (
	"path/filepath"
	"testing"

	"virtualpc/internal/config"
	"virtualpc/internal/daemon"
)

func TestMachineLifecycle(t *testing.T) {
	d := t.TempDir()
	svc, err := daemon.New(config.Config{UnixSocket: filepath.Join(d, "x.sock"), DataPath: filepath.Join(d, "state.json"), FirecrackerDir: filepath.Join(d, "fc")})
	if err != nil {
		t.Fatal(err)
	}
	m, err := svc.CreateMachine("minimal-shell")
	if err != nil {
		t.Fatal(err)
	}
	m, err = svc.StartMachine(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if m.State != "running" {
		t.Fatalf("expected running got %s", m.State)
	}
	m, err = svc.StopMachine(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if m.State != "stopped" {
		t.Fatalf("expected stopped got %s", m.State)
	}
}
