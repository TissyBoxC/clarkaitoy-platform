[CmdletBinding()]
param(
    [string]$OutputDirectory = (Join-Path $PSScriptRoot '..\migration-data'),
    [string]$PostgresContainer = 'sprout-local-postgres-1',
    [string]$Sub2APIContainer = 'sprout-local-sub2api-1',
    [string]$Sub2APIVolume = 'sprout-local_sprout_sub2api_data'
)

$ErrorActionPreference = 'Stop'

function Assert-ContainerRunning {
    param([string]$Name)

    $state = & docker inspect --format '{{.State.Running}}' $Name 2>$null
    if ($LASTEXITCODE -ne 0 -or $state.Trim() -ne 'true') {
        throw "Container is not running: $Name"
    }
}

Assert-ContainerRunning -Name $PostgresContainer
Assert-ContainerRunning -Name $Sub2APIContainer

$outputPath = [System.IO.Path]::GetFullPath($OutputDirectory)
New-Item -ItemType Directory -Force -Path $outputPath | Out-Null

Write-Host 'Exporting sprout_device_platform'
& docker exec $PostgresContainer pg_dump -U sprout -Fc -d sprout_device_platform `
    -f /tmp/sprout_device_platform.dump
if ($LASTEXITCODE -ne 0) { throw 'pg_dump failed for sprout_device_platform' }
& docker cp "${PostgresContainer}:/tmp/sprout_device_platform.dump" `
    (Join-Path $outputPath 'sprout_device_platform.dump')
if ($LASTEXITCODE -ne 0) { throw 'docker cp failed for sprout_device_platform' }

Write-Host 'Exporting sprout_sub2api'
& docker exec $PostgresContainer pg_dump -U sprout -Fc -d sprout_sub2api `
    -f /tmp/sprout_sub2api.dump
if ($LASTEXITCODE -ne 0) { throw 'pg_dump failed for sprout_sub2api' }
& docker cp "${PostgresContainer}:/tmp/sprout_sub2api.dump" `
    (Join-Path $outputPath 'sprout_sub2api.dump')
if ($LASTEXITCODE -ne 0) { throw 'docker cp failed for sprout_sub2api' }

Write-Host 'Exporting Sub2API non-secret application data'
$archivePath = Join-Path $outputPath 'sprout_sub2api_data.tar.gz'
$archivePathInContainer = '/tmp/sprout_sub2api_data.tar.gz'
# config.yaml and .installed belong to one server's credentials. The cloud
# stack regenerates them from .env, so only pricing and runtime assets travel.
& docker exec $Sub2APIContainer sh -c "tar -czf $archivePathInContainer -C /app/data model_pricing.json model_pricing.sha256 pages plugins"
if ($LASTEXITCODE -ne 0) { throw 'tar failed for Sub2API data directory' }
& docker cp "${Sub2APIContainer}:$archivePathInContainer" $archivePath
if ($LASTEXITCODE -ne 0) { throw 'docker cp failed for Sub2API data directory' }

# Record the volume name so operators can verify the exported data came from
# the expected Compose project before uploading it.
Write-Host "Source volume: $Sub2APIVolume"

$checksumPath = Join-Path $outputPath 'SHA256SUMS'
$files = Get-ChildItem -LiteralPath $outputPath -File |
    Where-Object { $_.Name -ne 'SHA256SUMS' } |
    Sort-Object Name
$lines = foreach ($file in $files) {
    $hash = (Get-FileHash -LiteralPath $file.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
    "$hash  $($file.Name)"
}
$checksumText = ($lines -join "`n") + "`n"
[System.IO.File]::WriteAllText($checksumPath, $checksumText, [System.Text.Encoding]::ASCII)

Write-Host "Export complete: $outputPath"
