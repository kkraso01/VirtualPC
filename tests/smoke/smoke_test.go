package smoke

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

func TestCreateSnapshotFork(t *testing.T) {
	d := t.TempDir()
	fc := writeExecutable(t, d, "fake-firecracker.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	agent := writeExecutable(t, d, "fake-agent.sh", "#!/usr/bin/env bash\nwhile true; do sleep 3600; done\n")
	svc, _ := daemon.New(config.Config{UnixSocket: filepath.Join(d, "x.sock"), DataPath: filepath.Join(d, "state.json"), FirecrackerDir: filepath.Join(d, "fc"), FirecrackerBin: fc, AgentBin: agent})
	m, _ := svc.CreateMachine("minimal-shell")
	_, _ = svc.StartMachine(m.ID)
	snap, err := svc.CreateSnapshot(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ForkMachine(snap.ID); err != nil {
		t.Fatal(err)
	}
}
