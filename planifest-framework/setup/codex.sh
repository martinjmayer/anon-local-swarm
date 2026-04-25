# OpenAI Codex - tool configuration
# https://openai.com/codex
#
# Skills:    .agents/skills/{name}/SKILL.md       (auto-discovered)
# Workflows: .agents/workflows/{name}.md          (auto-discovered)
# Boot file: AGENTS.md                            (project root)

TOOL_SKILLS_DIR=".agents/skills"
TOOL_WORKFLOWS_DIR=".agents/workflows"

TOOL_BOOT_FILE="AGENTS.md"

TOOL_BOOT_TEMPLATE="planifest-framework/templates/standard-boot.md"

# Enforcement tier — Bash-only adapter; no settings.json wiring (ADR-001, REQ-010)
PLANIFEST_TIER=1b
TOOL_HOOK_ADAPTER_SRC="hooks/adapters/codex.mjs"
TOOL_HOOK_ADAPTER_DEST=".agents/hooks/adapters/codex.mjs"
TOOL_HOOKS_INSTALL_DIR=".agents/hooks"