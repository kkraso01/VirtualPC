package capabilities

func Merge(groups ...[]Capability) []Capability {
	idx := map[string]Capability{}
	for _, g := range groups {
		for _, c := range g {
			idx[c.ID] = c
		}
	}
	out := make([]Capability, 0, len(idx))
	for _, c := range idx {
		out = append(out, c)
	}
	return out
}
