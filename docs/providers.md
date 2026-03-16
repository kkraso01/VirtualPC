# Provider profiles

Provider profiles are reusable YAML files in `agent/providers/profiles/`.

Use:
- `vpc provider list`
- `vpc provider inspect ollama`
- `vpc agent start --machine <id> --goal "..." --provider-profile ollama`

Controller startup fails fast if required capabilities (for example, tool calling) are unsupported by the selected profile.
