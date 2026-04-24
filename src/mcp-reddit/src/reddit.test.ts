/**
 * reddit.test.ts — Unit tests for the RedditClient.
 *
 * All tests use a stub fetch function — no real network calls.
 */

import { RedditClient, RedditRateLimitError } from "./reddit.js";

const CREDS = {
  clientId: "test-client-id",
  clientSecret: "test-client-secret",
  userAgent: "als-test/0.1.0",
};

/** Builds a stub fetch that returns the given responses in order. */
function stubFetch(responses: Array<{ status: number; body: unknown; headers?: Record<string, string> }>): typeof fetch {
  let callIndex = 0;
  return async (_url: RequestInfo | URL, _init?: RequestInit): Promise<Response> => {
    const resp = responses[callIndex++] ?? { status: 200, body: {} };
    return {
      ok: resp.status >= 200 && resp.status < 300,
      status: resp.status,
      headers: {
        get: (name: string) => resp.headers?.[name] ?? null,
      } as unknown as Headers,
      json: async () => resp.body,
    } as unknown as Response;
  };
}

/** Standard OAuth token response. */
const TOKEN_RESP = {
  status: 200,
  body: { access_token: "test-token", expires_in: 3600 },
};

// req-012: reddit_search returns posts matching query.
test("req-012: search returns posts", async () => {
  const fetch = stubFetch([
    TOKEN_RESP,
    {
      status: 200,
      body: {
        data: {
          children: [
            {
              kind: "t3",
              data: {
                id: "abc123",
                title: "Test post",
                selftext: "body text",
                url: "https://reddit.com/r/test/abc123",
                subreddit: "test",
                author: "user1",
                score: 42,
                num_comments: 5,
                created_utc: 1700000000,
                permalink: "/r/test/comments/abc123/test_post/",
              },
            },
          ],
        },
      },
    },
  ]);

  const client = new RedditClient(CREDS, fetch);
  const result = await client.search("test query", null, 10, "relevance");

  expect(result.posts).toHaveLength(1);
  expect(result.posts[0]!.id).toBe("abc123");
  expect(result.posts[0]!.title).toBe("Test post");
});

// req-012: reddit_get_post returns post and top comments.
test("req-012: getPost returns post and comments", async () => {
  const fetch = stubFetch([
    TOKEN_RESP,
    {
      status: 200,
      body: [
        {
          data: {
            children: [
              {
                kind: "t3",
                data: {
                  id: "post1",
                  title: "Post title",
                  selftext: "Post body",
                  url: "https://reddit.com/r/sub/post1",
                  subreddit: "sub",
                  author: "author1",
                  score: 100,
                  num_comments: 10,
                  created_utc: 1700000000,
                  permalink: "/r/sub/comments/post1/",
                },
              },
            ],
          },
        },
        {
          data: {
            children: [
              {
                kind: "t1",
                data: {
                  id: "comment1",
                  author: "commenter",
                  body: "Great post!",
                  score: 10,
                  created_utc: 1700000100,
                },
              },
            ],
          },
        },
      ],
    },
  ]);

  const client = new RedditClient(CREDS, fetch);
  const result = await client.getPost("post1");

  expect(result.post.id).toBe("post1");
  expect(result.comments).toHaveLength(1);
  expect(result.comments[0]!.body).toBe("Great post!");
});

// req-012: reddit_hot returns hot posts from subreddit.
test("req-012: hot returns posts from subreddit", async () => {
  const fetch = stubFetch([
    TOKEN_RESP,
    {
      status: 200,
      body: {
        data: {
          children: [
            {
              kind: "t3",
              data: {
                id: "hot1",
                title: "Hot post",
                selftext: "",
                url: "https://i.redd.it/image.jpg",
                subreddit: "programming",
                author: "dev",
                score: 500,
                num_comments: 30,
                created_utc: 1700001000,
                permalink: "/r/programming/comments/hot1/",
              },
            },
          ],
        },
      },
    },
  ]);

  const client = new RedditClient(CREDS, fetch);
  const result = await client.hot("programming", 5);

  expect(result.posts).toHaveLength(1);
  expect(result.posts[0]!.subreddit).toBe("programming");
});

