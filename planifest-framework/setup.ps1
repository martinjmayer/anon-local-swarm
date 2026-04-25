<#
.SYNOPSIS
    Planifest Setup - Configures skills for your agentic coding tool.

.DESCRIPTION
    Copies Planifest skills into the directory structure each coding tool expects.
    Each tool's specific config lives in setup/<tool>.ps1.
    This script handles shared logic only.

.EXAMPLE
    .\planifest-framework\setup.ps1 claude-code
    .\planifest-framework\setup.ps1 claude-code --context-mode-mcp
    .\planifest-framework\setup.ps1 all
#>

# Manual arg parsing — supports --flag style for cross-platform consistency
$Tool = $null
$ContextModeMcp = $false
$StructuredTelemetryMcp = $false
$BackendUrl = 'http://localhost:3741'
$i = 0
while ($i -lt $args.Count) {
    switch ($args[$i]) {
        '--context-mode-mcp'          { $ContextModeMcp = $true; $i++ }
        '--structured-telemetry-mcp'  { $StructuredTelemetryMcp = $true; $i++ }
        '--backend-url' {
            $i++
            if ($i -ge $args.Count) { Write-Host "Error: --backend-url requires a value"; exit 1 }
            $BackendUrl = $args[$i]; $i++
        }
        default {
            if ($args[$i] -like '-*') { Write-Host "Unknown flag: $($args[$i])"; exit 1 }
            else { $Tool = $args[$i]; $i++ }
        }
    }
}

$ErrorActionPreference = 'Stop'

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$SkillsSrc = Join-Path $ScriptDir 'skills'
$WorkflowsSrc = Join-Path $ScriptDir 'workflows'
$SetupDir = Join-Path $ScriptDir 'setup'

$ValidTools = @('claude-code', 'cursor', 'codex', 'antigravity', 'copilot', 'windsurf', 'cline')

# --- Shared functions ---

