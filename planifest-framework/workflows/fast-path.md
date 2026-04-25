---
name: fast-path
description: Apply a trivial fix directly without a Feature Brief, Execution Plan, or ADR. Use ONLY when the change meets all Fast Path criteria defined in the orchestrator skill.
---

# Fast Path

Executes a direct, lightweight change for trivial fixes. Bypasses the full pipeline.

## Prerequisites

- An existing feature with code at `src/{component-id}/`
- The request meets **ALL** Fast Path criteria (verified by the orchestrator)

## Fast Path Criteria

The orchestrator must confirm all four before this workflow is used:

1. Does **not** introduce new external dependencies
2. Does **not** alter, add, or remove database schemas or data models
3. Does **not** change security parameters, authentication, or routing logic
4. Is confined to: UI styling, copy changes, or isolated pure-function logic bugs

**If any criterion fails → use `/change-pipeline` instead.**

## Steps

1. **Orchestrator evaluates the request** against the Fast Path criteria above.
   - If all criteria pass: proceed.
   - If any criterion fails: stop. Use `/change-pipeline`.

2. **Implement the fix directly.** No Feature Brief, Execution Plan, or ADR required.
   - Apply the minimum change to address the request.
   - Do not refactor beyond the scope of the fix.

3. **Validate.** Run CI checks (lint, typecheck, test, build). Self-correct up to 3 times. If CI still fails after 3 attempts, escalate to the Change Pipeline.

4. **Update `component.yml`.**
   - Increment the patch version (e.g. `1.2.3` → `1.2.4`)
   - Update `metadata.updatedAt` to today's ISO-8601 date

5. **Log the change.** Append an entry to `plan/changelog/{feature-id}-{YYYY-MM-DD}.md`:
   ```markdown
   ## {YYYY-MM-DD} - Fast Path: {description}
   - Change: {what was changed}
   - Component: {component-id}
   - Reason: {why}
   ```

6. **Commit** using the fast-path convention:
   ```
   fix(fast-path): {description}
   ```
   The pre-push hook and CI workflow recognise this prefix and require only `component.yml` or `plan/changelog/` to be updated - not full `plan/` or `docs/` changes.

## What is NOT touched

- `plan/current/design.md` - no modification
- `plan/current/design-requirements.md` - no modification
- `plan/current/adr/` - no new ADRs
- `docs/` - no modification (unless the fix directly corrects a documentation error)

