package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Provider              string
	Model                 string
	MaxIterations         int
	MaxRuntimeMinutes     int
	DangerousCommandMode  string
	MaxCommandsPerSession int
	MaxDiskUsageMB        int
	MaxMemoryMB           int
	MaxProcesses          int
	MaxFailures           int
	MaxRepeatedCommand    int
	RequestsPerMinute     int
	WritableRoot          string
	ApprovalRequiredTools map[string]bool
}

func Default() Config {
	return Config{
		Provider:              "openai",
		Model:                 "gpt-4o",
		MaxIterations:         40,
		MaxRuntimeMinutes:     20,
		DangerousCommandMode:  "block",
		MaxCommandsPerSession: 80,
		MaxDiskUsageMB:        2048,
		MaxMemoryMB:           2048,
		MaxProcesses:          128,
		MaxFailures:           6,
		MaxRepeatedCommand:    3,
		RequestsPerMinute:     60,
		WritableRoot:          "/workspace",
		ApprovalRequiredTools: map[string]bool{"destroy_machine": true, "snapshot_machine": false, "fork_machine": false},
	}
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
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(strings.Trim(parts[1], "\"'"))
		switch key {
		case "provider":
			cfg.Provider = value
		case "model":
			cfg.Model = value
		case "max_iterations":
			cfg.MaxIterations = atoi(value, cfg.MaxIterations)
		case "max_runtime_minutes":
			cfg.MaxRuntimeMinutes = atoi(value, cfg.MaxRuntimeMinutes)
		case "dangerous_command_mode":
			cfg.DangerousCommandMode = value
		case "max_commands_per_session":
			cfg.MaxCommandsPerSession = atoi(value, cfg.MaxCommandsPerSession)
		case "max_disk_usage":
			cfg.MaxDiskUsageMB = parseSizeMB(value, cfg.MaxDiskUsageMB)
		case "max_memory_usage":
			cfg.MaxMemoryMB = parseSizeMB(value, cfg.MaxMemoryMB)
		case "max_processes":
			cfg.MaxProcesses = atoi(value, cfg.MaxProcesses)
		case "max_failures":
			cfg.MaxFailures = atoi(value, cfg.MaxFailures)
		case "max_repeated_command":
			cfg.MaxRepeatedCommand = atoi(value, cfg.MaxRepeatedCommand)
		case "requests_per_minute":
			cfg.RequestsPerMinute = atoi(value, cfg.RequestsPerMinute)
		case "writable_root":
			cfg.WritableRoot = value
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	return cfg, nil
}

func (c Config) RuntimeLimit() time.Duration {
	return time.Duration(c.MaxRuntimeMinutes) * time.Minute
}

func atoi(v string, fallback int) int {
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func parseSizeMB(v string, fallback int) int {
	trim := strings.TrimSpace(strings.ToLower(v))
	trim = strings.TrimSuffix(trim, "mb")
	return atoi(trim, fallback)
}
