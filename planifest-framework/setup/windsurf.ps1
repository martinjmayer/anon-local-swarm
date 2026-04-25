# Windsurf - tool configuration
# https://docs.windsurf.com
#
# Skills:    .windsurf/skills/{name}/SKILL.md       (auto-discovered via memories/rules)
# Workflows: (none - Windsurf uses rules, not a separate workflow directory)
# Boot file: .windsurfrules                         (project root - always-on rules file)

@{
    SkillsDir    = '.windsurf\skills'
    WorkflowsDir = ''
    BootFile     = '.windsurfrules'
    BootTemplate   = "planifest-framework/templates/standard-boot.md"
    # context-mode MCP routing rules — installed when --context-mode-mcp is passed
    AgentsFile     = '.windsurf\rules\context-mode.md'
    AgentsTemplate = "planifest-framework/templates/context-mode-agents.md"
}
