---
title: "Requirement: REQ-009 - MCP Tool Integration"
summary: "MCP client registration, tool dispatch, and graceful degradation"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-009 - MCP Tool Integration

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — MCP tool layer, 10 MCP servers, MCP client component
**Priority:** must-have

---

## Functional Requirements
- The orchestrator MUST embed an MCP client that registers the following servers at startup:
  1. Brave Search (`BRAVE_API_KEY`)
  2. GitHub MCP (`GITHUB_TOKEN`)
  3. context-mode (no credentials)
  4. mcp-server-fetch (no credentials)
  5. Wikipedia MCP (no credentials)
  6. Hacker News MCP (no credentials)
  7. mcp-reddit — custom server (`REDDIT_CLIENT_ID`, `REDDIT_CLIENT_SECRET`)
  8. ArXiv MCP (no credentials)
  9. YouTube Transcript MCP (no credentials)
  10. Memory MCP (no credentials)
- MCP server credentials MUST be read from environment variables only — never hardcoded
- The MCP client MUST make tools available to RESEARCH and GO_CODE task prompts via tool-use context injection
- If an MCP server fails to register at startup, the orchestrator MUST log the failure and continue — tool availability is degraded, not fatal
- If an MCP tool call fails at runtime, the orchestrator MUST log the failure and allow the task to proceed without that tool's output
- The MCP client MUST support adding new servers via config without code changes

## Acceptance Criteria
- [ ] All 10 MCP servers register successfully when credentials are present
- [ ] Missing BRAVE_API_KEY causes Brave Search registration to fail gracefully (logged, not crash)
- [ ] RESEARCH task prompt includes available MCP tool context
- [ ] Failed MCP tool call is logged; task continues and produces output
- [ ] New MCP server can be added via config file without recompilation

## Dependencies
- REQ-003 (MCP tool context injected into Ollama prompts)
- REQ-012 (mcp-reddit custom server must be running)
