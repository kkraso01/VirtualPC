package integration

import (
	"path/filepath"
	"testing"

	"virtualpc/internal/config"
	"virtualpc/internal/daemon"
)

func TestServiceLifecycleRequiresGuestRuntime(t *testing.T) {
	d := t.TempDir()
	svc, _ := daemon.New(config.Config{UnixSocket: filepath.Join(d, "x.sock"), DataPath: filepath.Join(d, "state.json"), FirecrackerDir: filepath.Join(d, "fc")})
	m, _ := svc.CreateMachine("minimal-shell")
	_, err := svc.CreateService(m.ID, "db", "postgres:16")
	if err == nil {
		t.Fatal("expected create service to fail without guest runtime")
	}
}
