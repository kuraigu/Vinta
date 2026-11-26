# Vinta Backend

The backend layer is written in Go and interfaces with Windows APIs, the Wails v3 application runtime, and an in-process native C++/WinRT bridge (`vintaservices.dll`).

## Architecture Overview

```
backend/
├── api_client.go      # Shared HTTP client with buffer pooling and timeout defaults
├── app.go             # Application struct, lifecycle hooks, and initialization
├── app_window.go      # Win32 HWND management, Raw Input listener, foreground hooks, display metrics, and power throttling
├── app_hittest.go     # Event-driven WM_INPUT hit-testing reconciliation and multi-rect input coordination
├── app_launcher.go    # Installed desktop app discovery and launching
├── app_settings.go    # Settings JSON persistence and hardware acceleration configuration
├── app_notes.go       # Sticky note storage, legacy migration, and file system operations
├── app_shortcuts.go   # Global accelerator registration and hotkey dispatching
├── app_storage.go     # Calendar, timer, flashcards, habits, and DPAPI-encrypted notification persistence
├── app_system.go      # Autostart registration, system settings launcher, and hardware status
├── app_media.go       # Media session playback, thumbnail extraction, and app focus bridge
├── cache.go           # File cache helpers
├── cloud_sync.go      # Google OAuth PKCE sign-in and Firebase Cloud Sync backup/restore
├── credentials.go     # Windows DPAPI-encrypted local credential vault
├── dev_integrations.go # GitHub Device Flow auth, Developer Hub queries, and AI provider usage
├── env_config.go      # Build-time and .env runtime configuration loader
├── google_calendar.go # Google Calendar API v3 sync and token refresh
├── holidays.go        # Public holiday fetching and regional caching
├── icon_manager.go    # Application icon extraction and cache indexing
├── logger.go          # Runtime diagnostic logger
├── theme.go           # Theme engine service for user-installed .vintatheme packages
├── updater.go         # Background updater and release installer
├── version.go         # Version stamping and repository metadata
└── services/          # In-process C++/WinRT native bridge and IPC channels
```

## Key Subsystems

### 1. Transparent Canvas & Multi-Rect Hit Testing
Vinta uses a single-window transparent architecture across multi-monitor virtual desktop coordinates. Rather than clipping physical window regions via Win32 `SetWindowRgn` (which clips animations and shadows) or installing a global `WH_MOUSE_LL` hook (which sits on the OS input path and can be silently unhooked by Windows), the overlay registers mouse Raw Input (`WM_INPUT` with `RIDEV_INPUTSINK`) on its hidden Win32 message window to reconcile `WS_EX_TRANSPARENT` in under 1ms whenever the cursor moves or `SetHitRects` updates. A lightweight 150ms housekeeping loop backs up Raw Input for touch/pen input, satellite gap-hold expiry, force-interactive timeouts, and HWND recovery (falling back to 16ms polling if Raw Input registration is unavailable). `setClickThrough` is the single writer of `WS_EX_TRANSPARENT` and checks the live window style so click-through state cannot drift.

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
- `themes/`: Custom community themes and unpacked `.vintatheme` packages.
