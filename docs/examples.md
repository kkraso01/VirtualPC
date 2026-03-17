# Examples (release/v1)

## A) Ollama + coding skill + built-in VM tools

```bash
vpc provider inspect ollama-local
vpc agent start --machine <id> --goal "write a test for dispatcher allowlist" --provider-profile ollama-local --skill coding --approval
vpc agent approvals
```

Expected state: session enters `running`, dangerous calls show as pending approvals, approving resumes execution.

## B) OpenAI + coding skill

```bash
export OPENAI_API_KEY=...
vpc provider inspect openai-default
vpc agent start --machine <id> --goal "fix failing go test in capabilities" --provider-profile openai-default --skill coding --approval
```

Expected state: tool-call loop with run/read/write tools constrained by policy.

## C) Anthropic + research skill + MCP

```bash
export ANTHROPIC_API_KEY=...
vpc mcp list
vpc agent start --machine <id> --goal "summarize open issues" --provider-profile anthropic-default --skill research --mcp github --approval
```

Expected state: MCP resource/tool calls resolve through capability registry and dispatcher.

## D) Custom HTTP tool (`docs_search`)

```bash
vpc tool inspect docs_search
vpc agent start --machine <id> --goal "find runtime docs for snapshot" --provider-profile openai-default --skill research --approval
```

Expected state: HTTP call allowed only for host allowlist in manifest; blocked host attempts are denied.
