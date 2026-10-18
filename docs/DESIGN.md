# Vinta Design System Guidelines

This document defines the visual guidelines, visual aesthetics, theme compliance rules, and icon standards for building components and pages in Vinta.

All UI contributions must follow these guidelines to maintain a clean and responsive desktop experience.

## Core Design Principles

### 1. Theme CSS Variables

Prefer using standard `--vinta-*` theme CSS variables for core UI surfaces, typography, and interactive elements so themes adapt smoothly across presets. Component-specific accent colors or custom values are permitted when unique visual requirements call for them, but referencing standard tokens is strongly encouraged for visual cohesion across themes.

Standard token groups:
- **Accents:** `--vinta-accent-primary`, `--vinta-accent-hover`, `--vinta-accent-active`, `--vinta-accent-muted`, `--vinta-accent-purple`, `--vinta-accent-danger`, `--vinta-accent-warning`, `--vinta-accent-success`
- **Surfaces:** `--vinta-bg`, `--vinta-bg-opaque`, `--vinta-surface-card`, `--vinta-surface-elevated`, `--vinta-dock-popover-bg` (must always remain an opaque hex/rgb color to prevent WebView2 transparent desktop bleed-through)
- **Borders:** `--vinta-border`, `--vinta-border-light`, `--vinta-border-accent`
- **Interactions:** `--vinta-hover`, `--vinta-item-bg`
- **Typography:** `--vinta-text-primary`, `--vinta-text-secondary`, `--vinta-text-muted`
- **Geometry:** `--vinta-radius-island`, `--vinta-radius-card`, `--vinta-radius-button`, `--vinta-radius-pill`, `--vinta-font-family`

### 2. Iconography Standardsp

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
- **Custom Tooltip Mandate for Interactive Controls:** Always use Vinta's custom tooltip mechanism via the `use:islandTooltip` action (`frontend/src/services/islandTooltip.ts`) or the `<CustomTooltip>` component (`frontend/src/components/CustomTooltip.svelte`) on interactive buttons, icon triggers, and toggles. Never use native HTML `title` attributes on interactive elements, as native OS tooltips introduce delayed hover popups and fail to integrate with Vinta design tokens and hit testing.
- **Truncation Fallback:** Retain native `title={text}` attributes strictly on truncated, non-interactive text spans to provide a simple accessible fallback without cluttering action tooltips.

### 8. Non-Distracting Indicator States

- Avoid overly fast or annoying blinking animations. Keep indicator states clean and steady to maintain a calm user experience.

### 9. Custom Form Control Standards (Select, DatePicker, TimePicker)

- **Never use native HTML `<select>` elements:** Standard `<select>` elements create unstyled operating system dialogs that lack theme styling and break within scrollable containers. Always import and use `CustomSelect` from `frontend/src/components/CustomSelect.svelte`.
- **Swipe Protection in Select:** `<CustomSelect>` includes `.no-swipe` and pointer event suppression to ensure opening and scrolling options inside carousel viewports (`ComponentGallery`) does not trigger unwanted slide navigation.
- **Never use native HTML `<input type="date">` elements:** Native date inputs spawn operating system dialogs that conflict with window transparency and lack theme variables. Always import and use `CustomDatePicker` from `frontend/src/components/CustomDatePicker.svelte`.
- **Never use native HTML `<input type="time">` elements:** Native time inputs disrupt keyboard navigation and do not support dynamic theme accents. Always import and use `CustomTimePicker` from `frontend/src/components/CustomTimePicker.svelte`.

### 10. Component & Page Stylesheet Splitting

- **Scoped Styling via `<style src="...">`**: When a page or component grows in length and styling complexity, extract its `<style>` block into a dedicated `.css` file.
- **File Naming & Directory Structure**:
  - Keep stylesheet filenames identical to component filenames (`<Name>.svelte` ➔ `<Name>.css`).
  - **Pages**: Place extracted stylesheets in `frontend/src/pages/styles/<PageName>.css` (linked via `<style src="./styles/<PageName>.css"></style>`).
  - **Components & Sub-groups**: Place extracted stylesheets under `frontend/src/components/styles/` mirroring the folder structure (e.g. `src/components/notes/NoteLive.svelte` ➔ `src/components/styles/notes/NoteLive.css` linked via `<style src="../styles/notes/NoteLive.css"></style>`).
- **No Global Script Imports**: Never use `import './styles.css'` inside `<script>` for component styling, as it bypasses scoping and leaks styles globally. Always use `<style src="...">` to retain Svelte's compile-time scoping hashes.

### 11. Floating Satellites and Companion Docks

Satellites are secondary peripheral surfaces that orbit the perimeter of the Dynamic Island. They provide quick-access controls, status indicators, and companion toolbars while keeping the primary island canvas focused and uncluttered.

#### 1. Purpose and Visual Hierarchy

- **Secondary Role:** The Dynamic Island is always the primary visual anchor. Satellites must look and feel secondary in weight, scale, and density.
- **Visual Separation:** Satellites must remain optically detached from the main island, separated by an intentional air gap (typically `10px` to `12px`), rather than touching or overlapping the main capsule.
- **Compact Scale:** Interactive targets inside satellites should be scaled down relative to primary island buttons (typically `26px` to `32px` touch targets with `14px` to `18px` iconography) to maintain clear visual hierarchy.

