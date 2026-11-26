# Vinta Build and Automation Scripts

This directory contains PowerShell and Node.js automation scripts used by `Taskfile.yml` and the root `scripts.bat` wrapper.

## Script Catalog

### `setup.ps1`
First-time environment setup script. Checks required developer toolchains (Go 1.26+, Node.js 20+, CMake, MSVC Build Tools, Wails v3 CLI, and optional NSIS / LLVM `clang-format`), installs Taskfile if missing, restores frontend npm packages, and configures CMake.
- Invocation: `.\scripts.bat setup [--install-tools]` or `task setup -- --install-tools`

### `doctor.ps1`
Environment diagnostic script. Validates installed dependencies, version constraints, PATH availability, and CMake generator state.
- Invocation: `task doctor` or `.\scripts.bat doctor`

### `check-tokens.js` & `pack-theme.ps1`
Validates all `var(--vinta-*)` references across `frontend/src/` against the canonical base theme (`frontend/src/themes/presets/dark-gold/style.css`), audits secondary presets and `templates/theme` for missing tokens, and packages custom themes into `.vintatheme` archives.
- Invocation: `task check:themes` (`.\scripts.bat check-themes`) and `task theme:pack -- <path>` (`.\scripts.bat pack-theme <path>`)

### `check-locales.js` & `sync-locales.js`
Validates translation key parity, placeholder tokens, `@metadata` schema, and `i18n.ts` registration across `frontend/src/locales/*.json`, reporting per-locale coverage. `sync-locales.js` propagates new keys from `en.json` to all secondary catalogs.
- Invocation: `task check:locales -- [--all]` and `task sync:locales -- [--prune] [--dry-run]`

### `check-settings.js`
Verifies unique `vinta_*` setting keys, schema migrations, and audits direct `localStorage` usage.
- Invocation: `task check:settings -- [--defaults]`

### `check-transitions.js` & `lib/lint-utils.js`
Audits production `.css` and `.svelte` files (excluding `devtest`) for non-composited CSS transitions and verifies `prefers-reduced-motion` support. Intentional outer container-bounds transitions can be documented with `/* vinta-ignore: transition */`.
- Invocation: `task check:transitions -- [--all] [--strict]` or `.\scripts.bat check-transitions [--all] [--strict]`

### `check-hardcoded-ui.js`
Scans `frontend/src/components` and `frontend/src/pages` for hardcoded colors, native HTML controls, JS CSS imports, banned icons, and hardcoded UI attribute strings.
- Invocation: `task check:hardcoded-ui -- [--all] [--changed] [--strict]`

### `resource-profile.ps1`
Samples live unscaled CPU% and working-set RAM across Vinta's Go and WebView2 process tree.
- Invocation: `task profile:resources` or `.\scripts.bat resource-profile`

### `generate_icon.ps1`
Generates multi-resolution Windows icon assets (`build/windows/icon.ico`) from `build/appicon.png` covering standard mipmap resolutions (16x16, 24x24, 32x32, 48x48, 64x64, 128x128, 256x256).
- Invocation: `task icon`

### `increment_build.ps1`
Automatically increments the build metadata counter in `build/windows/info.json` during release builds.
- Invocation: Executed automatically during `task build`

### `package_installer.ps1`
Packages the release executable and native dependencies into a standalone NSIS installer (`build/bin/Vinta-Setup-x64.exe`).
- Invocation: `task package:installer` or `.\scripts.bat installer`

## Development Workflow

For daily development tasks, use `Taskfile` (`task dev`, `task run`, `task lint`, `task build`). Pass flags to Taskfile tasks using `--` (for example, `task lint -- --all`).

To view help for an individual task (single source of truth in `scripts.bat`):
- `task <task> -- --help` (forwards `--help` to `scripts.bat`)
- `task help -- <command>` or `.\scripts.bat help <command>` (defers to `scripts.bat <command> --help`)
