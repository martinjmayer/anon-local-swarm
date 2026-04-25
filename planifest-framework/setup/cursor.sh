# Cursor - tool configuration
# https://docs.cursor.com
#
# Skills:    .cursor/skills/{name}/SKILL.md       (auto-discovered)
# Workflows: embedded in .cursor/rules/*.mdc      (Cursor uses rules, not a separate workflow dir)
# Boot file: .cursor/rules/planifest.mdc

TOOL_SKILLS_DIR=".cursor/skills"
TOOL_WORKFLOWS_DIR=""

TOOL_BOOT_FILE=".cursor/rules/planifest.mdc"

TOOL_BOOT_TEMPLATE="planifest-framework/templates/cursor-boot.md"

# context-mode MCP routing rules — installed when --context-mode-mcp is passed
TOOL_AGENTS_FILE=".cursor/rules/context-mode.mdc"
TOOL_AGENTS_TEMPLATE="planifest-framework/templates/context-mode-agents.md"

# Enforcement tier — native hooks adapter (ADR-001, REQ-009, REQ-013)
PLANIFEST_TIER=1
TOOL_HOOK_ADAPTER_SRC="hooks/adapters/cursor.mjs"
TOOL_HOOK_ADAPTER_DEST=".cursor/hooks/adapters/cursor.mjs"
TOOL_HOOKS_INSTALL_DIR=".cursor/hooks"