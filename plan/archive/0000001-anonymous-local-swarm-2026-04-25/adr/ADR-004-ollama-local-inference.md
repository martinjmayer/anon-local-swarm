---
title: "ADR-004: Ollama for Local LLM Inference"
summary: "Ollama is chosen as the local LLM inference runtime, accessed via its HTTP API."
status: "accepted"
version: "0.1.0"
---
# ADR-004 - Ollama for Local LLM Inference

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator
**Status:** accepted
**Date:** 2026-04-22

---

## Context

The swarm must run 3-7B parameter models locally within an 8GB VRAM constraint. A runtime is needed that: manages model loading/unloading from GPU memory, exposes a stable HTTP API for inference calls, supports `keep_alive` control for VRAM management, and runs as a separate process the orchestrator communicates with over HTTP. The models required are: phi-4:mini, llama3.1:8b, qwen2.5-coder:7b, smollm3:135m.

---

## Decision

Use **Ollama** as the local LLM inference runtime, called via its HTTP API at `http://localhost:11434`.

Ollama is the de-facto standard for running open-weight models locally. It supports all required models, exposes `keep_alive: 0` for explicit VRAM unloading between model swaps, and provides a simple REST API (`/api/generate`, `/api/chat`). Running as a separate process means the orchestrator is decoupled from inference — a model crash does not crash the orchestrator. The Ollama API is stable and widely documented.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| llama.cpp directly | Maximum control, no intermediary | Complex API, no model management, requires per-model compilation flags, no keep_alive abstraction | Operational complexity too high for a single-operator local tool |
| LM Studio | User-friendly GUI, good model management | No headless/CLI mode suitable for programmatic use; not designed as a daemon called by Go HTTP client | Not headless-capable |
| Hugging Face Transformers (Python) | Largest model ecosystem | Python runtime dependency; complex VRAM management; no natural Go integration | Language and runtime mismatch |
| vLLM | High-throughput inference, production-grade | Designed for multi-user serving; heavy setup; no `keep_alive` equivalent for VRAM cycling | Overkill for single-operator local use; setup burden |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | All inference routed through Ollama HTTP API |
| router | Model routing table maps task types to Ollama model identifiers |
| variational-engine | Runs N inference calls via Ollama; judge call also via Ollama |

---

## Consequences

**Positive:**
- Decoupled inference — orchestrator resilient to model-level failures
- `keep_alive: 0` enables deterministic VRAM unloading between swaps
- Operator manages models independently (`ollama pull`, `ollama list`) without touching the orchestrator
- Stable, well-documented HTTP API

**Negative:**
- Orchestrator depends on Ollama being running before startup — adds a startup dependency
- Ollama version changes could alter `keep_alive` behaviour (mitigated by version check at startup)
- Network overhead of HTTP vs in-process inference (negligible at localhost)

**Risks:**
- Ollama does not honour `keep_alive: 0` in all versions/configurations — VRAM OOM risk; mitigated by startup version check and monitoring

---

## Related ADRs

- ADR-005 - related-to (MCP tool context injected into Ollama prompts)
- ADR-010 - depends-on (VRAM strategy relies on Ollama's keep_alive mechanism)

---
