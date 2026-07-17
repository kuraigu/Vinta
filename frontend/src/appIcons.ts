// This file maps Windows application display names to curated icons and categories.
//
// ARCHITECTURE NOTE: Vinta uses the official WinRT UserNotificationListener API
// (Windows.UI.Notifications.Management) to intercept toast notifications. The backend
// resolves each notification's app name via IAppDisplayInfo.get_DisplayName — so matchers
// here run against the real, human-readable OS display name (e.g. "Discord", "NVIDIA App")
// rather than raw system IDs.
//
// WHY THIS FILE STILL EXISTS:
// The WinRT API gives us app names but not icons. GetLogo() exists but returns a raw byte
// stream that would require expensive Base64 encoding on every notification. Instead, we
// match the display name against this dictionary to load a high-quality CDN icon and assign
// a semantic category (used for fallback icon rendering).
//
//
// We primarily use thesvg.org CDN for major developer brands and applications.
// For video games and apps not in that CDN, the source may vary.
//
//
// Licenses & Fair Use:
// All brand icons and logos linked in this dictionary are the property of their
// respective trademark and copyright holders. They are loaded dynamically for
// identification purposes only. Their usage here constitutes "fair use" under
// copyright law, as they are utilized solely to visually identify the source
// of incoming notifications on the user's personal operating system.
//
//
// NOTE TO ALL CONTRIBUTORS:
// PLEASE AVOID PUTTING ANY FILE IMAGE OF THE LOGOS IN OUR REPO

export type AppCategory =
  | 'communication' // Messaging, email, & video calls (Discord, Telegram, Outlook)
  | 'gaming' // Games & game launchers (Steam, CS2, Valorant)
  | 'media' // Audio, video, & streaming tools (Spotify, OBS Studio)
  | 'productivity' // Office, calendars, utilities (Word, Excel)
  | 'browser' // Web browsers (Chrome, Edge)
  | 'system' // Drivers & OS utilities (NVIDIA, AMD, Snipping Tool)
  | 'dev' // Coding tools, terminals, IDEs (VS Code, Git) - for future use

export interface AppBrand {
  name: string
  icon: string
  matchers: string[]
  category?: AppCategory
}

// IMPORTANT MATCHING BEHAVIOR & HOW TO ADD APPS:
// 1. Matchers run against the lowercased WinRT display name (e.g. 'discord', 'nvidia app').
//    Use `.includes()` semantics — partial matches work, so 'discord' matches 'Discord'.
// 2. The frontend loops through gameIcons FIRST, then appIcons.
//    Order matters: place more specific matchers above general ones.
// 3. Because the OS display name is already human-readable, matchers are usually just
//    the lowercased app name (e.g. 'discord', 'spotify', 'google chrome').
//
// HOW TO LOCATE A MATCHER FOR A NEW APP:
// - Method A: Receive a notification from the app while Vinta is running. The debug log at
//   %LOCALAPPDATA%\vinta_debug.log will show: Emitting notification: app="..."
//   Use that lowercased app name string as your matcher.
// - Method B: Check Windows Settings → Apps → Notifications to see the exact display name
//   Windows uses for the app.
//
// NOTE TO ALL CONTRIBUTORS:
// 1. WE DO NOT NEED TO ADD EVERY SINGLE APPLICATION — THERE IS A FALLBACK ICON.
//    ONLY ADD APPS THAT ACTUALLY SEND NOTIFICATIONS.
// 2. ALL VIDEO GAMES MUST BE PLACED IN THE `gameIcons` ARRAY.
//    Keep general client launchers (Steam, Epic, EA App) in `appIcons`. Because `gameIcons`
//    is evaluated first, putting games there guarantees no conflicts with generic launchers.

