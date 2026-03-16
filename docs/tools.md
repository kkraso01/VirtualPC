# Custom tools

Custom tools are manifest-driven (`tools/local/*.yaml`, `tools/http/*.yaml`).

Required fields: `name`, `description`, `schema`, `backend_type`, command/URL, execution location, policy requirements, timeout and retries.

Use:
- `vpc tool list`
- `vpc tool inspect jira_search`

## Executable custom tools

Custom tool manifests now execute at runtime via normalized runners:
- local command-backed tools (validated args, timeout, allowlisted executable, restricted env)
- HTTP-backed tools (validated request templating, destination allowlist, timeout)

All tool calls pass through capability policy and audit.
