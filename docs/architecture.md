# Architecture

```text
vpc CLI -> virtualpcd (unix API) -> firecracker runtime manager -> Firecracker process
                                           \-> guest vsock client -> vpc-agent RPC server
```

## Daemon
- API server over unix socket.
- State store for machines/snapshots/projects/services/tasks/profiles.
- Runtime orchestrator for VM process lifecycle and guest RPC.

## Firecracker runtime
- `process_manager.go`: StartVM/StopVM/KillVM/InspectVM/VMStatus.
- Persists `vm_state.json` per machine for daemon restart recovery.
- Cleans sockets, process metadata, network artifacts on shutdown.

## Guest architecture
- `vpc-agent` runs in guest rootfs context.
- JSON RPC protocol over vsock bridge socket.
- Supports command execution, PTY stub control, file copy, process list, and container operations.

## Container runtime
- Guest RPC handlers call container actions (backed by nerdctl-compatible command model).
- Service layer maps project services to guest containers.

## Task execution
- Tasks transition: created -> running -> success/failed.
- Runner executes goal through guest `ExecCommand` and stores logs/artifacts metadata.


## Agent capability layer

The VirtualPC runtime architecture is unchanged: `LLM -> Agent Controller (optional) -> VirtualPC API -> virtualpcd -> Firecracker VM`.

The optional controller now resolves skills, custom tools, MCP integrations, provider profiles, and policy bindings via a capability registry before tool execution.
