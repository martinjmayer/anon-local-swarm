# Security Report - anonymous-local-swarm

**Date:** 2026-04-25
**Reviewer:** planifest-security-agent
**Scope:** `src/obs/`, `src/orchestrator/`, `src/mcp-reddit/`, `Dockerfile`, `docker-compose.yml`
**Overall Risk Rating:** Medium

---

## Threat Model (STRIDE)

| Threat | Category | Severity | Mitigation |
|---|---|---|---|
| LLM-generated artifact filename escapes output directory via `../` traversal | Tampering | Medium | Not mitigated — `filepath.Join` normalises but does not contain paths |
| Container process runs as root; volume-mounted files owned by root on host | Elevation of Privilege | Medium | Not mitigated — no `USER` directive in Dockerfile |
| Operator accidentally commits `.env` containing API credentials | Information Disclosure | Medium | Not mitigated — `.env` absent from `.gitignore` |
| Prompt injection via malicious task payload manipulates LLM behaviour | Tampering | Medium | Partially mitigated — skills system prompts provide structure; no hard guardrails |
| Credentials leak into obs.db `detail` JSON via non-obvious key names | Information Disclosure | Low | Partially mitigated — `scrubMap` redacts credential-keyed fields; payload/message keys not scrubbed |
| Reddit API credentials exposed in mcp-reddit error output | Information Disclosure | Low | Mitigated — `scrubCredentials()` applied to all error messages before return |
| Unauthenticated HTTP to Ollama intercepted on local network | Spoofing | Low | Accepted — local-only deployment; no TLS or auth to Ollama is standard practice |
| mcp-reddit write guard bypassed via unexpected tool name | Elevation of Privilege | Low | Partially mitigated — keyword heuristic guard present; strict allowlist preferred |
| `post_id` input to `reddit_get_post` sent to Reddit API without format validation | Tampering | Low | Not mitigated — no format check before API call |
| No audit trail for failed container starts or missing config | Repudiation | Low | Accepted — container logs capture stderr; obs.db records all post-start events |
| Runtime databases (swarm.db, obs.db) committed to version control | Information Disclosure | Low | Not mitigated — neither file is in `.gitignore` |

---

## Dependency Audit

### Go (`src/orchestrator/go.mod`, `src/obs/go.mod`)

| Package | Version | Notes |
|---|---|---|
| `github.com/mattn/go-sqlite3` | current | CGO-based; no known active CVEs; well-maintained |
| `github.com/marcboeker/go-duckdb` | current | CGO-based; actively maintained |
| `gopkg.in/yaml.v3` | current | No known CVEs; widely used |
| `github.com/google/uuid` | current | No known CVEs |

No concerning Go dependencies identified. `go mod download` completes cleanly.

### Node.js (`src/mcp-reddit/package.json`)

| Package | Notes |
|---|---|
| `@modelcontextprotocol/sdk` | Official SDK; actively maintained by Anthropic |
| `node-fetch` / built-in fetch | Node 20 uses built-in fetch; no external HTTP lib introduced |

**Action required:** No `package-lock.json` was found in the repository. Without a lockfile, `npm install` resolves to latest compatible versions — dependency pinning is absent. Run `npm install --package-lock-only` and commit the lockfile.

Build warning noted: `inflight@1.0.6` (transitive dep) is deprecated and leaks memory; `glob@7.2.3` contains publicised security vulnerabilities. These are transitive and not directly importable by application code, but should be resolved by updating the direct dependency tree.

---

## Secrets Management

**Correctly implemented:**
- All five credential env vars (`BRAVE_API_KEY`, `GITHUB_TOKEN`, `REDDIT_CLIENT_ID`, `REDDIT_CLIENT_SECRET`, `OBSIDIAN_API_KEY`) are sourced exclusively from environment variables injected via `docker-compose.yml`.
- `src/mcp-reddit/src/index.ts:209` — `scrubCredentials()` strips credential values from all error messages before returning to the caller.
- `src/obs/obs.go:193` — `scrubMap()` redacts values for credential-keyed fields before writing to obs.db.
- mcp-reddit fails fast (`process.exit(1)`) if credentials are missing — no silent degraded mode.

**Gap — `.env` not in `.gitignore`:**
`C:/d/anon-local-swarm/.gitignore` does not contain `.env`. An operator who copies `.env.example` to `.env` and runs `git add .` will commit live credentials. Add `.env` to `.gitignore` immediately.

**Gap — runtime databases not in `.gitignore`:**
`swarm.db` and `obs.db` are not excluded from version control. `obs.db` records task payloads and detail fields. If committed, historical runs are exposed. Add both to `.gitignore`.

**Gap — partial `scrubMap` coverage:**
`src/obs/obs.go:193` — `scrubMap` redacts credential-keyed fields by key name heuristic. If a credential value flows into a detail field with a non-standard key (e.g., `"message"`, `"payload"`, `"raw"`), it will not be redacted. Credentials do not currently flow through task payloads (they are confined to MCP server env vars), so the practical risk is low — but the coverage should be documented as a known boundary.

---

## Authentication & Authorisation Review

This system has no HTTP API surface exposed externally. There are no inbound endpoints to assess for auth gaps.

**Ollama:** Plain HTTP to `http://localhost:11434` (or `host.docker.internal:11434` in Docker). Ollama does not require authentication by default. Acceptable for local operator deployment; not acceptable if the host's Ollama port is ever exposed to a network.

**Reddit OAuth:** App-only `client_credentials` grant. Token refresh on 401 is implemented in `src/mcp-reddit/src/reddit.ts`. No user-level OAuth is present. ✓

---

## Input Validation Review

