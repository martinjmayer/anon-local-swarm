# Cline / Roo Code - tool configuration
# https://github.com/cline/cline  |  https://roosoft.com
#
# Skills:    .clinerules/skills/{name}/SKILL.md     (loaded via .clinerules context)
# Workflows: (none - Cline uses .clinerules for persistent instructions)
# Boot file: .clinerules                            (project root - always-on rules file)

TOOL_SKILLS_DIR=".clinerules/skills"
TOOL_WORKFLOWS_DIR=""

TOOL_BOOT_FILE=".clinerules"

TOOL_BOOT_TEMPLATE="planifest-framework/templates/standard-boot.md"

# context-mode MCP routing rules — installed when --context-mode-mcp is passed
TOOL_AGENTS_FILE=".clinerules/context-mode.md"
TOOL_AGENTS_TEMPLATE="planifest-framework/templates/context-mode-agents.md"

# Enforcement tier — native hooks adapter (ADR-001, REQ-009, REQ-013)
PLANIFEST_TIER=1
TOOL_HOOK_ADAPTER_SRC="hooks/adapters/cline.mjs"
TOOL_HOOK_ADAPTER_DEST=".clinerules/hooks/adapters/cline.mjs"
TOOL_HOOKS_INSTALL_DIR=".clinerules/hooks"