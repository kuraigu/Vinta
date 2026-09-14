# Build Directory

The build directory houses application build configurations, assets, and platform manifests for Vinta (Wails v3).

## Directory Structure

* `bin/` - Output directory for compiled application binaries (`Vinta.exe`) and installers.
* `config.yml` - Wails v3 application metadata and development mode orchestration.
* `appicon.png` - Source high-resolution master application icon.
* `native/` - CMake build tree for the native C++/WinRT bridge (`vintaservices.dll`).
* `windows/` - Windows platform resource definitions:
  - `icon.ico` - Multi-resolution icon generated via `scripts/generate_icon.ps1`.
  - `info.json` - PE VERSIONINFO string table metadata embedded via `wails3 generate syso`.
  - `wails.exe.manifest` - Windows application execution and HiDPI per-monitor manifest.
  - `installer/` - NSIS Windows installer script (`project.nsi`) and helper macros (`wails_tools.nsh`).

## Build Commands

Vinta uses **[Taskfile](https://taskfile.dev)** for standardized build tasks:

- `task dev` - Starts live development mode with Vite and Wails v3.
- `task build` - Generates native bridge DLL, multi-res icon, frontend production bundle, Windows syso resource, and compiles `build/bin/Vinta.exe`.
- `task package:installer` - Builds the application and compiles `build/bin/Vinta-Setup-x64.exe` using NSIS (`makensis`).