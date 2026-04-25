#!/usr/bin/env node
/**
 * Stop hook: phase_end telemetry emission.
 *
 * Fires at the end of each response turn. Reads the start timestamp from the
 * phase-start flag file to compute duration_ms (ADR-003, REQ-002).
 *
 * Usage:  node emit-phase-end.mjs <phase>
 *   e.g.  node emit-phase-end.mjs spec
 *
 * Silent on all errors (ADR-005). No retries. 3-second abort on HTTP.
 */

import { existsSync, readFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const BACKEND_URL = process.env.PLANIFEST_TELEMETRY_URL;
const PHASE = process.argv[2];

function readStdin() {
  return new Promise((resolve) => {
    let data = "";
    process.stdin.setEncoding("utf-8");
    process.stdin.on("data", (chunk) => { data += chunk; });
    process.stdin.on("end", () => resolve(data.replace(/^\uFEFF/, "")));
    process.stdin.resume();
  });
}

function getSessionId(input, cwd) {
  if (process.env.PLANIFEST_SESSION_ID) return process.env.PLANIFEST_SESSION_ID;
  if (input?.session_id) return input.session_id;
  if (input?.transcript_path) {
    const m = input.transcript_path.match(/([a-f0-9-]{36})\.jsonl$/i);
    if (m) return m[1];
  }
  try {
    const sessionFile = join(cwd, ".claude", ".planifest-session");
    if (existsSync(sessionFile)) return readFileSync(sessionFile, "utf-8").trim();
  } catch { /* silent */ }
  return `pid-${process.pid}`;
}

function getFlagPath(sessionId) {
  return join(tmpdir(), "planifest-telemetry", `phase-start-${sessionId}-${PHASE}`);
}

try {
  // Sentinel check: no telemetry URL or no phase arg = silent exit (REQ-004)
  if (!BACKEND_URL || !PHASE) process.exit(0);

  const raw = await readStdin();
  const input = JSON.parse(raw);
  const cwd = input?.cwd ?? process.cwd();
  const sessionId = getSessionId(input, cwd);
  const now = Date.now();

  // Read start timestamp from flag file for duration_ms (ADR-003)
  let duration_ms;
  try {
    const flagPath = getFlagPath(sessionId);
    if (existsSync(flagPath)) {
      const startTs = new Date(readFileSync(flagPath, "utf-8").trim()).getTime();
      if (!isNaN(startTs)) duration_ms = now - startTs;
    }
  } catch { /* no flag file = omit duration */ }

  const event = {
    schema_version: "1.0",
    event: "phase_end",
    session_id: sessionId,
    phase: PHASE,
    agent: `planifest-${PHASE}-agent`,
    tool: process.env.PLANIFEST_TOOL ?? "claude-code",
    model: process.env.CLAUDE_API_MODEL ?? "unknown",
    mcp_mode: "none",
    timestamp: new Date().toISOString(),
    data: {
      phase_name: PHASE,
      status: "pass",
      ...(duration_ms !== undefined ? { duration_ms } : {}),
    },
  };

  // Fire-and-forget: abort after 3 s (ADR-005, NFR)
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
  // Stop hook must never block the session — silent fallback (ADR-005).
}
