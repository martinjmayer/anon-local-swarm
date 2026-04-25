---
title: "Requirement: REQ-004 - VRAM Management"
summary: "Strict single-model VRAM constraint and keep_alive enforcement"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-004 - VRAM Management

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — 8GB VRAM guardrails, sticky model strategy, MAX_VRAM_MODELS: 1
**Priority:** must-have

---

## Functional Requirements
- The orchestrator MUST enforce `MAX_VRAM_MODELS = 1` — only one model may be loaded in Ollama at any time
- Every Ollama inference request MUST include `keep_alive: 0` to ensure the model is unloaded immediately after the response
- The scheduler MUST implement the sticky model strategy: when multiple tasks are pending, tasks sharing the same `model_tag` as the currently active task MUST be prioritised to avoid unnecessary model swaps
- The orchestrator MUST NOT initiate a new model inference call while a model swap (unload + load) is in progress
- VRAM swap tax MUST be logged to the observability store (model unloaded, model loaded, duration)

## Acceptance Criteria
- [ ] Two tasks with different model_tags are not dispatched simultaneously
- [ ] keep_alive: 0 is present in every Ollama API request body
- [ ] When 3 tasks are pending (2 phi-4:mini, 1 llama3.1:8b), the second phi-4:mini task is dispatched before the llama3.1:8b task
- [ ] VRAM swap events appear in obs.db
- [ ] No OOM error occurs during a model swap sequence under normal operation

## Dependencies
- REQ-002 (scheduler controls dispatch order)
- REQ-011 (observability logs swap events)
