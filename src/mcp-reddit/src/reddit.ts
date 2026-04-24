/**
 * reddit.ts — Reddit OAuth2 app-only client with automatic 429 backoff.
 *
 * REQ-012: Authenticates using client_credentials grant; credentials read from
 * environment variables only. On 429, reads x-ratelimit-reset header and retries
 * once. Credentials never appear in responses or error messages.
 */

export interface RedditCredentials {
  clientId: string;
  clientSecret: string;
  userAgent: string;
}

export interface RedditPost {
  id: string;
  title: string;
  selftext: string;
  url: string;
  subreddit: string;
  author: string;
  score: number;
  num_comments: number;
  created_utc: number;
  permalink: string;
}

export interface RedditComment {
  id: string;
  author: string;
  body: string;
  score: number;
  created_utc: number;
}

export interface SearchResult {
  posts: RedditPost[];
}

export interface PostResult {
  post: RedditPost;
  comments: RedditComment[];
}

export interface HotResult {
  posts: RedditPost[];
}

export class RedditRateLimitError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "RedditRateLimitError";
  }
}

export class RedditAuthError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "RedditAuthError";
  }
}

/** Thin HTTP fetch wrapper type so tests can inject a stub. */
export type FetchFn = typeof fetch;

export class RedditClient {
  private accessToken: string | null = null;
  private tokenExpiresAt = 0;

  constructor(
    private readonly creds: RedditCredentials,
    private readonly fetchFn: FetchFn = fetch,
  ) {}

  /** Ensures a valid access token is available, refreshing if needed. */
  private async ensureToken(): Promise<void> {
    if (this.accessToken && Date.now() < this.tokenExpiresAt) {
      return;
    }

    // app-only OAuth2: client_credentials grant
    const encoded = Buffer.from(
      `${this.creds.clientId}:${this.creds.clientSecret}`,
    ).toString("base64");

    const resp = await this.fetchFn(
      "https://www.reddit.com/api/v1/access_token",
      {
        method: "POST",
        headers: {
          Authorization: `Basic ${encoded}`,
          "Content-Type": "application/x-www-form-urlencoded",
          "User-Agent": this.creds.userAgent,
        },
        body: "grant_type=client_credentials",
      },
    );

    if (!resp.ok) {
      // REQ-012: credentials must never appear in error messages
      throw new RedditAuthError(
        `Reddit OAuth failed: HTTP ${resp.status} (credentials redacted)`,
      );
    }

    const data = (await resp.json()) as {
      access_token: string;
      expires_in: number;
    };
    this.accessToken = data.access_token;
    // Expire 60 seconds early to avoid using a stale token
    this.tokenExpiresAt = Date.now() + (data.expires_in - 60) * 1000;
  }

  /**
   * Makes an authenticated GET request to the Reddit OAuth API.
   * REQ-012: On 429, reads x-ratelimit-reset and retries once.
   */
  private async get<T>(path: string): Promise<T> {
    await this.ensureToken();
    const url = `https://oauth.reddit.com${path}`;
    const result = await this.doGet<T>(url, false);
    return result;
  }

  private async doGet<T>(url: string, isRetry: boolean): Promise<T> {
    const resp = await this.fetchFn(url, {
      headers: {
        Authorization: `Bearer ${this.accessToken}`,
        "User-Agent": this.creds.userAgent,
      },
    });

    if (resp.status === 429) {
      if (isRetry) {
        // Second 429 — return error to caller, no infinite loop (REQ-012)
        throw new RedditRateLimitError(
          "Reddit rate limit exceeded after retry",
        );
      }
      // First 429 — read reset header and wait (REQ-012)
      const resetHeader = resp.headers.get("x-ratelimit-reset");
      const waitMs = resetHeader ? parseFloat(resetHeader) * 1000 : 60_000;
      await sleep(waitMs);
      // Refresh token in case it expired during wait
      this.accessToken = null;
      await this.ensureToken();
      return this.doGet<T>(url, true);
    }

    if (!resp.ok) {
      throw new Error(`Reddit API error: HTTP ${resp.status}`);
    }

    return resp.json() as Promise<T>;
  }

  /**
   * search — REQ-012: reddit_search tool.
   * Searches Reddit posts by query with optional subreddit filter.
   */
  async search(
    query: string,
    subreddit: string | null,
    limit: number,
    sort: string,
  ): Promise<SearchResult> {
    const sub = subreddit ? `/r/${subreddit}` : "";
    const params = new URLSearchParams({
      q: query,
      limit: String(Math.min(limit, 100)),
      sort,
      type: "link",
    });
    const path = `${sub}/search.json?${params}`;
    const data = await this.get<RedditListingResponse>(path);
    return { posts: data.data.children.map((c) => normalisePost(c.data)) };
  }

  /**
   * getPost — REQ-012: reddit_get_post tool.
   * Fetches full post content and top 10 comments by post ID.
   */
  async getPost(postId: string): Promise<PostResult> {
    // Reddit API requires /comments/<id>.json
    const path = `/comments/${postId}.json?limit=10&depth=1`;
    const data = await this.get<[RedditListingResponse, RedditListingResponse]>(
      path,
    );

    const [postListing, commentListing] = data;
    const postData = postListing.data.children[0]?.data;
    if (!postData) {
      throw new Error(`Post ${postId} not found`);
    }

    const comments: RedditComment[] = commentListing.data.children
      .filter((c) => c.kind === "t1")
      .slice(0, 10)
      .map((c) => ({
        id: c.data.id as string,
        author: c.data.author as string,
        body: c.data.body as string,
        score: c.data.score as number,
        created_utc: c.data.created_utc as number,
      }));

    return { post: normalisePost(postData), comments };
  }

  /**
   * hot — REQ-012: reddit_hot tool.
   * Fetches hot posts from a subreddit.
   */
  async hot(subreddit: string, limit: number): Promise<HotResult> {
    const params = new URLSearchParams({ limit: String(Math.min(limit, 100)) });
    const path = `/r/${subreddit}/hot.json?${params}`;
    const data = await this.get<RedditListingResponse>(path);
    return { posts: data.data.children.map((c) => normalisePost(c.data)) };
  }
}

// ── Internal helpers ──────────────────────────────────────────────────────────

interface RedditListingResponse {
  data: {
    children: Array<{ kind: string; data: Record<string, unknown> }>;
  };
}

function normalisePost(raw: Record<string, unknown>): RedditPost {
  return {
    id: String(raw["id"] ?? ""),
    title: String(raw["title"] ?? ""),
    selftext: String(raw["selftext"] ?? ""),
    url: String(raw["url"] ?? ""),
    subreddit: String(raw["subreddit"] ?? ""),
    author: String(raw["author"] ?? ""),
    score: Number(raw["score"] ?? 0),
    num_comments: Number(raw["num_comments"] ?? 0),
    created_utc: Number(raw["created_utc"] ?? 0),
    permalink: String(raw["permalink"] ?? ""),
  };
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
