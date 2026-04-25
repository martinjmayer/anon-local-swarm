#!/usr/bin/env node
/**
 * UserPromptSubmit hook: scope context injection.
 *
 * Reads plan/current/design.md and injects the component paths / scope section
 * as additionalContext so the agent always has confirmed scope visible at the
 * start of every turn (REQ-005, ADR-004).
 *
 * Receives: JSON on stdin (Claude Code UserPromptSubmit hook input or common envelope)
 * Outputs: JSON { additionalContext: "<string>" } on stdout, or nothing on skip.
 *
 * Always exits 0 — never blocks a turn (ADR-005).
 */

import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";

// Headings that introduce the scope section in design.md
const SCOPE_SECTION_RE = /^##\s+(Component Paths|Scope)\s*$/im;

function readStdin() {
  return new Promise((resolve) => {
    let data = "";
    process.stdin.setEncoding("utf-8");
    process.stdin.on("data", (chunk) => { data += chunk; });
    process.stdin.on("end", () => resolve(data.replace(/^\uFEFF/, "")));
    process.stdin.resume();
  });
}

function extractScope(content) {
  const match = SCOPE_SECTION_RE.exec(content);
  if (!match) return null;

  // Capture heading + content until next h2
  const afterStart = content.slice(match.index);
  const lines = afterStart.split("\n");
  const scopeLines = [lines[0]];

  for (let i = 1; i < lines.length; i++) {
    if (/^##\s/.test(lines[i]) && i > 1) break;
    scopeLines.push(lines[i]);
  }

  return scopeLines.join("\n").trim() || null;
}

try {
  const raw = await readStdin();
  const input = JSON.parse(raw);
  const cwd = input?.cwd ?? process.cwd();

  const designPath = join(cwd, "plan", "current", "design.md");
  if (!existsSync(designPath)) process.exit(0);

  let designContent;
  try {
    designContent = readFileSync(designPath, "utf-8");
  } catch {
    process.exit(0);
  }

  const scopeSection = extractScope(designContent);
  if (!scopeSection) process.exit(0);

  const additionalContext =
    "[Planifest] Confirmed scope from plan/current/design.md:\n\n" + scopeSection;

  process.stdout.write(JSON.stringify({ additionalContext }));
  process.exit(0);
} catch {
  // UserPromptSubmit must never block a turn — silent fallback (ADR-005).
  process.exit(0);
}
