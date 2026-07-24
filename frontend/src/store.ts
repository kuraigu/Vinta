import { writable, get } from 'svelte/store'

export const devTestingStore = writable<boolean>(false)

function persistentWritable<T>(key: string, initialValue: T) {
  const stored = localStorage.getItem(key)
  const initial = stored ? JSON.parse(stored) : initialValue
  const store = writable<T>(initial)
  store.subscribe((value) => {
    // Only persist to localStorage if NOT currently in DevTest testing mode
    if (!get(devTestingStore)) {
      localStorage.setItem(key, JSON.stringify(value))
    }
  })
  return store
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
export const autoHideStore = persistentWritable<boolean>(
  'vinta_autohide',
  false
)
export const wakeMethodStore = persistentWritable<'hover' | 'click'>(
  'vinta_wakemethod',
  'hover'
)
export const dndStore = persistentWritable<boolean>('vinta_dnd', false)

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
  { name: 'Manila', lat: 14.5995, lon: 120.9842 }
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
