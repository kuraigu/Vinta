import { get } from 'svelte/store'
import appIcon from '@appicon'
import {
  notifData,
  dndStore,
  notifSoundStore,
  notifVolumeStore,
  MAX_NOTIFICATIONS
} from '../store'
import { playNotifChime } from './audioService'

export interface VintaNotificationOptions {
  title: string
  message: string
  subApp?: string
  icon?: string
  playSound?: boolean
}

/**
 * Dispatches an official Vinta notification to the in-app notification center and OS desktop toast.
 * Guarantees that internal notifications always attach the official Vinta application icon.
 *
 * @param options - Notification parameters including title, message body, and optional sub-app name.
 */
export function sendVintaNotification(options: VintaNotificationOptions): void {
  const { title, message, subApp, playSound = true } = options
  const icon = options.icon || appIcon
  const appName = subApp ? `Vinta ${subApp}` : 'Vinta'

  // 1. In-app notification store
  notifData.update((items) => [
    {
      id: Date.now() + Math.floor(Math.random() * 1000),
      appName,
      title,
      message,
      icon,
      appLogo: icon,
      timestamp: Date.now()
    },
    ...items.slice(0, MAX_NOTIFICATIONS - 1)
  ])

  // 2. Desktop toast notification
  if (
    typeof Notification !== 'undefined' &&
    Notification.permission === 'granted'
  ) {
    try {
      new Notification(title, { body: message, icon })
    } catch {
      // Ignore desktop notification failure
    }
  }

  // 3. Audio feedback chime if enabled and DND is inactive
  if (playSound && get(notifSoundStore) && !get(dndStore)) {
    try {
      playNotifChime(get(notifVolumeStore))
    } catch {
      // Ignore audio playback failure
    }
  }
}
