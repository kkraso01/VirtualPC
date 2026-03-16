package mcp

import "context"

type Discovery struct {
	Tools     []DiscoveredTool
	Resources []DiscoveredResource
	Prompts   []DiscoveredPrompt
}

type DiscoveredTool struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type DiscoveredResource struct {
	Name        string
	Description string
}

type DiscoveredPrompt struct {
	Name        string
	Description string
}

type Client interface {
	Discover(ctx context.Context, server ServerConfig) (Discovery, error)
}

type NoopClient struct{}

func (NoopClient) Discover(_ context.Context, _ ServerConfig) (Discovery, error) {
	return Discovery{}, nil
}
