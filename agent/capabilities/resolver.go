package capabilities

import "fmt"

func Resolve(reg *Registry, name string) (Capability, error) {
	c, err := reg.FindByName(name)
	if err != nil {
		return Capability{}, err
	}
	if !c.Enabled {
		return Capability{}, fmt.Errorf("capability disabled: %s", name)
	}
	return c, nil
}
