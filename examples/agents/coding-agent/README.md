# Coding Agent Example

Goal: `fix failing tests`.

```bash
vpc agent start --machine <machine-id> --goal "fix failing tests"
```

The controller will typically use `run_command`, `read_file`, and `write_file` while respecting command policy and filesystem guardrails.
