---
title: "Requirement: REQ-012 - mcp-reddit"
summary: "Custom TypeScript MCP server wrapping Reddit OAuth API with rate-limit backoff"
status: "active"
version: "0.1.0"
---
# Requirement: REQ-012 - mcp-reddit

**Skill:** spec-agent
**Feature:** 0000001-autonomous-local-swarm
**Source:** Design — mcp-reddit custom component, Reddit OAuth API, rate-limit backoff
**Priority:** must-have

---

## Functional Requirements
- `mcp-reddit` MUST implement the MCP server protocol using `@modelcontextprotocol/sdk`
- The server MUST authenticate with Reddit using app-only OAuth2 (`client_credentials` grant)
- Credentials MUST be read from `REDDIT_CLIENT_ID` and `REDDIT_CLIENT_SECRET` environment variables only
- The server MUST expose three tools:
  - `reddit_search`: search Reddit posts by query, optional subreddit filter, limit, sort
  - `reddit_get_post`: fetch full post content and top comments by post ID
  - `reddit_hot`: fetch hot posts from a specified subreddit
- On HTTP 429 (rate limit), the server MUST pause the request, read the `x-ratelimit-reset` header, wait the specified duration, then retry once
- If the retry also returns 429, the tool call MUST return an error to the caller
- Credentials MUST NEVER appear in tool responses, error messages, or stderr output
- The server MUST be read-only — no posting, voting, commenting, or moderation actions
- The server MUST start as a standalone Node.js process and communicate via MCP stdio transport

## Acceptance Criteria
- [ ] Server starts and registers 3 tools with MCP client
- [ ] reddit_search returns posts matching query
- [ ] reddit_get_post returns post body and top 10 comments
- [ ] reddit_hot returns top N posts from subreddit
- [ ] 429 response triggers automatic backoff and single retry
- [ ] Second consecutive 429 returns error to caller (no infinite loop)
- [ ] REDDIT_CLIENT_SECRET does not appear in any output or log
- [ ] Server rejects write-operation requests with a clear error
- [ ] Missing credentials cause server to exit with clear error message on startup

## Dependencies
- REQ-009 (orchestrator MCP client must register mcp-reddit as a server)
