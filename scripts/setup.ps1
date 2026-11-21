# Vinta Developer Environment Setup Script
param(
    [switch]$InstallTools,
    [switch]$SkipTools,
    [switch]$Yes,
    [switch]$Help
)

if ($Help) {
    Write-Host "Usage: scripts.bat setup [--install-tools] [--skip-tools] [--yes] [--help]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  --install-tools, -InstallTools  Install missing required and optional toolchains via winget"
    Write-Host "  --skip-tools, -SkipTools        Skip optional tool installation prompts"
    Write-Host "  --yes, -Yes                     Auto-accept optional tool installation prompts"
    Write-Host "  --help, -h                      Show this help message"
    exit 0
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "       Vinta Developer Setup            " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. Ensure Taskfile CLI (task) is installed
$taskCmd = Get-Command "task" -ErrorAction SilentlyContinue
if (-not $taskCmd) {
    $goPath = (go env GOPATH 2>$null)
    if ($goPath -and (Test-Path (Join-Path $goPath "bin\task.exe"))) {
        $taskCmd = (Join-Path $goPath "bin\task.exe")
    }
}

if (-not $taskCmd) {
    Write-Host "Installing Taskfile CLI (go-task)..." -ForegroundColor Cyan
    go install github.com/go-task/task/v3/cmd/task@latest
}

# 2. Check & Install Required System Tools if requested
if ($InstallTools) {
    Write-Host "Installing missing required system toolchains via winget..." -ForegroundColor Yellow

    if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
        Write-Host "Installing Go..." -ForegroundColor Cyan
        winget install GoLang.Go --accept-source-agreements --accept-package-agreements
    }
    if (-not (Get-Command "node" -ErrorAction SilentlyContinue)) {
        Write-Host "Installing Node.js LTS..." -ForegroundColor Cyan
        winget install OpenJS.NodeJS.LTS --accept-source-agreements --accept-package-agreements
    }
    if (-not (Get-Command "cmake" -ErrorAction SilentlyContinue)) {
        Write-Host "Installing CMake..." -ForegroundColor Cyan
        winget install Kitware.CMake --accept-source-agreements --accept-package-agreements
    }
    if (-not (Get-Command "ninja" -ErrorAction SilentlyContinue)) {
        Write-Host "Installing Ninja..." -ForegroundColor Cyan
        winget install Ninja-build.Ninja --accept-source-agreements --accept-package-agreements
    }
    if (-not (Get-Command "wails3" -ErrorAction SilentlyContinue)) {
        Write-Host "Installing Wails v3 CLI..." -ForegroundColor Cyan
        go install github.com/wailsapp/wails/v3/cmd/wails3@latest
    }
}

# 2b. Optional Tools: Prompt (or auto-install with --install-tools / --yes)
if (-not $SkipTools) {
    $hasNsis = (Get-Command "makensis" -ErrorAction SilentlyContinue) -or
               (Test-Path "C:\Program Files (x86)\NSIS\makensis.exe") -or
               (Test-Path "C:\Program Files\NSIS\makensis.exe")
    if (-not $hasNsis) {
        $installNsis = $InstallTools -or $Yes
        if (-not $installNsis -and [Environment]::UserInteractive -and -not [Console]::IsInputRedirected) {
            $reply = Read-Host "Optional tool 'NSIS' (installer packaging) is missing. Install via winget? [y/N]"
            $installNsis = $reply -match '^(y|yes)$'
        }
        if ($installNsis) {
            Write-Host "Installing NSIS..." -ForegroundColor Cyan
            winget install NSIS.NSIS --accept-source-agreements --accept-package-agreements
        }
    }

    $hasClangFormat = (Get-Command "clang-format" -ErrorAction SilentlyContinue) -or
                      (Test-Path "C:\Program Files\LLVM\bin\clang-format.exe") -or
                      (Test-Path "C:\Program Files (x86)\LLVM\bin\clang-format.exe")
    if (-not $hasClangFormat) {
        $installLlvm = $InstallTools -or $Yes
        if (-not $installLlvm -and [Environment]::UserInteractive -and -not [Console]::IsInputRedirected) {
            $reply = Read-Host "Optional tool 'LLVM (clang-format)' (C++ formatting) is missing. Install via winget? [y/N]"
            $installLlvm = $reply -match '^(y|yes)$'
        }
        if ($installLlvm) {
            Write-Host "Installing LLVM (clang-format)..." -ForegroundColor Cyan
            winget install LLVM.LLVM --accept-source-agreements --accept-package-agreements
        }
    }
}

# 3. Install Frontend Node Packages
Write-Host "Installing frontend dependencies (npm install)..." -ForegroundColor Cyan
cmd.exe /c "cd frontend && npm install"
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Failed to install frontend dependencies." -ForegroundColor Red
    exit 1
}

# 4. Download Go Modules
Write-Host "Downloading Go modules..." -ForegroundColor Cyan
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Failed to download Go modules." -ForegroundColor Red
    exit 1
}

# 5. Configure CMake Native Build Directory
Write-Host "Configuring CMake native C++ build tree..." -ForegroundColor Cyan
cmake -B build/native -S backend/services/native
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Failed to configure CMake." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Setup complete! You can now run '.\scripts.bat dev' or 'task dev' to start developing." -ForegroundColor Green
Write-Host "Note: If 'task' is not recognized in your current terminal, please restart your terminal or use '.\scripts.bat'." -ForegroundColor Gray
Write-Host ""
