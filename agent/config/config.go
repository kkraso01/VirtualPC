package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"virtualpc/agent/providers"
)

type Config struct {
	Provider                  string
	Model                     string
	BaseURL                   string
	APIKey                    string
	SupportsResponsesAPI      bool
	SupportsChatCompletions   bool
	SupportsToolCalling       bool
	SupportsStatefulResponses bool
	MaxIterations             int
	MaxRuntimeMinutes         int
	DangerousCommandMode      string
	MaxCommandsPerSession     int
	MaxDiskUsageMB            int
	MaxMemoryMB               int
	MaxProcesses              int
	MaxFailures               int
	MaxRepeatedCommand        int
	MaxMachinesCreated        int
	MaxForks                  int
	MaxContainers             int
	RequestsPerMinute         int
	WritableRoot              string
	ExtraWritableRoot         string
	ApprovalRequiredTools     map[string]bool
}

func Default() Config {
	return Config{Provider: "openai", Model: "gpt-4o", SupportsResponsesAPI: true, SupportsChatCompletions: true, SupportsToolCalling: true, SupportsStatefulResponses: true, MaxIterations: 40, MaxRuntimeMinutes: 20, DangerousCommandMode: "block", MaxCommandsPerSession: 80, MaxDiskUsageMB: 2048, MaxMemoryMB: 2048, MaxProcesses: 128, MaxFailures: 6, MaxRepeatedCommand: 3, MaxMachinesCreated: 1, MaxForks: 1, MaxContainers: 4, RequestsPerMinute: 60, WritableRoot: "/workspace", ExtraWritableRoot: "/tmp", ApprovalRequiredTools: map[string]bool{"destroy_machine": true, "fork_machine": true, "snapshot_machine": false}}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k, v := strings.TrimSpace(parts[0]), strings.TrimSpace(strings.Trim(parts[1], "\"'"))
		switch k {
		case "provider":
			cfg.Provider = v
		case "model":
			cfg.Model = v
		case "base_url":
			cfg.BaseURL = v
		case "api_key":
			cfg.APIKey = v
		case "supports_responses_api":
			cfg.SupportsResponsesAPI = parseBool(v, cfg.SupportsResponsesAPI)
		case "supports_chat_completions":
			cfg.SupportsChatCompletions = parseBool(v, cfg.SupportsChatCompletions)
		case "supports_tool_calling":
			cfg.SupportsToolCalling = parseBool(v, cfg.SupportsToolCalling)
		case "supports_stateful_responses":
			cfg.SupportsStatefulResponses = parseBool(v, cfg.SupportsStatefulResponses)
		case "max_iterations":
			cfg.MaxIterations = atoi(v, cfg.MaxIterations)
		case "max_runtime_minutes":
			cfg.MaxRuntimeMinutes = atoi(v, cfg.MaxRuntimeMinutes)
		case "dangerous_command_mode":
			cfg.DangerousCommandMode = v
		case "max_commands_per_session":
			cfg.MaxCommandsPerSession = atoi(v, cfg.MaxCommandsPerSession)
		case "max_disk_usage":
			cfg.MaxDiskUsageMB = parseSizeMB(v, cfg.MaxDiskUsageMB)
		case "max_memory_usage":
			cfg.MaxMemoryMB = parseSizeMB(v, cfg.MaxMemoryMB)
		case "max_processes":
			cfg.MaxProcesses = atoi(v, cfg.MaxProcesses)
		case "max_failures":
			cfg.MaxFailures = atoi(v, cfg.MaxFailures)
		case "max_repeated_command":
			cfg.MaxRepeatedCommand = atoi(v, cfg.MaxRepeatedCommand)
		case "max_machines_created":
			cfg.MaxMachinesCreated = atoi(v, cfg.MaxMachinesCreated)
		case "max_forks":
			cfg.MaxForks = atoi(v, cfg.MaxForks)
		case "max_containers":
			cfg.MaxContainers = atoi(v, cfg.MaxContainers)
		case "requests_per_minute":
			cfg.RequestsPerMinute = atoi(v, cfg.RequestsPerMinute)
		case "writable_root":
			cfg.WritableRoot = v
		}
	}
	if err := s.Err(); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	return cfg, nil
}
func (c Config) RuntimeLimit() time.Duration { return time.Duration(c.MaxRuntimeMinutes) * time.Minute }
func (c Config) ProviderCapabilities() providers.Capabilities {
	return providers.Capabilities{SupportsResponsesAPI: c.SupportsResponsesAPI, SupportsChatCompletions: c.SupportsChatCompletions, SupportsToolCalling: c.SupportsToolCalling, SupportsStatefulResponses: c.SupportsStatefulResponses}
}
func (c Config) WritableRoots() []string { return []string{c.WritableRoot, c.ExtraWritableRoot} }
func atoi(v string, f int) int {
	i, e := strconv.Atoi(v)
	if e != nil {
		return f
	}
	return i
}
func parseSizeMB(v string, f int) int {
	return atoi(strings.TrimSuffix(strings.ToLower(strings.TrimSpace(v)), "mb"), f)
}
func parseBool(v string, f bool) bool {
	b, e := strconv.ParseBool(strings.ToLower(v))
	if e != nil {
		return f
	}
	return b
}
