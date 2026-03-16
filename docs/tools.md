# Custom tools

Custom tools are manifest-driven (`tools/local/*.yaml`, `tools/http/*.yaml`).

Required fields: `name`, `description`, `schema`, `backend_type`, command/URL, execution location, policy requirements, timeout and retries.

Use:
- `vpc tool list`
- `vpc tool inspect jira_search`
