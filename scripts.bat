@echo off
:: ============================================================================
:: Vinta Build & Development Pipeline (Single Source of Truth)
::
:: All build, dev, packaging, and lint logic lives in this file so contributors
:: do not need external task runners installed. Taskfile.yml and CI delegate
:: directly to this script.
:: ============================================================================
setlocal enabledelayedexpansion

if "%~1"=="" goto usage

if "%~1"=="--help" goto usage
if "%~1"=="-h" goto usage
if "%~1"=="help" (
    if "%~2"=="" goto usage
    call "%~f0" "%~2" --help
    exit /b !ERRORLEVEL!
)

if "%~1"=="doctor" (
    if "%~2"=="--help" (
        echo Usage: scripts.bat doctor
        echo Checks required and optional development toolchains ^(Go, Node, Wails, CMake, Ninja, NSIS, LLVM/clang-format, Task, MSVC^).
        exit /b 0
    )
    if "%~2"=="-h" (
        echo Usage: scripts.bat doctor
        echo Checks required and optional development toolchains ^(Go, Node, Wails, CMake, Ninja, NSIS, LLVM/clang-format, Task, MSVC^).
        exit /b 0
    )
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\doctor.ps1"
    exit /b !ERRORLEVEL!
)

if "%~1"=="setup" (
    set "SETUP_ARGS="
    for %%A in (%2 %3 %4 %5) do (
        if "%%~A"=="--install-tools" (
            set "SETUP_ARGS=!SETUP_ARGS! -InstallTools"
        ) else if "%%~A"=="-InstallTools" (
            set "SETUP_ARGS=!SETUP_ARGS! -InstallTools"
        ) else if "%%~A"=="--skip-tools" (
            set "SETUP_ARGS=!SETUP_ARGS! -SkipTools"
        ) else if "%%~A"=="-SkipTools" (
            set "SETUP_ARGS=!SETUP_ARGS! -SkipTools"
        ) else if "%%~A"=="--yes" (
            set "SETUP_ARGS=!SETUP_ARGS! -Yes"
        ) else if "%%~A"=="-y" (
            set "SETUP_ARGS=!SETUP_ARGS! -Yes"
        ) else if "%%~A"=="--help" (
            set "SETUP_ARGS=!SETUP_ARGS! -Help"
        ) else if "%%~A"=="-h" (
            set "SETUP_ARGS=!SETUP_ARGS! -Help"
        ) else (
            set "SETUP_ARGS=!SETUP_ARGS! %%~A"
        )
    )
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\setup.ps1"!SETUP_ARGS!
    exit /b !ERRORLEVEL!
)

if "%~1"=="kill" (
    call :kill_app
    exit /b 0
)

if "%~1"=="icon" (
    call :generate_icon
    exit /b !ERRORLEVEL!
)

if "%~1"=="bindings" (
    call :generate_bindings
    exit /b !ERRORLEVEL!
)

if "%~1"=="syso" (
    call :generate_syso
    exit /b !ERRORLEVEL!
)

if "%~1"=="build-native" (
    call :kill_app
    call :build_native
    exit /b !ERRORLEVEL!
)

if "%~1"=="build-frontend" (
    call :generate_bindings
    if errorlevel 1 exit /b !ERRORLEVEL!
    call :build_frontend
    exit /b !ERRORLEVEL!
)

