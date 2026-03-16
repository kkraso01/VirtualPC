package controller

import (
	"virtualpc/agent/capabilities"
	"virtualpc/agent/skills"
)

type Context struct {
	Prompt        string
	Tools         []capabilities.Capability
	PolicySummary map[string]any
	Resources     []string
}

func BuildContext(basePrompt string, selectedSkills []string, manifests []skills.SkillManifest, caps []capabilities.Capability, history []string) Context {
	rt := skills.BuildRuntime(basePrompt, selectedSkills, manifests)
	prompt := mergePrompt(rt.EffectivePrompt, mergePrompt(history...))
	effective := caps
	if len(rt.Overlay.ToolAllow) > 0 {
		for i := range effective {
			if effective[i].Type == capabilities.TypeTool {
				if v, ok := rt.Overlay.ToolAllow[effective[i].Name]; ok {
					effective[i].Enabled = v
				} else if effective[i].Source == capabilities.SourceBuiltin {
					effective[i].Enabled = false
				}
			}
		}
	}
	return Context{Prompt: prompt, Tools: effectiveTools(effective), PolicySummary: map[string]any{"approval_mode": rt.Overlay.Policy.ApprovalMode, "allowed_network": rt.Overlay.Policy.AllowedNetwork}, Resources: rt.Overlay.Resources}
}
