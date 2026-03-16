package mcp

func ResourceCapabilities(server ServerConfig, d Discovery) []CapabilityView {
	return Normalize(server, Discovery{Resources: d.Resources})
}
