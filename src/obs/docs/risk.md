# Risk — obs

## Component-scoped risks

| ID | Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| R-obs-001 | DuckDB write performance degrades under high event volume | Low | Low | DuckDB is columnar append-only — handles high write throughput well; write latency monitored via obs timestamps |
| R-obs-002 | Credential leak if scrubber coverage has gaps | Low | High | `scrubMap` redacts credential-keyed fields; `scrubCredentials` in mcp-reddit prevents leakage at source; gap: non-standard key names (e.g. `"message"`) are not scrubbed — documented in security report |
| R-obs-003 | obs.db grows unbounded over long multi-project runs | Low | Medium | No rotation implemented (deferred); operator responsibility to archive or truncate between major runs |
| R-obs-004 | DuckDB schema incompatibility after go-duckdb update | Low | Medium | Schema is applied idempotently using `CREATE TABLE IF NOT EXISTS`; changes require a migration proposal |

## Cross-reference

Risks R-obs-001 and R-obs-002 correspond to R-008 and R-012 in the system risk register (`plan/current/risk-register.md`).
