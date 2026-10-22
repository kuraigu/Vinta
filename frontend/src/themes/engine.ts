import type { ThemeInfo } from './types.js'
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
    styleTag.textContent = `@layer vinta.theme {\n${darkGoldCss}\n}`
    return
  }

  if (normalizedId === 'nord-frost') {
    styleTag.textContent = `@layer vinta.theme {\n${nordFrostCss}\n}`
    return
  }

  if (normalizedId === 'colorblind-safe') {
    styleTag.textContent = `@layer vinta.theme {\n${colorblindSafeCss}\n}`
    return
  }

  try {
    const rawCss = await getThemeCss(normalizedId)
    if (rawCss) {
      styleTag.textContent = `@layer vinta.theme {\n${rawCss}\n}`
    } else {
      styleTag.textContent = ''
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
