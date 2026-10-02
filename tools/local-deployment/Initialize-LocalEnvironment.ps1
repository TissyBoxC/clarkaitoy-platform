[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$workspaceRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$examplePath = Join-Path $workspaceRoot 'deploy\.env.example'
$environmentPath = Join-Path $workspaceRoot 'deploy\.env'

if (-not (Test-Path -LiteralPath $examplePath)) {
    throw "Environment template not found: $examplePath"
}

if (Test-Path -LiteralPath $environmentPath) {
    throw "Local environment already exists: $environmentPath"
}

function New-HexSecret {
    param(
        [Parameter(Mandatory)]
        [ValidateRange(16, 128)]
        [int]$ByteCount
    )

    $bytes = New-Object byte[] $ByteCount
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    return [Convert]::ToHexString($bytes).ToLowerInvariant()
}

$replacements = [ordered]@{
    'replace-with-local-database-password' = New-HexSecret -ByteCount 24
    'replace-with-local-redis-password' = New-HexSecret -ByteCount 24
    'replace-with-local-admin-password' = New-HexSecret -ByteCount 24
    'replace-with-local-sub2api-key' = New-HexSecret -ByteCount 24
    'replace-with-at-least-32-random-hex-characters' = New-HexSecret -ByteCount 32
    'replace-with-a-different-32-character-secret' = New-HexSecret -ByteCount 32
}

$environmentLines = Get-Content -LiteralPath $examplePath
$renderedLines = foreach ($line in $environmentLines) {
    $outputLine = $line
    foreach ($placeholder in $replacements.Keys) {
        if ($outputLine -eq "$($placeholder.Split('=')[0])=") {
            continue
        }
        $outputLine = $outputLine.Replace($placeholder, $replacements[$placeholder])
    }
    $outputLine
}

$renderedLines | Set-Content -LiteralPath $environmentPath -Encoding utf8

Write-Host "Created local environment file: $environmentPath"
