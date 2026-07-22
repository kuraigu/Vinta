$ErrorActionPreference = 'Stop'

# Set working directory to the script's location
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location -Path $scriptDir

if (-not (Test-Path '.dotnet\dotnet.exe')) {
    Write-Host 'Downloading .NET SDK locally (does not install globally)...'
    Invoke-WebRequest 'https://dot.net/v1/dotnet-install.ps1' -OutFile 'dotnet-install.ps1'
    .\dotnet-install.ps1 -InstallDir .dotnet -Channel 8.0 -NoPath
}

Write-Host 'Building project...'
.\.dotnet\dotnet.exe publish -c Release -r win-x64 --self-contained true

if ($LASTEXITCODE -eq 0) {
    Write-Host 'Build successful. Executable is at bin\Release\net8.0-windows10.0.19041.0\win-x64\publish\VintaNotifications.exe'
} else {
    Write-Error 'Build failed.'
}
