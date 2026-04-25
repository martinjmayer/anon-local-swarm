# Risk — mcp-reddit

## Component-scoped risks

| ID | Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| R-mcp-001 | Reddit OAuth token expiry mid-run | Medium | Low | Token refresh on 401 implemented in `RedditClient` |
| R-mcp-002 | Reddit API response schema change breaks parser | Low | Medium | Response parsing isolated in `reddit.ts`; a schema change only requires updating one module |
| R-mcp-003 | Reddit API rate limiting (HTTP 429) | Medium | Low | Single retry with backoff on 429; if still rate-limited, tool returns error and task proceeds with degraded context |
| R-mcp-004 | mcp-reddit process crash leaves orchestrator without the tool | Medium | Low | MCP client detects timeout; tool call returns error; task proceeds without Reddit context |
| R-mcp-005 | Credentials appear in error output | Low | High | `scrubCredentials()` applied to all error messages before return; fail-fast on missing credentials at startup |

## Cross-reference

R-mcp-001 corresponds to R-007 in the system risk register. R-mcp-005 corresponds to R-012.
