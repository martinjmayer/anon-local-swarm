---
title: "ADR-011: Variational Execution with Judge Model for Output Quality"
summary: "For eligible tasks, the swarm generates N=3 candidate outputs at varied temperatures and uses phi-4:mini as a judge model to select the best, trading inference cost for output quality."
status: "accepted"
version: "0.1.0"
---
# ADR-011 - Variational Execution with Judge Model for Output Quality

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator
**Status:** accepted
**Date:** 2026-04-22

---

## Context

3-7B SLMs produce inconsistent output quality — a single inference may produce a poor result while a second attempt on the same prompt produces an excellent one. The swarm's primary risk (R-001) is SLM output quality. A mechanism was needed to improve output reliability beyond single-shot inference, within the VRAM and compute constraints of local hardware.

---

## Decision

Implement **variational execution**: for eligible tasks, run the same prompt `VARIATION_COUNT = 3` times with varied temperature settings, store each output as a `task_candidate`, then invoke phi-4:mini as a **judge model** to evaluate and select the best candidate.

Temperature variation induces output diversity — the same prompt at temperatures 0.3, 0.7, 1.0 produces meaningfully different candidates. The judge model evaluates quality against a rubric in its system prompt (correctness, structure, relevance) and selects one. The selected candidate becomes the task output. All candidates are retained for audit. This pattern — ensemble generation + judge selection — is well-established for improving SLM reliability.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| Single-shot inference (no variational) | Simplest, fastest, lowest VRAM usage | Does not address R-001; quality variance is the primary risk | Fails to address primary risk |
| Majority voting (N outputs, pick most common) | No judge model needed | Works for classification/short answers; poor fit for Go code generation or research synthesis where outputs are not comparable by equality | Wrong comparison method for open-ended outputs |
| RLHF-tuned models | Permanently improved base quality | Requires fine-tuning infrastructure, datasets, and GPU time far beyond scope | Out of scope for local tool |
| Beam search / best-of-N at Ollama level | Model-native quality improvement | Not exposed in Ollama's API | Not available |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | Implements variational loop; manages candidate generation and judge invocation |
| variational-engine | Package within orchestrator implementing the N-run + judge pattern |
| task_candidates table | Stores all N candidates with scores; is_selected marks the chosen one |

---

## Consequences

**Positive:**
- Directly addresses R-001 (SLM output quality) — ensemble + judge reliably outperforms single-shot for 3-7B models
- All candidates retained — operator can inspect rejected outputs in obs.db
- Configurable — `VARIATION_COUNT` can be set to 1 to disable variational execution per task type

**Negative:**
- 3x inference cost for variational tasks + 1 judge call = 4 Ollama calls per eligible task
- 3 model loads if the task model differs from the judge model (phi-4:mini) — VRAM swap tax multiplied
- Increases end-to-end latency for eligible tasks

**Risks:**
- Judge model (phi-4:mini) selects a poor candidate due to weak evaluation rubric — mitigated by careful judge system prompt in `als-viability-review` skill; rubric is versioned and editable

---

## Related ADRs

- ADR-004 - depends-on (all candidates generated via Ollama)
- ADR-010 - related-to (each variational run triggers VRAM swap if model differs from judge)
- ADR-007 - related-to (judge system prompt defined in Agent Skills format)

---