function Copy-PlanifestSkills {
    param($TargetDir)

    Get-ChildItem -Path $SkillsSrc -Directory | ForEach-Object {
        $skillName = $_.Name
        $srcDir = $_.FullName
        $destDir = Join-Path $TargetDir $skillName
        
        $srcSkillMd = Join-Path $srcDir "SKILL.md"
        if (Test-Path $srcSkillMd) {
            New-Item -ItemType Directory -Path $destDir -Force | Out-Null
            Copy-Item -Path $srcSkillMd -Destination $destDir -Force

            # Rewrite relative paths to match bundled directory structure
            $skillMdPath = Join-Path $destDir "SKILL.md"
            $skillContent = Get-Content -Path $skillMdPath -Raw
            $skillContent = $skillContent -replace '\.\./templates/', './assets/templates/'
            $skillContent = $skillContent -replace '\.\./standards/reference/', './references/reference/'
            $skillContent = $skillContent -replace '\.\./standards/', './references/'
            $skillContent = $skillContent -replace '\.\./schemas/', './assets/schemas/'
            Set-Content -Path $skillMdPath -Value $skillContent -NoNewline -Encoding UTF8

            Write-Host "  + $skillName/SKILL.md"
            
            foreach ($optDir in @('scripts', 'assets', 'references')) {
                $srcOptDir = Join-Path $srcDir $optDir
                if (Test-Path $srcOptDir) {
                    Copy-Item -Path $srcOptDir -Destination $destDir -Recurse -Force
                }
            }

            # Parse bundle_templates and bundle_standards from SKILL.md frontmatter
            $rawContent = Get-Content -Path $srcSkillMd -Raw
            $bundleTemplates = @()
            $bundleStandards = @()
            if ($rawContent -match '(?m)^bundle_templates:\s*\[([^\]]*)\]') {
                $bundleTemplates = $Matches[1].Trim() -split '\s*,\s*' | Where-Object { $_ }
            }
            if ($rawContent -match '(?m)^bundle_standards:\s*\[([^\]]*)\]') {
                $bundleStandards = $Matches[1].Trim() -split '\s*,\s*' | Where-Object { $_ }
            }

            # Bundle only declared templates (or all if no manifest found)
            $templatesSrc = Join-Path $ScriptDir "templates"
            if (Test-Path $templatesSrc) {
                $destTemplates = Join-Path $destDir "assets\templates"
                New-Item -ItemType Directory -Path $destTemplates -Force | Out-Null
                if ($bundleTemplates.Count -gt 0) {
                    foreach ($tpl in $bundleTemplates) {
                        $tplPath = Join-Path $templatesSrc $tpl
                        if (Test-Path $tplPath) {
                            Copy-Item -Path $tplPath -Destination $destTemplates -Force
                        }
                    }
                    Write-Host "    templates: selective ($($bundleTemplates.Count) files)"
                }
                else {
                    Copy-Item -Path "$templatesSrc\*" -Destination $destTemplates -Recurse -Force
                    Write-Host "    templates: all (no manifest)"
                }
            }

            # Always bundle schemas (small, universally needed)
            $schemasSrc = Join-Path $ScriptDir "schemas"
            if (Test-Path $schemasSrc) {
                $destSchemas = Join-Path $destDir "assets\schemas"
                New-Item -ItemType Directory -Path $destSchemas -Force | Out-Null
                Copy-Item -Path "$schemasSrc\*" -Destination $destSchemas -Recurse -Force
            }

            # Bundle only declared standards (or all top-level if no manifest found)
            $standardsSrc = Join-Path $ScriptDir "standards"
            if (Test-Path $standardsSrc) {
                $destRefs = Join-Path $destDir "references"
                New-Item -ItemType Directory -Path $destRefs -Force | Out-Null
                if ($bundleStandards.Count -gt 0) {
                    foreach ($std in $bundleStandards) {
                        $stdPath = Join-Path $standardsSrc $std
                        if (Test-Path $stdPath) {
                            Copy-Item -Path $stdPath -Destination $destRefs -Force
                        }
                    }
                    Write-Host "    standards: selective ($($bundleStandards.Count) files)"
                }
                else {
                    # No manifest - copy all top-level standards (skip reference/ subdirectory)
                    Get-ChildItem -Path $standardsSrc -File | ForEach-Object {
                        Copy-Item -Path $_.FullName -Destination $destRefs -Force
                    }
                    Write-Host "    standards: all top-level (no manifest)"
                }
            }
        }
    }
}

function Write-PlanifestBootFile {
    param($RelPath, $Content)

    $fullPath = Join-Path $ProjectRoot $RelPath
    $dir = Split-Path -Parent $fullPath
    New-Item -ItemType Directory -Path $dir -Force | Out-Null

    if (-not (Test-Path $fullPath)) {
        Set-Content -Path $fullPath -Value $Content -Encoding UTF8
        Write-Host "  + $RelPath (created)"
    }
    else {
        Write-Host "  - $RelPath (already exists, skipped)"
    }
}

function Copy-PlanifestWorkflow {
    param($WorkflowFile, $TargetDir)

    $name = [System.IO.Path]::GetFileNameWithoutExtension($WorkflowFile)
    $destFile = Join-Path $TargetDir "$name.md"

    New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
    Copy-Item -Path $WorkflowFile -Destination $destFile -Force
    Write-Host "  + workflows/$name.md"
}

