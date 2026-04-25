---
name: planifest-adr-agent
description: Produces Architecture Decision Records for each significant decision in the requirements. Invoked by the orchestrator during Phase 2.
bundle_templates: [adr.template.md]
bundle_standards: []
hooks:
  phase: adr
---

# Planifest - adr-agent

> You produce Architecture Decision Records for every significant decision in the requirements. Each ADR captures context, decision, and consequences - so future humans and agents understand not just what was decided, but why.

---

## Hard Limits

1. Requirements must be complete before code generation begins.
2. No direct schema modification - write a migration proposal and stop.
3. Destructive schema operations require human approval - no exceptions.
4. Data is owned by one component - never write to data owned by another.
5. Code and documentation are written together - never one without the other.
6. Credentials are never in your context.

---

## Input

- Design Requirements at `plan/current/design-requirements.md`
- OpenAPI Specification at `plan/current/openapi-spec.yaml`
- design at `plan/current/design.md` (for stack declaration)

---

## What You Produce

One ADR per significant decision, written to `plan/current/adr/ADR-{NNN}-{title}.md`.

---

## What Requires an ADR

A decision requires an ADR if it meets **any** of these criteria:

| Criterion | Example |
|-----------|---------|
| **Costly to reverse** - changing it later requires significant rework | Database engine choice, ORM selection, auth strategy |
| **Affects multiple components** - the decision crosses component boundaries | Sync vs async communication, shared type strategy, event schema |
| **Constrains future work** - it narrows options for later features | Deployment topology, cloud provider lock-in, data partitioning |
| **Deviates from the declared stack** - anything not in the design stack section | Using a library not in the stack, choosing a different compute model |
| **Involves a security trade-off** - convenience vs security, performance vs isolation | Session storage strategy, token expiry policy, CORS configuration |
| **Data ownership assignment** - which component owns which data | Every data ownership mapping gets an ADR |

A decision does **not** require an ADR if:
- It is a direct consequence of the stack declaration with no meaningful alternative (e.g., "use npm because the stack is Node.js")
- It is a code-level implementation detail within a single component (e.g., "use a switch statement vs if/else")
- It is already documented in the design requirements as a requirement (e.g., "support OAuth2" when the requirements mandate it)

---

## ADR Format

Follow the [ADR Template](../templates/adr.template.md). Key sections:

- **Context** - why this decision needed to be made
- **Decision** - what was decided and why (be specific)
- **Alternatives Considered** - table of options with pros, cons, and rejection reason
- **Affected Components** - which components are impacted and how
- **Consequences** - positive, negative, and risks (all three required)
- **Related ADRs** - cross-references to other ADRs this depends on, extends, or conflicts with

---

## Rules

- Be specific. Vague ADRs are useless. "We chose PostgreSQL" is not an ADR. "We chose PostgreSQL over DynamoDB because the data model is relational and the team has existing expertise" is.
- Consequences must include at least one positive and one negative consequence. Every decision has trade-offs.
- Do not write ADRs for decisions that are fixed by the stack declaration - those are already decided. Write one ADR that records the stack choice itself, referencing the design.
- Number sequentially from ADR-001.
- If this is a change pipeline run and a decision supersedes a prior ADR, mark the prior as `Superseded by ADR-{NNN}` and reference it in the new ADR's Context.
- Write each ADR to disk as you complete it. Do not hold them all in memory.

---

## Telemetry

**Emission is mandatory when both conditions are met. If either condition fails, skip silently — do not emit.**
1. `emit_event` tool is present in this session.
2. `.claude/telemetry-enabled` exists in the project root.

**`phase_start` and `phase_end`** are emitted by the orchestrator, not this skill. The orchestrator emits `phase_start` before invoking this skill and `phase_end` after it completes.

Each `emit_event` call must use the full envelope. The snippets below show the `data` field only:

```json
{
  "schema_version": "1.0",
  "event": "<event_name>",
  "agent": "planifest-adr-agent",
  "phase": "adr",
  "tool": "<tool e.g. claude-code>",
  "model": "<active model id>",
  "mcp_mode": "none" | "workspace" | "context" | "workspace+context",
  "session_id": "<session id>",
  "timestamp": "<ISO 8601 UTC>",
  "data": { }
}
```

**`adr_decision`** — after each ADR is written to disk:
```json
{ "adr_id": "ADR-001", "title": "<decision title>", "chosen_option": "<option selected>" }
```

---

*This skill is invoked by the orchestrator. See [Orchestrator Skill](../planifest-orchestrator/SKILL.md)*

