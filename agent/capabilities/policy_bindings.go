package capabilities

import "virtualpc/agent/config"

func ApplyPolicyBindings(cfg *config.Config, caps []Capability) {
	if cfg.ApprovalRequiredTools == nil {
		cfg.ApprovalRequiredTools = map[string]bool{}
	}
	for _, c := range caps {
		if c.Type != TypeTool || !c.Enabled {
			continue
		}
		if c.ApprovalRequired || len(c.Policy.ApprovalsRequired) > 0 {
			cfg.ApprovalRequiredTools[c.Name] = true
		}
		if len(c.FilesystemScope) > 0 {
			cfg.ExtraWritableRoot = c.FilesystemScope[0]
		}
	}
}