function Merge-HookSettings {
    # Merge PreToolUse hook entries into .claude/settings.json (REQ-004)
    # Additive merge: existing content preserved; Grep/Bash/WebFetch entries
    # removed then re-added for idempotency on re-run.
    param(
        [string]$SettingsPath,
        [string]$HooksDir   # relative path used in command values
    )

    $newEntries = @(
        @{ matcher = "Grep";     hooks = @(@{ type = "command"; command = "$HooksDir/block-grep.sh" }) }
        @{ matcher = "Bash";     hooks = @(@{ type = "command"; command = "$HooksDir/block-bash.sh" }) }
        @{ matcher = "WebFetch"; hooks = @(@{ type = "command"; command = "$HooksDir/block-webfetch.sh" }) }
    )

    if (Test-Path $SettingsPath) {
        # Additive merge using PowerShell JSON handling
        $existing = Get-Content -Raw -Path $SettingsPath | ConvertFrom-Json

        # Ensure hooks.PreToolUse exists
        if (-not $existing.hooks) {
            $existing | Add-Member -NotePropertyName 'hooks' -NotePropertyValue ([PSCustomObject]@{}) -Force
        }
        if (-not $existing.hooks.PreToolUse) {
            $existing.hooks | Add-Member -NotePropertyName 'PreToolUse' -NotePropertyValue @() -Force
        }

        # Remove existing Grep/Bash/WebFetch entries, then append new ones
        $toRemove = @('Grep', 'Bash', 'WebFetch')
        $filtered = @($existing.hooks.PreToolUse | Where-Object { $toRemove -notcontains $_.matcher })
        $existing.hooks.PreToolUse = $filtered + $newEntries

        $existing | ConvertTo-Json -Depth 10 | Set-Content -Path $SettingsPath -Encoding UTF8
        Write-Host "  ~ .claude/settings.json (context-mode hook entries merged)"
    }
    else {
        $dir = Split-Path -Parent $SettingsPath
        if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }

        $settings = [PSCustomObject]@{
            hooks = [PSCustomObject]@{
                PreToolUse = $newEntries
            }
        }
        $settings | ConvertTo-Json -Depth 10 | Set-Content -Path $SettingsPath -Encoding UTF8
        Write-Host "  + .claude/settings.json (created with context-mode hook entries)"
    }
}

function Install-ContextModeHooks {
    # Copy enforcement hook scripts and wire settings.json (REQ-004)
    param(
        [string]$HooksSrcRel,    # relative to ScriptDir  e.g. hooks/context-mode
        [string]$HooksDirRel,    # relative to ProjectRoot e.g. .claude/hooks/context-mode
        [string]$SettingsRel     # relative to ProjectRoot e.g. .claude/settings.json
    )

    $src      = Join-Path $ScriptDir $HooksSrcRel
    $dest     = Join-Path $ProjectRoot $HooksDirRel
    $settings = Join-Path $ProjectRoot $SettingsRel

    if (-not (Test-Path $src)) {
        Write-Host "  ! Warning: hook scripts not found at $src — skipping hook installation"
        return
    }

    Write-Host ""
    Write-Host "  Installing context-mode enforcement hooks"

    # Create target directory
    New-Item -ItemType Directory -Path $dest -Force | Out-Null

    # Copy each script
    Get-ChildItem -Path $src -Filter '*.sh' | ForEach-Object {
        $destFile = Join-Path $dest $_.Name
        Copy-Item -Path $_.FullName -Destination $destFile -Force
        Write-Host "  + $HooksDirRel/$($_.Name)"
    }

    # Merge settings.json wiring
    Merge-HookSettings -SettingsPath $settings -HooksDir $HooksDirRel
}

