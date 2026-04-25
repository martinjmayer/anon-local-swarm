---
title: "ADR-008: DAG + blocked_by Dependency Model"
summary: "Task dependencies are modelled as a DAG via a blocked_by foreign key on the tasks table, enabling declarative dependency resolution without a separate graph structure."
status: "accepted"
version: "0.1.0"
---
# ADR-008 - DAG + blocked_by Dependency Model

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator
**Status:** accepted
**Date:** 2026-04-22

---

## Context

The swarm's tasks have dependencies — a GO_CODE task cannot begin until its VIABILITY_REVIEW passes; a RESEARCH task cannot begin until DECOMPOSE has produced its subtasks. A model was needed for expressing and resolving these dependencies. The model must support: blocking a task until a prerequisite completes, cascading prune when a prerequisite fails, and depth tracking for recursion control. The dependency graph must be acyclic (no circular dependencies) by construction.

---

## Decision

Model task dependencies as a **DAG via a `blocked_by` foreign key** on the `tasks` table — a self-referential FK pointing to the prerequisite task.

A task's `status = blocked` means it has an unresolved `blocked_by`. The scheduler query for dispatchable tasks is: `WHERE status = 'pending' AND (blocked_by IS NULL OR blocked_by IN (SELECT id FROM tasks WHERE status = 'completed'))`. When a task completes, the scheduler unblocks its dependents by checking `blocked_by = completed_task_id`. Cascading prune traverses the `blocked_by` graph recursively. This is declarative — the dependency structure is intrinsic to the data model, not managed by a separate graph library.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| Separate edges table (task_id, depends_on_id) | Supports many-to-many dependencies | More complex queries; more join overhead; overkill for the swarm's one-to-one dependency pattern | The swarm's tasks each have at most one blocking prerequisite (the viability review); many-to-many adds unneeded complexity |
| In-memory graph (e.g. topological sort at startup) | Fast traversal; no DB queries for dependency resolution | State lost on restart; must rebuild graph from DB on recovery | Recovery complexity; DB is the source of truth anyway |
| Workflow engine (e.g. Temporal, Cadence) | Full-featured DAG scheduling, fault tolerance, replay | Server dependency; heavy operational footprint; overkill for a local tool | Violates local-first, zero-server constraint |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | Scheduler query uses blocked_by for dispatch decisions; prune engine traverses via blocked_by |
| swarm.db | `tasks.blocked_by` FK is the structural expression of the DAG |

---

## Consequences

**Positive:**
- Dependency structure is intrinsic to the data model — no separate graph to maintain
- Recovery is automatic — on restart, scheduler re-reads DB and resolves current state correctly
- Simple scheduler query: one SQL WHERE clause resolves all blocking logic
- Cascading prune traversal is a standard recursive SQL query or Go recursion over FK

**Negative:**
- Self-referential FK means circular dependencies are possible if task creation logic has a bug — must be prevented in application code, not DB constraints (SQLite does not enforce acyclicity)
- `blocked_by` is a single FK — a task can only be directly blocked by one predecessor; complex multi-prerequisite patterns would require the edges table alternative

**Risks:**
- Bug in DECOMPOSE output parser creates a circular dependency (A blocked_by B, B blocked_by A) — scheduler would never dispatch either; mitigated by cycle detection in the scheduler's unblocking query

---

## Related ADRs

- ADR-002 - depends-on (SQLite stores the DAG structure)

---
