package controller

import (
	"testing"
	"virtualpc/agent/capabilities"
	"virtualpc/agent/skills"
)

func TestBuildContextSkillOverlay(t *testing.T) {
	caps := []capabilities.Capability{
		{ID: "builtin.tool.run_command", Name: "run_command", Type: capabilities.TypeTool, Source: capabilities.SourceBuiltin, Enabled: true},
		{ID: "builtin.tool.read_file", Name: "read_file", Type: capabilities.TypeTool, Source: capabilities.SourceBuiltin, Enabled: true},
	}
	man := []skills.SkillManifest{{Name: "coding", Prompt: "coding prompt", Tools: []skills.ToolBinding{{Name: "run_command", Enabled: true}}}}
	ctx := BuildContext("base", []string{"coding"}, man, caps, []string{"history"})
	if len(ctx.Tools) != 1 || ctx.Tools[0].Name != "run_command" {
		t.Fatalf("unexpected tools: %#v", ctx.Tools)
	}
	if ctx.Prompt == "" {
		t.Fatal("prompt empty")
	}
}
