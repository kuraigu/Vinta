# scripts/sign.ps1
# Authenticode-signs one or more files using the certificate in the CurrentUser\My store.
#
# The certificate is selected by thumbprint (-Thumbprint, or $env:SIGNING_THUMBPRINT).
# If no thumbprint is configured the script does nothing, so local builds stay unsigned.
#
# Usage:
#   ./scripts/sign.ps1 build/bin/Vinta.exe
#   ./scripts/sign.ps1 build/bin/*.dll -Optional

param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string[]]$Path,

    [string]$Thumbprint = $env:SIGNING_THUMBPRINT,

    # Skip paths that don't exist instead of failing
    [switch]$Optional,

    [string[]]$TimestampServers = @(
        "http://timestamp.digicert.com",
        "http://timestamp.sectigo.com",
        "http://timestamp.globalsign.com/tsa/r6advanced1"
    )
)

$ErrorActionPreference = "Stop"

if (-not $Thumbprint) {
    Write-Host "No signing certificate configured. Skipping signing." -ForegroundColor Yellow
    return
}

$cert = Get-Item -Path "Cert:\CurrentUser\My\$Thumbprint" -ErrorAction SilentlyContinue
if (-not $cert) {
    throw "Signing certificate $Thumbprint was not found in Cert:\CurrentUser\My."
}

# Resolve targets (wildcards allowed)
$files = [System.Collections.Generic.List[System.IO.FileInfo]]::new()
foreach ($p in $Path) {
    $matches_ = @(Get-Item -Path $p -ErrorAction SilentlyContinue | Where-Object { -not $_.PSIsContainer })
    if ($matches_.Count -eq 0) {
        if ($Optional) { continue }
        throw "File to sign was not found: $p"
    }
    foreach ($m in $matches_) { $files.Add($m) }
}

if ($files.Count -eq 0) {
    Write-Host "Nothing to sign." -ForegroundColor Yellow
    return
}

foreach ($file in $files) {
    Write-Host "Signing $($file.FullName) ..." -ForegroundColor Cyan
    $signed = $false

    foreach ($ts in $TimestampServers) {
        try {
            $sig = Set-AuthenticodeSignature -FilePath $file.FullName -Certificate $cert `
                -HashAlgorithm SHA256 -TimestampServer $ts
            # NOTE: Status is "UnknownError" for self-signed certs because the root isn't
            # trusted, even when signing worked. Check the signer and timestamp instead.
            if ($sig.SignerCertificate -and
                $sig.SignerCertificate.Thumbprint -eq $cert.Thumbprint -and
                $sig.TimeStamperCertificate) {
                $signed = $true
                break
            }
            Write-Warning "Signing with timestamp server $ts did not produce a timestamped signature."
        } catch {
            Write-Warning "Timestamp server $ts failed: $($_.Exception.Message)"
        }
    }

    if (-not $signed) {
        throw "Failed to sign $($file.Name)."
    }

    $check = Get-AuthenticodeSignature -FilePath $file.FullName
    Write-Host "  Signed by: $($check.SignerCertificate.Subject)" -ForegroundColor Green
}