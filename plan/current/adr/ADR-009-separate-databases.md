---
title: "ADR-009: Separate Databases for App State and Observability"
summary: "swarm.db (SQLite) and obs.db (DuckDB) are kept as separate files with separate ownership, enforcing the data ownership rule between orchestrator and obs."
status: "accepted"
version: "0.1.0"
---
# ADR-009 - Separate Databases for App State and Observability

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator, obs
**Status:** accepted
**Date:** 2026-04-22

---

## Context

Two persistent stores are needed: application state (projects, tasks, candidates) and observability (event log). A decision was needed on whether these should share a single database file or be separated. The Planifest hard limit states data is owned by exactly one component. The two stores have different workload profiles: app state is OLTP (frequent point reads/writes), observability is append-only analytical.

---

## Decision

Keep app state and observability in **separate database files**: `swarm.db` (SQLite, owned by orchestrator) and `obs.db` (DuckDB, owned by obs).

Each component owns its data — no component writes to another's database. This enforces the Planifest data ownership rule at the filesystem level. The different workload profiles map to the right engines: SQLite for OLTP task state, DuckDB for append-heavy event log. A corrupted or deleted `obs.db` does not affect swarm operation. `obs.db` can be purged and recreated between runs without affecting task state.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| Single SQLite file for both | One technology, simpler setup, SQLite handles both patterns adequately | Mixes concerns; events table in same file as tasks table; workload mismatch for event log queries; harder to purge/archive observability independently | Violates data ownership rule; wrong workload fit for events |
| Single DuckDB file for both | One technology; DuckDB can handle OLTP | DuckDB is not optimised for frequent point-update OLTP patterns (task status transitions); risk of write contention | Wrong fit for task state workload |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | Owns swarm.db exclusively; never writes to obs.db |
| obs | Owns obs.db exclusively; never reads from swarm.db |

---

## Consequences

**Positive:**
- Hard enforcement of data ownership — no accidental cross-component writes
- obs.db can be deleted/archived independently of task state
- Different engines chosen for their workload fit
- Cleaner component boundaries — obs is a pure write-only sink from orchestrator's perspective

**Negative:**
- Two database files to manage, backup, and configure
- Two Go drivers (SQLite + DuckDB) in the dependency graph

**Risks:**
- Operator confusion about which file contains what — mitigated by clear naming (`swarm.db` vs `obs.db`) and documentation

---

## Related ADRs

- ADR-002 - extends (SQLite choice for app state)
- ADR-003 - extends (DuckDB choice for observability)

---
