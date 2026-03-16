package providers

import (
	"context"
	"encoding/json"
	"os"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"virtualpc/agent/tools"
)

type Anthropic struct {
	Model  string
	client anthropic.Client
	caps   Capabilities
}

func NewAnthropic(model string) *Anthropic {
	client := anthropic.NewClient(option.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")))
	return &Anthropic{Model: model, client: client, caps: Capabilities{SupportsChatCompletions: false, SupportsResponsesAPI: false, SupportsToolCalling: true, SupportsStatefulResponses: false}}
}

func (a *Anthropic) Name() string               { return "anthropic" }
func (a *Anthropic) Capabilities() Capabilities { return a.caps }

func (a *Anthropic) GeneratePlan(ctx context.Context, systemPrompt string, messages []Message, toolSchemas []tools.ToolSchema) (Response, error) {
	params := anthropic.MessageNewParams{Model: anthropic.Model(a.Model), MaxTokens: 512, System: []anthropic.TextBlockParam{{Type: "text", Text: systemPrompt}}}
	for _, m := range messages {
		if m.Role == "assistant" {
			params.Messages = append(params.Messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		} else {
			params.Messages = append(params.Messages, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		}
	}
	for _, s := range toolSchemas {
		params.Tools = append(params.Tools, anthropic.ToolUnionParam{OfTool: &anthropic.ToolParam{Name: s.Function.Name, Description: anthropic.String(s.Function.Description), InputSchema: anthropic.ToolInputSchemaParam{Type: "object", Properties: s.Function.Parameters["properties"]}}})
	}
	resp, err := a.client.Messages.New(ctx, params)
	if err != nil {
		return Response{}, err
	}
	for _, c := range resp.Content {
		if c.Type == "tool_use" {
			args := map[string]any{}
			_ = json.Unmarshal(c.Input, &args)
			return Response{Provider: a.Name(), ToolCall: tools.ToolCall{Name: c.Name, Arguments: args}}, nil
		}
	}
	return Response{Provider: a.Name(), Done: true}, nil
}

func (a *Anthropic) ExtractToolCall(response Response) (tools.ToolCall, bool, error) {
	if response.Done || response.ToolCall.Name == "" {
		return tools.ToolCall{}, true, nil
	}
	return response.ToolCall, false, nil
}
