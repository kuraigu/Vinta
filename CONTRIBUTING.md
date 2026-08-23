# Contributing to Vinta

Thank you for your interest in contributing to Vinta.

Whether you are submitting a bug fix, improving documentation, or proposing a new feature, your contributions help keep the project moving forward.

## Getting Credited

All contributors, whether you write code, update documentation, report bugs, or test release packages, will be listed in [docs/CONTRIBUTORS.md](docs/CONTRIBUTORS.md) and inside the Vinta desktop application under **Settings > About / Credits**!

## Understanding Vinta Architecture

### Main Application (`Vinta.exe`)

Vinta is built using **Go + Wails v2 + Svelte 5**. The Go backend handles native Win32 window region clipping, click-through bounds, and configuration storage, while Svelte renders the Dynamic Island interface.

### Native WinRT Services Bridge (`vintaservices.dll`)

- **What it is:** A high-performance, native C++/WinRT in-process library located in `backend/services/` (`winrt_bridge.cpp` and `winrt_bridge.h`).
- **In-Process Architecture:** `vintaservices.dll` is compiled alongside the application and dynamically loaded in-process by the Go backend via standard Windows syscalls. There are no secondary child processes or external .NET runtimes required.
- **Why C++/WinRT is used:** Windows provides native WinRT APIs (`Windows.UI.Notifications` for system toast listeners, `Windows.Media.Control` for Global System Media Transport Controls, and CoreAudio COM APIs for real-time peak audio metering). Native C++/WinRT provides direct access with zero inter-process overhead, instant event dispatching, and a minimal memory footprint (~3MB).
- **Direct Event Callbacks & Media Controls:** The bridge directly invokes Go callbacks for real-time media metadata, album art streams, toast notifications, and dynamic audio visualizer peaks. Control commands (`seek`, `play_pause`, `next`, `prev`, `set_volume`) are executed immediately on the active media session.

### Vinta Component vs Page

#### Vinta Component
- **What it is:** A reusable widget located in `frontend/src/components/`, such as `Clock.svelte`, `Weather.svelte`, `Calendar.svelte`, `MediaWidget.svelte`, or `BentoGrid.svelte`.
- **How it works:** Components are designed to sit inside pages or carousels like `ComponentGallery.svelte`. They accept props such as `expanded` to adapt between minimal and expanded views.

#### Vinta Page
- **What it is:** A top-level view located in `frontend/src/pages/`, such as `HomePage.svelte`, `MediaPage.svelte`, `AppsPage.svelte`, `NotificationPage.svelte`, or `SettingsPage.svelte`.
- **How it works:** Pages sit inside `App.svelte` and handle overall screen navigation, island height, and component layout.

### Page Hierarchy & Minimization Standard

To maintain a consistent, predictable user experience when closing, collapsing, or minimizing sub-apps and overlays, Vinta uses a standardized page hierarchy managed via `AppControls.getFallbackPage()` in `frontend/src/utils/AppControls.svelte`:

1. **Page Priority Stack**:
   - **`HomePage`**: Base island minimal page when system is idle.
   - **`MediaPage`**: Preferred minimal page whenever audio/video media is actively playing.
   - **`NotificationPage`**: Temporary alert overlay page displayed when an incoming notification arrives.
   - **`AppsPage` / `SettingsPage`**: Expanded sub-applications.

2. **Standardized Minimization (`AppControls.doMinimize(navigate, e)`)**:
   - Never hardcode manual `if/else` checks for `HomePage` or `MediaPage` inside components.
   - Always call `AppControls.doMinimize(navigate, e)` or `AppControls.getFallbackPage()` when closing or minimizing a page.
   - **Automated Fallback**: If media is playing, the island automatically minimizes back to **`MediaPage`**; if media is stopped or paused, it minimizes to **`HomePage`**.

## How to Create New Components, Pages, and Themes

### Creating a New Component

1. Create a new `.svelte` file inside `frontend/src/components/` (such as `MyWidget.svelte`).
2. Accept the `expanded` prop: `let { expanded = false } = $props<{ expanded?: boolean }>()`.
3. Build both minimal view (`{#if !expanded}`) and expanded view (`{#else}`).
4. Use design system CSS theme variables like `var(--c-primary)` and `var(--c-item-bg)` described in [docs/DESIGN.md](docs/DESIGN.md).
5. Add your new component to `ComponentGallery.svelte` or `HomePage.svelte`.

### Creating a New Page

1. Create a new `.svelte` file inside `frontend/src/pages/` (such as `MyPage.svelte`).
2. Style the page container using design system CSS variables.
3. Import and arrange your components.
4. Register the new page route inside `frontend/src/store.ts` and `frontend/src/App.svelte`.

### Component & Page CSS Splitting Guidelines

To keep code readable, modular, and easy to maintain as features grow:

- **When to Split**: CSS and Svelte do not necessarily need to be separated for small, concise components (inline `<style>` is completely fine for simple widgets). However, as a page or component grows in length and styling complexity, its styles should be extracted into a companion `.css` file.
- **File Naming & Organization**:
  - The companion `.css` file name must match its `.svelte` counterpart exactly (e.g. `HomePage.svelte` ➔ `HomePage.css`).
  - **Pages**: Place extracted stylesheets in `frontend/src/pages/styles/` and reference them via `<style src="./styles/PageName.css"></style>`.
  - **Components & Sub-groups**: Place extracted stylesheets under `frontend/src/components/styles/` mirroring any folder hierarchy (e.g. `src/components/notes/NoteLive.svelte` ➔ `src/components/styles/notes/NoteLive.css` referenced via `<style src="../styles/notes/NoteLive.css"></style>`).
