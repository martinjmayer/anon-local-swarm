---
title: "ADR-010: Sticky Model Strategy and keep_alive: 0 for VRAM Management"
summary: "The swarm enforces MAX_VRAM_MODELS=1 via keep_alive: 0 on every Ollama call and a sticky model scheduler that batches tasks by model to minimise swap frequency."
status: "accepted"
version: "0.1.0"
---
# ADR-010 - Sticky Model Strategy and keep_alive: 0 for VRAM Management

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator
**Status:** accepted
**Date:** 2026-04-22

---

## Context

The target hardware has 8GB VRAM. The swarm uses five models ranging from 135M to 8B parameters. Loading multiple models simultaneously would cause OOM. A strategy was needed for ensuring only one model is in VRAM at a time, while minimising the latency cost of model swaps (the "loading tax" of SSD → VRAM transfers).

---

## Decision

Enforce **MAX_VRAM_MODELS = 1** via two mechanisms:
1. **`keep_alive: 0`** on every Ollama inference request — forces model unload immediately after each response
2. **Sticky model scheduler** — when multiple tasks are pending, the scheduler prioritises tasks sharing the same `model_tag` as the most recently active model, batching them before switching

Together these ensure: no two models coexist in VRAM, and the number of costly model swaps is minimised.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| Let Ollama manage VRAM (default keep_alive: 5m) | No orchestrator logic needed | Ollama may keep a model loaded while the next task needs a different model; OOM risk on 8GB systems | Gives up VRAM control; OOM risk is unacceptable |
| Pre-load all models at startup | Eliminates swap latency | 4 models × 3-8GB each = impossible on 8GB VRAM | Physically impossible |
| Quantised models to fit multiple in VRAM | Reduced swap frequency | Quality degradation from aggressive quantisation; complex model selection logic | Quality trade-off too significant for a quality-critical system |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | Scheduler implements sticky model prioritisation |
| ollama-client | Sets keep_alive: 0 on every request |
| obs | Logs model load/unload events for swap monitoring |

---

## Consequences

**Positive:**
- Deterministic VRAM behaviour — OOM prevented by design
- Sticky batching reduces swap frequency, improving throughput when multiple tasks of the same type are queued
- Observable — every swap logged to obs.db

**Negative:**
- `keep_alive: 0` means every inference call pays the model load cost (SSD → VRAM), even for back-to-back calls of the same model. Sticky batching mitigates but does not eliminate this
- Orchestrator complexity: scheduler must track model_tag of active/last-active task

**Risks:**
- Ollama ignores `keep_alive: 0` in some versions — model stays loaded, next swap causes OOM. Mitigated by Ollama version check at startup and VRAM monitoring

---

## Related ADRs

- ADR-004 - depends-on (Ollama's keep_alive mechanism is the enforcement mechanism)
- ADR-008 - related-to (scheduler's sticky strategy integrates with DAG dispatch priority)

---
