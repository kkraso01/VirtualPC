package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndValidate(t *testing.T) {
	d := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(d, "coding"), 0o755))
	must(t, os.WriteFile(filepath.Join(d, "coding", "SKILL.md"), []byte("---\ndescription: coding\n---"), 0o644))
	must(t, os.WriteFile(filepath.Join(d, "coding", "prompt.md"), []byte("prompt"), 0o644))
	must(t, os.WriteFile(filepath.Join(d, "coding", "tools.yaml"), []byte("tools:\n  - name: run_command\n    enabled: true\n"), 0o644))
	must(t, os.WriteFile(filepath.Join(d, "coding", "policies.yaml"), []byte("allowed_paths: ['/workspace']"), 0o644))
	m, err := LoadOne(filepath.Join(d, "coding"))
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "coding" || len(m.Tools) == 0 {
		t.Fatalf("invalid manifest: %+v", m)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
