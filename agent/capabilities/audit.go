package capabilities

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type AuditRecord struct {
	Time       time.Time      `json:"time"`
	SessionID  string         `json:"session_id"`
	Capability string         `json:"capability"`
	Allowed    bool           `json:"allowed"`
	Reason     string         `json:"reason,omitempty"`
	Args       map[string]any `json:"args,omitempty"`
}

type Auditor struct{ path string }

func NewAuditor(path string) *Auditor { return &Auditor{path: path} }

func (a *Auditor) Record(r AuditRecord) {
	if a == nil || a.path == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(a.path), 0o755)
	f, err := os.OpenFile(a.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	b, _ := json.Marshal(r)
	_, _ = f.Write(append(b, '\n'))
}
