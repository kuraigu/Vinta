import { writable, get } from 'svelte/store'
import { Events } from '@wailsio/runtime'
import {
  GetSettings,
  SaveSettings,
  SetAutoHideBorderless,
  ExportSettingsFileDialog,
  ImportSettingsFileDialog,
  ApplyImportedSettings
} from '../../bindings/Vinta/backend/app.js'
import { devTestingStore } from './appState'

// ============================================================================
// Types
// ============================================================================
export type AppTheme = string
export type GlowIntensity = 'disabled' | 'subtle' | 'vibrant'
export type AutoHideDisplayWidget = 'dash' | 'clock'
export type HiddenPillPosition = 'center' | 'top-left' | 'top-right'
export type TimeFormatOption = '12h' | '24h'
export type { DateFormatOption } from '../utils/dateFormatter'
import type { DateFormatOption } from '../utils/dateFormatter'

export type IslandConfig = {
  minimal: { w: number; h: number; r: number }
  expanded: {
    w: number
    h: number
    r: number
    minH?: number
    maxH?: number
    minW?: number
    maxW?: number
  }
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
  actionCenter: boolean
}

export const defaultHomeWidgetsVisibility: HomeWidgetsVisibility = {
  clock: true,
  calendar: true,
  weather: true,
  battery: true,
  wireless: true,
  actionCenter: true
}

export interface PageVisibility {
  notes: boolean
  media: boolean
  notifications: boolean
  actioncenter: boolean
  calculator: boolean
  focus?: boolean
  timer: boolean
  calendar: boolean
  weather: boolean
}

export const defaultPageVisibility: PageVisibility = {
  notes: true,
  media: true,
  notifications: true,
  actioncenter: true,
  calculator: true,
  focus: true,
  timer: true,
  calendar: true,
  weather: true
}

export interface DockPinnedApps {
  calendar: boolean
  notes: boolean
  calculator: boolean
  timer: boolean
  weather: boolean
}

export const defaultDockPinnedApps: DockPinnedApps = {
  calendar: true,
  notes: true,
  calculator: true,
  timer: false,
  weather: false
}

// ============================================================================
// Persistence Engine (Sparse Schema with Fallback Defaults)
// ============================================================================
const userOverridesCache: Record<string, unknown> = {}
let saveSettingsTimeout: ReturnType<typeof setTimeout> | null = null
let isInitialDiskLoadComplete = false

/**
 * Performs a deep equality comparison between two arbitrary values.
 *
 * @param a - First value to compare.
 * @param b - Second value to compare.
 * @returns True if both structures are deeply identical.
 */
function isDeepEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true
  if (
    typeof a !== 'object' ||
    a === null ||
    typeof b !== 'object' ||
    b === null
  ) {
    return false
  }
  const keysA = Object.keys(a as Record<string, unknown>)
  const keysB = Object.keys(b as Record<string, unknown>)
  if (keysA.length !== keysB.length) return false
  for (const k of keysA) {
    if (!Object.prototype.hasOwnProperty.call(b, k)) return false
    if (
      !isDeepEqual(
        (a as Record<string, unknown>)[k],
        (b as Record<string, unknown>)[k]
      )
    ) {
      return false
    }
  }
  return true
}

/**
 * Flushes active user overrides to the disk settings.json file.
 */
