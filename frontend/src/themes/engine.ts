import type { ThemeInfo } from './types.js'
import tokensRawCss from './tokens.css?raw'
import darkGoldManifest from './presets/dark-gold/manifest.json'
import darkGoldCss from './presets/dark-gold/style.css?raw'
import nordFrostManifest from './presets/nord-frost/manifest.json'
import nordFrostCss from './presets/nord-frost/style.css?raw'
import colorblindSafeManifest from './presets/colorblind-safe/manifest.json'
import colorblindSafeCss from './presets/colorblind-safe/style.css?raw'
import {
  ListThemes as listThemesFromBackend,
  GetThemeCSS as getThemeCssFromBackend,
  ImportThemeDialog,
  OpenThemesFolder
} from '../../bindings/Vinta/backend/themeservice.js'

const STYLE_TAG_ID = 'vinta-active-theme'

/**
 * Built-in themes packaged directly with the frontend distribution.
 */
export const BUILTIN_THEMES: ThemeInfo[] = [
  {
    manifest: darkGoldManifest as ThemeInfo['manifest'],
    isBuiltIn: true,
    path: 'frontend/src/themes/presets/dark-gold'
  },
  {
    manifest: nordFrostManifest as ThemeInfo['manifest'],
    isBuiltIn: true,
    path: 'frontend/src/themes/presets/nord-frost'
  },
  {
    manifest: colorblindSafeManifest as ThemeInfo['manifest'],
    isBuiltIn: true,
    path: 'frontend/src/themes/presets/colorblind-safe'
  }
]

/**
 * Retrieves built-in presets combined with custom themes installed on disk.
 */
export async function listThemes(): Promise<ThemeInfo[]> {
  try {
    const customThemes = await listThemesFromBackend()
    if (Array.isArray(customThemes) && customThemes.length > 0) {
      return [...BUILTIN_THEMES, ...(customThemes as unknown as ThemeInfo[])]
    }
  } catch (err) {
    console.warn('Failed to load custom themes from backend:', err)
  }

  return BUILTIN_THEMES
}

/**
 * Prompts the user with a native file dialog to import a .vintatheme package.
 */
export async function importTheme(): Promise<ThemeInfo | null> {
  return (await ImportThemeDialog()) as unknown as ThemeInfo | null
}

/**
 * Opens the local themes storage directory in Windows Explorer.
 */
export async function openThemesFolder(): Promise<void> {
  await OpenThemesFolder()
}

/**
 * Retrieves the entry CSS stylesheet content for a specific theme ID.
 */
export async function getThemeCss(id: string): Promise<string> {
  if (id === 'dark-gold') {
    return darkGoldCss
  }
  if (id === 'nord-frost') {
    return nordFrostCss
  }
  if (id === 'colorblind-safe') {
    return colorblindSafeCss
  }
  return await getThemeCssFromBackend(id)
}

/**
 * Extracts all declared `--vinta-*` token names from a raw CSS stylesheet.
 *
 * @param css - Raw CSS string to scan for custom property declarations.
 * @returns Unique set of declared `--vinta-*` token names.
 */
function extractDeclaredTokens(css: string): Set<string> {
  const matches = css.match(/--vinta-[a-z0-9-]+(?=\s*:)/g) ?? []
  return new Set(matches)
}

/**
 * Logs a development-only console warning if a theme stylesheet omits tokens declared in `tokens.css`.
 *
 * @param themeId - Active theme identifier.
 * @param themeCss - Raw CSS content of the active theme.
 */
function auditThemeTokensInDev(themeId: string, themeCss: string): void {
  if (!import.meta.env.DEV) {
    return
  }
  const declared = extractDeclaredTokens(tokensRawCss)
  const implemented = extractDeclaredTokens(themeCss)
  const missing = [...declared].filter((token) => !implemented.has(token))
  if (missing.length > 0) {
    console.warn(
      `[Vinta Theme Audit] Theme "${themeId}" is missing ${missing.length} token(s) declared in tokens.css (using dark-gold fallback):`,
      missing
    )
  }
}

/**
 * Applies a theme to the current DOM document by injecting CSS into the vinta.theme cascade layer.
 */
export async function applyThemeToDOM(themeId: string): Promise<void> {
  if (typeof document === 'undefined') {
    return
  }

  let styleTag = document.getElementById(
    STYLE_TAG_ID
  ) as HTMLStyleElement | null
  if (!styleTag) {
    styleTag = document.createElement('style')
    styleTag.id = STYLE_TAG_ID
    document.head.appendChild(styleTag)
  }

  const normalizedId = themeId || 'dark-gold'
  document.documentElement.dataset.theme = normalizedId

  if (normalizedId === 'dark-gold') {
    auditThemeTokensInDev(normalizedId, darkGoldCss)
    styleTag.textContent = `@layer vinta.theme {\n${darkGoldCss}\n}`
    return
  }

  if (normalizedId === 'nord-frost') {
    auditThemeTokensInDev(normalizedId, nordFrostCss)
    styleTag.textContent = `@layer vinta.theme {\n${nordFrostCss}\n}`
    return
  }

  if (normalizedId === 'colorblind-safe') {
    auditThemeTokensInDev(normalizedId, colorblindSafeCss)
    styleTag.textContent = `@layer vinta.theme {\n${colorblindSafeCss}\n}`
    return
  }

  try {
    const rawCss = await getThemeCss(normalizedId)
    if (rawCss) {
      auditThemeTokensInDev(normalizedId, rawCss)
      styleTag.textContent = `@layer vinta.theme {\n${rawCss}\n}`
    } else {
      styleTag.textContent = `@layer vinta.theme {\n${darkGoldCss}\n}`
      document.documentElement.dataset.theme = 'dark-gold'
    }
  } catch (err) {
    console.warn(
      `Failed to load theme "${normalizedId}", falling back to dark-gold:`,
      err
    )
    styleTag.textContent = `@layer vinta.theme {\n${darkGoldCss}\n}`
    document.documentElement.dataset.theme = 'dark-gold'
  }
}
