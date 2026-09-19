# Vinta Design System Guidelines

This document defines the visual guidelines, visual aesthetics, theme compliance rules, and icon standards for building components and pages in Vinta.

All UI contributions must follow these guidelines to maintain a clean and responsive desktop experience.

## Core Design Principles

### 1. Theme CSS Variables

Never use hardcoded hex or RGB colors like `#ffffff` or `rgb(30, 30, 30)` inside component styles. All visual elements must use theme CSS tokens so themes can switch cleanly.

### 2. Iconography Standards

- **Primary Icon Library:** Always use **Lucide Icons** (`lucide-svelte`) for standard UI icons.
- **Custom SVG Icons:** If a required icon is not available in Lucide, place custom SVG helpers inside `frontend/src/assets/`. Application icons for notifications and media sessions are extracted authentically at runtime from Windows executables and packaged Store apps via `frontend/src/services/iconService.ts`.
- **SVG Color Rule:** Custom SVG paths must use `fill="currentColor"` or `stroke="currentColor"` so they adapt to parent theme CSS variables.

### 3. Custom Themes and Asset Rules

Vinta supports custom community themes (including anime-styled or custom background themes).

- **Asset Hosting:** Theme background images and assets belong in `frontend/src/assets/themes/` or can reference external user URLs and local file paths.
- **Copyrighted Materials Rule:** Do not commit copyrighted artwork, anime images, or trademarked media directly into the official core repository. Themes requiring custom artwork should use public domain or CC0 licensed media, or allow users to supply their own local file paths.

### 4. Glassmorphism and Backdrops

- Use subtle translucency with `backdrop-filter: blur(16px)` for floating island containers.
- Use radial glow backdrops (`glow-backdrop`) with translucent accent gradients to create depth.
- Keep border radii smooth and consistent (`12px` to `20px`).

### 5. Dynamic Container Bounds and Heights

- Calculate element bounds dynamically from wrapped content instead of hardcoding static pixel offsets.
- When expanding the island container for new widgets, adjust the page container height smoothly.

### 6. Gestures and Micro-Animations

- Use the reusable `use:swipe` Svelte action (`frontend/src/utils/useSwipe.ts`) for touch and mouse drag interactions.
- Keep animations at 60fps using GPU-accelerated CSS properties like `transform` and `opacity`.
- Use smooth cubic-bezier transitions (`transition: transform 0.4s cubic-bezier(0.22, 1, 0.36, 1)`).

### 7. Typography, Truncation, and Tooltips

- Use clean sans-serif system font stacks with explicit font weights (`500`, `600`, `700`, `800`).
- Apply `text-overflow: ellipsis`, `overflow: hidden`, and `white-space: nowrap` with a `max-width` on dynamic text fields (such as city names and timezones) to prevent overflow.
- Always include a native `title={text}` attribute on truncated elements so users can hover to read full strings.

### 8. Non-Distracting Indicator States

- Avoid overly fast or annoying blinking animations. Keep indicator states clean and steady to maintain a calm user experience.

### 9. Custom Select Component Standards

- **Never use native HTML `<select>` elements**: Standard `<select>` elements create native OS popups that bleed through transparent Tauri/Wails webview boundaries, clip inside scrollable containers (`overflow: hidden` / `overflow-y: auto`), and break when hovering over background elements.
- **Use `<CustomSelect>`**: Always import and use `CustomSelect` from `frontend/src/components/CustomSelect.svelte`.
- **Swipe Protection**: `<CustomSelect>` includes `.no-swipe` and pointer event suppression to ensure opening and scrolling options inside carousel viewports (`ComponentGallery`) does not trigger unwanted slide navigation.

### 10. Component & Page Stylesheet Splitting

- **Scoped Styling via `<style src="...">`**: When a page or component grows in length and styling complexity, extract its `<style>` block into a dedicated `.css` file.
- **File Naming & Directory Structure**:
  - Keep stylesheet filenames identical to component filenames (`<Name>.svelte` ➔ `<Name>.css`).
  - **Pages**: Place extracted stylesheets in `frontend/src/pages/styles/<PageName>.css` (linked via `<style src="./styles/<PageName>.css"></style>`).
  - **Components & Sub-groups**: Place extracted stylesheets under `frontend/src/components/styles/` mirroring the folder structure (e.g. `src/components/notes/NoteLive.svelte` ➔ `src/components/styles/notes/NoteLive.css` linked via `<style src="../styles/notes/NoteLive.css"></style>`).