if "%~1"=="dev" (
    if "%~2"=="frontend" (
        pushd "%~dp0frontend"
        call npm run dev -- --port 9245 --strictPort
        set "ERR=!ERRORLEVEL!"
        popd
        exit /b !ERR!
    )
    if "%~2"=="--frontend" (
        pushd "%~dp0frontend"
        call npm run dev -- --port 9245 --strictPort
        set "ERR=!ERRORLEVEL!"
        popd
        exit /b !ERR!
    )

    set "DO_BUILD_NATIVE=1"
    set "DO_PROFILE=0"
    for %%A in (%2 %3 %4) do (
        if "%%~A"=="--help" (
            echo Usage: scripts.bat dev [--no-build] [--profile] [--frontend]
            exit /b 0
        ) else if "%%~A"=="-h" (
            echo Usage: scripts.bat dev [--no-build] [--profile] [--frontend]
            exit /b 0
        ) else if "%%~A"=="no_build" (
            set "DO_BUILD_NATIVE=0"
        ) else if "%%~A"=="--no-build" (
            set "DO_BUILD_NATIVE=0"
        ) else if "%%~A"=="profile" (
            set "DO_PROFILE=1"
        ) else if "%%~A"=="--profile" (
            set "DO_PROFILE=1"
        ) else (
            echo Error: Unknown argument '%%~A' for 'dev'.
            echo Allowed options for 'dev': [--no-build ^| --profile ^| --frontend]
            exit /b 1
        )
    )

    call :kill_app
    if "!DO_BUILD_NATIVE!"=="1" (
        call :build_native
        if errorlevel 1 exit /b !ERRORLEVEL!
    )
    call :generate_syso
    if errorlevel 1 exit /b !ERRORLEVEL!
    set "VINTA_SHOWCASE="
    set "VINTA_NO_DEVTEST="
    if "!DO_PROFILE!"=="1" (
        set "WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS=--remote-debugging-port=9222"
        echo Remote debugging enabled: edge://inspect ^(port 9222^)
    ) else (
        set "WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS="
    )
    echo Dev server: http://localhost:9245
    wails3 dev -config ./build/config.yml -port 9245
    set "DEV_ERR=!ERRORLEVEL!"
    if exist "wails_windows_amd64.syso" del /f /q "wails_windows_amd64.syso"
    exit /b !DEV_ERR!
)

if "%~1"=="showcase" (
    call :kill_app
    call :build_native
    if errorlevel 1 exit /b !ERRORLEVEL!
    call :generate_syso
    if errorlevel 1 exit /b !ERRORLEVEL!
    set "VINTA_SHOWCASE=1"
    set "VINTA_NO_DEVTEST="
    wails3 dev -config ./build/config.yml -port 9245
    set "DEV_ERR=!ERRORLEVEL!"
    if exist "wails_windows_amd64.syso" del /f /q "wails_windows_amd64.syso"
    exit /b !DEV_ERR!
)

if "%~1"=="build" (
    set "DEV_BUILD=0"
    if "%~2"=="--help" (
        echo Usage: scripts.bat build [--dev]
        exit /b 0
    ) else if "%~2"=="-h" (
        echo Usage: scripts.bat build [--dev]
        exit /b 0
    ) else if "%~2"=="dev" (
        set "DEV_BUILD=1"
    ) else if "%~2"=="--dev" (
        set "DEV_BUILD=1"
    ) else if "%~2" NEQ "" (
        echo Error: Unknown argument '%~2' for 'build'. Allowed: [--dev]
        exit /b 1
    )

    call :do_build !DEV_BUILD!
    exit /b !ERRORLEVEL!
)

if "%~1"=="run" (
    set "DO_BUILD=1"
    set "DEBUG_FLAG="

    for %%A in (%2 %3 %4) do (
        if "%%~A"=="--help" (
            echo Usage: scripts.bat run [--no-build] [--debug]
            exit /b 0
        ) else if "%%~A"=="-h" (
            echo Usage: scripts.bat run [--no-build] [--debug]
            exit /b 0
        ) else if "%%~A"=="no_build" (
            set "DO_BUILD=0"
        ) else if "%%~A"=="--no-build" (
            set "DO_BUILD=0"
        ) else if "%%~A"=="debug" (
            set "DEBUG_FLAG=-debug"
        ) else if "%%~A"=="--debug" (
            set "DEBUG_FLAG=-debug"
        ) else (
            echo Error: Unknown argument '%%~A' for 'run'.
            echo Allowed options for 'run': [--no-build] [--debug]
            exit /b 1
        )
    )

    call :kill_app
    if "!DO_BUILD!"=="1" (
        if defined DEBUG_FLAG (
            call :do_build 1
        ) else (
            call :do_build 0
        )
        if errorlevel 1 exit /b !ERRORLEVEL!
    )

    if defined DEBUG_FLAG (
        if "!DO_BUILD!"=="0" set "VINTA_NO_DEVTEST=1"
        set "VINTA_SHOWCASE="
        start "" "build\bin\Vinta.exe" -debug
    ) else (
        if "!DO_BUILD!"=="0" (
            if defined FRONTEND_DEVSERVER_URL (
                "build\bin\Vinta.exe"
            ) else (
                set "VINTA_NO_DEVTEST=1"
                set "VINTA_SHOWCASE="
                start "" "build\bin\Vinta.exe"
            )
        ) else (
            start "" "build\bin\Vinta.exe"
        )
    )
    exit /b 0
)

