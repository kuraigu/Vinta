import { writable } from 'svelte/store'
import {
  GetNotifications,
  SaveNotifications
} from '../../bindings/Vinta/backend/app.js'

// ============================================================================
// Notifications
// ============================================================================
export type AppNotification = {
  id: number
  nativeId?: number
  appName?: string
  aumid?: string
  icon?: string
  appLogo?: string
  title: string
  message: string
  heroImage?: string
  timestamp: number
  canOpen?: boolean
  isUpdate?: boolean
}

export const MAX_NOTIFICATIONS = 999

function loadInitialNotifications(): AppNotification[] {
  try {
    const raw = localStorage.getItem('vinta_notifications')
    if (raw) {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) return parsed.slice(0, MAX_NOTIFICATIONS)
    }
  } catch {
    // Ignore initial storage read error
  }
  return []
}

export const notifData = writable<AppNotification[]>(loadInitialNotifications())
export const notificationPermissionDenied = writable<boolean>(false)

let saveNotificationsTimeout: ReturnType<typeof setTimeout> | null = null

function syncNotificationsToDisk(items: AppNotification[]) {
  if (typeof SaveNotifications === 'function') {
    if (saveNotificationsTimeout) clearTimeout(saveNotificationsTimeout)
    saveNotificationsTimeout = setTimeout(() => {
      SaveNotifications(
        JSON.stringify(items.slice(0, MAX_NOTIFICATIONS))
      ).catch(() => {})
    }, 400)
  }
}

// Hydrate persisted notifications from backend storage on application launch
if (typeof GetNotifications === 'function') {
  GetNotifications()
    .then((jsonStr) => {
      if (jsonStr && jsonStr !== '[]') {
        try {
          const parsed = JSON.parse(jsonStr)
          if (Array.isArray(parsed) && parsed.length > 0) {
            notifData.set(parsed.slice(0, MAX_NOTIFICATIONS))
            try {
              localStorage.setItem(
                'vinta_notifications',
                JSON.stringify(parsed.slice(0, MAX_NOTIFICATIONS))
              )
            } catch {
              // Ignore localStorage write error
            }
          }
        } catch {
          // Ignore parse errors from corrupted storage
        }
      }
    })
    .catch(() => {})
}

// Subscribe to persist notification list changes capped at 999 items
notifData.subscribe((items) => {
  if (Array.isArray(items)) {
    const capped = items.slice(0, MAX_NOTIFICATIONS)
    try {
      localStorage.setItem('vinta_notifications', JSON.stringify(capped))
    } catch {
      // Ignore storage write error
    }
    syncNotificationsToDisk(capped)
  }
})

// ============================================================================
// Media Player & Audio Peaks
// ============================================================================
export interface MediaSessionItem {
  id: string
  title: string
  artist: string
  sourceApp: string
  isPlaying: boolean
}

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
  sessions?: MediaSessionItem[]
  activeSessionId?: string
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

// ============================================================================
// Battery
// ============================================================================
export interface BatteryData {
  hasBattery: boolean
  isCharging: boolean
  percentage: number
  batterySaver: boolean
  lifeTimeSeconds: number
  state: 'charging' | 'discharging' | 'full' | 'desktop' | 'unknown'
}

export const defaultBatteryData: BatteryData = {
  hasBattery: false,
  isCharging: true,
  percentage: 100,
  batterySaver: false,
  lifeTimeSeconds: -1,
  state: 'desktop'
}

export const batteryDataStore = writable<BatteryData>(defaultBatteryData)

// ============================================================================
// Wireless (Wi-Fi & Bluetooth)
// ============================================================================
export interface WirelessData {
  hasWifi: boolean
  wifiEnabled: boolean
  wifiConnected: boolean
  wifiSsid: string
  wifiSignal: number // 0-100

  hasBluetooth: boolean
  bluetoothEnabled: boolean
  bluetoothConnected: boolean
  bluetoothDeviceName: string
  bluetoothBattery?: number

  airplaneMode: boolean
}

export const defaultWirelessData: WirelessData = {
  hasWifi: true,
  wifiEnabled: true,
  wifiConnected: true,
  wifiSsid: 'Wi-Fi',
  wifiSignal: 85,

  hasBluetooth: false,
  bluetoothEnabled: false,
  bluetoothConnected: false,
  bluetoothDeviceName: '',
  bluetoothBattery: undefined,

  airplaneMode: false
}

export const wirelessDataStore = writable<WirelessData>(defaultWirelessData)
