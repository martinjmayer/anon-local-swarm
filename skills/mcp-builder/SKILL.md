---
name: mcp-builder
description: Best practices for building MCP servers — used by GO_CODE tasks that generate MCP server output
bundle: anthropic
source: https://github.com/anthropics/skills/tree/main/skills/mcp-builder
task_types: [GO_CODE]
---

# MCP Builder

Reference skill for GO_CODE tasks that generate MCP server implementations.

## When this skill applies

Load this skill alongside `als-go-code` when the GO_CODE task payload references:
- Building an MCP server
- Implementing MCP tools or resources
- Wrapping an API as an MCP server

## MCP server anatomy

An MCP server implements the [Model Context Protocol](https://modelcontextprotocol.io/):

```
Server
├── Tools        — functions the model can call (tools/list, tools/call)
├── Resources    — data the model can read (resources/list, resources/read)
└── Prompts      — reusable prompt templates (prompts/list, prompts/get)
```

## Tool design principles

1. **One action per tool** — tools should do one thing and return structured output.
2. **Descriptive schemas** — every parameter needs a `description` field; models use these to decide how to call the tool.
3. **Return text content** — always return `{ content: [{ type: "text", text: "..." }] }`.
4. **Signal errors clearly** — return `isError: true` with a human-readable message on failure.
5. **No credentials in responses** — scrub API keys from all output and error messages.
6. **Read-only by default** — only expose write operations if explicitly required.

## stdio transport (Node.js)

```typescript
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

const server = new Server({ name: "my-server", version: "0.1.0" }, { capabilities: { tools: {} } });
const transport = new StdioServerTransport();
await server.connect(transport);
```

## Rate limiting pattern

```typescript
async function withRateLimit<T>(fn: () => Promise<T>, retryAfterMs: number): Promise<T> {
  try {
    return await fn();
  } catch (err) {
    if (isRateLimitError(err)) {
      await sleep(retryAfterMs);
      return fn(); // single retry
    }
    throw err;
  }
}
```

## Checklist before emitting an MCP server

- [ ] All tool `inputSchema` properties have `description` fields
- [ ] Credentials are read from environment variables only
- [ ] Credentials never appear in tool responses or error messages
- [ ] Server exits with a non-zero code and clear message if required env vars are missing
- [ ] All tools are read-only unless the spec explicitly requires writes
- [ ] Server handles disconnection gracefully (process exits on stdin close)
