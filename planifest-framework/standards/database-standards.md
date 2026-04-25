# Planifest Database Standards

> The data layer is the hardest part to change. Schema decisions are expensive to reverse, migrations are risky, and data loss is unrecoverable. These standards ensure agent-generated database code is safe and well-structured.

---

## 1. Schema Design

- **Naming:** snake_case for tables and columns. Plural table names (`users`, `orders`).
- **Primary keys:** Every table has a primary key. Prefer UUIDs for distributed systems, auto-increment for single-instance.
- **Foreign keys:** Always define foreign key constraints. Never rely on application-level referential integrity alone.
- **Timestamps:** Every table includes `created_at` and `updated_at` columns (timestamptz in PostgreSQL).
- **Soft deletes:** Use a `deleted_at` column instead of hard deletes, unless the data contract explicitly requires hard deletes.
- **Indexing:** Create indexes for every foreign key and every column used in WHERE clauses. Document index rationale.

---

## 2. Migrations

- All schema changes go through migration files - never modify the database directly
- Migrations are forward-only (up only). Rollbacks are separate migration files.
- Migration files are numbered sequentially and never reordered
- Every migration is idempotent - running it twice produces the same result

**Destructive operations** (drop column, drop table, rename) require:
1. A migration proposal at `src/{component-id}/docs/migrations/proposed-{desc}.md`
2. Human approval
3. A data backup plan

---

## 3. Data Contracts

- Every data-owning component has a data contract at `src/{component-id}/docs/data-contract.md`
- The data contract defines: tables, columns, types, constraints, relationships, invariants
- The data contract is the source of truth - the ORM schema must match it exactly
- Changes to the data contract require a migration proposal

---

## 4. Query Patterns

- Use parameterized queries - never string interpolation for SQL
- Use the ORM declared in the stack for standard CRUD operations
- Raw SQL is acceptable for complex queries - document the rationale
- N+1 queries are defects - use eager loading or batch queries
- All queries must have a timeout configured

---

## 5. Connection Management

- Use connection pooling for all database connections
- Configure pool size based on the compute model (serverless needs smaller pools)
- Handle connection errors gracefully - retry with backoff, then fail with a clear error
- Close connections properly - use `finally` blocks or connection managers

---

*Referenced by codegen-agent and spec-agent. Source of truth: `planifest-framework/standards/database-standards.md`*
