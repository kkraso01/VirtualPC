package integration

import (
	"os"
	"path/filepath"
	"testing"

	"virtualpc/internal/config"
	"virtualpc/internal/daemon"
)

func writeExecutable(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMachineLifecycle(t *testing.T) {
	d := t.TempDir()
	fc := writeExecutable(t, d, "fake-firecracker.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	agent := writeExecutable(t, d, "fake-agent.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	svc, err := daemon.New(config.Config{UnixSocket: filepath.Join(d, "x.sock"), DataPath: filepath.Join(d, "state.json"), FirecrackerDir: filepath.Join(d, "fc"), FirecrackerBin: fc, AgentBin: agent})
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
