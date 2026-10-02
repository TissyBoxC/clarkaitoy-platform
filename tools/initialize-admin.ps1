[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [string]$Email,

    [Parameter(Mandatory)]
    [string]$DisplayName,

    [Parameter(Mandatory)]
    [string]$Password
)

$ErrorActionPreference = 'Stop'

$workspaceRoot = Split-Path -Parent $PSScriptRoot
$composePath = Join-Path $workspaceRoot 'deploy\docker-compose.yml'
$serviceName = 'device_platform'
$profileArguments = @('--profile', 'ai', '--profile', 'services')

function Invoke-Compose {
    param([string[]]$Arguments)

    & docker compose -f $composePath @profileArguments @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose command failed: $($Arguments -join ' ')"
    }
}

function Get-ServiceEnvironmentValue {
    param([string]$Name)

    $containerID = (& docker compose -f $composePath @profileArguments ps -q $serviceName).Trim()
    if ([string]::IsNullOrWhiteSpace($containerID)) {
        throw "Service container is not running: $serviceName"
    }

    $value = (& docker inspect --format "{{range .Config.Env}}{{println .}}{{end}}" $containerID |
        Where-Object { $_ -like "$Name=*" } |
        Select-Object -First 1)
    if ([string]::IsNullOrWhiteSpace($value)) {
        throw "Required container environment variable is missing: $Name"
    }

    return $value.Substring($Name.Length + 1)
}

$databaseDSN = Get-ServiceEnvironmentValue -Name 'DEVICE_PLATFORM_DATABASE_DSN'
$mfaCredentialKey = Get-ServiceEnvironmentValue -Name 'DEVICE_PLATFORM_MFA_CREDENTIAL_KEY'
$accessTokenSecret = Get-ServiceEnvironmentValue -Name 'DEVICE_PLATFORM_AUTH_ACCESS_TOKEN_SECRET'

$arguments = @(
    'exec',
    '--env', "DEVICE_PLATFORM_DATABASE_DSN=$databaseDSN",
    '--env', "DEVICE_PLATFORM_MFA_CREDENTIAL_KEY=$mfaCredentialKey",
    '--env', "DEVICE_PLATFORM_AUTH_ACCESS_TOKEN_SECRET=$accessTokenSecret",
    '--env', "DEVICE_PLATFORM_ADMIN_PASSWORD=$Password",
    $serviceName,
    '/device-platform-admin',
    'bootstrap',
    '--email', $Email,
    '--display-name', $DisplayName
)

$output = & docker compose -f $composePath @arguments 2>&1
if ($LASTEXITCODE -ne 0) {
    throw "Administrator initialization failed: $($output -join [Environment]::NewLine)"
}

$output
