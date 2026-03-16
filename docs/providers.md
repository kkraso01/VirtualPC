# Agent Provider Support Matrix

## Architecture

```text
LLM
 ↓
Agent Controller
 ↓
VirtualPC API
 ↓
virtualpcd
 ↓
Firecracker VM
```

The controller is optional and does not alter the VirtualPC runtime internals.

## Provider strategy

| Provider | Type | Transport | Tool calling | Notes |
|---|---|---|---|---|
| OpenAI | Official SDK | `github.com/openai/openai-go` | Yes | Uses Chat Completions tool calls. |
| Anthropic | Official SDK | `github.com/anthropics/anthropic-sdk-go` | Yes | Uses Messages API `tool_use` blocks. |
| OpenAI-compatible | Compatibility adapter | `/v1/chat/completions` | Configurable | Generic path used for Ollama/vLLM/local endpoints. |
| Ollama | OpenAI-compatible profile | `http://localhost:11434/v1` | Model dependent | Experimental OpenAI-compat support; non-stateful controller mode recommended. |
| vLLM | OpenAI-compatible profile | `http://localhost:8000/v1` | Yes (with flags) | Requires tool-call parser flags for auto tool choice.

## Capability flags

```yaml
provider: openai_compatible
base_url: http://localhost:11434/v1
api_key: ""
model: llama3.1
supports_responses_api: false
supports_chat_completions: true
supports_tool_calling: true
supports_stateful_responses: false
```

Controller behavior:
- If `supports_tool_calling: false`, start fails clearly.
- If responses API is unavailable, Chat Completions mode is used.
- Non-stateful providers rely on persisted local session history.

## Verified references

- OpenAI official Go SDK README and function calling examples: https://github.com/openai/openai-go
- Anthropic official Go SDK README and tool examples: https://github.com/anthropics/anthropic-sdk-go
- Ollama OpenAI compatibility announcement: https://ollama.com/blog/openai-compatibility
- Ollama API tool-calling docs (`tools`, `tool_calls`): https://github.com/ollama/ollama/blob/main/docs/api.md
- vLLM OpenAI-compatible serving docs: https://github.com/vllm-project/vllm/blob/main/docs/serving/openai_compatible_server.md
- vLLM tool-calling flags and parser requirements: https://github.com/vllm-project/vllm/blob/main/docs/features/tool_calling.md
