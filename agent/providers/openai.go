package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"virtualpc/agent/tools"
)

type OpenAI struct {
	Model  string
	client openai.Client
	caps   Capabilities
}

func NewOpenAI(model string) *OpenAI {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAI{Model: model, client: client, caps: Capabilities{SupportsResponsesAPI: true, SupportsChatCompletions: true, SupportsToolCalling: true, SupportsStatefulResponses: true}}
}

func (o *OpenAI) Name() string               { return "openai" }
func (o *OpenAI) Capabilities() Capabilities { return o.caps }

func (o *OpenAI) GeneratePlan(ctx context.Context, systemPrompt string, messages []Message, toolSchemas []tools.ToolSchema) (Response, error) {
	params := openai.ChatCompletionNewParams{Model: o.Model, Messages: []openai.ChatCompletionMessageParamUnion{openai.SystemMessage(systemPrompt)}}
	for _, m := range messages {
		params.Messages = append(params.Messages, openAIMessage(m))
	}
	for _, s := range toolSchemas {
		params.Tools = append(params.Tools, openai.ChatCompletionToolParam{Type: "function", Function: openai.FunctionDefinitionParam{Name: s.Function.Name, Description: openai.String(s.Function.Description), Parameters: s.Function.Parameters}})
	}
	resp, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return Response{}, err
	}
	if len(resp.Choices) == 0 {
		return Response{Provider: o.Name(), Done: true}, nil
	}
	choice := resp.Choices[0].Message
	if len(choice.ToolCalls) == 0 {
		return Response{Provider: o.Name(), Done: true, RawText: choice.Content}, nil
	}
	args := map[string]any{}
	if err := json.Unmarshal([]byte(choice.ToolCalls[0].Function.Arguments), &args); err != nil {
		return Response{}, fmt.Errorf("decode tool args: %w", err)
	}
	return Response{Provider: o.Name(), ToolCall: tools.ToolCall{Name: choice.ToolCalls[0].Function.Name, Arguments: args}, Done: false, RawText: choice.Content}, nil
}

func (o *OpenAI) ExtractToolCall(response Response) (tools.ToolCall, bool, error) {
	if response.Done {
		return tools.ToolCall{}, true, nil
	}
	if response.ToolCall.Name == "" {
		return tools.ToolCall{}, true, nil
	}
	return response.ToolCall, false, nil
}

func openAIMessage(m Message) openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case "system":
		return openai.SystemMessage(m.Content)
	case "assistant":
		return openai.AssistantMessage(m.Content)
	case "tool":
		return openai.ToolMessage(m.Content, "tool_result")
	default:
		return openai.UserMessage(m.Content)
	}
}
