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
        Url = 'https://github.com/TissyBoxC/Clarkaitoy.git'
        Revision = 'ff3eb53'
    },
    @{
        Name = 'sub2api_fork'
        Path = 'services/sub2api_fork'
        Url = 'https://github.com/TissyBoxC/sub2api.git'
        Revision = '42bc7f6cf'
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

    if ($currentRevision -ne $repository.Revision) {
        Write-Warning "Checkout does not match workspace.lock.yaml: $($repository.Name)"
    }
}
