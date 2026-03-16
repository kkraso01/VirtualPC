package tools

type ToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
	Reason    string         `json:"reason,omitempty"`
}

type ToolResult struct {
	Tool    string `json:"tool"`
	Output  string `json:"output"`
	Success bool   `json:"success"`
}

type ToolSchema struct {
	Type     string         `json:"type"`
	Function SchemaFunction `json:"function"`
}

type SchemaFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}
