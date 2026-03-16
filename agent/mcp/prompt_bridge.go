package mcp

func PromptCapabilities(server ServerConfig, d Discovery) []CapabilityView {
	return Normalize(server, Discovery{Prompts: d.Prompts})
}
