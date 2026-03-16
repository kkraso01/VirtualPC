package capabilities

import (
	"context"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"virtualpc/agent/mcp"
	"virtualpc/agent/skills"
	"virtualpc/agent/tools"
)

type LoaderOptions struct {
	SkillsRoot           string
	LocalToolsRoot       string
	HTTPToolsRoot        string
	ProviderProfilesRoot string
	MCPConfigPath        string
	MCPClient            mcp.Client
}

type CustomToolManifest struct {
	Name              string             `yaml:"name"`
	Description       string             `yaml:"description"`
	Schema            map[string]any     `yaml:"schema"`
	BackendType       string             `yaml:"backend_type"`
	Command           string             `yaml:"command"`
	URL               string             `yaml:"url"`
	ExecutionLocation ExecutionLocation  `yaml:"execution_location"`
	Policy            PolicyRequirements `yaml:"policy"`
	TimeoutSeconds    int                `yaml:"timeout_seconds"`
	Retries           int                `yaml:"retries"`
	Auth              string             `yaml:"auth"`
	Enabled           bool               `yaml:"enabled"`
}

func Load(ctx context.Context, opts LoaderOptions) (*Registry, []skills.SkillManifest, []ProviderProfile, []mcp.ServerConfig, error) {
	builtins := builtinCapabilities()
	skillManifests, _ := skills.LoadAll(opts.SkillsRoot)
	skillCaps := []Capability{}
	for _, s := range skillManifests {
		skillCaps = append(skillCaps, Capability{ID: "skill." + s.Name, Name: s.Name, Type: TypeSkill, Source: SourceSkill, Description: s.Description, ExecutionLocation: LocationController, Enabled: true, Policy: PolicyRequirements{AllowedPaths: s.Policies.AllowedPaths, NetworkMode: s.Policies.AllowedNetwork}})
		for _, t := range s.Tools {
			skillCaps = append(skillCaps, Capability{ID: "skill." + s.Name + ".tool." + t.Name, Name: t.Name, Type: TypeTool, Source: SourceSkill, ExecutionLocation: LocationVM, Enabled: t.Enabled, Metadata: map[string]any{"skill": s.Name, "config": t.Config}})
		}
	}
	customCaps, _ := loadCustomTools(opts.LocalToolsRoot, SourceCustom)
	httpCaps, _ := loadCustomTools(opts.HTTPToolsRoot, SourceCustom)
	profiles, _ := loadProviderProfiles(opts.ProviderProfilesRoot)
	providerCaps := []Capability{}
	for _, p := range profiles {
		providerCaps = append(providerCaps, Capability{ID: "provider." + p.Name, Name: p.Name, Type: TypeProvider, Source: SourceCustom, Description: p.Provider + " profile", ExecutionLocation: LocationController, Enabled: true, Metadata: map[string]any{"profile": p}})
	}
	servers, _ := mcp.LoadConfig(opts.MCPConfigPath)
	mcpCaps := []Capability{}
	if opts.MCPClient == nil {
		opts.MCPClient = mcp.NoopClient{}
	}
	for _, s := range servers {
		d, _ := opts.MCPClient.Discover(ctx, s)
		mcpCaps = append(mcpCaps, NormalizeMCP(s, d)...)
	}
	all := Merge(builtins, skillCaps, customCaps, httpCaps, providerCaps, mcpCaps)
	return NewRegistry(all), skillManifests, profiles, servers, nil
}

func NormalizeMCP(s mcp.ServerConfig, d mcp.Discovery) []Capability {
	v := mcp.Normalize(s, d)
	out := []Capability{}
	for _, c := range v {
		typeName := TypePrompt
		switch c.Kind {
		case "tool":
			typeName = TypeTool
		case "resource":
			typeName = TypeResource
		}
		loc := LocationSidecar
		if c.Mode == "remote" {
			loc = LocationRemote
		}
		out = append(out, Capability{ID: c.ID, Name: c.Name, Type: typeName, Source: SourceMCP, Description: c.Description, Schema: c.Schema, ExecutionLocation: loc, Enabled: true, NetworkRequired: c.Mode == "remote"})
	}
	return out
}

func builtinCapabilities() []Capability {
	out := []Capability{}
	for _, t := range tools.Catalog() {
		out = append(out, Capability{ID: "builtin.tool." + t.Function.Name, Name: t.Function.Name, Type: TypeTool, Source: SourceBuiltin, Schema: t.Function.Parameters, Description: t.Function.Description, ExecutionLocation: LocationVM, Enabled: true})
	}
	return out
}

func loadCustomTools(root string, source CapabilitySource) ([]Capability, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := []Capability{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		var m CustomToolManifest
		b, _ := os.ReadFile(filepath.Join(root, e.Name()))
		if yaml.Unmarshal(b, &m) != nil || m.Name == "" {
			continue
		}
		if !m.Enabled {
			continue
		}
		if m.ExecutionLocation == "" {
			m.ExecutionLocation = LocationController
		}
		out = append(out, Capability{ID: "custom.tool." + m.Name, Name: m.Name, Type: TypeTool, Source: source, Schema: m.Schema, Description: m.Description, ExecutionLocation: m.ExecutionLocation, Policy: m.Policy, Enabled: m.Enabled, ApprovalRequired: len(m.Policy.ApprovalsRequired) > 0, NetworkRequired: m.BackendType == "http" || m.ExecutionLocation == LocationRemote, Metadata: map[string]any{"backend_type": m.BackendType, "command": m.Command, "url": m.URL, "timeout_seconds": m.TimeoutSeconds, "retries": m.Retries, "auth": m.Auth}})
	}
	return out, nil
}

func loadProviderProfiles(root string) ([]ProviderProfile, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := []ProviderProfile{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		var p ProviderProfile
		b, _ := os.ReadFile(filepath.Join(root, e.Name()))
		if yaml.Unmarshal(b, &p) == nil && p.Name != "" {
			out = append(out, p)
		}
	}
	return out, nil
}
