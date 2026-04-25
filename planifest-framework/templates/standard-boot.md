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
- **Commit message standard:** Follow `planifest-framework/standards/commit-standards.md`. Concise, scope-focused, imperative mood, ≤72 chars. No AI attribution. No affirmatory language.
- **Do not invent APIs:** Only call endpoints that explicitly exist in the provided OpenAPI interfaces.
- **Check `.cursorindexingignore`:** The `standards/reference/` directory and guide files are deliberately excluded from your semantic search index to preserve your context window. If you need deep domain knowledge about frameworks or pitfalls, explicitly read those files using `@` mentions.
- **Use context-mode MCP when available:** If `mcp__context-mode__ctx_batch_execute` is available, apply the routing rules in `AGENTS.md`. Key rules: use `ctx_batch_execute` for codebase discovery; `ctx_execute_file` for analysis-only file reads; `ctx_execute(language:"shell")` for large-output shell commands; `ctx_fetch_and_index` + `ctx_search` for web fetching. Use the `Read` tool only when you need file content in context for editing.

## Hook Enforcement

Planifest installs deterministic enforcement hooks via `setup.sh`. These run automatically:

- **gate-write** (PreToolUse: Write, Edit): Blocks writes to `src/` unless `plan/current/design.md` exists AND the target path matches a declared component. Writes to `plan/`, `docs/`, and other always-permitted paths are never blocked.
- **check-design** (UserPromptSubmit): Injects active component scope from `design.md` as additional context.
- **emit-phase-start / emit-phase-end**: Structured telemetry for pipeline phases (no-op if `PLANIFEST_TELEMETRY_URL` is unset).

Enforcement failures exit 2 and surface a human-readable message. All unexpected errors exit 0 — hooks never block your session unexpectedly.

Hook enforcement is active: `gate-write.mjs` blocks writes outside confirmed design scope; `check-design.mjs` injects scope context on every turn. Manual scope checks in these instructions are retained as documentation but are now redundant enforcement.

## Escalation

If you are blocked, unable to resolve tests after 5 attempts, or confused by conflicting requirements, **STOP**. Do not guess. State clearly what is blocking you and ask the human for direction.
