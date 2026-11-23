/**
 * @fileoverview User settings store with encrypted disk persistence.
 *
 * Stores user preferences (theme, auto-hide, 12h/24h time, weather location, hotkeys)
 * and automatically saves changes to encrypted storage via the backend.
 */

import { writable, get } from 'svelte/store'
import { Events } from '$lib/events'
import {
  GetSettings,
  SaveSettings,
  SetAutoHideBorderless,
  SetRelocateOnFullscreen,
  SetReturnOnFullscreenExit,
  ExportSettingsFileDialog,
  ImportSettingsFileDialog,
  ApplyImportedSettings,
  RestoreLatestSettingsBackup
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
import {
  inferDefaultTemperatureUnit,
  inferDefaultTimeFormat,
  inferDefaultDateFormat
} from '../utils/localeInference'

/** Standardized expanded island width tiers (px) to prevent horizontal jitter across navigation. */
export const ISLAND_WIDTH_TIERS = {
  compact: 340,
  media: 400,
  standard: 440,
  wide: 600
} as const

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
  media?: boolean
  systemStats?: boolean
  spotlight?: boolean
  announcements?: boolean
  notifications?: boolean
}

export const defaultHomeWidgetsVisibility: HomeWidgetsVisibility = {
  clock: true,
  calendar: true,
  weather: true,
  battery: true,
  wireless: true,
  actionCenter: true,
  media: true,
  systemStats: false,
  spotlight: true,
  announcements: true,
  notifications: true
}

export type WidgetSizePreset = '1x1' | '2x1' | '2x2'

export const defaultHomeWidgetsOrder: string[] = [
  'media',
  'brief',
  'actionCenter',
  'systemStats'
]

export const defaultHomeWidgetSizes: Record<string, WidgetSizePreset> = {
  weather: '2x1',
  media: '2x1',
  announcements: '2x1',
  notifications: '2x1',
  calendar: '2x1',
  actionCenter: '2x1',
  systemStats: '2x1'
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
  fileshelf: boolean
  performance?: boolean
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
  weather: true,
  fileshelf: true,
  performance: true
}

export interface DockPinnedApps {
  calendar: boolean
  notes: boolean
  calculator: boolean
  timer: boolean
  weather: boolean
  fileshelf: boolean
}

export const defaultDockPinnedApps: DockPinnedApps = {
  calendar: true,
  notes: true,
  calculator: true,
  timer: false,
  weather: false,
  fileshelf: false
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
export function syncSettingsToDisk(immediate: boolean = false): Promise<void> {
  if (saveSettingsTimeout) clearTimeout(saveSettingsTimeout)
  if (immediate) {
    try {
      if (typeof SaveSettings === 'function') {
        return SaveSettings(JSON.stringify(userOverridesCache)).catch(() => {})
      }
    } catch {
      // Ignore disk write failure
    }
    return Promise.resolve()
  }
  saveSettingsTimeout = setTimeout(() => {
    try {
      if (typeof SaveSettings === 'function') {
        SaveSettings(JSON.stringify(userOverridesCache)).catch(() => {})
      }
    } catch {
      // Ignore disk write failure
    }
  }, 300)
  return Promise.resolve()
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

      if (!isDeepEqual(value, initialValue)) {
        userOverridesCache[key] = value
      } else {
        delete userOverridesCache[key]
      }
      userOverridesCache['_sparse_migrated'] = true
      userOverridesCache['_schema_version'] = CURRENT_SETTINGS_SCHEMA_VERSION

      if (isInitialDiskLoadComplete) {
        syncSettingsToDisk()
      }
    }
  })
  return {
    ...store,
    defaultValue: initialValue,
    reset: () => store.set(initialValue)
  }
}

export const CURRENT_SETTINGS_SCHEMA_VERSION = 2

