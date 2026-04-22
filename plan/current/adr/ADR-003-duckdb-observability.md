---
title: "ADR-003: DuckDB for Observability"
summary: "DuckDB is chosen as the observability store (obs.db) — an append-only event log for all swarm activity."
status: "accepted"
version: "0.1.0"
---
# ADR-003 - DuckDB for Observability

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** obs
**Status:** accepted
**Date:** 2026-04-22

---

## Context

The swarm needs a persistent, queryable event log for all task state transitions and system events — separate from application state. Requirements: append-only writes (no updates/deletes), efficient time-range queries (operator reviewing a run's timeline), and good performance under rapid sequential appends. The operator queries the log ad-hoc via DuckDB CLI; no application query API is needed. The log must not share a file with swarm.db.

---

## Decision

Use **DuckDB** (`obs.db`) for the observability event log.

DuckDB is an embedded columnar database purpose-built for analytical, append-heavy workloads. An event log is exactly this pattern: sequential appends, batch time-range queries, no row-level updates. DuckDB's columnar storage compresses repetitive event data (repeated task_type, status values) efficiently. The operator can query obs.db directly with the DuckDB CLI without any application layer. The Go driver (`marcboeker/go-duckdb`) is functional for write-only workloads. DuckDB is embedded like SQLite — no server, no configuration.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| SQLite (same as app state) | One technology, simpler setup | Row-based storage is less efficient for append-heavy analytical queries; mixing app state and observability violates separation of concerns | Wrong fit for append-heavy analytical pattern; see ADR-009 |
| TimescaleDB | Purpose-built time-series, excellent temporal queries | Requires PostgreSQL server; heavy dependency for a local tool | Server dependency; overkill |
| Plain log files | Zero dependencies, human-readable | Not queryable; no schema; difficult to analyse runs programmatically | No structured query capability |
| InfluxDB | Purpose-built time-series | Requires running server; complex setup | Server dependency |

---

## Affected Components

| Component | Impact |
|---|---|
| obs | Owns and exclusively writes obs.db using DuckDB Go driver |
| orchestrator | Calls obs.WriteEvent() — never touches obs.db directly |

---

## Consequences

**Positive:**
- Columnar storage is efficient for append-heavy event log workloads
- Operator can query obs.db directly with DuckDB CLI — no application layer needed
- Embedded — no server, single file, trivially portable
- Compresses repetitive column values (task_type, status) well

**Negative:**
- `go-duckdb` requires CGO, complicating cross-platform builds
- DuckDB Go driver is less battle-tested than SQLite drivers for write-heavy workloads
- DuckDB is optimised for read-heavy analytics; high-frequency small appends (every task transition) may not fully leverage columnar advantages

**Risks:**
- DuckDB Go driver maturity: `go-duckdb` is functional but less mature than `mattn/go-sqlite3`; may have edge cases under rapid concurrent appends (mitigated by serialising obs writes through a single channel)

---

## Related ADRs

- ADR-001 - related-to (CGO requirement from DuckDB driver relevant to Go build toolchain)
- ADR-009 - depends-on (decision to keep observability separate from app state)

---
