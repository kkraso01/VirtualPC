# Release Checklist (Single-node Self-hosted RC)

Status legend:
- ✅ Passing now
- ❌ Failing / launch blocker
- ⚠️ Deferred with documented risk

## Hard release gates

| Gate | Status | Evidence / notes |
|---|---|---|
| Installability | ✅ | `scripts/install.sh`, `scripts/upgrade.sh`, `scripts/uninstall.sh` added for Linux hosts with systemd flow. |
| Host preflight (`vpc doctor`) | ✅ | Real-host e2e starts with doctor and hard-fails if not healthy. |
| VM boot success rate | ⚠️ | Covered by `scripts/e2e-real-host.sh` and `scripts/soak.sh`; threshold policy still operator-defined. |
| Guest-agent connect success rate | ⚠️ | Proxy validated via `machine exec` in e2e/soak; statistical SLO threshold not yet encoded. |
| File transfer success | ✅ | e2e and soak include upload/download and integrity checks. |
| In-guest service lifecycle success | ✅ | e2e + soak exercise create/list/logs/stop/destroy. |
| Snapshot/fork integrity | ⚠️ | Snapshot/fork executed; deeper disk-content integrity verification still pending. |
| Daemon restart reconciliation | ✅ | e2e restarts daemon and validates machine state still inspectable. |
| Crash recovery | ⚠️ | Soak contains destructive loops and process-kill cleanup accounting; automated Firecracker kill/reconcile assertions still incomplete. |
| Network policy enforcement | ❌ | Validation test marks allowlist enforcement path as blocker until full offline/nat/allowlist probes are implemented. |
| Log/artifact collection | ✅ | E2E and soak persist logs and reports under `.tmp/`. |
| Uninstall/cleanup | ⚠️ | `uninstall.sh` stops services and removes sockets/processes; deep host-network artifact cleanup remains operator-verified. |

## Launch decision guidance

Public launch requires all hard gates to be ✅.
Current blockers:
1. Network policy enforcement validation for offline/nat/allowlist.
2. Quantified boot/agent success-rate thresholds.
3. Full crash-recovery assertions under forced runtime failure.
