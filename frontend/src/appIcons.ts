// This file maps Windows application identifiers and display names to curated icons,
// categories, and process targeting metadata.
//
// ARCHITECTURE NOTE:
// This catalog and resolver (resolveAppBrand) are shared across two core Vinta features:
// 1. Notification Center: Matches human-readable OS display names intercepted via WinRT
//    UserNotificationListener (Windows.UI.Notifications.Management) to display brand icons
//    and group notification cards.
// 2. Media Player: Matches session source identifiers from WinRT SMTC
//    (GlobalSystemMediaTransportControlsSession.SourceAppUserModelId) to resolve the media
//    brand, badge icon, and executable process token for source app window focusing.
//
// WHY THIS FILE EXISTS:
// Windows APIs frequently expose disparate identifiers across Win32, UWP, and WinRT sessions
// (e.g., "Google Chrome", "Google.Chrome", "chrome.exe", or raw package IDs), and GetLogo() byte
// streams incur expensive encoding overhead. This centralized dictionary maps those variations
// to lightweight CDN SVGs, semantic categories, and executable process names.
//
// Licenses & Fair Use:
// All brand icons and logos linked in this dictionary are the property of their
// respective trademark and copyright holders. They are loaded dynamically for
// identification purposes only under fair use for visual representation within the OS.
//
// NOTE TO CONTRIBUTORS:
// DO NOT STORE STATIC LOGO IMAGE FILES IN THIS REPOSITORY. All assets must be loaded via CDN.

export type AppCategory =
  | 'communication' // Messaging, email, & video calls (Discord, Telegram, Outlook)
  | 'gaming' // Games & game launchers (Steam, CS2, Valorant)
  | 'media' // Audio, video, & streaming tools (Spotify, Netflix, OBS Studio)
  | 'productivity' // Office, calendars, utilities (Word, Excel)
  | 'browser' // Web browsers (Chrome, Edge)
  | 'system' // Drivers & OS utilities (NVIDIA, AMD, Snipping Tool)
  | 'dev' // Coding tools, terminals, IDEs (VS Code, Git)

export interface AppBrand {
  name: string
  icon: string
  matchers: string[]
  fallbackRegexes?: RegExp[] // Used for regex matching against notification content when the OS omits appName
  category?: AppCategory
}

// MATCHING BEHAVIOR & MATCHER RULES:
// 1. Matchers run against lowercased input strings using `.includes()` semantics.
// 2. Evaluation order: `gameIcons` runs first, followed by `appIcons`. Place specific matchers above general ones.
// 3. Process Targeting Rule for Media Player:
//    Because the Media Player switches to source applications via FocusMediaApp(), entries should include
//    both the full display name (e.g., 'google chrome') and the single-word executable token (e.g., 'chrome').
//    The resolver selects the single-word token (without spaces) as `processName` to target the executable image.
// 4. Video games must always be added to `gameIcons`, keeping general game launchers in `appIcons`.