function Merge-TelemetryHookSettings {
    # Merge PostToolUse context-pressure hook entry into .claude/settings.json
    # Idempotent: removes existing context-pressure entry before re-adding.
    param(
        [string]$SettingsPath,
        [string]$HooksDir,
        [string]$BackendUrl
    )

    $hookCmd = "PLANIFEST_TELEMETRY_URL=$BackendUrl node $HooksDir/context-pressure.mjs"
    $newEntry = @(
        @{
            matcher = ".*"
            hooks = @(@{ type = "command"; command = $hookCmd; async = $true; timeout = 5000 })
        }
    )

    if (Test-Path $SettingsPath) {
        $existing = Get-Content -Raw -Path $SettingsPath | ConvertFrom-Json

        if (-not $existing.hooks) {
            $existing | Add-Member -NotePropertyName 'hooks' -NotePropertyValue ([PSCustomObject]@{}) -Force
        }
        if (-not $existing.hooks.PostToolUse) {
            $existing.hooks | Add-Member -NotePropertyName 'PostToolUse' -NotePropertyValue @() -Force
        }

        # Remove existing context-pressure entries then append new one
        $filtered = @($existing.hooks.PostToolUse | Where-Object {
            $hooks = $_.hooks
            -not ($hooks | Where-Object { $_.command -match 'context-pressure' })
        })
        $existing.hooks.PostToolUse = $filtered + $newEntry

        $existing | ConvertTo-Json -Depth 10 | Set-Content -Path $SettingsPath -Encoding UTF8
        Write-Host "  ~ .claude/settings.json (telemetry PostToolUse hook merged)"
    }
    else {
        $dir = Split-Path -Parent $SettingsPath
        if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }

        $settings = [PSCustomObject]@{
            hooks = [PSCustomObject]@{ PostToolUse = $newEntry }
        }
        $settings | ConvertTo-Json -Depth 10 | Set-Content -Path $SettingsPath -Encoding UTF8
        Write-Host "  + .claude/settings.json (created with telemetry PostToolUse hook)"
    }
}

function Install-TelemetryHooks {
    # Copy context-pressure hook and wire PostToolUse in settings.json (REQ-008, REQ-010)
    # Only called when both --structured-telemetry-mcp and --context-mode-mcp are active.
    param(
        [string]$HooksSrcRel,    # relative to ScriptDir  e.g. hooks/telemetry
        [string]$HooksDirRel,    # relative to ProjectRoot e.g. .claude/hooks/telemetry
        [string]$SettingsRel,    # relative to ProjectRoot e.g. .claude/settings.json
        [string]$BackendUrl
    )

    $src      = Join-Path $ScriptDir $HooksSrcRel
    $dest     = Join-Path $ProjectRoot $HooksDirRel
    $settings = Join-Path $ProjectRoot $SettingsRel

    if (-not (Test-Path $src)) {
        Write-Host "  ! Warning: telemetry hook scripts not found at $src — skipping"
        return
    }

    Write-Host ""
    Write-Host "  Installing structured telemetry hooks"

    New-Item -ItemType Directory -Path $dest -Force | Out-Null

    Get-ChildItem -Path $src -Filter '*.mjs' | ForEach-Object {
        $destFile = Join-Path $dest $_.Name
        Copy-Item -Path $_.FullName -Destination $destFile -Force
        Write-Host "  + $HooksDirRel/$($_.Name)"
    }

    Merge-TelemetryHookSettings -SettingsPath $settings -HooksDir $HooksDirRel -BackendUrl $BackendUrl
}

function Invoke-PlanifestGuardrails {
    Write-Host ""
    Write-Host "  Activating Planifest Git Guardrails"

    # Point Git to the version-controlled hooks directory
    git config core.hooksPath planifest-framework/hooks
    Write-Host "  + git config core.hooksPath planifest-framework/hooks"

    # Note: chmod is not available on Windows; hooks are made executable by setup.sh on Unix.
    # On Windows, Git for Windows respects the executable bit stored in the repo,
    # so no additional step is required here.

    # Deploy the CI/CD pipeline workflow
    $githubWorkflows = Join-Path $ProjectRoot '.github\workflows'
    $workflowSrc = Join-Path $ScriptDir 'hooks\planifest.yml'
    if (Test-Path $workflowSrc) {
        New-Item -ItemType Directory -Path $githubWorkflows -Force | Out-Null
        $dest = Join-Path $githubWorkflows 'planifest.yml'
        if (-not (Test-Path $dest)) {
            Copy-Item -Path $workflowSrc -Destination $dest -Force
            Write-Host "  + .github/workflows/planifest.yml (created)"
        }
        else {
            Write-Host "  - .github/workflows/planifest.yml (already exists, skipped)"
        }
    }

    # Deploy .gitattributes to enforce LF endings on hook scripts.
    # Without this, Git for Windows re-adds CRLF on checkout, breaking the bash shebang.
    $gitattributesSrc = Join-Path $ScriptDir '.gitattributes'
    $gitattributesDest = Join-Path $ProjectRoot '.gitattributes'
    if (Test-Path $gitattributesSrc) {
        if (-not (Test-Path $gitattributesDest)) {
            Copy-Item -Path $gitattributesSrc -Destination $gitattributesDest -Force
            Write-Host "  + .gitattributes (created - enforces LF on hook scripts)"
        }
        else {
            Write-Host "  - .gitattributes (already exists, skipped)"
        }
    }

    Write-Host "  `u{2705} Git guardrails activated."
}

