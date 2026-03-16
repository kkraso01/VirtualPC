package capabilities

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeBuiltin struct{}

func (fakeBuiltin) ExecuteBuiltin(name string, _ map[string]any) (string, error) {
	return "ok:" + name, nil
}

type fakeMCP struct{}

func (fakeMCP) InvokeTool(_ context.Context, server, name string, _ map[string]any) (string, error) {
	return server + ":" + name, nil
}
func (fakeMCP) FetchResource(_ context.Context, server, name string, _ map[string]any) (string, error) {
	return "res:" + server + ":" + name, nil
}
func (fakeMCP) FetchPrompt(_ context.Context, server, name string, _ map[string]any) (string, error) {
	return "prompt:" + server + ":" + name, nil
}

type fakeApproval struct{ allow bool }

func (f fakeApproval) Require(_ context.Context, _, _, _ string, _ map[string]any) error {
	if f.allow {
		return nil
	}
	return context.DeadlineExceeded
}

func TestDispatchBuiltin(t *testing.T) {
	reg := NewRegistry([]Capability{{ID: "builtin.tool.run_command", Name: "run_command", Type: TypeTool, Source: SourceBuiltin, Enabled: true}})
	d := NewDispatcher(reg, NewExecutor(fakeBuiltin{}, fakeMCP{}), nil, nil)
	res, err := d.Dispatch(context.Background(), ExecutionRequest{SessionID: "s1", Name: "run_command", Args: map[string]any{}})
	if err != nil || !res.Success {
		t.Fatalf("dispatch failed: %v %#v", err, res)
	}
}

func TestDispatchHTTPWithAllowlist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) }))
	defer ts.Close()
	reg := NewRegistry([]Capability{{ID: "custom.tool.docs", Name: "docs", Type: TypeTool, Source: SourceCustom, Enabled: true, NetworkRequired: true, Metadata: map[string]any{"backend_type": "http", "url": ts.URL, "allowed_hosts": []any{"127.0.0.1"}}}})
	d := NewDispatcher(reg, NewExecutor(fakeBuiltin{}, fakeMCP{}), nil, nil)
	res, err := d.Dispatch(context.Background(), ExecutionRequest{SessionID: "s1", Name: "docs", Args: map[string]any{"q": "x"}})
	if err != nil || !res.Success {
		t.Fatalf("http failed %v %#v", err, res)
	}
}

func TestDispatchMCPPrompt(t *testing.T) {
	reg := NewRegistry([]Capability{{ID: "mcp.github.prompt.repo", Name: "repo", Type: TypePrompt, Source: SourceMCP, Enabled: true}})
	d := NewDispatcher(reg, NewExecutor(fakeBuiltin{}, fakeMCP{}), nil, nil)
	res, err := d.Dispatch(context.Background(), ExecutionRequest{SessionID: "s1", Name: "repo", Args: map[string]any{}})
	if err != nil || res.Output == "" {
		t.Fatalf("mcp prompt failed: %v %#v", err, res)
	}
}

func TestApprovalRequiredFlow(t *testing.T) {
	reg := NewRegistry([]Capability{{ID: "custom.tool.deploy", Name: "deploy", Type: TypeTool, Source: SourceCustom, Enabled: true, ApprovalRequired: true, Metadata: map[string]any{"backend_type": "local", "command": "/bin/echo"}}})
	d := NewDispatcher(reg, NewExecutor(fakeBuiltin{}, fakeMCP{}), fakeApproval{allow: false}, nil)
	res, err := d.Dispatch(context.Background(), ExecutionRequest{SessionID: "s1", Name: "deploy", Args: map[string]any{}})
	if err == nil || !res.Blocked {
		t.Fatalf("expected blocked approval")
	}
}