// DO NOT REMOVE THIS
// THIS ENSURES THAT PRETTIER FORMATTER WOULD NOT MESS UP WITH THE MINIFIED FORMAT
// prettier-ignore
export const appIcons: AppBrand[] = [
  // --- COMMUNICATION & CHAT (High Toast Volume) ---
  { name: 'Discord', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/discord/default.svg', matchers: ['discord'], category: 'communication' },
  { name: 'Viber', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/viber/default.svg', matchers: ['viber'], category: 'communication' },
  { name: 'Telegram', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/telegram/default.svg', matchers: ['telegram'], category: 'communication' },
  { name: 'WhatsApp', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/whatsapp/default.svg', matchers: ['whatsapp'], category: 'communication' },
  { name: 'Signal', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/signal/default.svg', matchers: ['signal'], category: 'communication' },
  { name: 'Skype', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/skype/default.svg', matchers: ['skype'], category: 'communication' },
  { name: 'Slack', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/slack/default.svg', matchers: ['slack'], category: 'communication' },
  { name: 'Microsoft Teams', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-teams/default.svg', matchers: ['microsoft teams', 'teams'], category: 'communication' },
  { name: 'Zoom', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/zoom/default.svg', matchers: ['zoom'], category: 'communication' },
  { name: 'Cisco Webex', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/webex/default.svg', matchers: ['cisco webex', 'webex'], category: 'communication' },
  { name: 'Proton Mail', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/protonmail/default.svg', matchers: ['proton mail', 'protonmail'], category: 'communication' },
  { name: 'Phone Link', icon: '', matchers: ['phone link', 'your phone'], category: 'communication' },
  { name: 'Microsoft Outlook', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-outlook/default.svg', matchers: ['microsoft outlook', 'outlook'], category: 'communication' },
  { name: 'Thunderbird', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/thunderbird/default.svg', matchers: ['thunderbird'], category: 'communication' },
  { name: 'Windows Mail', icon: '', matchers: ['windows mail', 'mail'], category: 'communication' },

  // --- BROWSERS (Web Push Notifications - Massive Toast Volume) ---
  { name: 'Google Chrome', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/google-chrome/default.svg', matchers: ['google chrome', 'chrome'], category: 'browser' },
  { name: 'Microsoft Edge', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-edge/default.svg', matchers: ['microsoft edge', 'msedge', 'edge'], category: 'browser' },
  { name: 'Mozilla Firefox', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/firefox/default.svg', matchers: ['mozilla firefox', 'firefox'], category: 'browser' },
  { name: 'Brave Browser', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/brave/default.svg', matchers: ['brave browser', 'brave'], category: 'browser' },
  { name: 'Opera', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/opera/default.svg', matchers: ['opera'], category: 'browser' },
  { name: 'Vivaldi', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/vivaldi/default.svg', matchers: ['vivaldi'], category: 'browser' },

  // --- SYSTEM, CLOUD, SECURITY & HARDWARE DRIVERS (Sync, Autofill, Battery & Clip Alerts) ---
  { name: 'Microsoft OneDrive', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-onedrive/default.svg', matchers: ['microsoft onedrive', 'onedrive'], category: 'system' },
  { name: 'Google Drive', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/google-drive/default.svg', matchers: ['google drive'], category: 'system' },
  { name: 'Dropbox', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/dropbox/default.svg', matchers: ['dropbox'], category: 'system' },
  { name: '1Password', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/1password/default.svg', matchers: ['1password'], category: 'system' },
  { name: 'Bitwarden', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/bitwarden/default.svg', matchers: ['bitwarden'], category: 'system' },
  { name: 'ShareX', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/sharex/default.svg', matchers: ['sharex'], category: 'system' },
  { name: 'NVIDIA App', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/nvidia/default.svg', matchers: ['nvidia app', 'geforce experience'], category: 'system' },
  { name: 'AMD Software', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/amd/default.svg', matchers: ['amd software', 'radeon software'], category: 'system' },
  { name: 'Logitech G HUB', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/logitech/default.svg', matchers: ['lghub', 'logitech g hub', 'logitech'], category: 'system' },
  { name: 'Razer Synapse', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/razer/default.svg', matchers: ['razer synapse', 'razer central', 'razer'], category: 'system' },
  { name: 'SteelSeries GG', icon: '', matchers: ['steelseries gg', 'steelseries'], category: 'system' },
  { name: 'Elgato Stream Deck', icon: '', matchers: ['elgato stream deck', 'stream deck', 'elgato'], category: 'system' },
  { name: 'NordVPN', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/nordvpn/default.svg', matchers: ['nordvpn'], category: 'system' },
  { name: 'Malwarebytes', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/malwarebytes/default.svg', matchers: ['malwarebytes'], category: 'system' },
  { name: 'Windows Security', icon: '', matchers: ['windows security'], category: 'system' },
  { name: 'Windows Clock', icon: '', matchers: ['windows clock', 'clock', 'alarms & clock'], category: 'system' },
  { name: 'Snipping Tool', icon: '', matchers: ['snipping tool', 'snip & sketch'], category: 'system' },

  // --- DEV TOOLS (Build & Container Alerts) ---
  { name: 'Docker Desktop', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/docker/default.svg', matchers: ['docker desktop', 'docker'], category: 'dev' },
  { name: 'Visual Studio Code', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/visual-studio-code/default.svg', matchers: ['visual studio code', 'vscode', 'code'], category: 'dev' },
  { name: 'GitHub Desktop', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/github/default.svg', matchers: ['github desktop'], category: 'dev' },
  { name: 'JetBrains Toolbox', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/jetbrains/default.svg', matchers: ['jetbrains toolbox', 'intellij', 'webstorm', 'pycharm'], category: 'dev' },

  // --- MEDIA, CAPTURE & STREAMING (Clip Saved & Recording Alerts) ---
  { name: 'Spotify', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/spotify/default.svg', matchers: ['spotify'], category: 'media' },
  { name: 'Apple Music', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/apple-music/default.svg', matchers: ['apple music', 'applemusic', 'appleinc.applemusic'], category: 'media' },
  { name: 'YouTube Music', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/youtube-music/default.svg', matchers: ['youtube music', 'youtubemusic'], category: 'media' },
  { name: 'TIDAL', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/tidal/default.svg', matchers: ['tidal'], category: 'media' },
  { name: 'SoundCloud', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/soundcloud/default.svg', matchers: ['soundcloud'], category: 'media' },
  { name: 'Deezer', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/deezer/default.svg', matchers: ['deezer'], category: 'media' },
  { name: 'VLC Media Player', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/vlc/default.svg', matchers: ['vlc', 'videolan'], category: 'media' },
  { name: 'Windows Media Player', icon: '', matchers: ['windows media player', 'wmplayer', 'zunemusic', 'microsoft.zunemusic'], category: 'media' },
  { name: 'foobar2000', icon: '', matchers: ['foobar2000'], category: 'media' },
  { name: 'OBS Studio', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/obs-studio/default.svg', matchers: ['obs studio', 'obs'], category: 'media' },
  { name: 'Medal.tv', icon: '', matchers: ['medal.tv', 'medal'], category: 'media' },
  { name: 'Outplayed', icon: '', matchers: ['outplayed', 'overwolf'], category: 'media' },
  { name: 'Loom', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/loom/default.svg', matchers: ['loom'], category: 'media' },
  { name: 'Netflix', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/netflix/default.svg', matchers: ['netflix'], category: 'media' },

  // --- PRODUCTIVITY, OFFICE & TASKS (Reminders & Alerts) ---
  { name: 'Notion', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/notion/default.svg', matchers: ['notion'], category: 'productivity' },
  { name: 'Todoist', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/todoist/default.svg', matchers: ['todoist'], category: 'productivity' },
  { name: 'Figma', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/figma/default.svg', matchers: ['figma'], category: 'productivity' },
  { name: 'Trello', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/trello/default.svg', matchers: ['trello'], category: 'productivity' },
  { name: 'Windows Calendar', icon: '', matchers: ['windows calendar', 'calendar'], category: 'productivity' },
  { name: 'Grass', icon: 'https://cdn.jsdelivr.net/gh/selfhst/icons/svg/grass-io.svg', matchers: ['grass'], category: 'system' },

  // --- GAMING LAUNCHERS & CLIENTS (Native Toast Emitters Only) ---
  { name: 'Epic Games Launcher', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/epic-games/default.svg', matchers: ['epic games launcher', 'epic games'], category: 'gaming' },
  { name: 'Xbox App', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/xbox/default.svg', matchers: ['xbox app', 'xbox'], category: 'gaming' },
  { name: 'EA App', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/ea/default.svg', matchers: ['ea app', 'electronic arts'], category: 'gaming' },
  { name: 'Ubisoft Connect', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/ubisoft/default.svg', matchers: ['ubisoft connect', 'uplay'], category: 'gaming' },
  { name: 'GOG Galaxy', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/gog/default.svg', matchers: ['gog galaxy'], category: 'gaming' },
  { name: 'Riot Client', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/riot-games/default.svg', matchers: ['riot client', 'riot games'], category: 'gaming' },
  { name: 'HoYoPLAY', icon: '', matchers: ['hoyoplay', 'mihoyo', 'hoyoverse'], category: 'gaming' }
]

// DO NOT REMOVE THIS
// THIS ENSURES THAT PRETTIER FORMATTER WOULD NOT MESS UP WITH THE MINIFIED FORMAT
// prettier-ignore
export const gameIcons: AppBrand[] = [
  // --- PC GAMES WITH ACTIVE OS TOAST NOTIFICATIONS ---
  { name: 'Rise of Kingdoms', icon: 'https://static.wikia.nocookie.net/riseofcivilizations/images/e/e6/Site-logo.png', matchers: ['rise of kingdoms', 'riseofkingdoms', 'rok'], fallbackRegexes: [/^dear .*, your troops have returned to the city\.?$/i, /^.* has launched a \d+ (minute|hour) rally!?$/i, /^.* training complete!?$/i, /^resource collection$/i], category: 'gaming' },
  { name: 'Clash of Clans', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/clash-of-clans/default.svg', matchers: ['clash of clans', 'google play games'], category: 'gaming' },
  { name: 'Roblox', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/roblox/default.svg', matchers: ['roblox'], category: 'gaming' },
  { name: 'Genshin Impact', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/genshin-impact/default.svg', matchers: ['genshin impact', 'genshinimpact'], category: 'gaming' },
  { name: 'Honkai: Star Rail', icon: '', matchers: ['honkai: star rail', 'star rail'], category: 'gaming' },
  { name: 'Zenless Zone Zero', icon: '', matchers: ['zenless zone zero', 'zzz'], category: 'gaming' },
  { name: 'RAID: Shadow Legends', icon: '', matchers: ['raid: shadow legends', 'plarium play'], category: 'gaming' },
  { name: 'Warframe', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/warframe/default.svg', matchers: ['warframe'], category: 'gaming' },
  { name: 'Albion Online', icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/albion-online/default.svg', matchers: ['albion online', 'albion'], category: 'gaming' }
]

// TO ADD
// ADOBE READER AND ACROBAT MANAGER
// FORTINET CLIENT
// OTHER VPNS

import { writable } from 'svelte/store'
import { CacheIcons } from '../wailsjs/go/backend/Cache.js'
import { GetAppIcon } from '../wailsjs/go/backend/App.js'

// In-memory client cache for native resolved icons
const nativeIconCache = new Map<string, string>()
const pendingRequests = new Set<string>()

// Reactive store for extracted icons across components
export const nativeIconsStore = writable<Record<string, string>>({})

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
      (window as unknown as { go?: { backend?: { App?: { GetAppIcon?: (a: string, b: string) => Promise<string> } } } })?.go?.backend?.App?.GetAppIcon
    ) {
      uri = await (window as unknown as { go: { backend: { App: { GetAppIcon: (a: string, b: string) => Promise<string> } } } }).go.backend.App.GetAppIcon(sApp, sAumid)
    }
    const res = uri || ''
    nativeIconCache.set(key, res)
    if (res) {
      nativeIconsStore.update((curr) => ({ ...curr, [key]: res }))
    }
    return res
  } catch (err) {
    nativeIconCache.set(key, '')
    return ''
  } finally {
    pendingRequests.delete(key)
  }
}

export function requestNativeAppIcon(sourceApp?: string, aumid?: string): string {
  const sApp = sourceApp?.trim() || ''
  const sAumid = aumid?.trim() || ''
  if (!sApp && !sAumid) return ''

  const key = `${sApp.toLowerCase()}|${sAumid.toLowerCase()}`
  if (nativeIconCache.has(key)) {
    return nativeIconCache.get(key)!
  }

  if (!pendingRequests.has(key)) {
    pendingRequests.add(key)
    getNativeAppIcon(sApp, sAumid).catch(() => {})
  }
  return ''
}

export async function preloadIcons() {
  const urls: string[] = []
  const allIcons = [...appIcons, ...gameIcons]

  for (const brand of allIcons) {
    if (brand.icon && brand.icon.startsWith('http')) {
      urls.push(brand.icon)
    }
  }

  if (urls.length > 0) {
    try {
      const cachedMap = await CacheIcons(urls)
      for (const brand of allIcons) {
        if (brand.icon && cachedMap[brand.icon]) {
          brand.icon = cachedMap[brand.icon]
        }
      }
    } catch (err) {
      console.error('Failed to preload SVG icons:', err)
    }
  }
}

export function findAppIconConflicts(): string[] {
  const conflicts: string[] = []
  const allIcons = [...gameIcons, ...appIcons]
  for (let i = 0; i < allIcons.length; i++) {
    const brand = allIcons[i]
    for (const matcher of brand.matchers) {
      for (let j = 0; j < i; j++) {
        const prevBrand = allIcons[j]
        for (const prevMatcher of prevBrand.matchers) {
          if (matcher === prevMatcher) {
            conflicts.push(
              `Shadowed: '${brand.name}' matcher '${matcher}' is intercepted by '${prevBrand.name}' matcher '${prevMatcher}' (placed earlier)`
            )
          }
        }
      }
    }
  }
  return conflicts
}

// In-memory Base64 image cache to prevent redundant DOM re-allocations
const imageCache = new Map<string, string>()

export function getCachedBase64Url(url: string | undefined): string {
  if (!url) return ''
  if (!url.startsWith('data:image')) return url
  if (imageCache.has(url)) return imageCache.get(url)!

  imageCache.set(url, url)
  if (imageCache.size > 100) {
    const oldestKey = imageCache.keys().next().value
    if (oldestKey) imageCache.delete(oldestKey)
  }
  return url
}

export function resolveAppBrand(
  sourceApp?: string,
  aumidOverride?: string
): {
  name: string
  icon: string
  processName?: string
  aumid?: string
  isCurated?: boolean
} {
  if (!sourceApp && !aumidOverride) {
    return { name: 'Media Player', icon: '', isCurated: false }
  }
  const s = (sourceApp || '').toLowerCase()
  const allIcons = [...appIcons, ...gameIcons]

  let matchedBrand: (typeof allIcons)[0] | undefined
  for (const brand of allIcons) {
    const matched = brand.matchers.find((m) => s.includes(m.toLowerCase()))
    if (matched) {
      matchedBrand = brand
      break
    }
  }

  let processName = ''
  let aumid =
    aumidOverride || (sourceApp && sourceApp.includes('!') ? sourceApp : '')
  let cleanName = ''

  if (matchedBrand) {
    cleanName = matchedBrand.name
    processName =
      matchedBrand.matchers.find(
        (m) => !m.includes(' ') && !m.includes('!')
      ) || ''
    if (!aumid) {
      aumid = matchedBrand.matchers.find((m) => m.includes('!')) || ''
    }
  } else {
    cleanName = sourceApp || aumidOverride || ''
    const lastSlash = Math.max(
      cleanName.lastIndexOf('/'),
      cleanName.lastIndexOf('\\')
    )
    if (lastSlash >= 0) cleanName = cleanName.substring(lastSlash + 1)
    const dotExe = cleanName.toLowerCase().indexOf('.exe')
    if (dotExe >= 0) cleanName = cleanName.substring(0, dotExe)
    const bang = cleanName.lastIndexOf('!')
    if (bang >= 0) cleanName = cleanName.substring(bang + 1)
    const underscore = cleanName.indexOf('_')
    if (underscore >= 0) cleanName = cleanName.substring(0, underscore)

    processName = cleanName.toLowerCase()
    if (cleanName.length > 0) {
      cleanName = cleanName.charAt(0).toUpperCase() + cleanName.slice(1)
    } else {
      cleanName = 'Media Player'
    }
  }

  // Priority 1 (Option A): Request or check native extracted icon from Windows (.exe / AUMID)
  const searchApp = processName || sourceApp || ''
  const nativeIcon =
    requestNativeAppIcon(searchApp, aumid) ||
    (sourceApp && sourceApp !== searchApp
      ? requestNativeAppIcon(sourceApp, aumid)
      : '')

  // Priority 2 (Option B backup): Curated vector SVG if native icon is not yet ready or unavailable
  const effectiveIcon = nativeIcon || matchedBrand?.icon || ''

  return {
    name: cleanName,
    icon: effectiveIcon,
    processName,
    aumid,
    isCurated: Boolean(matchedBrand?.icon)
  }
}