- **No Global Script Imports**: Never use `import './styles.css'` inside `<script>` for component styling, as it bypasses scoping and leaks styles globally. Always use `<style src="...">` to retain Svelte's compile-time scoping hashes.

### 11. Floating Satellites and Docks

Vinta supports anchored satellite docks that float beside or below the main Dynamic Island while retaining transparent desktop pass-through in visual gaps.

#### Architectural Principles

1. **Single-Window Multi-Region Model:** Rather than spawning multiple OS windows or secondary WebView2 runtimes (which consume 40MB to 90MB of RAM per instance and suffer from mouse drag lag and DWM frame tearing), Vinta renders all floating elements inside a single shared webview. Visual gaps between the main island and satellite docks are truly transparent to mouse clicks via Win32 `CombineRgn(..., RGN_OR)` with `CreateRoundRectRgn`.
2. **Mounting via `<SatelliteHost>`:** Never use ad-hoc positioned containers for floating widgets. Always mount floating satellites using the declarative `<SatelliteHost>` component (`frontend/src/components/SatelliteHost.svelte`).
3. **Automatic Backend Synchronization:** `<SatelliteHost>` automatically synchronizes its geometry with the Go backend via `SetSatelliteBounds` and cleans up on unmount via `RemoveSatellite`. The Go compositor creates an `HRGN` for each active satellite and merges it into the primary window hit region.
4. **Snappy Transitions:** Exit animations must use `cubicOut` easing with short durations (110ms exit, 220ms enter). The non-zero initial slope of `cubicOut` ($f'(0) = 3$) eliminates perceived delay when collapsing.

#### Component Props and Parameters

The `<SatelliteHost>` component accepts the following configuration:

| Prop | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `id` | `string` | **required** | Unique identifier for the satellite (for example, `"controls"` or `"dev-tools"`). |
| `anchor` | `'left' \| 'right' \| 'bottom' \| 'top'` | `'right'` | The edge of the main island container to anchor against. |
| `width` | `number` | **required** | Visual width in CSS pixels. |
| `height` | `number` | **required** | Visual height in CSS pixels. |
| `radius` | `number` | `20` | Corner radius for the Win32 rounded hit-test region. |
| `gap` | `number` | `12` | Distance in pixels between the island boundary and the satellite. |
| `offsetX` | `number` | `0` | Additional horizontal pixel offset (positive values extend outward). |
| `offsetY` | `number` | `0` | Additional vertical pixel offset. |
| `active` | `boolean` | **required** | Controls both visual display and backend Win32 click region registration. |

#### Multi-Dock Stacking Model

When multiple satellites share an anchor (for example, two satellites anchored to the right flank), their positions can be coordinated using `offsetX` and `offsetY`:

- **Horizontal Row Stacking (Side by Side):**
  - First dock: `offsetX = 0`
  - Second dock: `offsetX = firstWidth + gap`
  - Total flank footprint: `firstWidth + secondWidth + (gap * 2)`
- **Vertical Column Stacking (Top to Bottom):**
  - First dock: `offsetY = 0`
  - Second dock: `offsetY = firstHeight + gap`
  - Total flank footprint: `max(firstWidth, secondWidth) + gap`

#### Edge Safety and Clamping

Desktop edge boundaries are dynamically calculated using active satellite footprints. The helper `getActiveSatelliteFootprint(anchor)` in `frontend/src/stores/satellites.ts` scans all active satellites and determines the maximum outward reach. The main island's horizontal drag clamping automatically reserves this margin so that floating satellites never clip off the physical screen.

#### Standard Window Controls Satellite (`FloatingControls.svelte`)

The window controls dock floats on the right flank and provides direct system controls:

1. **Button Order (Top to Bottom):**
   - **Quit** (`Power` icon, red hover accent): Located at the top to prevent accidental clicks.
   - **Hide** (`EyeOff` icon, green hover accent): Minimizes the widget to the system notification tray.
   - **Collapse / Minimize** (`Minimize2` icon, yellow hover accent): Located at the bottom nearest to everyday island content for optimal ergonomics.
2. **Accessibility Standards:** The controls container must include `role="toolbar"`, `aria-label="Window controls"`, and `tabindex="-1"`.
3. **Dynamic Tooltips:** Tooltips adapt directionally (`tooltipPosition="left" | "right"`) based on proximity to the right monitor edge to prevent tooltip clipping.

