---
title: "Requirement: REQ-003 - Ollama Integration and Model Routing"
summary: "Ollama HTTP client, semantic router, and model-to-task-type mapping"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-003 - Ollama Integration and Model Routing

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — Semantic Router, Ollama Client, Intent-Based Routing table
**Priority:** must-have

---

## Functional Requirements
- The orchestrator MUST route tasks to Ollama models according to the following fixed mapping:
  - `DECOMPOSE` → `phi-4:mini`
  - `RESEARCH` → `llama3.1:8b`
  - `GO_CODE` → `qwen2.5-coder:7b`
  - `VIABILITY_REVIEW` → `phi-4:mini`
  - `SUMMARY` → `smollm3:135m`
- The Ollama client MUST call the Ollama HTTP API at `http://localhost:11434`
- Each inference call MUST include the system prompt injected by the orchestrator per task type
- The Ollama client MUST set `keep_alive: 0` on every request to enforce model unloading after use (see REQ-004)
- The router MUST attach relevant MCP tool context to RESEARCH and GO_CODE task prompts
- Model routing MUST be config-driven (not hardcoded in switch statements) to allow future remapping without code changes
- Inference responses MUST be parsed for structured output markers: `NEW_TASK:` prefix for subtask emission, `REVIEW_REQUIRED:` prefix for viability gate trigger

## Acceptance Criteria
- [ ] DECOMPOSE task calls phi-4:mini with correct system prompt
- [ ] RESEARCH task calls llama3.1:8b and includes MCP tool context
- [ ] GO_CODE task calls qwen2.5-coder:7b and includes MCP tool context
- [ ] VIABILITY_REVIEW task calls phi-4:mini and returns a numeric score
- [ ] SUMMARY task calls smollm3:135m and returns compressed context
- [ ] All requests include keep_alive: 0
- [ ] NEW_TASK: prefix in output triggers subtask creation
- [ ] REVIEW_REQUIRED: prefix triggers viability gate
- [ ] Model config can be changed via config file without recompilation

## Dependencies
- REQ-004 (VRAM management is part of every Ollama call)
- REQ-009 (MCP tools attached to RESEARCH / GO_CODE prompts)
