@echo off
setlocal
cd /d "%~dp0\..\..\.."
if not exist "build\native" (
    cmake -B build\native -S backend\services\native
)
cmake --build build\native --config Release
exit /b %ERRORLEVEL%
