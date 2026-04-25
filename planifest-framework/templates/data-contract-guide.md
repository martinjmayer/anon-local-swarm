# Data Contract - Guide

> How the codegen-agent defines schema ownership and how migration proposals work.

*Related: [Codegen Agent Skill](../skills/codegen-agent-SKILL.md) | [Component Guide](component-guide.md)*

---

## Purpose

The Data Contract is the authoritative schema definition for a component's data. It defines which tables exist, what columns they have, what invariants must hold, and who owns the data. It is the single source of truth - the database schema is derived from this document, not the other way around.

---

## Who Writes It

The **codegen-agent** creates it during Phase 3 when building data-owning components. It's updated exclusively through **migration proposals** - never by directly editing the schema.

---

## The Ownership Rule

Data is owned by exactly one component. This is a Planifest hard limit.

- The owning component reads and writes its data
- Other components access it via the owner's API (REST, gRPC, events)
- No component bypasses the owner to query another component's database directly

The `Owner` field in the data contract header must match the `id` in the component's `component.yml`.

---

## Section-by-Section Guidance

### Tables

Each table gets a full column listing with types, nullability, defaults, and constraints. Be precise - the migration tooling reads this.

**Indexes** - list every index. Don't rely on the ORM to "figure it out". Explicit indexes are reproducible.

**Relationships** - type (one-to-many, many-to-many), target table, and foreign key. This feeds into the OpenAPI schema and the component's integration points.

### Invariants

Rules the data must always satisfy, regardless of how it's accessed. These are the data's own rules, not the application's business logic.

| ❌ Application Logic | ✅ Data Invariant |
|---------------------|-------------------|
| "Orders must be approved before shipping" | "order.total_amount must equal the sum of order_items.amount for all items in the order" |
| "Users can't delete their own account" | "user.email must be unique across all non-deleted rows" |

Invariants become CHECK constraints, triggers, or application-layer validations depending on the stack.

### Migration History

Every schema change is recorded here. The migration proposal process:

1. An agent needs to change the schema
2. It writes a migration proposal to `src/{component-id}/docs/migrations/proposed-{description}.md`
3. It **stops** and reports the proposal to the orchestrator
4. The human reviews and approves (or rejects)
5. If approved, the agent applies the migration and updates the data contract

**Destructive operations** (drop column, drop table, rename) additionally require explicit human approval - this is a hard limit.

---

## Common Mistakes

1. **Direct schema edits.** Never modify the data contract directly during codegen. Always go through a migration proposal.
2. **Missing invariants.** If the data has rules (uniqueness, referential integrity, value ranges), they must be here - not just in application code.
3. **Shared ownership.** If two components "need" to own the same data, the architecture is wrong. One owns it; the other consumes it via API.
4. **No migration history.** Every change, no matter how small, gets a migration entry. This is the audit trail for schema evolution.

---

## File Location

`src/{component-id}/docs/data-contract.md`

Migration proposals: `src/{component-id}/docs/migrations/proposed-{description}.md`

---

*Template: [data-contract.template.md](data-contract.template.md)*
