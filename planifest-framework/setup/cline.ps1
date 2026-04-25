# Cline / Roo Code - tool configuration
# https://github.com/cline/cline  |  https://roosoft.com
#
# Skills:    .clinerules/skills/{name}/SKILL.md     (loaded via .clinerules context)
# Workflows: (none - Cline uses .clinerules for persistent instructions)
# Boot file: .clinerules                            (project root - always-on rules file)

@{
    SkillsDir    = '.clinerules\skills'
    WorkflowsDir = ''
    BootFile     = '.clinerules'
    BootTemplate   = "planifest-framework/templates/standard-boot.md"
    # context-mode MCP routing rules — installed when --context-mode-mcp is passed
    AgentsFile     = '.clinerules\context-mode.md'
    AgentsTemplate = "planifest-framework/templates/context-mode-agents.md"
}
