import { writable } from 'svelte/store'

// ============================================================================
// Notifications
// ============================================================================
export type AppNotification = {
  id: number
  appName?: string
  aumid?: string
  icon?: string
  title: string
  message: string
  heroImage?: string
  timestamp: number
  canOpen?: boolean
}

export const notifData = writable<AppNotification[]>([])
export const notificationPermissionDenied = writable<boolean>(false)

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
