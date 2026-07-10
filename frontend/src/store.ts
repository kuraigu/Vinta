import { writable } from 'svelte/store'

export type AppNotification = {
  id: number
  title: string
  message: string
}

export const notifData = writable<AppNotification[]>([
  { id: 1, title: "Notification", message: "This is what a real notification will look like." }
])
