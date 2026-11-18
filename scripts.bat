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

if "%~1"=="doctor" (
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\doctor.ps1"
    exit /b !ERRORLEVEL!
)

if "%~1"=="setup" (
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\setup.ps1" %2 %3
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

    set "DO_BUILD_NATIVE=1"
    set "DO_PROFILE=0"
    for %%A in (%2 %3 %4) do (
        if "%%~A"=="no_build" (
            set "DO_BUILD_NATIVE=0"
        ) else if "%%~A"=="profile" (
            set "DO_PROFILE=1"
        ) else (
            echo Error: Unknown argument '%%~A' for 'dev'.
            echo Allowed options for 'dev': [no_build ^| profile ^| frontend]
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
    if "%~2"=="dev" (
        set "DEV_BUILD=1"
    ) else if "%~2" NEQ "" (
        echo Error: Unknown argument '%~2' for 'build'. Allowed: [dev]
        exit /b 1
    )

    call :do_build !DEV_BUILD!
    exit /b !ERRORLEVEL!
)

if "%~1"=="run" (
    set "DO_BUILD=1"
    set "DEBUG_FLAG="

    for %%A in (%2 %3 %4) do (
        if "%%~A"=="no_build" (
            set "DO_BUILD=0"
        ) else if "%%~A"=="debug" (
            set "DEBUG_FLAG=-debug"
        ) else (
            echo Error: Unknown argument '%%~A' for 'run'.
            echo Allowed options for 'run': [no_build] [debug]
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
        if "%%~A"=="no_build" (
            set "DO_BUILD=0"
        ) else (
            echo Error: Unknown argument '%%~A' for 'test'.
            echo Allowed options for 'test': [no_build]
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
    if "%~2"=="native" (
        call :lint_native 0
        exit /b !ERRORLEVEL!
    )

    set "CI_MODE=0"
    if "%~2"=="ci" (
        set "CI_MODE=1"
    ) else if "%~2" NEQ "" (
        echo Error: Unknown argument '%~2' for 'lint'. Allowed: [ci ^| native]
        exit /b 1
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
    node "%~dp0scripts\check-transitions.js"
    if errorlevel 1 exit /b !ERRORLEVEL!
    exit /b 0
)

if "%~1"=="check-themes" (
    call :check_themes
    exit /b !ERRORLEVEL!
)

if "%~1"=="pack-theme" (
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\pack-theme.ps1" %2 %3
    exit /b !ERRORLEVEL!
)

if "%~1"=="check-locales" (
    node "%~dp0scripts\check-locales.js"
    exit /b !ERRORLEVEL!
)

if "%~1"=="check-transitions" (
    node "%~dp0scripts\check-transitions.js" %2
    exit /b !ERRORLEVEL!
)

if "%~1"=="check-locals" (
    node "%~dp0scripts\check-locales.js"
    exit /b !ERRORLEVEL!
)

if "%~1"=="sync-locales" (
    node "%~dp0scripts\sync-locales.js" %2
    exit /b !ERRORLEVEL!
)

if "%~1"=="sync-locals" (
    node "%~dp0scripts\sync-locales.js" %2
    exit /b !ERRORLEVEL!
)

if "%~1"=="installer" (
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

if "%~1"=="contributor" (
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
echo   setup [-InstallTools]    Sets up and configures the developer environment
echo   kill                     Terminates any running Vinta.exe processes
echo   dev [no_build^|frontend]  Runs Vite / Wails live development server
echo   showcase                 Runs live development server with Showcase Studio enabled
echo   build [dev]              Builds binary (production by default, or devtest with 'dev')
echo   build-native             Builds native C++/WinRT bridge DLL
echo   build-frontend           Builds frontend production bundle using Vite
echo   run [no_build] [debug]   Runs binary (optionally skip build, optionally enable debug mode)
echo   test [no_build]          Runs test suite with DevTest panel
echo   icon                     Generates build\windows\icon.ico from build\appicon.png
echo   bindings                 Generates TypeScript bindings for frontend from Go backend
echo   syso                     Generates Windows syso resource file
echo   installer                Builds standalone production binary and Windows NSIS installer
echo   pack-theme ^<path^>        Packages a theme directory into a .vintatheme bundle
echo   check-locales            Validates key parity and schema across all locale catalogs
echo   sync-locales [--prune]   Propagates new keys from en.json to all other locale catalogs
echo   lint [ci^|native]         Runs Go, C++, Frontend, Theme, and Locale linters
echo   contributor ^<args^>       Adds a contributor via scripts\add_contributor.ps1
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
node "%~dp0scripts\check-tokens.js"
exit /b !ERRORLEVEL!

:generate_syso
wails3 generate syso -arch amd64 -icon build/windows/icon.ico -manifest build/windows/wails.exe.manifest -info build/windows/info.json -out wails_windows_amd64.syso
exit /b !ERRORLEVEL!

:do_build
set "IS_DEV=%~1"
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
exit /b !BUILD_ERR!

:lint_native
set "IS_CI=%~1"
where clang-format >nul 2>&1
if not errorlevel 1 (
    for %%f in (backend\services\native\*.cpp backend\services\native\*.h) do (
        if "!IS_CI!"=="1" (
            clang-format --dry-run --Werror "%%f"
            if errorlevel 1 exit /b !ERRORLEVEL!
        ) else (
            clang-format -i "%%f"
        )
    )
) else (
    echo clang-format not found on PATH, skipping native C++ formatting.
)
exit /b 0