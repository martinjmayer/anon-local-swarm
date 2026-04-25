# Component Registry — Autonomous Local Swarm

**Feature:** `anonymous-local-swarm`
**Status:** built
**Last updated:** 2026-04-25

## Components

| ID | Name | Type | Language | Status | Summary | Purpose Doc |
|---|---|---|---|---|---|---|
| `obs` | ALS Observability | microservice | Go | built | Owns obs.db (DuckDB); writes all swarm events as an append-only log, mirrored to stdout via log/slog | [purpose.md](../src/obs/docs/purpose.md) |
| `orchestrator` | ALS Orchestrator | microservice | Go | built | DAG scheduler, semantic router, Ollama client, variational engine, self-correction, MCP client, output deposit | [purpose.md](../src/orchestrator/docs/purpose.md) |
| `mcp-reddit` | Reddit MCP Server | microservice | TypeScript | built | Read-only MCP server wrapping Reddit OAuth API; exposes search, post fetch, and hot feed tools | [purpose.md](../src/mcp-reddit/docs/purpose.md) |

## Data ownership

| Component | Database | Owns |
|---|---|---|
| obs | obs.db (DuckDB) | `events` table |
| orchestrator | swarm.db (SQLite) | `projects`, `tasks`, `task_candidates` tables |
| mcp-reddit | — | No data ownership |

## Component locations

| ID | Source | Manifest |
|---|---|---|
| obs | `src/obs/` | `src/obs/component.yml` |
| orchestrator | `src/orchestrator/` | `src/orchestrator/component.yml` |
| mcp-reddit | `src/mcp-reddit/` | `src/mcp-reddit/component.yml` |