if "%~1"=="test" (
    set "DO_BUILD=1"
    for %%A in (%2 %3) do (
        if "%%~A"=="--help" (
            echo Usage: scripts.bat test [--no-build]
            exit /b 0
        ) else if "%%~A"=="-h" (
            echo Usage: scripts.bat test [--no-build]
            exit /b 0
        ) else if "%%~A"=="no_build" (
            set "DO_BUILD=0"
        ) else if "%%~A"=="--no-build" (
            set "DO_BUILD=0"
        ) else (
            echo Error: Unknown argument '%%~A' for 'test'.
            echo Allowed options for 'test': [--no-build]
            exit /b 1
        )
    )
    call :kill_app
    if "!DO_BUILD!"=="1" (
        call :do_build 1
        if errorlevel 1 exit /b !ERRORLEVEL!
    )
    start "" "build\bin\Vinta.exe"
    exit /b 0
)

if "%~1"=="lint" (
    set "CI_MODE=0"
    set "NATIVE_ONLY=0"
    set "CHECK_ALL_FLAG="
    set "CHECK_STRICT_FLAG="
    for %%A in (%2 %3 %4 %5) do (
        if "%%~A"=="--help" (
            echo Usage: scripts.bat lint [--ci] [--native] [--all] [--strict]
            echo   --ci      Non-mutating check mode ^(gofmt -l, format:check, clang-format --dry-run^)
            echo   --native  Run only C++ clang-format check
            echo   --all     Print all warnings from check-transitions and check-hardcoded-ui
            echo   --strict  Exit non-zero if check-transitions or check-hardcoded-ui reports warnings
            exit /b 0
        ) else if "%%~A"=="-h" (
            echo Usage: scripts.bat lint [--ci] [--native] [--all] [--strict]
            exit /b 0
        ) else if "%%~A"=="native" (
            set "NATIVE_ONLY=1"
        ) else if "%%~A"=="--native" (
            set "NATIVE_ONLY=1"
        ) else if "%%~A"=="ci" (
            set "CI_MODE=1"
        ) else if "%%~A"=="--ci" (
            set "CI_MODE=1"
        ) else if "%%~A"=="--all" (
            set "CHECK_ALL_FLAG=--all"
        ) else if "%%~A"=="--strict" (
            set "CHECK_STRICT_FLAG=--strict"
        ) else (
            echo Error: Unknown argument '%%~A' for 'lint'. Allowed: [--ci ^| --native ^| --all ^| --strict]
            exit /b 1
        )
    )
    if "!NATIVE_ONLY!"=="1" (
        call :lint_native !CI_MODE!
        exit /b !ERRORLEVEL!
    )

    echo Running Go Lint...
    if "!CI_MODE!"=="1" (
        for /f "delims=" %%i in ('gofmt -l .') do (
            echo Unformatted Go files found. Run 'scripts.bat lint' locally.
            gofmt -l .
            exit /b 1
        )
    ) else (
        gofmt -w .
        if errorlevel 1 exit /b !ERRORLEVEL!
    )
    golangci-lint run
    if errorlevel 1 exit /b !ERRORLEVEL!

    echo Running Native C++ Lint...
    call :lint_native !CI_MODE!
    if errorlevel 1 exit /b !ERRORLEVEL!

    call :generate_bindings
    if errorlevel 1 exit /b !ERRORLEVEL!

    echo Running Frontend Lint ^& Checks...
    pushd "%~dp0frontend"
    call npm run check
    if errorlevel 1 ( popd & exit /b !ERRORLEVEL! )
    call npm run lint
    if errorlevel 1 ( popd & exit /b !ERRORLEVEL! )
    if "!CI_MODE!"=="1" (
        call npm run format:check
    ) else (
        call npm run format
    )
    if errorlevel 1 ( popd & exit /b !ERRORLEVEL! )
    echo Checking AppIcon Conflicts...
    call npx tsx scripts/checkConflicts.ts
    if errorlevel 1 ( popd & exit /b !ERRORLEVEL! )
    popd

    echo Checking Themes ^& Locales...
    call :check_themes
    if errorlevel 1 exit /b !ERRORLEVEL!
    node "%~dp0scripts\check-locales.js"
    if errorlevel 1 exit /b !ERRORLEVEL!
    node "%~dp0scripts\check-settings.js"
    if errorlevel 1 exit /b !ERRORLEVEL!
    node "%~dp0scripts\check-transitions.js" !CHECK_ALL_FLAG! !CHECK_STRICT_FLAG!
    if errorlevel 1 exit /b !ERRORLEVEL!
    node "%~dp0scripts\check-hardcoded-ui.js" !CHECK_ALL_FLAG! !CHECK_STRICT_FLAG!
    if errorlevel 1 exit /b !ERRORLEVEL!
    exit /b 0
)

