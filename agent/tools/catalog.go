package tools

func Catalog() []ToolSchema {
	return []ToolSchema{
		createMachineSchema(),
		startMachineSchema(),
		RunCommandSchema(),
		openShellSchema(),
		WriteFileSchema(),
		ReadFileSchema(),
		uploadFileSchema(),
		downloadFileSchema(),
		StartServiceSchema(),
		stopServiceSchema(),
		SnapshotMachineSchema(),
		ForkMachineSchema(),
		destroyMachineSchema(),
	}
}

func createMachineSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{Name: "create_machine", Description: "Create a machine profile in VirtualPC.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"profile": map[string]any{"type": "string"}}, "required": []string{"profile"}}}}
}
func startMachineSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{Name: "start_machine", Description: "Start a machine in VirtualPC.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"machine_id": map[string]any{"type": "string"}}, "required": []string{"machine_id"}}}}
}
func openShellSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{Name: "open_shell", Description: "Open non-interactive shell endpoint for machine.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"machine_id": map[string]any{"type": "string"}}, "required": []string{"machine_id"}}}}
}
func uploadFileSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{Name: "upload_file", Description: "Alias of write_file using cp-to.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"machine_id": map[string]any{"type": "string"}, "local_src": map[string]any{"type": "string"}, "remote_dst": map[string]any{"type": "string"}}, "required": []string{"machine_id", "local_src", "remote_dst"}}}}
}
func downloadFileSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{Name: "download_file", Description: "Alias of read_file using cp-from.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"machine_id": map[string]any{"type": "string"}, "remote_src": map[string]any{"type": "string"}, "local_dst": map[string]any{"type": "string"}}, "required": []string{"machine_id", "remote_src", "local_dst"}}}}
}
func stopServiceSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{Name: "stop_service", Description: "Stop a service by id.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"service_id": map[string]any{"type": "string"}}, "required": []string{"service_id"}}}}
}
func destroyMachineSchema() ToolSchema {
	return ToolSchema{Type: "function", Function: SchemaFunction{Name: "destroy_machine", Description: "Destroy machine resource in VirtualPC.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"machine_id": map[string]any{"type": "string"}}, "required": []string{"machine_id"}}}}
}
