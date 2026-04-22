# Design - Autonomous Local Swarm (ALS)

## Feature
- Problem: A human operator seeds a problem brief; the swarm autonomously decomposes, researches, and builds Go-based assets until a terminal state is reached and outputs are deposited.
- Adoption mode: greenfield
- Feature ID: 0000001-autonomous-local-swarm

## Product Layer
- User stories confirmed: 1
  - As a Human on the Loop, I want the swarm to tackle a given problem and asynchronously work on it until it has the answer.
- Acceptance criteria confirmed: 1
  - The swarm reaches a terminal state (`completed` or `pruned`) and deposits all outputs (artifacts, reports) into the project directory per the job brief's delivery spec.
- Constraints:
  - Must operate within 8GB VRAM (max 1 model loaded at a time)
  - Local-first; no external dispatch — swarm writes outputs, other systems handle delivery
  - Max task recursion depth: 5
  - Viability threshold: 5/10 (below triggers cascading prune)
  - Variational execution candidate count: 3
- Integrations:
  - Ollama (local LLM inference)
  - MCP tool layer (Brave Search, GitHub, context-mode, fetch, Wikipedia, Hacker News)

## Architecture Layer
- Latency target: Scheduler poll interval ≤ 2 seconds for unblocked task pickup
- Availability target: local process — no SLO; operator restarts on failure
- Scalability target: single-machine, 8GB VRAM hard ceiling; depth cap enforces recursion bound
- Security: no auth required; local-only; no credentials stored in swarm; operator supplies MCP API keys via env vars
- Data privacy: no PII, no regulated data; all data is operator-generated task content
- Observability: structured event log in DuckDB (`obs.db`), mirrored to stdout via `log/slog`; every task state transition is a row
- Cost boundary: local hardware only; no cloud spend

## Engineering Layer
- Stack:
  - Language: Go (orchestrator), TypeScript/Node (mcp-reddit)
  - Runtime: local binary (Go), Node process (TypeScript MCP server)
  - Database (app state): SQLite (`swarm.db`)
  - Database (observability): DuckDB (`obs.db`)
  - LLM inference: Ollama (HTTP API)
  - MCP client: embedded in orchestrator
  - Skills: Agent Skills format (agentskills.io) — `skills/` directory; skill-loader in orchestrator
  - CI: not applicable (local tool)
  - IaC: not applicable
- Components:
  - `orchestrator` — main loop; DAG scheduler; task dispatch; poll interval 2s
  - `decomposer` — wraps Phi-4 Mini via Ollama; shatters seed into atomic task JSON
  - `router` — intent-based semantic router; maps task type to model and tool set
  - `ollama-client` — Ollama HTTP wrapper; manages `keep_alive: 0` between model swaps
  - `mcp-client` — MCP protocol client; registers and calls external tool servers
  - `variational-engine` — runs N candidates per task; invokes judge model for selection
  - `validator` — wraps `go build` via `os/exec`; emits self-correction tasks on failure
  - `obs` — DuckDB event logger; all state transitions; mirrored to stdout
  - `mcp-reddit` — TypeScript MCP server; wraps Reddit OAuth API; pauses on rate limit (429 backoff); exposes search/post/comment tools
- Data ownership:
  - `orchestrator` owns `swarm.db` (projects, tasks, task_candidates)
  - `obs` owns `obs.db` (events)
  - No other component writes to either database
- Deployment: local binary, single process, operator-launched
- API versioning: not applicable — no external API exposed

