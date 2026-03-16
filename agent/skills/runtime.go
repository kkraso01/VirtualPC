package skills

type RuntimeState struct {
	EffectivePrompt string
	Overlay         Overlay
}

func BuildRuntime(basePrompt string, selected []string, manifests []SkillManifest) RuntimeState {
	ov := BuildOverlay(selected, manifests)
	return RuntimeState{
		EffectivePrompt: ComposePrompt(basePrompt, ov.PromptFragments),
		Overlay:         ov,
	}
}
