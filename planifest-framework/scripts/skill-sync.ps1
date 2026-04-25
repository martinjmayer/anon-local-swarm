# Planifest skill-sync — install, remove, and sync external skills (REQ-024, REQ-025)
# PowerShell equivalent of skill-sync.sh
#
# Usage:
#   skill-sync.ps1 add     <skill-name> <tool> [-From <url>] [-Authorized]
#   skill-sync.ps1 install <skill-name> <tool>
#   skill-sync.ps1 remove  <skill-name> <tool>
#   skill-sync.ps1 sync    <tool>
#   skill-sync.ps1 preserve   <skill-name> <tool>
#   skill-sync.ps1 unpreserve <skill-name> <tool>

param(
  [Parameter(Mandatory)][string]$Operation,
  [string]$SkillName,
  [string]$Tool,
  [string]$From,
  [switch]$Authorized
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$ScriptDir         = Split-Path -Parent $MyInvocation.MyCommand.Path
$FrameworkDir      = Split-Path -Parent $ScriptDir
$ProjectRoot       = Split-Path -Parent $FrameworkDir
$SetupDir          = Join-Path $FrameworkDir "setup"
$Manifest          = Join-Path $FrameworkDir "external-skills.json"
$PlanSkillsDir     = Join-Path $ProjectRoot "plan\current\external-skills"
$PreservedSkillsDir= Join-Path $FrameworkDir "external-skills"
$AnthropicRawBase  = "https://raw.githubusercontent.com/anthropics/skills/main/skills"

function Die([string]$Msg) {
  Write-Error "  [skill-sync] Error: $Msg"
  exit 1
}

function Info([string]$Msg) {
  Write-Host "  [skill-sync] $Msg"
}

function Resolve-ToolSkillsDir([string]$ToolName) {
  $config = Join-Path $SetupDir "$ToolName.ps1"
  if (-not (Test-Path $config)) { Die "Unknown tool '$ToolName'." }
  $line = (Get-Content $config | Where-Object { $_ -match '^\$TOOL_SKILLS_DIR\s*=' } | Select-Object -First 1)
  if (-not $line) { Die "TOOL_SKILLS_DIR not set in $config" }
  $rel = $line -replace '^\$TOOL_SKILLS_DIR\s*=\s*["\x27]?([^"\\x27\s]+)["\x27]?.*','$1'
  return Join-Path $ProjectRoot $rel
}

function Detect-Tool {
  foreach ($config in (Get-ChildItem "$SetupDir\*.ps1")) {
    $toolName = [System.IO.Path]::GetFileNameWithoutExtension($config.Name)
    $line = (Get-Content $config | Where-Object { $_ -match '^\$TOOL_SKILLS_DIR\s*=' } | Select-Object -First 1)
    if (-not $line) { continue }
    $rel = $line -replace '^\$TOOL_SKILLS_DIR\s*=\s*["\x27]?([^"\\x27\s]+)["\x27]?.*','$1'
    if (Test-Path (Join-Path $ProjectRoot $rel)) { return $toolName }
  }
  Die "No tool skills directory detected. Run setup.ps1 <tool> first."
}

function Ensure-Manifest {
  if (-not (Test-Path $Manifest)) {
    '{"skills":[]}' | Set-Content -Path $Manifest -Encoding UTF8
  }
}

function Invoke-Node([string]$Script) {
  $result = node -e $Script 2>&1
  return $result
}

function Skill-InManifest([string]$Name) {
  if (-not (Test-Path $Manifest)) { return $false }
  $m = Get-Content $Manifest -Raw | ConvertFrom-Json
  return ($m.skills | Where-Object { $_.name -eq $Name }).Count -gt 0
}

function Get-SkillScope([string]$Name) {
  if (-not (Test-Path $Manifest)) { return "" }
  $m = Get-Content $Manifest -Raw | ConvertFrom-Json
  $entry = $m.skills | Where-Object { $_.name -eq $Name } | Select-Object -First 1
  if ($entry) { return $entry.scope } else { return "" }
}

function Add-ToManifest([string]$Name, [string]$Source, [bool]$Trusted, [string]$Scope, [string]$FeatureId = "") {
  Ensure-Manifest
  $m = Get-Content $Manifest -Raw | ConvertFrom-Json
  $m.skills = @($m.skills | Where-Object { $_.name -ne $Name })
  $entry = [PSCustomObject]@{
    name        = $Name
    source      = $Source
    trusted     = $Trusted
    installedAt = (Get-Date -Format "yyyy-MM-dd")
    scope       = $Scope
  }
  if ($FeatureId) { $entry | Add-Member -NotePropertyName featureId -NotePropertyValue $FeatureId }
  $m.skills = @($m.skills) + $entry
  $m | ConvertTo-Json -Depth 5 | Set-Content -Path $Manifest -Encoding UTF8
}

function Remove-FromManifest([string]$Name) {
  if (-not (Test-Path $Manifest)) { return }
  $m = Get-Content $Manifest -Raw | ConvertFrom-Json
  $m.skills = @($m.skills | Where-Object { $_.name -ne $Name })
  if ($m.skills.Count -eq 0) { Remove-Item $Manifest -Force }
  else { $m | ConvertTo-Json -Depth 5 | Set-Content -Path $Manifest -Encoding UTF8 }
}

function Update-Scope([string]$Name, [string]$NewScope) {
  $m = Get-Content $Manifest -Raw | ConvertFrom-Json
  $entry = $m.skills | Where-Object { $_.name -eq $Name } | Select-Object -First 1
  if ($entry) { $entry.scope = $NewScope }
  $m | ConvertTo-Json -Depth 5 | Set-Content -Path $Manifest -Encoding UTF8
}

function Fetch-Skill([string]$Name, [string]$SourceUrl, [string]$Scope) {
  $destDir = if ($Scope -eq "preserved") { Join-Path $PreservedSkillsDir $Name } else { Join-Path $PlanSkillsDir $Name }
  New-Item -ItemType Directory -Path $destDir -Force | Out-Null

  $rawUrl = "$AnthropicRawBase/$Name/SKILL.md"
  if ($SourceUrl -notmatch "anthropics/skills" -and $SourceUrl -notmatch "raw.githubusercontent.com") {
    $rawUrl = "$SourceUrl/SKILL.md"
  }

  try {
    Invoke-WebRequest -Uri $rawUrl -OutFile (Join-Path $destDir "SKILL.md") -UseBasicParsing
    Info "Fetched: $rawUrl"
  } catch {
    Remove-Item $destDir -Recurse -Force -ErrorAction SilentlyContinue
    Die "Failed to fetch skill '$Name' from $rawUrl — check the skill name and network."
  }
}

function Cmd-Install([string]$Name, [string]$ToolName) {
  $toolSkillsDir = Resolve-ToolSkillsDir $ToolName
  $dest = Join-Path $toolSkillsDir $Name

  $src = ""
  if (Test-Path (Join-Path $PreservedSkillsDir $Name)) { $src = Join-Path $PreservedSkillsDir $Name }
  elseif (Test-Path (Join-Path $PlanSkillsDir $Name))  { $src = Join-Path $PlanSkillsDir $Name }
  else { Die "Skill '$Name' not found in either storage tier. Run: skill-sync.ps1 add $Name $ToolName" }

  New-Item -ItemType Directory -Path $dest -Force | Out-Null
  Copy-Item "$src\*" -Destination $dest -Recurse -Force
  Info "Installed: $Name → $dest"
}

function Cmd-Remove([string]$Name, [string]$ToolName) {
  $toolSkillsDir = Resolve-ToolSkillsDir $ToolName
  $dest = Join-Path $toolSkillsDir $Name
  if (Test-Path $dest) { Remove-Item $dest -Recurse -Force; Info "Removed from tool skills: $Name" }
  else { Info "Warning: $Name not found in $toolSkillsDir — skipping" }
}

function Cmd-Sync([string]$ToolName) {
  if (-not (Test-Path $Manifest)) { Info "No manifest found — nothing to sync."; return }
  $m = Get-Content $Manifest -Raw | ConvertFrom-Json
  if ($m.skills.Count -eq 0) { Info "Manifest is empty — nothing to sync."; return }
  foreach ($entry in $m.skills) {
    if ($entry.scope -eq "plan" -and -not (Test-Path (Join-Path $PlanSkillsDir $entry.name))) {
      Info "Re-fetching plan-scoped skill: $($entry.name)"
      Fetch-Skill $entry.name $entry.source "plan"
    }
    Cmd-Install $entry.name $ToolName
  }
}

function Cmd-Add([string]$Name, [string]$ToolName) {
  if (Skill-InManifest $Name) { Info "Warning: '$Name' is already installed — skipping."; return }

  $trusted = $false; $sourceUrl = ""
  if (-not $From) {
    $trusted   = $true
    $sourceUrl = "https://github.com/anthropics/skills/tree/main/skills/$Name"
  } else {
    $sourceUrl = $From
    if (-not $Authorized) {
      Die "Non-Anthropic source requires human approval.`n  Source: $From`n  Skill:  $Name`n  Re-run with -Authorized once the human has confirmed."
    }
  }

  $featureId = ""
  $fidFile = Join-Path $ProjectRoot "plan\current\.feature-id"
  if (Test-Path $fidFile) { $featureId = (Get-Content $fidFile -Raw).Trim() }

  Info "Fetching skill: $Name"
  Fetch-Skill $Name $sourceUrl "plan"
  Ensure-Manifest
  Add-ToManifest $Name $sourceUrl $trusted "plan" $featureId
  Info "Recorded in manifest."
  Cmd-Install $Name $ToolName
  Info "Skill '$Name' installed for tool '$ToolName'."
}

function Cmd-Preserve([string]$Name, [string]$ToolName) {
  if (-not (Skill-InManifest $Name)) { Die "Skill '$Name' not found in manifest." }
  if ((Get-SkillScope $Name) -eq "preserved") { Info "'$Name' is already preserved."; return }
  $src = Join-Path $PlanSkillsDir $Name
  if (Test-Path $src) {
    New-Item -ItemType Directory -Path $PreservedSkillsDir -Force | Out-Null
    Copy-Item $src -Destination (Join-Path $PreservedSkillsDir $Name) -Recurse -Force
    Remove-Item $src -Recurse -Force
  }
  Update-Scope $Name "preserved"
  Info "'$Name' is now preserved — will survive P7 archive."
}

function Cmd-Unpreserve([string]$Name, [string]$ToolName) {
  if (-not (Skill-InManifest $Name)) { Die "Skill '$Name' not found in manifest." }
  if ((Get-SkillScope $Name) -eq "plan") { Info "'$Name' is already plan-scoped."; return }
  $src = Join-Path $PreservedSkillsDir $Name
  if (Test-Path $src) {
    New-Item -ItemType Directory -Path $PlanSkillsDir -Force | Out-Null
    Copy-Item $src -Destination (Join-Path $PlanSkillsDir $Name) -Recurse -Force
    Remove-Item $src -Recurse -Force
  }
  Update-Scope $Name "plan"
  Info "'$Name' will be removed at P7 archive."
}

# ── Resolve tool ────────────────────────────────────────────────────────────
if ($Operation -eq "sync") {
  if (-not $Tool) { $Tool = Detect-Tool }
  Cmd-Sync $Tool; exit 0
}
if (-not $SkillName) { Die "Skill name required." }
if (-not $Tool)      { $Tool = Detect-Tool }

switch ($Operation) {
  "add"        { Cmd-Add        $SkillName $Tool }
  "install"    { Cmd-Install    $SkillName $Tool }
  "remove"     { Cmd-Remove     $SkillName $Tool }
  "preserve"   { Cmd-Preserve   $SkillName $Tool }
  "unpreserve" { Cmd-Unpreserve $SkillName $Tool }
  default      { Die "Unknown operation '$Operation'. Valid: add install remove sync preserve unpreserve" }
}
