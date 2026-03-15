package smoke

import (
	"path/filepath"
	"testing"

	"virtualpc/internal/config"
	"virtualpc/internal/daemon"
)

func TestCreateExecSnapshotFork(t *testing.T) {
	d := t.TempDir()
	svc, _ := daemon.New(config.Config{UnixSocket: filepath.Join(d, "x.sock"), DataPath: filepath.Join(d, "state.json"), FirecrackerDir: filepath.Join(d, "fc")})
	m, _ := svc.CreateMachine("minimal-shell")
	_, _ = svc.StartMachine(m.ID)
	if _, err := svc.Exec(m.ID, []string{"echo", "hello"}); err != nil {
		t.Fatal(err)
	}
	snap, err := svc.CreateSnapshot(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ForkMachine(snap.ID); err != nil {
		t.Fatal(err)
	}
}
