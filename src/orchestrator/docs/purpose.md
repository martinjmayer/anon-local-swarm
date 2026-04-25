# Purpose — orchestrator

## What this component exists to do

The orchestrator is the central nervous system of the Autonomous Local Swarm. It owns `swarm.db` (SQLite), runs the DAG scheduler, routes tasks to Ollama models, manages the full task lifecycle, and deposits output artifacts when a project reaches terminal state.

It is the only component that reads from and writes to `swarm.db`. It calls `obs.WriteEvent` to record every state change but never reads from `obs.db`. It spawns MCP server subprocesses and communicates with them via the MCP stdio protocol to give LLM agents access to external tools.

## Core responsibilities

| Subsystem | Package | What it does |
|---|---|---|
| Scheduler | `internal/scheduler` | Polls `swarm.db` every N ms; dispatches unblocked pending tasks; enforces max depth; recovers from panics |
| Router | `internal/router` | Maps task type → Ollama model; builds prompts from skills; handles output markers |
| Ollama client | `internal/ollama` | HTTP POST to Ollama `/api/generate`; enforces `keep_alive: 0`; parses output markers |
| Variational engine | `internal/variational` | Runs N parallel inference calls; collects candidates; invokes Judge model |
| Validator | `internal/validator` | Runs `go build` on GO_CODE output; constructs self-correction payloads |
| Context sidecar | `internal/sidecar` | Writes per-task context files; detects threshold breach; emits SUMMARY tasks |
| MCP client | `internal/mcp` | Spawns and manages MCP server processes; dispatches tool calls; injects results into prompts |
| Deposit | `internal/deposit` | Writes terminal-state output artifacts to the project output directory |
| DB | `internal/db` | All swarm.db reads and writes; parameterised queries; WAL mode; foreign keys |
| Config | `internal/config` | Loads `config.yaml`; resolves env var refs; validates all fields |
| Skills loader | `internal/skills` | Reads SKILL.md files from `skills/`; exposes system prompts by task type |

## Why it is a single binary

Per ADR-001, Go was chosen for the orchestrator because goroutine-per-task concurrency maps naturally to the parallel DAG dispatch model, and a single binary simplifies local deployment (no runtime to install beyond the Docker image).
