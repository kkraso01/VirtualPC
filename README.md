# VirtualPC

VirtualPC is a daemon-first, CLI-first runtime for running AI agents inside Firecracker-backed machines.

Runtime architecture (unchanged):

```text
LLM
 ↓
Agent Controller (optional)
 ↓
Capability Registry
 ↓
Capability Dispatcher
 ↓
Execution backends
 ↓
VirtualPC API
 ↓
virtualpcd
 ↓
Firecracker VM
```

## Fastest useful path (release/v1)

1. Install/build
```bash
make build
./scripts/install.sh
```

2. Start daemon
```bash
virtualpcd
vpc daemon status
```

3. Create/start machine
```bash
MID=$(vpc machine create --profile minimal-shell | jq -r '.id')
vpc machine start "$MID"
```

4. Start agent with provider + skill
```bash
vpc provider list
vpc skill list
vpc agent start --machine "$MID" --goal "fix a failing test" --provider-profile ollama-local --skill coding --approval
```

5. Inspect approvals/logs
```bash
vpc agent approvals
vpc agent approve <session-id> <approval-id>
vpc agent logs <session-id>
```

See:
- `docs/getting-started.md`
- `docs/examples.md`
- `docs/approvals.md`