export interface MigrationResult {
  migrated: Record<string, unknown>
  changed: boolean
  fromVersion: number
  toVersion: number
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
  vinta_notif_satellite_enabled: true,
  vinta_media_autowake: true,
  vinta_notif_sound: true,
  vinta_notif_volume: 0.8,
  vinta_time_format: '12h',
  vinta_date_format: 'default',
  vinta_weather_unit: 'celsius',
  vinta_weather_animations: true,
  vinta_autostart: false,
  vinta_hardware_accel: true,
  autoCheckUpdates: true,
  vinta_update_channel: 'stable',
  vinta_custom_update_repo: '',
  vinta_ui_scale: 1.0,
  vinta_auto_discard_past_events: false,
  vinta_locale: 'en'
}

type MigrationFn = (settings: Record<string, unknown>) => Record<string, unknown>

const migrations: Record<number, MigrationFn> = {
  1: (settings) => {
    const updated = { ...settings }
    if (!updated['_sparse_migrated']) {
      for (const [k, legacyVal] of Object.entries(legacyV1Defaults)) {
        if (k in updated && isDeepEqual(updated[k], legacyVal)) {
          delete updated[k]
        }
      }
      updated['_sparse_migrated'] = true
    }
    updated['_schema_version'] = 1
    return updated
  },
  2: (settings) => {
    const updated = { ...settings }
    if (updated['vinta_home_widget_menu_enabled'] === false) {
      delete updated['vinta_home_widget_menu_enabled']
    }
    updated['_schema_version'] = 2
    return updated
  }
}

/**
 * Executes pending schema migrations on a raw settings object up to CURRENT_SETTINGS_SCHEMA_VERSION.
 *
 * @param raw - Raw parsed settings configuration.
 * @param fallbackVersion - Optional version from an export envelope ($schemaVersion).
 * @returns Migration result containing the updated settings object and migration status.
 */
export function migrateSettings(
  raw: Record<string, unknown>,
  fallbackVersion?: number
): MigrationResult {
  let current = { ...raw }
  let fromVersion = 0

  if (typeof current['_schema_version'] === 'number') {
    fromVersion = current['_schema_version']
  } else if (typeof fallbackVersion === 'number') {
    fromVersion = fallbackVersion
  } else if (current['_sparse_migrated']) {
    fromVersion = 0
  }

  let changed = false
  for (let v = fromVersion + 1; v <= CURRENT_SETTINGS_SCHEMA_VERSION; v++) {
    if (migrations[v]) {
      current = migrations[v](current)
      changed = true
    }
  }

  if (current['_schema_version'] !== CURRENT_SETTINGS_SCHEMA_VERSION) {
    current['_schema_version'] = CURRENT_SETTINGS_SCHEMA_VERSION
    changed = true
  }
  if (!current['_sparse_migrated']) {
    current['_sparse_migrated'] = true
    changed = true
  }

  return {
    migrated: current,
    changed,
    fromVersion,
    toVersion: CURRENT_SETTINGS_SCHEMA_VERSION
  }
}

