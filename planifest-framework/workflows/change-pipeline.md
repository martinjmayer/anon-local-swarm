---
name: change-pipeline
description: Modify an existing feature - implements the minimum change, validates, and updates documentation. Use this instead of the full pipeline for changes to existing work.
---

# Change Pipeline

Execute the confirmed design change pipeline for modifications to an existing feature.

## Prerequisites

- An existing feature with a confirmed design at `plan/current/design.md`
- The component(s) affected by the change exist in `src/{component-id}/`

## Input

Provide all three:
- **Feature ID**: which feature to modify
- **Component ID**: which component(s) are affected
- **Change request**: what to change and why

## Steps

1. **Load the orchestrator skill**
2. **Confirm scope** - the orchestrator confirms:
   - Which feature?
   - Which component(s)?
   - What is the change?
   - If ambiguous, clarify - one question at a time
3. **Invoke the change-agent** - it handles:
   - Load domain context (existing spec, manifests, ADRs)
   - Implement the minimum necessary change
   - Run validation (lint, typecheck, test, build)
   - Check for contract or schema changes - if found, propose migration and stop
   - Update documentation
4. **Review** - the change-agent produces a summary of what changed and why via a change log entry at `plan/changelog/`.
5. **Phase 7 - Human Review and Filing** (Post-Review Action)
   - The human reviews the changes and the active plan.
   - Upon acceptance, the active plan (brief, spec, ADRs) is moved from `plan/current/` to `plan/_archive/{feature-id}/` for historical tracking.

