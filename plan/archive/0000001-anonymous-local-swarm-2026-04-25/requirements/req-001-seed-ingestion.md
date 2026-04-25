---
title: "Requirement: REQ-001 - Seed Ingestion"
summary: "Seed task creation and project initialisation in swarm.db"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-001 - Seed Ingestion

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** AS-001 — Human on the Loop seeds a problem; swarm runs until done
**Priority:** must-have

---

## Functional Requirements
- The orchestrator MUST accept a seed task inserted directly into `swarm.db` as a row in the `tasks` table with `type = DECOMPOSE` and `status = pending`
- A seed task MUST reference a valid `project_id` in the `projects` table
- The seed task `payload` MUST contain the high-level goal as a plain-text string
- The seed task `depth` MUST be set to `0`
- The orchestrator MUST detect the new pending task within 2 seconds of insertion (scheduler poll interval)
- The orchestrator MUST NOT require a restart to pick up a new seed task

## Acceptance Criteria
- [ ] Insert a DECOMPOSE task with status=pending; scheduler picks it up within 2 seconds
- [ ] Seed task with missing project_id is rejected (FK constraint)
- [ ] Seed task with depth ≠ 0 is rejected or normalised to 0
- [ ] Orchestrator running with no tasks waits idle without crashing

## Dependencies
- REQ-002 (DAG scheduler must be running to detect the seed)
