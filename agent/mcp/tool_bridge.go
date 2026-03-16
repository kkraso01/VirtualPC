package mcp

func ToolCapabilities(server ServerConfig, d Discovery) []CapabilityView {
	return Normalize(server, Discovery{Tools: d.Tools})
}
