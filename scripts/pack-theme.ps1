# Vinta Theme Packager & Token Auditor
# Compresses a theme directory into a valid .vintatheme bundle and warns about unapplied tokens

param(
    [Parameter(Mandatory = $false, Position = 0)]
    [string]$ThemePath,

    [Parameter(Mandatory = $false, Position = 1)]
    [string]$OutputPath,

    [Parameter(Mandatory = $false)]
    [switch]$CheckOnly
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$tokensCssPath = Join-Path $repoRoot "frontend\src\themes\tokens.css"

function Get-DeclaredVintaTokens {
    param([string]$CssFilePath)
    if (-not (Test-Path $CssFilePath -PathType Leaf)) {
        return @()
    }
    $content = Get-Content -Raw -Path $CssFilePath -Encoding UTF8
    $matchesList = [regex]::Matches($content, '--vinta-[a-z0-9-]+(?=\s*:)')
    return @($matchesList | ForEach-Object { $_.Value } | Select-Object -Unique)
}

function Test-ThemeTokens {
    param(
        [string]$TargetDir,
        [string[]]$BaseTokens
    )
    $manifestFile = Join-Path $TargetDir "manifest.json"
    $entryRel = "style.css"
    $themeName = Split-Path -Leaf $TargetDir

    if (Test-Path $manifestFile -PathType Leaf) {
        try {
            $m = Get-Content -Raw -Path $manifestFile -Encoding UTF8 | ConvertFrom-Json
            if ($m.entry) { $entryRel = $m.entry }
            if ($m.id) { $themeName = "$($m.name) ($($m.id))" }
        } catch {
            # Ignore manifest parse failure during token audit
        }
    }

    $cssFile = Join-Path $TargetDir $entryRel
    if (-not (Test-Path $cssFile -PathType Leaf)) {
        Write-Warning "[Theme Audit] Entry stylesheet '$entryRel' not found in $TargetDir"
        return
    }

    $implementedTokens = Get-DeclaredVintaTokens -CssFilePath $cssFile
    $missingTokens = @($BaseTokens | Where-Object { $implementedTokens -notcontains $_ })

    if ($missingTokens.Count -gt 0) {
        Write-Host "WARNING: Theme '$themeName' is missing $($missingTokens.Count) token(s) declared in tokens.css (will fall back to dark-gold):" -ForegroundColor Yellow
        foreach ($t in $missingTokens) {
            Write-Host "  - $t" -ForegroundColor Yellow
        }
    } else {
        Write-Host "[Theme Audit] '$themeName' implements all $($BaseTokens.Count) tokens declared in tokens.css." -ForegroundColor DarkGray
    }
}

$baseTokens = Get-DeclaredVintaTokens -CssFilePath $tokensCssPath

if ($CheckOnly) {
    if ($ThemePath) {
        if (-not (Test-Path $ThemePath -PathType Container)) {
            Write-Warning "Theme directory not found: $ThemePath"
            exit 0
        }
        Test-ThemeTokens -TargetDir (Resolve-Path $ThemePath).Path -BaseTokens $baseTokens
    } else {
        $presetsDir = Join-Path $repoRoot "frontend\src\themes\presets"
        if (Test-Path $presetsDir -PathType Container) {
            Get-ChildItem -Path $presetsDir -Directory | ForEach-Object {
                Test-ThemeTokens -TargetDir $_.FullName -BaseTokens $baseTokens
            }
        }
        $templateDir = Join-Path $repoRoot "templates\theme"
        if (Test-Path $templateDir -PathType Container) {
            Test-ThemeTokens -TargetDir $templateDir -BaseTokens $baseTokens
        }
    }
    exit 0
}

if (-not $ThemePath) {
    Write-Error "ThemePath is required when packaging a theme (or pass -CheckOnly to audit built-in presets)."
    exit 1
}

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

# Audit declared tokens against tokens.css (non-blocking warning)
if ($baseTokens.Count -gt 0) {
    Test-ThemeTokens -TargetDir $resolvedThemePath -BaseTokens $baseTokens
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
