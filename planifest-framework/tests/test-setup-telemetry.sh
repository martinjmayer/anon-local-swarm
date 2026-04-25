#!/usr/bin/env bash
# Tests for --structured-telemetry-mcp flag in setup.sh
# Covers: REQ-001 (idempotency), REQ-002 (--backend-url), REQ-004 (sentinel),
#         REQ-008 (PostToolUse wiring), REQ-010 (both-flags gate)

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers/assert.sh"

FRAMEWORK_SRC="$(cd "$SCRIPT_DIR/.." && pwd)"

# -----------------------------------------------------------------------
# Helpers
# -----------------------------------------------------------------------

make_workspace() {
  local dir
  dir=$(mktemp -d -t planifest_tel_test_XXXXXX)
  cp -r "$FRAMEWORK_SRC" "$dir/planifest-framework"
  # git init so activate_guardrails() (git config core.hooksPath) doesn't fail
  git init "$dir" >/dev/null 2>&1
  # Allow git to operate in this temp dir (Windows safe.directory guard)
  git config --global --add safe.directory "$dir" >/dev/null 2>&1 || true
  echo "$dir"
}

get_posttooluse_json() {
  local settings_file="$1"
  node -e "
    const fs = require('fs');
    const raw = fs.readFileSync('$settings_file', 'utf8').replace(/^\uFEFF/,'');
    const j = JSON.parse(raw);
    console.log(JSON.stringify(j?.hooks?.PostToolUse ?? []));
  "
}

count_context_pressure_entries() {
  local settings_file="$1"
  node -e "
    const fs = require('fs');
    const raw = fs.readFileSync('$settings_file', 'utf8').replace(/^\uFEFF/,'');
    const j = JSON.parse(raw);
    const entries = (j?.hooks?.PostToolUse ?? []).filter(e =>
      JSON.stringify(e).includes('context-pressure')
    );
    console.log(entries.length);
  "
}

# -----------------------------------------------------------------------

echo ""
echo "=== REQ-004: sentinel created with --structured-telemetry-mcp ==="

WS=$(make_workspace); cd "$WS"
bash planifest-framework/setup.sh claude-code --structured-telemetry-mcp >/dev/null 2>&1
assert_exit_zero $? "setup exits 0"
[ -f ".claude/telemetry-enabled" ] && assert_equals "0" "0" "REQ-004: .claude/telemetry-enabled created" \
  || assert_equals "1" "0" "REQ-004: .claude/telemetry-enabled created"
rm -rf "$WS"

# -----------------------------------------------------------------------

echo ""
echo "=== REQ-004: sentinel NOT created without flag ==="

WS=$(make_workspace); cd "$WS"
bash planifest-framework/setup.sh claude-code >/dev/null 2>&1
[ ! -f ".claude/telemetry-enabled" ] && assert_equals "0" "0" "REQ-004: no sentinel without flag" \
  || assert_equals "1" "0" "REQ-004: no sentinel without flag"
rm -rf "$WS"

# -----------------------------------------------------------------------

echo ""
echo "=== REQ-010: hook NOT installed with --structured-telemetry-mcp alone ==="

WS=$(make_workspace); cd "$WS"
bash planifest-framework/setup.sh claude-code --structured-telemetry-mcp >/dev/null 2>&1
[ ! -f ".claude/hooks/telemetry/context-pressure.mjs" ] \
  && assert_equals "0" "0" "REQ-010: no hook without --context-mode-mcp" \
  || assert_equals "1" "0" "REQ-010: no hook without --context-mode-mcp"
rm -rf "$WS"

# -----------------------------------------------------------------------

echo ""
echo "=== REQ-010: hook installed with both flags ==="

WS=$(make_workspace); cd "$WS"
bash planifest-framework/setup.sh claude-code --context-mode-mcp --structured-telemetry-mcp >/dev/null 2>&1
assert_exit_zero $? "setup exits 0 with both flags"
[ -f ".claude/hooks/telemetry/context-pressure.mjs" ] \
  && assert_equals "0" "0" "REQ-010: context-pressure.mjs installed" \
  || assert_equals "1" "0" "REQ-010: context-pressure.mjs installed"
rm -rf "$WS"

# -----------------------------------------------------------------------

echo ""
echo "=== REQ-008: PostToolUse entry wired in settings.json with both flags ==="

WS=$(make_workspace); cd "$WS"
bash planifest-framework/setup.sh claude-code --context-mode-mcp --structured-telemetry-mcp >/dev/null 2>&1
PT=$(get_posttooluse_json ".claude/settings.json")
assert_contains "context-pressure" "$PT" "REQ-008: PostToolUse references context-pressure.mjs"
assert_contains "localhost:3741"    "$PT" "REQ-008: default backend URL present"
assert_contains "async"             "$PT" "REQ-008: hook is async"
rm -rf "$WS"

# -----------------------------------------------------------------------

echo ""
echo "=== REQ-002: --backend-url overrides URL in settings.json ==="

WS=$(make_workspace); cd "$WS"
bash planifest-framework/setup.sh claude-code --context-mode-mcp --structured-telemetry-mcp \
  --backend-url http://localhost:9999 >/dev/null 2>&1
PT=$(get_posttooluse_json ".claude/settings.json")
assert_contains "localhost:9999" "$PT" "REQ-002: custom URL embedded in command"
rm -rf "$WS"

# -----------------------------------------------------------------------

echo ""
echo "=== REQ-001: idempotency — re-run produces exactly one PostToolUse entry ==="

WS=$(make_workspace); cd "$WS"
bash planifest-framework/setup.sh claude-code --context-mode-mcp --structured-telemetry-mcp >/dev/null 2>&1
bash planifest-framework/setup.sh claude-code --context-mode-mcp --structured-telemetry-mcp >/dev/null 2>&1
COUNT=$(count_context_pressure_entries ".claude/settings.json")
assert_equals "1" "$COUNT" "REQ-001: exactly one context-pressure entry after two runs"
rm -rf "$WS"

# -----------------------------------------------------------------------

print_summary
