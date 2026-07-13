import { writable } from 'svelte/store'

function persistentWritable<T>(key: string, initialValue: T) {
  const stored = localStorage.getItem(key)
  const initial = stored ? JSON.parse(stored) : initialValue
  const store = writable<T>(initial)
  store.subscribe(value => {
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
export const autoHideStore = persistentWritable<boolean>('vinta_autohide', false)
export const wakeMethodStore = persistentWritable<'hover' | 'click'>('vinta_wakemethod', 'hover')

export type AppNotification = {
  id: number
  appName?: string
  icon?: string
  title: string
  message: string
  time?: string
}

export const notifData = writable<AppNotification[]>([
  {
    id: 1,
    appName: 'Messages',
    icon: '💬',
    title: 'Sarah Jenkins',
    message: 'Are we still on for the design review later?',
    time: 'Now'
  },
  {
    id: 2,
    appName: 'System',
    icon: '⚙️',
    title: 'Update Available',
    message: 'A new system update is ready to install.',
    time: '5m'
  }
])
