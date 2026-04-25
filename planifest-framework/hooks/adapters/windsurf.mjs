#!/usr/bin/env node
/**
 * Windsurf hook adapter — Tier 1 (ADR-001, ADR-002).
 *
 * Translates Windsurf's native hook envelope to the Planifest common envelope,
 * then delegates to the appropriate shared hook script.
 *
 * Windsurf PreToolUse envelope shape:
 *   { tool_name, tool_input, session_id?, workspace_root? }
 *
 * Usage: node windsurf.mjs <script> [phase]
 *   e.g.  node windsurf.mjs gate-write
 *         node windsurf.mjs emit-phase-start spec
 */

import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { dirname, join } from "node:path";

const SCRIPT_NAME = process.argv[2];
const PHASE_ARG = process.argv[3];

const ADAPTER_DIR = dirname(new URL(import.meta.url).pathname);
const HOOKS_DIR = dirname(ADAPTER_DIR);

async function readStdin() {
  return new Promise((resolve) => {
    let data = "";
    process.stdin.setEncoding("utf-8");
    process.stdin.on("data", (c) => { data += c; });
    process.stdin.on("end", () => resolve(data.replace(/^\uFEFF/, "")));
    process.stdin.resume();
  });
}

try {
  if (!SCRIPT_NAME) process.exit(0);

  const raw = await readStdin();
  const windsurfInput = JSON.parse(raw);

  // Translate Windsurf envelope → Planifest common envelope (ADR-002)
  const envelope = {
    session_id: windsurfInput.session_id ?? windsurfInput.sessionId,
    cwd: windsurfInput.workspace_root ?? windsurfInput.workspaceRoot ?? windsurfInput.cwd ?? process.cwd(),
    tool_input: windsurfInput.tool_input ?? windsurfInput.toolInput ?? {},
    event: "PreToolUse",
  };

  const scriptSubdir = SCRIPT_NAME.startsWith("emit-") ? "telemetry" : "enforcement";
  const scriptPath = join(HOOKS_DIR, scriptSubdir, `${SCRIPT_NAME}.mjs`);

  if (!existsSync(scriptPath)) process.exit(0);

  const args = PHASE_ARG ? [scriptPath, PHASE_ARG] : [scriptPath];
  const result = spawnSync(process.execPath, args, {
    input: JSON.stringify(envelope),
    encoding: "utf-8",
    stdio: ["pipe", "pipe", "pipe"],
  });

  if (result.stdout) process.stdout.write(result.stdout);
  if (result.stderr) process.stderr.write(result.stderr);
  process.exit(result.status ?? 0);
} catch {
  process.exit(0);
}
