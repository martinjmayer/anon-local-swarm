# Scope — orchestrator

## In scope

- DAG scheduler with `blocked_by` dependency resolution
- Semantic router mapping task types to Ollama models
- Ollama HTTP client with `keep_alive: 0` enforcement
- Variational execution (N candidates + Judge selection)
- Go build validation and self-correction loop
- Context threshold detection and SUMMARY task emission
- MCP client with subprocess management for up to 10 servers
- Output artifact deposit with path traversal protection
- SQLite app state (swarm.db) — projects, tasks, task_candidates
- Config loading from `config.yaml` with env var resolution
- Skills loader reading SKILL.md system prompts
- All observability event emission to obs component

## Out of scope

- Observability storage — owned by obs (`obs.db`)
- Reddit API access — owned by mcp-reddit
- External dispatch (email, Dropbox, webhooks)
- HTTP API exposure to external systems
- Ollama model management (pulling, deleting)
- MCP server protocol implementation (orchestrator is MCP client only)

## Deferred

- Formal job brief schema validation (currently free-text payload)
- Playwright and Docker MCP server registration
- Distributed / multi-host execution
- Persistent context store beyond per-task sidecar files
