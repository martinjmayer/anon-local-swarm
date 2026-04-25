# Claude Code - tool configuration
# https://docs.anthropic.com/en/docs/claude-code
#
# Skills:    .claude/skills/{name}/SKILL.md      (auto-discovered)
# Workflows: .claude/commands/{name}.md           (becomes /name slash command)
# Boot file: CLAUDE.md                            (project root)

TOOL_SKILLS_DIR=".claude/skills"
TOOL_WORKFLOWS_DIR=".claude/commands"

TOOL_BOOT_FILE="CLAUDE.md"

TOOL_BOOT_TEMPLATE="planifest-framework/templates/standard-boot.md"

# context-mode MCP routing rules — installed as AGENTS.md when context-mode is configured
TOOL_AGENTS_FILE="AGENTS.md"
TOOL_AGENTS_TEMPLATE="planifest-framework/templates/context-mode-agents.md"

# context-mode enforcement hooks — installed when --context-mode-mcp is passed
# Source scripts live here (within the framework); copied to TOOL_HOOKS_DIR at setup time
TOOL_HOOKS_SRC="hooks/context-mode"
TOOL_HOOKS_DIR=".claude/hooks/context-mode"
TOOL_SETTINGS_FILE=".claude/settings.json"

# structured telemetry hooks — installed when both --structured-telemetry-mcp and --context-mode-mcp are passed
TOOL_TELEMETRY_HOOKS_SRC="hooks/telemetry"
TOOL_TELEMETRY_HOOKS_DIR=".claude/hooks/telemetry"