if "%~1"=="check-themes" (
    call :check_themes %2 %3
    exit /b !ERRORLEVEL!
)

if "%~1"=="pack-theme" (
    if "%~2"=="--help" (
        echo Usage: scripts.bat pack-theme ^<path^>
        echo Validates and packages a theme directory into a .vintatheme bundle.
        exit /b 0
    )
    if "%~2"=="-h" (
        echo Usage: scripts.bat pack-theme ^<path^>
        echo Validates and packages a theme directory into a .vintatheme bundle.
        exit /b 0
    )
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\pack-theme.ps1" %2 %3
    exit /b !ERRORLEVEL!
)

if "%~1"=="check-locales" (
    node "%~dp0scripts\check-locales.js" %2 %3 %4 %5
    exit /b !ERRORLEVEL!
)

if "%~1"=="check-transitions" (
    node "%~dp0scripts\check-transitions.js" %2 %3 %4 %5
    exit /b !ERRORLEVEL!
)

if "%~1"=="check-settings" (
    node "%~dp0scripts\check-settings.js" %2 %3 %4 %5
    exit /b !ERRORLEVEL!
)

if "%~1"=="check-hardcoded-ui" (
    node "%~dp0scripts\check-hardcoded-ui.js" %2 %3 %4 %5
    exit /b !ERRORLEVEL!
)

if "%~1"=="resource-profile" (
    if "%~2"=="--help" (
        echo Usage: scripts.bat resource-profile [options]
        echo Monitors live unscaled CPU%% and RAM across Vinta's process tree.
        exit /b 0
    )
    if "%~2"=="-h" (
        echo Usage: scripts.bat resource-profile [options]
        echo Monitors live unscaled CPU%% and RAM across Vinta's process tree.
        exit /b 0
    )
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\resource-profile.ps1" %2
    exit /b !ERRORLEVEL!
)

if "%~1"=="check-locals" (
    node "%~dp0scripts\check-locales.js" %2 %3 %4 %5
    exit /b !ERRORLEVEL!
)

if "%~1"=="sync-locales" (
    node "%~dp0scripts\sync-locales.js" %2 %3
    exit /b !ERRORLEVEL!
)

if "%~1"=="sync-locals" (
    node "%~dp0scripts\sync-locales.js" %2 %3
    exit /b !ERRORLEVEL!
)