function Initialize-PlanifestRepo {
    Write-Host ""
    Write-Host "  Initializing Repository Structure"

    $gitignoreSrc = Join-Path $ScriptDir ".gitignore"
    $gitignoreDest = Join-Path $ProjectRoot ".gitignore"
    
    if (Test-Path $gitignoreSrc) {
        if (-not (Test-Path $gitignoreDest)) {
            Copy-Item -Path $gitignoreSrc -Destination $gitignoreDest
            Write-Host "  + .gitignore (copied)"
        }
        else {
            Write-Host "  - .gitignore (already exists at root, skipped)"
        }
    }
    else {
        Write-Host "  ! Warning: .gitignore not found in framework directory ($gitignoreSrc)"
    }

    $srcDir = Join-Path $ProjectRoot "src"
    if (-not (Test-Path $srcDir)) {
        New-Item -ItemType Directory -Path $srcDir -Force | Out-Null
        Write-Host "  + src/ (created)"
    }
    
    $srcReadme = Join-Path $srcDir "README.md"
    if (-not (Test-Path $srcReadme)) {
        Set-Content -Path $srcReadme -Value @'
# src/

Components live here. Each component is a subfolder with a `component.yml` manifest.

See [planifest/spec/feature-structure.md](../planifest/spec/feature-structure.md) for the canonical layout.
'@ -Encoding UTF8
        Write-Host "  + src/README.md (created)"
    }

    $planDir = Join-Path $ProjectRoot "plan"
    if (-not (Test-Path $planDir)) {
        New-Item -ItemType Directory -Path $planDir -Force | Out-Null
        Write-Host "  + plan/ (created)"
    }
    
    $planReadme = Join-Path $planDir "README.md"
    if (-not (Test-Path $planReadme)) {
        Set-Content -Path $planReadme -Value @'
# plan/

Feature specifications live here. Each feature gets a subfolder.

See [plan/feature-structure.md](feature-structure.md) for the canonical layout.
'@ -Encoding UTF8
        Write-Host "  + plan/README.md (created)"
    }

    $planStructure = Join-Path $planDir "feature-structure.md"
    if (-not (Test-Path $planStructure)) {
        Set-Content -Path $planStructure -Value @'
# Planifest - Repository Structure

> The canonical layout for a Planifest-managed repository. Three top-level folders, three concerns.

---

## The Three Folders

```
repo/
+-- planifest-framework/        <- The framework (skills, templates, schemas, standards)
|                                  Drop this in. Don't modify it per-project.
|
+-- plan/                       <- The specifications (organized by feature)
|                                  Plans, briefs, specs, ADRs, risk, scope, glossary.
|                                  Everything that describes WHAT to build and WHY.
|
+-- src/                        <- The code (organized by component)
                                   Implementation, tests, config, manifests.
                                   Everything that IS the built thing.
```

---

## `planifest-framework/` - The Framework

This folder is the Planifest framework itself. It is the same across every project. You do not modify it per-feature - you update it when the framework evolves.

```
planifest/
+-- skills/           <- Agent instructions (orchestrator + phase skills)
+-- templates/        <- File format templates for every artifact
+-- schemas/          <- JSON Schema validation definitions
+-- standards/        <- Code quality standards
+-- spec/             <- This file - the canonical structure definition
```

---

## `plan/` - The Plan/Specifications

Organized by feature. Each feature gets a subfolder. This is where humans write briefs and agents write specs. No code lives here.

```
plan/
+-- {feature-id}/
    +-- feature-brief.md          <- Human input (start here)
    +-- design.md                 <- Validated plan (orchestrator output)
    +-- pipeline-run.md              <- Audit trail (per run)
    +-- pipeline-run-phase-2.md      <- Phase 2 audit (if phased)
    |
    +-- design-requirements.md               <- Functional & non-functional requirements
    +-- design-spec-phase-2.md       <- Phase 2 spec (if phased)
    +-- openapi-spec.yaml            <- API contract
    +-- scope.md                     <- In / Out / Deferred
    +-- risk-register.md             <- Risk items with likelihood & impact
    +-- domain-glossary.md           <- Ubiquitous language
    +-- security-report.md           <- Security review findings
    +-- quirks.md                    <- Quirks and workarounds
    +-- recommendations.md           <- Improvement suggestions
    |
    +-- adr/
        +-- ADR-001-{title}.md       <- Architecture decision records
        +-- ADR-002-{title}.md
        +-- ...
```

### Path Rules - plan/

1. **Feature ID** follows the format `{0000000}-{kebab-case-name}` - a 7-digit zero-padded number prefix for chronological ordering, followed by a human-chosen kebab-case name.
2. **No nesting** - specs, ADRs, and supporting docs are flat within the feature folder. One level of subfolders only (adr/).
3. **No code** - nothing executable lives in `plan/`. If it runs, it belongs in `src/`.
4. **Phased features** append the phase number: `design-spec-phase-2.md`, `pipeline-run-phase-2.md`. The `design.md` is updated per phase, not duplicated.
5. **ADRs** are numbered sequentially. Never renumber. Superseded ADRs stay with `status: superseded`.

---

## `src/` - The Code

Organized by component. Each component is a subfolder at the top level of `src/`. The component manifest lives with the code, not with the plan.

```
src/
+-- {component-id}/
    +-- component.yml               <- Component manifest (from template)
    +-- package.json                  <- (or equivalent for the stack)
    |
    +-- src/                          <- Implementation (structure varies by stack)
    |   +-- ...
    |
    +-- tests/                        <- Tests
    |   +-- ...
    |
    +-- docs/
        +-- data-contract.md          <- Schema ownership & invariants
        +-- migrations/
            +-- proposed-{desc}.md    <- Migration proposals
```

### Path Rules - src/

1. **Component ID** is kebab-case, matches the `id` in `component.yml`.
2. **component.yml is mandatory** - every component has one. Read it before any work; update it after every build.
3. **Component-specific docs** live with the component at `src/{component-id}/docs/`. These describe the component's data contract, migrations, and technical specifics.
4. **Feature-level docs** live in `plan/`. The component's `component.yml` references the feature via the `feature` field.
5. **Existing components** that predate Planifest are retrofitted by adding a `component.yml` at their root.

---

## How the Three Folders Connect

```
plan/current/design.md
    +-- lists component IDs -> src/{component-id}/component.yml
                                    +-- references feature -> plan/

plan/current/design-requirements.md
    +-- functional requirements -> implemented in -> src/{component-id}/src/

plan/current/adr/ADR-001-*.md
    +-- decisions -> followed by -> src/{component-id}/src/

plan/current/openapi-spec.yaml
    +-- API contract -> implemented in -> src/{component-id}/src/
```

The relationship is bidirectional:
- `design.md` lists all component IDs
- Each `component.yml` references its feature ID
- The plan describes WHAT; the code IS the WHAT

---

## Retrofit Ã¢â‚¬â€ Adding Planifest to an Existing Repo

If the repo already has code:

1. Drop `planifest/` into the repo root
2. Create `plan/` for the first feature
3. Move existing components under `src/` (or leave them if they're already there)
4. Add a `component.yml` to each existing component
5. The orchestrator's retrofit mode will read the codebase and infer the existing architecture

---

*Templates for each file are in [planifest/templates/](../templates/). Skills reference these paths.*
'@ -Encoding UTF8
        Write-Host "  + plan/feature-structure.md (created)"
    }

    # Add tool ignore rules to keep context windows lean
    $ignoreContent = @"

# Planifest - Token Reduction (keeps agent semantic search from bloating context)
plan/_archive/
node_modules/
dist/
build/
out/
.next/
"@

    foreach ($ignoreFile in @('.cursorignore', '.claudeignore', '.windsurfignore', '.clineignore')) {
        $ignorePath = Join-Path $ProjectRoot $ignoreFile
        if (-not (Test-Path $ignorePath)) {
            Set-Content -Path $ignorePath -Value $ignoreContent -Encoding UTF8
            Write-Host "  + $ignoreFile (created)"
        }
        else {
            $existing = Get-Content -Path $ignorePath -Raw
            if ($existing -notmatch "Planifest - Token Reduction") {
                Add-Content -Path $ignorePath -Value $ignoreContent -Encoding UTF8
                Write-Host "  + $ignoreFile (appended Planifest ignore rules)"
            }
        }
    }

    # Deploy .cursorindexingignore - excludes large reference docs from semantic
    # search indexing but keeps them accessible via explicit @ mention
    $indexingIgnoreContent = @"

# Planifest - Indexing Exclusions (files accessible via @ mention but excluded from search)
*-evaluation.md
*-guide.md
tool-setup-reference.md
getting-started.md
"@

    $indexingIgnorePath = Join-Path $ProjectRoot ".cursorindexingignore"
    if (-not (Test-Path $indexingIgnorePath)) {
        Set-Content -Path $indexingIgnorePath -Value $indexingIgnoreContent -Encoding UTF8
        Write-Host "  + .cursorindexingignore (created)"
    }
    else {
        $existing = Get-Content -Path $indexingIgnorePath -Raw
        if ($existing -notmatch "Planifest - Indexing Exclusions") {
            Add-Content -Path $indexingIgnorePath -Value $indexingIgnoreContent -Encoding UTF8
            Write-Host "  + .cursorindexingignore (appended Planifest rules)"
        }
    }
}

function Invoke-PlanifestSetup {
    param($ToolName)

    $toolConfigPath = Join-Path $SetupDir "$ToolName.ps1"
    if (-not (Test-Path $toolConfigPath)) {
        Write-Host "Error: no config file at setup/$ToolName.ps1"
        exit 1
    }

    # Load tool-specific config
    $toolConfig = & $toolConfigPath

    $skillsDir = Join-Path $ProjectRoot $toolConfig.SkillsDir

    Write-Host ""
    Write-Host "  Setting up $ToolName"
    Write-Host "  Skills directory: $($toolConfig.SkillsDir)/"

    # Copy skills (now automatically bundles supporting files)
    Copy-PlanifestSkills -TargetDir $skillsDir

    # Copy workflows (if tool defines a workflow dir)
    if ($toolConfig.WorkflowsDir -and (Test-Path $WorkflowsSrc)) {
        $workflowsDir = Join-Path $ProjectRoot $toolConfig.WorkflowsDir
        Get-ChildItem -Path $WorkflowsSrc -Filter '*.md' | ForEach-Object {
            Copy-PlanifestWorkflow -WorkflowFile $_.FullName -TargetDir $workflowsDir
        }
    }

    # Create boot file (if tool defines one)
    if ($toolConfig.BootFile) {
        $bootContent = $toolConfig.BootContent
        if (-not $bootContent -and $toolConfig.BootTemplate) {
            $bootContentPath = Join-Path $ProjectRoot $toolConfig.BootTemplate
            $bootContent = Get-Content -Raw -Path $bootContentPath
        }
        Write-PlanifestBootFile -RelPath $toolConfig.BootFile -Content $bootContent
    }

    # Install context-mode MCP routing rules if --context-mode-mcp flag is set
    if ($ContextModeMcp -and $toolConfig.AgentsFile -and $toolConfig.AgentsTemplate) {
        $agentsContentPath = Join-Path $ProjectRoot $toolConfig.AgentsTemplate
        $agentsContent = Get-Content -Raw -Path $agentsContentPath
        Write-PlanifestBootFile -RelPath $toolConfig.AgentsFile -Content $agentsContent
    }

    # Install context-mode enforcement hooks if --context-mode-mcp flag is set (REQ-004)
    if ($ContextModeMcp -and $toolConfig.HooksSrc -and $toolConfig.HooksDir -and $toolConfig.SettingsFile) {
        Install-ContextModeHooks `
            -HooksSrcRel  $toolConfig.HooksSrc `
            -HooksDirRel  $toolConfig.HooksDir `
            -SettingsRel  $toolConfig.SettingsFile
    }

    # Write telemetry opt-in sentinel so skills know emission is authorised (REQ-004)
    if ($StructuredTelemetryMcp) {
        $sentinel = Join-Path $ProjectRoot '.claude\telemetry-enabled'
        $sentinelDir = Split-Path -Parent $sentinel
        if (-not (Test-Path $sentinelDir)) { New-Item -ItemType Directory -Path $sentinelDir -Force | Out-Null }
        if (-not (Test-Path $sentinel)) {
            New-Item -ItemType File -Path $sentinel -Force | Out-Null
            Write-Host "  + .claude/telemetry-enabled (telemetry opt-in sentinel)"
        } else {
            Write-Host "  - .claude/telemetry-enabled (already exists)"
        }
    }

    # Install telemetry hooks only when BOTH flags are active (REQ-010)
    if ($StructuredTelemetryMcp -and $ContextModeMcp -and
        $toolConfig.TelemetryHooksSrc -and $toolConfig.TelemetryHooksDir -and $toolConfig.SettingsFile) {
        Install-TelemetryHooks `
            -HooksSrcRel  $toolConfig.TelemetryHooksSrc `
            -HooksDirRel  $toolConfig.TelemetryHooksDir `
            -SettingsRel  $toolConfig.SettingsFile `
            -BackendUrl   $BackendUrl
    }

    Write-Host "  Done."
}

# --- Main ---

if (-not $Tool) {
    Write-Host ""
    Write-Host "Planifest Setup"
    Write-Host ""
    Write-Host "Usage: .\planifest-framework\setup.ps1 [tool] [--context-mode-mcp]"
    Write-Host ""
    Write-Host "Tools:"
    foreach ($t in $ValidTools) {
        Write-Host "  $t"
    }
    Write-Host "  all"
    Write-Host ""
    Write-Host "Flags:"
    Write-Host "  --context-mode-mcp           Install context-mode MCP routing rules file"
    Write-Host "                               (only needed if context-mode MCP plugin is installed)"
    Write-Host "                               See: https://github.com/mksglu/context-mode"
    Write-Host "  --structured-telemetry-mcp   Install structured telemetry hooks"
    Write-Host "                               Requires --context-mode-mcp to also be set."
    Write-Host "                               Context-pressure hook installed when both flags are active."
    Write-Host "  --backend-url <url>          Override telemetry backend URL (default: http://localhost:3741)"
    Write-Host ""
    Write-Host "Run from the repository root."
    Write-Host "Each tool's config: planifest-framework\setup\[tool].ps1"
    exit 0
}

Write-Host "Planifest Setup"
Write-Host ("=" * 40)

Initialize-PlanifestRepo
Invoke-PlanifestGuardrails

$ToolLower = $Tool.ToLower()

if ($ToolLower -eq 'all') {
    foreach ($t in $ValidTools) {
        Invoke-PlanifestSetup -ToolName $t
    }
}
elseif ($ValidTools -contains $ToolLower) {
    Invoke-PlanifestSetup -ToolName $ToolLower
}
else {
    Write-Host "Unknown tool: $Tool"
    Write-Host "Valid tools: $($ValidTools -join ', '), all"
    exit 1
}

Write-Host ""
Write-Host "Setup complete."
Write-Host "  Source of truth: planifest-framework/"
Write-Host "  Re-run after updating framework files."

