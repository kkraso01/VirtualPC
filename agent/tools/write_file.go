package tools

func WriteFileSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{
		Name:        "write_file",
		Description: "Write or upload a local file to a machine path (maps to vpc machine cp-to).",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"machine_id": map[string]any{"type": "string"},
			"local_src":  map[string]any{"type": "string"},
			"remote_dst": map[string]any{"type": "string"},
		}, "required": []string{"machine_id", "local_src", "remote_dst"}},
	}}
}
