# Dependency Graph — Autonomous Local Swarm

## Component relationships

```mermaid
graph TD
    Operator([Operator]) -->|inserts seed task| swarmdb[(swarm.db\nSQLite)]
    Operator -->|queries| obsdb[(obs.db\nDuckDB)]
    Operator -->|reads| output[/output/\nFilesystem]

    subgraph Docker Container
        orchestrator[orchestrator\nGo binary]
        mcp-reddit[mcp-reddit\nNode.js subprocess]
        obs[obs\nGo package]
    end

    orchestrator -->|reads/writes| swarmdb
    orchestrator -->|WriteEvent| obs
    obs -->|writes| obsdb
    obs -->|mirrors| stdout[stdout\nlog/slog]
    orchestrator -->|spawns + MCP stdio| mcp-reddit
    orchestrator -->|HTTP POST /api/generate| Ollama([Ollama\nhost:11434])
    mcp-reddit -->|HTTPS OAuth2| Reddit([Reddit API])

    orchestrator -->|writes artifacts| output
```

## Dependency matrix

| From | To | Type | Protocol |
|---|---|---|---|
| orchestrator | obs | Go import | In-process function call |
| orchestrator | mcp-reddit | Subprocess | MCP stdio (JSON-RPC) |
| orchestrator | swarm.db | File | SQLite (CGO) |
| orchestrator | Ollama | Network | HTTP |
| obs | obs.db | File | DuckDB (CGO) |
| mcp-reddit | Reddit API | Network | HTTPS OAuth2 |
| Operator | swarm.db | File | sqlite3 CLI |
| Operator | obs.db | File | DuckDB CLI |

## Data flow summary

1. Operator inserts seed → `swarm.db`
2. Scheduler polls `swarm.db` → dispatches task to Router
3. Router builds prompt → calls Ollama `/api/generate`
4. Router optionally calls mcp-reddit via MCP client → Reddit API
5. Router parses output → updates `swarm.db`, calls `obs.WriteEvent`
6. obs writes to `obs.db` + mirrors to stdout
7. On terminal state → Deposit writes artifacts to `output/`
