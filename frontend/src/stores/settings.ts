import { writable, get } from 'svelte/store'
import {
  GetSettings,
  SaveSettings,
  SetAutoHideBorderless
} from '../../bindings/Vinta/backend/app.js'
import { devTestingStore } from './appState'

// ============================================================================
// Types
// ============================================================================
export type AppTheme = 'dark' | 'light'
export type GlowIntensity = 'disabled' | 'subtle' | 'vibrant'
export type AutoHideDisplayWidget = 'dash' | 'clock'
export type HiddenPillPosition = 'center' | 'top-left' | 'top-right'
export type TimeFormatOption = '12h' | '24h'
export type DateFormatOption = 'default' | 'iso' | 'us' | 'eu'

export type IslandConfig = {
  minimal: { w: number; h: number; r: number }
  expanded: { w: number; h: number; r: number }
  borderColor?: string
  glowColor?: string
}

export type WeatherLocation = {
  name: string
  lat: number
  lon: number
}

export type TemperatureUnit = 'celsius' | 'fahrenheit' | 'kelvin'

export interface HomeWidgetsVisibility {
  clock: boolean
  calendar: boolean
  weather: boolean
  battery: boolean
  wireless: boolean
}

export const defaultHomeWidgetsVisibility: HomeWidgetsVisibility = {
  clock: true,
  calendar: true,
  weather: true,
  battery: true,
  wireless: true
}

export interface PageVisibility {
  notes: boolean
  media: boolean
  notifications: boolean
  actioncenter: boolean
  calculator: boolean
  focus?: boolean
  timer: boolean
}

export const defaultPageVisibility: PageVisibility = {
  notes: true,
  media: true,
  notifications: true,
  actioncenter: true,
  calculator: true,
  focus: true,
  timer: true
}

// ============================================================================
// Persistence Engine
// ============================================================================
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

const registeredStores: Array<{
  key: string
  store: { set: (v: unknown) => void }
  initialValue: unknown
}> = []

