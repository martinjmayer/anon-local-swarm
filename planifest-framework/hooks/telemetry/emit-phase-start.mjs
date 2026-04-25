#!/usr/bin/env node
/**
 * PreToolUse hook: phase_start telemetry emission.
 *
 * Fires on first tool use within a phase. Guards against re-emission using a
 * flag file keyed by session_id + phase (DD-001, ADR-003).
 *
 * Usage:  node emit-phase-start.mjs <phase>
 *   e.g.  node emit-phase-start.mjs spec
 *
 * Session ID fallback: reads/creates {cwd}/.claude/.planifest-session when
 * PLANIFEST_SESSION_ID is absent (R-005 mitigation, ADR-003).
 *
 * Silent on all errors (ADR-005). No retries. 3-second abort on HTTP.
 */

import { existsSync, mkdirSync, readFileSync, renameSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { randomUUID } from "node:crypto";

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
  // Priority 1: explicit env var
  if (process.env.PLANIFEST_SESSION_ID) return process.env.PLANIFEST_SESSION_ID;
  // Priority 2: hook input session_id field
  if (input?.session_id) return input.session_id;
  // Priority 3: UUID from transcript path filename
  if (input?.transcript_path) {
    const m = input.transcript_path.match(/([a-f0-9-]{36})\.jsonl$/i);
    if (m) return m[1];
  }
  // Priority 4: project-scoped session file (R-005 mitigation)
  try {
    const sessionFile = join(cwd, ".claude", ".planifest-session");
    if (existsSync(sessionFile)) return readFileSync(sessionFile, "utf-8").trim();
    const id = randomUUID();
    mkdirSync(dirname(sessionFile), { recursive: true });
    writeFileSync(sessionFile, id);
    return id;
  } catch {
    return `pid-${process.pid}`;
  }
}

function getFlagPath(sessionId) {
  const dir = join(tmpdir(), "planifest-telemetry");
  return join(dir, `phase-start-${sessionId}-${PHASE}`);
}

try {
  // Sentinel check: no telemetry URL or no phase arg = silent exit (REQ-004)
  if (!BACKEND_URL || !PHASE) process.exit(0);

  const raw = await readStdin();
  const input = JSON.parse(raw);
  const cwd = input?.cwd ?? process.cwd();
  const sessionId = getSessionId(input, cwd);
  const flagPath = getFlagPath(sessionId);

  // Deduplication guard — exit 0 if already emitted this session+phase (ADR-003)
  if (existsSync(flagPath)) process.exit(0);

  // Write flag file atomically with ISO 8601 start timestamp (used by emit-phase-end.mjs)
  const timestamp = new Date().toISOString();
  const flagDir = dirname(flagPath);
  mkdirSync(flagDir, { recursive: true });
  const tmpPath = `${flagPath}.tmp`;
  writeFileSync(tmpPath, timestamp);
  renameSync(tmpPath, flagPath);

  const event = {
    schema_version: "1.0",
    event: "phase_start",
    session_id: sessionId,
    phase: PHASE,
    agent: `planifest-${PHASE}-agent`,
    tool: process.env.PLANIFEST_TOOL ?? "claude-code",
    model: process.env.CLAUDE_API_MODEL ?? "unknown",
    mcp_mode: "none",
    timestamp,
    data: { phase_name: PHASE },
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
  // PreToolUse must never block the session — silent fallback (ADR-005).
}