export const appIcons: AppBrand[] = [
  // --- COMMUNICATION & CHAT (High Toast Volume) ---
  {
    name: 'Discord',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/discord/default.svg',
    matchers: [
      'com.squirrel.discord.discord',
      'com.squirrel.discord',
      'discord',
      'discord.exe'
    ],
    category: 'communication'
  },
  {
    name: 'Viber',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/viber/default.svg',
    matchers: ['viber', 'viber.exe', 'com.viber.voip'],
    category: 'communication'
  },
  {
    name: 'Telegram',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/telegram/default.svg',
    matchers: [
      'telegram.telegramdesktop.web',
      'telegramdesktop',
      'telegram.exe',
      'telegram'
    ],
    category: 'communication'
  },
  {
    name: 'WhatsApp',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/whatsapp/default.svg',
    matchers: [
      '5319275a.whatsappdesktop_cv1g1gvanyjgm!app',
      'whatsapp',
      'whatsapp.exe'
    ],
    category: 'communication'
  },
  {
    name: 'Signal',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/signal/default.svg',
    matchers: ['org.whispersystems.signal-desktop', 'signal', 'signal.exe'],
    category: 'communication'
  },
  {
    name: 'Skype',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/skype/default.svg',
    matchers: ['microsoft.skypeapp_8wekyb3d8bbwe!app', 'skype', 'skype.exe'],
    category: 'communication'
  },
  {
    name: 'Slack',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/slack/default.svg',
    matchers: [
      'com.squirrel.slack.slack',
      'slack',
      'slack.exe',
      'com.tinyspeck.slackmacgap'
    ],
    category: 'communication'
  },
  {
    name: 'Microsoft Teams',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-teams/default.svg',
    matchers: [
      'msteams_8wekyb3d8bbwe!msteams',
      'microsoft.teams',
      'msteams.exe',
      'com.squirrel.teams.teams',
      'teams.exe'
    ],
    category: 'communication'
  },
  {
    name: 'Zoom',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/zoom/default.svg',
    matchers: ['zoom.us', 'zoom', 'zoom.exe', 'us.zoom.zoom'],
    category: 'communication'
  },
  {
    name: 'Phone Link',
    icon: '',
    matchers: ['microsoft.yourphone_8wekyb3d8bbwe!app'],
    category: 'communication'
  },
  {
    name: 'Microsoft Outlook',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-outlook/default.svg',
    matchers: [
      'microsoft.outlookforwindows_8wekyb3d8bbwe!microsoft.outlookforwindows',
      'microsoft.office.outlook.exe.16',
      'microsoft.office.outlook.exe.15',
      'outlook.exe'
    ],
    category: 'communication'
  },
  {
    name: 'Thunderbird',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/thunderbird/default.svg',
    matchers: ['org.mozilla.thunderbird', 'thunderbird', 'thunderbird.exe'],
    category: 'communication'
  },
  {
    name: 'Windows Mail',
    icon: '',
    matchers: [
      'microsoft.windowscommunicationsapps_8wekyb3d8bbwe!microsoft.windowslive.mail'
    ],
    category: 'communication'
  },

  // --- BROWSERS (Web Push Notifications) ---
  {
    name: 'Google Chrome',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/google-chrome/default.svg',
    matchers: ['chrome', 'google.chrome', 'google chrome', 'chrome.exe'],
    category: 'browser'
  },
  {
    name: 'Microsoft Edge',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-edge/default.svg',
    matchers: [
      'microsoft.microsoftedge_8wekyb3d8bbwe!microsoftedge',
      'msedge',
      'msedge.exe'
    ],
    category: 'browser'
  },

  // --- SYSTEM, CLOUD & OS UTILITIES (Sync, Autofill & Security Alerts) ---
  {
    name: 'Microsoft OneDrive',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-onedrive/default.svg',
    matchers: ['microsoft.skydrive', 'onedrive', 'onedrive.exe'],
    category: 'system'
  },
  {
    name: 'Google Drive',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/google-drive/default.svg',
    matchers: ['googledrivefs.exe', 'google drive', 'backup and sync'],
    category: 'system'
  },
  {
    name: 'Dropbox',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/dropbox/default.svg',
    matchers: ['dropbox', 'dropbox.exe', 'com.dropbox.client'],
    category: 'system'
  },
  {
    name: '1Password',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/1password/default.svg',
    matchers: ['1password', '1password.exe', 'com.agilebits.onepassword'],
    category: 'system'
  },
  {
    name: 'Bitwarden',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/bitwarden/default.svg',
    matchers: ['bitwarden', 'bitwarden.exe', 'com.bitwarden.desktop'],
    category: 'system'
  },
  {
    name: 'ShareX',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/sharex/default.svg',
    matchers: ['sharex', 'sharex.exe'],
    category: 'system'
  },
  {
    name: 'NVIDIA App',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/nvidia/default.svg',
    matchers: [
      'nvapp',
      'nvidia geforce experience',
      'nvidia corporation.geforceexperience',
      'nvidia.geforceexperience'
    ],
    category: 'system'
  },
  {
    name: 'AMD Software',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/amd/default.svg',
    matchers: [
      'advancedmicrodevicesinc-2.amdradeonsoftware_0a9344xs7nr4m!amdradeonsoftware',
      'radeonsoftware.exe'
    ],
    category: 'system'
  },
  {
    name: 'Windows Security',
    icon: '',
    matchers: [
      'microsoft.windows.sechealthui_8wekyb3d8bbwe!sechealthui',
      'securityhealthsystray.exe'
    ],
    category: 'system'
  },
  {
    name: 'Windows Clock',
    icon: '',
    matchers: ['microsoft.windowsalarms_8wekyb3d8bbwe!app'],
    category: 'system'
  },
  {
    name: 'Snipping Tool',
    icon: '',
    matchers: ['microsoft.screensketch_8wekyb3d8bbwe!app'],
    category: 'system'
  },

  // --- DEV TOOLS (Build & Container Alerts) ---
  {
    name: 'Docker Desktop',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/docker/default.svg',
    matchers: ['docker desktop', 'docker desktop.exe', 'com.docker.docker'],
    category: 'dev'
  },
  {
    name: 'Postman',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/postman/default.svg',
    matchers: ['postman', 'postman.exe', 'com.postmanlabs.mac'],
    category: 'dev'
  },

  // --- MEDIA & STREAMING ---
  {
    name: 'Spotify',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/spotify/default.svg',
    matchers: [
      'spotifyab.spotifymusic_zpdnekdrzrea0!spotify',
      'spotify',
      'spotify.exe'
    ],
    category: 'media'
  },
  {
    name: 'OBS Studio',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/obs-studio/default.svg',
    matchers: ['obsstudio', 'obs64.exe', 'obs-studio'],
    category: 'media'
  },

  // --- PRODUCTIVITY, OFFICE & TASKS (Reminders & Alerts) ---
  {
    name: 'Notion',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/notion/default.svg',
    matchers: ['notion.id', 'notion', 'notion.exe'],
    category: 'productivity'
  },
  {
    name: 'Todoist',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/todoist/default.svg',
    matchers: ['todoist.exe', 'com.todoist'],
    category: 'productivity'
  },
  {
    name: 'Microsoft Word',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-word/default.svg',
    matchers: [
      'microsoft.office.winword.exe.16',
      'microsoft.office.winword.exe.15',
      'winword.exe'
    ],
    category: 'productivity'
  },
  {
    name: 'Microsoft Excel',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/microsoft-excel/default.svg',
    matchers: [
      'microsoft.office.excel.exe.16',
      'microsoft.office.excel.exe.15',
      'excel.exe'
    ],
    category: 'productivity'
  },
  {
    name: 'Windows Calendar',
    icon: '',
    matchers: [
      'microsoft.windowscommunicationsapps_8wekyb3d8bbwe!microsoft.windowslive.calendar'
    ],
    category: 'productivity'
  },

  // --- GAMING LAUNCHERS (Download, Friend & Achievement Toasts) ---
  {
    name: 'Steam',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/steam/default.svg',
    matchers: ['valve.steam.client', 'steam', 'steam.exe', 'valve.steam'],
    category: 'gaming'
  },
  {
    name: 'Epic Games Launcher',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/epic-games/default.svg',
    matchers: [
      'com.epicgames.epicgameslauncher',
      'epicgameslauncher',
      'epicgameslauncher.exe'
    ],
    category: 'gaming'
  },
  {
    name: 'Xbox App',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/xbox/default.svg',
    matchers: ['microsoft.gamingapp_8wekyb3d8bbwe!microsoft.xbox.app', 'xbox'],
    category: 'gaming'
  },
  {
    name: 'Battle.net',
    icon: '',
    matchers: [
      'blizzard entertainment.battle.net',
      'battle.net',
      'battle.net.exe'
    ],
    category: 'gaming'
  }
]

