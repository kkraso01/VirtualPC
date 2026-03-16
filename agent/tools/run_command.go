package tools

func RunCommandSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{
		Name:        "run_command",
		Description: "Run a shell command inside a VirtualPC machine (maps to vpc machine exec).",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"machine_id": map[string]any{"type": "string"},
			"command":    map[string]any{"type": "string"},
		}, "required": []string{"machine_id", "command"}},
	}}
}
