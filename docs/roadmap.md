# Roadmap

Completed in this refactor:
- daemon-first and CLI-first architecture
- Firecracker runtime manager abstraction
- machine/project/service/snapshot/task durable models
- local Unix-socket API with CLI workflows

Stubbed or partial (explicitly non-core for this pass):
- remote gateway multi-user auth boundary (`vpc-gateway`)
- full guest file streaming and interactive PTY tunnel
- full Firecracker snapshot restore orchestration
- production policy enforcement engine and secrets manager integration