if "%~1"=="installer" (
    if "%~2"=="--help" (
        echo Usage: scripts.bat installer
        echo Builds standalone production binary and Windows NSIS installer.
        exit /b 0
    )
    if "%~2"=="-h" (
        echo Usage: scripts.bat installer
        echo Builds standalone production binary and Windows NSIS installer.
        exit /b 0
    )
    call :kill_app
    call :build_native
    if errorlevel 1 exit /b !ERRORLEVEL!
    call :generate_icon
    if errorlevel 1 exit /b !ERRORLEVEL!
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\increment_build.ps1"
    if errorlevel 1 exit /b !ERRORLEVEL!
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\package_installer.ps1"
    exit /b !ERRORLEVEL!
)

if "%~1"=="new-ui" (
    if "%~2"=="--help" (
        node "%~dp0scripts\new-ui-item.js" --help
        exit /b !ERRORLEVEL!
    )
    if "%~2"=="-h" (
        node "%~dp0scripts\new-ui-item.js" --help
        exit /b !ERRORLEVEL!
    )
    node "%~dp0scripts\new-ui-item.js" %2 %3 %4
    exit /b !ERRORLEVEL!
)

if "%~1"=="contributor" (
    if "%~2"=="--help" (
        echo Usage: scripts.bat contributor ^<username^> ^<roles^> [display_name]
        echo Example: scripts.bat contributor octocat "code,doc" "The Octocat"
        exit /b 0
    )
    if "%~2"=="-h" (
        echo Usage: scripts.bat contributor ^<username^> ^<roles^> [display_name]
        echo Example: scripts.bat contributor octocat "code,doc" "The Octocat"
        exit /b 0
    )
    if "%~4"=="" (
        powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\add_contributor.ps1" -Username "%~2" -Roles "%~3"
    ) else (
        powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\add_contributor.ps1" -Username "%~2" -Roles "%~3" -DisplayName "%~4"
    )
    exit /b !ERRORLEVEL!
)

echo Error: Unknown command '%~1'.
:usage
echo Usage: scripts.bat ^<command^> [options]
echo.
echo Commands:
echo   doctor                   Checks development environment toolchains and dependencies
echo   setup [--install-tools]  Sets up and configures the developer environment
echo   kill                     Terminates any running Vinta.exe processes
echo   dev [--no-build^|--profile^|--frontend]  Runs Vite / Wails live development server
echo   showcase                 Runs live development server with Showcase Studio enabled
echo   build [--dev]            Builds binary (production by default, or devtest with --dev)
echo   build-native             Builds native C++/WinRT bridge DLL
echo   build-frontend           Builds frontend production bundle using Vite
echo   run [--no-build] [--debug]  Runs binary (optionally skip build, optionally enable debug mode)
echo   test [--no-build]        Runs test suite with DevTest panel
echo   icon                     Generates build\windows\icon.ico from build\appicon.png
echo   bindings                 Generates TypeScript bindings for frontend from Go backend
echo   syso                     Generates Windows syso resource file
echo   installer                Builds standalone production binary and Windows NSIS installer
echo   pack-theme ^<path^>        Packages a theme directory into a .vintatheme bundle
echo   check-themes [path]      Audits theme presets against dark-gold and validates --vinta-* tokens
echo   check-locales [--all]    Validates key parity, registration, and schema across all locale catalogs
echo   check-settings [default] Validates setting keys, defaults, and localStorage usage
echo   check-transitions [--all] [--strict]  Audits CSS transitions and reduced-motion coverage
echo   check-hardcoded-ui [--all] [--strict] Audits components/pages for hardcoded colors, strings, and native controls
echo   sync-locales [--prune] [--dry-run]    Propagates new keys from en.json to all other locale catalogs
echo   resource-profile [opts]  Monitors live unscaled CPU%% and RAM across Vinta's process tree
echo   lint [--ci^|--native^|--all^|--strict]  Runs Go, C++, Frontend, Theme, and Locale linters
echo   new-ui ^<type^> ^<Name^>     Scaffolds a component, widget, page, or overlay from templates/
echo   contributor ^<args^>       Adds a contributor via scripts\add_contributor.ps1
echo.
echo Run 'scripts.bat ^<command^> --help' for command-specific flags.
exit /b 1

