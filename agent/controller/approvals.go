package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalDenied   ApprovalStatus = "denied"
)

type Approval struct {
	ID           string         `json:"id"`
	SessionID    string         `json:"session_id"`
	CapabilityID string         `json:"capability_id"`
	Reason       string         `json:"reason"`
	Status       ApprovalStatus `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type ApprovalManager struct{}

func NewApprovalManager() *ApprovalManager { return &ApprovalManager{} }

func ensureApprovals(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS approvals(id TEXT PRIMARY KEY, session_id TEXT, capability_id TEXT, reason TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
	return err
}

func (m *ApprovalManager) Require(_ context.Context, sessionID, capabilityID, reason string, _ map[string]any) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := ensureApprovals(db); err != nil {
		return err
	}
	var id, status string
	err = db.QueryRow(`SELECT id,status FROM approvals WHERE session_id=? AND capability_id=? ORDER BY updated_at DESC LIMIT 1`, sessionID, capabilityID).Scan(&id, &status)
	if err == nil {
		switch ApprovalStatus(status) {
		case ApprovalApproved:
			return nil
		case ApprovalDenied:
			return errors.New("action denied by operator")
		default:
			return fmt.Errorf("approval pending: %s", id)
		}
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	now := time.Now().UTC()
	id = fmt.Sprintf("appr-%d", now.UnixNano())
	_, err = db.Exec(`INSERT INTO approvals(id,session_id,capability_id,reason,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, id, sessionID, capabilityID, reason, string(ApprovalPending), now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	if err != nil {
		return err
	}
	return fmt.Errorf("approval pending: %s", id)
}

func ListApprovals(sessionID string) ([]Approval, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := ensureApprovals(db); err != nil {
		return nil, err
	}
	q := `SELECT id,session_id,capability_id,reason,status,created_at,updated_at FROM approvals`
	args := []any{}
	if sessionID != "" {
		q += ` WHERE session_id=?`
		args = append(args, sessionID)
	}
	q += ` ORDER BY updated_at DESC`
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Approval{}
	for rows.Next() {
		var a Approval
		var c, u string
		if err := rows.Scan(&a.ID, &a.SessionID, &a.CapabilityID, &a.Reason, &a.Status, &c, &u); err == nil {
			a.CreatedAt, _ = time.Parse(time.RFC3339Nano, c)
			a.UpdatedAt, _ = time.Parse(time.RFC3339Nano, u)
			out = append(out, a)
		}
	}
	return out, rows.Err()
}

func ResolveApproval(sessionID, approvalID string, approve bool) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := ensureApprovals(db); err != nil {
		return err
	}
	status := ApprovalDenied
	if approve {
		status = ApprovalApproved
	}
	_, err = db.Exec(`UPDATE approvals SET status=?,updated_at=? WHERE id=? AND session_id=?`, string(status), time.Now().UTC().Format(time.RFC3339Nano), approvalID, sessionID)
	return err
}
