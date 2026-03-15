# Security

## Isolation layers
- Firecracker VM boundary (or mock fallback when KVM unavailable).
- Dedicated runtime directory per machine (`data/firecracker/<machine-id>`).
- Per-machine network policy file and mode (`offline`, `nat`, `allowlist`).
- Guest filesystem operations restricted to machine guest root.

## Resource controls
Profiles define vCPU, memory, and disk intent. These values are stored and enforced during VM config generation.
Network mode is enforced by runtime network setup and can be audited via runtime metadata.

## Trust boundaries
- Host daemon is trusted control plane.
- Guest agent is trusted only inside machine boundary.
- Host filesystem mounts are explicit and scoped; no automatic host root mount into guest.
