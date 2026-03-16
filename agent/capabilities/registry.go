package capabilities

import (
	"fmt"
	"slices"
)

type Registry struct {
	items []Capability
}

func NewRegistry(items []Capability) *Registry {
	return &Registry{items: items}
}

func (r *Registry) All() []Capability { return slices.Clone(r.items) }

func (r *Registry) EnabledTools() []Capability {
	out := []Capability{}
	for _, c := range r.items {
		if c.Type == TypeTool && c.Enabled {
			out = append(out, c)
		}
	}
	return out
}

func (r *Registry) FindByName(name string) (Capability, error) {
	for _, c := range r.items {
		if c.Name == name {
			return c, nil
		}
	}
	return Capability{}, fmt.Errorf("capability %s not found", name)
}
