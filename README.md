# VirtualPC

VirtualPC is a **daemon-first, CLI-first, self-hosted runtime platform** for AI agents.

Primary product shape:
- `virtualpcd` control daemon
- `vpc` operator CLI
- Firecracker microVM per VirtualPC machine
- `vpc-agent` inside each guest VM
- inner project services run inside the guest via containerd

## Quick start

```bash
make build
make dev-up
make smoke
```

## Required launch-path commands

```bash
vpc daemon status
vpc machine create --profile minimal-shell
vpc machine list
vpc machine inspect <id>
vpc machine start <id>
vpc machine exec <id> -- echo hello
vpc machine shell <id>
vpc project create myproj
vpc machine assign <id> --project <project-id>
vpc service create --machine <id> --name db --image postgres:16
vpc service list --machine <id>
vpc snapshot create <id>
vpc machine fork <snapshot-id>
vpc task create --machine <id> --goal "build and run a small service"
vpc task run <task-id>
```

## Repository layout

- `cmd/virtualpcd`: control daemon
- `cmd/vpc`: CLI
- `cmd/vpc-agent`: guest agent
- `internal/runtime/firecracker`: primary runtime backend
- `packaging/docker/compose.dev.yml`: local bring-up
- `db/migrations`: durable schema for launch target
- `tests`: unit/integration/smoke coverage


## Agent capability layer

The VirtualPC runtime architecture is unchanged: `LLM -> Agent Controller (optional) -> VirtualPC API -> virtualpcd -> Firecracker VM`.

The optional controller now resolves skills, custom tools, MCP integrations, provider profiles, and policy bindings via a capability registry before tool execution.
