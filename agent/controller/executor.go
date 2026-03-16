package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"virtualpc/agent/capabilities"
	"virtualpc/agent/config"
	"virtualpc/agent/mcp"
	"virtualpc/agent/safety"
	"virtualpc/agent/tools"
	"virtualpc/internal/cli"
)

type Executor struct {
	client      *cli.Client
	commandSafe safety.CommandPolicy
	fsGuard     safety.FilesystemGuard
	approval    bool
	cfg         config.Config
	registry    *tools.Registry
	dispatcher  *capabilities.Dispatcher
}

type builtinExec struct{ e *Executor }

func (b builtinExec) ExecuteBuiltin(name string, args map[string]any) (string, error) {
	return b.e.dispatch(tools.ToolCall{Name: name, Arguments: args})
}

func NewExecutor(client *cli.Client, writableRoots []string, requireApproval bool, cfg config.Config, capReg *capabilities.Registry, servers []mcp.ServerConfig) *Executor {
	e := &Executor{client: client, commandSafe: safety.DefaultCommandPolicy(), fsGuard: safety.NewFilesystemGuard(writableRoots...), approval: requireApproval, cfg: cfg, registry: tools.NewRegistry()}
	appr := NewApprovalManager()
	if !requireApproval {
		appr = nil
	}
	mcpRuntime := mcp.NewRuntime(servers, mcp.NoopInvoker{})
	exec := capabilities.NewExecutor(builtinExec{e}, mcpRuntime)
	e.dispatcher = capabilities.NewDispatcher(capReg, exec, appr, capabilities.NewAuditor(SessionLogPath("capability-audit")))
	return e
}

func (e *Executor) Execute(sessionID string, call tools.ToolCall, dangerousMode string) (tools.ToolResult, error) {
	result := tools.ToolResult{Tool: call.Name, Success: false}
	if err := e.registry.Validate(call); err == nil {
		if call.Name == "run_command" {
			command, _ := call.Arguments["command"].(string)
			decision, reason := e.commandSafe.Evaluate(command, dangerousMode)
			if decision == safety.DecisionBlock {
				return result, fmt.Errorf("blocked command: %s", reason)
			}
			if decision == safety.DecisionApprove {
				return result, fmt.Errorf("approval required: %s", reason)
			}
		}
		if err := e.validateFS(call); err != nil {
			return result, err
		}
	}
	execRes, err := e.dispatcher.Dispatch(context.Background(), capabilities.ExecutionRequest{SessionID: sessionID, Name: call.Name, Args: call.Arguments})
	if err != nil {
		if execRes.Blocked {
			return result, fmt.Errorf("blocked: %s", execRes.Reason)
		}
		return result, err
	}
	result.Success = execRes.Success
	result.Output = execRes.Output
	if !result.Success {
		return result, fmt.Errorf("execution failed")
	}
	return result, nil
}

func (e *Executor) validateFS(call tools.ToolCall) error {
	for _, k := range []string{"remote_dst", "remote_src", "path", "local_dst", "local_src"} {
		if p, ok := call.Arguments[k].(string); ok && strings.HasPrefix(p, "/") {
			if err := e.fsGuard.Validate(p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Executor) dispatch(call tools.ToolCall) (string, error) {
	m := getString(call.Arguments, "machine_id")
	switch call.Name {
	case "create_machine":
		out := map[string]any{}
		err := e.client.Do("POST", "/v1/machines", map[string]string{"profile": getString(call.Arguments, "profile")}, &out)
		return stringify(out), err
	case "start_machine":
		out := map[string]any{}
		err := e.client.Do("POST", "/v1/machines/"+m+"/start", map[string]string{}, &out)
		return stringify(out), err
	case "run_command":
		out := map[string]any{}
		cmd := strings.Fields(getString(call.Arguments, "command"))
		err := e.client.Do("POST", "/v1/machines/"+m+"/exec", map[string]any{"command": cmd}, &out)
		return stringify(out), err
	case "open_shell":
		out := map[string]any{}
		err := e.client.Do("POST", "/v1/machines/"+m+"/shell", map[string]any{}, &out)
		return stringify(out), err
	case "write_file", "upload_file":
		out := map[string]any{}
		err := e.client.Do("POST", "/v1/machines/"+m+"/cp-to", map[string]any{"src": getString(call.Arguments, "local_src"), "dst": getString(call.Arguments, "remote_dst")}, &out)
		return stringify(out), err
	case "read_file", "download_file":
		out := map[string]any{}
		err := e.client.Do("POST", "/v1/machines/"+m+"/cp-from", map[string]any{"src": getString(call.Arguments, "remote_src"), "dst": getString(call.Arguments, "local_dst")}, &out)
		return stringify(out), err
	case "start_service":
		out := map[string]any{}
		in := map[string]string{"MachineID": m, "Name": getString(call.Arguments, "name"), "Image": getString(call.Arguments, "image")}
		err := e.client.Do("POST", "/v1/services", in, &out)
		return stringify(out), err
	case "stop_service":
		out := map[string]any{}
		err := e.client.Do("POST", "/v1/services/"+getString(call.Arguments, "service_id")+"/stop", map[string]any{}, &out)
		return stringify(out), err
	case "snapshot_machine":
		out := map[string]any{}
		err := e.client.Do("POST", "/v1/snapshots", map[string]string{"machine_id": m}, &out)
		return stringify(out), err
	case "fork_machine":
		out := map[string]any{}
		err := e.client.Do("POST", "/v1/fork", map[string]string{"snapshot_id": getString(call.Arguments, "snapshot_id")}, &out)
		return stringify(out), err
	case "destroy_machine":
		out := map[string]any{}
		err := e.client.Do("POST", "/v1/machines/"+m+"/destroy", map[string]string{}, &out)
		return stringify(out), err
	default:
		return "", fmt.Errorf("unknown tool: %s", call.Name)
	}
}
func getString(m map[string]any, key string) string { v, _ := m[key].(string); return v }
func stringify(v any) string                        { b, _ := json.Marshal(v); return string(b) }
