# Test Coverage — mcp-reddit

## Summary

| Type | Coverage | Notes |
|---|---|---|
| Unit | ~60% | `reddit.test.ts` covers RedditClient HTTP interactions via mock server |
| Integration | 0% | No live Reddit API tests |
| E2E | 0% | No end-to-end tests |

## What is tested

- `RedditClient` authentication and token acquisition
- `reddit_search` request shape and response parsing
- `reddit_get_post` request and response parsing
- `scrubCredentials` — credentials redacted from error output
- HTTP 429 retry behaviour

## What is not tested

- `reddit_hot` tool (not covered in current test suite — gap)
- MCP stdio transport layer
- Write guard (keyword rejection)
- Credential absence → process.exit(1) path

## Test framework

Vitest (`src/mcp-reddit/src/reddit.test.ts`)

## Note on Docker build

The TypeScript tests are **not run as part of the Docker build** (`npm run build` compiles only). Tests must be run separately on the host:

```bash
cd src/mcp-reddit
npm install
npm test
```
