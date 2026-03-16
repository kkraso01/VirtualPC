package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"virtualpc/agent/tools"
)

type OpenAI struct {
	Model string
	Key   string
	URL   string
}

func NewOpenAI(model string) *OpenAI {
	return &OpenAI{Model: model, Key: os.Getenv("OPENAI_API_KEY"), URL: "https://api.openai.com/v1/chat/completions"}
}

func (o *OpenAI) NextToolCall(systemPrompt string, messages []Message, toolSchemas []tools.ToolSchema) (tools.ToolCall, bool, error) {
	if o.Key == "" {
		return tools.ToolCall{}, true, nil
	}
	in := map[string]any{"model": o.Model, "messages": []map[string]string{{"role": "system", "content": systemPrompt}}, "tools": toolSchemas, "tool_choice": "auto"}
	for _, m := range messages {
		in["messages"] = append(in["messages"].([]map[string]string), map[string]string{"role": m.Role, "content": m.Content})
	}
	b, _ := json.Marshal(in)
	req, _ := http.NewRequest("POST", o.URL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+o.Key)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return tools.ToolCall{}, false, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return tools.ToolCall{}, false, fmt.Errorf("openai error status: %d", res.StatusCode)
	}
	var out struct {
		Choices []struct {
			Message struct {
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
		return tools.ToolCall{}, false, err
	}
	if len(out.Choices) == 0 || len(out.Choices[0].Message.ToolCalls) == 0 {
		return tools.ToolCall{}, true, nil
	}
	call := out.Choices[0].Message.ToolCalls[0].Function
	args := map[string]any{}
	_ = json.Unmarshal([]byte(call.Arguments), &args)
	return tools.ToolCall{Name: call.Name, Arguments: args}, false, nil
}
