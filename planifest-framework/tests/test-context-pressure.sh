#!/usr/bin/env bash
# Tests for planifest-framework/hooks/telemetry/context-pressure.mjs
# Covers: REQ-008 (emission logic, threshold, event shape), NFR-001 (silent failure)

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers/assert.sh"

HOOK="$(cd "$SCRIPT_DIR/.." && pwd)/hooks/telemetry/context-pressure.mjs"
MOCK_PORT=19741

# Write the mock HTTP server to a temp file so it can be run as a plain node process.
# Using -t flag for mktemp ensures the path lands in the real Windows temp dir on Git Bash.
MOCK_SETUP_DIR=$(mktemp -d -t planifest_mock_XXXXXX)
MOCK_JS="$MOCK_SETUP_DIR/mock-server.mjs"

cat > "$MOCK_JS" << 'MOCK_SERVER_EOF'
import http from 'node:http';
import { writeFileSync } from 'node:fs';
const [,, receivedFile, port, readyFile] = process.argv;
const server = http.createServer((req, res) => {
  let body = '';
  req.on('data', c => body += c);
  req.on('end', () => {
    writeFileSync(receivedFile, body);
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end('{"ok":true,"id":"mock-id"}');
  });
});
server.listen(parseInt(port), '127.0.0.1', () => {
  writeFileSync(readyFile, 'ready');
});
MOCK_SERVER_EOF

wait_for_file() {
  local file="$1"
  local max_attempts="${2:-40}"  # 40 × 50ms = 2s
  local i=0
  while [ ! -f "$file" ] && [ "$i" -lt "$max_attempts" ]; do
    sleep 0.05
    i=$((i + 1))
  done
  [ -f "$file" ]
}

# -----------------------------------------------------------------------

echo ""
echo "=== NFR-001: no transcript_path — exits 0, no output ==="

INPUT='{"session_id":"test-001","hook_event_name":"PostToolUse","tool_name":"Read"}'
output=$(printf '%s' "$INPUT" | node "$HOOK" 2>&1)
assert_exit_zero $? "NFR-001: exits 0 when transcript_path absent"
assert_equals "" "$output" "NFR-001: no output when transcript_path absent"

# -----------------------------------------------------------------------

echo ""
echo "=== NFR-001: nonexistent transcript_path — exits 0, no output ==="

INPUT='{"session_id":"test-002","hook_event_name":"PostToolUse","tool_name":"Read","transcript_path":"/tmp/does-not-exist-at-all.jsonl"}'
output=$(printf '%s' "$INPUT" | node "$HOOK" 2>&1)
assert_exit_zero $? "NFR-001: exits 0 on missing transcript file"
assert_equals "" "$output" "NFR-001: no output on missing transcript file"

# -----------------------------------------------------------------------

echo ""
echo "=== REQ-008: below threshold (100 KB) — no POST, exits 0 ==="

SMALL_DIR=$(mktemp -d -t planifest_small_XXXXXX)
SMALL_FILE="$SMALL_DIR/session.jsonl"
# 100 KB << 630 KB threshold (70% of 900 KB)
node -e "require('fs').writeFileSync(process.argv[1], Buffer.alloc(100 * 1024))" "$SMALL_FILE"

INPUT="{\"session_id\":\"test-003\",\"hook_event_name\":\"PostToolUse\",\"tool_name\":\"Read\",\"transcript_path\":\"$SMALL_FILE\"}"
output=$(printf '%s' "$INPUT" | PLANIFEST_TELEMETRY_URL=http://127.0.0.1:19999 node "$HOOK" 2>&1)
assert_exit_zero $? "REQ-008: exits 0 below threshold"
assert_equals "" "$output" "REQ-008: no output below threshold"
rm -rf "$SMALL_DIR"

# -----------------------------------------------------------------------

echo ""
echo "=== REQ-008: above threshold (700 KB) — POSTs correct event envelope ==="

MOCK_RUN_DIR=$(mktemp -d -t planifest_mock_run_XXXXXX)
RECEIVED_FILE="$MOCK_RUN_DIR/received.json"
READY_FILE="$MOCK_RUN_DIR/ready"

# Start mock server — runs as a background node process in the current shell
node "$MOCK_JS" "$RECEIVED_FILE" "$MOCK_PORT" "$READY_FILE" &
MOCK_PID=$!
wait_for_file "$READY_FILE" 40

TRANSCRIPT_DIR=$(mktemp -d -t planifest_transcript_XXXXXX)
UUID="aabbccdd-1122-3344-5566-778899aabbcc"
LARGE_FILE="$TRANSCRIPT_DIR/${UUID}.jsonl"
# 700 KB > 630 KB threshold
node -e "require('fs').writeFileSync(process.argv[1], Buffer.alloc(700 * 1024))" "$LARGE_FILE"

INPUT="{\"session_id\":\"ignored\",\"hook_event_name\":\"PostToolUse\",\"tool_name\":\"Read\",\"transcript_path\":\"$LARGE_FILE\"}"
printf '%s' "$INPUT" | PLANIFEST_TELEMETRY_URL=http://127.0.0.1:$MOCK_PORT node "$HOOK" 2>/dev/null || true
# Give fetch time to complete before killing mock
wait_for_file "$RECEIVED_FILE" 20

kill "$MOCK_PID" 2>/dev/null || true

if [ -f "$RECEIVED_FILE" ]; then
  BODY=$(cat "$RECEIVED_FILE")
  assert_contains '"context_pressure"'    "$BODY" "REQ-008: event type is context_pressure"
  assert_contains '"schema_version"'      "$BODY" "REQ-008: envelope has schema_version"
  assert_contains '"1.0"'                 "$BODY" "REQ-008: schema_version is 1.0"
  assert_contains '"context_fill_pct"'    "$BODY" "REQ-008: data includes context_fill_pct"
  assert_contains '"threshold_exceeded"'  "$BODY" "REQ-008: trigger is threshold_exceeded"
  assert_contains '"unused_sources"'      "$BODY" "REQ-008: data includes unused_sources"
  assert_contains "$UUID"                 "$BODY" "REQ-008: session_id extracted from transcript UUID"
  assert_contains '"monitoring"'          "$BODY" "REQ-008: phase is monitoring"
else
  echo "  FAIL: mock server did not receive a request within timeout"
  ((FAIL++)) || true
fi
rm -rf "$MOCK_RUN_DIR" "$TRANSCRIPT_DIR"

# -----------------------------------------------------------------------

echo ""
echo "=== NFR-001: dead backend — exits 0, no output ==="

DEAD_DIR=$(mktemp -d -t planifest_dead_XXXXXX)
UUID2="11223344-aabb-ccdd-eeff-001122334455"
DEAD_FILE="$DEAD_DIR/${UUID2}.jsonl"
node -e "require('fs').writeFileSync(process.argv[1], Buffer.alloc(700 * 1024))" "$DEAD_FILE"

INPUT="{\"session_id\":\"test-005\",\"hook_event_name\":\"PostToolUse\",\"tool_name\":\"Read\",\"transcript_path\":\"$DEAD_FILE\"}"
output=$(printf '%s' "$INPUT" | PLANIFEST_TELEMETRY_URL=http://127.0.0.1:19999 node "$HOOK" 2>&1)
assert_exit_zero $? "NFR-001: exits 0 when backend is unreachable"
assert_equals "" "$output" "NFR-001: no output when backend is unreachable"
rm -rf "$DEAD_DIR"

# -----------------------------------------------------------------------

rm -rf "$MOCK_SETUP_DIR"
print_summary
