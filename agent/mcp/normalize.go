package mcp

type CapabilityView struct {
	ID          string
	Name        string
	Kind        string
	Description string
	Schema      map[string]any
	Mode        string
}

func Normalize(server ServerConfig, d Discovery) []CapabilityView {
	out := []CapabilityView{}
	for _, t := range d.Tools {
		out = append(out, CapabilityView{ID: "mcp." + server.Name + ".tool." + t.Name, Name: t.Name, Kind: "tool", Description: t.Description, Schema: t.InputSchema, Mode: server.Mode})
	}
	for _, r := range d.Resources {
		out = append(out, CapabilityView{ID: "mcp." + server.Name + ".resource." + r.Name, Name: r.Name, Kind: "resource", Description: r.Description, Mode: server.Mode})
	}
	for _, p := range d.Prompts {
		out = append(out, CapabilityView{ID: "mcp." + server.Name + ".prompt." + p.Name, Name: p.Name, Kind: "prompt", Description: p.Description, Mode: server.Mode})
	}
	return out
}
