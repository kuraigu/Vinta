import { writable, get } from 'svelte/store'
import { GetSettings, SaveSettings } from '../wailsjs/go/backend/App.js'

export const devTestingStore = writable<boolean>(false)
export const activeSettingsTabStore = writable<string>('appearance')

const allSettingsCache: Record<string, unknown> = {}
let saveSettingsTimeout: ReturnType<typeof setTimeout> | null = null

function syncSettingsToDisk() {
  if (saveSettingsTimeout) clearTimeout(saveSettingsTimeout)
  saveSettingsTimeout = setTimeout(() => {
    try {
      if (typeof SaveSettings === 'function') {
        SaveSettings(JSON.stringify(allSettingsCache)).catch(() => {})
      }
    } catch {
      // Ignore disk write failure
    }
  }, 300)
}

// Registry of stores for initial loading from disk
const registeredStores: Array<{
  key: string
  store: { set: (v: unknown) => void }
  initialValue: unknown
}> = []

function persistentWritable<T>(key: string, initialValue: T) {
  const stored = localStorage.getItem(key)
  const initial = stored ? JSON.parse(stored) : initialValue
  allSettingsCache[key] = initial
  const store = writable<T>(initial)

  registeredStores.push({
    key,
    store: store as unknown as { set: (v: unknown) => void },
    initialValue
  })

  store.subscribe((value) => {
    allSettingsCache[key] = value
    if (!get(devTestingStore)) {
      try {
        localStorage.setItem(key, JSON.stringify(value))
      } catch {
        // Ignore storage write error
      }
      syncSettingsToDisk()
    }
  })
  return store
}

// Load disk settings on startup and update stores
if (typeof window !== 'undefined') {
  setTimeout(async () => {
    try {
      if (typeof GetSettings === 'function') {
        const json = await GetSettings()
        if (json && json !== '{}') {
          const parsed = JSON.parse(json)
          for (const item of registeredStores) {
            if (item.key in parsed) {
              const diskVal = parsed[item.key]
              localStorage.setItem(item.key, JSON.stringify(diskVal))
              item.store.set(diskVal)
            }
          }
        }
      }
    } catch {
      // Ignore disk load error
    }
  }, 0)
}

export type AppTheme = 'dark' | 'light'
export type IslandShape = 'pill' | 'box'
export type IslandConfig = {
  minimal: { w: number; h: number; r: number }
  expanded: { w: number; h: number; r: number }
  borderColor?: string
  glowColor?: string
}

export const themeStore = persistentWritable<AppTheme>('vinta_theme', 'dark')
export const shapeStore = persistentWritable<IslandShape>('vinta_shape', 'pill')
export const opacityMinimalStore = persistentWritable<number>(
  'vinta_opacity_minimal',
  1.0
)
export const opacityExpandedStore = persistentWritable<number>(
  'vinta_opacity_expanded',
  1.0
)
export const opacityHiddenStore = persistentWritable<number>(
  'vinta_opacity_hidden',
  1.0
)
export const hiddenPillOffsetXStore = persistentWritable<number>(
  'vinta_hidden_pill_offset_x',
  0
)
export const allowDragPositionStore = persistentWritable<boolean>(
  'vinta_allow_drag_position',
  true
)
export const hardwareAccelStore = persistentWritable<boolean>(
  'vinta_hardware_accel',
  true
)
export const notifSoundStore = persistentWritable<boolean>(
  'vinta_notif_sound',
  true
)
export const notifVolumeStore = persistentWritable<number>(
  'vinta_notif_volume',
  0.8
)
export type GlowIntensity = 'disabled' | 'subtle' | 'vibrant'
export const glowIntensityStore = persistentWritable<GlowIntensity>(
  'vinta_glow_intensity',
  'subtle'
)

