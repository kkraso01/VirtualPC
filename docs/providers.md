# Provider profiles

Provider profiles are reusable YAML files in `agent/providers/profiles/`.

Use:
- `vpc provider list`
- `vpc provider inspect ollama`
- `vpc agent start --machine <id> --goal "..." --provider-profile ollama`

Controller startup fails fast if required capabilities (for example, tool calling) are unsupported by the selected profile.

## Provider profiles in capability execution

Provider profiles continue to select model/provider behavior; capability execution remains controller-mediated with a normalized effective tool inventory passed to the provider.
