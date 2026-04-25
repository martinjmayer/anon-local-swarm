# Interface Contract — obs

## Public surface

### `obs.Open(path string) (*Logger, error)`
Opens (or creates) `obs.db` at `path`, applies the schema, and returns a ready `Logger`. Idempotent — safe to call on an existing database.

### `(*Logger).WriteEvent(e Event)`
Persists `e` to `obs.db` and mirrors it to stdout. Never returns an error — failures are logged to stderr and the write is dropped rather than crashing the caller. Callers must not depend on WriteEvent for error propagation.

### `(*Logger).Close() error`
Flushes and closes the DuckDB connection. Must be called on shutdown.

## `Event` struct fields

| Field | Type | Required | Description |
|---|---|---|---|
| `Type` | `EventType` (string) | ✓ | Event classifier — see event types below |
| `TaskID` | `*string` | — | Task ID if event is task-scoped |
| `TaskType` | `*string` | — | Task type (DECOMPOSE, RESEARCH, etc.) |
| `FromStatus` | `*string` | — | Previous task status (for transition events) |
| `ToStatus` | `*string` | — | New task status (for transition events) |
| `ModelTag` | `*string` | — | Ollama model tag |
| `Depth` | `*int` | — | Task depth in the DAG |
| `Detail` | `map[string]any` | — | Arbitrary JSON payload; credential keys are scrubbed before write |

## Event types

| Constant | Description |
|---|---|
| `EventOrchestratorStart` | Swarm process started |
| `EventOrchestratorStop` | Swarm process stopped cleanly |
| `EventTaskTransition` | Task status changed (FromStatus → ToStatus) |
| `EventModelDispatch` | Ollama inference started for a task |
| `EventModelComplete` | Ollama inference completed |
| `EventSelfCorrection` | Go build validation result (success or failure) |
| `EventCascadingPrune` | Task pruned due to failed viability review upstream |
| `EventContextSummary` | SUMMARY task emitted due to context threshold exceeded |
| `EventMCPToolCall` | MCP tool invoked |
| `EventOutputDeposit` | Artifact written to output directory |

## Breaking change policy

Any change to the `Event` struct that removes or renames a field, or changes the `obs.db` schema, requires a migration proposal at `src/obs/docs/migrations/` and human approval before merging. Additive changes (new optional fields, new event types) are non-breaking.

## Consumers

| Consumer | How it uses obs |
|---|---|
| orchestrator | Imports `als/obs`; calls `WriteEvent` throughout the task lifecycle |
| Operator | Queries `obs.db` via DuckDB CLI — never writes to obs.db |
