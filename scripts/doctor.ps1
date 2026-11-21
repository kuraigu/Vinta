Write-Host "========================================" -ForegroundColor Cyan
Write-Host "       Vinta Environment Doctor         " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Refresh PATH from Machine + User registry so newly installed tools are detected immediately
$env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")

$allPassed = $true

function Test-Tool {
    param(
        [string]$Name,
        [string]$Command,
        [string]$VersionArg = "--version",
        [string]$InstallHint = "",
        [string[]]$FallbackPaths = @(),
        [switch]$Optional
    )

    $cmd = Get-Command $Command -ErrorAction SilentlyContinue
    if (-not $cmd -and $FallbackPaths.Count -gt 0) {
        foreach ($fb in $FallbackPaths) {
            if ($fb -and (Test-Path $fb)) {
                $cmd = Get-Item $fb
                $Command = $fb
                break
            }
        }
    }

    if ($cmd) {
        $resolvedPath = if ($cmd.Source) { $cmd.Source } else { $cmd.FullName }
        try {
            $global:LASTEXITCODE = 0
            $verOutput = (& $Command $VersionArg 2>&1 | Out-String).Trim()
            if ($LASTEXITCODE -eq 0 -and $verOutput) {
                $firstLine = ($verOutput -split "`r?`n")[0].Trim()
                Write-Host " [OK] " -ForegroundColor Green -NoNewline
                Write-Host "${Name}: " -ForegroundColor White -NoNewline
                Write-Host "$firstLine" -ForegroundColor Gray
            } else {
                Write-Host " [OK] " -ForegroundColor Green -NoNewline
                Write-Host "$Name (found at $resolvedPath)" -ForegroundColor White
            }
        } catch {
            Write-Host " [OK] " -ForegroundColor Green -NoNewline
            Write-Host "$Name (found at $resolvedPath)" -ForegroundColor White
        }
    } else {
        if ($Optional) {
            Write-Host " [OPTIONAL] " -ForegroundColor Yellow -NoNewline
            Write-Host "$Name" -ForegroundColor Yellow
            if ($InstallHint) {
                Write-Host "    Install hint: $InstallHint" -ForegroundColor DarkGray
            }
        } else {
            Write-Host " [MISSING] " -ForegroundColor Red -NoNewline
            Write-Host "$Name" -ForegroundColor Yellow
            if ($InstallHint) {
                Write-Host "    Install hint: $InstallHint" -ForegroundColor DarkGray
            }
            $script:allPassed = $false
        }
    }
}

Test-Tool -Name "Go Toolchain" -Command "go" -VersionArg "version" -InstallHint "winget install GoLang.Go"
Test-Tool -Name "Node.js Runtimes" -Command "node" -VersionArg "-v" -InstallHint "winget install OpenJS.NodeJS.LTS"
Test-Tool -Name "NPM Package Manager" -Command "npm" -VersionArg "-v" -InstallHint "Included with Node.js"
Test-Tool -Name "Wails v3 CLI" -Command "wails3" -VersionArg "version" -InstallHint "go install github.com/wailsapp/wails/v3/cmd/wails3@latest"
Test-Tool -Name "CMake" -Command "cmake" -VersionArg "--version" -InstallHint "winget install Kitware.CMake"
Test-Tool -Name "Ninja Build" -Command "ninja" -VersionArg "--version" -InstallHint "winget install Ninja-build.Ninja"
Test-Tool -Name "NSIS Installer Compiler" -Command "makensis" -VersionArg "/VERSION" -FallbackPaths @("C:\Program Files (x86)\NSIS\makensis.exe", "C:\Program Files\NSIS\makensis.exe") -InstallHint "winget install NSIS.NSIS" -Optional
Test-Tool -Name "Clang Format (LLVM)" -Command "clang-format" -VersionArg "--version" -FallbackPaths @("C:\Program Files\LLVM\bin\clang-format.exe", "C:\Program Files (x86)\LLVM\bin\clang-format.exe") -InstallHint "winget install LLVM.LLVM" -Optional
Test-Tool -Name "Taskfile CLI (go-task)" -Command "task" -VersionArg "--version" -InstallHint "winget install Task.Task (or use ./scripts.bat)" -Optional

# Check MSVC / Visual Studio C++ Build Tools
$vswhere = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer\vswhere.exe"
$msvcFound = $false
if (Test-Path $vswhere) {
    $vsPath = (& $vswhere -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath 2>$null)
    if ($vsPath -and (Test-Path $vsPath)) {
        Write-Host " [OK] " -ForegroundColor Green -NoNewline
        Write-Host "MSVC C++ Toolchain: " -ForegroundColor White -NoNewline
        Write-Host "$vsPath" -ForegroundColor Gray
        $msvcFound = $true
    }
}

if (-not $msvcFound) {
    $clCmd = Get-Command "cl.exe" -ErrorAction SilentlyContinue
    if ($clCmd) {
        Write-Host " [OK] " -ForegroundColor Green -NoNewline
        Write-Host "MSVC Compiler (cl.exe on PATH): " -ForegroundColor White -NoNewline
        Write-Host "$($clCmd.Source)" -ForegroundColor Gray
    } else {
        Write-Host " [MISSING] " -ForegroundColor Red -NoNewline
        Write-Host "Visual Studio C++ Build Tools & Windows SDK" -ForegroundColor Yellow
        Write-Host "    Install hint: winget install Microsoft.VisualStudio.2022.BuildTools --override `"--add Microsoft.VisualStudio.Workload.VCTools --add Microsoft.VisualStudio.Component.Windows11SDK.22621 --passive`"" -ForegroundColor DarkGray
        $allPassed = $false
    }
}

Write-Host ""
if ($allPassed) {
    Write-Host "All required development toolchains are installed and ready." -ForegroundColor Green
} else {
    Write-Host "Some toolchains are missing. You can run './scripts.bat setup --install-tools' to install them automatically via winget." -ForegroundColor Yellow
}
Write-Host ""
