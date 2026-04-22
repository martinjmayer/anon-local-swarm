# Design - reddit-mcp

## Component
- Purpose: TypeScript MCP server wrapping the Reddit OAuth API; exposes search, post, and comment retrieval tools with automatic rate-limit backoff
- Parent feature: 0000001-autonomous-local-swarm
- Component ID: reddit-mcp
- Adoption mode: greenfield

## Product Layer
- User stories confirmed: 1
  - As the ALS orchestrator, I want to query Reddit via MCP tools so that RESEARCH tasks can retrieve community discussions, niche signals, and trend data.
- Acceptance criteria:
  - MCP server registers and responds to tool calls from the orchestrator's MCP client
  - Search tool returns posts for a given query and subreddit filter
  - Post tool returns full post content + top comments by post ID
  - On HTTP 429, server pauses the request, waits for the `x-ratelimit-reset` header duration, then retries transparently — caller receives the result, not the error
  - Credentials are read from env vars only; never logged or returned in tool output
- Constraints:
  - Reddit OAuth API only — no scraping
  - Max 1 retry per rate-limited request (pause → retry → fail if still 429)
  - Credentials: `REDDIT_CLIENT_ID`, `REDDIT_CLIENT_SECRET` (app-only OAuth, no user login required)

## Architecture Layer
- Latency: best-effort; rate-limit pause is transparent to caller
- Security: credentials in env vars only; never in logs, tool output, or error messages
- Observability: stderr logging for rate-limit events and errors; MCP protocol handles tool response

## Engineering Layer
- Language: TypeScript
- Runtime: Node.js 20+
- MCP SDK: `@modelcontextprotocol/sdk`
- HTTP client: `node-fetch` or native `fetch` (Node 20)
- Auth: Reddit app-only OAuth2 (`client_credentials` grant)
- Entry point: `src/index.ts`
- Build: `tsc` → `dist/`
- Run: `node dist/index.js` (registered as MCP server process)

## Tools Exposed
| Tool name | Description | Key inputs |
|---|---|---|
| `reddit_search` | Search Reddit posts | `query`, `subreddit?`, `limit?`, `sort?` |
| `reddit_get_post` | Fetch post + top comments by ID | `post_id`, `comment_limit?` |
| `reddit_hot` | Hot posts from a subreddit | `subreddit`, `limit?` |

## Scope
- In: search, post retrieval, hot feed, rate-limit backoff, app-only OAuth
- Out: user authentication, posting/voting/commenting (read-only), pagination beyond single response
- Deferred: subreddit metadata, user profile lookup

## Risks
- Reddit API changes OAuth token format — likelihood: low; impact: medium; mitigated by isolating auth in a single module
- Rate limits tightened further — likelihood: low; impact: low; backoff handler already covers this

## Confirmation
Human confirmed this design before proceeding: no
Date confirmed: —
