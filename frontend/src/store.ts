import { writable } from 'svelte/store'

export type AppTheme = 'dark' | 'light'
export type IslandShape = 'pill' | 'box'
export type IslandConfig = {
  minimal: { w: number; h: number; r: number }
  expanded: { w: number; h: number; r: number }
  borderColor?: string
}

export const themeStore = writable<AppTheme>('dark')
export const shapeStore = writable<IslandShape>('pill')

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
