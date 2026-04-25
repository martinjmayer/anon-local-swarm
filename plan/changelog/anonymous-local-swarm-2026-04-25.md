# Iteration Log - anonymous-local-swarm

**Date:** 2026-04-25
**Feature ID:** anonymous-local-swarm
**Tool:** claude-code

---

## Iteration Steps completed

- [x] Specification (Phase 1) — 12 requirements, scope, risk register, domain glossary, operational model, SLO definitions, cost model
- [x] Architecture Decisions (Phase 2) — 12 ADRs (Go orchestrator, SQLite, DuckDB, Ollama, MCP, TypeScript MCP Reddit, skills/system prompts, DAG model, separate DBs, VRAM strategy, variational execution, context sidecar)
- [x] Code Generation (Phase 3) — 3 components: obs (Go), orchestrator (Go), mcp-reddit (TypeScript)
- [x] Validation (Phase 4) — Docker build passing; 10 test packages green; 3 test fixes required (see self-correct log)
- [x] Security Assessment (Phase 5) — 5 findings; 3 fixed inline; 2 deferred to recommendations
- [x] Documentation (Phase 6) — per-component docs, system registry, dependency graph, iteration log

---

## Assumptions made

- A-001: Ollama is running and reachable at localhost:11434 before swarm starts
- A-002: All required Ollama models are pre-pulled by the operator
- A-003: The operator's machine has a GPU with sufficient VRAM for the configured models (or CPU fallback is acceptable at reduced speed)
- A-004: Reddit app credentials are obtained by the operator (free Reddit API account)
- A-005: The operator runs Docker Desktop on Windows with WSL2 backend, data root on C:\DockerData

---

## Infrastructure notes

Docker Desktop setup required non-trivial configuration during P4 validation:
- WSL2 backend required; Hyper-V backend rejects VHD creation on external SSD
- Docker Desktop reinstalled with `--wsl-default-data-root=C:\DockerData` to prevent tmpfs exhaustion
- `.wslconfig` swap file reference to R: drive removed to avoid WSL parse errors
- `docker-credential-desktop` missing from PATH on fresh install — resolved by removing `credsStore` from `~/.docker/config.json`

---

## Quirks

- **Q-001 (obs, orchestrator):** go-duckdb requires CGO + GCC. On Windows host, MinGW-w64 required. Docker build handles this automatically.
- **Q-002 (orchestrator):** Scheduler depth-pruning test uses `maxDepth=4` (not 5) because the SQLite CHECK constraint prevents inserting `depth=6`. Production config defaults to `max_depth: 5`.
- **Q-003 (orchestrator):** `router.unblockDependents` has an N+1 query pattern — see recommendations.

---

## Security fixes applied

1. `.gitignore` — added `.env`, `swarm.db`, `obs.db`, `output/`
2. `deposit/deposit.go` — path traversal guard: `strings.HasPrefix` check before `os.WriteFile`
3. `Dockerfile` — non-root runtime user (`USER swarm`)

---

## Recommendations

See `plan/current/recommendations.md` for 9 specific improvement items.

Top 3:
1. Commit `package-lock.json` for mcp-reddit (dependency pinning)
2. Add `post_id` format validation in mcp-reddit
3. Replace `router.unblockDependents` N+1 query pattern

---

## Self-correct log

| Step | What failed | Fix applied |
|---|---|---|
| P4 build — step 1 | `obs_test.go:193` — assignment mismatch: 2 variables but `QueryRow.Scan` returns 1 | Removed spurious `_` from `if _, err :=` |
| P4 build — step 2 | `t.Context()` undefined in validator_test.go and ollama/client_test.go — Go 1.24 method, image uses 1.23 | Added `"context"` import; replaced `t.Context()` with `context.Background()` |
| P4 build — step 3 | `TestScheduler_req002_PrunesExceedingMaxDepth` fails — SQLite CHECK rejects depth=6, task stays at depth=5, 5 > 5 is false | Changed test to `maxDepth=4`; depth=5 > 4 exercises the same code path |
