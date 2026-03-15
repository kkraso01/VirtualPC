package integration

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"virtualpc/internal/api"
	"virtualpc/internal/cli"
	"virtualpc/internal/config"
	"virtualpc/internal/daemon"
)

func TestCLIToDaemonFlow(t *testing.T) {
	d := t.TempDir()
	sock := filepath.Join(d, "v.sock")
	cfg := config.Config{UnixSocket: sock, DataPath: filepath.Join(d, "state.json"), FirecrackerDir: filepath.Join(d, "fc")}
	svc, err := daemon.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sock)
	defer l.Close()
	go http.Serve(l, api.New(svc).Handler())
	c := cli.New(sock)
	var st map[string]any
	if err := c.Do("GET", "/v1/daemon/status", nil, &st); err != nil {
		t.Fatal(err)
	}
	if st["machines"] == nil {
		b, _ := json.Marshal(st)
		t.Fatalf("bad status %s", string(b))
	}
}
