package providers

import (
	"context"
	"fmt"

	"virtualpc/agent/tools"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Capabilities struct {
	SupportsResponsesAPI      bool `json:"supports_responses_api"`
	SupportsChatCompletions   bool `json:"supports_chat_completions"`
	SupportsToolCalling       bool `json:"supports_tool_calling"`
	SupportsStatefulResponses bool `json:"supports_stateful_responses"`
}

type Response struct {
	Provider string         `json:"provider"`
	Done     bool           `json:"done"`
	ToolCall tools.ToolCall `json:"tool_call"`
	RawText  string         `json:"raw_text,omitempty"`
}

type Provider interface {
	GeneratePlan(ctx context.Context, systemPrompt string, messages []Message, toolSchemas []tools.ToolSchema) (Response, error)
	ExtractToolCall(response Response) (tools.ToolCall, bool, error)
	Capabilities() Capabilities
	Name() string
}

func EnsureToolCallingAvailable(p Provider) error {
	if p == nil {
		return nil
	}
	if !p.Capabilities().SupportsToolCalling {
		return fmt.Errorf("provider %s does not support tool calling", p.Name())
	}
	return nil
}
