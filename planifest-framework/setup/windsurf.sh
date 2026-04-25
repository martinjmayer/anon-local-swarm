# Windsurf - tool configuration
# https://docs.windsurf.com
#
# Skills:    .windsurf/skills/{name}/SKILL.md       (auto-discovered via memories/rules)
# Workflows: (none - Windsurf uses rules, not a separate workflow directory)
# Boot file: .windsurfrules                         (project root - always-on rules file)

TOOL_SKILLS_DIR=".windsurf/skills"
TOOL_WORKFLOWS_DIR=""

TOOL_BOOT_FILE=".windsurfrules"

TOOL_BOOT_TEMPLATE="planifest-framework/templates/standard-boot.md"

# context-mode MCP routing rules — installed when --context-mode-mcp is passed
TOOL_AGENTS_FILE=".windsurf/rules/context-mode.md"
TOOL_AGENTS_TEMPLATE="planifest-framework/templates/context-mode-agents.md"

# Enforcement tier — native hooks adapter (ADR-001, REQ-009, REQ-013)
PLANIFEST_TIER=1
TOOL_HOOK_ADAPTER_SRC="hooks/adapters/windsurf.mjs"
TOOL_HOOK_ADAPTER_DEST=".windsurf/hooks/adapters/windsurf.mjs"
TOOL_HOOKS_INSTALL_DIR=".windsurf/hooks"