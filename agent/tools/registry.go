package tools

import (
	"fmt"
)

type Registry struct{ byName map[string]ToolSchema }

func NewRegistry() *Registry {
	r := &Registry{byName: map[string]ToolSchema{}}
	for _, s := range Catalog() {
		r.byName[s.Function.Name] = s
	}
	return r
}

func (r *Registry) Validate(call ToolCall) error {
	schema, ok := r.byName[call.Name]
	if !ok {
		return fmt.Errorf("unknown tool: %s", call.Name)
	}
	params, _ := schema.Function.Parameters["properties"].(map[string]any)
	required, _ := schema.Function.Parameters["required"].([]string)
	for _, req := range required {
		if _, ok := call.Arguments[req]; !ok {
			return fmt.Errorf("missing required argument %s for %s", req, call.Name)
		}
	}
	for k, v := range call.Arguments {
		p, ok := params[k].(map[string]any)
		if !ok {
			continue
		}
		typeName, _ := p["type"].(string)
		if !matchesType(typeName, v) {
			return fmt.Errorf("invalid argument type for %s.%s", call.Name, k)
		}
	}
	return nil
}

func matchesType(typeName string, v any) bool {
	switch typeName {
	case "string":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "number", "integer":
		switch v.(type) {
		case int, int64, float64, float32:
			return true
		}
		return false
	default:
		return true
	}
}
