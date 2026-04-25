# Scope - Guide

> How to write the scope document that keeps agents within boundaries.

*Related: [Spec Agent Skill](../skills/spec-agent-SKILL.md) | [design requirements Guide](design-spec-guide.md)*

---

## Purpose

The Scope document is the boundary fence. It tells the agent what to build, what NOT to build, and what's been deliberately pushed to later. Without it, agents drift - they add features not asked for, miss features that were, or build deferred items prematurely.

All three sections must be present. If nothing is deferred, state "Nothing deferred."

---

## Who Writes It

The **spec-agent** produces this during Phase 1, derived from the Feature Brief's scope boundaries. The human seeds the scope in the brief; the spec-agent formalises it.

---

## The Three Sections

### In Scope

What this feature or phase **will** deliver. Be specific - "authentication" is vague; "JWT-based auth with refresh tokens, scoped to the auth-service component" is clear.

Each item should be traceable to a feature in the Feature Brief.

### Out of Scope

What this feature **will NOT** deliver. This is equally important as In Scope - arguably more so, because it prevents the agent from building things you didn't ask for.

Good "Out of Scope" items are things someone might reasonably expect to be included but aren't:
- "Email notifications - notifications are a separate feature"
- "Admin dashboard - API-only in this phase"
- "Multi-tenancy - single-tenant deployment only"

### Deferred

What might be delivered later but not now. The critical addition: note what is **blocked** until each deferred item is resolved.

- "Database migration tooling - deferred to Phase 2. Blocked: no automated rollback until this ships."
- "Rate limiting - deferred. Blocked: the API is unprotected against abuse until this ships."

This turns deferred items into tracked risks.

---

## Common Mistakes

1. **Missing "Out of Scope".** If everything is in scope, the agent has no boundaries. Always state what's excluded.
2. **Deferred without blockers.** "Deferred" without noting what's blocked is wishful thinking, not project management.
3. **Scope creep via omission.** If the brief says "MVP" but the scope doesn't say what's excluded from the MVP, the agent may build the full product.

---

## File Location

`plan/current/scope.md`

If phased: `plan/scope-phase-{n}.md`

---

*Template: [scope.template.md](scope.template.md)*

