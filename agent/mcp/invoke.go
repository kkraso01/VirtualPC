package mcp

import "context"

func (r *Runtime) InvokeTool(ctx context.Context, server, name string, args map[string]any) (string, error) {
	s, err := r.server(server)
	if err != nil {
		return "", err
	}
	return r.client.InvokeTool(ctx, s, name, args)
}
func (r *Runtime) FetchResource(ctx context.Context, server, name string, args map[string]any) (string, error) {
	s, err := r.server(server)
	if err != nil {
		return "", err
	}
	return r.client.FetchResource(ctx, s, name, args)
}
func (r *Runtime) FetchPrompt(ctx context.Context, server, name string, args map[string]any) (string, error) {
	s, err := r.server(server)
	if err != nil {
		return "", err
	}
	return r.client.FetchPrompt(ctx, s, name, args)
}
