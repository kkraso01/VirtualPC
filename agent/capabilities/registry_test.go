package capabilities

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"virtualpc/agent/config"
	"virtualpc/agent/mcp"
)

type fakeMCP struct{}

func (fakeMCP) Discover(_ context.Context, _ mcp.ServerConfig) (mcp.Discovery, error) {
	return mcp.Discovery{Tools: []mcp.DiscoveredTool{{Name: "mcp_tool", InputSchema: map[string]any{"type": "object"}}}}, nil
}

func TestMergeOverride(t *testing.T) {
	out := Merge([]Capability{{ID: "a", Name: "x"}}, []Capability{{ID: "a", Name: "y"}})
	if len(out) != 1 || out[0].Name != "y" {
		t.Fatalf("unexpected merge result: %+v", out)
	}
}

func TestLoaderAndPolicyBindings(t *testing.T) {
	d := t.TempDir()
	mustWrite(t, filepath.Join(d, "skills/coding/SKILL.md"), "---\ndescription: coding\n---")
	mustWrite(t, filepath.Join(d, "skills/coding/tools.yaml"), "tools:\n  - name: run_command\n    enabled: true\n")
	mustWrite(t, filepath.Join(d, "skills/coding/prompt.md"), "x")
	mustWrite(t, filepath.Join(d, "skills/coding/policies.yaml"), "allowed_paths: ['/workspace']")
	mustWrite(t, filepath.Join(d, "tools/local/test.yaml"), "name: risky\ndescription: x\nbackend_type: local\nenabled: true\npolicy:\n  approvals_required: ['human']\n")
	mustWrite(t, filepath.Join(d, "tools/http/h.yaml"), "name: web\ndescription: y\nbackend_type: http\nenabled: true\n")
	mustWrite(t, filepath.Join(d, "profiles/p.yaml"), "name: p\nprovider: openai\nmodel: gpt-4o\nsupports_tool_calling: true\n")
	mustWrite(t, filepath.Join(d, "mcp.yaml"), "mcp_servers:\n  - name: m\n    mode: remote\n    url: http://localhost")
	reg, _, _, _, err := Load(context.Background(), LoaderOptions{SkillsRoot: filepath.Join(d, "skills"), LocalToolsRoot: filepath.Join(d, "tools/local"), HTTPToolsRoot: filepath.Join(d, "tools/http"), ProviderProfilesRoot: filepath.Join(d, "profiles"), MCPConfigPath: filepath.Join(d, "mcp.yaml"), MCPClient: fakeMCP{}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reg.FindByName("mcp_tool"); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	ApplyPolicyBindings(&cfg, reg.All())
	if !cfg.ApprovalRequiredTools["risky"] {
		t.Fatalf("expected approval binding")
	}
}

func mustWrite(t *testing.T, p, v string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(v), 0o644); err != nil {
		t.Fatal(err)
	}
}
