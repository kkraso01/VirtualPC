package controller

import "virtualpc/agent/capabilities"

func effectiveTools(caps []capabilities.Capability) []capabilities.Capability {
	out := []capabilities.Capability{}
	for _, c := range caps {
		if c.Enabled && c.Type == capabilities.TypeTool {
			out = append(out, c)
		}
	}
	return out
}
