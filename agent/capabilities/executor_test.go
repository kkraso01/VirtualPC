package capabilities

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type noOpBuiltin struct{}

func (noOpBuiltin) ExecuteBuiltin(_ string, _ map[string]any) (string, error) { return "", nil }

type noOpMCP struct{}

func (noOpMCP) InvokeTool(_ context.Context, _, _ string, _ map[string]any) (string, error) {
	return "", nil
}
func (noOpMCP) FetchResource(_ context.Context, _, _ string, _ map[string]any) (string, error) {
	return "", nil
}
func (noOpMCP) FetchPrompt(_ context.Context, _, _ string, _ map[string]any) (string, error) {
	return "", nil
}

func TestLocalToolTimeout(t *testing.T) {
	t.Parallel()
	e := NewExecutor(noOpBuiltin{}, noOpMCP{})
	cap := Capability{ID: "custom.tool.sleep", Name: "sleep", Metadata: map[string]any{"backend_type": "local", "command": "/bin/sh", "args_template": []any{"-c", "sleep 2"}, "timeout_seconds": 1}}
	_, err := e.runLocal(context.Background(), cap, map[string]any{})
	if err == nil || err.Error() != "local tool timed out after 1s" {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestHTTPToolAllowlistBlocked(t *testing.T) {
	t.Parallel()
	reg := NewRegistry([]Capability{{ID: "custom.tool.http", Name: "http_tool", Type: TypeTool, Source: SourceCustom, Enabled: true, NetworkRequired: true, Metadata: map[string]any{"backend_type": "http", "url": "https://example.com", "allowed_hosts": []any{"api.test.local"}}}})
	d := NewDispatcher(reg, NewExecutor(noOpBuiltin{}, noOpMCP{}), nil, NewAuditor(t.TempDir()+"/audit.log"))
	res, err := d.Dispatch(context.Background(), ExecutionRequest{SessionID: "s1", Name: "http_tool", Args: map[string]any{}})
	if err == nil || !res.Blocked {
		t.Fatalf("expected blocked response, got err=%v res=%+v", err, res)
	}
}

func TestHTTPStructuredErrorIncludesBody(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request"))
	}))
	defer ts.Close()
	e := NewExecutor(noOpBuiltin{}, noOpMCP{})
	cap := Capability{ID: "custom.tool.http", Name: "http", Metadata: map[string]any{"backend_type": "http", "url": ts.URL, "method": "POST"}}
	_, err := e.runHTTP(context.Background(), cap, map[string]any{"q": "x"})
	if err == nil || err.Error() != "http status 400: invalid request" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInterpolateStrictFailsOnMissingVariable(t *testing.T) {
	t.Parallel()
	_, err := interpolateStrict("{{missing}}", map[string]any{"x": 1})
	if err == nil || !strings.Contains(err.Error(), "template interpolation failed") {
		t.Fatalf("expected interpolation error")
	}
}

func TestRunHTTPCancelsByContext(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		_, _ = w.Write([]byte("late"))
	}))
	defer ts.Close()
	e := NewExecutor(noOpBuiltin{}, noOpMCP{})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	cap := Capability{ID: "custom.tool.http", Name: "http", Metadata: map[string]any{"backend_type": "http", "url": ts.URL, "method": "GET"}}
	_, err := e.runHTTP(ctx, cap, map[string]any{})
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