// req-012: 429 triggers backoff and single retry.
test("req-012: 429 triggers backoff and single retry", async () => {
  let callCount = 0;
  const mockFetch = async (_url: RequestInfo | URL, _init?: RequestInit): Promise<Response> => {
    callCount++;
    if (callCount === 1) {
      // First call: token
      return {
        ok: true,
        status: 200,
        headers: { get: () => null } as unknown as Headers,
        json: async () => ({ access_token: "tok", expires_in: 3600 }),
      } as unknown as Response;
    }
    if (callCount === 2) {
      // Second call: API — 429 with reset header of 0.01s
      return {
        ok: false,
        status: 429,
        headers: {
          get: (name: string) => name === "x-ratelimit-reset" ? "0.01" : null,
        } as unknown as Headers,
        json: async () => ({}),
      } as unknown as Response;
    }
    if (callCount === 3) {
      // Third call: re-auth after backoff
      return {
        ok: true,
        status: 200,
        headers: { get: () => null } as unknown as Headers,
        json: async () => ({ access_token: "tok2", expires_in: 3600 }),
      } as unknown as Response;
    }
    // Fourth call: retry succeeds
    return {
      ok: true,
      status: 200,
      headers: { get: () => null } as unknown as Headers,
      json: async () => ({
        data: { children: [] },
      }),
    } as unknown as Response;
  };

  const client = new RedditClient(CREDS, mockFetch);
  const result = await client.search("test", null, 5, "relevance");
  expect(result.posts).toHaveLength(0);
  expect(callCount).toBeGreaterThanOrEqual(3); // token + 429 + retry
});

// req-012: Second consecutive 429 returns error — no infinite loop.
test("req-012: second 429 throws RedditRateLimitError", async () => {
  let callCount = 0;
  const mockFetch = async (_url: RequestInfo | URL, _init?: RequestInit): Promise<Response> => {
    callCount++;
    if (callCount === 1 || callCount === 3) {
      // Token responses
      return {
        ok: true,
        status: 200,
        headers: { get: () => null } as unknown as Headers,
        json: async () => ({ access_token: "tok", expires_in: 3600 }),
      } as unknown as Response;
    }
    // Both API calls return 429
    return {
      ok: false,
      status: 429,
      headers: {
        get: (name: string) => name === "x-ratelimit-reset" ? "0.01" : null,
      } as unknown as Headers,
      json: async () => ({}),
    } as unknown as Response;
  };

  const client = new RedditClient(CREDS, mockFetch);
  await expect(client.search("test", null, 5, "relevance")).rejects.toThrow(
    RedditRateLimitError,
  );
});

// req-012: REDDIT_CLIENT_SECRET must not appear in error messages.
test("req-012: credentials do not appear in auth error message", async () => {
  const mockFetch = async (): Promise<Response> =>
    ({
      ok: false,
      status: 401,
      headers: { get: () => null } as unknown as Headers,
      json: async () => ({}),
    }) as unknown as Response;

  const client = new RedditClient(CREDS, mockFetch);
  try {
    await client.search("anything", null, 1, "relevance");
    fail("Expected error");
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    expect(msg).not.toContain("test-client-secret");
    expect(msg).not.toContain("test-client-id");
  }
});

// req-012: search respects limit cap of 100.
test("req-012: search caps limit at 100", async () => {
  let capturedUrl = "";
  const mockFetch = async (url: RequestInfo | URL, _init?: RequestInit): Promise<Response> => {
    capturedUrl = url.toString();
    if (capturedUrl.includes("access_token")) {
      return {
        ok: true, status: 200,
        headers: { get: () => null } as unknown as Headers,
        json: async () => ({ access_token: "tok", expires_in: 3600 }),
      } as unknown as Response;
    }
    return {
      ok: true, status: 200,
      headers: { get: () => null } as unknown as Headers,
      json: async () => ({ data: { children: [] } }),
    } as unknown as Response;
  };

  const client = new RedditClient(CREDS, mockFetch);
  await client.search("test", null, 500, "relevance");
  expect(capturedUrl).toContain("limit=100");
});