// Load disk settings on startup, apply schema migrations, and synchronize stores
if (typeof window !== 'undefined') {
  setTimeout(async () => {
    try {
      if (typeof GetSettings === 'function') {
        const json = await GetSettings()
        if (json && json !== '{}') {
          const parsed: Record<string, unknown> = JSON.parse(json)
          const migration = migrateSettings(parsed)

          for (const [k, v] of Object.entries(migration.migrated)) {
            userOverridesCache[k] = v
          }

          if (migration.changed) {
            syncSettingsToDisk()
          }

          for (const item of registeredStores) {
            if (item.key in migration.migrated && migration.migrated[item.key] !== undefined) {
              let diskVal = migration.migrated[item.key]
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

  const flushPendingSettings = () => {
    if (isInitialDiskLoadComplete && !get(devTestingStore)) {
      syncSettingsToDisk(true)
    }
  }
  window.addEventListener('pagehide', flushPendingSettings)
  window.addEventListener('beforeunload', flushPendingSettings)
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
export const animScaleStore = persistentWritable<number>('vinta_anim_scale', 1.0)
export const monitorUiScalesStore = persistentWritable<Record<string, number>>(
  'vinta_monitor_ui_scales',
  {}
)

/**
 * Returns a stable storage key for a monitor's custom UI scale override.
 */
export function getMonitorScaleKey(
  monitor?: {
    name?: string
    left?: number
    top?: number
    isPrimary?: boolean
  } | null
): string {
  if (!monitor || monitor.isPrimary) return 'primary'
  return `${monitor.name || 'display'}_${monitor.left ?? 0}_${monitor.top ?? 0}`
}

/**
 * Computes the baseline physical scale for a monitor relative to the 1080p design base,
 * ensuring 1.0x looks proportionally consistent on 1440p/4K displays while respecting OS DPI.
 */
export function getMonitorBasePhysicalScale(
  width: number = 1920,
  height: number = 1080,
  dpiScale: number = 1.0
): number {
  const shortEdge = Math.min(width || 1920, height || 1080)
  const resBase = Math.max(1.0, shortEdge / 1080)
  const osDpi = dpiScale > 0 ? dpiScale : 1.0
  return Math.max(osDpi, resBase)
}

/**
 * Computes the maximum safe user UI scale multiplier on top of the monitor's baseline scale.
 */
export function getMaxScaleForMonitor(
  width: number = 1920,
  height: number = 1080,
  dpr?: number
): number {
  const shortEdge = Math.min(width || 1920, height || 1080)
  const basePhysical = getMonitorBasePhysicalScale(width, height, dpr)
  const effectiveShortEdge = shortEdge / basePhysical
  if (effectiveShortEdge <= 900) return 1.1
  if (shortEdge <= 1080) return 1.2
  if (shortEdge <= 1440) return 1.25
  return 1.5
}

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
export const clickOutsideMinimizeStore = persistentWritable<boolean>(
  'vinta_click_outside_minimize',
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
export const fullscreenBehaviorStore = persistentWritable<'relocate' | 'hide'>('vinta_fullscreen_behavior', 'relocate')
export const returnOnFullscreenExitStore = persistentWritable<boolean>('vinta_return_on_fullscreen_exit', false)

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

fullscreenBehaviorStore.subscribe((val) => {
  if (typeof SetRelocateOnFullscreen === 'function') {
    SetRelocateOnFullscreen(val === 'relocate').catch(() => {})
  }
})

returnOnFullscreenExitStore.subscribe((val) => {
  if (typeof SetReturnOnFullscreenExit === 'function') {
    SetReturnOnFullscreenExit(val).catch(() => {})
  }
})

// ============================================================================
// 3. Notifications & Do Not Disturb
// ============================================================================
export const dndStore = persistentWritable<boolean>('vinta_dnd', false)
export const notifAutoWakeStore = persistentWritable<boolean>('vinta_notif_autowake', true)
export const notifSatelliteEnabledStore = persistentWritable<boolean>('vinta_notif_satellite_enabled', true)
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
export const timeFormatStore = persistentWritable<TimeFormatOption>('vinta_time_format', inferDefaultTimeFormat())
export const timezoneStore = persistentWritable<string>('vinta_timezone', getInitialTimezone())
export const dateFormatStore = persistentWritable<DateFormatOption>('vinta_date_format', inferDefaultDateFormat())
export const autoDiscardPastEventsStore = persistentWritable<boolean>('vinta_auto_discard_past_events', false)

// ============================================================================
// 5. Weather
// ============================================================================
export const weatherLocationStore = persistentWritable<WeatherLocation>('vinta_weather_location', { name: 'Manila, Philippines', lat: 14.5995, lon: 120.9842 })
export const weatherUnitStore = persistentWritable<TemperatureUnit>('vinta_weather_unit', inferDefaultTemperatureUnit())
export const weatherAnimationsStore = persistentWritable<boolean>('vinta_weather_animations', true)

// ============================================================================
// 6. Home & Page Visibility
// ============================================================================
export const homeWidgetsVisibilityStore = persistentWritable<HomeWidgetsVisibility>('vinta_home_widgets_visibility', defaultHomeWidgetsVisibility)
export const homeWidgetsOrderStore = persistentWritable<string[]>(
  'vinta_home_widgets_order',
  defaultHomeWidgetsOrder
)
export const homeWidgetSizesStore = persistentWritable<Record<string, WidgetSizePreset>>('vinta_home_widget_sizes', defaultHomeWidgetSizes)
export const homeWidgetMenuEnabledStore = persistentWritable<boolean>('vinta_home_widget_menu_enabled', true)

/**
 * Normalizes a persisted Home widget order array against known default widget IDs.
 */
export function normalizeHomeWidgetsOrder(order?: string[] | null): string[] {
  const seen = new Set<string>()
  const result: string[] = []
  if (Array.isArray(order)) {
    for (const id of order) {
      if (defaultHomeWidgetsOrder.includes(id) && !seen.has(id)) {
        seen.add(id)
        result.push(id)
      }
    }
  }
  for (const id of defaultHomeWidgetsOrder) {
    if (!seen.has(id)) {
      seen.add(id)
      result.push(id)
    }
  }
  return result
}

/**
 * Moves a modular Home Bento widget up, down, or to the top of the Bento grid.
 */
export function moveHomeWidget(
  widgetId: string,
  direction: 'up' | 'down' | 'top'
): void {
  homeWidgetsOrderStore.update((prev) => {
    const order = normalizeHomeWidgetsOrder(prev)
    const idx = order.indexOf(widgetId)
    if (idx < 0) return order
    const next = [...order]
    if (direction === 'top') {
      if (idx === 0) return order
      next.splice(idx, 1)
      next.unshift(widgetId)
      return next
    }
    const targetIdx = direction === 'up' ? idx - 1 : idx + 1
    if (targetIdx < 0 || targetIdx >= next.length) return order
    ;[next[idx], next[targetIdx]] = [next[targetIdx], next[idx]]
    return next
  })
}

/**
 * Reorders a dragged Home Bento widget before or after a target widget ID.
 */
export function reorderHomeWidgets(sourceId: string, targetId: string): void {
  if (!sourceId || !targetId || sourceId === targetId) return
  homeWidgetsOrderStore.update((prev) => {
    const order = normalizeHomeWidgetsOrder(prev)
    const fromIdx = order.indexOf(sourceId)
    const toIdx = order.indexOf(targetId)
    if (fromIdx < 0 || toIdx < 0) return order
    const next = [...order]
    const [moved] = next.splice(fromIdx, 1)
    next.splice(toIdx, 0, moved)
    return next
  })
}
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
export const hasCompletedOnboardingStore = persistentWritable<boolean>('vinta_has_completed_onboarding', false)
/**
 * Resolves the initial display language by checking host system language preferences.
 */
function resolveInitialLocale(): string {
  if (typeof navigator !== 'undefined') {
    const list =
      navigator.languages && navigator.languages.length > 0
        ? navigator.languages
        : [navigator.language]
    for (const lang of list) {
      if (!lang) continue
      const lower = lang.toLowerCase().replace(/_/g, '-')
      if (lower.startsWith('fil') || lower.startsWith('tl')) return 'fil'
      if (lower.startsWith('de')) return 'de'
      if (lower.startsWith('hi')) return 'hi'
      if (lower.startsWith('es-ar')) return 'es-AR'
      if (lower.startsWith('es-es')) return 'es-ES'
      if (lower.startsWith('es-')) return 'es-419'
      if (lower === 'es') return 'es-ES'
      if (lower.startsWith('en-gb') || lower.startsWith('en-uk')) return 'en-GB'
      if (lower.startsWith('en')) return 'en'
    }
  }
  return 'en'
}

export const localeStore = persistentWritable<string>('vinta_locale', resolveInitialLocale())

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
 * Reactive store holding the dynamically measured client height of the Home Page feed.
 */
export const homeMeasuredHeightStore = writable<number>(0)

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
  const GAP = 10

  // 1. Header segment (Weather hero full-bleed or salutation greeting)
  let totalHeight = visibility.weather ? 158 : (visibility.clock ? 94 : 72)

  // 2. Update notice banner
  if (hasUpdate) {
    totalHeight += GAP + 54
  }

  // 3. Spotlight carousel banner
  if (visibility.spotlight !== false) {
    totalHeight += GAP + 72
  }

  // 4. Announcements & Release Notes glance card (fixed 76px height)
  if (visibility.announcements !== false) {
    totalHeight += GAP + 76
  }

  // 5. Notification glance card (only renders if unread notifications exist)
  if (hasNotification) {
    totalHeight += GAP + 78
  }

  // 6. Calendar glance agenda
  if (visibility.calendar) {
    totalHeight += GAP + 82
  }

  // 7. Quick Controls / Action Center glance
  if (visibility.actionCenter) {
    totalHeight += GAP + 96
  }

  // 8. System Stats glance card
  if (visibility.systemStats) {
    totalHeight += GAP + 128
  }

  // 9. Studio Branding Footer ("Vinta by FreeMatrix Studios & Contributors")
  totalHeight += GAP + 24

  // 10. Bottom clearance for feed padding & island corner curvature
  const BOTTOM_CLEARANCE = 10

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

      if (key === 'vinta_home_widgets_order') {
        sanitized[key] = Array.isArray(val)
          ? normalizeHomeWidgetsOrder(val.filter((x): x is string => typeof x === 'string'))
          : defaultHomeWidgetsOrder
        continue
      }

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
          } else if (key === 'vinta_anim_scale') {
            sanitized[key] = Math.max(0, Math.min(2.0, Number(val)))
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
  sanitized['_schema_version'] = CURRENT_SETTINGS_SCHEMA_VERSION
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

async function applySanitizedSettingsPayload(
  rawJSON: string
): Promise<SettingsImportResult> {
  let parsed: Record<string, unknown>
  try {
    parsed = JSON.parse(rawJSON)
  } catch {
    return { success: false, error: 'Selected file is not valid JSON' }
  }

  const envelopeVersion =
    typeof parsed.$schemaVersion === 'number' ? parsed.$schemaVersion : undefined

  const settingsPayload =
    parsed.settings && typeof parsed.settings === 'object' && !Array.isArray(parsed.settings)
      ? (parsed.settings as Record<string, unknown>)
      : parsed

  const migrated = migrateSettings(settingsPayload, envelopeVersion)
  const sanitized = sanitizeSettingsPayload(migrated.migrated)
  const sanitizedKeys = Object.keys(sanitized).filter(
    (k) => k !== '_sparse_migrated' && k !== '_schema_version'
  )

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

    return await applySanitizedSettingsPayload(rawJSON)
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err)
    return { success: false, error: msg }
  }
}

/**
 * Restores the most recent automatic settings snapshot and updates reactive state in-place.
 *
 * @returns Result object with status and restored settings count.
 */
export async function restoreLatestSettingsBackup(): Promise<SettingsImportResult> {
  try {
    if (
      typeof RestoreLatestSettingsBackup !== 'function' ||
      typeof ApplyImportedSettings !== 'function'
    ) {
      return { success: false, error: 'Backup restore function not available' }
    }

    const rawJSON = await RestoreLatestSettingsBackup()
    if (!rawJSON) {
      return { success: false, error: 'No automatic settings backup found' }
    }

    return await applySanitizedSettingsPayload(rawJSON)
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err)
    return { success: false, error: msg }
  }
}

