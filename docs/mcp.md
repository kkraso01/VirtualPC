# MCP integration

MCP servers are configured in `mcp/servers.yaml` and normalized into the same capability registry path as built-in tools.

Supported server modes:
- `stdio`
- `remote`

Use:
- `vpc mcp list`
- `vpc mcp inspect github`

## MCP runtime invocation

The controller now uses an MCP runtime with server rehydration and unified invocation for:
- MCP tools
- MCP resources
- MCP prompts

MCP invocations are routed through the same capability dispatcher and approval/policy checks.

## Release/v1 hardening

MCP invocations now enforce per-call timeout, explicit unavailable-server errors, and response-size limits for tools/resources/prompts.
