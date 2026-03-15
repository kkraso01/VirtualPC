# Security

- Primary isolation boundary: Firecracker microVM.
- Inner service isolation: containerd namespaces inside guest.
- Policy modes: offline, allow-all, allowlist.
- Audit events persist all core operations.
- Secrets injection is runtime-scoped; snapshot exclusion is required behavior for launch hardening.
