import { writable } from 'svelte/store'

export type AppTheme = 'dark' | 'light'
export type IslandShape = 'pill' | 'box'
export type IslandConfig = {
  minimal: { w: number; h: number; r: number }
  expanded: { w: number; h: number; r: number }
}

export const themeStore = writable<AppTheme>('dark')
export const shapeStore = writable<IslandShape>('pill')

export type AppNotification = {
  id: number
  title: string
  message: string
}

export const notifData = writable<AppNotification[]>([
  {
    id: 1,
    title: 'Notification',
    message: 'This is what a real notification will look like.'
  }
])
