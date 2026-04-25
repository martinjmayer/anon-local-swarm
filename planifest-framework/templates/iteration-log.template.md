---
title: "Iteration Log - {{feature-id}}"
summary: "Execution log for the agent session."
status: "active"
version: "0.1.0"
---
# Iteration Log - {{feature-id}}

**Skill:** [docs-agent](../skills/planifest-docs-agent/SKILL.md) (or whichever agent completes the final iteration step)
**Date:** {{ISO-8601}}
**Tool:** {{agentic-tool-name}} (local)
**Model:** {{model-name-and-version, e.g. claude-sonnet-4-20250514, gpt-4.1, gemini-2.5-pro}}
**Phase:** {{phase-number}} (if phased)

---

## Iteration Steps Completed

| Phase | Status | Gate Result | Notes |
|-------|--------|-------------|-------|
| 0 - Assess & Coach | {{pass/skip}} | Design confirmed: {{yes/no}} | {{coaching rounds count}} |
| 1 - Specification | {{pass/fail/skip}} | All artifacts produced: {{yes/no}} | |
| 2 - ADRs | {{pass/fail/skip}} | {{n}} ADRs generated | |
| 3 - Code Generation | {{pass/fail/skip}} | Implementation complete: {{yes/no}} | {{deviations count}} |
| 4 - Validation | {{pass/fail/blocked}} | CI clean: {{yes/no}} | {{self-correct cycles}} cycles |
| 5 - Security | {{pass/fail/skip}} | Critical findings: {{count}} | |
| 6 - Docs & Ship | {{pass/fail/skip}} | All docs synced: {{yes/no}} | |

---

## Requirement Changes During Run

| Change | Phase Active | Classification | Action Taken |
|--------|-------------|----------------|-------------|
| {{description}} | {{phase number}} | cosmetic / additive / contradictory | {{what was re-run}} |

---

## Self-Correct Log

{{what failed and how it was fixed - each attempt with the error and the fix}}

---

## Quirks

{{anything unusual noticed during the run - written to docs/quirks.md and component.yml}}

---

## Recommended Improvements

{{what should be reviewed before the PR - these are not blockers, but flagged for human attention}}

---

## Next Step

```bash
git push origin feature/{{feature-id}}
```

---

*Written by the agent at the end of every Agentic Iteration Loop. This is the audit trail.*

