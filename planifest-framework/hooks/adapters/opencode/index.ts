/**
 * @planifest/opencode-hooks — Tier 2 OpenCode plugin shim (ADR-001, ADR-002).
 *
 * Implements the OpenCode plugin interface to bridge OpenCode's JS/TS plugin
 * API to Planifest's shared enforcement and telemetry hook scripts.
 *
 * Registered in opencode.json under the "plugins" array by setup/opencode.sh.
 *
 * Requires: Bun runtime (bundled with OpenCode). No external npm dependencies.
 *
 * Hook script delegation uses Bun.spawnSync so exit code 2 propagates as a
 * block back to OpenCode (ADR-001, ADR-005).
 */

import { existsSync, readFileSync, writeFileSync, mkdirSync } from "node:fs";
import { join, dirname } from "node:path";
import { randomUUID } from "node:crypto";

// Resolve paths relative to this plugin file
const PLUGIN_DIR = dirname(new URL(import.meta.url).pathname);
const HOOKS_DIR = dirname(PLUGIN_DIR); // planifest-framework/hooks/

const BACKEND_URL = process.env.PLANIFEST_TELEMETRY_URL;
const ACTIVE_PHASE = process.env.PLANIFEST_PHASE ?? "codegen";

function getSessionId(cwd: string): string {
  if (process.env.PLANIFEST_SESSION_ID) return process.env.PLANIFEST_SESSION_ID;
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

function delegateToScript(scriptName: string, phase: string | undefined, envelope: object): number {
  const subdir = scriptName.startsWith("emit-") ? "telemetry" : "enforcement";
  const scriptPath = join(HOOKS_DIR, subdir, `${scriptName}.mjs`);

  if (!existsSync(scriptPath)) return 0;

  const args = phase ? [scriptPath, phase] : [scriptPath];
  const result = Bun.spawnSync([process.execPath, ...args], {
    stdin: new TextEncoder().encode(JSON.stringify(envelope)),
    stdout: "pipe",
    stderr: "pipe",
  });

  // Pass through stdout (gate-write messages, check-design additionalContext)
  if (result.stdout) process.stdout.write(result.stdout);
  if (result.stderr) process.stderr.write(result.stderr);

  return result.exitCode ?? 0;
}

function buildEnvelope(event: Record<string, unknown>, cwd: string): object {
  return {
    session_id: getSessionId(cwd),
    cwd,
    tool_input: event.toolInput ?? event.tool_input ?? {},
    event: event.type ?? "pre_tool_use",
  };
}

// OpenCode plugin export — implements the OpenCode plugin API
export default {
  name: "@planifest/opencode-hooks",

  /**
   * Fires before every tool use.
   * Delegates to gate-write.mjs for Write/Edit; emit-phase-start.mjs for all tools.
   */
  async onPreToolUse(event: Record<string, unknown>) {
    const cwd = (event.cwd as string) ?? process.cwd();
    const toolName = (event.toolName as string) ?? (event.tool_name as string) ?? "";
    const envelope = buildEnvelope(event, cwd);

    // Telemetry: emit phase_start (deduplicated by flag file)
    delegateToScript("emit-phase-start", ACTIVE_PHASE, envelope);

    // Enforcement: gate write/edit operations
    if (/write|edit/i.test(toolName)) {
      const exitCode = delegateToScript("gate-write", undefined, envelope);
      if (exitCode === 2) {
        // Block the tool call — OpenCode reads this as a cancellation
        return { block: true };
      }
    }

    return { block: false };
  },

  /**
   * Fires after every tool use (including Stop/turn-end equivalent).
   * Delegates to emit-phase-end.mjs.
   */
  async onPostToolUse(event: Record<string, unknown>) {
    if (!BACKEND_URL) return;
    const cwd = (event.cwd as string) ?? process.cwd();
    const envelope = buildEnvelope(event, cwd);
    delegateToScript("emit-phase-end", ACTIVE_PHASE, envelope);
  },

  /**
   * Fires on each user prompt submission.
   * Delegates to check-design.mjs for scope context injection.
   */
  async onUserPrompt(event: Record<string, unknown>) {
    const cwd = (event.cwd as string) ?? process.cwd();
    const envelope = buildEnvelope(event, cwd);
    // check-design.mjs writes additionalContext JSON to stdout
    delegateToScript("check-design", undefined, envelope);
  },
};