- **Preserve Scoping**: Always use `<style src="...">` rather than raw script imports to preserve Svelte's compile-time component scoping and prevent global CSS leakage.
- **Theme Compliance**: All extracted styles must strictly consume CSS variables defined in [docs/DESIGN.md](docs/DESIGN.md).

### Creating a Theme

Vinta uses a tokenized CSS variable system so that all UI components adapt cleanly between themes without hardcoding colors inside components.

1. **Create Theme CSS File**: Add a new `.css` file inside `frontend/src/themes/` (e.g. `mytheme.css`).
2. **Define Theme Tokens**: Target the root selector `:root[data-theme='mytheme']` and declare all semantic design tokens:
   ```css
   :root[data-theme='mytheme'] {
     /* ACCENT COLORS */
     --c-primary: #6366f1;
     --c-danger: #ff3b30;
     --c-warning: #ffcc00;
     --c-success: #34c759;
     --c-purple: #a78bfa;

     /* BACKGROUNDS & SURFACES */
     --c-bg: rgba(15, 15, 20, 1);
     --c-border: rgba(255, 255, 255, 0.1);
     --c-hover: rgba(255, 255, 255, 0.08);
     --c-item-bg: rgba(255, 255, 255, 0.05);

     /* TYPOGRAPHY */
     --c-text-primary: #ffffff;
     --c-text-secondary: #b0b0b0;
     --c-text-muted: #888888;

     /* MEDIA & AMBIENT SCRIM */
     --c-scrim: rgba(0, 0, 0, 0.72);
   }
   ```
3. **Import in `style.css`**: Add `@import './themes/mytheme.css';` at the top of `frontend/src/style.css`.
4. **Register Theme Option**: Add your theme definition to `AppTheme` in `frontend/src/store.ts` and append `{ label: 'My Theme', value: 'mytheme' }` to `themeOptions` in `frontend/src/pages/SettingsPage.svelte`.
5. **Verify Design Guidelines**: Ensure components consume `var(--c-token)` variables without hardcoded hex values (see [docs/DESIGN.md](docs/DESIGN.md)).

## Helper Scripts & Debugging

Use `.\scripts.bat` in your terminal for development, testing, and debugging workflows:

- `.\scripts.bat dev`: Runs the live development server with hot reloading and developer debug tools automatically enabled.
- `.\scripts.bat run`: Builds the application and launches `Vinta.exe` in standard production mode.
- `.\scripts.bat run debug`: Builds and launches `Vinta.exe` with developer debug tools and hotkeys enabled.
- `.\scripts.bat run no_build`: Launches the pre-built `Vinta.exe` directly without rebuilding.
- `.\scripts.bat run no_build debug`: Launches the pre-built executable with debug mode enabled.
- `.\scripts.bat build`: Compiles the release binary to `build/bin/Vinta.exe`.
- `.\scripts.bat test`: Runs test suite checks with the DevTest panel.
- `.\scripts.bat lint`: Runs Go formatting, `golangci-lint`, and frontend type-checking/linters.

### Developer Debug Hotkeys

When running in `dev` mode or launched via `.\scripts.bat run debug`:
- **`Ctrl + Shift + D`** (or **`Ctrl + Alt + B`**): Toggles the physical Win32 window clipping region (`SetWindowRgn`) highlight overlay in real-time.
- **`Ctrl + Shift + L`** (or **`Ctrl + Alt + L`**): Toggles verbose console logging for window bounds and IPC communication.

Please remember to run `.\scripts.bat lint` before opening a Pull Request.

## Working on Issues & Claim Workflow

To keep development organized, prevent duplicated effort, and ensure fair opportunities for contributors:

1. **Claim before you code:** Before starting work or opening a Pull Request, comment on the issue asking to be assigned (e.g. "I'd like to work on this"). A maintainer will assign the issue to you once confirmed. Please do not open a PR for an issue you have not been assigned to first to avoid duplicating someone else's in-progress work.
2. **`available` Label & One Assignee at a Time:** Unclaimed issues are automatically tagged with the `available` label. Once assigned, the `available` label is removed and the issue is considered claimed. If you are interested in an already-assigned issue, comment to queue up as next in line if it reopens.
3. **Inactivity Counting & Auto-Avail (5 & 7-Day Policy):** Assigned issues require active progress. Inactivity is tracked using the latest assignee comment, commit reference, or status update:
   - **5 Days Inactive:** An automated reminder comment is posted on the issue and tagged `stale-claim`.
   - **7 Days Inactive:** The issue claim expires. It is automatically unassigned, tagged `available`, and reopened for anyone to claim.
4. **Unclaimed Issues:** Issues labeled `available` (or with no assignees) are open for anyone to claim, first-come first-served via an issue comment.

## Testing and Bug Reporting

You do not need to write code to contribute! Testing is just as important:

1. Download a pre-release package or run Vinta locally.
2. Test the UI, theme switching, swipe gestures, and widgets.
3. If you find a bug, open a GitHub Issue with clear steps on how to reproduce it.
4. Once verified, maintainers will tag your issue and add you to the **Testers & QA Hall of Fame**!

## Submitting a Pull Request

1. Fork the repository and create your branch from `main`.
2. Make your changes using the rules in [docs/DESIGN.md](docs/DESIGN.md).
3. Run `.\scripts.bat lint` and test your work.
4. Fill out the PR template in `.github/PULL_REQUEST_TEMPLATE.md`.
5. Maintainers will review your PR and assign your contributor role upon merge.
