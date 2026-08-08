@echo off
setlocal

if "%1"=="dev" (
    call :build_native
    if errorlevel 1 exit /b %ERRORLEVEL%
    wails dev
    if errorlevel 1 exit /b %ERRORLEVEL%
) else if "%1"=="build" (
    call :build_native_force
    if errorlevel 1 exit /b %ERRORLEVEL%
    wails build
    if errorlevel 1 exit /b %ERRORLEVEL%
) else if "%1"=="run" (
    taskkill /IM Vinta.exe /F 2>nul
    if "%2" NEQ "no_build" (
        call :build_native
        if errorlevel 1 exit /b %ERRORLEVEL%
        wails build
        if errorlevel 1 exit /b %ERRORLEVEL%
    )
    start "" "build\bin\Vinta.exe"
) else if "%1"=="test" (
    taskkill /IM Vinta.exe /F 2>nul
    call :build_native
    if errorlevel 1 exit /b %ERRORLEVEL%
    wails build
    if errorlevel 1 exit /b %ERRORLEVEL%
    start "" "build\bin\Vinta.exe" -test
) else if "%1"=="lint" (
    echo Running Go Lint...
    gofmt -w .
    if errorlevel 1 exit /b %ERRORLEVEL%
    golangci-lint run
    if errorlevel 1 exit /b %ERRORLEVEL%
    echo Running Frontend Lint...
    cd frontend
    call npm run check
    if errorlevel 1 ( cd .. & exit /b %ERRORLEVEL% )
    call npm run lint
    if errorlevel 1 ( cd .. & exit /b %ERRORLEVEL% )
    call npm run format
    if errorlevel 1 ( cd .. & exit /b %ERRORLEVEL% )
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
where cl.exe >nul 2>&1
if errorlevel 1 (
    set "VCVARS="
    set "VSWHERE=%ProgramFiles(x86)%\Microsoft Visual Studio\Installer\vswhere.exe"
    if exist "%VSWHERE%" (
        for /f "usebackq tokens=*" %%i in (`"%VSWHERE%" -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath 2^>nul`) do (
            if exist "%%i\VC\Auxiliary\Build\vcvars64.bat" set "VCVARS=%%i\VC\Auxiliary\Build\vcvars64.bat"
        )
    )
    if defined VCVARS (
        call "%VCVARS%"
    ) else if exist "D:\Programs\Microsoft Visual Studio\18\BuildTools\VC\Auxiliary\Build\vcvars64.bat" (
        call "D:\Programs\Microsoft Visual Studio\18\BuildTools\VC\Auxiliary\Build\vcvars64.bat"
    ) else if exist "C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools\VC\Auxiliary\Build\vcvars64.bat" (
        call "C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools\VC\Auxiliary\Build\vcvars64.bat"
    ) else if exist "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat" (
        call "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat"
    ) else if exist "C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Auxiliary\Build\vcvars64.bat" (
        call "C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Auxiliary\Build\vcvars64.bat"
    ) else if exist "C:\Program Files\Microsoft Visual Studio\2022\Enterprise\VC\Auxiliary\Build\vcvars64.bat" (
        call "C:\Program Files\Microsoft Visual Studio\2022\Enterprise\VC\Auxiliary\Build\vcvars64.bat"
    )
)
where cl.exe >nul 2>&1
if errorlevel 1 (
    echo Failed to initialize MSVC build environment: cl.exe not found.
    exit /b 1
)
if not exist "build\bin" mkdir "build\bin"
cl.exe /nologo /std:c++20 /EHa /O2 /LD backend\services\native\winrt_bridge.cpp backend\services\native\notification_service.cpp backend\services\native\media_service.cpp backend\services\native\audio_peak_service.cpp /Fe:build\bin\vintaservices.dll /link /MACHINE:X64 WindowsApp.lib Ole32.lib OleAut32.lib RuntimeObject.lib Advapi32.lib User32.lib
if errorlevel 1 (
    echo Native C++/WinRT compilation failed.
    exit /b 1
)
if exist "backend\services\native\*.obj" del /Q "backend\services\native\*.obj"
if exist "build\bin\*.exp" del /Q "build\bin\*.exp"
if exist "build\bin\*.lib" del /Q "build\bin\*.lib"
echo Native C++/WinRT bridge built successfully.
exit /b 0
