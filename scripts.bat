@echo off
if "%1"=="dev" (
    wails dev
) else if "%1"=="build" (
    wails build
) else if "%1"=="run" (
    taskkill /IM Vinta.exe /F 2>nul
    wails build
    if errorlevel 1 exit /b 1
    start "" "build\bin\Vinta.exe"
) else if "%1"=="test" (
    taskkill /IM Vinta.exe /F 2>nul
    wails build
    if errorlevel 1 exit /b 1
    start "" "build\bin\Vinta.exe" -test
) else if "%1"=="lint" (
    echo Running Go Lint...
    gofmt -w .
    golangci-lint run
    echo Running Frontend Lint...
    cd frontend
    call npm run check
    call npm run lint
    call npm run format
) else (
    echo Usage: scripts.bat [dev^|build^|run^|test^|lint]
)