**SQL injection — clean:**
All database operations in `src/orchestrator/internal/db/db.go` use parameterised queries via `database/sql` placeholders (`?`). No string interpolation into SQL was found. ✓

**Path traversal in deposit — not mitigated (Medium):**
`src/orchestrator/internal/deposit/deposit.go:33`:
```go
path := filepath.Join(outputDir, a.Filename)
os.WriteFile(path, []byte(a.Content), 0644)
```
`a.Filename` originates from LLM output (task `Payload` field, parsed by `ollama.ParseOutput`). `filepath.Join` cleans the path but does not prevent escaping the base directory — `filepath.Join("/data/output/proj", "../../swarm.db")` resolves to `/data/swarm.db`. A prompt-injected or adversarially-guided LLM could overwrite the application database or any volume-mounted file. **Fix:** validate that the cleaned path has `outputDir` as a prefix before writing.

**Prompt injection — partially mitigated (Medium):**
`task.Payload` is inserted into Ollama prompts without sanitisation. The system prompt from the skills file provides structural context, but there are no hard guardrails preventing a crafted payload from instructing the LLM to emit `NEW_TASK:` lines with malicious content, manipulate `VIABILITY_SCORE:`, or cause excessive task fan-out. This is inherent to LLM-based systems and is mitigated operationally by the seed being operator-controlled and maxDepth enforced. Acceptable for the current threat model (trusted operator seed).

**mcp-reddit input limits — absent (Low):**
`src/mcp-reddit/src/index.ts` — `query` in `reddit_search` and `post_id` in `reddit_get_post` have no length or format validation before being sent to the Reddit API. Reddit's API will reject malformed requests, but defensive validation (max length, alphanumeric check for `post_id`) should be added.

**LLM-generated Go code execution — not executed:**
`src/orchestrator/internal/validator/validator.go` writes LLM output to a temp directory and runs `go build`. The code is compiled but not executed. Build artefacts are discarded after the check. The temp directory is isolated by `os.MkdirTemp`. ✓

---

## Network Policy

| Connection | Direction | Protocol | Auth |
|---|---|---|---|
| Container → Ollama (host:11434) | Egress | HTTP | None |
| Container → Reddit API | Egress | HTTPS | OAuth2 Bearer |
| Container → Brave/GitHub/Obsidian (via MCP) | Egress | HTTPS | API key (env var) |
| Host → Container | None — no `ports:` declared | — | — |

**No inbound ports are exposed.** The container has no network server. ✓

**`extra_hosts: host.docker.internal:host-gateway`** — this allows the container to reach any port on the Windows host, not just Ollama. This is standard Docker Desktop practice and acceptable for a local tool, but means a compromised container can probe host services.

---

## Infrastructure as Code Review

No cloud IaC (Terraform, Pulumi, CDK). Docker Compose is the only infrastructure definition.

**Container runs as root — not mitigated (Medium):**
`Dockerfile` has no `USER` directive. The `swarm` binary and Node.js MCP subprocess both run as UID 0 inside the container. Volume-mounted files (`swarm.db`, `obs.db`, `output/`) are created with root ownership on the host. While a local-only tool, best practice is `USER nonroot`. Add to the runtime stage:
```dockerfile
RUN adduser --disabled-password --no-create-home swarm
USER swarm
```

**Base images — acceptable:**
- `golang:1.23-bullseye` — pinned by digest in build stage. ✓
- `node:20-bullseye-slim` — LTS; pinned by digest. ✓
- No `latest` tags used in production stages. ✓

**Volume mounts — acceptable:**
- `config.yaml` mounted `:ro`. ✓
- `skills/` mounted `:ro`. ✓
- `swarm.db` and `obs.db` are read-write (required). Acceptable.

---

## Risk Register Cross-Reference

| Risk ID | Description | Status |
|---|---|---|
| R-012 | Credential leakage into obs.db / payloads | **Mitigated** — scrubCredentials + scrubMap implemented |
| R-013 | mcp-reddit exposes write endpoint | **Mitigated** — read-only tool registration + write keyword guard |
| R-007 | Reddit OAuth token expiry | **Mitigated** — token refresh on 401 in RedditClient |
| R-009 | Missing Ollama models at startup | **Open** — model check at startup logged in risk register; verify in operator guide |
| R-010 | Output dir not writable | **Open** — writability check noted; verify implementation in router |
| R-004 | SQLite write contention during prune | **Mitigated** — WAL mode enabled in `db.Open()` |

---

## Summary

**Overall risk rating: Medium**

The system is a local, operator-controlled tool with no inbound network surface. No hardcoded credentials, no SQL injection, no externally exposed ports. The credential scrubbing pipeline is well-implemented.

**Top actions before production:**

1. **Add `.env`, `swarm.db`, and `obs.db` to `.gitignore`** (`C:/d/anon-local-swarm/.gitignore`) — prevents credential and data commits. Trivial fix; high impact.

2. **Fix path traversal in `deposit.go`** (`src/orchestrator/internal/deposit/deposit.go:33`) — validate that `filepath.Join(outputDir, a.Filename)` stays within `outputDir` before writing. One-line check with `strings.HasPrefix`.

3. **Add non-root user to Dockerfile** — add `RUN adduser` + `USER swarm` to the runtime stage to avoid root-owned volume files and reduce container escape impact.

4. **Commit `package-lock.json` for mcp-reddit** — run `npm install --package-lock-only` in `src/mcp-reddit/` and commit. Pins transitive dependencies.

5. **Add `post_id` format validation in mcp-reddit** (`src/mcp-reddit/src/index.ts`) — validate `post_id` matches Reddit's alphanumeric ID format (`/^[a-z0-9]{1,10}$/i`) before calling the API.
