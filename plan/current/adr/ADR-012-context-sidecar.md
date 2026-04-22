---
title: "ADR-012: Per-Branch JSON Context Sidecar"
summary: "Long-term branch context is stored in per-branch JSON sidecar files on disk, keeping swarm.db lightweight while providing compressible, model-accessible context."
status: "accepted"
version: "0.1.0"
---
# ADR-012 - Per-Branch JSON Context Sidecar

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator
**Status:** accepted
**Date:** 2026-04-22

---

## Context

As tasks in a branch execute, they accumulate context: extracted entities, business constraints, prior research findings, and compression summaries. This context must be available to each subsequent task in the branch as part of its prompt. The context grows over time and must be compressible when it approaches a model's context window limit. Two complementary stores exist: the Memory MCP (persistent cross-run knowledge graph) and a branch-local mechanism for the current run's accumulated context. A decision was needed for the branch-local mechanism.

---

## Decision

Store per-branch context in **JSON sidecar files** on disk, with the file path recorded in `tasks.context_path`.

Each branch has one JSON sidecar file. The file stores: extracted entities, constraints, prior task summaries, and the most recent compression output. The orchestrator reads the sidecar before constructing a task's prompt, appending it as context. When the sidecar's estimated token count exceeds 70% of the active model's context window, a SUMMARY task is emitted; its output replaces the sidecar content. The path is stored in `tasks.context_path` so any task in the branch can locate it without querying for lineage.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| Store context in tasks.meta_data (SQLite JSON column) | No extra files; single store | SQLite rows grow unbounded; JSON column in SQLite is not compressed; harder to stream large context to file I/O | DB bloat; worse compression |
| Memory MCP only | Persistent cross-run; no custom implementation | Memory MCP is for long-term cross-run knowledge, not ephemeral branch-local context; mixing concerns would pollute the knowledge graph | Wrong abstraction; different lifecycle |
| In-memory only (no persistence) | Simplest | Lost on restart or crash; branch context cannot be recovered | No resilience |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | Reads sidecar before prompt construction; writes updated sidecar after task completion |
| tasks table | `context_path` column stores the sidecar file path per task |

---

## Consequences

**Positive:**
- Context survives orchestrator restart — branch can be resumed without re-running prior tasks
- JSON files are inspectable by the operator — context state is human-readable
- Compression (SUMMARY task) replaces file content in place — no DB migration needed
- Decoupled from DB size — large contexts don't bloat swarm.db

**Negative:**
- Additional filesystem I/O on every task dispatch (read sidecar) and completion (write sidecar)
- Sidecar files accumulate in the project directory — operator must clean up manually
- Token counting for context threshold detection requires an approximation (character count / 4) or a tokeniser library

**Risks:**
- Sidecar file corruption (partial write on crash) leaves branch context inconsistent — mitigated by atomic file writes (write to temp file, rename)

---

## Related ADRs

- ADR-004 - related-to (sidecar content injected into Ollama prompts)
- ADR-008 - related-to (sidecar is per-branch; DAG structure defines branch boundaries)

---
