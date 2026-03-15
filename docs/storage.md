# Storage

- Durable daemon metadata is persisted to local state file in launch scaffold.
- Launch target durable systems:
  - PostgreSQL for relational metadata
  - MinIO for artifacts/snapshots/log bundles
  - NATS JetStream for event replay streams
  - Temporal for workflow durability
