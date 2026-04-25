#!/usr/bin/env bash
# Test assertion helpers for context-mode enforcement hook tests
# Usage: source this file, call assert_* functions, call print_summary at the end.
#
# JSON helpers use jq if available, fall back to Node.js — matching the hook scripts.

PASS=0
FAIL=0

# -----------------------------------------------------------------------
# JSON helpers (jq preferred, node fallback)
# -----------------------------------------------------------------------

get_permission_decision() {
  local output="$1"
  if command -v jq >/dev/null 2>&1; then
    printf '%s' "$output" | jq -r '.hookSpecificOutput.permissionDecision // ""'
  else
    printf '%s' "$output" | node -e \
      "let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',()=>{const j=JSON.parse(d);console.log(j?.hookSpecificOutput?.permissionDecision??'');});"
  fi
}

get_permission_reason() {
  local output="$1"
  if command -v jq >/dev/null 2>&1; then
    printf '%s' "$output" | jq -r '.hookSpecificOutput.permissionDecisionReason // ""'
  else
    printf '%s' "$output" | node -e \
      "let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',()=>{const j=JSON.parse(d);console.log(j?.hookSpecificOutput?.permissionDecisionReason??'');});"
  fi
}

get_hook_event_name() {
  local output="$1"
  if command -v jq >/dev/null 2>&1; then
    printf '%s' "$output" | jq -r '.hookSpecificOutput.hookEventName // ""'
  else
    printf '%s' "$output" | node -e \
      "let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',()=>{const j=JSON.parse(d);console.log(j?.hookSpecificOutput?.hookEventName??'');});"
  fi
}

validate_json() {
  # Returns 0 if input is valid JSON, non-zero otherwise
  local output="$1"
  if command -v jq >/dev/null 2>&1; then
    printf '%s' "$output" | jq empty 2>/dev/null
  else
    printf '%s' "$output" | node -e \
      "let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',()=>{JSON.parse(d);});" 2>/dev/null
  fi
}

# -----------------------------------------------------------------------
# Assertion helpers
# -----------------------------------------------------------------------

assert_equals() {
  local expected="$1"
  local actual="$2"
  local message="${3:-assertion}"

  if [ "$expected" = "$actual" ]; then
    echo "  PASS: $message"
    ((PASS++)) || true
  else
    echo "  FAIL: $message"
    echo "        expected: $expected"
    echo "        actual:   $actual"
    ((FAIL++)) || true
  fi
}

assert_contains() {
  local needle="$1"
  local haystack="$2"
  local message="${3:-contains assertion}"

  if printf '%s' "$haystack" | grep -qF "$needle"; then
    echo "  PASS: $message"
    ((PASS++)) || true
  else
    echo "  FAIL: $message"
    echo "        needle:   $needle"
    echo "        haystack: $haystack"
    ((FAIL++)) || true
  fi
}

assert_exit_zero() {
  local actual_exit="$1"
  local message="${2:-exit code assertion}"

  if [ "$actual_exit" -eq 0 ]; then
    echo "  PASS: $message"
    ((PASS++)) || true
  else
    echo "  FAIL: $message — expected exit 0, got $actual_exit"
    ((FAIL++)) || true
  fi
}

print_summary() {
  echo ""
  if [ "$FAIL" -eq 0 ]; then
    echo "Results: $PASS passed, $FAIL failed ✓"
    exit 0
  else
    echo "Results: $PASS passed, $FAIL failed ✗"
    exit 1
  fi
}
