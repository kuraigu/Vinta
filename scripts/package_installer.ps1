# scripts/package_installer.ps1
# Builds production binary and packages Windows NSIS installer

$nsisPaths = @(
    "C:\Program Files (x86)\NSIS",
    "C:\Program Files\NSIS"
)

foreach ($p in $nsisPaths) {
    if (Test-Path $p) {
        $env:PATH = "$p;$env:PATH"
    }
}

& "$PSScriptRoot\generate_icon.ps1"

wails build -nsis
