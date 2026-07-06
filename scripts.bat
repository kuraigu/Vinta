@echo off
if "%1"=="dev" (
    wails dev
) else if "%1"=="build" (
    wails build
) else if "%1"=="run" (
    taskkill /IM Vinta.exe /F 2>nul
    wails build
    if %errorlevel% neq 0 exit /b %errorlevel%
    start "" "build\bin\Vinta.exe"
) else if "%1"=="lint" (
    echo Running Go Lint...
    gofmt -w .
    golangci-lint run
    echo Running Frontend Lint...
    cd frontend
    call npm run lint
    call npm run format
) else (
    echo Usage: scripts.bat [dev^|build^|run^|lint]
)
