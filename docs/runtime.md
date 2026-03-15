# Runtime

## Machine lifecycle
1. `create` persists machine metadata and profile.
2. `start` creates runtime dir, network config, starts Firecracker process supervisor, and launches `vpc-agent`.
3. `running` machine accepts guest RPC for exec, file transfer, containers, services, and tasks.
4. `stop` sends TERM to guest agent and VM process and records stopped state.
5. `destroy` kills remaining processes and removes sockets/runtime artifacts.

## Guest agent
`vpc-agent server` listens on a per-machine unix socket used as host-vsock bridge.
RPC envelope fields: `id`, `method`, `payload`.
Implemented methods:
- ExecCommand
- OpenPTY/SendInput/ClosePTY
- ReadFile/WriteFile/UploadFile/DownloadFile
- ListProcesses/GetMachineInfo
- StartContainer/StopContainer/ListContainers/ContainerLogs
- FetchLogs

## Container services
Services map to guest container lifecycle calls.
`vpc service create` calls `StartContainer` and persists service metadata.
`vpc service stop|destroy|logs` are wired through guest RPC.
