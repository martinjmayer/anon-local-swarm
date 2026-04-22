---
title: "Requirement: REQ-011 - Observability"
summary: "DuckDB event log and stdout mirror via log/slog"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-011 - Observability

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — DuckDB obs.db, log/slog stdout, every task state transition logged
**Priority:** must-have

---

## Functional Requirements
- Every task state transition MUST be written as a row to `obs.db` (DuckDB) with: timestamp, task_id, task_type, from_status, to_status, model_tag, depth, and detail (JSON)
- The following additional events MUST also be logged: model load/unload (VRAM swap), cascading prune triggered, self-correction attempt, context compression triggered, MCP tool call (success/failure), output artifact deposited
- Every event written to DuckDB MUST also be mirrored to stdout via `log/slog` in structured text format
- `obs.db` MUST be a separate file from `swarm.db` — observability data and application state MUST NOT share a database
- Credentials (API keys, tokens) MUST NEVER appear in any log event, DuckDB row, or stdout output
- `obs.db` MUST be append-only — no updates or deletes to event rows
- The obs component MUST handle DuckDB write failures gracefully — a logging failure MUST NOT crash the orchestrator

## Acceptance Criteria
- [ ] Every task status change produces a row in obs.db
- [ ] obs.db and swarm.db are separate files
- [ ] BRAVE_API_KEY does not appear in any log output
- [ ] DuckDB write failure is caught; orchestrator continues running
- [ ] VRAM swap events appear in obs.db
- [ ] Prune events appear in obs.db with score and pruned count
- [ ] All obs.db events also appear on stdout via log/slog

## Dependencies
- None (observability is a cross-cutting concern used by all other components)
