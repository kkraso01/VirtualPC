package tools

func ReadFileSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{
		Name:        "read_file",
		Description: "Download a file from machine to local path (maps to vpc machine cp-from).",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"machine_id": map[string]any{"type": "string"},
			"remote_src": map[string]any{"type": "string"},
			"local_dst":  map[string]any{"type": "string"},
		}, "required": []string{"machine_id", "remote_src", "local_dst"}},
	}}
}