#### 2. Form Factor and Surface Aesthetics

- **Capsule Curvature:** Satellites must share the island's rounded geometry. Use pill or capsule silhouettes with border radii between `16px` and `20px` (or full circular radius for small single-action toggles).
- **Surface Tokens:** Satellite containers use a clean, borderless surface without elevations:
  - Background: `var(--vinta-dock-popover-bg)`
  - Border: `none` (satellites omit outer borders and highlights to maintain visual distinction from the primary island)
  - Elevation: `none` (shadows are omitted on secondary satellites)
  - Desktop Opacity: Satellites floating over the transparent desktop canvas must use solid opaque surfaces. CSS backdrop blur cannot sample or blur native OS windows beneath transparent webviews, so semi-transparency causes unblurred desktop bleed-through.

#### 3. Idle vs. Active States

- **Subtle Idle Presence:** When persistent or idle alongside the island, collapsed toggles and background tools maintain a quiet visual presence. Subtlety is achieved through subdued iconography (`var(--vinta-text-secondary)` or `var(--vinta-text-muted)`) rather than container transparency against the desktop.
- **Hover Transitions:** On cursor proximity, hover, or focus-within, satellites transition smoothly (`0.2s` cubic-bezier) to full primary iconography (`var(--vinta-text-primary)`) and elevated button states without container highlights or drop shadows.

#### 4. Spatial Anchoring and Alignment

- **Cardinal Anchors:** Satellites anchor along the cardinal flanks of the island (`left`, `right`, `top`, `bottom`).
- **Flank Alignment:**
  - When paired with a minimal island pill, flank satellites must be vertically centered relative to the minimal island height to preserve balance and symmetry.
  - When paired with an expanded card, flank satellites should align with the top edge or upper interactive zone of the card.
- **Polar Alignment:** Satellites anchored to the `top` or `bottom` must be horizontally centered relative to the island width.
- **Flank Discipline:** Avoid populating all four flanks simultaneously. Group companion controls onto a single flank to avoid cluttering the desktop workspace.

#### 5. Drawer Morphing and State Transformations

- **Compact Toggles to Expanded Drawers:** When a satellite provides multiple actions in a compact footprint, design it with a folded toggle state that expands on interaction:
  - **Folded State:** A single compact circular or pill toggle (such as a directional glyph or status indicator) that minimizes desktop footprint.
  - **Unfolded State:** A horizontal or vertical drawer revealing individual action buttons with clear spacing (`4px` to `6px`).
- **Smooth Container Morphing:** The container width, height, and padding must animate together with a smooth easing curve (`0.22s cubic-bezier(0.22, 1, 0.36, 1)`). Child buttons should transition smoothly without jumping.
- **Clear Retraction:** Every expandable drawer must provide an intuitive retract or fold trigger so the user can restore the compact idle footprint at any time.

#### 6. Screen Edge Proximity and Margin Safety

- **Display Perimeter Clearance:** Satellites must never clip off the screen edge or intersect the display boundary.
- **Dynamic Margin Clamping:** The island's allowed movement range must account for the full bounding footprint of all active satellites plus a safety margin (`12px`).
- **Responsive Edge Evasion:** If a satellite expands horizontally while near a screen edge, the island must smoothly slide inward to maintain full visibility and breathing room for the satellite.

#### 7. Tooltips and Popovers on Satellites

- **Directional Awareness:** Tooltips originating from satellites must orient away from the main island or adapt dynamically (`top`, `bottom`, `left`, `right`) depending on proximity to monitor edges.
- **Clear Placement:** Satellite tooltips float outside the satellite container while remaining fully visible within the active monitor workspace.

#### 8. Accessibility and Hit Targets

- **Accessible Roles:** Satellite toolbars must include appropriate semantic attributes (`role="toolbar"` and a descriptive `aria-label`).
- **Focus Rings:** All satellite buttons must display standard keyboard focus indicators (`:focus-visible` with `2px` accent outline and `1px` outline offset).
- **Target Spacing:** Keep at least `4px` of spacing between interactive items to prevent misclicks.

### 12. Bento Cards and Settings Standardization

#### 1. Bento Card Architecture and Surface Depth
- **Curvature Radii:** Primary dashboard and feed widgets use a generous `22px` squircle border radius. Settings items, action center tiles, and secondary cards use `16px` to `20px`.
- **Luminous Mesh Surfaces:** Bento cards avoid flat opaque grey backgrounds. Instead, cards use multi-stop dark surface mesh gradients (`linear-gradient(145deg, rgba(255, 255, 255, 0.065) 0%, rgba(255, 255, 255, 0.02) 100%), #15161c`).
- **Specular Rim Lighting:** Every bento card incorporates a subtle top specular rim highlight (`box-shadow: inset 0 1px 0 0 rgba(255, 255, 255, 0.12), 0 4px 18px rgba(0, 0, 0, 0.18)`), creating tactile elevation.
- **Micro-Interaction Hover Lift:** On hover, interactive cards transition with GPU-accelerated lift (`transform: translateY(-1px)`) and deepened ambient shadows (`0 6px 22px rgba(0, 0, 0, 0.24)`).

