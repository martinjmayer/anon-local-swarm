# Google Antigravity - tool configuration
# https://antigravity.google
#
# Skills:    .gemini/skills/{name}/SKILL.md       (auto-discovered)
# Workflows: .agent/workflows/{name}.md           (becomes /name slash command)
# Boot file: GEMINI.md                            (project root)

@{
    SkillsDir    = '.gemini\skills'
    WorkflowsDir = '.agent\workflows'
    BootFile     = 'GEMINI.md'
    BootTemplate   = "planifest-framework/templates/standard-boot.md"
    # context-mode MCP routing rules — installed when --context-mode-mcp is passed
    AgentsFile     = '.gemini\context-mode.md'
    AgentsTemplate = "planifest-framework/templates/context-mode-agents.md"
}
