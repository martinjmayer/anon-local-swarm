# Purpose — mcp-reddit

## What this component exists to do

`mcp-reddit` is a read-only MCP server that gives the swarm's LLM agents access to Reddit content. It wraps the Reddit OAuth2 API and exposes three tools via the MCP stdio protocol: search posts, fetch a single post with comments, and retrieve hot posts from a subreddit.

Its single job is to provide grounded, current information from Reddit to RESEARCH tasks, enabling the swarm to validate hypotheses, gauge community interest, and find prior art — without the orchestrator needing to know anything about Reddit's API.

## Why it exists as a separate component

Per ADR-006, the Reddit integration is isolated in a TypeScript MCP server for three reasons:

1. **Protocol isolation** — MCP servers communicate with the orchestrator over stdio; a crash or hang in mcp-reddit does not crash the orchestrator.
2. **Credential isolation** — Reddit OAuth credentials live only in this process's environment; they never appear in Go code, swarm.db, or obs.db.
3. **Language fit** — The MCP TypeScript SDK is the reference implementation; using it here gives the most stable and spec-compliant behaviour.

## Who uses it

- **orchestrator** — spawns mcp-reddit as a subprocess and calls its tools during RESEARCH task dispatch.
- **Operator** — configures it via `config.yaml` MCP server registration and `.env` credentials.
