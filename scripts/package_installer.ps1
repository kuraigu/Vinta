# scripts/package_installer.ps1
# Packages the production binary into the Windows NSIS installer.
#
# Usage:
#   ./scripts/package_installer.ps1 -Version v0.2.0
#
# Signing is NOT done here. The release workflow signs Vinta.exe before this script runs
# and signs the finished installer afterwards (see scripts/sign.ps1).

param(
    [string]$Version = "0.1.0",
    [string]$ExpectedPublisher = "FreeMatrix Studios"
)

$ErrorActionPreference = "Stop"
$repoRoot     = (Resolve-Path "$PSScriptRoot\..").Path
$installerDir = Join-Path $repoRoot "build\windows\installer"
$binDir       = Join-Path $repoRoot "build\bin"
$isCI         = [bool]$env:CI

# --- NSIS -------------------------------------------------------------------
foreach ($p in @("C:\Program Files (x86)\NSIS", "C:\Program Files\NSIS")) {
    if (Test-Path $p) { $env:PATH = "$p;$env:PATH" }
}

if (-not (Get-Command "makensis" -ErrorAction SilentlyContinue)) {
    Write-Error "makensis (NSIS) was not found on PATH. Install it with 'winget install NSIS.NSIS' or add it to your PATH."
}

# --- Production binary ------------------------------------------------------
if (-not (Test-Path (Join-Path $binDir "Vinta.exe"))) {
    Write-Host "Building Vinta production binary first..." -ForegroundColor Cyan
    task build
    if ($LASTEXITCODE -ne 0) { Write-Error "Failed to build Vinta production binary." }
}

# --- Publisher sanity check -------------------------------------------------
# The "Publisher" shown in Add/Remove Programs and the exe properties comes from
# build/windows/info.json (CompanyName). Warn if it isn't the expected name.
$infoJson = Join-Path $repoRoot "build\windows\info.json"
if (Test-Path $infoJson) {
    $infoText = Get-Content $infoJson -Raw
    if ($infoText -match '"CompanyName"\s*:\s*"([^"]*)"') {
        if ($Matches[1] -ne $ExpectedPublisher) {
            Write-Warning "build/windows/info.json CompanyName is '$($Matches[1])', expected '$ExpectedPublisher'. The installer will show that publisher name."
        }
    } else {
        Write-Warning "No CompanyName found in build/windows/info.json. The installer may show a blank or default publisher."
    }
}

# --- WebView2 bootstrapper --------------------------------------------------
# Download with retries, and verify it is really signed by Microsoft before it gets
# bundled into the installer.
$webviewBootstrapper = Join-Path $installerDir "tmp\MicrosoftEdgeWebview2Setup.exe"

function Test-MicrosoftSigned([string]$file) {
    $sig = Get-AuthenticodeSignature -FilePath $file
    return ($sig.Status -eq "Valid" -and $sig.SignerCertificate.Subject -match "O=Microsoft Corporation")
}

if ((Test-Path $webviewBootstrapper) -and -not (Test-MicrosoftSigned $webviewBootstrapper)) {
    Write-Warning "Cached WebView2 bootstrapper failed signature verification. Re-downloading."
    Remove-Item $webviewBootstrapper -Force
}

if (-not (Test-Path $webviewBootstrapper)) {
    Write-Host "Downloading Microsoft Edge WebView2 Evergreen Bootstrapper..." -ForegroundColor Cyan
    $tmpDir = Split-Path $webviewBootstrapper -Parent
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    $downloaded = $false
    foreach ($attempt in 1..3) {
        try {
            Invoke-WebRequest -Uri "https://go.microsoft.com/fwlink/p/?LinkId=2124703" -OutFile $webviewBootstrapper -UseBasicParsing
            if (Test-MicrosoftSigned $webviewBootstrapper) { $downloaded = $true; break }
            Write-Warning "Downloaded WebView2 bootstrapper is not validly signed by Microsoft."
            Remove-Item $webviewBootstrapper -Force -ErrorAction SilentlyContinue
        } catch {
            Write-Warning "Attempt $attempt failed: $($_.Exception.Message)"
            Start-Sleep -Seconds (2 * $attempt)
        }
    }

    if (-not $downloaded) {
        $msg = "Could not obtain a verified WebView2 bootstrapper. Users without WebView2 would not be able to launch Vinta after install."
        if ($isCI) { Write-Error $msg } else { Write-Warning $msg }
    }
}

# --- Version numbers --------------------------------------------------------
# Strip a leading 'v' and build the X.X.X.X form NSIS needs for fixed file info.
$cleanVer = $Version.TrimStart('v')
if ($cleanVer -match '^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:\.(\d+))?') {
    $parts = 1..4 | ForEach-Object { if ($Matches[$_]) { $Matches[$_] } else { "0" } }
    $numericVersion = $parts -join "."
} else {
    $numericVersion = "0.1.0.0"
}

# --- Compile ----------------------------------------------------------------
# Remove stale installers so we never pick up or ship an old one.
Get-ChildItem $binDir -Filter "Vinta-*.exe" -ErrorAction SilentlyContinue | Remove-Item -Force

Write-Host "Compiling Windows NSIS installer (version: $Version, numeric: $numericVersion)..." -ForegroundColor Cyan
Push-Location $installerDir
try {
    & makensis /DINFO_PRODUCTVERSION="$Version" /DINFO_PRODUCTVERSION_NUMERIC="$numericVersion" /DARG_WAILS_AMD64_BINARY=..\..\bin\Vinta.exe project.nsi
    if ($LASTEXITCODE -ne 0) { Write-Error "NSIS compilation failed." }
} finally {
    Pop-Location
}

# --- Verify output ----------------------------------------------------------
$installer = Get-ChildItem $binDir -File -ErrorAction SilentlyContinue |
    Where-Object { $_.Name -eq "Vinta-Setup-x64.exe" -or $_.Name -like "Vinta-*-installer.exe" } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 1

if (-not $installer) {
    Write-Error "NSIS reported success, but no installer was found in build/bin."
}

$sizeMB = [math]::Round($installer.Length / 1MB, 1)
$hash   = (Get-FileHash $installer.FullName -Algorithm SHA256).Hash.ToLower()
Write-Host "Installer generated: $($installer.FullName)" -ForegroundColor Green
Write-Host "  Size:   $sizeMB MB"
Write-Host "  SHA256: $hash"

if ($env:GITHUB_STEP_SUMMARY) {
    @(
        "### Installer built",
        "",
        "| File | Size | SHA256 |",
        "| --- | --- | --- |",
        "| ``$($installer.Name)`` | $sizeMB MB | ``$hash`` |"
    ) | Out-File -FilePath $env:GITHUB_STEP_SUMMARY -Append -Encoding utf8
}