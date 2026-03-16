# Capability Registry

VirtualPC's optional agent controller now resolves capabilities from a single registry layer before any execution.

## Architecture

LLM -> Agent Controller -> Capability Registry (built-in tools, skills, MCP, custom tools, provider profiles, policy bindings) -> VirtualPC API -> virtualpcd -> Firecracker VM.

## Execution placement

- `controller`: prompt assembly, provider SDK calls, policy checks.
- `vm`: machine-local operations (shell, files, process state).
- `sidecar`: local stdio MCP servers and reusable helper services.
- `remote`: remote APIs, remote MCP endpoints, hosted providers.

## Security model

Capabilities are deny-by-default unless registered and enabled. Every capability includes policy metadata used for approval checks, filesystem scopes, and network posture.

## Runtime execution (v1)

Capabilities are now executed through a single dispatcher path (`resolve -> validate -> policy -> approval -> execute -> audit`).
Execution supports built-in VM tools, custom local/HTTP tools, MCP tool/resource/prompt capabilities, and skill overlays.
Unknown or disabled capabilities are denied by default.
