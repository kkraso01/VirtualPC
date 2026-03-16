package tools

func StartServiceSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{
		Name:        "start_service",
		Description: "Start/create a service on a machine through VirtualPC service APIs.",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"machine_id": map[string]any{"type": "string"},
			"name":       map[string]any{"type": "string"},
			"image":      map[string]any{"type": "string"},
		}, "required": []string{"machine_id", "name", "image"}},
	}}
}
