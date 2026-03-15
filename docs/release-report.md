# VirtualPC v1.0 Release Report

## Final release decision

**STATE B: RELEASE BLOCKED**

## Gate results

- Network policy enforcement (offline/nat/allowlist): **BLOCKED IN CURRENT HOST RUN** (runtime binaries available but Firecracker missing, so e2e probes cannot complete).
- Reliability thresholds: **BLOCKED IN CURRENT HOST RUN** (`make soak` cannot execute real VM loops without Firecracker).
- Crash recovery guarantee: **IMPLEMENTED IN HARNESS**, blocked in this host execution because VM runtime cannot launch.
- Snapshot/fork integrity: **IMPLEMENTED IN HARNESS**, blocked in this host execution because VM runtime cannot launch.
- Install/upgrade/uninstall validation: **PASSING in automated script test harness**.
- Release build: **IMPLEMENTED** (`scripts/build-release.sh` now emits manifest + checksums + reproducible tar flags).

## Exact failing gates from this run

1. `make e2e` failed because Firecracker binary was unavailable on host (`/usr/local/bin/firecracker` missing), so real-host network policy and crash/snapshot checks could not pass.
2. `make soak` failed for the same host runtime dependency, preventing threshold validation.

A release host with KVM + Firecracker installed is required to move this report to **STATE A: RELEASE READY**.
