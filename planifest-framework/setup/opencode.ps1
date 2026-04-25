# OpenCode - tool configuration (Tier 2: Bun/TS plugin shim)
# Windows PowerShell equivalent of setup/opencode.sh

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$OpenCodeJson = Join-Path $ProjectRoot "opencode.json"
$PluginSrc = Join-Path $ScriptDir "hooks\adapters\opencode"
$PluginDest = Join-Path $ProjectRoot ".opencode\plugins\@planifest\opencode-hooks"
$HooksDest = Join-Path $ProjectRoot ".opencode\hooks"

Write-Host ""
Write-Host "  Setting up OpenCode (Tier 2: plugin shim)"

# Verify Bun
if (-not (Get-Command bun -ErrorAction SilentlyContinue)) {
    Write-Host "  ! Warning: bun not found in PATH. Ensure OpenCode is installed and bun is in your PATH."
}

# Copy plugin source
New-Item -ItemType Directory -Force -Path $PluginDest | Out-Null
Copy-Item "$PluginSrc\index.ts" "$PluginDest\index.ts" -Force
Copy-Item "$PluginSrc\package.json" "$PluginDest\package.json" -Force
Write-Host "  + .opencode\plugins\@planifest\opencode-hooks\"

# Copy shared hook scripts
New-Item -ItemType Directory -Force -Path "$HooksDest\enforcement" | Out-Null
New-Item -ItemType Directory -Force -Path "$HooksDest\telemetry" | Out-Null
Get-ChildItem "$ScriptDir\hooks\enforcement\*.mjs" | ForEach-Object {
    Copy-Item $_.FullName "$HooksDest\enforcement\$($_.Name)" -Force
    Write-Host "  + .opencode\hooks\enforcement\$($_.Name)"
}
Get-ChildItem "$ScriptDir\hooks\telemetry\emit-phase-*.mjs" | ForEach-Object {
    Copy-Item $_.FullName "$HooksDest\telemetry\$($_.Name)" -Force
    Write-Host "  + .opencode\hooks\telemetry\$($_.Name)"
}

# Register plugin in opencode.json (idempotent)
$pluginRef = ".opencode/plugins/@planifest/opencode-hooks/index.ts"
if (Test-Path $OpenCodeJson) {
    $json = Get-Content $OpenCodeJson -Raw | ConvertFrom-Json
    if (-not $json.plugins) { $json | Add-Member -MemberType NoteProperty -Name plugins -Value @() }
    if ($json.plugins -notcontains $pluginRef) {
        $json.plugins += $pluginRef
        $json | ConvertTo-Json -Depth 10 | Set-Content $OpenCodeJson
        Write-Host "  ~ opencode.json (plugin registered)"
    } else {
        Write-Host "  - opencode.json (plugin already registered)"
    }
} else {
    @{ plugins = @($pluginRef) } | ConvertTo-Json | Set-Content $OpenCodeJson
    Write-Host "  + opencode.json (created)"
}

Write-Host "  [Planifest] OpenCode hooks installed."
