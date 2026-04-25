#!/usr/bin/env bash
# context-mode enforcement hook — block-bash.sh
#
# Intercepts Bash tool calls that match context-flooding patterns.
# Commands whose leading token is in the allowlist are always permitted.
# Pattern matching applies only after the allowlist check.
#
# Feature:  0000001-context-mode-enforcement-hooks
# ADR-001:  hookSpecificOutput deny format
# ADR-002:  hardcoded allowlist (configurable allowlist deferred)
# ADR-003:  pattern-based blocking for Bash (Bash has a safe majority)
# Upstream: to be contributed to https://github.com/mksglu/context-mode
#
# Input  (stdin): Claude Code PreToolUse JSON payload
# Output (stdout): hookSpecificOutput JSON deny decision, or empty (allow)
# Exit:   0 always
#
# Runtime dependency: jq (preferred) OR node (Claude Code always ships with Node.js)
set -euo pipefail

input=$(cat)

# Extract tool_input.command using jq or node
if command -v jq >/dev/null 2>&1; then
  command=$(printf '%s' "$input" | jq -r '.tool_input.command // ""')
else
  command=$(printf '%s' "$input" | node -e \
    "let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',()=>{const j=JSON.parse(d);process.stdout.write((j.tool_input?.command??'')+'\n');})")
fi

# -----------------------------------------------------------------------
# ALLOWLIST CHECK (ADR-002)
# Checks the leading command token only — not the full pipeline.
# Consequence: `git log | grep feat` is allowed (git is leading).
# See quirks.md Q-001 for the `ls | grep` edge case.
# -----------------------------------------------------------------------

leading=$(printf '%s' "$command" | awk '{print $1}')

case "$leading" in
  git|mkdir|rm|mv|cd|ls)
    exit 0
    ;;
esac

leading2=$(printf '%s' "$command" | awk '{print $1, $2}')
case "$leading2" in
  "npm install"|"pip install")
    exit 0
    ;;
esac

# -----------------------------------------------------------------------
# BLOCKED PATTERN CHECK (ADR-003)
# Uses whole-word (-w) matching to avoid false positives on substrings.
# e.g. `cargo` does not match `rg`; `--arg` does not match `rg`.
# -----------------------------------------------------------------------

redirect_type=""

if printf '%s' "$command" | grep -qwE 'grep|rg'; then
  redirect_type="search"
elif printf '%s' "$command" | grep -qwE 'curl|wget' \
  || printf '%s' "$command" | grep -qE 'https?://'; then
  redirect_type="network"
fi

# No blocked pattern — allow
if [ -z "$redirect_type" ]; then
  exit 0
fi

# -----------------------------------------------------------------------
# DENY with redirect instruction (ADR-001)
# -----------------------------------------------------------------------

if command -v jq >/dev/null 2>&1; then
  if [ "$redirect_type" = "search" ]; then
    reason="context-mode: Blocked Bash command. Use ctx_execute(language:\"shell\", code:\"grep ...\") to keep search output in the context-mode sandbox. Original command: $command"
  else
    reason="context-mode: Blocked Bash command. Use ctx_fetch_and_index(url:\"...\") to ingest the URL, then ctx_search(queries:[\"...\"]) to retrieve content. Original command: $command"
  fi
  jq -cn --arg reason "$reason" \
    '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":$reason}}'
else
  # Fallback — Node.js
  if [ "$redirect_type" = "search" ]; then
    redirect_msg='Use ctx_execute(language:"shell", code:"grep ...") to keep search output in the context-mode sandbox.'
  else
    redirect_msg='Use ctx_fetch_and_index(url:"...") to ingest the URL, then ctx_search(queries:["..."]) to retrieve content.'
  fi
  # Pass command and redirect via env vars to avoid shell-quoting issues in node -e
  CMD="$command" REDIRECT="$redirect_msg" node -e \
    "const reason='context-mode: Blocked Bash command. '+process.env.REDIRECT+' Original command: '+process.env.CMD;process.stdout.write(JSON.stringify({hookSpecificOutput:{hookEventName:'PreToolUse',permissionDecision:'deny',permissionDecisionReason:reason}})+'\n');"
fi
