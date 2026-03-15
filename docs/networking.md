# Networking

- CLI communicates with daemon over Unix socket.
- Firecracker guest networking is host-bridged/tap-managed by runtime module.
- Guest-internal service networking handled via containerd namespace.
