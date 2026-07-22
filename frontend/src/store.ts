import { writable } from 'svelte/store'

function persistentWritable<T>(key: string, initialValue: T) {
  const stored = localStorage.getItem(key)
  const initial = stored ? JSON.parse(stored) : initialValue
  const store = writable<T>(initial)
  store.subscribe((value) => {
    localStorage.setItem(key, JSON.stringify(value))
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
