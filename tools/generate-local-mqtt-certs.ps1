[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$workspaceRoot = Split-Path -Parent $PSScriptRoot
$certificateDirectory = Join-Path $workspaceRoot 'deploy/mosquitto/certs'

New-Item -ItemType Directory -Force -Path $certificateDirectory | Out-Null

function ConvertTo-Pem {
    param(
        [Parameter(Mandatory)]
        [string] $Label,

        [Parameter(Mandatory)]
        [byte[]] $Bytes
    )

    $base64 = [Convert]::ToBase64String($Bytes)
    $lines = [regex]::Matches($base64, '.{1,64}') | ForEach-Object { $_.Value }
    return @(
        "-----BEGIN $Label-----"
        $lines
        "-----END $Label-----"
    ) -join "`n"
}

function Export-PrivateKey {
    param(
        [Parameter(Mandatory)]
        [System.Security.Cryptography.RSA] $PrivateKey
    )

    $pem = ConvertTo-Pem -Label 'PRIVATE KEY' -Bytes $PrivateKey.ExportPkcs8PrivateKey()
    return "$pem`n"
}

function New-RandomSerialNumber {
    $serialNumber = [byte[]]::new(16)
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($serialNumber)
    return $serialNumber
}

$caKey = [System.Security.Cryptography.RSA]::Create(3072)
$serverKey = $null
$clientKey = $null

try {
    $caRequest = [System.Security.Cryptography.X509Certificates.CertificateRequest]::new(
        'CN=Sprout Local MQTT CA',
        $caKey,
        [System.Security.Cryptography.HashAlgorithmName]::SHA256,
        [System.Security.Cryptography.RSASignaturePadding]::Pkcs1
    )
    $caRequest.CertificateExtensions.Add(
        [System.Security.Cryptography.X509Certificates.X509BasicConstraintsExtension]::new(
            $true,
            $false,
            0,
            $true
        )
    )
    $caRequest.CertificateExtensions.Add(
        [System.Security.Cryptography.X509Certificates.X509KeyUsageExtension]::new(
            [System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]::KeyCertSign -bor
                [System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]::CrlSign,
            $true
        )
    )
    $caCertificate = $caRequest.CreateSelfSigned(
        [DateTimeOffset]::UtcNow.AddDays(-1),
        [DateTimeOffset]::UtcNow.AddYears(1)
    )

    $serverKey = [System.Security.Cryptography.RSA]::Create(3072)
    $serverRequest = [System.Security.Cryptography.X509Certificates.CertificateRequest]::new(
        'CN=mqtt',
        $serverKey,
        [System.Security.Cryptography.HashAlgorithmName]::SHA256,
        [System.Security.Cryptography.RSASignaturePadding]::Pkcs1
    )
    $serverSan = [System.Security.Cryptography.X509Certificates.SubjectAlternativeNameBuilder]::new()
    $serverSan.AddDnsName('mqtt')
    $serverSan.AddIpAddress([System.Net.IPAddress]::Loopback)
    $serverRequest.CertificateExtensions.Add($serverSan.Build())
    $serverRequest.CertificateExtensions.Add(
        [System.Security.Cryptography.X509Certificates.X509KeyUsageExtension]::new(
            [System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]::DigitalSignature -bor
                [System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]::KeyEncipherment,
            $true
        )
    )
    $serverUsages = [System.Security.Cryptography.OidCollection]::new()
    $serverUsages.Add([System.Security.Cryptography.Oid]::new('1.3.6.1.5.5.7.3.1')) | Out-Null
    $serverRequest.CertificateExtensions.Add(
        [System.Security.Cryptography.X509Certificates.X509EnhancedKeyUsageExtension]::new(
            $serverUsages,
            $true
        )
    )
    $serverCertificate = $serverRequest.Create(
        $caCertificate,
        [DateTimeOffset]::UtcNow.AddDays(-1),
        [DateTimeOffset]::UtcNow.AddYears(1),
        (New-RandomSerialNumber)
    )

    $clientKey = [System.Security.Cryptography.RSA]::Create(3072)
    $clientRequest = [System.Security.Cryptography.X509Certificates.CertificateRequest]::new(
        'CN=local-device',
        $clientKey,
        [System.Security.Cryptography.HashAlgorithmName]::SHA256,
        [System.Security.Cryptography.RSASignaturePadding]::Pkcs1
    )
    $clientRequest.CertificateExtensions.Add(
        [System.Security.Cryptography.X509Certificates.X509KeyUsageExtension]::new(
            [System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]::DigitalSignature,
            $true
        )
    )
    $clientUsages = [System.Security.Cryptography.OidCollection]::new()
    $clientUsages.Add([System.Security.Cryptography.Oid]::new('1.3.6.1.5.5.7.3.2')) | Out-Null
    $clientRequest.CertificateExtensions.Add(
        [System.Security.Cryptography.X509Certificates.X509EnhancedKeyUsageExtension]::new(
            $clientUsages,
            $true
        )
    )
    $clientCertificate = $clientRequest.Create(
        $caCertificate,
        [DateTimeOffset]::UtcNow.AddDays(-1),
        [DateTimeOffset]::UtcNow.AddYears(1),
        (New-RandomSerialNumber)
    )

    $certificates = @{
        'ca.crt' = $caCertificate.ExportCertificatePem()
        'server.crt' = $serverCertificate.ExportCertificatePem()
        'device.crt' = $clientCertificate.ExportCertificatePem()
    }
    $privateKeys = @{
        'ca.key' = Export-PrivateKey -PrivateKey $caKey
        'server.key' = Export-PrivateKey -PrivateKey $serverKey
        'device.key' = Export-PrivateKey -PrivateKey $clientKey
    }

    foreach ($entry in $certificates.GetEnumerator()) {
        Set-Content -LiteralPath (Join-Path $certificateDirectory $entry.Key) `
            -Value "$($entry.Value)`n" -Encoding utf8NoBOM -NoNewline
    }
    foreach ($entry in $privateKeys.GetEnumerator()) {
        Set-Content -LiteralPath (Join-Path $certificateDirectory $entry.Key) `
            -Value $entry.Value -Encoding utf8NoBOM -NoNewline
    }
}
finally {
    $caKey.Dispose()
    if ($null -ne $serverKey) {
        $serverKey.Dispose()
    }
    if ($null -ne $clientKey) {
        $clientKey.Dispose()
    }
}

Write-Host "Generated local MQTT certificates in $certificateDirectory"
