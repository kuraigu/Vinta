# Vinta Frontend

The frontend is a single-page desktop interface built with Svelte 5, TypeScript, and Vite. It communicates with the Go backend through Wails v3 generated bindings.

## Directory Structure

```
frontend/
├── src/
│   ├── App.svelte          # Root island orchestrator and drag gesture handler
│   ├── components/         # Reusable widgets and subcomponents
│   │   ├── settings/       # Settings tabs and bento cards
│   │   ├── timer/          # Stopwatch, countdown, focus, and alarm views
│   │   ├── notes/          # Notes editor, list drawer, and formatting bar
│   │   └── styles/         # Extracted stylesheets for complex components
│   ├── pages/              # Top-level island views (Home, Media, Apps, etc.)
│   │   └── styles/         # Extracted stylesheets for pages
│   ├── stores/             # Svelte reactive state stores
│   ├── services/           # Window bounds, gestures, tooltips, and weather services
│   ├── illustrations/      # Volumetric vector artwork for weather and widgets
│   ├── themes/             # Theme engine, presets (oled-dark, nord-frost, colorblind-safe), base styles
│   └── utils/              # Gesture math, date formatting, and application helpers
├── bindings/               # Generated Wails v3 Go to TypeScript bindings
└── index.html              # HTML shell configured for transparent webviews
```

## Development Guidelines

### 1. Svelte 5 Runes

All components use Svelte 5 runes (`$state`, `$derived`, `$props`, `$effect`):

- When declaring `$effect` runes, attach a dedicated named function to the rune for clarity and readability.
- Keep derived derivations concise and scoped to component needs.

### 2. Design System & CSS Variables

- Strict compliance with `docs/DESIGN.md` is mandatory.
- Use standard `--vinta-*` semantic CSS variables (such as `var(--vinta-accent-primary)` and `var(--vinta-item-bg)`).
- Layer hierarchy uses CSS Cascade Layers (`@layer vinta.base, vinta.components, vinta.theme;`).
- Use `lucide-svelte` for standard iconography.

### 3. Stylesheet Organization

- Scoped styling via `<style src="...">` is required when components or pages exceed concise inline lengths.
- Never use script-level CSS imports (`import './styles.css'`) for component styles, as this breaks Svelte compile-time scoping.

### 4. Custom Form Controls

Native browser `<select>`, `<input type="date">`, and `<input type="time">` elements must not be used. Use `CustomSelect`, `CustomDatePicker`, and `CustomTimePicker` to ensure visual harmony with themes and avoid operating system popups.
