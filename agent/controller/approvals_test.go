package controller

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApprovalPendingApproveDenyFlow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := NewApprovalManager()
	err := m.Require(context.Background(), "s1", "custom.tool.deploy", "human", map[string]any{"env": "dev"})
	if err == nil || !strings.Contains(err.Error(), "approval pending") {
		t.Fatalf("expected pending approval, got %v", err)
	}
	items, err := ListApprovals("s1")
	if err != nil || len(items) != 1 {
		t.Fatalf("expected one approval, err=%v items=%d", err, len(items))
	}
	if err := ResolveApproval("s1", items[0].ID, true); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if err := m.Require(context.Background(), "s1", "custom.tool.deploy", "human", nil); err != nil {
		t.Fatalf("approval should pass after decision: %v", err)
	}

	err = m.Require(context.Background(), "s2", "custom.tool.destroy", "human", nil)
	if err == nil {
		t.Fatal("expected pending on s2")
	}
	items, err = ListApprovals("s2")
	if err != nil || len(items) != 1 {
		t.Fatalf("expected one approval for s2, err=%v items=%d", err, len(items))
	}
	if err := ResolveApproval("s2", items[0].ID, false); err != nil {
		t.Fatalf("deny failed: %v", err)
	}
	err = m.Require(context.Background(), "s2", "custom.tool.destroy", "human", nil)
	if err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("expected denied, got %v", err)
	}
}

func TestSessionRecoveryWithPendingApproval(t *testing.T) {
	h := t.TempDir()
	t.Setenv("HOME", h)
	s := &Session{SessionID: "recover-1", MachineID: "m-1", Goal: "test", Status: "running", ResolvedTools: []string{"run_command"}}
	if err := s.Save(); err != nil {
		t.Fatalf("save session: %v", err)
	}
	m := NewApprovalManager()
	if err := m.Require(context.Background(), s.SessionID, "custom.tool.deploy", "human", nil); err == nil {
		t.Fatal("expected pending approval")
	}
	loaded, err := LoadSession("recover-1")
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if loaded.SessionID != s.SessionID || len(loaded.ResolvedTools) != 1 {
		t.Fatalf("unexpected loaded state: %+v", loaded)
	}
	if _, err := os.Stat(filepath.Join(h, ".virtualpc", "agent", "sessions.db")); err != nil {
		t.Fatalf("expected session db present: %v", err)
	}
	if items, err := ListApprovals(s.SessionID); err != nil || len(items) != 1 {
		t.Fatalf("expected pending approval persisted, err=%v items=%d", err, len(items))
	}
}
