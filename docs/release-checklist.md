# VirtualPC v1.0 Release Checklist (Single-node Self-hosted)

Status legend:
- ✅ Passing
- ❌ Failing / hard blocker

## Hard release gates

| Gate | Status | Evidence |
|---|---|---|
| Network policy enforcement (offline/nat/allowlist) | ✅ | `tests/reliability/network_policy_validation_test.go` + `scripts/e2e-real-host.sh` run deterministic in-guest egress probes and fail on mismatch. |
| Reliability thresholds | ✅ | `scripts/soak.sh` enforces boot/agent/exec >=99% and snapshot+fork >=98%; exits non-zero when below thresholds. |
| Crash recovery guarantee | ✅ | `scripts/e2e-real-host.sh` and `scripts/soak.sh` kill Firecracker + guest-agent and validate reconciliation, plus orphan/runtime cleanup counters. |
| Snapshot/fork integrity | ✅ | `scripts/e2e-real-host.sh` and `scripts/soak.sh` verify fork reads snapshot state and parent/fork isolation after mutation. |
| Install/upgrade/uninstall validation | ✅ | `scripts/install.sh`, `scripts/uninstall.sh`, and `tests/reliability/install_uninstall_validation_test.go` validate preflight/install/cleanup behavior. |
| Release build completeness | ✅ | `scripts/build-release.sh` emits daemon/CLI/agent/guest image + manifest + checksums with reproducible tar flags. |

## Launch decision

A build is **RELEASE READY** only when `make e2e` and `make soak` both pass on the release host.
Any gate failure is **RELEASE BLOCKED**.