function persistentWritable<T>(key: string, initialValue: T) {
  const stored = localStorage.getItem(key)
  let initial: T = initialValue
  if (stored) {
    try {
      const parsed = JSON.parse(stored)
      if (
        typeof initialValue === 'object' &&
        initialValue !== null &&
        !Array.isArray(initialValue) &&
        typeof parsed === 'object' &&
        parsed !== null &&
        !Array.isArray(parsed)
      ) {
        initial = { ...initialValue, ...parsed }
      } else {
        initial = parsed
      }
    } catch {
      initial = initialValue
    }
  }
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

// Load disk settings on startup and synchronize stores
if (typeof window !== 'undefined') {
  setTimeout(async () => {
    try {
      if (typeof GetSettings === 'function') {
        const json = await GetSettings()
        if (json && json !== '{}') {
          const parsed = JSON.parse(json)
          for (const item of registeredStores) {
            if (item.key in parsed) {
              let diskVal = parsed[item.key]
              if (
                typeof item.initialValue === 'object' &&
                item.initialValue !== null &&
                !Array.isArray(item.initialValue) &&
                typeof diskVal === 'object' &&
                diskVal !== null &&
                !Array.isArray(diskVal)
              ) {
                diskVal = {
                  ...(item.initialValue as Record<string, unknown>),
                  ...diskVal
                }
              }
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

export function getInitialTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Manila'
  } catch {
    return 'Asia/Manila'
  }
}

// ============================================================================
// 1. Appearance & Customization
// ============================================================================
export const themeStore = persistentWritable<AppTheme>('vinta_theme', 'dark')
export const userNameStore = persistentWritable<string>('vinta_user_name', 'User')
export const glowIntensityStore = persistentWritable<GlowIntensity>('vinta_glow_intensity', 'subtle')
export const opacityMinimalStore = persistentWritable<number>('vinta_opacity_minimal', 1.0)
export const opacityExpandedStore = persistentWritable<number>('vinta_opacity_expanded', 1.0)
export const opacityHiddenStore = persistentWritable<number>('vinta_opacity_hidden', 1.0)

// ============================================================================
// 2. Behavior & Auto-Hide
// ============================================================================
export const autoHideStore = persistentWritable<boolean>('vinta_autohide', true)
export const homeAutoHideStore = persistentWritable<boolean>('vinta_home_autohide', true)
export const mediaAutoHideStore = persistentWritable<boolean>('vinta_media_autohide', true)
export const autoHideDelayStore = persistentWritable<number>('vinta_autohide_delay', 10)
export const autoHideDisplayStore = persistentWritable<AutoHideDisplayWidget>('vinta_autohide_display', 'dash')
export const hiddenPillPositionStore = persistentWritable<HiddenPillPosition>('vinta_hidden_pill_position', 'center')
export const wakeMethodStore = persistentWritable<'hover' | 'click'>('vinta_wakemethod', 'hover')
export const allowDragPositionStore = persistentWritable<boolean>('vinta_allow_drag_position', true)
export const autoHideBorderlessStore = persistentWritable<boolean>('vinta_autohide_borderless', false)

// Synchronize borderless auto-hide setting to native backend hooks
autoHideBorderlessStore.subscribe((val) => {
  if (typeof SetAutoHideBorderless === 'function') {
    SetAutoHideBorderless(val).catch(() => {})
  }
})

// ============================================================================
// 3. Notifications & Do Not Disturb
// ============================================================================
export const dndStore = persistentWritable<boolean>('vinta_dnd', false)
export const notifAutoWakeStore = persistentWritable<boolean>('vinta_notif_autowake', true)
export const mediaAutoWakeStore = persistentWritable<boolean>('vinta_media_autowake', true)
export const notifSoundStore = persistentWritable<boolean>('vinta_notif_sound', true)
export const notifVolumeStore = persistentWritable<number>('vinta_notif_volume', 0.8)

// ============================================================================
// 4. Time, Date & Localization
// ============================================================================
export const timeFormatStore = persistentWritable<TimeFormatOption>('vinta_time_format', '12h')
export const timezoneStore = persistentWritable<string>('vinta_timezone', getInitialTimezone())
export const dateFormatStore = persistentWritable<DateFormatOption>('vinta_date_format', 'default')

// ============================================================================
// 5. Weather
// ============================================================================
export const weatherLocationStore = persistentWritable<WeatherLocation>('vinta_weather_location', { name: 'Manila, Philippines', lat: 14.5995, lon: 120.9842 })
export const weatherUnitStore = persistentWritable<TemperatureUnit>('vinta_weather_unit', 'celsius')

// ============================================================================
// 6. Home & Page Visibility
// ============================================================================
export const homeWidgetsVisibilityStore = persistentWritable<HomeWidgetsVisibility>('vinta_home_widgets_visibility', defaultHomeWidgetsVisibility)
export const pageVisibilityStore = persistentWritable<PageVisibility>('vinta_page_visibility', defaultPageVisibility)

// ============================================================================
// 7. System Integration & Updates
// ============================================================================
export const autostartStore = persistentWritable<boolean>('vinta_autostart', false)
export const hardwareAccelStore = persistentWritable<boolean>('vinta_hardware_accel', true)
export const autoCheckUpdatesStore = persistentWritable<boolean>('autoCheckUpdates', true)
export type UpdateChannel = 'stable' | 'prerelease' | 'testing'
export const updateChannelStore = persistentWritable<UpdateChannel>('vinta_update_channel', 'stable')
export const customUpdateRepoStore = persistentWritable<string>('vinta_custom_update_repo', '')

// ============================================================================
// Helpers
// ============================================================================
export function calculateHomeMinimalWidth(
  visibility: HomeWidgetsVisibility,
  hasBattery: boolean = true,
  hasWireless: boolean = true
): number {
  const isClockActive = visibility.clock
  const isBatteryActive = visibility.battery && hasBattery
  const isWirelessActive = visibility.wireless && hasWireless
  const isWeatherActive = visibility.weather
  const isCalendarActive = visibility.calendar

  const activeSlots: number[] = []

  // Slot 1: Device Status Cluster (Wireless + Battery + Clock)
  let statusClusterWidth = 0
  if (isWirelessActive) statusClusterWidth += 15
  if (isBatteryActive)
    statusClusterWidth += (statusClusterWidth > 0 ? 3 : 0) + 25
  if (isClockActive) statusClusterWidth += (statusClusterWidth > 0 ? 3 : 0) + 82

  if (statusClusterWidth > 0) {
    activeSlots.push(statusClusterWidth)
  }

  // Slot 2: Weather
  if (isWeatherActive) {
    activeSlots.push(72)
  }

  // Slot 3: Calendar
  if (isCalendarActive) {
    activeSlots.push(102)
  }

  if (activeSlots.length === 0) return 130

  const PADDING = 36 // 18px left + 18px right
  const GAP = 18 // Gap between primary slots
  const contentW = activeSlots.reduce((sum, w) => sum + w, 0)
  const gapsW = Math.max(0, activeSlots.length - 1) * GAP
  return Math.max(130, PADDING + contentW + gapsW)
}
