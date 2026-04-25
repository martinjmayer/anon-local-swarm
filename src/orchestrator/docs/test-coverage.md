# Test Coverage — orchestrator

## Summary

| Type | Coverage | Notes |
|---|---|---|
| Unit | ~75% | All packages with logic have tests; router has partial coverage |
| Integration | 0% | No cross-component integration tests |
| E2E | 0% | No end-to-end tests (no HTTP surface; full run requires live Ollama) |

## Coverage by package

| Package | Test file | Key scenarios covered |
|---|---|---|
| `internal/config` | `config_test.go` | Default loading, env var override, validation errors |
| `internal/db` | `db_test.go` | Insert/get/update task and project, PendingTasks, prune |
| `internal/deposit` | `deposit_test.go` | Artifact write, path traversal guard |
| `internal/ollama` | `client_test.go` | keep_alive enforcement, stream=false, system prompt, error propagation, output parsing |
| `internal/scheduler` | `scheduler_test.go` | Task dispatch, depth pruning, panic recovery |
| `internal/sidecar` | `sidecar_test.go` | Context write/read, threshold detection |
| `internal/skills` | `loader_test.go` | SKILL.md loading, missing skill fallback |
| `internal/validator` | `validator_test.go` | SelfCorrectionPayload, RetryCount, IncrementRetryMeta, Validate (valid/invalid code) |
| `internal/variational` | `variational_test.go` | Candidate generation, selection |
| `cmd/swarm` | — | No test files (entry point wiring only) |
| `internal/mcp` | — | No test files (subprocess I/O hard to unit test) |
| `internal/router` | — | No test files (requires Ollama; covered by scheduler integration) |

## What is not tested

- Full router dispatch (requires live Ollama)
- MCP client subprocess lifecycle
- Multi-task DAG execution with real blocked_by chains
- Cascading prune across a real project tree

## Test files

`src/orchestrator/internal/*/` — each package has a `*_test.go` file in the same directory.
