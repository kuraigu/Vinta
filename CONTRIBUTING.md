# Contributing to Vinta

Thanks for helping make Vinta better. Whether you are fixing a bug, building a new widget, designing a theme, or just testing out builds, your work is appreciated.

Everyone who contributes (code, docs, bug reports, or QA testing) gets listed in [docs/CONTRIBUTORS.md](docs/CONTRIBUTORS.md) and inside the app under **Settings > About / Credits**.

## Quick Index
- [1. Ways to Contribute & Claiming Issues](#1-ways-to-contribute--claiming-issues) (testing, bug reports, claiming workflow)
- [2. Developer Setup & Commands](#2-developer-setup--commands) (setup, Taskfile commands, debug hotkeys)
- [3. How Vinta Works (Architecture)](#3-how-vinta-works-architecture) (backend, C++ bridge, satellites, page hierarchy)
- [4. Building Components & Pages](#4-building-components--pages) (component-first rules, templates, runes, scoping)
- [5. Creating & Packaging Themes](#5-creating--packaging-themes) (presets, theme packages, manifest format)
- [6. Submitting a Pull Request](#6-submitting-a-pull-request) (branching, guidelines, PR checklist)
- [7. Maintainer Release & Verification Guide](#7-maintainer-release--verification-guide) (releases, attestations, signatures)

---

## 1. Ways to Contribute & Claiming Issues

### Testing & Bug Reports (No Coding Needed)

You don't need to write code to help out:

1. Download a pre-release build or run Vinta locally.
2. Play around with the UI, switch themes, test swipe gestures, and try out the widgets.
3. If something breaks, open a GitHub Issue with clear steps to reproduce it.
4. Once confirmed, we'll tag your issue and add you to the **Testers & QA Hall of Fame**.

### Working on Issues (Claim Workflow)

To keep people from accidentally doing the same work twice, we use a simple claim system:

1. **Claim before you code:** Comment on the issue you want to tackle (e.g., _"I'd like to take this on"_). Wait for a maintainer to assign it to you before opening a Pull Request.
2. **Look for the `available` label:** Unclaimed issues are tagged `available` and are first-come, first-served. Once assigned, the label is removed. If an issue is already taken, you can still comment to get in line in case it opens back up.
3. **5 & 7-Day Inactivity Rule:** We track activity through comments, commits, or status updates so issues don't get stuck in limbo:
   - **5 days without updates:** A bot posts a quick check-in reminder and adds the `stale-claim` label.
   - **7 days without updates:** The issue is automatically unassigned and marked `available` again for someone else to grab.

---

## 2. Developer Setup & Commands

### First-Time Setup

Run the setup script to install dependencies, configure CMake, and grab the `task` CLI:

```powershell
.\scripts.bat setup
```

To make sure your compilers, SDKs, and toolchains are good to go:

```powershell
task doctor
# Or: .\scripts.bat doctor
```

### Everyday Commands

We recommend using **[go-task / Taskfile](https://taskfile.dev)**, though `.\scripts.bat` works as a fallback for everything:

| Task Command             | Batch Fallback                     | What it does                                                        |
| :----------------------- | :--------------------------------- | :------------------------------------------------------------------ |
| `task dev`               | `.\scripts.bat dev`                | Starts the live dev server with hot reload and the DevTest panel    |
| `task dev:nobuild`       | `.\scripts.bat dev no_build`       | Starts the dev server without recompiling the native C++ DLL        |
| `task run`               | `.\scripts.bat run`                | Builds the release binary and runs `Vinta.exe` in production mode   |
| `task run:debug`         | `.\scripts.bat run debug`          | Builds and runs `Vinta.exe` with debug hotkeys turned on            |
| `task run:nobuild`       | `.\scripts.bat run no_build`       | Launches the existing `.exe` without rebuilding                     |
| `task run:nobuild:debug` | `.\scripts.bat run no_build debug` | Launches the existing `.exe` with debug hotkeys turned on           |
| `task build`             | `.\scripts.bat build`              | Compiles the release binary to `build/bin/Vinta.exe`                |
| `task package:installer` | `.\scripts.bat installer`          | Builds the release binary and bundles the Windows NSIS installer    |
| `task test`              | `.\scripts.bat test`               | Builds and runs the test suite binary with the DevTest panel        |
| `task lint`              | `.\scripts.bat lint`               | Runs Go formatters, `golangci-lint`, and frontend type/style checks |
| `task doctor`            | `.\scripts.bat doctor`             | Checks your local environment, runtimes, and Windows SDKs           |
| `task setup`             | `.\scripts.bat setup`              | Installs dependencies and configures CMake                          |
| `task help`              | `task --list`                      | Lists all available tasks                                           |

### Debug Hotkeys

When running in `dev` mode or via `task run:debug`:

- **`Ctrl + Shift + D`** (or **`Ctrl + Alt + B`**): Highlights the physical Win32 window clipping region (`SetWindowRgn`) in real time so you can see where clicks register.
- **`Ctrl + Shift + I`** (or **`Ctrl + Alt + I`**): Toggles the real-time **Class Name Inspector** overlay. Hovering over any element in the main island highlights its bounding box, displays its tag, semantic class names, and `data-vinta` selector in an inline HUD pill, and streams the full CSS selector to the DevTest window where it can be copied with one click.
- **`Ctrl + Shift + L`** (or **`Ctrl + Alt + L`**): Toggles verbose console logs for window bounds and IPC messages.

---

## 3. How Vinta Works (Architecture)

### Core Stack (`Vinta.exe`)

Vinta is built with **Go + Wails v3 + Svelte 5+**. The Go backend handles native Win32 window management, transparent hit-testing, and config files, while Svelte renders the Dynamic Island UI.

### Native WinRT Bridge (`vintaservices.dll`)

Located in `backend/services/` (`winrt_bridge.cpp` and `winrt_bridge.h`), this is a native C++/WinRT library compiled alongside the app and loaded directly into the Go process via Windows syscalls.

- **Why C++/WinRT?** Windows exposes system notifications (`Windows.UI.Notifications`), media controls (`Windows.Media.Control`), and real-time audio peaks (CoreAudio COM) through native APIs. Loading a C++ DLL in-process skips child processes and .NET runtimes entirely, giving us instant callbacks and a tiny memory footprint (~3MB).
- **Direct Callbacks:** The bridge pushes media metadata, album art streams, toast notifications, and audio visualizer peaks straight to Go. Media commands (`seek`, `play_pause`, `next`, `prev`, `set_volume`) fire immediately on the active Windows session.

### Satellite Dock System

Satellites are floating utility panels and action buttons that sit around the main Dynamic Island.

- **Single-Window Multi-Region:** Spawning a separate WebView2 window for every floating button would eat 40MB to 90MB of RAM each and cause drag lag. Instead, Vinta renders the main island and all satellites inside one large transparent canvas, using dynamic multi-rect hit-testing (`SetHitRects`). Only the visible UI boxes catch mouse clicks; the empty gaps let clicks pass right through to the desktop.
- **Declarative Host:** Wrap satellite UI in `<SatelliteHost>` in Svelte. It automatically tells the Go backend its geometry (`anchor`, `width`, `height`, `radius`, `gap`, `offsetX`, `offsetY`, `active`) over IPC.
- **Screen Edge Clamping:** The island's movement bounds adjust automatically based on active satellites (`getActiveSatelliteFootprint`) so docks never get pushed off-screen.

### Components vs. Pages

- **Components (`frontend/src/components/`):** Reusable widgets like `Clock.svelte`, `Weather.svelte`, or `MediaWidget.svelte`. They live inside pages or carousels (`ComponentGallery.svelte`) and take props like `expanded` to switch between compact and full views.
- **Pages (`frontend/src/pages/`):** Top-level views like `HomePage.svelte`, `MediaPage.svelte`, or `SettingsPage.svelte`. They sit inside `App.svelte` and control overall navigation, layout, and island dimensions.

### Page Hierarchy & Minimizing

To keep navigation consistent when closing or minimizing overlays, Vinta uses a priority stack managed by `AppControls.getFallbackPage()` in `frontend/src/utils/AppControls.svelte`:

1. **`HomePage`**: Default minimal pill when the system is idle.
2. **`MediaPage`**: Default minimal pill whenever audio or video is playing.
3. **`NotificationPage`**: Temporary popup when a notification comes in.
4. **`AppsPage` / `SettingsPage`**: Full expanded views.

**Rule for minimizing:** Never write manual `if/else` checks to decide whether to return to `HomePage` or `MediaPage`. Always call `AppControls.doMinimize(navigate, e)` or `AppControls.getFallbackPage()`. It automatically drops the user back to `MediaPage` if music is playing, or `HomePage` if not.

---

## 4. Building Components & Pages

Before writing UI code, read [docs/DESIGN.md](docs/DESIGN.md) for our component-first workflow, token rules, and stylesheet splitting guidelines.

### Component-First Rule
Always check `frontend/src/components/` before creating a new component. Reuse existing primitives (such as `<BentoCard>`, `<BackButton>`, `<CustomSelect>`, `<ScrollView>`) or extend them with variants rather than writing duplicate markup and styles.

### Starter Templates

Boilerplates are available in `templates/`:

| Template      | Location                                                                                   | What's inside                                                                                                       |
| :------------ | :----------------------------------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------------------ |
| **Component** | [templates/component/ExampleComponent.svelte](templates/component/ExampleComponent.svelte) | Svelte 5+ runes (`$props`), `<BentoCard>`, and `--vinta-*` tokens.                                                  |
| **Page**      | [templates/page/ExamplePage.svelte](templates/page/ExamplePage.svelte)                     | Module `IslandConfig` bounds, `<BackButton>`, `AppControls.doMinimize` setup, `data-vinta`, and 56px dock clearance. |
| **Theme**     | [templates/theme/](templates/theme/)                                                       | Ready-to-edit `manifest.json` and `style.css` with token overrides and page-scoped selector examples.                 |

### Adding a New Component

1. Check `frontend/src/components/` first to ensure an existing component cannot handle the use case.
2. Create your `.svelte` file in `frontend/src/components/` (or copy `ExampleComponent.svelte`).
3. Use **Svelte 5+ runes** for reactivity:
   - Props: `let { expanded = false, title }: Props = $props()`
   - State: `let activeIndex = $state(0)`
   - Derived values: `let countLabel = $derived(...)`
   - Effects: When using `$effect`, pass a dedicated named function so stack traces stay easy to read.
4. Include both the compact view (`{#if !expanded}`) and expanded view (`{#else}`) if building an expandable island widget.
5. Style with `--vinta-*` variables. Reusable child components do not need `data-vinta` attributes; they inherit themes automatically via CSS tokens.
6. Register your widget in `ComponentGallery.svelte` or drop it into its parent page.

### Adding a New Page

1. Create your `.svelte` file in `frontend/src/pages/` (or copy `ExamplePage.svelte`).
2. Export the island's dimensions in a `<script module>` block:

   ```svelte
   <script module lang="ts">
     import type { IslandConfig } from '../store'

     export const config: IslandConfig = {
       minimal: { w: 320, h: 42, r: 21 },
       expanded: { w: 440, h: 420, r: 28 },
       borderColor: 'var(--vinta-border)',
       glowColor: 'transparent'
     }
   </script>
   ```

3. Add a root `data-vinta="<page-name>"` attribute on the outermost container for theme scoping.
4. Wire up your close or back navigation using `<BackButton />` or `AppControls.doMinimize(navigate, e)`.
5. Add at least `56px` of bottom padding (`padding-bottom: 56px;`) on the main scroll container so bottom dock buttons do not cover content.
6. Register the route in `frontend/src/store.ts` and `frontend/src/App.svelte`.

> **Note on large `<style>` blocks:** Inline `<style>` is fine for small widgets. Once a component or page gets complex, split its CSS into `frontend/src/components/styles/` or `frontend/src/pages/styles/` using `<style src="..."></style>`. See [docs/DESIGN.md](docs/DESIGN.md) for folder conventions.

---

## 5. Creating & Packaging Themes

Vinta supports two kinds of themes, which use the exact same file structure (`manifest.json` + `style.css`):

| Feature             | Built-in Presets (First-Party)                         | Packaged Themes (`.vintatheme`)                               |
| :------------------ | :----------------------------------------------------- | :------------------------------------------------------------ |
| **Where they live** | `frontend/src/themes/presets/<theme-id>/`              | `%APPDATA%/FreeMatrix Studios/Vinta/themes/<theme-id>/`       |
| **How they ship**   | Compiled directly into `Vinta.exe`                     | Shared as standalone `.vintatheme` zip files                  |
| **How they load**   | Statically imported in `frontend/src/themes/engine.ts` | Loaded at runtime by the Go backend (hot-reloads in Settings) |
| **Best for**        | Core brand themes and accessibility presets            | Community themes and personal setups                          |

### 1. Theme Files (`manifest.json` & `style.css`)

Every theme folder needs two files (grab the starter from [templates/theme/](templates/theme/)):

**`manifest.json`**

```json
{
	"$schema": "./schemas/vintatheme.schema.json",
	"schemaVersion": 1,
	"id": "my-custom-theme",
	"name": "My Custom Theme",
	"author": "Your Name",
	"version": "1.0.0",
	"description": "A tailored dark theme with warm amber accents.",
	"entry": "style.css",
	"accentColor": "#f59e0b",
	"mode": "dark"
}
```

**`style.css`**
Override any `--vinta-*` variables inside `:root`. Anything omitted falls back to the default built-in theme in `frontend/src/themes/presets/dark-gold/style.css` (while `tokens.css` holds base `#ff00ff` missing-token sentinels).

### 2. Targeting Specific Widgets with `data-vinta`

Themes customize specific pages and widgets by scoping selectors to their `data-vinta` attributes (`calendar-page`, `weather-page`, `action-center-page`, `notification-page`, `notes-page`, `calculator-page`, `media-page`, `main-island`, `apps-page`, etc.):

```css
[data-vinta="calendar-page"] .calendar-agenda-tray {
	background: #24252f !important;
	border: 1px solid var(--vinta-border-accent) !important;
}

[data-vinta="calendar-page"] .tray-add-btn {
	background: var(--vinta-accent-primary) !important;
	color: #000000 !important;
}
```

### 3. Testing & Packaging a Custom Theme

You don't need to rebuild Vinta to test a custom theme:

1. Drop your theme folder into `%APPDATA%/FreeMatrix Studios/Vinta/themes/<theme-id>/`.
2. Open Vinta, go to **Settings > Appearance**, hit **Refresh**, and select your theme.
3. When you are ready to share it, pack it into a `.vintatheme` file:
   ```powershell
   task theme:pack -- "path/to/my-theme"
   ```
   Users can install the resulting file via **Settings > Appearance > Import Theme**.

### 4. Adding a Built-In Preset to the Core Repo

If you are contributing a new first-party preset directly into Vinta:

1. Put the folder in `frontend/src/themes/presets/<theme-id>/`.
2. Register it in `frontend/src/themes/engine.ts` (import the manifest and `?raw` CSS, add it to `BUILTIN_THEMES`, and add a return branch in `getThemeCss`).
3. Add the option to the fallback `themeOptions` array in `frontend/src/components/settings/tabs/AppearanceTab.svelte` so it renders immediately on startup.

---

## 6. Submitting a Pull Request

1. Fork the repo and create a branch off `main`.
2. Follow the design and CSS rules in [docs/DESIGN.md](docs/DESIGN.md).
3. Run `task lint` (or `.\scripts.bat lint`) and test your changes locally.
4. Open a PR and fill out `.github/PULL_REQUEST_TEMPLATE.md`.
5. Once reviewed and merged, we'll add your contributor role and credit!

---

## 7. Maintainer Release & Verification Guide

Official releases are built on isolated GitHub Actions runners so binaries stay reproducible and verifiable.

### Cutting a Release (Maintainers Only)

1. Go to **Actions > Release Vinta** on GitHub.
2. Click **Run workflow**, enter the version tag (like `v0.2.0`), and keep **Create as a Private Draft Release** checked.
3. The runner compiles the C++ DLL, builds `Vinta-Setup-x64.exe`, generates `SHA256SUMS.txt`, and signs the build provenance.
4. Download and test the draft installer. Once everything looks good, click **Publish release**.

### Verifying Official Builds & Fork Protection

Anyone can verify that an installer came straight from the official repo without being tampered with:

- **GitHub Attestation (Sigstore):**
  ```bash
  gh attestation verify Vinta-Setup-x64.exe --owner kuraigu
  ```
  _Note on forks:_ Even though forks inherit our GitHub Actions workflows, their builds are signed with the fork owner's identity and don't have access to our signing secrets. Running the command above against a fork's binary will immediately fail. Furthermore, Vinta's built-in updater is hardcoded to only check `kuraigu/Vinta`.
- **SHA-256 Checksums:**
  ```powershell
  # PowerShell
  Get-FileHash .\Vinta-Setup-x64.exe -Algorithm SHA256
  Get-Content .\SHA256SUMS.txt
  ```
  ```bash
  # Linux / Git Bash
  sha256sum -c SHA256SUMS.txt --ignore-missing
  ```
- **Authenticode Signature (if signed with an official certificate):**
  ```powershell
  Get-AuthenticodeSignature .\Vinta-Setup-x64.exe
  ```
