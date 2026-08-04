@echo off
call "D:\Programs\Microsoft Visual Studio\18\BuildTools\VC\Auxiliary\Build\vcvars64.bat" >nul 2>&1
cl.exe /nologo /std:c++20 /EHa /O2 /LD backend\services\native\winrt_bridge.cpp /Fe:backend\services\vintaservices.dll /link /MACHINE:X64 WindowsApp.lib Ole32.lib OleAut32.lib RuntimeObject.lib Advapi32.lib User32.lib
if %ERRORLEVEL% equ 0 (
    echo [OK] Built backend\services\vintaservices.dll successfully.
    if exist "build\bin" (
        copy /y "backend\services\vintaservices.dll" "build\bin\vintaservices.dll" >nul
        echo [OK] Copied to build\bin\vintaservices.dll
    )
) else (
    echo [ERROR] Failed to build vintaservices.dll
    exit /b 1
)
