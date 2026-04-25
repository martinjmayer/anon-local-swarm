#!/usr/bin/env bash
# context-mode enforcement hook — block-webfetch.sh
#
# Intercepts the WebFetch tool and redirects to ctx_fetch_and_index + ctx_search.
# All WebFetch calls are denied unconditionally — WebFetch always returns the full
# response body into the context window with no size control.
#
# Feature:  0000001-context-mode-enforcement-hooks
# ADR-001:  hookSpecificOutput deny format (not deprecated top-level decision)
# ADR-003:  unconditional block strategy for WebFetch
# Upstream: to be contributed to https://github.com/mksglu/context-mode
#
# Input  (stdin): Claude Code PreToolUse JSON payload
# Output (stdout): hookSpecificOutput JSON deny decision
# Exit:   0 always
#
# Runtime dependency: jq (preferred) OR node (Claude Code always ships with Node.js)
set -euo pipefail

input=$(cat)

if command -v jq >/dev/null 2>&1; then
  # Fast path — jq available
  url=$(printf '%s' "$input" | jq -r '.tool_input.url // "URL"')
  reason="context-mode: Do not use WebFetch. Instead: (1) ctx_fetch_and_index(url:\"$url\") to ingest and index the page content, then (2) ctx_search(queries:[\"...\"]) to retrieve relevant sections. This keeps the response body in the sandbox and protects the context window."
  jq -cn --arg reason "$reason" \
    '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":$reason}}'
else
  # Fallback — Node.js (always co-present with Claude Code)
  printf '%s' "$input" | node -e \
    "let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',()=>{const j=JSON.parse(d);const url=j.tool_input?.url??'URL';const reason='context-mode: Do not use WebFetch. Instead: (1) ctx_fetch_and_index(url:\"'+url+'\") to ingest and index the page content, then (2) ctx_search(queries:[\"...\"]) to retrieve relevant sections. This keeps the response body in the sandbox and protects the context window.';process.stdout.write(JSON.stringify({hookSpecificOutput:{hookEventName:'PreToolUse',permissionDecision:'deny',permissionDecisionReason:reason}})+'\n');})"
fi
