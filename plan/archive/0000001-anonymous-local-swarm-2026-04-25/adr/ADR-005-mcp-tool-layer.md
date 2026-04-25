---
title: "ADR-005: MCP as the Tool Integration Layer"
summary: "The Model Context Protocol is adopted as the tool integration layer, replacing direct API clients for all external capabilities."
status: "accepted"
version: "0.1.0"
---
# ADR-005 - MCP as the Tool Integration Layer

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** orchestrator
**Status:** accepted
**Date:** 2026-04-22

---

## Context

The swarm's RESEARCH and GO_CODE tasks need access to external capabilities: web search, GitHub code retrieval, URL fetching, Reddit, ArXiv, Wikipedia, Hacker News, YouTube transcripts, and a persistent knowledge graph. Two architectural approaches were considered: direct Go HTTP clients per capability, or a pluggable tool layer using a standard protocol. The choice affects how the capabilities are called from Ollama prompts, how new tools are added, and whether the tool ecosystem is reusable across future swarm projects.

---

## Decision

Use **MCP (Model Context Protocol)** as the tool integration layer. The orchestrator embeds an MCP client; each external capability is an MCP server process. Tools are registered at startup via config and called by the orchestrator on behalf of task prompts.

MCP is an open standard with a growing ecosystem of pre-built servers (Brave Search, GitHub, Wikipedia, HN, ArXiv, YouTube, Memory, fetch). Adopting MCP means the orchestrator needs only one integration pattern — the MCP client protocol — rather than ten different API clients. New tools are added by registering a new MCP server in config, with no orchestrator code changes. The tool call/response model maps cleanly to the prompt injection pattern: the orchestrator injects available tool schemas into the model prompt, the model issues a tool call, the orchestrator executes it via MCP and returns the result.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| Direct Go HTTP clients per capability | Simple, no protocol overhead, full control | Ten different API integrations to maintain; adding a new tool requires code changes; no standard interface | Not pluggable; high maintenance surface |
| Function calling via Ollama only | Model-native tool calling | 3-7B models have inconsistent function calling reliability; ties tool dispatch to model capability | Model quality risk for tool selection; not all models support function calling reliably |
| LangChain/LlamaIndex agent framework | Rich tooling ecosystem, built-in tool abstractions | Python runtime dependency; heavy framework overhead; couples orchestrator to a specific framework | Language mismatch; framework lock-in |

---

## Affected Components

| Component | Impact |
|---|---|
| orchestrator | Embeds MCP client; registers all tool servers at startup |
| mcp-reddit | Custom MCP server — must implement MCP server protocol (see ADR-006) |
| router | Tool availability injected into RESEARCH and GO_CODE prompts |

---

## Consequences

**Positive:**
- Pluggable — new tools added via config with no orchestrator code changes
- Pre-built MCP servers available for most required capabilities (9 of 10)
- Standard protocol — tool servers can be swapped independently
- LLM token cost is identical to direct API calls (tool results fed back as context, not protocol overhead)

**Negative:**
- MCP server processes must be running before orchestrator startup — adds operational complexity
- MCP protocol adds a process boundary between orchestrator and tools (latency negligible at localhost)
- Debugging tool failures requires understanding two processes (orchestrator + MCP server)

**Risks:**
- MCP ecosystem servers are community-maintained; quality and reliability varies; mitigated by graceful degradation (task proceeds if a tool fails)

---

## Related ADRs

- ADR-004 - related-to (Ollama prompts receive MCP tool context)
- ADR-006 - depends-on (mcp-reddit must implement MCP protocol)

---
