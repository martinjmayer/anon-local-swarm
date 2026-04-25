# Interface Contract — mcp-reddit

## Transport

MCP stdio protocol. The orchestrator spawns `node dist/index.js` and communicates via JSON-RPC over stdin/stdout. stderr is used for diagnostic output only.

## Tools

### `reddit_search`
Search Reddit posts by keyword.

| Parameter | Type | Required | Constraints |
|---|---|---|---|
| `query` | string | ✓ | Search query text |
| `subreddit` | string | — | Restrict to a specific subreddit |
| `limit` | number | — | Max results (default 10, max 100) |
| `sort` | enum | — | `relevance` \| `hot` \| `top` \| `new` \| `comments` (default: `relevance`) |

Returns: JSON array of post objects (title, author, score, url, selftext preview).

### `reddit_get_post`
Fetch full post content and top 10 comments.

| Parameter | Type | Required | Constraints |
|---|---|---|---|
| `post_id` | string | ✓ | Reddit post ID (alphanumeric, max 10 chars) |

Returns: JSON object with post body and top-level comments.

### `reddit_hot`
Get hot posts from a subreddit.

| Parameter | Type | Required | Constraints |
|---|---|---|---|
| `subreddit` | string | ✓ | Subreddit name (without `r/`) |
| `limit` | number | — | Max results (default 10) |

Returns: JSON array of hot post objects.

## Error handling

- Auth errors return `isError: true` with message `Auth error: ...` (credentials scrubbed).
- Rate limit errors return `isError: true` with message `Rate limit error: ...`.
- All error messages are passed through `scrubCredentials()` before return.
- Write-like tool names are rejected defensively with a static error message.

## Credentials

Read from environment at startup:
- `REDDIT_CLIENT_ID` — required
- `REDDIT_CLIENT_SECRET` — required
- `REDDIT_USER_AGENT` — optional (default: `als-swarm/0.1.0`)

Process exits with code 1 if credentials are missing.

## Breaking change policy

Adding new tools is non-breaking. Removing or renaming tools, or changing required parameters, is breaking and requires human approval.
