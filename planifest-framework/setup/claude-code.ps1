# Claude Code - tool configuration
# https://docs.anthropic.com/en/docs/claude-code
#
# Skills:    .claude/skills/{name}/SKILL.md      (auto-discovered)
# Workflows: .claude/commands/{name}.md           (becomes /name slash command)
# Boot file: CLAUDE.md                            (project root)

@{
    SkillsDir    = '.claude\skills'
    WorkflowsDir = '.claude\commands'
    BootFile     = 'CLAUDE.md'
    BootTemplate   = "planifest-framework/templates/standard-boot.md"
    # context-mode MCP routing rules — installed as AGENTS.md when context-mode is configured
    AgentsFile     = 'AGENTS.md'
    AgentsTemplate = "planifest-framework/templates/context-mode-agents.md"
    # context-mode enforcement hooks — installed when --context-mode-mcp is passed
    HooksSrc       = 'hooks/context-mode'
    HooksDir       = '.claude/hooks/context-mode'
    SettingsFile   = '.claude/settings.json'
    # structured telemetry hooks — installed when both flags are passed
    TelemetryHooksSrc = 'hooks/telemetry'
    TelemetryHooksDir = '.claude/hooks/telemetry'
}
