# Dependencies — mcp-reddit

## Runtime dependencies

| Dependency | Version | Purpose |
|---|---|---|
| `@modelcontextprotocol/sdk` | latest | MCP server and stdio transport |
| Node.js | 20 LTS | Runtime |
| Reddit OAuth2 API | — | `https://oauth.reddit.com` — data source |

## Build dependencies

| Dependency | Purpose |
|---|---|
| TypeScript | Compilation (`tsc`) |

## What mcp-reddit consumes

- Reddit REST API (HTTPS, outbound only)
- Environment variables: `REDDIT_CLIENT_ID`, `REDDIT_CLIENT_SECRET`, `REDDIT_USER_AGENT`

## What depends on mcp-reddit

| Component | How |
|---|---|
| orchestrator | Spawns as subprocess via `internal/mcp`; calls tools over MCP stdio |

## Dependency direction

```
orchestrator → mcp-reddit (subprocess)
mcp-reddit → Reddit API (HTTPS)
```

## Lockfile note

`package-lock.json` should be committed to pin transitive dependencies. See security report finding #4.
