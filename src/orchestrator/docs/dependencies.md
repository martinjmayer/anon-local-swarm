# Dependencies — orchestrator

## External dependencies

| Dependency | Type | Purpose |
|---|---|---|
| `github.com/mattn/go-sqlite3` | Go module | swarm.db driver (CGO) |
| `gopkg.in/yaml.v3` | Go module | config.yaml parsing |
| `github.com/google/uuid` | Go module | Task/project ID generation |
| `als/obs` | Internal Go module | WriteEvent calls throughout |
| Ollama HTTP API | Runtime | LLM inference (`POST /api/generate`) |

## Internal component dependencies

| Component | How consumed |
|---|---|
| obs | Imported as `als/obs`; `WriteEvent` called from router, scheduler, deposit, mcp client |
| mcp-reddit | Spawned as a subprocess by `internal/mcp`; communicates via MCP stdio protocol |

## What depends on orchestrator

Nothing. The orchestrator is the top-level component. It is the entry point for the swarm.

## Dependency direction

```
orchestrator → obs
orchestrator → mcp-reddit (subprocess)
orchestrator → Ollama (HTTP)
orchestrator → Reddit API (via mcp-reddit)
```

## Build-time requirements

- Go 1.23+
- CGO enabled
- GCC / libc-dev (for go-sqlite3 and go-duckdb via obs)
- `go.work` workspace linking `als/obs` and `als/orchestrator`
