# Purpose — obs

## What this component exists to do

`obs` is the observability backbone of the Autonomous Local Swarm. It owns `obs.db` — a DuckDB append-only event log — and exposes a single write surface (`WriteEvent`) that every other component calls to record state transitions, errors, and operational signals.

Its two responsibilities are inseparable:
1. **Persist** every swarm event to `obs.db` in a structured, queryable form.
2. **Mirror** every event to stdout via `log/slog` so the operator has a live console view without needing a database client.

`obs` deliberately owns nothing else. It does not own application state (`swarm.db` is the orchestrator's concern), does not provide a query API, and does not send alerts. Its job is to be a reliable, credential-safe, append-only sink.

## Why it exists as a separate component

Per ADR-009, the observability database (DuckDB, columnar, analytics-optimised) is kept strictly separate from the application state database (SQLite, transactional). Mixing them would couple the write path of the scheduler to the write path of the event log, creating contention and complicating schema evolution.

Separating `obs` as its own Go package also enforces a clean dependency direction: the orchestrator imports `obs`, never the reverse.

## Who uses it

- **orchestrator** — calls `WriteEvent` on every task state transition, model dispatch, prune event, self-correction, and context summary.
- **Operator** — queries `obs.db` directly with the DuckDB CLI to monitor run progress and diagnose failures.
