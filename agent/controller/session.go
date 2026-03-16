package controller

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"virtualpc/agent/tools"
)

type Session struct {
	ID               string             `json:"id"`
	MachineID        string             `json:"machine_id"`
	Goal             string             `json:"goal"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	Status           string             `json:"status"`
	Iterations       int                `json:"iterations"`
	ToolFailures     int                `json:"tool_failures"`
	LastCommand      string             `json:"last_command"`
	RepeatedCommands int                `json:"repeated_commands"`
	CommandHistory   []string           `json:"command_history"`
	FilesModified    []string           `json:"files_modified"`
	ServicesStarted  []string           `json:"services_started"`
	SnapshotsCreated []string           `json:"snapshots_created"`
	ToolResults      []tools.ToolResult `json:"tool_results"`
}

func (s *Session) Record(result tools.ToolResult) {
	s.UpdatedAt = time.Now()
	s.ToolResults = append(s.ToolResults, result)
}

func sessionDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".virtualpc", "agent", "sessions")
}

func (s *Session) Save() error {
	if err := os.MkdirAll(sessionDir(), 0o755); err != nil {
		return err
	}
	path := filepath.Join(sessionDir(), s.ID+".json")
	b, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(path, b, 0o644)
}

func LoadSession(id string) (*Session, error) {
	path := filepath.Join(sessionDir(), id+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func SessionLogPath(id string) string {
	return filepath.Join(sessionDir(), id+".log")
}
