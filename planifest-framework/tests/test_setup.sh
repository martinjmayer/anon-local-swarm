#!/usr/bin/env bash
set -euo pipefail

echo "Running setup.sh tests..."

# Create a temporary test workspace
TEST_DIR=$(mktemp -d -t planifest_test_XXXXXX)
trap 'rm -rf "$TEST_DIR"' EXIT

# Copy the framework into the test workspace
FRAMEWORK_SRC=$(cd "$(dirname "$0")/.." && pwd)
cp -r "$FRAMEWORK_SRC" "$TEST_DIR/planifest-framework"

cd "$TEST_DIR"

echo "--- Testing: claude-code ---"
./planifest-framework/setup.sh claude-code

if [ ! -d ".claude/skills" ]; then
    echo "FAIL: .claude/skills directory not created"
    exit 1
fi

if [ ! -f ".claude/skills/planifest-orchestrator/SKILL.md" ]; then
    echo "FAIL: SKILL.md not copied for claude-code"
    exit 1
fi

if [ ! -f "CLAUDE.md" ]; then
    echo "FAIL: CLAUDE.md boot file not created"
    exit 1
fi

if [ ! -f ".claude/commands/feature-pipeline.md" ]; then
    echo "FAIL: Workflow command not created for claude-code"
    exit 1
fi

echo "--- Testing: cursor ---"
./planifest-framework/setup.sh cursor

if [ ! -d ".cursor/skills" ]; then
    echo "FAIL: .cursor/skills directory not created"
    exit 1
fi

if [ ! -f ".cursor/skills/planifest-orchestrator/SKILL.md" ]; then
    echo "FAIL: SKILL.md not copied for cursor"
    exit 1
fi

echo "SUCCESS: All setup.sh tests passed."

