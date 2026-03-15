# Operations

## Host prerequisites

- Linux host with KVM enabled
- Firecracker binary and compatible kernel/rootfs assets
- container runtime for daemon packaging

## Local run

- `make build`
- `make dev-up`
- `make smoke`

## Data paths

- daemon state: `./data/state.json`
- runtime artifacts: `./data/firecracker/<machine-id>/`
