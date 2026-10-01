[CmdletBinding()]
param(
    [switch]$Update
)

$ErrorActionPreference = 'Stop'
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$lockFilePath = Join-Path $workspaceRoot 'workspace.lock.yaml'

if (-not (Test-Path -LiteralPath $lockFilePath)) {
    throw "Workspace lock file not found: $lockFilePath"
}

$repoDefinitions = @()
$currentRepository = $null
$insideRepositories = $false

# The lock file has a small, fixed schema. Parsing only the repository block
# avoids adding a YAML runtime dependency to a lightweight checkout script.
foreach ($line in Get-Content -LiteralPath $lockFilePath) {
    if ($line -match '^repositories:\s*$') {
        $insideRepositories = $true
        continue
    }
    if ($insideRepositories -and $line -match '^\S') {
        break
    }
    if (-not $insideRepositories -or $line -match '^\s*#') {
        continue
    }

    if ($line -match '^  ([a-z0-9_]+):\s*$') {
        if ($null -ne $currentRepository) {
            $repoDefinitions += $currentRepository
        }
        $currentRepository = @{ Name = $Matches[1] }
        continue
    }
    if ($null -eq $currentRepository) {
        continue
    }
    if ($line -match '^\s{4}(path|url|branch|revision):\s*(.+?)\s*$') {
        $currentRepository[$Matches[1]] = $Matches[2].Trim('"', "'")
    }
}

if ($null -ne $currentRepository) {
    $repoDefinitions += $currentRepository
}

foreach ($repository in $repoDefinitions) {
    if (-not $repository.ContainsKey('path') -or
        -not $repository.ContainsKey('url') -or
        -not $repository.ContainsKey('revision')) {
        throw "Workspace lock entry is incomplete: $($repository.Name)"
    }
}

foreach ($repository in $repoDefinitions) {
    $destination = Join-Path $workspaceRoot $repository.Path

    if (-not (Test-Path -LiteralPath $destination)) {
        Write-Host "Cloning $($repository.Name) into $destination"
        git clone --branch main $repository.Url $destination
    } elseif ($Update) {
        Write-Host "Fetching $($repository.Name)"
        git -C $destination fetch origin
    } else {
        Write-Host "Keeping existing checkout: $($repository.Name)"
    }

    $currentRevision = (git -C $destination rev-parse HEAD).Trim()
    Write-Host "Locked revision: $($repository.Revision)"
    Write-Host "Current revision: $currentRevision"

    if (-not $currentRevision.StartsWith(
        $repository.Revision,
        [System.StringComparison]::OrdinalIgnoreCase
    )) {
        Write-Warning "Checkout does not match workspace.lock.yaml: $($repository.Name)"
    }
}
