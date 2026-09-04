import { writable } from 'svelte/store'
import { GetAppIcon } from '../../wailsjs/go/backend/App.js'

// In-memory client cache for native resolved icons (key -> base64 data URI)
const nativeIconCache = new Map<string, string>()
const pendingRequests = new Set<string>()

// Reactive store for extracted icons across components (key -> data URI)
export const nativeIconsStore = writable<Record<string, string>>({})

/**
 * Directly fetches the native icon for a given process name or AUMID as a base64 data URI.
 */
export async function getNativeAppIcon(
  sourceApp?: string,
  aumid?: string
): Promise<string> {
  const sApp = sourceApp?.trim() || ''
  const sAumid = aumid?.trim() || ''
  if (!sApp && !sAumid) return ''

  const key = `${sApp.toLowerCase()}|${sAumid.toLowerCase()}`
  if (nativeIconCache.has(key)) {
    return nativeIconCache.get(key)!
  }

  try {
    let uri = ''
    if (typeof GetAppIcon === 'function') {
      uri = await GetAppIcon(sApp, sAumid)
    } else if (
      typeof window !== 'undefined' &&
      (
        window as unknown as {
          go?: {
            backend?: {
              App?: { GetAppIcon?: (a: string, b: string) => Promise<string> }
            }
          }
        }
      )?.go?.backend?.App?.GetAppIcon
    ) {
      uri = await (
        window as unknown as {
          go: {
            backend: {
              App: { GetAppIcon: (a: string, b: string) => Promise<string> }
            }
          }
        }
      ).go.backend.App.GetAppIcon(sApp, sAumid)
    }
    const res = uri || ''
    nativeIconCache.set(key, res)
    nativeIconsStore.update((store) => ({ ...store, [key]: res }))
    return res
  } catch (err) {
    console.warn(
      `[IconService] Failed to extract native icon for '${sApp}':`,
      err
    )
    nativeIconCache.set(key, '')
    return ''
  }
}

/**
 * Returns the cached native icon immediately if present;
 * otherwise dispatches an asynchronous background fetch and updates nativeIconsStore.
 */
export function requestNativeAppIcon(
  sourceApp?: string,
  aumid?: string
): string {
  const sApp = sourceApp?.trim() || ''
  const sAumid = aumid?.trim() || ''
  if (!sApp && !sAumid) return ''

  const key = `${sApp.toLowerCase()}|${sAumid.toLowerCase()}`
  if (nativeIconCache.has(key)) {
    return nativeIconCache.get(key)!
  }

  if (!pendingRequests.has(key)) {
    pendingRequests.add(key)
    getNativeAppIcon(sApp, sAumid).finally(() => {
      pendingRequests.delete(key)
    })
  }

  return ''
}

/**
 * Universally parses Windows SourceAppUserModelId (AUMID) or process strings into
 * clean display names and executable tokens without hardcoding vendor names.
 *
 * Handles:
 * 1. Chromium browsers: <AppName>.UserData.<Profile> (e.g. Chrome.UserData.Profile1)
 * 2. UWP / Store apps: <PackageFamilyName>!<AppId> (e.g. SpotifyAB.SpotifyMusic_zpdnekdrzrea0!Spotify)
 * 3. Reverse-domain / Win32 strings: (e.g. Microsoft.ZuneMusic -> ZuneMusic, vlc.exe -> Vlc)
 */
export function parseAumid(raw?: string): {
  displayName: string
  processName: string
} {
  if (!raw) return { displayName: 'Notification', processName: '' }

  let s = raw.trim()

  // Drop composite session suffix if present (e.g. AUMID::index::title)
  const doubleColonIdx = s.indexOf('::')
  if (doubleColonIdx >= 0) {
    s = s.substring(0, doubleColonIdx)
  }

  // 1. Drop filesystem path if present
  const lastSlash = Math.max(s.lastIndexOf('/'), s.lastIndexOf('\\'))
  if (lastSlash >= 0) s = s.substring(lastSlash + 1)

  // 2. Chromium browsers: <AppName>.UserData.<Profile>
  // Split on case-insensitive ".userdata." and keep first segment
  const userDataIdx = s.toLowerCase().indexOf('.userdata')
  if (userDataIdx > 0) {
    s = s.substring(0, userDataIdx)
  }

  // 3. UWP / Store apps: <PackageFamilyName>!<AppId>
  // Isolate app identifier after '!'
  const bangIdx = s.lastIndexOf('!')
  if (bangIdx >= 0 && bangIdx + 1 < s.length) {
    s = s.substring(bangIdx + 1)
  }

  // 4. Strip .exe if present
  if (s.toLowerCase().endsWith('.exe')) {
    s = s.substring(0, s.length - 4)
  }

  // 5. Reverse-domain / vendor prefix strings: take last dot-separated token
  const dotIdx = s.lastIndexOf('.')
  if (dotIdx >= 0 && dotIdx + 1 < s.length) {
    s = s.substring(dotIdx + 1)
  }

  // 6. Strip package suffix after underscore
  const underscoreIdx = s.indexOf('_')
  if (underscoreIdx > 0) {
    s = s.substring(0, underscoreIdx)
  }

  const processName = s.toLowerCase()
  const displayName =
    s.length > 0 ? s.charAt(0).toUpperCase() + s.slice(1) : 'Notification'

  return { displayName, processName }
}

/**
 * Resolves the application brand, display name, and authentic native icon for a media session or notification.
 */
export function resolveAppBrand(
  sourceApp?: string,
  aumidOverride?: string
): {
  name: string
  icon: string
  processName?: string
  aumid?: string
} {
  if (!sourceApp && !aumidOverride) {
    return { name: 'Media Player', icon: '' }
  }

  const rawInput = sourceApp || aumidOverride || ''
  const parsed = parseAumid(rawInput)

  let aumid =
    aumidOverride || (sourceApp && sourceApp.includes('!') ? sourceApp : '')
  if (aumid.includes('::')) {
    aumid = aumid.substring(0, aumid.indexOf('::'))
  }

  const searchApp = parsed.processName || sourceApp || ''
  const nativeIcon =
    requestNativeAppIcon(searchApp, aumid) ||
    (parsed.processName && parsed.processName !== searchApp
      ? requestNativeAppIcon(parsed.processName, aumid)
      : '') ||
    (sourceApp && sourceApp !== searchApp
      ? requestNativeAppIcon(sourceApp, aumid)
      : '')

  return {
    name: parsed.displayName,
    icon: nativeIcon || '',
    processName: parsed.processName,
    aumid
  }
}
