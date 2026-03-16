package skills

type Overlay struct {
	PromptFragments []string
	ToolAllow       map[string]bool
	Policy          SkillPolicies
	Resources       []string
}

func BuildOverlay(selected []string, manifests []SkillManifest) Overlay {
	ov := Overlay{ToolAllow: map[string]bool{}}
	for _, name := range selected {
		for _, m := range manifests {
			if m.Name != name {
				continue
			}
			if m.Prompt != "" {
				ov.PromptFragments = append(ov.PromptFragments, m.Prompt)
			}
			for _, t := range m.Tools {
				ov.ToolAllow[t.Name] = t.Enabled
			}
			ov.Resources = append(ov.Resources, m.Resources...)
			ov.Policy = m.Policies
		}
	}
	return ov
}
