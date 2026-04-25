---
title: "ADR-006: TypeScript for mcp-reddit"
summary: "TypeScript/Node is chosen for the custom mcp-reddit server, creating a deliberate polyglot split from the Go orchestrator."
status: "accepted"
version: "0.1.0"
---
# ADR-006 - TypeScript for mcp-reddit

**Skill:** adr-agent
**Tool:** claude-code
**Model:** claude-sonnet-4-5
**Feature:** 0000001-autonomous-local-swarm
**Component:** mcp-reddit
**Status:** accepted
**Date:** 2026-04-22

---

## Context

The custom Reddit MCP server must implement the MCP server protocol. A language choice was needed. The orchestrator is Go; the MCP ecosystem is primarily TypeScript/Node. The `@modelcontextprotocol/sdk` npm package is the reference implementation of the MCP server protocol — mature, well-documented, and used by all Anthropic-published MCP servers. A Go MCP server SDK exists but is less mature.

---

## Decision

Use **TypeScript/Node** for `mcp-reddit`, diverging from the Go orchestrator.

The `@modelcontextprotocol/sdk` TypeScript package is the canonical MCP server implementation. Using it means the mcp-reddit server follows identical patterns to all other MCP servers in the ecosystem (Brave, GitHub, etc.), making it easier to maintain, contribute to, and audit. The Reddit OAuth flow and HTTP client code is straightforward TypeScript. The operational model is simple: `node dist/index.js` started as a separate process, registered by the Go orchestrator over MCP stdio transport — no cross-language calls, just process boundaries.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|---|---|---|---|
| Go (`mcp-go` or hand-rolled) | Single language across all components | `mcp-go` is less mature; would require hand-rolling MCP protocol handling; maintenance burden | Protocol implementation risk; ecosystem mismatch |
| Python | Simple HTTP client code; good Reddit library (`praw`) | Third runtime (Python + Node already excluded); adds another dependency | Unnecessary third language |
| Go with TypeScript SDK via subprocess | Stays in Go | Deeply wrong architecture — calling TypeScript from Go via subprocess is worse than just using TypeScript | Anti-pattern |

---

## Affected Components

| Component | Impact |
|---|---|
| mcp-reddit | Written in TypeScript; requires Node.js 20+ at runtime |
| orchestrator | Starts mcp-reddit as a subprocess; communicates via MCP stdio — no language coupling |

---

## Consequences

**Positive:**
- Canonical MCP server SDK (`@modelcontextprotocol/sdk`) — well-tested, consistent with ecosystem
- Operator can understand mcp-reddit by comparing it to any other MCP server
- TypeScript type safety catches Reddit API response shape issues at compile time

**Negative:**
- Polyglot project: operator needs both Go toolchain and Node.js 20+ installed
- Two build systems (Go modules + npm/tsc) to understand and maintain
- Two separate test suites

**Risks:**
- Node.js version mismatch on operator machine — mitigated by `.nvmrc` or `engines` field in `package.json`

---

## Related ADRs

- ADR-001 - related-to (Go choice creates the polyglot split this ADR explains)
- ADR-005 - depends-on (MCP protocol is the reason TypeScript is justified here)

---
