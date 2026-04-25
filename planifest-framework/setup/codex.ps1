# OpenAI Codex - tool configuration
# https://openai.com/codex
#
# Skills:    .agents/skills/{name}/SKILL.md       (auto-discovered)
# Workflows: .agents/workflows/{name}.md          (auto-discovered)
# Boot file: AGENTS.md                            (project root)

@{
    SkillsDir    = '.agents\skills'
    WorkflowsDir = '.agents\workflows'
    BootFile     = 'AGENTS.md'
    BootTemplate = "planifest-framework/templates/standard-boot.md"
}
