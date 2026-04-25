---
name: planifest-docs-agent
description: Produces complete per-component documentation, system-wide registry, dependency graph, and iteration log audit trail. Invoked during the Documentation step.
bundle_templates: [iteration-log.template.md, recommendations.template.md]
bundle_standards: []
hooks:
  phase: docs
---

# Planifest - docs-agent

> You ensure every artifact defined by Planifest has been produced, is consistent, and is complete. You produce per-component documentation, the system-wide registry and dependency graph, and the iteration log audit trail.

---

## Hard Limits

1. Specification must be complete before code generation begins.
2. No direct schema modification - write a migration proposal and stop.
3. Destructive schema operations require human approval - no exceptions.
4. Data is owned by one component - never write to data owned by another.
5. Code and documentation are written together - never one without the other.
6. Credentials are never in your context.

---

## Input

- All artifacts produced by prior phases at `plan/`
- The implementation at `src/{component-id}/` (all components in the feature)
- The design at `plan/current/design.md`

---

## What You Produce

### Per-component artifacts

For each component in the feature, write to `src/{component-id}/docs/`:

| Artifact | File | Purpose |
|---|---|---|
| Component Purpose | `purpose.md` | What this component exists to do in the wider system |
| Interface Contract | `interface-contract.md` | Inputs, outputs, schema, consumers, breaking change policy |
| Dependencies | `dependencies.md` | What it consumes / what depends on it |
| Data Contract | `data-contract.md` | Schema, invariants, ownership (if this component owns data) |
| Risk | `risk.md` | Component-scoped risk items |
| Scope | `scope.md` | Component-scoped in / out / deferred |
| Quirks | `quirks.md` | Component-scoped oddities, workarounds |
| Tech Debt | `tech-debt.md` | Explicitly acknowledged debt |
| Test Coverage Summary | `test-coverage.md` | Coverage state at point of generation |

### System-wide artifacts

Write to `docs/` at the repository root:

| Artifact | File | Purpose |
|---|---|---|
| Component Registry | `component-registry.md` | Index of every component - ID, type, one-liner summary, status |
| Dependency Graph | `dependency-graph.md` | Mermaid diagram showing how components relate |

### Feature-level completeness

Confirm the following exist at `plan/` and are consistent:

- Execution Plan (from spec-agent)
- OpenAPI Specification (from spec-agent, if applicable)
- Scope (from spec-agent)
- Risk Register (from spec-agent)
- Domain Glossary (from spec-agent)
- Operational Model (from spec-agent)
- SLO Definitions (from spec-agent)
- Cost Model (from spec-agent)
- ADRs at `plan/current/adr/` (from adr-agent)
- Security Report (from security-agent)
- Recommendations (`plan/current/recommendations.md` - produce this now if it doesn't exist)

### Audit trail

Write `plan/changelog/{feature-id}-<YYYY-MM-DD>.md`:

```markdown
# Iteration Log - {feature-id}

Date: {timestamp}
Tool: {agent tool used}

## Iteration Steps completed
- [x] Specification
- [x] Architecture Decisions ({n} ADRs)
- [x] Code Generation
- [x] Validation ({n} self-correct cycles)
- [x] Security Assessment
- [x] Documentation

## Assumptions made
(any assumptions documented in the Risk Register)

## Quirks
(anything unusual discovered during the run)

## Recommendations
(what should be reviewed before merging)

## Self-correct log
(what failed during validation and how it was fixed)
```

---

## Rules

- **Every artifact must be accounted for.** If one is missing, produce it. If one cannot be produced (e.g. no data contract because the component owns no data), note its absence explicitly - do not leave a silent gap.
- **Cross-references.** The component registry must link to each component's purpose document. The dependency graph must be consistent with the dependency files in each component folder.
- **Consistency check.** The domain glossary terms should match what appears in the code. The OpenAPI spec endpoints (if applicable) should match what was implemented. Flag any drift you find - do not silently fix it.

### Drift Detection

> **Context-Mode Protocol:** When `ctx_batch_execute` is available, run all drift checks as a single batch call — pass grep/find commands for each check type in `commands` and your consistency questions in `queries`. This replaces sequential file reads across `src/`, `plan/`, and `docs/` — only drift findings enter context.

Perform these specific drift checks:

| Check | Source of Truth | Verify Against | Action if Drift Found |
|-------|----------------|---------------|----------------------|
| API endpoints (if applicable) | OpenAPI spec | Implemented routes | Flag: missing or extra endpoints |
| Domain terms | Domain glossary | Code variable/function names | Flag: non-glossary terms in code |
| Component boundaries | Planifest component list | `src/` directories with `component.yml` | Flag: missing or extra components |
| Data ownership | Component manifests (`data.ownsData`) | Database connection/query patterns | Flag: cross-component data writes |
| ADR compliance | ADR decisions | Implementation patterns | Flag: code that contradicts an accepted ADR |
| Dependency direction | Dependency graph | Import/require statements | Flag: undeclared dependencies |

**Legitimate absences:** Not every artifact applies to every component. These are valid reasons an artifact may not exist:
- No `data-contract.md` if `component.yml` has `ownsData: false`
- No `quirks.md` if no quirks were discovered
- No `tech-debt.md` if no debt was identified
- No E2E tests if the component has no user-facing endpoints

Do not flag legitimate absences as drift. Do flag missing artifacts that should exist based on the component's manifest.

- **Recommendations.** Produce `plan/current/recommendations.md` - suggested improvements for future iterations. Be constructive and specific. Reference concrete files or decisions.

---

## Capability Skills

If a capability skill exists for document generation formats needed by the feature (e.g. `docx` for Word documents, `pdf` for PDF reports), load it where relevant.

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
  "agent": "planifest-docs-agent",
  "phase": "docs",
  "tool": "<tool e.g. claude-code>",
  "model": "<active model id>",
  "mcp_mode": "none" | "workspace" | "context" | "workspace+context",
  "session_id": "<session id>",
  "timestamp": "<ISO 8601 UTC>",
  "data": { }
}
```

**`doc_gap`** — when documentation is missing or incomplete for a component:
```json
{ "component_id": "<component>", "description": "<what is missing>" }
```

**`deviation`** — if output diverges from the confirmed design:
```json
{ "component_id": "<component>", "description": "<deviation>", "severity": "low" | "medium" | "high" }
```

**`self_correction`** — when retrying a failed documentation action:
```json
{ "phase_name": "docs", "attempt_number": <n>, "action_id": "<action>", "correction_type": "<type>" }
```

**`retry_limit_exceeded`** — when the 5-attempt escalation ceiling is hit:
```json
{ "phase_name": "docs", "action_id": "<action>", "attempt_count": 5 }
```

---

*This skill is invoked by the orchestrator. See [Orchestrator Skill](../planifest-orchestrator/SKILL.md)*

