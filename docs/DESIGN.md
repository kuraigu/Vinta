# Vinta Design System Guidelines

Welcome to the Vinta design guide. This document covers how we style components, manage themes, and keep the desktop experience feeling fast, cohesive, and native.

If you are contributing to Vinta's UI, please follow these rules so everything fits together naturally.

---

## 1. Theming & Tokens

Vinta uses a dynamic theme engine built on CSS Cascade Layers (`vinta.base`, `vinta.components`, `vinta.theme`) and `.vintatheme` packages.

### CSS Variables & Where to Edit

Always prefer `--vinta-*` CSS variables for core UI surfaces, text, and interactive states so components adapt automatically across themes. You can use custom one-off colors if a specific component genuinely needs it, but standard tokens should be your default.

- **Common examples:** `--vinta-accent-primary`, `--vinta-accent-hover`, `--vinta-accent-active`, `--vinta-accent-muted`, `--vinta-accent-purple`, `--vinta-accent-danger`, `--vinta-accent-warning`, `--vinta-accent-success`.
- **`tokens.css` is for definitions only:** All base tokens live in `frontend/src/themes/tokens.css`. **Do not** edit `tokens.css` to tweak an existing theme's look. Only touch this file when introducing a brand-new token to the system.
- **Editing Built-in & Accessibility Presets:** Vinta ships with built-in themes (including dedicated accessibility presets calibrated with Okabe-Ito color scales). To adjust the colors or styling of any built-in preset, edit that preset's individual `style.css` file inside its respective theme directory rather than `tokens.css`.

### Cascade Layer Hierarchy

1. `@layer vinta.base`: Core foundational token defaults (`frontend/src/themes/tokens.css`).
2. `@layer vinta.components`: Scoped component styles compiled via Svelte `<style src="...">`.
3. `@layer vinta.theme`: Active theme overrides injected at runtime into `<style id="vinta-active-theme">`. Anything declared here overrides base and component fallbacks.

### Custom Themes & Copyright Rules

Vinta supports community themes, including custom backgrounds.

- **Asset Hosting:** Theme assets belong in `frontend/src/assets/themes/`, or they can point to external URLs and local user file paths.
- **No Copyrighted Media in Core:** Never commit copyrighted artwork, anime images, or trademarked media to the repository. Official repo themes must use CC0/public domain media, or let users supply their own local files. If possible, use SVG.

### Theme Manifest (`manifest.json`)

Every theme directory or `.vintatheme` package requires a root `manifest.json` that follows `schemas/vintatheme.schema.json`:

- `schemaVersion` (integer): Manifest schema version (defaults to `1`). Only bump this for breaking token removals or structural manifest changes. Adding new optional CSS tokens does not require a version bump, since missing tokens safely fall back to `vinta.base`. Outdated schemas automatically fall back to the default built-in theme.
- `id` (string): Unique ID (lowercase alphanumeric, hyphens, underscores).
- `name` & `author` (string): Creator name/handle.
- `version` (string): Semantic version (e.g., `1.0.0`).
- `entry` (string): Relative path to the stylesheet (defaults to `style.css`).
- `preview` (string, optional): Path to the preview thumbnail.
- `accentColor` (string, optional): Hex color for theme picker badges.
- `mode` (string): Classification (`dark`, `light`, `oled`, `custom`).

---

## 2. Visual Aesthetics & Surfaces

### Glassmorphism, Backdrops & Opaque Overlays

- **Island Translucency:** Use `backdrop-filter: blur(16px)` for floating island containers, paired with radial glow backdrops (`glow-backdrop`) using translucent accent gradients for depth.
- **The Opaque Desktop Rule:** While internal island elements can use blur, `--vinta-dock-popover-bg` and floating desktop satellites **must be 100% solid and opaque**. Windows WebView2 cannot blur native OS windows sitting behind a transparent webview, so semi-transparent floating popovers will just bleed raw, unblurred desktop wallpaper through.

### Bento Cards

