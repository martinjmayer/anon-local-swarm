$ErrorActionPreference = 'Stop'

Write-Host "Running setup.ps1 tests..."

# Create a temporary test workspace
$tempDirPattern = "planifest_test_" + [guid]::NewGuid().ToString()
$testDir = Join-Path $env:TEMP $tempDirPattern
New-Item -ItemType Directory -Path $testDir -Force | Out-Null

$originalLocation = Get-Location

try {
    # Copy the framework into the test workspace
    $scriptPath = $MyInvocation.MyCommand.Path
    $frameworkSrc = Split-Path (Split-Path $scriptPath -Parent) -Parent
    
    Copy-Item -Path $frameworkSrc -Destination (Join-Path $testDir "planifest-framework") -Recurse -Force
    
    Set-Location $testDir

    Write-Host "--- Testing: claude-code ---"
    & .\planifest-framework\setup.ps1 claude-code

    if (-not (Test-Path ".claude\skills")) {
        throw "FAIL: .claude\skills directory not created"
    }

    if (-not (Test-Path ".claude\skills\planifest-orchestrator\SKILL.md")) {
        throw "FAIL: SKILL.md not copied for claude-code"
    }

    if (-not (Test-Path "CLAUDE.md")) {
        throw "FAIL: CLAUDE.md boot file not created"
    }

    if (-not (Test-Path ".claude\commands\feature-pipeline.md")) {
        throw "FAIL: Workflow command not created for claude-code"
    }

    Write-Host "--- Testing: cursor ---"
    & .\planifest-framework\setup.ps1 cursor

    if (-not (Test-Path ".cursor\skills")) {
        throw "FAIL: .cursor\skills directory not created"
    }

    if (-not (Test-Path ".cursor\skills\planifest-orchestrator\SKILL.md")) {
        throw "FAIL: SKILL.md not copied for cursor"
    }

    Write-Host "SUCCESS: All setup.ps1 tests passed."
}
finally {
    Set-Location $originalLocation
    if (Test-Path $testDir) {
        Remove-Item -Path $testDir -Recurse -Force
    }
}

