---
title: "Requirement: REQ-005 - Variational Execution"
summary: "Multi-candidate generation and judge-model selection for high-priority tasks"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-005 - Variational Execution

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — Variational Execution, Ensemble Voting, task_candidates table
**Priority:** must-have

---

## Functional Requirements
- For tasks flagged for variational execution, the orchestrator MUST run the same prompt `VARIATION_COUNT = 3` times with varied temperature settings
- Each candidate output MUST be stored as a row in `task_candidates` with `output_text`, `evaluation_score`, and `is_selected = false`
- After all candidates are generated, the orchestrator MUST invoke the judge model (phi-4:mini) to evaluate and select the best candidate
- The judge MUST set `is_selected = true` on exactly one `task_candidates` row
- The selected candidate's output MUST be used as the task's final output
- Non-selected candidates MUST be retained in `task_candidates` for audit purposes (never deleted)
- Tasks eligible for variational execution are determined by task type and project priority (config-driven)
- The variational loop MUST complete within the same task lifecycle — it does not spawn child tasks

## Acceptance Criteria
- [ ] Variational task produces exactly 3 rows in task_candidates
- [ ] Exactly one row has is_selected = true after judge evaluation
- [ ] Non-selected candidates remain in task_candidates after task completion
- [ ] Judge model call uses phi-4:mini regardless of the original task's model_tag
- [ ] Variational execution is skipped when VARIATION_COUNT = 1 (config)

## Dependencies
- REQ-003 (Ollama client required for candidate generation and judge call)
- REQ-004 (each candidate generation respects VRAM constraints)
