import { get } from 'svelte/store'
import { Events } from '@wailsio/runtime'
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
  const newItem = {
    id: Date.now() + Math.floor(Math.random() * 1000),
    appName,
    title,
    message,
    icon,
    appLogo: icon,
    timestamp: Date.now()
  }

  notifData.update((items) => {
    // If an identical notification from Vinta exists, discard the old one and replace with latest
    const filtered = items.filter(
      (item) =>
        !(
          item.title === title &&
          (item.appName === appName || item.appName.startsWith('Vinta'))
        )
    )
    return [newItem, ...filtered].slice(0, MAX_NOTIFICATIONS)
  })

  // 2. Desktop OS toast notification (via native WinRT bridge)
  try {
    Events.Emit('send-native-notification', { title, message })
  } catch {
    // Ignore native notification event failure
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
