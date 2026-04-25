# Quirks — orchestrator

## Q-001: Go toolchain required at build time, not just runtime

The Docker build compiles the binary inside the container. On the host, running `go mod tidy` or `go test` requires Go 1.23+ and MinGW-w64 (for CGO on Windows). The Docker build handles this automatically, but host-side development requires the full toolchain.

## Q-002: go-duckdb requires CGO

`als/obs` (imported by orchestrator) depends on `github.com/marcboeker/go-duckdb`, which requires CGO enabled and GCC at build time. All builds must set `CGO_ENABLED=1`. The Dockerfile builder stage installs `gcc libc-dev` to satisfy this.

## Q-003: `unblockDependents` issues an extra GetTask query per dependent

`router.go` calls `GetTask` for each dependent task when unblocking after completion. This is an N+1 pattern — a single `UPDATE tasks SET status='pending' WHERE blocked_by=? AND status='blocked'` would be more efficient. Logged as a future optimisation; at current task volumes (<100 tasks per run) this is not a performance concern. See `src/orchestrator/docs/` for the tech-debt note if this is promoted.

## Q-004: Scheduler test uses `maxDepth=4` not `maxDepth=5`

The SQLite CHECK constraint limits `depth` to 0–5. The depth-pruning test (`TestScheduler_req002_PrunesExceedingMaxDepth`) uses `maxDepth=4` with a task at `depth=5` to exercise the prune path, since inserting `depth=6` is rejected by the constraint. The production max depth is configured in `config.yaml` and defaults to 5.
