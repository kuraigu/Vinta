@echo off
setlocal
set "SCRIPT_DIR=%~dp0"
set "ROOT_DIR=%SCRIPT_DIR%..\..\.."

where cl.exe >nul 2>&1
if %ERRORLEVEL% neq 0 (
    if exist "D:\Programs\Microsoft Visual Studio\18\BuildTools\VC\Auxiliary\Build\vcvars64.bat" (
        call "D:\Programs\Microsoft Visual Studio\18\BuildTools\VC\Auxiliary\Build\vcvars64.bat" >nul 2>&1
    ) else if exist "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat" (
        call "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat" >nul 2>&1
    )
)

pushd "%ROOT_DIR%"
cl.exe /nologo /std:c++20 /EHa /O2 /LD backend\services\native\winrt_bridge.cpp backend\services\native\notification_service.cpp backend\services\native\media_service.cpp backend\services\native\audio_peak_service.cpp /Fe:backend\services\vintaservices.dll /link /MACHINE:X64 WindowsApp.lib Ole32.lib OleAut32.lib RuntimeObject.lib Advapi32.lib User32.lib
if %ERRORLEVEL% equ 0 (
    echo [OK] Built backend\services\vintaservices.dll successfully.
    if exist "build\bin" (
        copy /y "backend\services\vintaservices.dll" "build\bin\vintaservices.dll" >nul
        echo [OK] Copied to build\bin\vintaservices.dll
    )
    popd
    exit /b 0
) else (
    echo [ERROR] Failed to build vintaservices.dll
    popd
    exit /b 1
)
