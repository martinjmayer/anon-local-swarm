#!/usr/bin/env node
/**
 * PostToolUse hook: context pressure monitor for structured telemetry.
 *
 * Emits a `context_pressure` event to the structured telemetry MCP backend
 * when estimated context fill exceeds THRESHOLD_PCT (default: 70%).
 *
 * Installed only when both --structured-telemetry-mcp and --context-mode-mcp
 * are active at setup time. See plan/current/design.md — Context Pressure Hook.
 *
 * Fill % is estimated from transcript file size. This is a proxy metric —
 * it grows proportionally with context use within a session and resets at
 * session start. It does not account for compaction events.
 *
 * Silent on all errors (NFR-001). No retries. No local fallback (NFR-002).
 */

import { statSync } from "node:fs";

const THRESHOLD_PCT = 70;
// Rough estimate: ~900 KB of JSONL transcript ≈ full 200K token context window.
// 70% threshold ≈ 630 KB.
const ESTIMATED_MAX_BYTES = 900_000;
const BACKEND_URL = process.env.PLANIFEST_TELEMETRY_URL ?? "http://localhost:3741";

function readStdin() {
  return new Promise((resolve, reject) => {
    let data = "";
    process.stdin.setEncoding("utf-8");
    process.stdin.on("data", (chunk) => { data += chunk; });
    process.stdin.on("end", () => resolve(data.replace(/^\uFEFF/, "")));
    process.stdin.on("error", reject);
    process.stdin.resume();
  });
}

function getSessionId(input) {
  if (input.transcript_path) {
    const match = input.transcript_path.match(/([a-f0-9-]{36})\.jsonl$/i);
    if (match) return match[1];
  }
  if (input.session_id) return input.session_id;
  return `pid-${process.ppid}`;
}

try {
  const raw = await readStdin();
  const input = JSON.parse(raw);

  if (!input.transcript_path) {
    process.exit(0);
  }

  let transcriptBytes;
  try {
    transcriptBytes = statSync(input.transcript_path).size;
  } catch {
    process.exit(0);
  }

  const context_fill_pct =
    Math.min(100, Math.round((transcriptBytes / ESTIMATED_MAX_BYTES) * 1000) / 10);

  if (context_fill_pct <= THRESHOLD_PCT) {
    process.exit(0);
  }

  const event = {
    schema_version: "1.0",
    event: "context_pressure",
    session_id: getSessionId(input),
    phase: "monitoring",
    agent: "context-pressure-hook",
    tool: "claude-code",
    model: process.env.CLAUDE_API_MODEL ?? "unknown",
    mcp_mode: "workspace+context",
    timestamp: new Date().toISOString(),
    data: {
      context_fill_pct,
      unused_sources: [],
      trigger: "threshold_exceeded",
    },
  };

  // Fire-and-forget: abort after 3 s to keep the hook fast.
  const ac = new AbortController();
  const timer = setTimeout(() => ac.abort(), 3_000);
  try {
    await fetch(`${BACKEND_URL}/emit`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(event),
      signal: ac.signal,
    });
  } finally {
    clearTimeout(timer);
  }
} catch {
  // PostToolUse must never block the session — silent fallback (NFR-001).
}
