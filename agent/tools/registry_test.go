package tools

import "testing"

func TestRegistryValidation(t *testing.T) {
	r := NewRegistry()
	call := ToolCall{Name: "run_command", Arguments: map[string]any{"machine_id": "m1", "command": "ls"}}
	if err := r.Validate(call); err != nil {
		t.Fatal(err)
	}
	bad := ToolCall{Name: "run_command", Arguments: map[string]any{"machine_id": 5, "command": "ls"}}
	if err := r.Validate(bad); err == nil {
		t.Fatal("expected type error")
	}
}
