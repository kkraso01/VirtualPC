# Runtime

## Machine lifecycle states

`pending -> booting -> running -> stopping -> stopped -> deleted` (with `failed` fallback)

## Firecracker path

- create per-machine runtime directory
- materialize vm config and workspace refs
- start vm process
- handshake with `vpc-agent`
- execute commands/files/services via guest agent APIs

## Guest baseline

- Linux minimal image
- `vpc-agent`
- workspace layout
- containerd + nerdctl
- shell and diagnostics tooling
