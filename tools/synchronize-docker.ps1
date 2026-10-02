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

$env:SPROUT_PLATFORM_VERSION = $platformVersion
$env:SPROUT_SUB2API_VERSION = $gatewayVersion

$profileArguments = @('--profile', 'ai', '--profile', 'services')

if (-not $SkipBuild) {
    Invoke-DockerCompose -Arguments ($profileArguments + @('build', '--pull'))
}

if (-not $NoRestart) {
    Invoke-DockerCompose -Arguments ($profileArguments + @('up', '-d', '--remove-orphans'))
    Test-ServiceReady -Name 'device platform' -Url 'http://127.0.0.1:8081/readyz'
    Test-ServiceReady -Name 'voice gateway' -Url 'http://127.0.0.1:8082/readyz'
    Test-ServiceReady -Name 'sub2api' -Url 'http://127.0.0.1:8080/health'
    Test-ServiceReady -Name 'admin console' -Url 'http://127.0.0.1:8083/healthz'
}

Write-Host "Docker synchronized: platform=$platformVersion gateway=$gatewayVersion"
