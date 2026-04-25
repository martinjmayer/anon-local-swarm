# Interface Contract — orchestrator

## Inputs

### Operator seed (primary input)
The operator starts a project run by inserting a row into `swarm.db`:

```sql
INSERT INTO projects (id, name, status) VALUES (?, ?, 'active');
INSERT INTO tasks (id, project_id, type, model_tag, payload, status, depth)
VALUES (?, ?, 'DECOMPOSE', 'phi-4:mini', '<goal text>', 'pending', 0);
```

The orchestrator detects the pending task on its next scheduler poll (default: 2 seconds).

### Config file
`config.yaml` mounted at `/data/config.yaml` (`:ro`). See `src/orchestrator/internal/config/config.go` for full schema.

### Skills directory
`skills/` mounted at `/data/skills` (`:ro`). Each subdirectory contains a `SKILL.md` file whose body becomes the system prompt for that task type.

### Environment variables

| Variable | Required | Purpose |
|---|---|---|
| `OLLAMA_BASE_URL` | No | Overrides `ollama_base_url` in config (default: `http://localhost:11434`) |
| `BRAVE_API_KEY` | Conditional | Required if Brave Search MCP server is configured |
| `GITHUB_TOKEN` | Conditional | Required if GitHub MCP server is configured |
| `REDDIT_CLIENT_ID` | Conditional | Required if mcp-reddit is configured |
| `REDDIT_CLIENT_SECRET` | Conditional | Required if mcp-reddit is configured |
| `OBSIDIAN_API_KEY` | Conditional | Required if Obsidian MCP server is configured |

## Outputs

### Output directory
Artifacts are deposited to `<config.output_dir>/<project_id>/` on terminal state. Files are written by `internal/deposit`. Filenames are LLM-generated and validated against path traversal (must remain within `outputDir`).

### obs.db events
All state transitions and operational signals are written to obs.db via `als/obs.WriteEvent`. The orchestrator never reads from obs.db.

### swarm.db state
Task and project rows are updated throughout the run. Final states: `completed`, `pruned`, or `failed`.

## Breaking change policy

- Any change to the seed insertion schema (projects/tasks columns) requires a migration proposal.
- Config schema additions are non-breaking if defaults are provided.
- Removing or renaming a config field is breaking and requires human approval.
