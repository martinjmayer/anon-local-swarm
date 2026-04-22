---
title: "ADR-001: Go for the Orchestrator"
summary: "Go is chosen as the primary language for the ALS orchestrator — the DAG scheduler, Ollama client, MCP client, and all coordination logic."
status: "accepted"
version: "0.1.0"
---
# ADR-001 - Go for the Orchestrator

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator
**Status:** accepted
**Date:** 2026-04-22

---

## Context

The ALS orchestrator is a long-running local daemon that: polls SQLite every 2 seconds, manages concurrent task dispatch across multiple Ollama models, enforces VRAM constraints, and coordinates I/O across SQLite, DuckDB, the filesystem, and HTTP (Ollama API, MCP servers). A language choice was needed that handles concurrent I/O efficiently, compiles to a single self-contained binary, and has strong SQLite and HTTP library support.

---

## Decision

Use **Go** for the orchestrator and all its internal packages (scheduler, router, Ollama client, MCP client, variational engine, validator, obs).

Go's goroutine model maps naturally to the swarm's concurrent task dispatch pattern. The scheduler polls SQLite while multiple tasks may be active; goroutines handle each without thread-pool overhead. The compiled binary has no runtime dependency — the operator runs a single executable. Go's standard library covers HTTP (Ollama API), file I/O (context sidecars, output deposit), and process execution (`os/exec` for `go build`). SQLite (via `modernc.org/sqlite` or `mattn/go-sqlite3`) and DuckDB (via `marcboeker/go-duckdb`) both have mature Go drivers.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| Python | Rapid development, rich ML/AI ecosystem, mature SQLite support | GIL limits true concurrency; runtime dependency (Python version, venv); slower startup | Concurrency model is wrong for a tight poll loop with concurrent dispatch |
| TypeScript/Node | Single language across orchestrator and mcp-reddit; large ecosystem | Event loop not ideal for CPU-bound coordination; requires Node runtime; less ergonomic for binary distribution | Runtime dependency; less natural for systems-level process management |
| Rust | Maximum performance, memory safety guarantees, single binary | Steep learning curve; longer development time; smaller LLM-generated code quality for Rust | Development speed matters more than raw performance for a local tool at this scale |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | Primary language — all internal packages written in Go |
| obs | Written in Go as a package within the orchestrator module |
| mcp-reddit | Not affected — remains TypeScript (see ADR-006) |

---

## Consequences

**Positive:**
- Goroutine-based concurrency handles scheduler loop + concurrent task dispatch without complexity
- Single compiled binary; no runtime dependency for the operator
- Strong standard library covers all I/O needs without heavy third-party dependencies
- `go build` validation in the self-correction loop is natural — the orchestrator already runs Go

**Negative:**
- Polyglot project (Go + TypeScript) adds build and toolchain complexity
- Go's error handling verbosity increases code size vs Python equivalent

**Risks:**
- DuckDB Go driver (`go-duckdb`) requires CGO; cross-compilation becomes non-trivial if the operator is on a different OS

---

## Related ADRs

- ADR-006 - depends-on (TypeScript choice for mcp-reddit creates the polyglot situation)
- ADR-003 - related-to (DuckDB driver requires CGO, relevant to Go choice)

---
