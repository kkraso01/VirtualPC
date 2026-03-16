package mcp

import "context"

func DiscoverAll(ctx context.Context, c Client, servers []ServerConfig) ([]Discovery, error) {
	out := make([]Discovery, 0, len(servers))
	for _, s := range servers {
		d, err := c.Discover(ctx, s)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}
