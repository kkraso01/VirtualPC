package controller

import (
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
	return &Controller{
		planner:   NewPlanner(provider),
		executor:  NewExecutor(c, cfg.WritableRoot, approvalMode),
		limiter:   safety.NewRateLimiter(cfg.RequestsPerMinute),
		config:    cfg,
		logs:      f,
		systemMsg: systemMsg,
		limits:    safety.ResourceLimits{MaxCommandsPerSession: cfg.MaxCommandsPerSession, MaxRuntime: cfg.RuntimeLimit(), MaxDiskUsageMB: cfg.MaxDiskUsageMB, MaxMemoryMB: cfg.MaxMemoryMB, MaxProcesses: cfg.MaxProcesses, MaxIterations: cfg.MaxIterations, MaxFailures: cfg.MaxFailures, MaxRepeatedCommand: cfg.MaxRepeatedCommand},
	}, nil
}

func (c *Controller) Close() error { return c.logs.Close() }

func (c *Controller) Run(session *Session) error {
	session.Status = "running"
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	for {
		session.Iterations++
		if err := c.limits.Validate(len(session.CommandHistory), session.CreatedAt, session.Iterations, session.ToolFailures, session.RepeatedCommands); err != nil {
			session.Status = "stopped"
			c.logf(session.ID, "loop stop: %v", err)
			return err
		}
		if err := c.limiter.Allow(); err != nil {
			session.Status = "stopped"
			c.logf(session.ID, "rate limit stop: %v", err)
			return err
		}
		call, done, err := c.planner.NextAction(c.systemMsg, session)
		if err != nil {
			session.ToolFailures++
			c.logf(session.ID, "planner error: %v", err)
			continue
		}
		if done {
			session.Status = "completed"
			c.logf(session.ID, "session completed")
			return nil
		}
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
		res, err := c.executor.Execute(call, c.config.DangerousCommandMode)
		if err != nil {
			session.ToolFailures++
			c.logf(session.ID, "safety violation / execution error (%s): %v", call.Name, err)
			session.Record(tools.ToolResult{Tool: call.Name, Success: false, Output: err.Error()})
			_ = session.Save()
			continue
		}
		c.trackState(session, call, res)
		session.Record(res)
		_ = session.Save()
		c.logf(session.ID, "tool executed: %s", call.Name)
	}
}

func (c *Controller) trackState(s *Session, call tools.ToolCall, _ tools.ToolResult) {
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
		s.SnapshotsCreated = append(s.SnapshotsCreated, fmt.Sprintf("iteration-%d", s.Iterations))
	}
}

func (c *Controller) logf(sessionID, format string, args ...any) {
	line := fmt.Sprintf("%s [%s] %s\n", time.Now().Format(time.RFC3339), sessionID, fmt.Sprintf(format, args...))
	_, _ = c.logs.WriteString(line)
}

func defaultSystemPrompt() string {
	return "You are operating a Linux machine through a tool interface. You can run commands, edit files, and manage services. The machine is disposable. Use commands responsibly. Avoid destructive operations unless necessary. Never attempt to escape the environment."
}
