# Risk — orchestrator

## Component-scoped risks

| ID | Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| R-orch-001 | SLM output quality — malformed or non-conformant LLM responses | Medium | High | ParseOutput extracts markers defensively; malformed output falls back to marking task failed rather than crashing |
| R-orch-002 | VRAM OOM if Ollama does not honour `keep_alive: 0` | Medium | High | `keep_alive: 0` enforced in every request (tested); Ollama version check at startup recommended (R-009) |
| R-orch-003 | SQLite write contention during cascading prune | Low | Medium | WAL mode enabled; prune writes are serialised through UpdateTaskStatus calls |
| R-orch-004 | Operator forgets to pre-pull required Ollama models | High | High | Orchestrator checks model availability at startup; logs missing models and exits |
| R-orch-005 | Output directory not writable at deposit time | Medium | Medium | Writability check at project start; fail fast with clear error |
| R-orch-006 | Self-correction loop exhausts MAX_DEPTH before fixing a build error | Medium | Medium | Task marked `failed`; operator can inspect stderr chain in obs.db and re-seed |
| R-orch-007 | LLM-generated artifact filename escapes output directory | Low | Medium | Path traversal guard in `deposit.go` — `strings.HasPrefix` check against `outputDir` (security fix applied) |
| R-orch-008 | MCP server process crashes or becomes unresponsive | Medium | Low | MCP client detects timeout; tool call returns error; task proceeds with degraded context |

## Cross-reference

Corresponds to R-001 through R-011 in the system risk register (`plan/current/risk-register.md`).
