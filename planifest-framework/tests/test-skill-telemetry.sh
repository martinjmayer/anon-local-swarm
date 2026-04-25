#!/usr/bin/env bash
# Validates that all 8 SKILL.md files contain the required Telemetry section.
# Covers: REQ-004 (gate text), REQ-006 (event coverage per skill)

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers/assert.sh"

SKILLS_DIR="$(cd "$SCRIPT_DIR/../skills" && pwd)"

SKILLS=(
  planifest-orchestrator
  planifest-spec-agent
  planifest-adr-agent
  planifest-codegen-agent
  planifest-validate-agent
  planifest-change-agent
  planifest-security-agent
  planifest-docs-agent
)

check_skill() {
  local skill="$1"
  local file="$SKILLS_DIR/$skill/SKILL.md"

  echo ""
  echo "=== $skill ==="

  if [ ! -f "$file" ]; then
    echo "  FAIL: SKILL.md not found at $file"
    ((FAIL++)) || true
    return
  fi

  local content
  content=$(cat "$file")

  assert_contains "## Telemetry"            "$content" "$skill: has ## Telemetry section"
  assert_contains "emit_event"              "$content" "$skill: gate references emit_event"
  assert_contains "telemetry-enabled"       "$content" "$skill: gate references telemetry-enabled"
  assert_contains "skip silently"           "$content" "$skill: gate specifies silent skip"
  assert_contains "phase_start"             "$content" "$skill: emits phase_start"
  assert_contains "phase_end"               "$content" "$skill: emits phase_end"
}

# -----------------------------------------------------------------------
# All 8 skills: common gate + phase lifecycle
# -----------------------------------------------------------------------

for skill in "${SKILLS[@]}"; do
  check_skill "$skill"
done

# -----------------------------------------------------------------------
# Skill-specific event coverage
# -----------------------------------------------------------------------

echo ""
echo "=== Skill-specific events ==="

check_event() {
  local skill="$1"
  local event="$2"
  local content
  content=$(cat "$SKILLS_DIR/$skill/SKILL.md")
  assert_contains "$event" "$content" "$skill: emits $event"
}

check_event planifest-orchestrator    "phase_skip"
check_event planifest-orchestrator    "spec_gap"
check_event planifest-orchestrator    "mcp_impact"
check_event planifest-spec-agent      "spec_gap"
check_event planifest-adr-agent       "adr_decision"
check_event planifest-codegen-agent   "deviation"
check_event planifest-codegen-agent   "migration_proposal"
check_event planifest-codegen-agent   "self_correction"
check_event planifest-codegen-agent   "retry_limit_exceeded"
check_event planifest-validate-agent  "validation_failure"
check_event planifest-validate-agent  "self_correction"
check_event planifest-validate-agent  "retry_limit_exceeded"
check_event planifest-change-agent    "deviation"
check_event planifest-change-agent    "migration_proposal"
check_event planifest-change-agent    "retry_limit_exceeded"
check_event planifest-security-agent  "security_finding"
check_event planifest-security-agent  "deviation"
check_event planifest-docs-agent      "doc_gap"
check_event planifest-docs-agent      "deviation"

# -----------------------------------------------------------------------

print_summary
