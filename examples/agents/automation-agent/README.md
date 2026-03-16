# Automation Agent Example

Goal: `start service and verify health`.

```bash
vpc agent start --machine <machine-id> --goal "start service and verify health"
```

Expected tools: `start_service`, `run_command`, `snapshot_machine`.
