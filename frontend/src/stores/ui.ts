/**
 * @fileoverview UI view state and developer debug flags.
 *
 * Tracks the current active page, whether the island is minimized into the pill,
 * media drawer visibility, and test simulation overrides (mock battery, offline mode).
 */

import { writable } from 'svelte/store'
import type { WeatherData } from '../services/weatherService'

export type OverlayPageId = 'notifications' | 'apps' | null
export type NotificationsOverlayTab =
  | 'notifications'
  | 'announcements'
  | 'releases'

export const activeSettingsTabStore = writable<string>('appearance')
export const activePageIdStore = writable<string>('home')
export const activeOverlayStore = writable<OverlayPageId>(null)
export const notificationsOverlayTabStore =
  writable<NotificationsOverlayTab>('notifications')
export const hiddenPillOffsetXStore = writable<number>(0)
export const isIslandHiddenStore = writable<boolean>(false)
export const isMediaDrawerOpenStore = writable<boolean>(false)

export const simulateOfflineStore = writable<boolean>(false)
export const simulateWeatherStore = writable<WeatherData | null>(null)
export const isSimulatingBatteryStore = writable<boolean>(false)
export const isSimulatingWirelessStore = writable<boolean>(false)
