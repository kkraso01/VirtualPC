package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"virtualpc/agent/tools"
)

type Anthropic struct {
	Model string
	Key   string
	URL   string
}

func NewAnthropic(model string) *Anthropic {
	return &Anthropic{Model: model, Key: os.Getenv("ANTHROPIC_API_KEY"), URL: "https://api.anthropic.com/v1/messages"}
}

func (a *Anthropic) NextToolCall(systemPrompt string, messages []Message, toolSchemas []tools.ToolSchema) (tools.ToolCall, bool, error) {
	if a.Key == "" {
		return tools.ToolCall{}, true, nil
	}
	in := map[string]any{"model": a.Model, "max_tokens": 512, "system": systemPrompt, "messages": []map[string]string{}, "tools": toolSchemas}
	for _, m := range messages {
		in["messages"] = append(in["messages"].([]map[string]string), map[string]string{"role": m.Role, "content": m.Content})
	}
	b, _ := json.Marshal(in)
	req, _ := http.NewRequest("POST", a.URL, bytes.NewReader(b))
	req.Header.Set("x-api-key", a.Key)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return tools.ToolCall{}, false, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return tools.ToolCall{}, false, fmt.Errorf("anthropic error status: %d", res.StatusCode)
	}
	var out struct {
		Content []struct {
			Type  string         `json:"type"`
			Name  string         `json:"name"`
			Input map[string]any `json:"input"`
		} `json:"content"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return tools.ToolCall{}, false, err
	}
	for _, c := range out.Content {
		if c.Type == "tool_use" {
			return tools.ToolCall{Name: c.Name, Arguments: c.Input}, false, nil
		}
	}
	return tools.ToolCall{}, true, nil
}
