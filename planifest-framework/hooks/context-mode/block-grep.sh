#!/usr/bin/env bash
# context-mode enforcement hook — block-grep.sh
#
# Intercepts the Grep tool and redirects to ctx_execute shell.
# All Grep calls are denied unconditionally — Grep has no safe subset
# in a context-mode environment (any file search floods the context window).
#
# Feature:  0000001-context-mode-enforcement-hooks
# ADR-001:  hookSpecificOutput deny format (not deprecated top-level decision)
# ADR-003:  unconditional block strategy for Grep
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
  pattern=$(printf '%s' "$input" | jq -r '.tool_input.pattern // "PATTERN"')
  path=$(printf '%s' "$input" | jq -r '.tool_input.path // "PATH"')
  reason="context-mode: Do not use Grep. Use ctx_execute(language:\"shell\", code:\"grep '$pattern' $path\") to keep search output in the context-mode sandbox and protect the context window."
  jq -cn --arg reason "$reason" \
    '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":$reason}}'
else
  # Fallback — Node.js (always co-present with Claude Code)
  printf '%s' "$input" | node -e \
    "let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',()=>{const j=JSON.parse(d);const pattern=j.tool_input?.pattern??'PATTERN';const path=j.tool_input?.path??'PATH';const reason='context-mode: Do not use Grep. Use ctx_execute(language:\"shell\", code:\"grep \\''+pattern+'\\' '+path+'\") to keep search output in the context-mode sandbox and protect the context window.';process.stdout.write(JSON.stringify({hookSpecificOutput:{hookEventName:'PreToolUse',permissionDecision:'deny',permissionDecisionReason:reason}})+'\n');})"
fi
