#!/usr/bin/env node
/**
 * Codex CLI hook adapter — Tier 1b (ADR-001, ADR-002).
 *
 * Translates Codex CLI's native hook envelope to the Planifest common envelope,
 * then delegates to the appropriate shared hook script.
 *
 * NOTE: Codex CLI hooks are Bash-only. Write interception works on macOS/Linux.
 * Windows is not supported (REQ-010). This adapter exits 0 silently on Windows.
 *
 * Codex pre_tool_use envelope shape:
 *   { tool, input, session_id?, cwd? }
 *
 * Requires: features.codex_hooks = true in .codex/config.toml (REQ-010)
 *
 * Usage: node codex.mjs <script> [phase]
 */

import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { dirname, join } from "node:path";
import { platform } from "node:os";

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
  // Windows: Tier 1b does not support write gating on Windows (REQ-010)
  if (platform() === "win32") process.exit(0);

  if (!SCRIPT_NAME) process.exit(0);

  const raw = await readStdin();
  const codexInput = JSON.parse(raw);

  // Translate Codex envelope → Planifest common envelope (ADR-002)
  const envelope = {
    session_id: codexInput.session_id ?? codexInput.sessionId,
    cwd: codexInput.cwd ?? process.cwd(),
    tool_input: codexInput.input ?? codexInput.tool_input ?? codexInput.toolInput ?? {},
    event: "pre_tool_use",
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
