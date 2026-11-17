/**
 * @fileoverview Reactive translation and localization service.
 *
 * Handles translation lookups, {token} interpolation, English fallbacks,
 * and loads user-added JSON translation files from `%APPDATA%/FreeMatrix Studios/Vinta/locales`.
 */

import { writable, derived, get } from 'svelte/store'
import { localeStore } from '../stores/settings'
import enCatalog from '../locales/en.json'
import enGbCatalog from '../locales/en-GB.json'
import deCatalog from '../locales/de.json'
import filCatalog from '../locales/fil.json'
import es419Catalog from '../locales/es-419.json'
import esEsCatalog from '../locales/es-ES.json'
import esArCatalog from '../locales/es-AR.json'
import hiCatalog from '../locales/hi.json'
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

export interface LocaleMetadata {
  authors: string[]
  lastUpdated: string
}

/** Standard embedded locale metadata */
export const BUILTIN_LOCALES: LocaleOption[] = [
  { id: 'en', name: 'English (US)', flag: 'en' },
  { id: 'en-GB', name: 'English (UK)', flag: 'en-gb' },
  { id: 'fil', name: 'Filipino', flag: 'fil' },
  { id: 'de', name: 'Deutsch', flag: 'de' },
  { id: 'es-419', name: 'Español (Latinoamérica)', flag: 'es-419' },
  { id: 'es-ES', name: 'Español (España)', flag: 'es-es' },
  { id: 'es-AR', name: 'Español (Argentina)', flag: 'es-ar' },
  { id: 'hi', name: 'हिन्दी', flag: 'hi' }
]

type TranslationCatalog = Record<string, any>

const embeddedCatalogs: Record<string, TranslationCatalog> = {
  en: enCatalog,
  'en-gb': enGbCatalog,
  'en-GB': enGbCatalog,
  fil: filCatalog,
  tl: filCatalog,
  de: deCatalog,
  'es-419': es419Catalog,
  'es-es': esEsCatalog,
  'es-ES': esEsCatalog,
  es: esEsCatalog,
  'es-ar': esArCatalog,
  'es-AR': esArCatalog,
  hi: hiCatalog
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
  const normalizedKey = currentLocale.toLowerCase()

  // 1. Try active locale (custom drop-in takes precedence, then embedded)
  const activeCatalog =
    customCatalogs[currentLocale] ||
    customCatalogs[normalizedKey] ||
    embeddedCatalogs[currentLocale] ||
    embeddedCatalogs[normalizedKey]
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
      const activeLocale = $locale || 'en'
      const normalizedKey = activeLocale.toLowerCase()
      const activeCatalog =
        $customCatalogs[activeLocale] ||
        $customCatalogs[normalizedKey] ||
        embeddedCatalogs[activeLocale] ||
        embeddedCatalogs[normalizedKey]
      let res = resolvePath(activeCatalog, key)

      if (!res && activeLocale !== 'en') {
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
 * Returns @metadata (authors and last-updated) for a given locale ID, if present.
 */
export function getLocaleMetadata(
  localeId: string,
  customCatalogs: Record<string, TranslationCatalog> = get(customCatalogsStore)
): LocaleMetadata | null {
  const normalized = localeId.toLowerCase()
  const catalog =
    customCatalogs[localeId] ||
    customCatalogs[normalized] ||
    embeddedCatalogs[localeId] ||
    embeddedCatalogs[normalized]
  const meta = catalog?.['@metadata']
  if (!meta || typeof meta !== 'object') return null
  const authors = Array.isArray(meta.authors)
    ? meta.authors.filter((a: unknown): a is string => typeof a === 'string' && a.trim().length > 0)
    : []
  const lastUpdated = typeof meta['last-updated'] === 'string' ? meta['last-updated'] : ''
  if (authors.length === 0 && !lastUpdated) return null
  return { authors, lastUpdated }
}

/**
 * Detects the user's preferred language from host system environment.
 */
export function detectSystemLocale(): string {
  if (typeof navigator === 'undefined' || !navigator.language) {
    return 'en'
  }

  const raw = navigator.language.toLowerCase().replace(/_/g, '-')
  const available = get(availableLocalesStore)

  const exact = available.find((l) => l.id.toLowerCase() === raw)
  if (exact) return exact.id

  if (raw.startsWith('es-ar')) return 'es-AR'
  if (raw.startsWith('es-es')) return 'es-ES'
  if (raw.startsWith('es-')) return 'es-419'
  if (raw === 'es') return 'es-ES'

  const base = raw.split('-')[0]
  const baseMatch = available.find((l) => l.id.toLowerCase() === base)
  return baseMatch ? baseMatch.id : 'en'
}

// Automatically load custom drop-in translations on client startup
if (typeof window !== 'undefined') {
  setTimeout(() => {
    void loadCustomLocales()
  }, 100)
}
