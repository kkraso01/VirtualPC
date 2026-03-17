package mcp

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeInvoker struct {
	sleep time.Duration
	err   error
	resp  string
}

func (f fakeInvoker) InvokeTool(ctx context.Context, _ ServerConfig, _ string, _ map[string]any) (string, error) {
	if f.sleep > 0 {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(f.sleep):
		}
	}
	if f.err != nil {
		return "", f.err
	}
	return f.resp, nil
}
func (f fakeInvoker) FetchResource(ctx context.Context, s ServerConfig, n string, a map[string]any) (string, error) {
	return f.InvokeTool(ctx, s, n, a)
}
func (f fakeInvoker) FetchPrompt(ctx context.Context, s ServerConfig, n string, a map[string]any) (string, error) {
	return f.InvokeTool(ctx, s, n, a)
}

func TestInvokeUnavailableServer(t *testing.T) {
	t.Parallel()
	r := NewRuntime(nil, fakeInvoker{})
	_, err := r.InvokeTool(context.Background(), "missing", "tool", nil)
	if err == nil {
		t.Fatal("expected unavailable server error")
	}
}

func TestInvokeTimeout(t *testing.T) {
	t.Parallel()
	r := NewRuntime([]ServerConfig{{Name: "x"}}, fakeInvoker{sleep: 16 * time.Second})
	_, err := r.InvokeTool(context.Background(), "x", "tool", nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestInvokeFailureWrapped(t *testing.T) {
	t.Parallel()
	r := NewRuntime([]ServerConfig{{Name: "x"}}, fakeInvoker{err: errors.New("boom")})
	_, err := r.InvokeTool(context.Background(), "x", "tool", nil)
	if err == nil {
		t.Fatal("expected wrapped execution error")
	}
}
