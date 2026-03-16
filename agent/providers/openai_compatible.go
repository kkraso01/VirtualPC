package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"virtualpc/agent/tools"
)

type OpenAICompatible struct {
	Model   string
	BaseURL string
	APIKey  string
	caps    Capabilities
	client  *http.Client
}

func NewOpenAICompatible(model, baseURL, apiKey string, caps Capabilities) *OpenAICompatible {
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/") + "/v1"
	}
	return &OpenAICompatible{Model: model, BaseURL: baseURL, APIKey: apiKey, caps: caps, client: http.DefaultClient}
}

func (o *OpenAICompatible) Name() string               { return "openai_compatible" }
func (o *OpenAICompatible) Capabilities() Capabilities { return o.caps }

func (o *OpenAICompatible) GeneratePlan(ctx context.Context, systemPrompt string, messages []Message, toolSchemas []tools.ToolSchema) (Response, error) {
	if !o.caps.SupportsToolCalling {
		return Response{}, fmt.Errorf("openai-compatible provider configured without tool calling")
	}
	if !o.caps.SupportsChatCompletions {
		return Response{}, fmt.Errorf("openai-compatible provider requires chat completions capability")
	}
	in := map[string]any{"model": o.Model, "messages": []map[string]string{{"role": "system", "content": systemPrompt}}, "tools": toolSchemas, "tool_choice": "auto"}
	for _, m := range messages {
		in["messages"] = append(in["messages"].([]map[string]string), map[string]string{"role": m.Role, "content": m.Content})
	}
	body, _ := json.Marshal(in)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if o.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.APIKey)
	}
	res, err := o.client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return Response{}, fmt.Errorf("openai-compatible status: %d", res.StatusCode)
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return Response{}, err
	}
	if len(out.Choices) == 0 || len(out.Choices[0].Message.ToolCalls) == 0 {
		return Response{Provider: o.Name(), Done: true, RawText: out.Choices[0].Message.Content}, nil
	}
	args := map[string]any{}
	_ = json.Unmarshal([]byte(out.Choices[0].Message.ToolCalls[0].Function.Arguments), &args)
	return Response{Provider: o.Name(), Done: false, ToolCall: tools.ToolCall{Name: out.Choices[0].Message.ToolCalls[0].Function.Name, Arguments: args}, RawText: out.Choices[0].Message.Content}, nil
}

func (o *OpenAICompatible) ExtractToolCall(response Response) (tools.ToolCall, bool, error) {
	if response.Done || response.ToolCall.Name == "" {
		return tools.ToolCall{}, true, nil
	}
	return response.ToolCall, false, nil
}
