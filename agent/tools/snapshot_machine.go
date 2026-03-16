package tools

func SnapshotMachineSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{
		Name:        "snapshot_machine",
		Description: "Create a machine snapshot via VirtualPC (maps to vpc snapshot create).",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"machine_id": map[string]any{"type": "string"},
		}, "required": []string{"machine_id"}},
	}}
}
