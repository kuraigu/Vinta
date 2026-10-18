# Vinta Backend

The backend layer is written in Go and interfaces with Windows APIs, the Wails v3 application runtime, and an in-process native C++/WinRT bridge (`vintaservices.dll`).

## Architecture Overview

```
backend/
├── app.go             # Application struct, lifecycle hooks, and initialization
├── app_window.go      # Win32 HWND management, foreground hooks, display metrics, and power throttling
├── app_hittest.go     # Transparent canvas hit-testing loop and multi-rect input coordination
├── app_settings.go    # Settings JSON persistence and hardware acceleration configuration
├── app_notes.go       # Sticky note storage, legacy migration, and file system operations
├── app_shortcuts.go   # Global accelerator registration and hotkey dispatching
├── app_storage.go     # Calendar, timer, and DPAPI-encrypted notification persistence
├── app_system.go      # Autostart registration, system settings launcher, and hardware status
├── app_media.go       # Media session playback, thumbnail extraction, and app focus bridge
├── cache.go           # File cache helpers
├── icon_manager.go    # Application icon extraction and cache indexing
├── logger.go          # Runtime diagnostic logger
├── updater.go         # Background updater and release installer
├── version.go         # Version stamping and repository metadata
└── services/          # In-process C++/WinRT native bridge and IPC channels
```

## Key Subsystems

### 1. Transparent Canvas & Multi-Rect Hit Testing
Vinta uses a single-window transparent architecture across multi-monitor virtual desktop coordinates. Rather than clipping physical window regions via Win32 `SetWindowRgn`, the application operates an unclipped transparent window and dynamically toggles `WS_EX_TRANSPARENT` via a 60Hz hit-testing loop (`SetHitRects`) based on cursor proximity to active visual elements.

### 2. Native WinRT Services Bridge (`services/`)
The native services layer compiles to `vintaservices.dll`, an in-process C++/WinRT library that interacts directly with:
- `Windows.Media.Control`: Global System Media Transport Controls (GSMTC) session tracking, playback controls, and album art streams.
- `Windows.UI.Notifications`: Native toast notification listener and badge retrieval.
- CoreAudio COM APIs: Real-time dynamic audio visualizer peaks.

### 3. Data Storage & Security
Configuration and runtime states are saved under `%APPDATA%/FreeMatrix Studios/Vinta`:
- `settings.json`: General preferences, theme selections, and custom hotkeys.
- `notifications.json`: Persisted notifications encrypted at rest via the Windows Data Protection API (DPAPI).
- `notes/`: Sticky notes saved in `.vnote` JSON format.
- `calendar.json` & `timer.json`: Events, alarms, and focus configuration.
