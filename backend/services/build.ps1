$ErrorActionPreference = "Stop"

Write-Host "Building VintaServices..." -ForegroundColor Cyan

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

dotnet publish VintaServices.csproj -c Release -r win-x64 --self-contained true -p:PublishSingleFile=true -p:IncludeNativeLibrariesForSelfExtract=true -p:EnableCompressionInSingleFile=true

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed." -ForegroundColor Red
    exit 1
}

$publishExe = "bin\Release\net8.0-windows10.0.19041.0\win-x64\publish\VintaServices.exe"
$buildBinDir = "..\..\build\bin"

if (Test-Path $publishExe) {
    if (-not (Test-Path $buildBinDir)) {
        New-Item -ItemType Directory -Path $buildBinDir -Force | Out-Null
    }
    Copy-Item $publishExe -Destination "$buildBinDir\VintaServices.exe" -Force
}

Write-Host "Build successful. Executable is at $publishExe and copied to build\bin\VintaServices.exe" -ForegroundColor Green
