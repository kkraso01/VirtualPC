package providers

import "virtualpc/agent/tools"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Provider interface {
	NextToolCall(systemPrompt string, messages []Message, toolSchemas []tools.ToolSchema) (tools.ToolCall, bool, error)
}
