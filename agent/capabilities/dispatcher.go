package capabilities

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ApprovalManager interface {
	Require(ctx context.Context, sessionID, capabilityID, reason string, args map[string]any) error
}

type Dispatcher struct {
	registry  *Registry
	exec      *Executor
	approvals ApprovalManager
	audit     *Auditor
}

func NewDispatcher(reg *Registry, exec *Executor, approvals ApprovalManager, audit *Auditor) *Dispatcher {
	return &Dispatcher{registry: reg, exec: exec, approvals: approvals, audit: audit}
}

func (d *Dispatcher) Dispatch(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	cap, err := Resolve(d.registry, req.Name)
	if err != nil {
		return d.blocked(req, req.Name, err), err
	}
	if err := validateArgs(cap.Schema, req.Args); err != nil {
		return d.blocked(req, cap.ID, err), err
	}
	if err := d.enforcePolicy(ctx, cap, req); err != nil {
		return d.blocked(req, cap.ID, err), err
	}
	res, err := d.exec.Execute(ctx, cap, req)
	if err != nil {
		d.audit.Record(AuditRecord{Time: now(), SessionID: req.SessionID, Capability: cap.ID, Allowed: true, Reason: err.Error(), Args: req.Args})
		return res, err
	}
	d.audit.Record(AuditRecord{Time: now(), SessionID: req.SessionID, Capability: cap.ID, Allowed: true, Args: req.Args})
	return res, nil
}

func (d *Dispatcher) blocked(req ExecutionRequest, capID string, err error) ExecutionResult {
	d.audit.Record(AuditRecord{Time: now(), SessionID: req.SessionID, Capability: capID, Allowed: false, Reason: err.Error(), Args: req.Args})
	return ExecutionResult{CapabilityID: capID, Name: req.Name, Blocked: true, Reason: err.Error(), Success: false}
}

func (d *Dispatcher) enforcePolicy(ctx context.Context, cap Capability, req ExecutionRequest) error {
	if cap.ApprovalRequired || len(cap.Policy.ApprovalsRequired) > 0 {
		if d.approvals == nil {
			return fmt.Errorf("approval required")
		}
		if err := d.approvals.Require(ctx, req.SessionID, cap.ID, strings.Join(cap.Policy.ApprovalsRequired, ","), req.Args); err != nil {
			return err
		}
	}
	if len(cap.FilesystemScope) > 0 {
		for _, k := range []string{"path", "remote_dst", "remote_src", "cwd"} {
			if v, ok := req.Args[k].(string); ok && strings.HasPrefix(v, "/") {
				if !pathAllowed(v, cap.FilesystemScope) {
					return fmt.Errorf("filesystem path blocked: %s", v)
				}
			}
		}
	}
	if bt, _ := cap.Metadata["backend_type"].(string); bt == "local" {
		cmd, _ := cap.Metadata["command"].(string)
		allow, _ := cap.Metadata["allowed_executables"].([]any)
		if len(allow) > 0 {
			ok := false
			for _, a := range allow {
				if filepath.Clean(fmt.Sprint(a)) == filepath.Clean(cmd) {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("executable not allowlisted: %s", cmd)
			}
		}
	}
	if cap.NetworkRequired || cap.Policy.NetworkMode == "approved-only" {
		if u, ok := cap.Metadata["url"].(string); ok && u != "" {
			if err := enforceURLAllowlist(u, cap); err != nil {
				return err
			}
		}
	}
	return nil
}

func enforceURLAllowlist(raw string, cap Capability) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	host := strings.ToLower(u.Hostname())
	allow, _ := cap.Metadata["allowed_hosts"].([]any)
	if len(allow) == 0 {
		if strings.Contains(host, "internal") || strings.HasPrefix(host, "127.") || host == "localhost" {
			return nil
		}
		return fmt.Errorf("remote destination blocked: %s", host)
	}
	for _, h := range allow {
		if strings.EqualFold(fmt.Sprint(h), host) {
			return nil
		}
	}
	return fmt.Errorf("remote destination blocked: %s", host)
}

func pathAllowed(path string, roots []string) bool {
	clean := filepath.Clean(path)
	for _, r := range roots {
		root := filepath.Clean(r)
		rel, err := filepath.Rel(root, clean)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func validateArgs(schema map[string]any, args map[string]any) error {
	if schema == nil {
		return nil
	}
	requiredAny, _ := schema["required"].([]any)
	for _, r := range requiredAny {
		if _, ok := args[fmt.Sprint(r)]; !ok {
			return fmt.Errorf("missing required argument %s", r)
		}
	}
	if requiredStr, ok := schema["required"].([]string); ok {
		for _, r := range requiredStr {
			if _, ok := args[r]; !ok {
				return fmt.Errorf("missing required argument %s", r)
			}
		}
	}
	return nil
}

var now = func() time.Time { return time.Now() }