export const gameIcons: AppBrand[] = [
  // --- PC GAMES WITH ACTIVE OS TOAST NOTIFICATIONS ---
  {
    name: 'Rise of Kingdoms',
    icon: 'https://static.wikia.nocookie.net/riseofcivilizations/images/e/e6/Site-logo.png',
    matchers: [
      'rise of kingdoms',
      'com.lilithgame.roc.gp',
      'riseofkingdoms.exe',
      'rok'
    ],
    category: 'gaming'
  },
  {
    name: 'Clash of Clans',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/clash-of-clans/default.svg',
    matchers: [
      'com.supercell.clashofclans',
      'clash of clans',
      'clashofclans.exe',
      'googleplaygames.exe'
    ],
    category: 'gaming'
  },
  {
    name: 'Roblox',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/roblox/default.svg',
    matchers: ['roblox.corporation.roblox', 'robloxplayerbeta.exe', 'roblox'],
    category: 'gaming'
  },
  {
    name: 'EVE Online',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/eve-online/default.svg',
    matchers: ['eve online', 'exefile.exe', 'evelauncher.exe', 'ccp games'],
    category: 'gaming'
  },
  {
    name: 'Warframe',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/warframe/default.svg',
    matchers: ['warframe.x64.exe', 'warframe', 'digitalextremes.warframe'],
    category: 'gaming'
  },
  {
    name: 'Albion Online',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/albion-online/default.svg',
    matchers: ['albion online', 'albion-online.exe', 'sandbox interactive'],
    category: 'gaming'
  },
  {
    name: 'Magic: The Gathering Arena',
    icon: 'https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/magic-the-gathering/default.svg',
    matchers: [
      'mtga.exe',
      'magic the gathering arena',
      'mtg arena',
      'wizards of the coast.mtga'
    ],
    category: 'gaming'
  }
]

import { CacheIcons } from '../wailsjs/go/backend/Cache.js'

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
          if (matcher.includes(prevMatcher)) {
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
