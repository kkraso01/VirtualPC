# VirtualPC Optional Agent Controller

## Architecture

The new agent layer is optional and runs **above** VirtualPC runtime components:

User/App → LLM Provider → `vpc-agent-controller` → VirtualPC API/CLI (`vpc`) → `virtualpcd` → Firecracker VM.

This preserves core runtime boundaries: the controller is a client and never mutates daemon internals directly.

## Tool Schemas

The controller exposes JSON-schema tool definitions compatible with OpenAI/Anthropic tool calling:

- `create_machine`
- `start_machine`
- `run_command`
- `open_shell`
- `write_file`
- `read_file`
- `upload_file`
- `download_file`
- `start_service`
- `stop_service`
- `snapshot_machine`
- `fork_machine`
- `destroy_machine`

Mappings:

- `run_command` → `vpc machine exec`
- `write_file`/`upload_file` → `vpc machine cp-to`
- `read_file`/`download_file` → `vpc machine cp-from`
- `snapshot_machine` → `vpc snapshot create`
- `fork_machine` → `vpc machine fork`

## Safety Enforcement

Safety is enforced in the controller before each tool execution:

1. **Command policy** (`agent/safety/command_policy.go`)
   - denylist blocks high-risk commands (`rm -rf /`, `mkfs`, `shutdown`, etc.)
   - allowlist supports common safe operations.
2. **Resource limits** (`agent/safety/resource_limits.go`)
   - max commands/session
   - max runtime minutes
   - iteration, repeated-command, failure thresholds.
3. **Filesystem guard** (`agent/safety/filesystem_guard.go`)
   - writable root defaults to `/workspace`
   - blocks `/proc`, `/sys`, `/dev`.
4. **Rate limits / network safety** (`agent/safety/rate_limits.go`)
   - request-per-minute limits to prevent runaway network behavior.
5. **Optional human approval mode**
   - enabled with `--approval`
   - destructive tools (e.g. `destroy_machine`) require operator approval.

## Session + Observability

Session state is persisted under `~/.virtualpc/agent/sessions/` with:

- machine id
- command history
- files modified
- services started
- snapshots created
- tool execution results

`vpc agent logs <session-id>` reads persisted session logs.

## Sample Session

```bash
vpc agent start --machine m-123 --goal "fix failing tests"
vpc agent attach s-1700000000
vpc agent logs s-1700000000
vpc agent stop s-1700000000
```

You can inspect tool schemas with:

```bash
vpc-agent-controller schemas
```
