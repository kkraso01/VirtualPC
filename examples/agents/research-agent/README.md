# Research Agent Example

Goal: `inspect runtime and summarize bottlenecks`.

```bash
vpc agent start --machine <machine-id> --goal "inspect runtime and summarize bottlenecks"
```

The agent should use non-destructive read-only commands and avoid network scanning.
