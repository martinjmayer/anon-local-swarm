---
description: Planifest framework - mandatory rules for all code generation
globs: ["**/*"]
alwaysApply: true
---

# Planifest - Mandatory Framework Rules

This project uses the confirmed design framework. These rules are non-negotiable and apply to every session.

## Hard Limits

1. **No code without a confirmed design.** You MUST NOT generate application code unless a confirmed design exists at plan/current/design.md. If none exists, load the planifest-orchestrator skill and begin Phase 0 (Assess and Coach). Do NOT skip to code generation.
2. **No code without documentation.** Every component MUST have a component.yml manifest and docs/ artifacts. Never produce code without its corresponding documentation.
3. **No direct schema modification.** Write a migration proposal at src/{component-id}/docs/migrations/proposed-{desc}.md and STOP for human approval.
4. **Destructive schema operations require human approval.** Drop column, drop table, rename - propose and stop. No exceptions.
5. **Data is owned by one component.** Never write to data owned by another component.
6. **No credentials in context.** If a credential appears anywhere, flag it immediately and do not use it.
7. **Update documentation after any deviation.** If implementation required deviating from the spec, plan, or design decisions, you MUST update the affected artifacts (plan/, docs/, or component.yml) to reflect what was actually built. Documentation must always match reality.

## Operational Directives

- **Be succinct:** Minimal prose. Format answers in markdown with brevity.
- **Do not invent APIs:** Only call endpoints that explicitly exist in the provided OpenAPI interfaces.
- **Check `.cursorindexingignore`:** The `standards/reference/` directory and guide files are deliberately excluded from your semantic search index to preserve your context window. If you need deep domain knowledge about frameworks or pitfalls, explicitly read those files using `@` mentions.

## Escalation

If you are blocked, unable to resolve tests after 5 attempts, or confused by conflicting requirements, **STOP**. Do not guess. State clearly what is blocking you and ask the human for direction.
