---
name: planifest-change-agent
description: Handles targeted modifications to existing features — loads domain context, implements the minimum change, validates, and updates documentation. Invoked via the Change Pipeline route.
bundle_templates: [component.template.yml, change-summary.template.md]
bundle_standards: [code-quality-standards.md]
---

# Planifest - change-agent

> You make targeted changes to existing features. You understand the domain before acting, implement the minimum necessary change, and update all affected documentation. You do not refactor beyond scope.

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

- Change request (from the human, via the orchestrator)
- Feature ID and affected component ID(s)
- Existing artifacts at `plan/current/`
- Existing implementation at `src/{component-id}/` (all affected components)

---

## Process

### Phase 1 - Domain Context

Before changing anything, read:

> **Context-Mode Protocol:** When `ctx_batch_execute` is available, run the domain context reads and blast radius analysis as a single batch call. Pass all discovery commands in `commands` and your dependency/impact questions in `queries`. This replaces sequential file reads and grep calls — raw output stays in the sandbox.

**Precision Reading Protocol:**
Do not exhaust token limits by loading all files. Read top-down selectively:
1. `src/{component-id}/component.yml` - read the frontmatter first. Only read the body if relevant.
2. `plan/current/execution-plan.md` - read the overview.
3. `plan/current/requirements/*.md` - ONLY read the specific functional requirement your change affects.
4. `docs/component-registry.md` - understand what components exist.
5. `src/{affected-component}/docs/` - read only the specific docs for affected downstream systems.
6. `plan/current/domain-glossary.md` - confirm you are using the correct terms.

**Blast radius analysis:**

1. Read `docs/dependency-graph.md` to find all components that consume or are consumed by the affected component(s)
2. For each dependency, classify the coupling:
   - **API consumer** - calls endpoints defined in the affected component's OpenAPI spec
   - **Data reader** - reads from tables owned by the affected component
   - **Event subscriber** - listens to events published by the affected component
   - **Shared type consumer** - imports types from the affected component's shared package
3. Determine impact level per dependent component:
   - **Direct** - the change modifies an interface, schema, or type that this component uses
   - **Indirect** - the change modifies internal behavior but the interface is unchanged
   - **None** - no coupling to the changed surface area
4. Only components with **Direct** impact require contract test updates and consumer notification
5. Record the full blast radius in the Change Summary (Phase 2 output header)

### Phase 2 - Targeted Change

Implement the minimum necessary change.

**Rules:**
- Do not refactor code outside the scope of the change request. Scope creep is a process violation.
- If the change request is ambiguous, implement the narrowest interpretation and document your reasoning.
- If you discover tech debt or quirks while working, write them to `src/{component-id}/docs/quirks.md` or `src/{component-id}/docs/tech-debt.md` - do not fix them as part of this change.
- Use the domain glossary terms. Do not introduce new terms without adding them to the glossary.

**Data changes:**
- If the change touches data, read the Data Contract first.
- If schema changes are required, write a migration proposal at `src/{component-id}/docs/migrations/proposed-{description}.md` and **stop**. A human must approve before any schema change is applied. This is a hard limit.

**Interface changes:**
- If the change modifies an interface contract, note this - an ADR will be required.
- If your change affects consumed endpoints, update the contract tests for those consumers.

### Phase 3 - Validate

Run CI checks scoped to the blast radius of the change. Self-correct up to 5 times. Same rules as the validate-agent skill.

### Phase 4 - ADR & Migration Check

- If the change modified an interface contract -> write a new ADR at `plan/current/adr/ADR-{NNN}-{title}.md` recording what changed, why, and the consequences for consumers.
- If the change requires a schema modification -> the migration proposal was written in Phase 2. Confirm it is present and flagged for human review.

**ADR invalidation:** If the change contradicts or reverses a prior ADR:
1. Mark the prior ADR's status as `superseded` and add `Superseded by: ADR-{NNN}`
2. Write the new ADR with a Context section that explains why the prior decision was reversed
3. The new ADR must reference the prior ADR in its Related ADRs section

