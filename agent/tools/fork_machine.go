package tools

func ForkMachineSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{
		Name:        "fork_machine",
		Description: "Fork from a snapshot (maps to vpc machine fork).",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"snapshot_id": map[string]any{"type": "string"},
		}, "required": []string{"snapshot_id"}},
	}}
}
