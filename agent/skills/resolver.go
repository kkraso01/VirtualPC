package skills

func ResolveByName(manifests []SkillManifest, names []string) []SkillManifest {
	out := []SkillManifest{}
	for _, n := range names {
		for _, m := range manifests {
			if m.Name == n {
				out = append(out, m)
			}
		}
	}
	return out
}
