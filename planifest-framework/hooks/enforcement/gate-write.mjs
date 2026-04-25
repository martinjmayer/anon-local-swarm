#!/usr/bin/env node
/**
 * PreToolUse hook: plan compliance write gate.
 *
 * Enforces three checks before allowing a Write or Edit tool call (ADR-004):
 *   1. Always-permitted paths (plan/, docs/, CLAUDE.md, AGENTS.md) → pass
 *   2. design.md must exist → block with exit 2 if absent
 *   3. Target path must match a component path prefix in design.md → block if not
 *
 * Receives: JSON on stdin conforming to the Planifest common envelope (ADR-002)
 * or the raw Claude Code PreToolUse hook input.
 *
 * Exit codes: 0 = pass, 2 = block (DD-004, ADR-005)
 * Silent on unexpected errors — always exit 0 on non-enforcement failures (ADR-005).
 */

import { existsSync, readFileSync } from "node:fs";
import { join, normalize, resolve } from "node:path";

// Always-permitted: planning/doc artefacts and Planifest internal files
const ALWAYS_PERMITTED_PREFIXES = ["plan/", "plan\\", "docs/", "docs\\"];
const ALWAYS_PERMITTED_FILES = [
  "claude.md", "agents.md", ".planifest-session", ".skips", ".feature-id",
  ".gitignore", ".gitattributes", ".claudeignore", ".cursorignore",
  ".windsurfignore", ".clineignore", ".cursorindexingignore",
];

// Headings that introduce the component paths list in design.md (ADR-004)
const PATHS_SECTION_RE = /^##\s+(Component Paths|Scope)\s*$/im;

function readStdin() {
  return new Promise((res) => {
    let data = "";
    process.stdin.setEncoding("utf-8");
    process.stdin.on("data", (c) => { data += c; });
    process.stdin.on("end", () => res(data.replace(/^\uFEFF/, "")));
    process.stdin.resume();
  });
}

function norm(p) {
  return normalize(p).replace(/\\/g, "/").toLowerCase();
}

function isAlwaysPermitted(relPath) {
  const n = norm(relPath);
  const base = n.split("/").pop() ?? "";
  if (ALWAYS_PERMITTED_FILES.includes(base)) return true;
  return ALWAYS_PERMITTED_PREFIXES.some((prefix) => n.startsWith(norm(prefix)));
}

function extractComponentPaths(content) {
  const match = PATHS_SECTION_RE.exec(content);
  if (!match) return null;

  const afterHeading = content.slice(match.index + match[0].length);
  const lines = afterHeading.split("\n");
  const paths = [];

  for (const line of lines) {
    if (/^##\s/.test(line)) break; // stop at next h2
    // Match: "- path", "* path", "- `path`", "| path |" table rows
    const m = line.match(/^[\s*\-|]+`?([^\s`|,{}]+)`?/);
    if (m?.[1] && !m[1].startsWith("{") && m[1].includes("/")) {
      paths.push(m[1].replace(/\/$/, "")); // strip trailing slash for normalised comparison
    }
  }

  return paths.length ? paths : null;
}

function matchesComponentPath(relPath, componentPaths) {
  const normTarget = norm(relPath);
  return componentPaths.some((cp) => {
    const normCp = norm(cp);
    // prefix match: target starts with component path (with separator)
    return normTarget === normCp || normTarget.startsWith(normCp + "/");
  });
}

try {
  const raw = await readStdin();
  const input = JSON.parse(raw);

  // Support both common envelope (ADR-002) and raw Claude Code hook input
  const cwd = input?.cwd ?? process.cwd();
  const toolInput = input?.tool_input ?? input;
  const rawTarget = toolInput?.path ?? toolInput?.file_path ?? "";

  // No target path = pass (not a file-writing tool call)
  if (!rawTarget) process.exit(0);

  // Resolve to a path relative to cwd for prefix matching
  const absTarget = resolve(cwd, rawTarget);
  const relTarget = absTarget.startsWith(cwd)
    ? absTarget.slice(cwd.length).replace(/^[/\\]/, "")
    : rawTarget;

  // Check 1 — always-permitted paths (ADR-004)
  if (isAlwaysPermitted(relTarget)) process.exit(0);

  // Check 2 — design.md must exist
  const designPath = join(cwd, "plan", "current", "design.md");
  if (!existsSync(designPath)) {
    process.stdout.write(
      "[Planifest] No confirmed design at plan/current/design.md. " +
      "Complete Phase 0 first.\n"
    );
    process.exit(2);
  }

  // Check 3 — path must be in component paths list
  let designContent;
  try {
    designContent = readFileSync(designPath, "utf-8");
  } catch {
    process.exit(0); // Can't read design → pass through (ADR-005)
  }

  const componentPaths = extractComponentPaths(designContent);
  if (!componentPaths) {
    // Section missing or unparseable → warn but pass through (ADR-005, R-006)
    process.stderr.write(
      "[Planifest] Warning: Could not parse component paths from " +
      "plan/current/design.md. Ensure a '## Component Paths' section exists. " +
      "Write allowed.\n"
    );
    process.exit(0);
  }

  if (!matchesComponentPath(relTarget, componentPaths)) {
    process.stdout.write(
      `[Planifest] Path '${relTarget}' is not covered by the confirmed design. ` +
      "Add it to the '## Component Paths' section of plan/current/design.md first.\n"
    );
    process.exit(2);
  }

  process.exit(0);
} catch {
  // Never block the session on unexpected errors (ADR-005).
  process.exit(0);
}
