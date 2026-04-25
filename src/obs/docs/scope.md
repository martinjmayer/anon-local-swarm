# Scope — obs

## In scope

- DuckDB-backed append-only event log (`obs.db`)
- Structured stdout mirror via `log/slog`
- Credential scrubbing before every write
- Graceful failure handling (write errors logged, never crash caller)
- Schema initialisation and idempotent migration

## Out of scope

- Application state storage — owned by orchestrator (`swarm.db`)
- Query API or dashboard — operator queries `obs.db` directly
- Log rotation or archival
- Alerting or notifications
- Reading from `obs.db` at runtime

## Deferred

- Log rotation / archival policy
- obs.db size threshold warning
- Structured query helper for common operator queries
