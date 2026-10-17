# Vinta Build and Automation Scripts

This directory contains PowerShell automation scripts used by `Taskfile.yml` and the root `scripts.bat` wrapper.

## Script Catalog

### `setup.ps1`
First-time environment setup script. Checks required developer toolchains (Go 1.26+, Node.js 20+, CMake, MSVC Build Tools, Wails v3 CLI), installs Taskfile if missing, restores frontend npm packages, configures CMake, and builds the native C++/WinRT bridge (`vintaservices.dll`).
- Invocation: `.\scripts.bat setup` or `task setup`

### `doctor.ps1`
Environment diagnostic script. Validates all installed dependencies, version constraints, PATH availability, and CMake generator state.
- Invocation: `task doctor` or `.\scripts.bat doctor`

### `generate_icon.ps1`
Generates multi-resolution Windows icon assets (`build/windows/icon.ico`) from `build/appicon.png` covering standard mipmap resolutions (16x16, 24x24, 32x32, 48x48, 64x64, 128x128, 256x256).
- Invocation: `task icon`

### `increment_build.ps1`
Automatically increments the build metadata counter in `build/windows/info.json` during release builds.
- Invocation: Executed automatically during `task build`

### `package_installer.ps1`
Packages the release executable and native dependencies into a standalone NSIS installer (`build/bin/Vinta-Setup-x64.exe`).
- Invocation: `task package:installer` or `.\scripts.bat package`

## Development Workflow

For daily development tasks, use `Taskfile` (`task dev`, `task run`, `task lint`, `task build`). The root `scripts.bat` script is primarily intended for initial developer onboarding via `.\scripts.bat setup` and as a fallback when Taskfile is not yet on the system PATH.