:: ============================================================================
:: Internal Subroutines
:: ============================================================================

:kill_app
taskkill /IM Vinta.exe /F /T >nul 2>&1
exit /b 0

:build_native
echo Building native C++/WinRT bridge via CMake...
if not exist "build\native" (
    cmake -B build\native -S backend\services\native -G "Visual Studio 18 2026" -A x64
    if errorlevel 1 exit /b !ERRORLEVEL!
)
cmake --build build\native --config Release
if errorlevel 1 (
    echo Native C++/WinRT compilation failed.
    exit /b !ERRORLEVEL!
)
echo Native C++/WinRT bridge built successfully.
exit /b 0

:generate_icon
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\generate_icon.ps1"
exit /b !ERRORLEVEL!

:generate_bindings
wails3 generate bindings -ts
exit /b !ERRORLEVEL!

:build_frontend
node "%~dp0scripts\check-tokens.js"
if errorlevel 1 exit /b !ERRORLEVEL!
pushd "%~dp0frontend"
call npm run build
set "ERR=!ERRORLEVEL!"
popd
exit /b !ERR!

:check_themes
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\pack-theme.ps1" -CheckOnly
if errorlevel 1 exit /b !ERRORLEVEL!
node "%~dp0scripts\check-tokens.js" %1 %2
exit /b !ERRORLEVEL!

:generate_syso
wails3 generate syso -arch amd64 -icon build/windows/icon.ico -manifest build/windows/wails.exe.manifest -info build/windows/info.json -out wails_windows_amd64.syso
exit /b !ERRORLEVEL!

:do_build
set "IS_DEV=%~1"
for /f "usebackq delims=" %%T in (`powershell.exe -NoProfile -Command "[int64]([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds())"`) do set "BUILD_START_MS=%%T"
call :kill_app
call :build_native
if errorlevel 1 exit /b !ERRORLEVEL!
call :generate_icon
if errorlevel 1 exit /b !ERRORLEVEL!
call :generate_bindings
if errorlevel 1 exit /b !ERRORLEVEL!
call :build_frontend
if errorlevel 1 exit /b !ERRORLEVEL!
call :generate_syso
if errorlevel 1 exit /b !ERRORLEVEL!
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\increment_build.ps1"
if errorlevel 1 exit /b !ERRORLEVEL!

if "!IS_DEV!"=="1" (
    go build -tags devtest -o "build\bin\Vinta.exe" .
) else (
    go build -trimpath -ldflags="-w -s -H windowsgui" -o "build\bin\Vinta.exe" .
)
set "BUILD_ERR=!ERRORLEVEL!"
if exist "wails_windows_amd64.syso" del /f /q "wails_windows_amd64.syso"
if "!BUILD_ERR!"=="0" (
    for /f "usebackq delims=" %%E in (`powershell.exe -NoProfile -Command "$elapsed = (([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()) - !BUILD_START_MS!) / 1000.0; $now = Get-Date -Format 'HH:mm:ss'; ('{0:N2}s at {1}' -f $elapsed, $now)"`) do (
        echo Build completed in %%E.
    )
)
exit /b !BUILD_ERR!

:lint_native
set "IS_CI=%~1"
set "CLANG_FMT="
where clang-format >nul 2>&1
if not errorlevel 1 (
    set "CLANG_FMT=clang-format"
) else if exist "C:\Program Files\LLVM\bin\clang-format.exe" (
    set "CLANG_FMT=C:\Program Files\LLVM\bin\clang-format.exe"
)
if defined CLANG_FMT (
    for %%f in (backend\services\native\*.cpp backend\services\native\*.h) do (
        if "!IS_CI!"=="1" (
            "!CLANG_FMT!" --dry-run --Werror "%%f"
            if errorlevel 1 exit /b !ERRORLEVEL!
        ) else (
            "!CLANG_FMT!" -i "%%f"
        )
    )
) else (
    echo clang-format not found on PATH, skipping native C++ formatting.
)
exit /b 0