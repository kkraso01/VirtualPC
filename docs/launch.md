# VirtualPC Launch Mode

## Scope
This release candidate is **single-node self-hosted** only.

## Supported environment
- Linux host with KVM (`/dev/kvm` required)
- Firecracker installed and accessible in PATH (or configured via `VPCD_FIRECRACKER_BIN`)
- systemd-managed daemon (`packaging/systemd/virtualpcd.service`)

## Runtime architecture in this launch
- `virtualpcd` daemon
- `vpc` CLI
- `vpc-agent`
- Firecracker VM runtime and guest-agent trusted path
- No host/dev fallback for real runtime operations

## Operator install and validation
1. Build binaries: `./scripts/build-binaries.sh`
2. Build guest assets: `./scripts/build-guest-image.sh`
3. Install: `sudo ./scripts/install.sh`
4. Start daemon: `sudo systemctl enable --now virtualpcd`
5. Validate host readiness: `/opt/virtualpc/bin/vpc doctor`
6. Run real-host e2e: `VPC_E2E_REAL_HOST=1 go test ./tests/e2e -run TestRealHostE2E -v`
7. Run soak: `VPC_SOAK_REAL_HOST=1 SOAK_LOOPS=25 ./scripts/soak.sh`

## Production-ready in this RC
- Core machine lifecycle create/start/exec/stop/destroy
- File upload/download command path
- Service lifecycle commands (create/list/logs/stop/destroy)
- Snapshot create and fork API path
- Install/upgrade/uninstall scripts
- Systemd operator path

## Alpha / known limitations
- Network policy enforcement validation (offline/nat/allowlist) is not fully complete; allowlist gate is blocking public launch.
- Crash-recovery and reconciliation under forced Firecracker failures need stricter automated assertions.
- Snapshot/fork deep data integrity checks are not yet exhaustive.
