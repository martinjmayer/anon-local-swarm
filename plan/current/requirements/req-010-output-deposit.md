---
title: "Requirement: REQ-010 - Output Artifact Deposit"
summary: "Terminal state detection and artifact deposit to project directory"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-010 - Output Artifact Deposit

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — output deposit to project directory, job brief delivery spec
**Priority:** must-have

---

## Functional Requirements
- When a project reaches terminal state (`completed` or `pruned`), the orchestrator MUST deposit all output artifacts into the project directory
- The output directory path MUST be configurable per project (stored in the job brief / project record)
- Output artifacts MUST include: final task outputs, research summaries, generated Go code (if any), and the viability review report
- Each artifact MUST be written as a discrete file (e.g., `research-summary.md`, `viability-report.md`, `main.go`)
- If the project is fully pruned (all branches pruned), the orchestrator MUST write a `prune-report.md` explaining which branches were pruned and why
- The orchestrator MUST NOT write to external systems (email, Dropbox) — deposit is to local directory only
- Artifact filenames MUST be deterministic and human-readable

## Acceptance Criteria
- [ ] Completed project deposits research-summary.md and viability-report.md to configured output directory
- [ ] Fully pruned project deposits prune-report.md with branch scores
- [ ] Generated Go code deposited as main.go (or named per task output)
- [ ] Output directory is writable before deposit (checked at project start, not at terminal state)
- [ ] No writes to email or Dropbox under any condition

## Dependencies
- REQ-006 (terminal state determined by viability and completion logic)