#### 2. Control Layout Standardization in Settings
- **Full-Width Action Placement:** Action buttons inside settings cards sit directly beneath the title and description using `flexCol={true}` and `fullWidth={true}`. This guarantees uniform interactive hit areas regardless of button label length.
- **Unified 32px Control Height:** Interactive action buttons (`SettingActionButton`) standardize on a `32px` height, matching `CustomSelect` dropdown triggers and text input fields.
- **Grouped Multi-Action Bentos:** When multiple actions relate to the same topic (such as documentation links or repository resources), group them under a single bento item with split or flex controls rather than creating fragmented individual cards.

#### 3. Contextual Atmospheric Gradient Banners
- **Banner Architecture:** Dynamic informational cards (such as the Weather widget, update available notices, and build warnings) feature rich, multi-stop atmospheric gradients.
- **Deep Luminance for Contrast:** Gradients must use deep, subdued color stops rather than neon or overly saturated hues. This preserves crisp text legibility in dark mode and eliminates glaring contrast fatigue.
- **Frosted Glass Accents:** Internal chips, badges, and action buttons within gradient banners use translucent frosted glass styling (`background: rgba(255, 255, 255, 0.12); border: 1px solid rgba(255, 255, 255, 0.2)`) with white typography to maintain contrast and visual cohesion.
- **Left-Aligned Content Flow:** Informational text, badges, and status descriptions inside atmospheric banners remain strictly left-aligned alongside their action buttons to sustain clean vertical reading lines.

### 13. Illustrations Architecture and Standards

#### 1. Directory Structure and Separation of Concerns
- **Location:** All thematic, decorative, and state illustrations reside in `frontend/src/illustrations/<category>/` (e.g. `frontend/src/illustrations/weather/`).
- **Distinction from Components:** Components in `frontend/src/components/` handle application logic, navigation, interactive controls, and user actions. Illustrations in `frontend/src/illustrations/` provide rich, non-interactive visual representations of data states (such as weather conditions, empty states, or achievement artwork).

#### 2. Vector Graphics and Scalability Standards
- **Vector Base:** Illustrations must be constructed using responsive SVG markup with clean `viewBox` coordinates to scale seamlessly across desktop window dimensions without raster artifacts.
- **Layered Volumetric Depth:** Avoid flat 2D silhouettes. Use multi-stop radial and linear gradients, layered overlapping shapes, soft ambient drop shadows, and specular highlights to create tangible visual depth.
- **Responsive Variants:** When an illustration is used in both hero cards and compact capsule views (such as the minimized Dynamic Island), provide a `variant` prop supporting `hero` (full volumetric artwork) and `compact` (clean, streamlined vector representation).
- **Theme and Ambient Harmony:** Illustration color stops must complement surrounding surface gradients, avoiding harsh neon tones that cause visual fatigue.

### 14. Theme Engine Architecture and Manifest Specification

Vinta features a dynamic theme engine built on CSS Cascade Layers (`vinta.base`, `vinta.components`, `vinta.theme`) and `.vintatheme` packages.

#### 1. Cascade Layer Hierarchy
1. `@layer vinta.base`: Core foundational tokens defined in `frontend/src/themes/oled-dark.css`.
2. `@layer vinta.components`: Hash-scoped component styles compiled via Svelte `<style src="...">`.
3. `@layer vinta.theme`: Active theme overrides injected dynamically into `<style id="vinta-active-theme">`. Properties declared in this layer take precedence over default base and component fallbacks.

#### 2. Manifest Schema (`manifest.json`)
Every theme directory or `.vintatheme` zip package must include a root `manifest.json` adhering to `schemas/vintatheme.schema.json`:
- `schemaVersion` (integer): Version of the manifest schema. Default is 1.
- `id` (string): Unique identifier (lowercase alphanumeric, hyphens, underscores).
- `name` (string): Human readable theme title.
- `author` (string): Author name or handle.
- `version` (string): Semantic version string (e.g. `1.0.0`).
- `entry` (string): Relative path to entry stylesheet (default `style.css`).
- `preview` (string, optional): Path to preview thumbnail.
- `accentColor` (string, optional): Hex color for theme picker badges.
- `mode` (string): Color scheme classification (`dark`, `light`, `oled`, `custom`).

#### 3. Schema Versioning Policy
The `schemaVersion` integer is incremented only when breaking token removals or structural manifest shifts occur. Additive changes (such as introducing new non-breaking optional CSS tokens) do not warrant a schema version bump because unspecified tokens cleanly fall back to `vinta.base` defaults. Outdated schema versions trigger fallback to `oled-dark` to prevent interface malfunction.

#### 4. Opaque Surface Rule for Desktop Overlays
The `--vinta-dock-popover-bg` token must always resolve to a solid, fully opaque color. Because Windows WebView2 cannot sample or blur native windows through transparent webview regions, translucent alpha channels on floating popovers cause unblurred desktop bleed-through.
