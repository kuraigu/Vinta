# scripts/generate_icon.ps1
# Automatically converts build/appicon.png into a high-quality multi-resolution build/windows/icon.ico

$ErrorActionPreference = 'Stop'

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptDir
$pngPath = Join-Path $repoRoot "build\appicon.png"
$icoPath = Join-Path $repoRoot "build\windows\icon.ico"

if (-not (Test-Path $pngPath)) {
    Write-Warning "Source image $pngPath not found."
    exit 0
}

Add-Type -AssemblyName System.Drawing

$src = [System.Drawing.Bitmap]::FromFile($pngPath)
$sizes = @(16, 24, 32, 48, 64, 128, 256)
$images = @()

foreach ($s in $sizes) {
    $bmp = New-Object System.Drawing.Bitmap($s, $s, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
    $g.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
    $g.Clear([System.Drawing.Color]::Transparent)
    $g.DrawImage($src, 0, 0, $s, $s)
    $g.Dispose()

    $ms = New-Object System.IO.MemoryStream
    $bmp.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
    $images += ,@($s, $ms.ToArray())
    $bmp.Dispose()
}
$src.Dispose()

$fs = [System.IO.File]::Create($icoPath)
$bw = New-Object System.IO.BinaryWriter($fs)

# ICONDIR Header: reserved (2 bytes), type (1 for icon, 2 bytes), count (2 bytes)
$bw.Write([uint16]0)
$bw.Write([uint16]1)
$bw.Write([uint16]$images.Count)

$offset = 6 + (16 * $images.Count)

# ICONDIRENTRY for each size
foreach ($item in $images) {
    $s = $item[0]
    $data = $item[1]
    $bw.Write([byte]$(if ($s -eq 256) { 0 } else { $s }))
    $bw.Write([byte]$(if ($s -eq 256) { 0 } else { $s }))
    $bw.Write([byte]0)
    $bw.Write([byte]0)
    $bw.Write([uint16]1)
    $bw.Write([uint16]32)
    $bw.Write([uint32]$data.Length)
    $bw.Write([uint32]$offset)
    $offset += $data.Length
}

# Write PNG byte arrays
foreach ($item in $images) {
    $bw.Write($item[1])
}

$bw.Flush()
$fs.Close()

Write-Host "Successfully generated multi-resolution icon: $icoPath" -ForegroundColor Green
