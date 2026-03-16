# VirtualPC Optional Agent Controller

## Architecture (unchanged runtime)

```text
LLM
 ↓
Agent Controller
 ↓
VirtualPC API
 ↓
virtualpcd
 ↓
Firecracker VM
```

The controller remains optional and acts as a client above the existing runtime.

## CLI

```bash
vpc agent start --machine <id> --goal "<goal>"
vpc agent start --machine <id> --goal "<goal>" --provider ollama --config agent/config/providers/ollama.yaml
vpc agent attach <session-id>
vpc agent logs <session-id>
vpc agent stop <session-id>
vpc agent list
```

## Session persistence

Session state is durable in sqlite (`~/.virtualpc/agent/sessions.db`) and includes:
- session_id
- provider/model
- machine_id
- history
- tool_calls
- snapshots
- services_started
- files_modified
- iteration_count
- status
- last_error

## Tool execution

- Central tool schema registry (`agent/tools/registry.go`)
- Strict argument validation before execution
- Schema enforcement independent of provider protocol
- Structured tool results (`tool`, `success`, `output`)

## Providers

See `docs/providers.md` for matrix and compatibility notes.

## Safeguards

See `docs/safeguards.md` for command policy, filesystem, loop and budget controls.
