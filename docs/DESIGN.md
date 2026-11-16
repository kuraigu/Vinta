# Vinta Design System Guidelines

A quick reference for styling, theming, layout, and component reuse across Vinta.

## Quick Index

- [1. Component-First Architecture](#1-component-first-architecture) (component discovery, reuse, creating new components)
- [2. Theming & Tokens](#2-theming--tokens) (CSS variables, cascade layers, theme manifests)
- [3. Scoping & `data-vinta`](#3-scoping--data-vinta) (page vs component scope, theme overrides)
- [4. Surfaces & Cards](#4-surfaces--cards) (Bento cards, opacity rules, desktop background bleed)
- [5. Typography, Icons & Illustrations](#5-typography-icons--illustrations) (fonts, Lucide icons, SVGs)
- [6. Interactive Controls & Forms](#6-interactive-controls--forms) (custom inputs, tooltips, standard heights)
- [7. Satellites & Motion](#7-satellites--motion) (floating docks, 60fps animations, hit-testing)
- [8. Code Organization & Stylesheets](#8-code-organization--stylesheets) (file locations, style scoping, no duplicate CSS)

---

## 1. Component-First Architecture

Always prioritize existing shared components before writing custom HTML or styles. Reusing existing building blocks maintains visual consistency, supports accessibility, and ensures full theme compatibility out of the box.

### Four-Tier Frontend Hierarchy

Vinta splits its Svelte UI into four clear layers so workspace state, overlays, and reusable controls stay decoupled:

- **Primitives (`frontend/src/components/`):** Reusable, token-driven building blocks (`BentoCard`, `CustomSelect`, `OverlayPage`, `ScrollView`, `SearchInput`). Primitives stay free of page-specific layout rules.
- **Widgets (`frontend/src/components/widgets/`):** Self-contained glance cards for the Home feed (`WeatherWidget`, `MediaGlanceWidget`, `CalendarGlanceWidget`, `SystemStatsWidget`, `ActionCenterWidget`, `AnnouncementWidget`, `NotificationWidget`). Widgets wrap `<WidgetFrame>` to carry `data-vinta-widget="<id>"` theme hooks, `contain: layout style` isolation, and Bento size presets (`1x1`, `2x1`, `2x2`).
- **Workspace Pages (`frontend/src/pages/`):** Persistent base views (`HomePage`, `MediaPage`, `NotesPage`, `CalculatorPage`, `CalendarPage`, `SettingsPage`) that own the island frame and keep user context.
- **Overlay Pages (`frontend/src/pages/overlays/`):** Top-level views (`NotificationsOverlayPage`, `AppsOverlayPage`) mounted inside `<OverlayPage>` above the active workspace. Closed overlays leave zero DOM nodes or listeners behind, and open overlays use solid elevated surfaces (`--vinta-bg-opaque` / `--vinta-surface-elevated`) instead of stacking a second `backdrop-filter: blur()` pass. Re-clicking the active dock trigger closes the overlay, and compact workspaces expand to at least `440px × 380px` while open before restoring exact dimensions on close or `Escape`.

### Where to Look

Check `frontend/src/components/` first:

- **Cards and Containers:** Look for `BentoCard.svelte`, `BentoGrid.svelte`, and `BentoItem.svelte`.
- **Navigation and Actions:** Look for `BackButton.svelte`, `DockButton.svelte`, and `RefreshButton.svelte`.
- **Form Controls:** Look for `CustomSelect.svelte`, `CustomDatePicker.svelte`, `CustomTimePicker.svelte`, and `SearchInput.svelte`.
- **Settings Primitives:** Check `frontend/src/components/settings/` for items, action buttons, sliders, and banners.
- **Drawers, Overlays, and Modals:** Look for `OverlayPage.svelte`, `DrawerPanel.svelte`, `EditorModal.svelte`, and `CustomContextMenu.svelte`.
- **Scrolling and Layout:** Look for `ScrollView.svelte` and `SidebarLayout.svelte`.
- **Accents and Badges:** Check `frontend/src/components/patterns/` and components like `ActionCapsule.svelte` or `DotIndicator.svelte`.

### Adding New Components

Only create a new component when existing primitives cannot satisfy the requirements.

- **Variant Over Duplication:** If two designs share structure or purpose, add a variant prop to the existing component instead of writing a new one.
- **Strict Tokenization:** Every new component must rely on `--vinta-*` CSS variables for background colors, borders, text, and active states.
- **No Inline Styles for Themes:** Never hardcode hex, rgb, or hsl values for core UI surfaces.

---

## 2. Theming & Tokens

Vinta uses a dynamic theme engine built on CSS Cascade Layers (`vinta.base`, `vinta.fallback`, `vinta.components`, `vinta.theme`) and `.vintatheme` packages.

### Token Catalog & Declarations

- **Source of Truth:** All base tokens are declared in `frontend/src/themes/tokens.css` with magenta (`#ff00ff`) missing-token sentinels. Check this file directly for the complete list of available `--vinta-*` variables.
- **Default Values:** Runtime default values are defined in the fallback preset stylesheet at `frontend/src/themes/presets/dark-gold/style.css`.
- **Adding Tokens:** Only edit `tokens.css` when introducing a brand new token to the system. Once declared with `#ff00ff`, provide concrete values inside the preset stylesheets under `frontend/src/themes/presets/*/style.css`.
- **Do Not Edit Sentinels for Theme Tweaks:** Never change `#ff00ff` in `tokens.css` to adjust colors for a theme. Adjust the preset stylesheets instead.

### Cascade Layer Hierarchy

1. `@layer vinta.base`: Foundational token declarations (`tokens.css`) using `#ff00ff` sentinels.
2. `@layer vinta.fallback`: Production fallback preset (`dark-gold/style.css`), ensuring missing tokens resolve cleanly.
3. `@layer vinta.components`: Scoped component styles compiled via Svelte `<style src="...">`.
4. `@layer vinta.theme`: Active runtime theme overrides injected into `<style id="vinta-active-theme">`.

### Theme Manifest (`manifest.json`)

Every theme directory or `.vintatheme` bundle requires a root `manifest.json` following `schemas/vintatheme.schema.json`:

- `schemaVersion` (integer): Version of the manifest schema (defaults to `1`).
- `id` (string): Unique identifier (lowercase alphanumeric, hyphens, underscores).
- `name` & `author` (string): Display name and creator handle.
- `version` (string): Semantic version (e.g., `1.0.0`).
- `entry` (string): Path to the stylesheet (defaults to `style.css`).
- `accentColor` (string, optional): Hex color for theme picker indicators.
- `mode` (string): Classification (`dark`, `light`, `oled`, `custom`).

Theme media assets belong in `frontend/src/assets/themes/` or point to local user paths. Never commit copyrighted media to the repository.

---

## 3. Scoping & `data-vinta`

Vinta scopes theme overrides cleanly without polluting every individual element.

### Page Containers Manage `data-vinta`

- Every top-level page in `frontend/src/pages/` must declare a root `data-vinta="<page-id>"` attribute (for example: `data-vinta="home-page"`, `data-vinta="settings-page"`, `data-vinta="calendar-page"`).
- Major top-level widget roots (`main-island`, `media-widget`, `notification`) also carry a `data-vinta` attribute.

### Components and Widgets

- `data-vinta` is optional for components and widgets in `frontend/src/components/`. Standalone widgets (such as `media-widget`, `calendar`, `notification`) or components requiring dedicated theme targeting may declare a root `data-vinta="<id>"` attribute.
- General child components remain token-driven without `data-vinta`, relying on semantic CSS class names and inherited `--vinta-*` variables.
- When a theme needs to customize how a component looks on a specific page, target it through the page container:
  ```css
  [data-vinta="calendar-page"] .tray-add-btn {
  	background: var(--vinta-accent-primary);
  }
  ```

---

## 4. Surfaces & Cards

### Bento Cards

Do not write custom card background formulas or box shadows from scratch. Use `BentoCard.svelte`:

- **Dashboard Widgets:** Use `<BentoCard variant="bento">`. It provides standard card radii, translucent surface fill, contrast borders, and hover elevation.
- **Settings Rows:** Use `<BentoCard variant="setting">` for horizontal key-value controls.
- **Announcement Banners:** Use `<BentoCard variant="announcement" color="blue|amber|emerald|danger|primary">` for status cards, weather warnings, or update alerts. Built-in gradients keep text readable while respecting theme boundaries.
- **Grid Layouts:** Wrap collections in `<BentoGrid>` with `<BentoItem>` to handle responsive columns and gaps automatically.

### Translucency & The Opaque Desktop Rule

- **Island Interiors:** Use `backdrop-filter: blur(16px)` for containers inside the main island.
- **Desktop Satellites Must Be Solid:** Windows WebView2 cannot blur native desktop windows located behind a transparent webview. Semi-transparent floating satellites will bleed unblurred wallpaper through. Keep `--vinta-dock-popover-bg` and peripheral satellite backgrounds 100% solid and opaque.

---

## 5. Typography, Icons & Illustrations

### Typography

- Use system sans-serif font stacks declared via `var(--vinta-font-family)` with explicit weights (`500`, `600`, `700`, `800`).
- Guard against dynamic text overflowing layouts by applying `text-overflow: ellipsis`, `overflow: hidden`, and `white-space: nowrap`.
- For truncated non-interactive text spans, supply a standard `title={text}` attribute for accessibility.

### Iconography

- **Primary Library:** Use Lucide icons from `lucide-svelte`.
- **Custom SVGs:** Place custom SVG icons in `frontend/src/assets/`. Use `fill="currentColor"` or `stroke="currentColor"` so icons inherit active theme colors.
- System and application icons for media sessions and notifications are resolved dynamically via `frontend/src/services/iconService.ts`.

### Illustrations

- Place non-interactive vectors in `frontend/src/illustrations/<category>/`.
- Build illustrations as responsive SVGs with clean `viewBox` coordinates. Use layered shapes, soft shadows, and gradients that complement dark surfaces.
- For illustrations that appear in both expanded cards and minimal island views, support a `variant` prop for `hero` and `compact`.

---

## 6. Interactive Controls & Forms

### Tooltips

Never use native HTML `title` attributes on interactive buttons or icon triggers. Native tooltips lag and ignore theme styling.

- Use the `use:islandTooltip` action from `frontend/src/services/islandTooltip.ts`.
- Alternatively, use `<CustomTooltip>` from `frontend/src/components/CustomTooltip.svelte`.
- **No Redundant Tooltips:** Never attach a tooltip to a button, tab, or control that already displays a visible text label (such as `"Start"`, `"Pause"`, `"Add Event"`, `"Focus"`, or `"Short Break"`) or an unambiguous symbol (`0` through `9`, `+`, `-`, `=`).
- **Icon-Only & Truncated Content Only:** Reserve tooltips strictly for icon-only actions or truncated text where extra context is genuinely helpful.
- **Keep Tooltips Concise:** Tooltip strings must be brief action labels (1 to 3 words, such as `"Clear all"`, `"Do Not Disturb"`, or `"Back"`), never full sentences.

### Custom Form Controls

Native HTML select menus and date pickers break window transparency and clash with theme tokens. Never use native form popups.

- Use `CustomSelect.svelte` instead of `<select>`.
- Use `CustomDatePicker.svelte` instead of `<input type="date">`.
- Use `CustomTimePicker.svelte` instead of `<input type="time">`.
- Use `CustomContextMenu.svelte` for right-click and overflow menus instead of browser context menus.
- Use `SearchInput.svelte` for search bars.
- Use the shared `use:portal` action and bounds helpers (`getUiScale`, `getIslandOrViewportBounds`, `clampToBounds`) from `frontend/src/utils/portal.ts` for tooltips, selects, and menus that must escape clipped island containers while respecting `--vinta-ui-scale`.
- Standard interactive controls standardize on a 32px height for consistent click targets.

---

## 7. Satellites & Motion

### Dynamic Island Motion

- Animate height and geometry smoothly using GPU-accelerated properties (`transform`, `opacity`) with cubic-bezier easing (`0.22, 1, 0.36, 1`).
- Multiply CSS and Svelte transition durations by `--vinta-anim-scale` (`animScaleStore`) so morph, overlay, and tab transitions follow the user's animation speed setting cleanly down to instant `0` transitions.
- Avoid rapid or distracting animations. Keep indicators and glimmers calm.

### Floating Satellites

Satellites are peripheral toolbars and docks orbiting the main island.

- **Hierarchy:** Keep satellites visually secondary with a 10px to 12px gap from the island.
- **Surface:** Use `var(--vinta-dock-popover-bg)` with no border and no drop shadow.
- **Multi-Rect Hit-Testing:** Wrap floating UI in `<SatelliteHost>` to sync hit-test geometry (`SetHitRects`) with the Go backend, allowing desktop clicks to pass through transparent gaps.
- **Screen Margins:** Satellites must never clip off-screen. Island positioning logic automatically clamps movement to keep satellites visible within monitor bounds.

---

## 8. Code Organization & Stylesheets

### Svelte Stylesheet Linking

- **Use `<style src="...">`:** Always link external component stylesheets using `<style src="./styles/<Name>.css"></style>`. Never use JavaScript `import './styles.css'` inside `<script>`, because script imports leak rules globally.
- **File Placement:**
  - Page stylesheets live in `frontend/src/pages/styles/<PageName>.css`.
  - Component stylesheets live in `frontend/src/components/styles/<ComponentName>.css` (mirroring subdirectories like `settings/SettingItem.css`).

### Stylesheet Hygiene

- Do not add repetitive CSS across multiple components. If styles are identical, consolidate them into a shared component or a variant.
- Keep styling scoped to classes within the component. Avoid universal or element-wide resets.
