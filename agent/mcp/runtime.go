package mcp

import (
	"context"
	"fmt"
)

type Runtime struct {
	servers map[string]ServerConfig
	client  Invoker
}

type Invoker interface {
	InvokeTool(ctx context.Context, server ServerConfig, name string, args map[string]any) (string, error)
	FetchResource(ctx context.Context, server ServerConfig, name string, args map[string]any) (string, error)
	FetchPrompt(ctx context.Context, server ServerConfig, name string, args map[string]any) (string, error)
}

func NewRuntime(servers []ServerConfig, inv Invoker) *Runtime {
	idx := map[string]ServerConfig{}
	for _, s := range servers {
		idx[s.Name] = s
	}
	if inv == nil {
		inv = NoopInvoker{}
	}
	return &Runtime{servers: idx, client: inv}
}

func (r *Runtime) server(name string) (ServerConfig, error) {
	s, ok := r.servers[name]
	if !ok {
		return ServerConfig{}, fmt.Errorf("unknown mcp server: %s", name)
	}
	return s, nil
}
