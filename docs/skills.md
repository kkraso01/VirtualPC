# Skills

Skill packs live in `skills/<name>/` and support `SKILL.md`, `prompt.md`, `tools.yaml`, `policies.yaml`, and optional `resources/`.

Use:
- `vpc skill list`
- `vpc skill inspect coding`
- `vpc agent start --machine <id> --goal "..." --skill coding`

## Runtime skill application

Attached skills now affect live sessions via prompt composition, tool filtering, policy overlays, and resource injection.
Effective capability inventory is computed per session from selected skills + enabled MCP + base registry.
