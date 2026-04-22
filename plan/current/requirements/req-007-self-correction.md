---
title: "Requirement: REQ-007 - Self-Correction Loop"
summary: "Go build validation and automatic self-correction task emission on failure"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-007 - Self-Correction Loop

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — Go Build Check, Self-Healing, os/exec wrapper
**Priority:** must-have

---

## Functional Requirements
- After every `GO_CODE` task completes, the orchestrator MUST run `go build` on the generated code via `os/exec`
- If `go build` succeeds, the task is marked `completed` and dependents are unblocked
- If `go build` fails, the orchestrator MUST emit a new `GO_CODE` task with:
  - The original code as part of the payload
  - The full `stderr` compiler output appended to the payload
  - `parent_id` set to the failing task
  - `depth` incremented by 1 from the failing task
- The self-correction loop MUST halt after 5 consecutive failures on the same branch (stored in `meta_data`)
- On halt, the task chain MUST be marked `failed` and an event logged to `obs.db`
- The depth increment on self-correction tasks counts toward the `MAX_DEPTH = 5` hard limit

## Acceptance Criteria
- [ ] Successful go build marks task completed and unblocks dependents
- [ ] Failed go build emits a new GO_CODE task with stderr in payload
- [ ] Self-correction task has parent_id pointing to the failing task
- [ ] After 5 consecutive failures, chain is marked failed (not pruned)
- [ ] Self-correction depth increments are capped by MAX_DEPTH
- [ ] Halt event logged to obs.db with failure count and final stderr

## Dependencies
- REQ-002 (scheduler dispatches self-correction tasks)
- REQ-003 (Ollama client handles the new GO_CODE task)
- REQ-011 (observability logs failures)
