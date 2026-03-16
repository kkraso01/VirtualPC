package capabilities

import "time"

type CapabilityType string

type CapabilitySource string

type ExecutionLocation string

const (
	TypeTool     CapabilityType = "tool"
	TypeResource CapabilityType = "resource"
	TypePrompt   CapabilityType = "prompt"
	TypeSkill    CapabilityType = "skill"
	TypeProvider CapabilityType = "provider"

	SourceBuiltin CapabilitySource = "builtin"
	SourceSkill   CapabilitySource = "skill"
	SourceMCP     CapabilitySource = "mcp"
	SourceCustom  CapabilitySource = "custom"

	LocationController ExecutionLocation = "controller"
	LocationVM         ExecutionLocation = "vm"
	LocationSidecar    ExecutionLocation = "sidecar"
	LocationRemote     ExecutionLocation = "remote"
)

type PolicyRequirements struct {
	ApprovalsRequired []string `json:"approvals_required" yaml:"approvals_required"`
	AllowedPaths      []string `json:"allowed_paths" yaml:"allowed_paths"`
	NetworkMode       string   `json:"network_mode" yaml:"network_mode"`
	RateLimitPerMin   int      `json:"rate_limit_per_min" yaml:"rate_limit_per_min"`
	BudgetTokens      int      `json:"budget_tokens" yaml:"budget_tokens"`
}

type Capability struct {
	ID                string             `json:"id" yaml:"id"`
	Name              string             `json:"name" yaml:"name"`
	Type              CapabilityType     `json:"type" yaml:"type"`
	Source            CapabilitySource   `json:"source" yaml:"source"`
	Schema            map[string]any     `json:"schema,omitempty" yaml:"schema,omitempty"`
	Description       string             `json:"description,omitempty" yaml:"description,omitempty"`
	ExecutionLocation ExecutionLocation  `json:"execution_location" yaml:"execution_location"`
	Policy            PolicyRequirements `json:"policy" yaml:"policy"`
	ApprovalRequired  bool               `json:"approval_required" yaml:"approval_required"`
	NetworkRequired   bool               `json:"network_required" yaml:"network_required"`
	FilesystemScope   []string           `json:"filesystem_scope,omitempty" yaml:"filesystem_scope,omitempty"`
	Enabled           bool               `json:"enabled" yaml:"enabled"`
	Metadata          map[string]any     `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type ProviderProfile struct {
	Name                      string `json:"name" yaml:"name"`
	Provider                  string `json:"provider" yaml:"provider"`
	Model                     string `json:"model" yaml:"model"`
	BaseURL                   string `json:"base_url" yaml:"base_url"`
	APIKeyEnv                 string `json:"api_key_env" yaml:"api_key_env"`
	SupportsChatCompletions   bool   `json:"supports_chat_completions" yaml:"supports_chat_completions"`
	SupportsToolCalling       bool   `json:"supports_tool_calling" yaml:"supports_tool_calling"`
	SupportsResponsesAPI      bool   `json:"supports_responses_api" yaml:"supports_responses_api"`
	SupportsStatefulResponses bool   `json:"supports_stateful_responses" yaml:"supports_stateful_responses"`
	SupportsParallelToolCalls bool   `json:"supports_parallel_tool_calls" yaml:"supports_parallel_tool_calls"`
}

type MCPEndpoint struct {
	Name    string   `json:"name" yaml:"name"`
	Mode    string   `json:"mode" yaml:"mode"`
	Command string   `json:"command,omitempty" yaml:"command,omitempty"`
	Args    []string `json:"args,omitempty" yaml:"args,omitempty"`
	URL     string   `json:"url,omitempty" yaml:"url,omitempty"`
}

type RegistrySnapshot struct {
	GeneratedAt  time.Time    `json:"generated_at"`
	Capabilities []Capability `json:"capabilities"`
}
