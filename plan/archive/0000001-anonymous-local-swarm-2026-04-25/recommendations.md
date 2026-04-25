# Recommendations — anonymous-local-swarm

**Generated:** 2026-04-25
**Phase:** Post-documentation

These are suggested improvements for future iterations. None are blockers for the current build.

---

## High priority

### 1. Add `package-lock.json` to `src/mcp-reddit/`
**File:** `src/mcp-reddit/`
**Why:** Without a lockfile, `npm install` resolves to latest compatible versions. A transitive dependency update could silently break the build. Run `npm install --package-lock-only` and commit.

### 2. Add `post_id` format validation in mcp-reddit
**File:** `src/mcp-reddit/src/index.ts`
**Why:** `post_id` is passed to the Reddit API without format validation. Add a check matching `/^[a-z0-9]{1,10}$/i` before calling `reddit.getPost()`.

### 3. Add `reddit_hot` unit test
**File:** `src/mcp-reddit/src/reddit.test.ts`
**Why:** `reddit_hot` is the only tool without a unit test. The pattern is identical to `reddit_search`.

---

## Medium priority

### 4. Replace `router.unblockDependents` N+1 query with a single SQL UPDATE
**File:** `src/orchestrator/internal/router/router.go`
**Why:** Current implementation calls `GetTask` for each dependent task when unblocking. A single `UPDATE tasks SET status='pending' WHERE blocked_by=? AND status='blocked'` eliminates the N+1 pattern. Not a concern at current task volumes but worth fixing before scaling.

### 5. Promote integration tests for the router
**File:** `src/orchestrator/internal/router/`
**Why:** The router is the most complex subsystem and has no automated tests. Consider a fake Ollama server (similar to the pattern in `client_test.go`) to drive the full dispatch pipeline without a live model.

### 6. Add obs.db rotation / size threshold warning
**File:** `src/obs/obs.go`
**Why:** `obs.db` grows unbounded. For long-running operators with many projects, this will become a problem. Add a configurable `obs_max_size_mb` that emits a warning event when the threshold is exceeded.

---

## Low priority

### 7. Replace mcp-reddit write guard keyword heuristic with strict allowlist
**File:** `src/mcp-reddit/src/index.ts`
**Why:** The current write guard checks tool names for write-like keywords. A strict `switch` on known tool names (rejecting the `default` case) is more robust.

### 8. Add structured startup health check to orchestrator
**File:** `src/orchestrator/cmd/swarm/main.go`
**Why:** At startup, validate: Ollama reachable, all configured models present, output directory writable, obs.db writable. Log each check as an obs event. Currently R-009 and R-010 are open risks.

### 9. Consider pinning base image digests in Dockerfile
**File:** `Dockerfile`
**Why:** `golang:1.23-bullseye` and `node:20-bullseye-slim` are already pinned by digest in the current build cache. Make this explicit in the Dockerfile `FROM` lines to prevent silent base image drift.
