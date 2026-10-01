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

$repoDefinitions = @(
    @{
        Name = 'firmware'
        Path = 'firmware'
        Url = 'https://github.com/TissyBoxC/sprout-firmware.git'
        Revision = 'e65fb9e937c6f243f82254aaeb8f84a7dc7a7f44'
    },
    @{
        Name = 'sub2api_fork'
        Path = 'services/sub2api_fork'
        Url = 'https://github.com/TissyBoxC/sprout-sub2api-fork.git'
        Revision = '00bd9edf13afec3d216fa96289718d0699553515'
    }
)

# This script intentionally keeps the clone URLs explicit because PowerShell
# cannot parse the YAML lock file without adding a runtime dependency.
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
