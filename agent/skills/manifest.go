package skills

type ToolBinding struct {
	Name    string         `yaml:"name" json:"name"`
	Enabled bool           `yaml:"enabled" json:"enabled"`
	Config  map[string]any `yaml:"config" json:"config"`
}

type SkillPolicies struct {
	BudgetDefaults struct {
		MaxIterations int `yaml:"max_iterations" json:"max_iterations"`
	} `yaml:"budget_defaults" json:"budget_defaults"`
	ApprovalMode      string   `yaml:"approval_mode" json:"approval_mode"`
	AllowedPaths      []string `yaml:"allowed_paths" json:"allowed_paths"`
	AllowedNetwork    string   `yaml:"allowed_network" json:"allowed_network"`
	AllowedRegistries []string `yaml:"allowed_registries" json:"allowed_registries"`
	AllowedImages     []string `yaml:"allowed_images" json:"allowed_images"`
}

type SkillManifest struct {
	Name        string        `json:"name"`
	Path        string        `json:"path"`
	Description string        `json:"description"`
	Purpose     string        `json:"purpose"`
	Workflows   []string      `json:"workflows"`
	Prompt      string        `json:"prompt"`
	Tools       []ToolBinding `json:"tools"`
	Policies    SkillPolicies `json:"policies"`
	Resources   []string      `json:"resources"`
}
