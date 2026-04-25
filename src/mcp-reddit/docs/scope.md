# Scope — mcp-reddit

## In scope

- `reddit_search` tool — keyword search across Reddit
- `reddit_get_post` tool — full post + top comments by ID
- `reddit_hot` tool — hot posts from a subreddit
- App-only OAuth2 client_credentials authentication
- HTTP 429 rate-limit backoff (single retry)
- Credential isolation — credentials never echoed in tool output or errors
- Defensive write guard — rejects write-like tool names

## Out of scope

- Write operations (posting, voting, commenting, editing, deleting)
- User OAuth (3-legged flow)
- Response caching
- Multi-page pagination (single response per tool call)
- Subreddit metadata tools
- User profile lookup

## Deferred

- `reddit_subreddit_info` tool
- `reddit_user_profile` tool
- Response caching with TTL
- Multi-page result support
