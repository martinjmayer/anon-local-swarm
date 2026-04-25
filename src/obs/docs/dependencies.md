# Dependencies — obs

## What obs consumes

| Dependency | Type | Version | Purpose |
|---|---|---|---|
| `github.com/marcboeker/go-duckdb` | Go module | latest | DuckDB driver for obs.db |
| `log/slog` | Go stdlib | — | Structured stdout mirror |
| CGO + GCC | Build toolchain | — | Required by go-duckdb |

## What depends on obs

| Component | Import path | Usage |
|---|---|---|
| orchestrator | `als/obs` | Calls `WriteEvent` throughout all subsystems |

## Dependency direction

```
orchestrator → obs
```

obs imports nothing from other ALS components. This is a hard rule — importing orchestrator from obs would create a cycle.

## Build-time note

`go-duckdb` requires CGO enabled and GCC available. On the host this means MinGW-w64 on Windows. In the Docker build this is satisfied by `apt-get install gcc libc-dev` in the builder stage.