- **Border Radii:** Primary dashboard and feed widgets use a `22px` squircle radius. Settings items, action center tiles, and secondary floating containers use `12px` to `20px`.
- **Luminous Mesh Surfaces:** Avoid flat grey boxes. Bento cards use multi-stop dark surface gradients (`linear-gradient(145deg, rgba(255, 255, 255, 0.065) 0%, rgba(255, 255, 255, 0.02) 100%), #15161c`).
- **Specular Rim Lighting:** Add a subtle top inner highlight for tactile depth (`box-shadow: inset 0 1px 0 0 rgba(255, 255, 255, 0.12), 0 4px 18px rgba(0, 0, 0, 0.18)`).
- **Hover Lift:** Interactive cards lift slightly on hover (`transform: translateY(-1px)`) with a deeper shadow (`0 6px 22px rgba(0, 0, 0, 0.24)`).

### Atmospheric Gradient Banners

Dynamic info cards (like Weather, update notices, and build warnings) use rich, multi-stop gradients:

- **Keep Luminance Deep:** Stick to subdued, deep color stops rather than neon or hyper-saturated hues so white text stays crisp and easy to read in dark mode.
- **Frosted Accents:** Inner chips, badges, and buttons inside banners use frosted glass (`background: rgba(255, 255, 255, 0.12); border: 1px solid rgba(255, 255, 255, 0.2)`) with white text.
- **Left-Aligned Flow:** Keep text, badges, and descriptions strictly left-aligned alongside their action buttons for clean vertical scanning.

---

## 3. Typography, Icons & Illustrations

### Typography & Text Truncation

- Use clean sans-serif system font stacks with explicit weights (`500`, `600`, `700`, `800`).
- Prevent dynamic text (like city names or timezones) from breaking layouts by applying `text-overflow: ellipsis`, `overflow: hidden`, `white-space: nowrap`, and a `max-width`.
- On truncated, non-interactive text spans, include a native `title={text}` attribute as a lightweight accessibility fallback.

### Iconography

- **Primary Library:** Always reach for **Lucide Icons** (`lucide-svelte`) first.
- **Custom SVGs & App Icons:** If Lucide doesn't have what you need, put custom SVG helpers in `frontend/src/assets/`. Always use `fill="currentColor"` or `stroke="currentColor"` so they inherit theme colors. (Note: OS app icons for notifications and media are pulled automatically at runtime via `frontend/src/services/iconService.ts`).

### Illustrations

- **Where They Live:** Non-interactive state artwork (weather conditions, empty states, achievements) lives in `frontend/src/illustrations/<category>/`. Keep interactive logic in `frontend/src/components/`.
- **Volumetric Vectors:** Build illustrations as responsive SVGs with clean `viewBox` coordinates. Avoid flat 2D silhouettes; use layered shapes, multi-stop gradients, soft shadows, and highlights that complement our dark surface palette.
- **Size Variants:** If an illustration appears in both full cards and the minimized Dynamic Island, support a `variant` prop for `hero` (full detail) and `compact` (streamlined for small sizes).

---

## 4. Interactive Controls & Forms

### Tooltips

- **Never use native `title` on interactive elements:** Native OS tooltips lag on hover and ignore our theme tokens. Always use the `use:islandTooltip` action (`frontend/src/services/islandTooltip.ts`) or `<CustomTooltip>` (`frontend/src/components/CustomTooltip.svelte`) for buttons, icon triggers, and toggles.

### Custom Form Controls

Native HTML form popups clash with window transparency, ignore theme variables, and break keyboard/swipe navigation. **Never use them.**

- Instead of `<select>`, use `CustomSelect` (`frontend/src/components/CustomSelect.svelte`). It includes `.no-swipe` pointer protection so scrolling a dropdown inside a carousel (`ComponentGallery`) won't accidentally swipe the slide.
- Instead of `<input type="date">`, use `CustomDatePicker` (`frontend/src/components/CustomDatePicker.svelte`).
- Instead of `<input type="time">`, use `CustomTimePicker` (`frontend/src/components/CustomTimePicker.svelte`).

### Settings Card Layout

