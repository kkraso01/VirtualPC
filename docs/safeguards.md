# Agent Safeguards

## Command policy
- Allowlist for common safe commands.
- Denylist blocks fork bombs, destructive filesystem operations, privileged container flags, mount/kernel/device operations.
- Approval mode can require explicit operator confirmation for sensitive actions.

## Filesystem guard
- Write paths constrained to `/workspace` and `/tmp` by default.
- Explicitly blocked: `/etc`, `/usr`, `/proc`, `/sys`, `/dev`.

## Loop / runaway protection
- Repeated identical command thresholds.
- Repeated failure thresholds.
- Iteration and command count hard caps.

## Resource budgets
- max commands
- max runtime
- max machines created
- max forks
- max containers
- rate-limited outbound provider requests

## Approval mode
Interactive approval blocks by default:
- `destroy_machine`
- `fork_machine`
- `snapshot_machine`
- `start_service` (external-facing risk)