## MCP Tool Servers (registered at launch)
| Server | Purpose | Credentials |
|---|---|---|
| Brave Search | Web search for RESEARCH tasks | `BRAVE_API_KEY` env var |
| GitHub MCP | Code search + repository pull | `GITHUB_TOKEN` env var |
| context-mode | Indexed retrieval + context management | none |
| mcp-server-fetch | Arbitrary URL fetching | none |
| Wikipedia MCP | Structured domain knowledge | none |
| Hacker News MCP | Market signals + trend research | none |
| mcp-reddit (custom) | Reddit search + post/comment retrieval; rate-limit aware | `REDDIT_CLIENT_ID`, `REDDIT_CLIENT_SECRET` env vars |
| ArXiv MCP | Academic and technical paper search + retrieval | none |
| YouTube Transcript MCP | Pull transcripts from video content for research | none |
| Memory MCP | Persistent knowledge graph across runs (complements per-branch JSON sidecar) | none |

## Task Type → Model Routing

System prompts are packaged as Agent Skills ([agentskills.io](https://agentskills.io)) in `skills/`. The router loads `name` + `description` at startup; full `SKILL.md` body is injected as the system prompt on task dispatch.

| Task Type | Model | Skill | Source |
|---|---|---|---|
| `DECOMPOSE` | Phi-4 Mini | `skills/als-decompose/` | custom |
| `RESEARCH` | Llama-3.1 8B | `skills/als-research/` | custom |
| `GO_CODE` | Qwen-2.5-Coder 7B | `skills/als-go-code/` | custom |
| `VIABILITY_REVIEW` | Phi-4 Mini | `skills/als-viability-review/` | custom |
| `SUMMARY` | SmolLM3 135M | `skills/als-summary/` | custom |

## Official Skills (Anthropic — subset)

Bundled from [anthropics/skills](https://github.com/anthropics/skills) as reference and supplementary instruction material.

| Skill | Path | Used by |
|---|---|---|
| `mcp-builder` | `skills/mcp-builder/` | `GO_CODE` tasks that generate MCP server output |
| `webapp-testing` | `skills/webapp-testing/` | `GO_CODE` validation and test-writing patterns |

## Scope
- In:
  - DAG scheduler with blocked_by dependency resolution
  - Cascading prune on VIABILITY_REVIEW score < 5
  - Variational execution with judge selection
  - Go build validation + self-correction feedback loop
  - Sticky model strategy (minimise VRAM swap tax)
  - MCP tool layer with 10 registered servers
  - Agent Skills (agentskills.io format): 5 custom task-type skills + 2 official Anthropic skills (`mcp-builder`, `webapp-testing`)
  - Skill-loader in orchestrator: reads SKILL.md, injects body as system prompt on task dispatch
  - Research-first task priority: DECOMPOSE → RESEARCH → VIABILITY_REVIEW is the primary flow; GO_CODE is secondary
  - DuckDB observability log + stdout mirror
  - Output artifact deposit to project directory
- Out:
  - Email dispatch (handled by external system)
  - Dropbox sync (handled by external system)
  - Authentication / access control
  - Cloud deployment
  - Multi-machine distribution
- Deferred:
  - Specific job brief schema (defined per-job, not in swarm core)
  - Additional MCP server registrations beyond initial 6

## Assumptions
- Ollama is running and accessible at localhost before swarm starts — impact if wrong: orchestrator fails at boot; operator must start Ollama first
- All specified models are pulled in Ollama before first run — impact if wrong: inference calls fail; operator must `ollama pull` each model
- MCP API keys are set as env vars before launch — impact if wrong: MCP tool calls fail gracefully; RESEARCH tasks degrade to Ollama-only output

## Risks
- SLM output quality: 3-7B models produce malformed JSON or invalid Go — likelihood: high; impact: medium; mitigated by decomposition, variational execution, and self-correction loop
- VRAM OOM if `keep_alive: 0` not enforced between swaps — likelihood: medium; impact: high; mitigated by strict ollama-client implementation
- Cascading prune over-fires on weak viability scorer — likelihood: medium; impact: medium; mitigated by configurable threshold

## Dependencies
- Upstream: Ollama (must be running), local models (must be pulled), MCP API keys (env vars)
- Downstream: project directory (output deposit target), external dispatch systems (email, Dropbox)

## Confirmation
Human confirmed this design before proceeding: yes
Date confirmed: 2026-04-22
