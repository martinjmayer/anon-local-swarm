#!/usr/bin/env bash
# Test runner for planifest-framework tests
# Usage: bash planifest-framework/tests/run-tests.sh

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

TOTAL_PASS=0
TOTAL_FAIL=0

run_suite() {
  local file="$1"
  local name
  name="$(basename "$file")"
  echo ""
  echo "--- $name ---"
  if bash "$file"; then
    TOTAL_PASS=$((TOTAL_PASS + 1))
  else
    TOTAL_FAIL=$((TOTAL_FAIL + 1))
  fi
}

run_suite "$SCRIPT_DIR/test-setup-telemetry.sh"
run_suite "$SCRIPT_DIR/test-context-pressure.sh"
run_suite "$SCRIPT_DIR/test-skill-telemetry.sh"

echo ""
echo "================================"
if [ "$TOTAL_FAIL" -eq 0 ]; then
  echo "All $TOTAL_PASS test files passed."
  exit 0
else
  echo "$TOTAL_PASS passed, $TOTAL_FAIL failed."
  exit 1
fi
