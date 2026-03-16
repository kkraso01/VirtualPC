package controller

import (
	"context"
	"fmt"
	"os"
	"strings"

	"virtualpc/agent/providers"
	"virtualpc/agent/tools"
)

type Planner struct {
	provider providers.Provider
}

func NewPlanner(provider providers.Provider) *Planner { return &Planner{provider: provider} }

func (p *Planner) NextAction(ctx context.Context, systemPrompt string, session *Session) (tools.ToolCall, bool, error) {
	if p.provider == nil {
		return fallbackCall(session)
	}
	messages := []providers.Message{{Role: "user", Content: fmt.Sprintf("Goal: %s", session.Goal)}}
	for _, r := range session.History {
		messages = append(messages, providers.Message{Role: "tool", Content: fmt.Sprintf("%s => %s", r.Tool, r.Output)})
	}
	resp, err := p.provider.GeneratePlan(ctx, systemPrompt, messages, tools.Catalog())
	if err != nil {
		return tools.ToolCall{}, false, err
	}
	call, done, err := p.provider.ExtractToolCall(resp)
	if err != nil {
		return tools.ToolCall{}, false, err
	}
	if done {
		return tools.ToolCall{}, true, nil
	}
	if call.Name == "" {
		return fallbackCall(session)
	}
	return call, false, nil
}

func fallbackCall(session *Session) (tools.ToolCall, bool, error) {
	if os.Getenv("VPC_AGENT_ALLOW_FALLBACK") == "0" {
		return tools.ToolCall{}, true, nil
	}
	if len(session.CommandHistory) > 0 {
		return tools.ToolCall{}, true, nil
	}
	cmd := "echo 'agent controller connected'"
	if strings.Contains(strings.ToLower(session.Goal), "test") {
		cmd = "cd /workspace && make test"
	}
	return tools.ToolCall{Name: "run_command", Arguments: map[string]any{"machine_id": session.MachineID, "command": cmd}, Reason: "fallback planner"}, false, nil
}
