# Vinta Theme Packager
# Compresses a theme directory into a valid .vintatheme bundle

param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$ThemePath,

    [Parameter(Mandatory = $false, Position = 1)]
    [string]$OutputPath
)

$ErrorActionPreference = 'Stop'

if (-not (Test-Path $ThemePath -PathType Container)) {
    Write-Error "Theme directory not found: $ThemePath"
    exit 1
}

$resolvedThemePath = (Resolve-Path $ThemePath).Path
$manifestPath = Join-Path $resolvedThemePath "manifest.json"

if (-not (Test-Path $manifestPath -PathType Leaf)) {
    Write-Error "manifest.json not found in $resolvedThemePath"
    exit 1
}

try {
    $manifestContent = Get-Content -Raw -Path $manifestPath -Encoding UTF8
    $manifest = $manifestContent | ConvertFrom-Json
} catch {
    Write-Error "Failed to parse manifest.json: $_"
    exit 1
}

# Validate required fields
$requiredFields = @("id", "name", "author", "version", "entry")
foreach ($field in $requiredFields) {
    if (-not $manifest.$field) {
        Write-Error "manifest.json is missing required field: '$field'"
        exit 1
    }
}

# Validate entry stylesheet exists
$entryFile = Join-Path $resolvedThemePath $manifest.entry
if (-not (Test-Path $entryFile -PathType Leaf)) {
    Write-Error "Entry stylesheet '$($manifest.entry)' not found in theme directory."
    exit 1
}

# Determine destination path
if (-not $OutputPath) {
    $OutputPath = Join-Path (Get-Location) "$($manifest.id).vintatheme"
} elseif (Test-Path $OutputPath -PathType Container) {
    $OutputPath = Join-Path $OutputPath "$($manifest.id).vintatheme"
}

if (Test-Path $OutputPath) {
    Remove-Item -Force $OutputPath
}

# Compress theme contents into .vintatheme (zip format)
$tempZip = [System.IO.Path]::ChangeExtension($OutputPath, ".zip")
if (Test-Path $tempZip) {
    Remove-Item -Force $tempZip
}

Compress-Archive -Path (Join-Path $resolvedThemePath "*") -DestinationPath $tempZip -CompressionLevel Optimal
Move-Item -Path $tempZip -Destination $OutputPath -Force

Write-Host "Successfully packaged Vinta theme:" -ForegroundColor Green
Write-Host "  Name:    $($manifest.name)" -ForegroundColor Cyan
Write-Host "  ID:      $($manifest.id)" -ForegroundColor Cyan
Write-Host "  Author:  $($manifest.author)" -ForegroundColor Cyan
Write-Host "  Version: $($manifest.version)" -ForegroundColor Cyan
Write-Host "  Output:  $OutputPath" -ForegroundColor Green
