/**
 * index.ts — MCP server entry point for mcp-reddit.
 *
 * REQ-012: Starts as a Node.js process using stdio MCP transport.
 * Exposes three read-only tools: reddit_search, reddit_get_post, reddit_hot.
 * Credentials are read from environment variables and never echoed.
 */

import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";

import { RedditClient, RedditAuthError, RedditRateLimitError } from "./reddit.js";

// ── Credentials — fail fast if missing (REQ-012) ─────────────────────────────
const clientId = process.env["REDDIT_CLIENT_ID"];
const clientSecret = process.env["REDDIT_CLIENT_SECRET"];

if (!clientId || !clientSecret) {
  process.stderr.write(
    "[mcp-reddit] ERROR: REDDIT_CLIENT_ID and REDDIT_CLIENT_SECRET must be set.\n",
  );
  process.exit(1);
}

const userAgent = process.env["REDDIT_USER_AGENT"] ?? "als-swarm/0.1.0";

const reddit = new RedditClient({ clientId, clientSecret, userAgent });

// ── Tool definitions ──────────────────────────────────────────────────────────
const TOOLS = [
  {
    name: "reddit_search",
    description:
      "Search Reddit posts by keyword query. Optionally filter by subreddit, result limit, and sort order.",
    inputSchema: {
      type: "object" as const,
      properties: {
        query: {
          type: "string",
          description: "Search query string",
        },
        subreddit: {
          type: "string",
          description: "Optional subreddit to restrict the search",
        },
        limit: {
          type: "number",
          description: "Maximum number of posts to return (default 10, max 100)",
          default: 10,
        },
        sort: {
          type: "string",
          enum: ["relevance", "hot", "top", "new", "comments"],
          description: "Sort order for results (default: relevance)",
          default: "relevance",
        },
      },
      required: ["query"],
    },
  },
  {
    name: "reddit_get_post",
    description:
      "Fetch full post content and top 10 comments for a Reddit post by its ID.",
    inputSchema: {
      type: "object" as const,
      properties: {
        post_id: {
          type: "string",
          description: "Reddit post ID (e.g. 'abc123' from reddit.com/r/sub/comments/abc123/...)",
        },
      },
      required: ["post_id"],
    },
  },
  {
    name: "reddit_hot",
    description: "Fetch the current hot posts from a subreddit.",
    inputSchema: {
      type: "object" as const,
      properties: {
        subreddit: {
          type: "string",
          description: "Subreddit name without the r/ prefix",
        },
        limit: {
          type: "number",
          description: "Number of posts to return (default 10, max 100)",
          default: 10,
        },
      },
      required: ["subreddit"],
    },
  },
];

// ── MCP Server ────────────────────────────────────────────────────────────────
const server = new Server(
  { name: "mcp-reddit", version: "0.1.0" },
  { capabilities: { tools: {} } },
);

server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: TOOLS,
}));

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  // REQ-012: server is read-only — reject any write-like tool names defensively.
  const WRITE_KEYWORDS = ["post", "vote", "comment", "delete", "edit", "submit"];
  if (WRITE_KEYWORDS.some((kw) => name.toLowerCase().includes(kw) && name !== "reddit_get_post")) {
    return {
      content: [{ type: "text", text: "Error: mcp-reddit is read-only. Write operations are not supported." }],
      isError: true,
    };
  }

  try {
    switch (name) {
      case "reddit_search": {
        const query = String(args?.["query"] ?? "");
        const subreddit = args?.["subreddit"] ? String(args["subreddit"]) : null;
        const limit = Number(args?.["limit"] ?? 10);
        const sort = String(args?.["sort"] ?? "relevance");
        const result = await reddit.search(query, subreddit, limit, sort);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
        };
      }

      case "reddit_get_post": {
        const postId = String(args?.["post_id"] ?? "");
        if (!postId) {
          return {
            content: [{ type: "text", text: "Error: post_id is required" }],
            isError: true,
          };
        }
        const result = await reddit.getPost(postId);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
        };
      }

      case "reddit_hot": {
        const subreddit = String(args?.["subreddit"] ?? "");
        if (!subreddit) {
          return {
            content: [{ type: "text", text: "Error: subreddit is required" }],
            isError: true,
          };
        }
        const limit = Number(args?.["limit"] ?? 10);
        const result = await reddit.hot(subreddit, limit);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
        };
      }

      default:
        return {
          content: [{ type: "text", text: `Unknown tool: ${name}` }],
          isError: true,
        };
    }
  } catch (err) {
    // REQ-012: credentials must never appear in error output.
    const message = scrubCredentials(
      err instanceof Error ? err.message : String(err),
      [clientId, clientSecret],
    );

    if (err instanceof RedditRateLimitError) {
      return {
        content: [{ type: "text", text: `Rate limit error: ${message}` }],
        isError: true,
      };
    }
    if (err instanceof RedditAuthError) {
      return {
        content: [{ type: "text", text: `Auth error: ${message}` }],
        isError: true,
      };
    }

    return {
      content: [{ type: "text", text: `Error: ${message}` }],
      isError: true,
    };
  }
});

// ── Start ─────────────────────────────────────────────────────────────────────
const transport = new StdioServerTransport();
await server.connect(transport);
process.stderr.write("[mcp-reddit] Server running on stdio\n");

// ── Helpers ───────────────────────────────────────────────────────────────────

/**
 * Replaces any occurrence of credential strings in a message with [REDACTED].
 * REQ-012: credentials must never appear in any output or log.
 */
function scrubCredentials(message: string, credentials: string[]): string {
  let result = message;
  for (const cred of credentials) {
    if (cred && cred.length > 0) {
      result = result.split(cred).join("[REDACTED]");
    }
  }
  return result;
}
