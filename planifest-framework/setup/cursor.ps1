# Cursor - tool configuration
# https://docs.cursor.com
#
# Skills:    .cursor/skills/{name}/SKILL.md       (auto-discovered)
# Workflows: embedded in .cursor/rules/*.mdc      (Cursor uses rules, not separate workflows)
# Boot file: .cursor/rules/planifest.mdc

@{
    SkillsDir    = '.cursor\skills'
    WorkflowsDir = ''
    BootFile     = '.cursor\rules\planifest.mdc'
    BootTemplate   = "planifest-framework/templates/cursor-boot.md"
    # context-mode MCP routing rules — installed when --context-mode-mcp is passed
    AgentsFile     = '.cursor\rules\context-mode.mdc'
    AgentsTemplate = "planifest-framework/templates/context-mode-agents.md"
}
