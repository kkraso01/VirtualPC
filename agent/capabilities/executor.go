package capabilities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ExecutionRequest struct {
	SessionID string
	Name      string
	Args      map[string]any
}

type ExecutionResult struct {
	CapabilityID string         `json:"capability_id"`
	Name         string         `json:"name"`
	Output       string         `json:"output"`
	Data         map[string]any `json:"data,omitempty"`
	Success      bool           `json:"success"`
	Blocked      bool           `json:"blocked"`
	Reason       string         `json:"reason,omitempty"`
}

type BuiltinExecutor interface {
	ExecuteBuiltin(name string, args map[string]any) (string, error)
}

type MCPExecutor interface {
	InvokeTool(ctx context.Context, server, name string, args map[string]any) (string, error)
	FetchResource(ctx context.Context, server, name string, args map[string]any) (string, error)
	FetchPrompt(ctx context.Context, server, name string, args map[string]any) (string, error)
}

type Executor struct {
	builtin BuiltinExecutor
	mcp     MCPExecutor
	http    *http.Client
}

func NewExecutor(builtin BuiltinExecutor, mcp MCPExecutor) *Executor {
	return &Executor{builtin: builtin, mcp: mcp, http: &http.Client{Timeout: 30 * time.Second}}
}

func (e *Executor) Execute(ctx context.Context, cap Capability, req ExecutionRequest) (ExecutionResult, error) {
	meta := cap.Metadata
	backend, _ := meta["backend_type"].(string)
	out := ExecutionResult{CapabilityID: cap.ID, Name: cap.Name, Success: false}
	switch {
	case cap.Source == SourceBuiltin:
		if e.builtin == nil {
			return out, fmt.Errorf("builtin executor unavailable")
		}
		s, err := e.builtin.ExecuteBuiltin(cap.Name, req.Args)
		out.Output = s
		out.Success = err == nil
		return out, err
	case cap.Source == SourceMCP:
		if e.mcp == nil {
			return out, fmt.Errorf("mcp executor unavailable")
		}
		server := strings.Split(strings.TrimPrefix(cap.ID, "mcp."), ".")[0]
		var s string
		var err error
		switch cap.Type {
		case TypeTool:
			s, err = e.mcp.InvokeTool(ctx, server, cap.Name, req.Args)
		case TypeResource:
			s, err = e.mcp.FetchResource(ctx, server, cap.Name, req.Args)
		case TypePrompt:
			s, err = e.mcp.FetchPrompt(ctx, server, cap.Name, req.Args)
		default:
			err = fmt.Errorf("unsupported mcp type: %s", cap.Type)
		}
		out.Output = s
		out.Success = err == nil
		return out, err
	case backend == "local":
		return e.runLocal(ctx, cap, req.Args)
	case backend == "http":
		return e.runHTTP(ctx, cap, req.Args)
	default:
		return out, fmt.Errorf("unsupported capability backend for %s", cap.Name)
	}
}

func (e *Executor) runLocal(ctx context.Context, cap Capability, args map[string]any) (ExecutionResult, error) {
	out := ExecutionResult{CapabilityID: cap.ID, Name: cap.Name}
	cmdPath, _ := cap.Metadata["command"].(string)
	if cmdPath == "" {
		return out, fmt.Errorf("local tool missing command")
	}
	templates, _ := cap.Metadata["args_template"].([]any)
	runArgs := []string{}
	for _, t := range templates {
		runArgs = append(runArgs, interpolate(fmt.Sprint(t), args))
	}
	timeout := intMeta(cap.Metadata, "timeout_seconds", 10)
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, cmdPath, runArgs...)
	cwd, _ := cap.Metadata["cwd"].(string)
	if cwd != "" {
		cmd.Dir = filepath.Clean(cwd)
	}
	cmd.Env = []string{"PATH=/usr/bin:/bin", "LANG=C"}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out.Output = strings.TrimSpace(stdout.String())
	if stderr.Len() > 0 {
		if out.Output != "" {
			out.Output += "\n"
		}
		out.Output += "stderr: " + strings.TrimSpace(stderr.String())
	}
	out.Success = err == nil
	return out, err
}

func (e *Executor) runHTTP(ctx context.Context, cap Capability, args map[string]any) (ExecutionResult, error) {
	out := ExecutionResult{CapabilityID: cap.ID, Name: cap.Name}
	method := strings.ToUpper(stringMeta(cap.Metadata, "method", "POST"))
	url := interpolate(stringMeta(cap.Metadata, "url", ""), args)
	bodyTemplate := stringMeta(cap.Metadata, "body_template", "")
	var body io.Reader
	if bodyTemplate != "" {
		body = strings.NewReader(interpolate(bodyTemplate, args))
	} else {
		b, _ := json.Marshal(args)
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", "application/json")
	if h, ok := cap.Metadata["headers"].(map[string]any); ok {
		for k, v := range h {
			req.Header.Set(k, interpolate(fmt.Sprint(v), args))
		}
	}
	resp, err := e.http.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	out.Output = string(b)
	out.Data = map[string]any{"status": resp.StatusCode}
	out.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	if !out.Success {
		return out, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return out, nil
}

func interpolate(tpl string, args map[string]any) string {
	out := tpl
	for k, v := range args {
		out = strings.ReplaceAll(out, "{{"+k+"}}", fmt.Sprint(v))
	}
	return out
}
func intMeta(m map[string]any, k string, def int) int {
	if m == nil {
		return def
	}
	switch v := m[k].(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return def
}
func stringMeta(m map[string]any, k, def string) string {
	if m == nil {
		return def
	}
	if s, ok := m[k].(string); ok && s != "" {
		return s
	}
	return def
}
