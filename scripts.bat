@echo off
setlocal enabledelayedexpansion

if "%~1"=="" goto usage

if "%~1"=="doctor" (
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\doctor.ps1"
    exit /b %ERRORLEVEL%
)

if "%~1"=="setup" (
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\setup.ps1" %2 %3
    exit /b %ERRORLEVEL%
)

if "%~1"=="dev" (
    set "DO_BUILD_NATIVE=1"

    for %%A in (%2 %3 %4) do (
        if "%%~A"=="no_build" (
            set "DO_BUILD_NATIVE=0"
        ) else (
            echo Error: Unknown argument '%%~A' for 'dev'.
            echo Allowed options for 'dev': [no_build]
            exit /b 1
        )
    )

    if "!DO_BUILD_NATIVE!"=="1" (
        call :build_native
        if errorlevel 1 exit /b !ERRORLEVEL!
    )
    taskkill /IM Vinta.exe /F 2>nul
    task dev
    if errorlevel 1 exit /b !ERRORLEVEL!
    exit /b 0
)

if "%~1"=="build" (
    if "%~2" NEQ "" (
        echo Error: 'build' does not take extra arguments.
        exit /b 1
    )
    task build
    if errorlevel 1 exit /b !ERRORLEVEL!
    exit /b 0
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

    taskkill /IM Vinta.exe /F 2>nul
    if "!DO_BUILD!"=="1" (
        task build
        if errorlevel 1 exit /b !ERRORLEVEL!
    )
    if defined DEBUG_FLAG (
        start "" "build\bin\Vinta.exe" !DEBUG_FLAG!
    ) else (
        start "" "build\bin\Vinta.exe"
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
    taskkill /IM Vinta.exe /F 2>nul
    if "!DO_BUILD!"=="1" (
        task build DEV=true
        if errorlevel 1 exit /b !ERRORLEVEL!
    )
    start "" "build\bin\Vinta.exe"
    exit /b 0
)

if "%~1"=="lint" (
    if "%~2" NEQ "" (
        echo Error: 'lint' does not take extra arguments.
        exit /b 1
    )
    echo Running Go Lint...
    gofmt -w .
    if errorlevel 1 exit /b !ERRORLEVEL!
    golangci-lint run
    if errorlevel 1 exit /b !ERRORLEVEL!
    echo Running Native C++ Lint...
    where clang-format >nul 2>&1
    if not errorlevel 1 (
        for %%f in (backend\services\native\*.cpp backend\services\native\*.h) do (
            clang-format -i "%%f"
        )
    ) else (
        echo clang-format not found on PATH, skipping native C++ formatting.
    )
    echo Running Frontend Lint...
    cd frontend
    call npm run check
    if errorlevel 1 ( cd .. & exit /b !ERRORLEVEL! )
    call npm run lint
    if errorlevel 1 ( cd .. & exit /b !ERRORLEVEL! )
    call npm run format
    if errorlevel 1 ( cd .. & exit /b !ERRORLEVEL! )
    cd ..
    exit /b 0
)

if "%~1"=="installer" (
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\package_installer.ps1"
    exit /b %ERRORLEVEL%
)

echo Error: Unknown command '%~1'.
:usage
echo Usage: scripts.bat ^<command^> [options]
echo.
echo Commands:
echo   doctor                   Checks development environment toolchains and dependencies
echo   setup [-InstallTools]    Sets up and configures the developer environment
echo   dev [no_build]           Runs Vite / Wails live development server (optionally skip native build)
echo   build                    Builds production binary
echo   run [no_build] [debug]   Runs binary (optionally skip build, optionally enable debug mode)
echo   test                     Runs test suite with DevTest panel (-test)
echo   installer                Builds standalone production binary and Windows NSIS installer
echo   lint                     Runs Go and Frontend linters and formatters
exit /b 1

:build_native
:build_native_force
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

:increment_build_seq
set "SEQ_DIR=%APPDATA%\FreeMatrix Studios\Vinta"
if not exist "%SEQ_DIR%" mkdir "%SEQ_DIR%"
set "SEQ_FILE=%SEQ_DIR%\build_seq"
set "BUILD_COUNT=1"
if exist "%SEQ_FILE%" (
    set /p CUR_COUNT=<"%SEQ_FILE%"
    set /a BUILD_COUNT=!CUR_COUNT!+1
)
<nul set /p="!BUILD_COUNT!" > "%SEQ_FILE%"
exit /b 0
