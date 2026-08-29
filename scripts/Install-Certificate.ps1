# Installs FreeMatrix Studios Root / Code Signing Certificate into CurrentUser Trusted Root Certification Authorities
param (
    [string]$CertPath = "$PSScriptRoot\..\build\FreeMatrixStudios.cer"
)

if (-not (Test-Path $CertPath)) {
    if (Test-Path "$PSScriptRoot\FreeMatrixStudios.cer") {
        $CertPath = "$PSScriptRoot\FreeMatrixStudios.cer"
    } else {
        Write-Host "Certificate file not found at $CertPath" -ForegroundColor Yellow
        exit 0
    }
}

try {
    $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($CertPath)
    $store = New-Object System.Security.Cryptography.X509Certificates.X509Store("Root", "CurrentUser")
    $store.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite)
    $store.Add($cert)
    $store.Close()
    Write-Host "Successfully installed FreeMatrix Studios certificate into CurrentUser\Root store." -ForegroundColor Green
} catch {
    Write-Error "Failed to install certificate: $_"
    exit 1
}
