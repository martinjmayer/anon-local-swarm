# Iteration Log - Guide

> The audit trail written at the end of every Agentic Iteration Loop.

*Related: [Docs Agent Skill](../skills/docs-agent-SKILL.md) | [Orchestrator Skill](../skills/orchestrator-SKILL.md)*

---

## Purpose

The Iteration Log is the audit trail. It records what happened during the Agentic Iteration Loop: which steps completed, what self-corrections were needed, what quirks were discovered, and what the agent recommends the human review. It's the first thing a human reads before approving a PR.

---

## Who Writes It

The **docs-agent** writes it during the final iteration step (Documentation & Ship). If the loop is interrupted before that step, the last active agent writes it.

---

## What It Captures

### Iteration Steps Completed

A checklist of all iteration steps. Unchecked items tell the reviewer what didn't run - and therefore what hasn't been validated.

### Self-Correct Log

Every failure and recovery. This is not embarrassing - it's valuable. Knowing that "the validate-agent caught a missing index and added it on attempt 3" tells the reviewer exactly what was auto-fixed.

Include:
- The error message
- What the agent tried
- Whether it succeeded
- How many attempts it took

### Quirks

Anything unusual the agent noticed during the run. These are also written to `quirks.md` and the component's `component.yml`. Examples:
- "The ORM generates a LEFT JOIN where an INNER JOIN would be correct - worked around with a raw query"
- "The test framework doesn't support parallel execution with this database driver"

### Recommended Improvements

Not blockers, but things the human should consider. These are flagged for attention at the PR gate:
- "The auth token TTL is hardcoded - consider making it configurable"
- "Test coverage for the error paths is thin - 3 out of 7 error cases have tests"

---

## Reading a Pipeline Run

When reviewing a PR that was produced by Planifest:

1. **Check the iteration steps** - are all boxes checked?
2. **Read the self-correct log** - were there many retries? Were they for the same issue? This signals a spec gap, not an agent failure.
3. **Read the quirks** - these are tech debt items. Decide whether to address them now or later.
4. **Read the recommendations** - these are the agent's judgment calls flagged for human review.

---

## File Location

`plan/changelog/{feature-id}-<YYYY-MM-DD>.md`

If phased: `plan/iteration-log-phase-{n}.md`

---

*Template: [iteration-log.template.md](iteration-log.template.md)*

