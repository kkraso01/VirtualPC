package controller

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
	"virtualpc/agent/tools"
)

type Session struct {
	SessionID        string             `json:"session_id"`
	Provider         string             `json:"provider"`
	Model            string             `json:"model"`
	MachineID        string             `json:"machine_id"`
	Goal             string             `json:"goal"`
	History          []tools.ToolResult `json:"history"`
	ToolCalls        []tools.ToolCall   `json:"tool_calls"`
	Snapshots        []string           `json:"snapshots"`
	ServicesStarted  []string           `json:"services_started"`
	FilesModified    []string           `json:"files_modified"`
	IterationCount   int                `json:"iteration_count"`
	Status           string             `json:"status"`
	LastError        string             `json:"last_error"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	ToolFailures     int                `json:"tool_failures"`
	LastCommand      string             `json:"last_command"`
	RepeatedCommands int                `json:"repeated_commands"`
	CommandHistory   []string           `json:"command_history"`
	StopRequested    bool               `json:"stop_requested"`
}

func sessionDir() string {
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".virtualpc", "agent")
}

func sessionDBPath() string { return filepath.Join(sessionDir(), "sessions.db") }

func openDB() (*sql.DB, error) {
	if err := os.MkdirAll(sessionDir(), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", sessionDBPath())
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sessions(id TEXT PRIMARY KEY, payload TEXT NOT NULL, updated_at TEXT NOT NULL);`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (s *Session) Save() error {
	s.UpdatedAt = time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = s.UpdatedAt
	}
	b, _ := json.Marshal(s)
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`INSERT INTO sessions(id,payload,updated_at) VALUES(?,?,?) ON CONFLICT(id) DO UPDATE SET payload=excluded.payload,updated_at=excluded.updated_at`, s.SessionID, string(b), s.UpdatedAt.Format(time.RFC3339Nano))
	return err
}

func LoadSession(id string) (*Session, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var payload string
	if err := db.QueryRow(`SELECT payload FROM sessions WHERE id=?`, id).Scan(&payload); err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal([]byte(payload), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func ListSessions() ([]Session, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT payload FROM sessions ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Session{}
	for rows.Next() {
		var p string
		_ = rows.Scan(&p)
		var s Session
		if json.Unmarshal([]byte(p), &s) == nil {
			out = append(out, s)
		}
	}
	return out, rows.Err()
}

func (s *Session) Record(result tools.ToolResult) {
	s.History = append(s.History, result)
	s.UpdatedAt = time.Now()
}

func SessionLogPath(id string) string { return filepath.Join(sessionDir(), id+".log") }
