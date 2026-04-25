---
title: "Requirement: REQ-002 - DAG Scheduler"
summary: "DAG scheduler loop, task dispatch, and blocked_by resolution"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-002 - DAG Scheduler

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — DAG Scheduler, blocked_by dependency resolution
**Priority:** must-have

---

## Functional Requirements
- The scheduler MUST poll `swarm.db` every 2 seconds for tasks with `status = pending` and no unresolved `blocked_by` dependency
- A task is considered unblocked when its `blocked_by` task has `status = completed` (specifically `COMPLETED_PASS` for VIABILITY_REVIEW gates)
- The scheduler MUST dispatch at most one task per model at a time (sticky model strategy — see REQ-004)
- The scheduler MUST NOT dispatch a task whose `blocked_by` task has `status = pruned`, `active`, `blocked`, or `failed`
- When a task completes, the scheduler MUST immediately re-evaluate its dependents for unblocking
- The scheduler MUST enforce the `MAX_DEPTH = 5` hard limit; tasks at depth > 5 MUST NOT be dispatched and MUST be marked `pruned`
- Task dispatch MUST be atomic: a task transitions from `pending` to `active` in a single DB write before any Ollama call is made
- The scheduler MUST handle concurrent task completions without race conditions (SQLite write serialisation)

## Acceptance Criteria
- [ ] Task with blocked_by pointing to a completed task is picked up on next poll
- [ ] Task with blocked_by pointing to a pruned task is never dispatched
- [ ] Task at depth 6 is marked pruned immediately on detection
- [ ] Two tasks of different model_tags can be dispatched concurrently
- [ ] Task transitions to active atomically before Ollama is called
- [ ] Scheduler survives a task that panics (recovers and continues polling)

## Dependencies
- REQ-001 (seed task must exist to schedule)
- REQ-003 (Ollama client required for dispatch)
- REQ-006 (prune logic triggered by viability review)