function syncSettingsToDisk() {
  if (saveSettingsTimeout) clearTimeout(saveSettingsTimeout)
  saveSettingsTimeout = setTimeout(() => {
    try {
      if (typeof SaveSettings === 'function') {
        SaveSettings(JSON.stringify(userOverridesCache)).catch(() => {})
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

/**
 * Creates a reactive writable store synchronized with localStorage and disk storage.
 * In accordance with the sparse configuration model, only user overrides differing from
 * default values are persisted to disk.
 *
 * @param key - Unique settings identifier.
 * @param initialValue - Default fallback value if unconfigured.
 * @returns Configured Svelte writable store.
 */
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
      if (key === 'vinta_theme' && (initial === 'dark' || initial === 'oled-dark')) {
        initial = 'dark-gold' as unknown as T
      }
    } catch {
      initial = initialValue
    }
  }

  if (key === 'vinta_theme' && (initial === 'dark' || initial === 'oled-dark')) {
    initial = 'dark-gold' as unknown as T
  }

  const store = writable<T>(initial)

  registeredStores.push({
    key,
    store: store as unknown as { set: (v: unknown) => void },
    initialValue
  })

  store.subscribe((value) => {
    if (!get(devTestingStore)) {
      try {
        localStorage.setItem(key, JSON.stringify(value))
      } catch {
        // Ignore storage write error
      }

      if (isInitialDiskLoadComplete) {
        if (!isDeepEqual(value, initialValue)) {
          userOverridesCache[key] = value
        } else {
          delete userOverridesCache[key]
        }
        userOverridesCache['_sparse_migrated'] = true
        syncSettingsToDisk()
      }
    }
  })
  return store
}

// Known legacy defaults to prune during one-time sparse migration
const legacyV1Defaults: Record<string, unknown> = {
  vinta_autohide_display: 'dash',
  vinta_theme: 'dark-gold',
  vinta_user_name: 'User',
  vinta_glow_intensity: 'subtle',
  vinta_opacity_minimal: 1.0,
  vinta_opacity_expanded: 1.0,
  vinta_opacity_hidden: 1.0,
  vinta_autohide: true,
  vinta_home_autohide: true,
  vinta_media_autohide: true,
  vinta_autohide_delay: 10,
  vinta_hidden_pill_position: 'center',
  vinta_wakemethod: 'click',
  vinta_allow_drag_position: true,
  vinta_autohide_borderless: false,
  vinta_dnd: false,
  vinta_notif_autowake: true,
  vinta_media_autowake: true,
  vinta_notif_sound: true,
  vinta_notif_volume: 0.8,
  vinta_time_format: '12h',
  vinta_date_format: 'default',
  vinta_weather_unit: 'celsius',
  vinta_autostart: false,
  vinta_hardware_accel: true,
  autoCheckUpdates: true,
  vinta_update_channel: 'stable',
  vinta_custom_update_repo: '',
  vinta_ui_scale: 1.0,
  vinta_auto_discard_past_events: false
}

// Load disk settings on startup, apply sparse migration if needed, and synchronize stores
if (typeof window !== 'undefined') {
  setTimeout(async () => {
    try {
      if (typeof GetSettings === 'function') {
        const json = await GetSettings()
        if (json && json !== '{}') {
          const parsed: Record<string, unknown> = JSON.parse(json)
          let shouldResavePruned = false

          if (!parsed['_sparse_migrated']) {
            parsed['_sparse_migrated'] = true
            shouldResavePruned = true

            for (const [k, legacyVal] of Object.entries(legacyV1Defaults)) {
              if (k in parsed && isDeepEqual(parsed[k], legacyVal)) {
                delete parsed[k]
              }
            }
          }

          for (const [k, v] of Object.entries(parsed)) {
            userOverridesCache[k] = v
          }

          if (shouldResavePruned) {
            syncSettingsToDisk()
          }

          for (const item of registeredStores) {
            if (item.key in parsed && parsed[item.key] !== undefined) {
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
              if (item.key === 'vinta_theme' && (diskVal === 'dark' || diskVal === 'oled-dark')) {
                diskVal = 'dark-gold'
              }
              localStorage.setItem(item.key, JSON.stringify(diskVal))
              item.store.set(diskVal)
            } else {
              localStorage.removeItem(item.key)
              item.store.set(item.initialValue)
            }
          }
        }
      }
    } catch {
      // Ignore disk load error
    } finally {
      isInitialDiskLoadComplete = true
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
export const themeStore = persistentWritable<string>('vinta_theme', 'dark-gold')
export const userNameStore = persistentWritable<string>(
  'vinta_user_name',
  'User'
)
export const glowIntensityStore = persistentWritable<GlowIntensity>(
  'vinta_glow_intensity',
  'subtle'
)
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
export const uiScaleStore = persistentWritable<number>('vinta_ui_scale', 1.0)

// ============================================================================
// 2. Behavior & Auto-Hide
// ============================================================================
export const autoHideStore = persistentWritable<boolean>(
  'vinta_autohide',
  true
)
export const homeAutoHideStore = persistentWritable<boolean>(
  'vinta_home_autohide',
  true
)
export const mediaAutoHideStore = persistentWritable<boolean>(
  'vinta_media_autohide',
  true
)
export const autoHideDelayStore = persistentWritable<number>(
  'vinta_autohide_delay',
  10
)
export const autoHideDisplayStore =
  persistentWritable<AutoHideDisplayWidget>('vinta_autohide_display', 'dash')
export const hiddenPillPositionStore = persistentWritable<HiddenPillPosition>('vinta_hidden_pill_position', 'center')
export const wakeMethodStore = persistentWritable<'hover' | 'click'>('vinta_wakemethod', 'click')
export const allowDragPositionStore = persistentWritable<boolean>('vinta_allow_drag_position', true)
export const autoHideBorderlessStore = persistentWritable<boolean>('vinta_autohide_borderless', false)
export const followActiveDisplayStore = persistentWritable<boolean>('vinta_follow_active_display', false)

// Synchronize borderless and active display settings to native backend hooks
autoHideBorderlessStore.subscribe((val) => {
  if (typeof SetAutoHideBorderless === 'function') {
    SetAutoHideBorderless(val).catch(() => {})
  }
  try {
    Events.Emit('set-autohide-borderless', val)
  } catch {
    // Ignore runtime event error
  }
})

followActiveDisplayStore.subscribe((val) => {
  try {
    Events.Emit('set-follow-active-display', val)
  } catch {
    // Ignore runtime event error
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

// Alarms & Reminders Audio Settings
export type { AlarmSoundPreset } from '../services/audioService'
import type { AlarmSoundPreset } from '../services/audioService'
export const alarmSoundPresetStore = persistentWritable<AlarmSoundPreset>('vinta_alarm_sound_preset', 'chime')
export const alarmVolumeStore = persistentWritable<number>('vinta_alarm_volume', 0.85)
export const alarmSnoozeMinutesStore = persistentWritable<number>('vinta_alarm_snooze_minutes', 5)
export const reminderSoundPresetStore = persistentWritable<AlarmSoundPreset>('vinta_reminder_sound_preset', 'chime')

// ============================================================================
// 4. Time, Date & Localization
// ============================================================================
export const timeFormatStore = persistentWritable<TimeFormatOption>('vinta_time_format', '12h')
export const timezoneStore = persistentWritable<string>('vinta_timezone', getInitialTimezone())
export const dateFormatStore = persistentWritable<DateFormatOption>('vinta_date_format', 'short')
export const autoDiscardPastEventsStore = persistentWritable<boolean>('vinta_auto_discard_past_events', false)

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
export const dockPinnedAppsStore = persistentWritable<DockPinnedApps>('vinta_dock_pinned_apps', defaultDockPinnedApps)
export const pinnedMediaSourceStore = persistentWritable<string | null>('vinta_pinned_media_source', null)
export const globalHotkeyStore = persistentWritable<string>('vinta_global_hotkey', 'Alt+Shift+V')
export const quickNotesHotkeyStore = persistentWritable<string>('vinta_hotkey_quick_notes', 'Alt+Shift+N')
export const modeSwitchHotkeyStore = persistentWritable<string>('vinta_hotkey_mode_switch', 'Alt+Shift+D')

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

/**
 * Calculates the dynamic height for the expanded Home Page based on active glance cards,
 * clamped between minimum and maximum bounds defined in the island configuration.
 */
export function calculateHomeExpandedHeight(
  visibility: HomeWidgetsVisibility,
  hasNotification: boolean = false,
  hasUpdate: boolean = false,
  minH: number = 260,
  maxH: number = 640
): number {
  const GAP = 12

  // 1. Header segment (Weather hero full-bleed or salutation greeting)
  let totalHeight = visibility.weather ? 178 : 96

  // 2. Update notice banner
  if (hasUpdate) {
    totalHeight += GAP + 68
  }

  // 3. Notification glance card (only renders if unread notifications exist)
  if (hasNotification) {
    totalHeight += GAP + 88
  }

  // 4. Calendar glance agenda
  if (visibility.calendar) {
    totalHeight += GAP + 104
  }

  // 5. Quick Controls / Action Center glance
  if (visibility.actionCenter) {
    totalHeight += GAP + 110
  }

  // 6. Studio Branding Footer ("Vinta by FreeMatrix Studios & Contributors")
  totalHeight += GAP + 36

  // 7. Bottom clearance for feed padding & island corner curvature
  const BOTTOM_CLEARANCE = 24

  const computedHeight = totalHeight + BOTTOM_CLEARANCE
  return Math.max(minH, Math.min(maxH, computedHeight))
}

// ============================================================================
// 8. Configuration Export, Import & Schema Sanitization
// ============================================================================

export interface SettingsExportResult {
  success: boolean
  path?: string
  error?: string
}

export interface SettingsImportResult {
  success: boolean
  count?: number
  error?: string
}

/**
 * Validates and sanitizes an imported configuration object.
 * Clamps numeric boundaries, enforces strict enum literals, and strips unrecognized fields.
 *
 * @param raw - Raw unverified settings payload.
 * @returns Cleaned and sanitized settings object.
 */
function sanitizeSettingsPayload(raw: Record<string, unknown>): Record<string, unknown> {
  const sanitized: Record<string, unknown> = {}

  for (const reg of registeredStores) {
    const key = reg.key
    if (key in raw) {
      const val = raw[key]
      const defaultVal = reg.initialValue

      if (key === 'vinta_pinned_media_source') {
        sanitized[key] = typeof val === 'string' && val.trim() ? val.trim() : null
        continue
      }

      if (key === 'vinta_global_hotkey') {
        sanitized[key] = typeof val === 'string' && val.trim() ? val.trim() : 'Alt+Shift+V'
        continue
      }

      if (key === 'vinta_hotkey_quick_notes') {
        sanitized[key] = typeof val === 'string' && val.trim() ? val.trim() : 'Alt+Shift+N'
        continue
      }

      if (key === 'vinta_hotkey_mode_switch') {
        sanitized[key] = typeof val === 'string' && val.trim() ? val.trim() : 'Alt+Shift+D'
        continue
      }

      if (key === 'vinta_ui_scale') {
        const num = typeof val === 'number' ? val : parseFloat(val as string)
        sanitized[key] = isNaN(num) ? 1.0 : Math.max(1.0, Math.min(1.25, num))
        continue
      }

      if (key === 'vinta_auto_discard_past_events') {
        sanitized[key] = Boolean(val)
        continue
      }

      if (typeof val === typeof defaultVal && val !== null) {
        if (typeof defaultVal === 'number') {
          if (key === 'vinta_autohide_delay') {
            sanitized[key] = Math.max(3, Math.min(60, Number(val)))
          } else if (key.startsWith('vinta_opacity_')) {
            sanitized[key] = Math.max(0.1, Math.min(1.0, Number(val)))
          } else {
            sanitized[key] = Number(val)
          }
        } else if (typeof defaultVal === 'string') {
          if (key === 'vinta_theme' && (val === 'dark' || val === 'light')) {
            sanitized[key] = val
          } else if (
            key === 'vinta_glow_intensity' &&
            (val === 'disabled' || val === 'subtle' || val === 'vibrant')
          ) {
            sanitized[key] = val
          } else if (key === 'vinta_autohide_display' && (val === 'dash' || val === 'clock')) {
            sanitized[key] = val
          } else if (
            key === 'vinta_hidden_pill_position' &&
            (val === 'center' || val === 'top-left' || val === 'top-right')
          ) {
            sanitized[key] = val
          } else if (key === 'vinta_wakemethod' && (val === 'hover' || val === 'click')) {
            sanitized[key] = val
          } else if (key === 'vinta_time_format' && (val === '12h' || val === '24h')) {
            sanitized[key] = val
          } else if (
            key === 'vinta_date_format' &&
            (val === 'default' || val === 'iso' || val === 'us' || val === 'eu')
          ) {
            sanitized[key] = val
          } else if (
            key === 'vinta_weather_unit' &&
            (val === 'celsius' || val === 'fahrenheit' || val === 'kelvin')
          ) {
            sanitized[key] = val
          } else if (
            key === 'vinta_update_channel' &&
            (val === 'stable' || val === 'prerelease' || val === 'testing')
          ) {
            sanitized[key] = val
          } else {
            sanitized[key] = String(val)
          }
        } else if (typeof defaultVal === 'boolean') {
          sanitized[key] = Boolean(val)
        } else if (
          typeof defaultVal === 'object' &&
          !Array.isArray(defaultVal) &&
          typeof val === 'object' &&
          !Array.isArray(val)
        ) {
          const mergedObj: Record<string, unknown> = { ...(defaultVal as Record<string, unknown>) }
          const valObj = val as Record<string, unknown>
          for (const subKey of Object.keys(mergedObj)) {
            if (subKey in valObj && typeof valObj[subKey] === typeof mergedObj[subKey]) {
              mergedObj[subKey] = valObj[subKey]
            }
          }
          sanitized[key] = mergedObj
        }
      }
    }
  }

  sanitized['_sparse_migrated'] = true
  return sanitized
}

/**
 * Prompts the user with a native file dialog to export their current settings to a JSON file.
 *
 * @returns Result object with status and chosen destination path.
 */
export async function exportSettingsToFile(): Promise<SettingsExportResult> {
  try {
    if (typeof ExportSettingsFileDialog !== 'function') {
      return { success: false, error: 'Export function not available' }
    }
    const targetPath = await ExportSettingsFileDialog('vinta-settings.json')
    if (!targetPath) {
      return { success: false }
    }
    return { success: true, path: targetPath }
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err)
    return { success: false, error: msg }
  }
}

/**
 * Prompts the user to pick a JSON configuration file, parses and deep-sanitizes its structure,
 * commits it atomically to disk, and updates reactive state in-place.
 *
 * @returns Result object with status and imported settings count.
 */
export async function importSettingsFromFile(): Promise<SettingsImportResult> {
  try {
    if (
      typeof ImportSettingsFileDialog !== 'function' ||
      typeof ApplyImportedSettings !== 'function'
    ) {
      return { success: false, error: 'Import function not available' }
    }

    const rawJSON = await ImportSettingsFileDialog()
    if (!rawJSON) {
      return { success: false }
    }

    let parsed: Record<string, unknown>
    try {
      parsed = JSON.parse(rawJSON)
    } catch {
      return { success: false, error: 'Selected file is not valid JSON' }
    }

    const settingsPayload =
      parsed.settings && typeof parsed.settings === 'object' && !Array.isArray(parsed.settings)
        ? (parsed.settings as Record<string, unknown>)
        : parsed

    const sanitized = sanitizeSettingsPayload(settingsPayload)
    const sanitizedKeys = Object.keys(sanitized).filter((k) => k !== '_sparse_migrated')

    if (sanitizedKeys.length === 0) {
      return { success: false, error: 'No recognizable settings found in file' }
    }

    await ApplyImportedSettings(JSON.stringify(sanitized))

    for (const key of Object.keys(userOverridesCache)) {
      delete userOverridesCache[key]
    }
    Object.assign(userOverridesCache, sanitized)

    for (const reg of registeredStores) {
      if (reg.key in sanitized) {
        reg.store.set(sanitized[reg.key])
      } else {
        reg.store.set(reg.initialValue)
      }
    }

    return { success: true, count: sanitizedKeys.length }
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err)
    return { success: false, error: msg }
  }
}