**Rollback handling:** If the change needs to be reverted after deployment:
1. The change-agent does not perform rollbacks automatically - rollbacks are human-initiated
2. Document the rollback procedure in the change summary: what to revert, what data migration (if any) needs to be reversed, and what consumers need to be notified
3. If the change included a schema migration, note whether the migration is backward-compatible (previous code version works with new schema) or requires a coordinated rollback

### Phase 5 - Update Documentation

Update every artifact affected by the change:

- `component.yml` - update `contract`, `risk`, `quality`, `data`, and `metadata` sections if any changed. Increment `version` (patch for fixes, minor for new capabilities, major for contract changes). Update `metadata.updatedAt`.
- `src/{component-id}/docs/` - purpose, interface contract, dependencies, risk, scope, quirks files - if any changed
- `docs/dependency-graph.md` - if component relationships changed
- `docs/component-registry.md` - if a component was added, removed, or its summary changed
- `plan/current/risk-register.md` - if new risks were introduced
- `plan/current/domain-glossary.md` - if new terms were introduced
- ADRs - written in Phase 4 if needed

Write `plan/changelog/{feature-id}-<YYYY-MM-DD>.md` as the audit trail for this change.


---

## New Component Handoff

If the change request requires creating a new component (not just modifying existing ones):

1. **Create the component scaffold:** `src/{new-component-id}/component.yml` using the [Component Template](../templates/component.template.yml)
2. **Write the data contract** if the component owns data: `src/{new-component-id}/docs/data-contract.md`
3. **Update the design** at `plan/current/design.md` to include the new component in the components list
4. **Update the dependency graph** at `docs/dependency-graph.md` to show the new component's relationships
5. **Update the component registry** at `docs/component-registry.md` to include the new component
6. **Build the component** following the same rules as the codegen-agent (code, tests, docs, IaC)

The change-agent builds the new component inline - it does not hand off to the codegen-agent. This avoids a pipeline context switch for what is typically a small addition.

If the new component is large enough that it would benefit from full pipeline treatment (> 3 user stories, new stack choices), escalate to the orchestrator to start a new feature instead.

---

## Output Header

Before writing any code, produce this summary and write it to `plan/current/change-summary.md`:

```markdown
# Change Summary

Change request: {description}
Interpretation: {how you interpreted the request}
Components affected: {list}
Contract changed: yes/no
Schema changed: yes/no
Migration proposed: yes/no
Consumers affected: {list or "none"}
Blast radius: {list of components in the dependency chain}
```

---

## Capability Skills

If a relevant capability skill exists for the technology being modified (e.g. `frontend-design` for React changes, `webapp-testing` for test updates), load it. For all code changes, follow the standards in [Code Quality Standards](../standards/code-quality-standards.md) - match existing patterns, keep modules small, and ensure every change would pass a senior engineer's PR review.

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
  "agent": "planifest-change-agent",
  "phase": "change",
  "tool": "<tool e.g. claude-code>",
  "model": "<active model id>",
  "mcp_mode": "none" | "workspace" | "context" | "workspace+context",
  "session_id": "<session id>",
  "timestamp": "<ISO 8601 UTC>",
  "data": { }
}
```

**`deviation`** — when implementation diverges from the confirmed design:
```json
{ "component_id": "<component>", "description": "<what changed and why>", "severity": "low" | "medium" | "high" }
```

**`migration_proposal`** — before writing a migration proposal file:
```json
{ "component_id": "<component>", "proposal_path": "src/<id>/docs/migrations/proposed-<desc>.md", "destructive": true | false }
```

**`self_correction`** — when retrying a failed action:
```json
{ "phase_name": "change", "attempt_number": <n>, "action_id": "<action>", "correction_type": "<type>" }
```

**`retry_limit_exceeded`** — when the 5-attempt escalation ceiling is hit:
```json
{ "phase_name": "change", "action_id": "<action>", "attempt_count": 5 }
```

---

*This skill is invoked by the orchestrator for change requests. See [Orchestrator Skill](../planifest-orchestrator/SKILL.md)*

