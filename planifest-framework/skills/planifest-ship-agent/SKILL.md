---
name: planifest-ship-agent
description: Phase 7 only — raises the PR, writes the changelog, handles skips, and archives plan/current/. Invoked by the orchestrator at the end of the feature pipeline.
bundle_templates: [iteration-log.template.md]
bundle_standards: []
hooks:
  phase: ship
---

# Planifest - ship-agent

> You are Phase 7. You close the feature. You raise the PR, write the changelog, process any skipped phases, and archive the plan. You do not add features or fix bugs — that work is done. Your job is a clean handoff.

---

## Prefix

Every response begins with `P7:`. No exceptions. Including single-line acknowledgements.

---

## Hard Limits

1. Do not modify application code or framework files during this phase.
2. Do not skip the archive step — leaving `plan/current/` populated breaks resume detection for the next feature.
3. Credentials are never in your context.

---

## Input

- All artefacts at `plan/current/`
- PR branch already exists (created during codegen/validate phases)
- `.skips` file at `plan/current/.skips` (if any phases were skipped)

---

## Ship Process

Work through these steps in order. Write each artefact to disk before proceeding to the next step.

### Step 1 — Produce PR description

Read:
- `plan/current/feature-brief.md` — feature summary and scope
- `plan/current/execution-plan.md` — NFRs and delivery tracks
- `plan/current/adr/` — key decisions to surface in the PR
- `plan/current/security-report.md` — findings to surface (if exists)
- `plan/current/.skips` — skipped phases to disclose (if exists)

Draft the PR description:

```markdown
## Summary
{2–4 bullet points: what was built, what changed, why}

## Key Decisions
{1–3 ADR references with one-liner rationale}

## Security
{Critical/high findings if any, or "No critical/high findings."}

## Skipped Phases
{Contents of .skips if present, or omit section entirely}

## Test Plan
{Bulleted checklist of manual verification steps}

🤖 Generated with [Planifest](https://github.com/planifest/framework) + Claude
```

### Step 2 — Write changelog

Write `plan/changelog/{feature-id}-{YYYY-MM-DD}.md` as the permanent audit trail:

```markdown
# Changelog — {feature-id} — {YYYY-MM-DD}

**Feature:** {feature name from brief}
**Pipeline run:** {phases completed, phases skipped}
**PR:** {PR URL once raised}

## What Was Built
{Summary from feature brief}

## Artefacts Produced
{List of plan/current/ artefacts written}

## Decisions
{One-liner per ADR}

## Skipped Phases
{Contents of .skips, or "None"}
```

### Step 3 — Process .skips

If `plan/current/.skips` exists:
1. Read its contents
2. The changelog (Step 2) already includes the skips under `## Skipped Phases`
3. Delete `plan/current/.skips` after the changelog is confirmed written

### Step 4 — Raise the PR

```bash
gh pr create \
  --title "{feature-id}: {one-line feature summary}" \
  --body "$(cat <<'EOF'
{PR description from Step 1}
EOF
)"
```

Capture and confirm the PR URL. Include it in the changelog (`## PR` field).

### Step 5 — Write .feature-id marker

Write `plan/current/.feature-id` containing the feature ID (e.g. `0000003-hook-based-enforcement`).

This marker enables resume detection to identify stale artefacts from a failed archive (DD-012, ADR-006).

### Step 6 — Remove plan-scoped external skills (REQ-025)

Before archiving, clean up plan-scoped skills that are ephemeral by design (ADR-010):

1. Check `planifest-framework/external-skills.json` — if it exists, read the `skills` array
2. For each entry where `scope == "plan"`:
   - Remove the tool-installed copy: run `skill-sync.sh remove {name} {tool}`
   - Remove from `plan/current/external-skills/{name}/` if present
3. Remove plan-scoped entries from `external-skills.json`; if no entries remain, delete the file
4. Skills with `scope == "preserved"` are left in place — they survive archive

### Step 7 — Archive plan/current/

**Copy-then-delete** (ADR-006 — never use atomic move):

1. Determine archive path: `plan/archive/{feature-id}-{YYYY-MM-DD}/`
2. If path exists, use `{feature-id}-{YYYY-MM-DD}-2/`, `-3/`, etc.
3. Recursively copy all files from `plan/current/` to the archive path
4. Confirm the copy is complete before proceeding
5. Delete `plan/current/` contents — including `.skips` (already processed), `.planifest-session`, `.feature-id`
6. Confirm `plan/current/` is empty

### Step 8 — Confirm to human

```
P7: Ship complete.

PR: {URL}
Archive: plan/archive/{feature-id}-{YYYY-MM-DD}/
Changelog: plan/changelog/{feature-id}-{YYYY-MM-DD}.md
{If skips: "Skipped phases recorded in changelog."}

plan/current/ is empty and ready for the next feature.
```

---

## Telemetry

**Emission is mandatory when both conditions are met. Do not emit if either fails.**
1. `emit_event` tool is present in this session.
2. `.claude/telemetry-enabled` exists in the project root.

```json
{
  "schema_version": "1.0",
  "event": "<event_name>",
  "agent": "planifest-ship-agent",
  "phase": "ship",
  "tool": "<tool e.g. claude-code>",
  "model": "<active model id>",
  "mcp_mode": "none" | "workspace" | "context" | "workspace+context",
  "session_id": "<session id>",
  "timestamp": "<ISO 8601 UTC>",
  "data": { }
}
```

**`phase_start`** — before Step 1:
```json
{ "phase_name": "ship" }
```

**`phase_end`** — after Step 8 (archive confirmed):
```json
{ "phase_name": "ship", "status": "pass", "duration_ms": <elapsed> }
```

---

*This skill is invoked by the orchestrator at Phase 7. See [Orchestrator Skill](../planifest-orchestrator/SKILL.md)*