- **Full-Width Actions:** Place action buttons directly below the card's title and description using `flexCol={true}` and `fullWidth={true}` so hit targets stay consistent regardless of label length.
- **32px Standard Height:** Interactive controls (`SettingActionButton`, `CustomSelect` triggers, and text inputs) all standardize on a `32px` height.
- **Group Related Actions:** If multiple actions belong to the same topic (like docs links or repo resources), group them into a single bento card with split/flex controls instead of making separate cards for each button.

---

## 5. Layout, Motion & Floating Satellites

### Dynamic Bounds & Micro-Animations

- **No Hardcoded Offsets:** Calculate element bounds dynamically from wrapped content. When the island expands for a new widget, animate the container height smoothly.
- **60fps Motion:** Stick to GPU-accelerated properties (`transform`, `opacity`) using smooth cubic-bezier easing (`transition: transform 0.4s cubic-bezier(0.22, 1, 0.36, 1)`).
- **Swipe Gestures:** Use the `use:swipe` action (`frontend/src/utils/useSwipe.ts`) for touch and mouse drag interactions.
- **Calm Indicators:** Avoid rapid or distracting blinking animations. Keep status indicators steady and calm.

### Floating Satellites & Companion Docks

Satellites are small peripheral toolbars and toggles that orbit the Dynamic Island, giving quick access to controls without cluttering the main canvas.

1. **Hierarchy & Scale:** The Dynamic Island is the star. Keep satellites visually secondary with a `10px` to `12px` air gap from the island. Scale down satellite targets (`26px` to `32px` buttons with `14px` to `18px` icons).
2. **Surface & Shape:** Match the island's pill curvature (`16px` to `20px` radius, or circular for single toggles). Use `var(--vinta-dock-popover-bg)` with **no border** and **no drop shadow**.
3. **Idle vs. Hover:** Keep idle satellites quiet by dimming the icons (`var(--vinta-text-secondary)` or `var(--vinta-text-muted)`), not by making the background transparent. On hover or focus-within, transition smoothly (`0.2s`) to `--vinta-text-primary`.
4. **Anchoring:** Satellites sit on the `left`, `right`, `top`, or `bottom` flanks.
   - _Left/Right flanks:_ Center vertically next to a minimal pill, or align with the top interactive zone of an expanded card.
   - _Top/Bottom flanks:_ Center horizontally relative to the island.
   - _Don't crowd the island:_ Group companion controls onto one flank rather than surrounding all four sides at once.
5. **Expandable Drawers:** Multi-action satellites should start as a compact toggle and unfold into a spaced drawer (`4px` to `6px` gap) using a `0.22s cubic-bezier(0.22, 1, 0.36, 1)` morph. Always include an obvious way to fold the drawer back up.
6. **Screen Edge Safety:** Satellites must never clip off-screen. Clamp the island's movement range to include active satellites plus a `12px` safety margin. If a satellite expands near a monitor edge, slide the island inward to make room.
7. **Tooltips & Accessibility:** Satellite tooltips should point away from the main island and stay inside the monitor bounds. Use `role="toolbar"`, a clear `aria-label`, at least `4px` between buttons, and `:focus-visible` outlines (`2px` accent outline with `1px` offset).

---

## 6. Code Organization & Stylesheets

When a component or page's `<style>` block gets long, split it into a dedicated `.css` file.

- **Use `<style src="...">`:** Always link external stylesheets using Svelte's `<style src="..."></style>` tag. **Never** use `import './styles.css'` inside your `<script>` block, because script imports bypass Svelte's style scoping and leak rules globally.
- **Where to Put CSS Files:** Match the `.svelte` filename (`<Name>.svelte` ➔ `<Name>.css`):
  - **Pages:** `frontend/src/pages/styles/<PageName>.css` (linked via `<style src="./styles/<PageName>.css"></style>`).
  - **Components:** Mirror the component folder structure inside `frontend/src/components/styles/` (e.g., `src/components/notes/NoteLive.svelte` links to `../styles/notes/NoteLive.css`).
