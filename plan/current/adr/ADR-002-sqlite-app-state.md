---
title: "ADR-002: SQLite for Application State"
summary: "SQLite is chosen as the database for swarm.db — the authoritative store for projects, tasks, and task_candidates."
status: "accepted"
version: "0.1.0"
---
# ADR-002 - SQLite for Application State

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator
**Status:** accepted
**Date:** 2026-04-22

---

## Context

The swarm needs a persistent store for application state: projects, tasks (with DAG dependency tracking), and task_candidates (variational execution). The store must support: relational queries (blocked_by resolution, cascading prune traversal), concurrent reads during scheduler polling, atomic writes for task status transitions, and foreign key constraints. The system is local-only, single-machine, and the data volume per project run is small (hundreds to low thousands of rows).

---

## Decision

Use **SQLite** (`swarm.db`) for application state.

SQLite is embedded — no separate server process, no network, no configuration. It handles the relational model (projects → tasks → task_candidates with FK constraints) and supports WAL mode for concurrent read/write without contention. The entire database is a single file, trivially backed up before a project run. The data volume (hundreds of rows per run) is well within SQLite's performance envelope. The Go SQLite driver is mature and widely used.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| PostgreSQL | Full RDBMS, excellent concurrency, rich query planner | Requires a running server; operator setup burden; overkill for local single-process workload | Server dependency eliminates the local-first, zero-dependency goal |
| DuckDB (for both) | Single DB technology for both app state and observability | DuckDB is optimised for analytical append-heavy workloads, not OLTP row-level updates; task status updates are frequent point writes | Wrong workload fit; task status transitions are OLTP, not analytical |
| Plain JSON files | Zero dependencies | No transactions, no FK constraints, no concurrent access safety, no query capability | Cannot safely handle concurrent reads/writes or relational integrity |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | Owns and exclusively writes swarm.db |
| obs | Must NOT write to swarm.db — owns obs.db separately (see ADR-009) |

---

## Consequences

**Positive:**
- Embedded — zero operator setup; single file backup
- WAL mode enables concurrent reads (scheduler) while writes occur (task transitions)
- FK constraints enforce data integrity (task → project, candidate → task)
- Familiar SQL query model for DAG traversal and prune logic

**Negative:**
- WAL mode still serialises writes — high-frequency concurrent writes could create queue contention during cascading prune
- SQLite is not suitable if the swarm ever needs multi-machine distribution

**Risks:**
- Cascading prune on a deep branch triggers many rapid writes; SQLite WAL serialisation could create a brief bottleneck (mitigated by single-goroutine prune execution per ADR-001 risk note)

---

## Related ADRs

- ADR-003 - related-to (DuckDB chosen separately for observability, keeping concerns separated)
- ADR-009 - depends-on (separation of app state and observability databases)

---
