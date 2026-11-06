/**
 * Vinta Reactive Internationalization (i18n) Engine.
 *
 * Provides reactive translation lookups, fallback resolution to English,
 * token interpolation, and dynamic drop-in locale loading from disk.
 */

import { writable, derived, get } from 'svelte/store'
import { localeStore } from '../stores/settings'
import enCatalog from '../locales/en.json'
import deCatalog from '../locales/de.json'
import filCatalog from '../locales/fil.json'
import {
  GetCustomLocales,
  OpenLocalesFolder
} from '../../bindings/Vinta/backend/app.js'

export interface LocaleOption {
  id: string
  name: string
  flag?: string
  isCustom?: boolean
}

/** Standard embedded locale metadata */
export const BUILTIN_LOCALES: LocaleOption[] = [
  { id: 'en', name: 'English (US)', flag: 'en' },
  { id: 'en-GB', name: 'English (UK)', flag: 'en-gb' },
  { id: 'fil', name: 'Filipino', flag: 'fil' },
  { id: 'de', name: 'Deutsch', flag: 'de' }
]

type TranslationCatalog = Record<string, any>

const embeddedCatalogs: Record<string, TranslationCatalog> = {
  en: enCatalog,
  'en-gb': enCatalog,
  fil: filCatalog,
  tl: filCatalog,
  de: deCatalog
}

/** Reactive store for custom drop-in translations loaded from disk */
export const customCatalogsStore = writable<Record<string, TranslationCatalog>>(
  {}
)

/**
 * Derived list of all available languages (built-in + local community drop-ins).
 */
export const availableLocalesStore = derived(
  customCatalogsStore,
  ($customCatalogs) => {
    const list: LocaleOption[] = [...BUILTIN_LOCALES]
    const customKeys = Object.keys($customCatalogs)

    for (const key of customKeys) {
      if (!list.some((l) => l.id.toLowerCase() === key.toLowerCase())) {
        list.push({
          id: key,
          name: `${key.toUpperCase()} (Custom)`,
          flag: key,
          isCustom: true
        })
      }
    }
    return list
  }
)

/**
 * Resolves a nested dot-notation path inside a catalog object.
 *
 * @param catalog - Translation dictionary.
 * @param path - Dot-delimited path (e.g. 'home.good_morning').
 * @returns Resolved string value or undefined.
 */
function resolvePath(
  catalog: TranslationCatalog | undefined,
  path: string
): string | undefined {
  if (!catalog) return undefined
  const parts = path.split('.')
  let current: any = catalog

  for (const part of parts) {
    if (current && typeof current === 'object' && part in current) {
      current = current[part]
    } else {
      return undefined
    }
  }

  return typeof current === 'string' ? current : undefined
}

/**
 * Interpolates tokens like {name} or {count} within a translation string.
 *
 * @param text - Raw translation string.
 * @param params - Key-value replacements.
 * @returns Formatted string.
 */
function interpolate(
  text: string,
  params?: Record<string, string | number>
): string {
  if (!params) return text
  return text.replace(/\{(\w+)\}/g, (_, key) => {
    return params[key] !== undefined ? String(params[key]) : `{${key}}`
  })
}

/**
 * Synchronous lookup function using the currently active locale.
 *
 * @param key - Translation key in dot notation (e.g. 'common.save').
 * @param params - Optional token interpolation values.
 * @returns Translated string with English fallback.
 */
export function t(
  key: string,
  params?: Record<string, string | number>
): string {
  const currentLocale = get(localeStore) || 'en'
  const customCatalogs = get(customCatalogsStore)

  // 1. Try active locale (custom drop-in takes precedence, then embedded)
  const activeCatalog =
    customCatalogs[currentLocale] || embeddedCatalogs[currentLocale]
  let result = resolvePath(activeCatalog, key)

  // 2. Fall back to embedded English
  if (!result && currentLocale !== 'en') {
    result = resolvePath(embeddedCatalogs.en, key)
  }

  // 3. Fall back to the raw key if untranslated
  if (!result) {
    result = key
  }

  return interpolate(result, params)
}

/**
 * Reactive Svelte translation store: `$t('common.save')`.
 * Re-evaluates automatically when the language is changed or custom translations reload.
 */
export const tStore = derived(
  [localeStore, customCatalogsStore],
  ([$locale, $customCatalogs]) => {
    return (key: string, params?: Record<string, string | number>): string => {
      const activeCatalog =
        $customCatalogs[$locale] || embeddedCatalogs[$locale]
      let res = resolvePath(activeCatalog, key)

      if (!res && $locale !== 'en') {
        res = resolvePath(embeddedCatalogs.en, key)
      }
      if (!res) {
        res = key
      }
      return interpolate(res, params)
    }
  }
)

/**
 * Scans %APPDATA%/FreeMatrix Studios/Vinta/locales/ via the Go backend
 * and loads any drop-in JSON files into memory.
 */
export async function loadCustomLocales(): Promise<void> {
  if (typeof GetCustomLocales !== 'function') return

  try {
    const rawMap = await GetCustomLocales()
    if (rawMap && typeof rawMap === 'object') {
      const parsed: Record<string, TranslationCatalog> = {}

      for (const [code, jsonStr] of Object.entries(rawMap)) {
        try {
          const data = JSON.parse(jsonStr)
          if (data && typeof data === 'object') {
            parsed[code] = data
          }
        } catch {
          // Ignore invalid JSON drop-in files
        }
      }

      customCatalogsStore.set(parsed)
    }
  } catch {
    // Ignore backend query failure
  }
}

/**
 * Opens the local community translations directory in Windows Explorer.
 */
export async function openTranslationsFolder(): Promise<void> {
  if (typeof OpenLocalesFolder === 'function') {
    try {
      await OpenLocalesFolder()
    } catch {
      // Ignore open folder failure
    }
  }
}

/**
 * Detects the user's preferred language from host system environment.
 */
export function detectSystemLocale(): string {
  if (typeof navigator === 'undefined' || !navigator.language) {
    return 'en'
  }

  const raw = navigator.language.toLowerCase()
  const code = raw.split(/[-_]/)[0]

  const available = get(availableLocalesStore)
  const match = available.find((l) => l.id.toLowerCase() === code)
  return match ? match.id : 'en'
}

// Automatically load custom drop-in translations on client startup
if (typeof window !== 'undefined') {
  setTimeout(() => {
    void loadCustomLocales()
  }, 100)
}
