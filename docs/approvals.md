# Approvals

Approvals are durable controller-side records for capability executions requiring operator confirmation.

## Commands

```bash
vpc agent approvals [session-id]
vpc agent approve <session-id> <approval-id>
vpc agent deny <session-id> <approval-id>
```

## Behavior

- Pending approvals are persisted in sqlite and survive controller restarts.
- Approval IDs are stable and human-readable (`appr-<timestamp-ns>`).
- Approved actions resume next loop iteration.
- Denied actions return explicit blocked errors and are audit-logged.

## Operator workflow

1. Run `vpc agent approvals` to list pending requests with capability/action context.
2. Approve or deny with session + approval ID.
3. Inspect `vpc agent logs <session-id>` for resume/blocked results.
