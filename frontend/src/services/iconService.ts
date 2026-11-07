/**
 * @fileoverview Extracts and caches native Windows application icons.
 *
 * Requests high-resolution icons for running apps and Windows Store packages from
 * the native backend, and parses Windows app IDs (AUMIDs) into clean display names.
 */

import { writable } from 'svelte/store'
import { GetAppIcon } from '../../bindings/Vinta/backend/app.js'

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

  // If already contains spaces and no path/bang, treat as formatted display name
  const isFormatted =
    s.includes(' ') && !s.includes('\\') && !s.includes('/') && !s.includes('!')

  // 1. Drop filesystem path if present
  const lastSlash = Math.max(s.lastIndexOf('/'), s.lastIndexOf('\\'))
  if (lastSlash >= 0) s = s.substring(lastSlash + 1)

  // 2. Chromium browsers: <AppName>.UserData.<Profile>
  // Split on case-insensitive ".userdata" and keep first segment
  const userDataIdx = s.toLowerCase().indexOf('.userdata')
  if (userDataIdx > 0) {
    s = s.substring(0, userDataIdx)
  }

  // 3. Known Windows system toast IDs
  const lowerRaw = s.toLowerCase()
  if (lowerRaw === 'windows.systemtoast.startupapp') {
    return { displayName: 'Startup Apps', processName: 'startupapp' }
  }
  if (
    lowerRaw === 'windows.systemtoast.bthprops' ||
    lowerRaw === 'windows.systemtoast.bthquickpair'
  ) {
    return { displayName: 'Bluetooth', processName: 'bthprops' }
  }
  if (lowerRaw === 'windows.systemtoast.audiotroubleshooter') {
    return {
      displayName: 'Audio Troubleshooter',
      processName: 'audiotroubleshooter'
    }
  }
  if (lowerRaw === 'windows.systemtoast.autoplay') {
    return { displayName: 'AutoPlay', processName: 'autoplay' }
  }
  if (lowerRaw === 'windows.systemtoast.backgroundaccess') {
    return {
      displayName: 'Background Apps',
      processName: 'backgroundaccess'
    }
  }
  if (
    lowerRaw === 'windows.systemtoast.calling' ||
    lowerRaw === 'windows.systemtoast.calling.systemalertnotification'
  ) {
    return { displayName: 'Phone Link', processName: 'calling' }
  }
  if (lowerRaw === 'windows.defender') {
    return { displayName: 'Windows Security', processName: 'defender' }
  }
  if (lowerRaw.includes('immersivecontrolpanel')) {
    return {
      displayName: 'Settings',
      processName: 'systemsettings'
    }
  }

  // 4. UWP / Store apps: <PackageFamilyName>!<AppId>
  const bangIdx = s.lastIndexOf('!')
  if (bangIdx >= 0 && bangIdx + 1 < s.length) {
    const appId = s.substring(bangIdx + 1)
    const lowerAppId = appId.toLowerCase()
    // If generic application identifier (App, Application, etc.), extract brand from package prefix
    if (
      lowerAppId === 'app' ||
      lowerAppId === 'application' ||
      lowerAppId === 'default' ||
      lowerAppId === 'views'
    ) {
      let pkg = s.substring(0, bangIdx)
      const underscoreIdx = pkg.indexOf('_')
      if (underscoreIdx > 0) pkg = pkg.substring(0, underscoreIdx)
      const dotIdx = pkg.lastIndexOf('.')
      if (dotIdx >= 0) pkg = pkg.substring(dotIdx + 1)
      s = pkg
    } else {
      s = appId
    }
  }

  // 5. Strip .exe or .app extensions
  if (s.toLowerCase().endsWith('.exe')) {
    s = s.substring(0, s.length - 4)
  } else if (s.toLowerCase().endsWith('.app')) {
    s = s.substring(0, s.length - 4)
  }

  // 6. Reverse-domain / vendor prefix strings: take last dot-separated token
  if (!isFormatted) {
    const dotIdx = s.lastIndexOf('.')
    if (dotIdx >= 0 && dotIdx + 1 < s.length) {
      s = s.substring(dotIdx + 1)
    }
  }

  // 7. Strip package suffix after underscore
  const underscoreIdx = s.indexOf('_')
  if (underscoreIdx > 0) {
    s = s.substring(0, underscoreIdx)
  }

  // 8. Split PascalCase / CamelCase words into spaced words (e.g. VisualStudioCode -> Visual Studio Code)
  if (!s.includes(' ')) {
    s = s
      .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
      .replace(/([A-Z]+)([A-Z][a-z])/g, '$1 $2')
      .trim()
  }

  const processName = s.toLowerCase().replace(/\s+/g, '')
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
