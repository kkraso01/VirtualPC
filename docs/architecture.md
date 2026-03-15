# Architecture

## Core

1. `virtualpcd` (containerized daemon) is the local control plane.
2. Every machine is a Firecracker-backed microVM boundary.
3. `vpc-agent` runs inside each guest and exposes machine operations.
4. `vpc` CLI is the first-class operator interface over Unix socket.

## Trust boundaries

- Host daemon orchestration != workload isolation.
- Workload isolation boundary is the microVM.
- Inner containerd runs workloads **inside guest**, not on host siblings by default.

## Components

- State: durable machine/project/task/snapshot metadata.
- Runtime manager: lifecycle abstraction with Firecracker primary backend.
- Workflow layer: task execution loop with durable orchestration hooks.
- Artifact layer: object storage integration for logs, bundles, exports.
- Event/audit layer: persisted operation history.

## Non-core

- `vpc-gateway` remote multi-user layer is intentionally stubbed.
- web console is not required for operation.
