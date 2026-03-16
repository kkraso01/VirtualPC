package mcp

import (
	"context"
	"fmt"
)

type NoopInvoker struct{}

func (NoopInvoker) InvokeTool(_ context.Context, server ServerConfig, name string, _ map[string]any) (string, error) {
	return fmt.Sprintf("mcp-stdio tool %s/%s invoked", server.Name, name), nil
}
func (NoopInvoker) FetchResource(_ context.Context, server ServerConfig, name string, _ map[string]any) (string, error) {
	return fmt.Sprintf("mcp-stdio resource %s/%s", server.Name, name), nil
}
func (NoopInvoker) FetchPrompt(_ context.Context, server ServerConfig, name string, _ map[string]any) (string, error) {
	return fmt.Sprintf("mcp-stdio prompt %s/%s", server.Name, name), nil
}
