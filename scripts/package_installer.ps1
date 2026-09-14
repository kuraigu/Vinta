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

Write-Host "Compiling Windows NSIS installer..." -ForegroundColor Cyan
Push-Location "$PSScriptRoot\..\build\windows\installer"
try {
    & makensis /DINFO_PRODUCTVERSION=$Version /DARG_WAILS_AMD64_BINARY=..\..\bin\Vinta.exe project.nsi
    if ($LASTEXITCODE -ne 0) {
        Write-Error "NSIS compilation failed."
        exit 1
    }
} finally {
    Pop-Location
}

Write-Host "Windows NSIS installer generated at build/bin/Vinta-Setup-x64.exe" -ForegroundColor Green
