---
title: "Requirement: REQ-006 - Viability Review and Cascading Prune"
summary: "VIABILITY_REVIEW gating and recursive pruning of failed branches"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-006 - Viability Review and Cascading Prune

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — Gatekeeper, Cascading Pruning, VIABILITY_THRESHOLD: 5
**Priority:** must-have

---

## Functional Requirements
- Every development branch MUST have at least one `VIABILITY_REVIEW` task tethered to it before implementation tasks are unblocked
- A `VIABILITY_REVIEW` task MUST return a numeric score between 0 and 10 in its output
- If the score is ≥ `VIABILITY_THRESHOLD` (5), the task completes with status `completed` and its dependents are unblocked
- If the score is < `VIABILITY_THRESHOLD`, the orchestrator MUST recursively mark all tasks in the branch as `pruned` (cascading prune)
- Cascading prune MUST traverse the full dependency graph depth-first and mark every descendant task `pruned` regardless of their current status (`pending`, `blocked`, `active`)
- A task with `status = active` at prune time MUST be cancelled (Ollama call aborted if possible, or result discarded on return)
- Pruned tasks MUST NOT be re-queued or reactivated
- The prune event MUST be logged to `obs.db` with the triggering VIABILITY_REVIEW score and the count of pruned tasks

## Acceptance Criteria
- [ ] VIABILITY_REVIEW score of 4 triggers cascading prune of all descendants
- [ ] VIABILITY_REVIEW score of 5 unblocks dependents normally
- [ ] Pruned tasks never appear in the scheduler's pending queue
- [ ] Cascading prune handles branches with depth 5 (max depth) correctly
- [ ] Prune event logged to obs.db with score and pruned task count
- [ ] Active task at prune time is marked pruned; its output is discarded

## Dependencies
- REQ-002 (scheduler must respect pruned status)
- REQ-011 (observability logs prune events)
