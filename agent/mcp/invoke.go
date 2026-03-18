package mcp

import (
	"context"
	"fmt"
	"time"
)

const (
	defaultInvocationTimeout = 15 * time.Second
	defaultMaxPayloadBytes   = 1 << 20
)

func (r *Runtime) InvokeTool(ctx context.Context, server, name string, args map[string]any) (string, error) {
	s, err := r.server(server)
	if err != nil {
		return "", fmt.Errorf("mcp server unavailable: %w", err)
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultInvocationTimeout)
	defer cancel()
	out, err := r.client.InvokeTool(callCtx, s, name, args)
	if err != nil {
		return "", fmt.Errorf("mcp tool execution failed: %w", err)
	}
	if len(out) > defaultMaxPayloadBytes {
		return "", fmt.Errorf("mcp response exceeds size limit")
	}
	return out, nil
}
func (r *Runtime) FetchResource(ctx context.Context, server, name string, args map[string]any) (string, error) {
	s, err := r.server(server)
	if err != nil {
		return "", fmt.Errorf("mcp server unavailable: %w", err)
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultInvocationTimeout)
	defer cancel()
	out, err := r.client.FetchResource(callCtx, s, name, args)
	if err != nil {
		return "", fmt.Errorf("mcp resource fetch failed: %w", err)
	}
	if len(out) > defaultMaxPayloadBytes {
		return "", fmt.Errorf("mcp response exceeds size limit")
	}
	return out, nil
}
func (r *Runtime) FetchPrompt(ctx context.Context, server, name string, args map[string]any) (string, error) {
	s, err := r.server(server)
	if err != nil {
		return "", fmt.Errorf("mcp server unavailable: %w", err)
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultInvocationTimeout)
	defer cancel()
	out, err := r.client.FetchPrompt(callCtx, s, name, args)
	if err != nil {
		return "", fmt.Errorf("mcp prompt fetch failed: %w", err)
	}
	if len(out) > defaultMaxPayloadBytes {
		return "", fmt.Errorf("mcp response exceeds size limit")
	}
	return out, nil
}
