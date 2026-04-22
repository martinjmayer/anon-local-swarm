# MCP Servers - Autonomous Local Swarm

**Feature:** 0000001-autonomous-local-swarm
**Version:** 0.1.0

> Reference document for all MCP servers registered by the orchestrator at startup. Each entry covers what the server does, how the swarm uses it, which task types invoke it, and what credentials are required.

---

## 1. Brave Search

**Source:** [brave/brave-search-mcp-server](https://github.com/brave/brave-search-mcp-server)
**Credentials:** `BRAVE_API_KEY` (free tier: 2,000 queries/month)

### What it does
Full web search via Brave's independent search index. Returns ranked web results with titles, URLs, and snippets.

### How the swarm uses it
Primary web search tool for `RESEARCH` tasks. When the orchestrator dispatches a RESEARCH task, the MCP client injects available Brave Search tools into the Ollama prompt. The model issues search queries as tool calls; results are returned as context for the model to synthesise.

**Primary task types:** `RESEARCH`
**Example invocations:** market sizing queries, competitor research, niche validation, price discovery, technology landscape surveys.

---

## 2. GitHub MCP

**Source:** [github/github-mcp-server](https://github.com/github/github-mcp-server)
**Credentials:** `GITHUB_TOKEN` (free, personal access token)

### What it does
Full GitHub API access: repository search, file reading, code search, issue and PR browsing, release listing.

### How the swarm uses it
Used by `RESEARCH` tasks to discover existing open-source projects, assess competition, and pull implementation references. Used by `GO_CODE` tasks to pull example code, check library APIs, and read documentation from source repositories.

**Primary task types:** `RESEARCH`, `GO_CODE`
**Example invocations:** searching for existing Go CLI tools in a niche, reading a library's source to understand its API, checking if a project already solves the target problem.

---

## 3. context-mode

**Source:** [mksglu/context-mode](https://github.com/mksglu/context-mode)
**Credentials:** none

### What it does
Context window optimisation: sandboxes large tool outputs, indexes content into a searchable knowledge base, and provides `ctx_search` for retrieving relevant chunks without flooding the model's context window.

### How the swarm uses it
Used across all task types to manage context efficiently. When a RESEARCH task pulls a large web page or document, context-mode indexes it and the model queries for relevant sections rather than consuming the full text. Particularly valuable for `SUMMARY` tasks — the model can search indexed prior context rather than re-reading everything.

**Primary task types:** `RESEARCH`, `DECOMPOSE`, `SUMMARY`
**Example invocations:** indexing a long web article then searching for the pricing section; retrieving only the relevant Go package documentation from a previously fetched repository.

---

## 4. mcp-server-fetch

**Source:** Standard MCP fetch server
**Credentials:** none

### What it does
Fetches arbitrary URLs and returns their content as plain text or markdown. Handles static HTML pages that don't require JavaScript rendering.

### How the swarm uses it
Fallback for RESEARCH tasks when Brave Search returns a URL but the model needs the full page content. Also used for pulling raw documentation, API references, pricing pages, and any URL that requires full-text reading.

**Primary task types:** `RESEARCH`, `GO_CODE`
**Example invocations:** fetching a SaaS pricing page to assess market rates; pulling a Go package's README from pkg.go.dev; retrieving a blog post referenced in a Brave Search result.

---

## 5. Wikipedia MCP

**Source:** Standard Wikipedia MCP server
**Credentials:** none

### What it does
Queries Wikipedia for structured encyclopaedic content. Returns article summaries, full article text, and related article links.

### How the swarm uses it
Used by `RESEARCH` and `DECOMPOSE` tasks to establish domain definitions, understand technical concepts, and build foundational knowledge about a target niche before deeper research begins. Wikipedia provides reliable, stable background context that anchors the model's understanding.

**Primary task types:** `DECOMPOSE`, `RESEARCH`
**Example invocations:** looking up the XBRL standard before decomposing a financial data task; understanding the legal niche landscape before planning research tasks; defining domain terminology.

---

## 6. Hacker News MCP

**Source:** Standard HN MCP server
**Credentials:** none

### What it does
Searches and retrieves Hacker News posts, comments, and Ask HN threads. Access to the HN Algolia search API.

### How the swarm uses it
Used by `RESEARCH` tasks for market signal and trend discovery. HN surfaces early-adopter opinions, developer pain points, and niche tool discussions that don't appear prominently in standard web search. Particularly useful for validating whether a problem is real and whether existing solutions are considered adequate.

**Primary task types:** `RESEARCH`, `VIABILITY_REVIEW`
**Example invocations:** searching "Ask HN: tools for X" to find underserved needs; finding discussions where developers complain about existing solutions; checking reception of competitor products.

---

## 7. mcp-reddit *(custom — built in this project)*

**Source:** `src/mcp-reddit/` — custom TypeScript MCP server
**Credentials:** `REDDIT_CLIENT_ID`, `REDDIT_CLIENT_SECRET` (free Reddit app registration)

### What it does
Wraps the Reddit OAuth API with three read-only tools: `reddit_search`, `reddit_get_post`, `reddit_hot`. Automatically pauses on HTTP 429 and retries once using the `x-ratelimit-reset` header.

### How the swarm uses it
Used by `RESEARCH` tasks to access community discussions, niche subreddits, and real-user feedback. Reddit is often more candid than formal web sources — users describe actual pain points, workarounds, and willingness to pay. Complements HN (developer-focused) with broader consumer and professional community signals.

**Primary task types:** `RESEARCH`
**Example invocations:** searching r/legaltech for complaints about existing tools; reading r/accounting for manual workflow discussions; finding niche communities to validate demand before VIABILITY_REVIEW.

---

## 8. ArXiv MCP

**Source:** Standard ArXiv MCP server
**Credentials:** none

### What it does
Searches ArXiv for academic papers by query, author, or category. Returns abstracts, metadata, and links to full PDFs.

### How the swarm uses it
Used by `RESEARCH` tasks when the target domain has an academic dimension — ML techniques, data formats, domain-specific algorithms. Particularly useful when a GO_CODE task requires implementing a non-trivial algorithm: the RESEARCH task can pull the source paper before code generation begins.

**Primary task types:** `RESEARCH`
**Example invocations:** finding papers on XBRL parsing algorithms before implementing a converter; searching for NLP techniques relevant to document classification; checking whether a proposed approach has academic precedent.

---

## 9. YouTube Transcript MCP

**Source:** Standard YouTube Transcript MCP server
**Credentials:** none

### What it does
Fetches transcripts from YouTube videos by URL or video ID. Returns the full spoken text of the video.

### How the swarm uses it
Used by `RESEARCH` tasks to extract insight from video content: product demos, tutorial channels, conference talks, and niche community content. Many subject-matter experts publish exclusively on YouTube; transcripts make their knowledge accessible to the swarm without requiring video playback.

**Primary task types:** `RESEARCH`
**Example invocations:** pulling transcript from a legal tech conference talk; extracting product positioning from a competitor's demo video; reading an accountant's walkthrough of a manual process that the swarm aims to automate.

---

## 10. Memory MCP

**Source:** Standard Memory MCP server (knowledge graph)
**Credentials:** none

### What it does
Provides a persistent knowledge graph that survives across swarm runs. Stores entities, relationships, and facts as structured nodes and edges. Queryable by entity name or relationship type.

### How the swarm uses it
Enables cross-run knowledge accumulation. When the swarm discovers a valuable entity (a competitor, a niche, a technical constraint, a pricing benchmark) in one project run, it can store it in the knowledge graph. Subsequent runs can query the graph rather than re-researching known facts, accelerating DECOMPOSE and RESEARCH tasks over time.

Complements (does not replace) the per-branch context sidecar — the sidecar is ephemeral within a run; Memory MCP is permanent across runs.

**Primary task types:** `DECOMPOSE`, `RESEARCH`
**Example invocations:** storing a confirmed niche opportunity after a successful viability review; recording a competitor's pricing for future reference; persisting a domain glossary entry discovered during research.

---

## Credential Summary

| Server | Env Var | Required | Source |
|---|---|---|---|
| Brave Search | `BRAVE_API_KEY` | Yes | [brave.com/search/api](https://brave.com/search/api/) — free tier |
| GitHub MCP | `GITHUB_TOKEN` | Yes | GitHub → Settings → Developer settings → Personal access tokens |
| mcp-reddit | `REDDIT_CLIENT_ID` | Yes | [reddit.com/prefs/apps](https://www.reddit.com/prefs/apps) — script app |
| mcp-reddit | `REDDIT_CLIENT_SECRET` | Yes | Same as above |
| All others | — | No | No credentials required |

---

*Generated by spec-agent as a supplementary reference document.*
