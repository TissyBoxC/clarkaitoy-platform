[CmdletBinding()]
param(
    [string]$PlatformVersion,
    [string]$GatewayVersion,
    [switch]$SkipBuild,
    [switch]$NoRestart
)

$ErrorActionPreference = 'Stop'
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$composePath = Join-Path $workspaceRoot 'deploy\docker-compose.yml'
$environmentPath = Join-Path $workspaceRoot 'deploy\.env'
$platformVersionPath = Join-Path $workspaceRoot 'VERSION'
$gatewayVersionPath = Join-Path $workspaceRoot 'services\sub2api_fork\backend\cmd\server\VERSION'

function Resolve-Version {
    param(
        [string]$ExplicitVersion,
        [string]$VersionPath,
        [string]$Label
    )

    $version = if ([string]::IsNullOrWhiteSpace($ExplicitVersion)) {
        (Get-Content -LiteralPath $VersionPath -Raw).Trim()
    } else {
        $ExplicitVersion.Trim()
    }

    if ($version -notmatch '^[0-9]+\.[0-9]+\.[0-9]+$') {
        throw "$Label version is invalid: $version"
    }

    return $version
}

function Invoke-DockerCompose {
    param([string[]]$Arguments)

    & docker compose -f $composePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose failed: $($Arguments -join ' ')"
    }
}

function Assert-PublishedRelease {
    param(
        [string]$RepositoryPath,
        [string]$Version,
        [string]$Label
    )

    & git -C $RepositoryPath ls-remote --exit-code --tags origin "refs/tags/v$Version" *> $null
    if ($LASTEXITCODE -ne 0) {
        throw "$Label release tag does not exist: v$Version"
    }
}

function Set-LocalEnvironmentVersion {
    param(
        [string]$Name,
        [string]$Version
    )

    if (-not (Test-Path -LiteralPath $environmentPath)) {
        throw "Local environment file not found: $environmentPath"
    }

    $content = Get-Content -LiteralPath $environmentPath -Raw
    $line = "$Name=$Version"
    $pattern = "(?m)^$([regex]::Escape($Name))=.*$"

    if ($content -match $pattern) {
        $content = [regex]::Replace($content, $pattern, $line)
    } else {
        $content = $content.TrimEnd() + [Environment]::NewLine + $line + [Environment]::NewLine
    }

    Set-Content -LiteralPath $environmentPath -Value $content -Encoding utf8 -NoNewline
}

function Test-ServiceReady {
    param(
        [string]$Name,
        [string]$Url
    )

    $deadline = (Get-Date).AddMinutes(3)
    do {
        try {
            $response = Invoke-WebRequest -Uri $Url -Method Get -TimeoutSec 5
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 300) {
                Write-Host "Ready: $Name ($Url)"
                return
            }
        } catch {
            Start-Sleep -Seconds 3
        }
    } while ((Get-Date) -lt $deadline)

    throw "Service did not become ready within three minutes: $Name ($Url)"
}

$platformVersion = Resolve-Version -ExplicitVersion $PlatformVersion -VersionPath $platformVersionPath -Label 'Platform'
$gatewayVersion = Resolve-Version -ExplicitVersion $GatewayVersion -VersionPath $gatewayVersionPath -Label 'AI gateway'

Assert-PublishedRelease -RepositoryPath $workspaceRoot -Version $platformVersion -Label 'Platform'
Assert-PublishedRelease -RepositoryPath (Join-Path $workspaceRoot 'services\sub2api_fork') -Version $gatewayVersion -Label 'AI gateway'

Set-LocalEnvironmentVersion -Name 'SPROUT_PLATFORM_VERSION' -Version $platformVersion
Set-LocalEnvironmentVersion -Name 'SPROUT_SUB2API_VERSION' -Version $gatewayVersion

$env:SPROUT_PLATFORM_VERSION = $platformVersion
$env:SPROUT_SUB2API_VERSION = $gatewayVersion

$profileArguments = @('--profile', 'ai', '--profile', 'services')

if (-not $SkipBuild) {
    Invoke-DockerCompose -Arguments (
        $profileArguments + @(
            'build',
            '--pull',
            '--build-arg:VITE_API_BASE_URL=',
            'admin_web'
        )
    )
}

if (-not $NoRestart) {
    Invoke-DockerCompose -Arguments ($profileArguments + @('up', '-d', '--remove-orphans'))
    Test-ServiceReady -Name 'device platform' -Url 'http://127.0.0.1:8081/readyz'
    Test-ServiceReady -Name 'voice gateway' -Url 'http://127.0.0.1:8082/readyz'
    Test-ServiceReady -Name 'sub2api' -Url 'http://127.0.0.1:8080/health'
    Test-ServiceReady -Name 'admin console' -Url 'http://127.0.0.1:8083/healthz'
}

Write-Host "Docker synchronized: platform=$platformVersion gateway=$gatewayVersion"
