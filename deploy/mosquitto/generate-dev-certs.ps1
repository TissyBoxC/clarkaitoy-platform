[CmdletBinding()]
param(
    [string]$OutputDirectory = (Join-Path $PSScriptRoot "certs")
)

$ErrorActionPreference = "Stop"

function Resolve-OpenSSL {
    $openssl = Get-Command openssl -ErrorAction SilentlyContinue
    if ($null -ne $openssl) {
        return $openssl.Source
    }

    $gitExecPath = git --exec-path
    if ($LASTEXITCODE -eq 0 -and $gitExecPath) {
        $gitRoot = Split-Path -Parent (Split-Path -Parent $gitExecPath)
        $gitOpenSSL = Join-Path $gitRoot "bin\openssl.exe"
        if (Test-Path -LiteralPath $gitOpenSSL -PathType Leaf) {
            return $gitOpenSSL
        }
    }

    $gitOpenSSL = Join-Path ${env:ProgramFiles} "Git\usr\bin\openssl.exe"
    if (Test-Path -LiteralPath $gitOpenSSL -PathType Leaf) {
        return $gitOpenSSL
    }

    throw "OpenSSL was not found. Install Git for Windows or add openssl to PATH."
}

function Invoke-OpenSSL {
    param(
        [Parameter(Mandatory)]
        [string]$OpenSSL,

        [Parameter(Mandatory)]
        [string[]]$Arguments
    )

    & $OpenSSL @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "OpenSSL failed with exit code $LASTEXITCODE."
    }
}

$openssl = Resolve-OpenSSL
$resolvedOutputDirectory = [System.IO.Path]::GetFullPath($OutputDirectory)
New-Item -ItemType Directory -Path $resolvedOutputDirectory -Force | Out-Null

$caKey = Join-Path $resolvedOutputDirectory "ca.key"
$caCertificate = Join-Path $resolvedOutputDirectory "ca.crt"
$serverKey = Join-Path $resolvedOutputDirectory "server.key"
$serverRequest = Join-Path $resolvedOutputDirectory "server.csr"
$serverCertificate = Join-Path $resolvedOutputDirectory "server.crt"
$serverExtensions = Join-Path $resolvedOutputDirectory "server.ext"
$deviceKey = Join-Path $resolvedOutputDirectory "device.key"
$deviceRequest = Join-Path $resolvedOutputDirectory "device.csr"
$deviceCertificate = Join-Path $resolvedOutputDirectory "device.crt"
$deviceExtensions = Join-Path $resolvedOutputDirectory "device.ext"

Invoke-OpenSSL $openssl @(
    "req", "-x509", "-newkey", "rsa:4096", "-sha256", "-days", "30",
    "-nodes", "-keyout", $caKey, "-out", $caCertificate,
    "-subj", "/CN=sprout-local-ca"
)

Invoke-OpenSSL $openssl @(
    "req", "-newkey", "rsa:2048", "-sha256", "-nodes",
    "-keyout", $serverKey, "-out", $serverRequest,
    "-subj", "/CN=mqtt"
)
@"
subjectAltName=DNS:mqtt,DNS:localhost,IP:127.0.0.1
extendedKeyUsage=serverAuth
"@ | Set-Content -LiteralPath $serverExtensions -Encoding ascii
Invoke-OpenSSL $openssl @(
    "x509", "-req", "-in", $serverRequest, "-CA", $caCertificate, "-CAkey", $caKey,
    "-CAcreateserial", "-out", $serverCertificate, "-days", "30", "-sha256",
    "-extfile", $serverExtensions
)

Invoke-OpenSSL $openssl @(
    "req", "-newkey", "rsa:2048", "-sha256", "-nodes",
    "-keyout", $deviceKey, "-out", $deviceRequest,
    "-subj", "/CN=device-platform"
)
@"
extendedKeyUsage=clientAuth
"@ | Set-Content -LiteralPath $deviceExtensions -Encoding ascii
Invoke-OpenSSL $openssl @(
    "x509", "-req", "-in", $deviceRequest, "-CA", $caCertificate, "-CAkey", $caKey,
    "-CAcreateserial", "-out", $deviceCertificate, "-days", "30", "-sha256",
    "-extfile", $deviceExtensions
)

Remove-Item -LiteralPath $serverRequest, $serverExtensions, $deviceRequest, $deviceExtensions -Force

Write-Host "Development MQTT certificates generated in $resolvedOutputDirectory"
Write-Host "These certificates are for local development only and must never be used in production."
