@echo off
setlocal

if "%1"=="dev" (
    call :build_native
    if errorlevel 1 exit /b 1
    wails dev
) else if "%1"=="build" (
    call :build_native_force
    if errorlevel 1 exit /b 1
    wails build
) else if "%1"=="run" (
    taskkill /IM Vinta.exe /F 2>nul
    if "%2" NEQ "no_build" (
        call :build_native
        if errorlevel 1 exit /b 1
        wails build
        if errorlevel 1 exit /b 1
    )
    start "" "build\bin\Vinta.exe"
) else if "%1"=="test" (
    taskkill /IM Vinta.exe /F 2>nul
    call :build_native
    if errorlevel 1 exit /b 1
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
    cd ..
) else (
    echo Usage: scripts.bat [dev^|build^|run^|test^|lint]
)
exit /b %ERRORLEVEL%

:build_native
call :build_native_force
exit /b %ERRORLEVEL%

:build_native_force
echo Building native C++/WinRT bridge...
call "D:\Programs\Microsoft Visual Studio\18\BuildTools\VC\Auxiliary\Build\vcvars64.bat" >nul 2>&1
if errorlevel 1 (
    echo Failed to initialize MSVC build environment.
    exit /b 1
)
if not exist "build\bin" mkdir "build\bin"
cl.exe /nologo /std:c++20 /EHa /O2 /LD backend\services\native\winrt_bridge.cpp /Fe:build\bin\vintaservices.dll /link /MACHINE:X64 WindowsApp.lib Ole32.lib OleAut32.lib RuntimeObject.lib Advapi32.lib User32.lib
if errorlevel 1 (
    echo Native C++/WinRT compilation failed.
    exit /b 1
)
if exist "backend\services\native\*.obj" del /Q "backend\services\native\*.obj"
if exist "build\bin\*.exp" del /Q "build\bin\*.exp"
if exist "build\bin\*.lib" del /Q "build\bin\*.lib"
echo Native C++/WinRT bridge built successfully.
exit /b 0
