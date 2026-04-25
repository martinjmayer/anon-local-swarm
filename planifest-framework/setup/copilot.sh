# GitHub Copilot - tool configuration
# https://docs.github.com/en/copilot
#
# Skills:    .github/skills/{name}/SKILL.md       (auto-discovered)
# Workflows: .github/copilot-workflows/{name}.md   (natural language workflows - avoids GitHub Actions conflict)
# Boot file: .github/copilot-instructions.md

TOOL_SKILLS_DIR=".github/skills"
TOOL_WORKFLOWS_DIR=".github/copilot-workflows"

TOOL_BOOT_FILE=".github/copilot-instructions.md"

TOOL_BOOT_TEMPLATE="planifest-framework/templates/standard-boot.md"

# context-mode MCP routing rules — installed when --context-mode-mcp is passed
TOOL_AGENTS_FILE=".github/instructions/context-mode.md"
TOOL_AGENTS_TEMPLATE="planifest-framework/templates/context-mode-agents.md"

# Enforcement tier — no native hook system; instructions-only (ADR-001, REQ-012)
PLANIFEST_TIER=3