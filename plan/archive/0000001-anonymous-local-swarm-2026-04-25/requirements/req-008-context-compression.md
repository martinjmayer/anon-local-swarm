---
title: "Requirement: REQ-008 - Context Compression"
summary: "Context threshold detection and SUMMARY task emission at 70% capacity"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-008 - Context Compression

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — CONTEXT_THRESHOLD: 0.70, SUMMARY task, Context Sidecar Pattern
**Priority:** must-have

---

## Functional Requirements
- Each task branch maintains a context sidecar — a JSON file at `context_path` storing accumulated context (entities, constraints, prior summaries)
- Before dispatching a task, the orchestrator MUST estimate the token count of the branch context sidecar
- If the estimated token count exceeds `CONTEXT_THRESHOLD = 70%` of the active model's context window, the orchestrator MUST emit a `SUMMARY` task before continuing
- The `SUMMARY` task MUST be dispatched to `smollm3:135m` with the full branch context as input
- The `SUMMARY` task output MUST replace the branch context sidecar content
- `SUMMARY` and `DECOMPOSE` tasks do NOT increment branch `depth`; they inherit the parent task's depth
- The context sidecar MUST be stored as a JSON file on disk at the path recorded in `tasks.context_path`

## Acceptance Criteria
- [ ] Branch context at 71% of model limit triggers SUMMARY task before next dispatch
- [ ] Branch context at 69% proceeds without SUMMARY task
- [ ] SUMMARY task output replaces (not appends to) the context sidecar
- [ ] SUMMARY task depth equals parent task depth (no increment)
- [ ] Context sidecar JSON file exists at the path stored in tasks.context_path

## Dependencies
- REQ-002 (scheduler checks context before dispatch)
- REQ-003 (SUMMARY dispatched via Ollama client to smollm3)
