---
title: "ADR-007: Agent Skills Format for System Prompts"
summary: "The agentskills.io SKILL.md format is adopted for all task-type system prompts, making prompts version-controlled, portable, first-class artifacts."
status: "accepted"
version: "0.1.0"
---
# ADR-007 - Agent Skills Format for System Prompts

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator
**Status:** accepted
**Date:** 2026-04-22

---

## Context

The swarm injects a system prompt into every Ollama inference call. Each of the five task types (DECOMPOSE, RESEARCH, GO_CODE, VIABILITY_REVIEW, SUMMARY) needs a distinct system prompt tuned to that model and task. A storage and management strategy was needed for these prompts. Options ranged from hardcoded strings in Go, to config file values, to structured files. System prompts are the primary lever for SLM output quality (risk R-001) — they are high-value artifacts that need to be editable, versioned, and reviewable without code changes.

---

## Decision

Adopt the **Agent Skills format** ([agentskills.io](https://agentskills.io)) for all task-type system prompts. Each task type has a `SKILL.md` file in `skills/{task-type}/` with YAML frontmatter (`name`, `description`) and a Markdown body containing the full system prompt instructions.

The orchestrator implements a skill-loader that: reads all `skills/*/SKILL.md` files at startup, indexes `name` + `description` (~100 tokens each) for routing metadata, and loads the full `SKILL.md` body as the Ollama system prompt at task dispatch. This is the progressive disclosure pattern from the Agent Skills spec — cheap at startup, full context loaded only when needed.

Two official Anthropic skills (`mcp-builder`, `webapp-testing`) are bundled as reference material for GO_CODE tasks.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| Hardcoded Go strings | Simple, no extra files | Prompts require code changes and recompilation to iterate; not reviewable by non-developers | Friction for prompt engineering iteration |
| Config file (YAML/TOML) | Simple, no new format | Multiline prompts are awkward in YAML/TOML; no standard structure; not portable | Poor ergonomics for long Markdown-formatted prompts |
| Plain .md files (no frontmatter) | Simple | No standard metadata; no ecosystem; not portable across agent tools | No standard; would invent a custom format without benefit |
| Agent Skills (chosen) | Open standard, progressive disclosure, portable, version-controlled, Markdown body | Adds a skill-loader to the orchestrator (trivial implementation) | — |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | Implements skill-loader; reads skills/ at startup; injects SKILL.md body as system prompt |
| router | Uses skill name + description for routing metadata (matches task type to skill) |

---

## Consequences

**Positive:**
- System prompts are first-class version-controlled artifacts — editable without touching Go code
- Standard format — prompts portable across any Agent Skills-compatible tool
- Progressive disclosure aligns with context efficiency goals
- Official Anthropic skills can be bundled directly from the public repo
- Prompt quality iteration is decoupled from the build/deploy cycle

**Negative:**
- Skill-loader adds a small amount of Go code to the orchestrator
- The agentskills.io standard targets AI coding agents (Claude Code, Cursor), not Go→Ollama pipelines; adoption is pattern-based, not native integration
- Routing in the swarm is deterministic (fixed task type → model mapping); the `description` field serves as documentation, not active routing logic

**Risks:**
- Agent Skills spec changes could diverge from swarm's usage — mitigated by the format being simple (SKILL.md + frontmatter) and unlikely to change in a breaking way

---

## Related ADRs

- ADR-004 - depends-on (SKILL.md body injected as Ollama system prompt)

---