export const autostartStore = persistentWritable<boolean>(
  'vinta_autostart',
  false
)
export const autoHideStore = persistentWritable<boolean>('vinta_autohide', true)
export const homeAutoHideStore = persistentWritable<boolean>(
  'vinta_home_autohide',
  true
)
export const mediaAutoHideStore = persistentWritable<boolean>(
  'vinta_media_autohide',
  true
)
export const notifAutoHideStore = persistentWritable<boolean>(
  'vinta_notif_autohide',
  false
)
export const autoHideDelayStore = persistentWritable<number>(
  'vinta_autohide_delay',
  10
)

export type AutoHideDisplayWidget = 'dash' | 'clock'
export const autoHideDisplayStore = persistentWritable<AutoHideDisplayWidget>(
  'vinta_autohide_display',
  'dash'
)

export type HiddenPillPosition = 'center' | 'top-left' | 'top-right'
export const hiddenPillPositionStore = persistentWritable<HiddenPillPosition>(
  'vinta_hidden_pill_position',
  'center'
)

export const wakeMethodStore = persistentWritable<'hover' | 'click'>(
  'vinta_wakemethod',
  'hover'
)
export const dndStore = persistentWritable<boolean>('vinta_dnd', false)
export const notifAutoWakeStore = persistentWritable<boolean>(
  'vinta_notif_autowake',
  true
)
export const mediaAutoWakeStore = persistentWritable<boolean>(
  'vinta_media_autowake',
  true
)

export type TimeFormatOption = '12h' | '24h'
export type DateFormatOption = 'default' | 'iso' | 'us' | 'eu'

export const timeFormatStore = persistentWritable<TimeFormatOption>(
  'vinta_time_format',
  '12h'
)

export function getInitialTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Manila'
  } catch {
    return 'Asia/Manila'
  }
}

// First-time launch auto-detects system timezone with Asia/Manila fallback.
export const timezoneStore = persistentWritable<string>(
  'vinta_timezone',
  getInitialTimezone()
)

export const dateFormatStore = persistentWritable<DateFormatOption>(
  'vinta_date_format',
  'default'
)

export type WeatherLocation = {
  name: string
  lat: number
  lon: number
}

export type TemperatureUnit = 'celsius' | 'fahrenheit' | 'kelvin'

export const weatherLocationStore = persistentWritable<WeatherLocation>(
  'vinta_weather_location',
  { name: 'Manila, Philippines', lat: 14.5995, lon: 120.9842 }
)

export const weatherUnitStore = persistentWritable<TemperatureUnit>(
  'vinta_weather_unit',
  'celsius'
)

export const simulateOfflineStore = writable<boolean>(false)

export type AppNotification = {
  id: number
  appName?: string
  icon?: string
  title: string
  message: string
  heroImage?: string
  timestamp: number
}

export const notifData = writable<AppNotification[]>([])
export const notificationPermissionDenied = writable<boolean>(false)

export interface MediaData {
  isAvailable: boolean
  sourceApp?: string
  appName?: string
  appIcon?: string
  title: string
  artist: string
  album: string
  albumArt: string
  isPlaying: boolean
  audioPeak?: number
  positionMs: number
  durationMs: number
  volume?: number
  isMuted?: boolean
  canGoNext: boolean
  canGoPrev: boolean
  isShuffleActive?: boolean
  autoRepeatMode?: number // 0: None, 1: Track, 2: List
  isShuffleEnabled?: boolean
  isRepeatEnabled?: boolean
}

export const globalMediaStore = writable<MediaData>({
  isAvailable: false,
  sourceApp: '',
  appName: 'Media Player',
  appIcon: '',
  title: '',
  artist: '',
  album: '',
  albumArt: '',
  isPlaying: false,
  audioPeak: 0,
  positionMs: 0,
  durationMs: 0,
  volume: 50,
  isMuted: false,
  canGoNext: true,
  canGoPrev: true,
  isShuffleActive: false,
  autoRepeatMode: 0
})

export const audioPeakStore = writable<number>(0)
