# scripts/package_installer.ps1
# Builds production binary and packages Windows NSIS installer

param(
    [string]$Version = "0.1.0"
)

$nsisPaths = @(
    "C:\Program Files (x86)\NSIS",
    "C:\Program Files\NSIS"
)

foreach ($p in $nsisPaths) {
    if (Test-Path $p) {
        $env:PATH = "$p;$env:PATH"
    }
}

if (-not (Get-Command "makensis" -ErrorAction SilentlyContinue)) {
    Write-Error "makensis (NSIS) was not found on PATH. Install NSIS via 'winget install NSIS.NSIS' or add it to your PATH."
    exit 1
}

# Ensure production binary exists
if (-not (Test-Path "build/bin/Vinta.exe")) {
    Write-Host "Building Vinta production binary first..." -ForegroundColor Cyan
    task build
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build Vinta production binary."
        exit 1
    }
}

# Ensure WebView2 bootstrapper exists for NSIS installer packaging
$installerDir = "$PSScriptRoot\..\build\windows\installer"
$webviewBootstrapper = Join-Path $installerDir "tmp\MicrosoftEdgeWebview2Setup.exe"
if (-not (Test-Path $webviewBootstrapper)) {
    Write-Host "Downloading Microsoft Edge WebView2 Evergreen Bootstrapper..." -ForegroundColor Cyan
    $tmpDir = Split-Path $webviewBootstrapper -Parent
    if (-not (Test-Path $tmpDir)) {
        New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
    }
    try {
        Invoke-WebRequest -Uri "https://go.microsoft.com/fwlink/p/?LinkId=2124703" -OutFile $webviewBootstrapper -UseBasicParsing
    } catch {
        Write-Warning "Could not download WebView2 bootstrapper: $($_.Exception.Message)"
    }
}

# Clean leading 'v' and extract major.minor.patch.build for NSIS fixed file info (X.X.X.X)
$cleanVer = $Version.TrimStart('v')
if ($cleanVer -match '^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:\.(\d+))?') {
    $v1 = if ($Matches[1]) { $Matches[1] } else { "0" }
    $v2 = if ($Matches[2]) { $Matches[2] } else { "0" }
    $v3 = if ($Matches[3]) { $Matches[3] } else { "0" }
    $v4 = if ($Matches[4]) { $Matches[4] } else { "0" }
    $numericVersion = "$v1.$v2.$v3.$v4"
} else {
    $numericVersion = "0.1.0.0"
}

Write-Host "Compiling Windows NSIS installer (version: $Version, numeric: $numericVersion)..." -ForegroundColor Cyan
Push-Location "$PSScriptRoot\..\build\windows\installer"
try {
    & makensis /DINFO_PRODUCTVERSION="$Version" /DINFO_PRODUCTVERSION_NUMERIC="$numericVersion" /DARG_WAILS_AMD64_BINARY=..\..\bin\Vinta.exe project.nsi
    if ($LASTEXITCODE -ne 0) {
        Write-Error "NSIS compilation failed."
        exit 1
    }
} finally {
    Pop-Location
}

Write-Host "Windows NSIS installer generated at build/bin/Vinta-Setup-x64.exe" -ForegroundColor Green
