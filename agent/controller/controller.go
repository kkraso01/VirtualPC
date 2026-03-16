package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"virtualpc/agent/config"
	"virtualpc/agent/providers"
	"virtualpc/agent/safety"
	"virtualpc/agent/tools"
	"virtualpc/internal/cli"
)

type Controller struct {
	planner   *Planner
	executor  *Executor
	limiter   *safety.RateLimiter
	limits    safety.ResourceLimits
	config    config.Config
	logs      *os.File
	systemMsg string
}

func New(cfg config.Config, c *cli.Client, provider providers.Provider, systemPromptPath string, approvalMode bool, logPath string) (*Controller, error) {
	systemMsg := defaultSystemPrompt()
	if b, err := os.ReadFile(systemPromptPath); err == nil {
		systemMsg = string(b)
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Controller{planner: NewPlanner(provider), executor: NewExecutor(c, cfg.WritableRoots(), approvalMode, cfg), limiter: safety.NewRateLimiter(cfg.RequestsPerMinute), config: cfg, logs: f, systemMsg: systemMsg, limits: safety.ResourceLimits{MaxCommandsPerSession: cfg.MaxCommandsPerSession, MaxRuntime: cfg.RuntimeLimit(), MaxDiskUsageMB: cfg.MaxDiskUsageMB, MaxMemoryMB: cfg.MaxMemoryMB, MaxProcesses: cfg.MaxProcesses, MaxIterations: cfg.MaxIterations, MaxFailures: cfg.MaxFailures, MaxRepeatedCommand: cfg.MaxRepeatedCommand, MaxMachinesCreated: cfg.MaxMachinesCreated, MaxForks: cfg.MaxForks, MaxContainers: cfg.MaxContainers}}, nil
}
func (c *Controller) Close() error { return c.logs.Close() }

func (c *Controller) Run(session *Session) error {
	session.Status = "running"
	_ = session.Save()
	c.logEvent(session.SessionID, "session_start", map[string]any{"machine_id": session.MachineID, "provider": session.Provider, "model": session.Model})
	for {
		session.IterationCount++
		if err := c.limits.Validate(len(session.CommandHistory), session.CreatedAt, session.IterationCount, session.ToolFailures, session.RepeatedCommands); err != nil {
			session.Status = "aborted"
			session.LastError = err.Error()
			_ = session.Save()
			c.logEvent(session.SessionID, "session_abort", map[string]any{"reason": err.Error()})
			return err
		}
		if err := c.limiter.Allow(); err != nil {
			session.Status = "aborted"
			session.LastError = err.Error()
			_ = session.Save()
			c.logEvent(session.SessionID, "policy_violation", map[string]any{"reason": err.Error()})
			return err
		}
		call, done, err := c.planner.NextAction(context.Background(), c.systemMsg, session)
		if err != nil {
			session.ToolFailures++
			session.LastError = err.Error()
			c.logEvent(session.SessionID, "provider_error", map[string]any{"error": err.Error()})
			_ = session.Save()
			continue
		}
		if done {
			session.Status = "completed"
			_ = session.Save()
			return nil
		}
		session.ToolCalls = append(session.ToolCalls, call)
		if call.Name == "run_command" {
			cmd, _ := call.Arguments["command"].(string)
			session.CommandHistory = append(session.CommandHistory, cmd)
			if strings.TrimSpace(cmd) == strings.TrimSpace(session.LastCommand) {
				session.RepeatedCommands++
			} else {
				session.RepeatedCommands = 0
			}
			session.LastCommand = cmd
		}
		c.logEvent(session.SessionID, "tool_call", map[string]any{"tool": call.Name, "arguments": call.Arguments})
		res, err := c.executor.Execute(call, c.config.DangerousCommandMode)
		if err != nil {
			session.ToolFailures++
			session.LastError = err.Error()
			session.Record(tools.ToolResult{Tool: call.Name, Success: false, Output: err.Error()})
			_ = session.Save()
			c.logEvent(session.SessionID, "blocked_action", map[string]any{"tool": call.Name, "error": err.Error()})
			continue
		}
		c.trackState(session, call)
		session.Record(res)
		_ = session.Save()
		c.logEvent(session.SessionID, "tool_result", map[string]any{"tool": call.Name, "success": res.Success})
	}
}

func (c *Controller) trackState(s *Session, call tools.ToolCall) {
	switch call.Name {
	case "write_file", "upload_file":
		if p, ok := call.Arguments["remote_dst"].(string); ok {
			s.FilesModified = append(s.FilesModified, p)
		}
	case "start_service":
		if name, ok := call.Arguments["name"].(string); ok {
			s.ServicesStarted = append(s.ServicesStarted, name)
		}
	case "snapshot_machine":
		s.Snapshots = append(s.Snapshots, fmt.Sprintf("iteration-%d", s.IterationCount))
	}
}

func (c *Controller) logEvent(sessionID, event string, fields map[string]any) {
	line := map[string]any{"time": time.Now().Format(time.RFC3339), "session_id": sessionID, "event": event, "fields": fields}
	b, _ := json.Marshal(line)
	_, _ = c.logs.Write(append(b, '\n'))
}

func defaultSystemPrompt() string {
	return "You are operating a Linux machine through a tool interface. Use tools safely and obey policy."
}
