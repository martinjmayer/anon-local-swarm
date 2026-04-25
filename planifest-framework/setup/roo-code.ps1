# Roo Code - tool configuration (Windows PowerShell)
# https://roosoft.com
#
# Roo Code is a VS Code extension (fork of Cline) that uses .roo/ for configuration.
#
# Skills:    .roo/skills/{name}/SKILL.md           (loaded via .roorules context)
# Boot file: .roorules                              (project root - always-on rules file)

$TOOL_SKILLS_DIR = ".roo/skills"
$TOOL_WORKFLOWS_DIR = ""

$TOOL_BOOT_FILE = ".roorules"

$TOOL_BOOT_TEMPLATE = "planifest-framework/templates/standard-boot.md"

# context-mode MCP routing rules — installed when --context-mode-mcp is passed
$TOOL_AGENTS_FILE = ".roo/context-mode.md"
$TOOL_AGENTS_TEMPLATE = "planifest-framework/templates/context-mode-agents.md"

# Enforcement tier — no deterministic hook support in public API; instructions-only (ADR-001, REQ-012)
$PLANIFEST_TIER = 3
