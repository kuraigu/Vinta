/**
 * @fileoverview Single registry and dispatcher for all internal Vinta notifications.
 *
 * Keeps notification copy, deduplication rules, avatar resolution, and sound settings
 * in one place so internal alerts stay consistent under the unified `Vinta` sender.
 */

import { get } from 'svelte/store'
import { Events } from '$lib/events'
import appIcon from '@appicon'
import {
  notifData,
  dndStore,
  notifSoundStore,
  notifVolumeStore,
  suppressExternalNotifications,
  MAX_NOTIFICATIONS
} from '../store'
import { playNotifChime } from './audioService'
import { getNativeAppIcon } from './iconService'

export interface VintaNotificationOptions {
  title: string
  message: string
  icon?: string
  playSound?: boolean
  emitNativeToast?: boolean
  canOpen?: boolean
  isUpdate?: boolean
  dedupeKey?: string
}

/**
 * Pushes a Vinta alert to the notification feed, desktop toast bridge, and chime player.
 */
export function sendVintaNotification(options: VintaNotificationOptions): void {
  if (get(suppressExternalNotifications)) {
    return
  }
  const {
    title,
    message,
    playSound = true,
    emitNativeToast = true,
    canOpen,
    isUpdate,
    dedupeKey
  } = options
  const icon = options.icon || appIcon
  const appName = 'Vinta'

  const newItem = {
    id: Date.now() + Math.floor(Math.random() * 1000),
    appName,
    title,
    message,
    icon,
    appLogo: appIcon,
    canOpen,
    isUpdate,
    dedupeKey,
    timestamp: Date.now()
  }

  notifData.update((items) => {
    const filtered = items.filter((item) => {
      if (dedupeKey && (item as { dedupeKey?: string }).dedupeKey) {
        return (item as { dedupeKey?: string }).dedupeKey !== dedupeKey
      }
      return !(
        item.title === title &&
        (item.appName === appName || item.appName.startsWith('Vinta'))
      )
    })
    return [newItem, ...filtered].slice(0, MAX_NOTIFICATIONS)
  })

  if (emitNativeToast) {
    try {
      Events.Emit('send-native-notification', { title, message })
    } catch {
      // Ignore native notification event failure
    }
  }

  if (playSound && get(notifSoundStore) && !get(dndStore)) {
    try {
      playNotifChime(get(notifVolumeStore))
    } catch {
      // Ignore audio playback failure
    }
  }
}

/**
 * Sends an auto-hide notice and attaches the foreground app's cached icon as the avatar.
 */
export async function notifyAutoHidden(params: {
  appName?: string
  reason?: string
}): Promise<void> {
  const appName = params.appName?.trim() || ''
  const reason = params.reason || ''

  let message = appName
    ? `Hid Vinta while ${appName} is in exclusive fullscreen. Switch it to borderless fullscreen if you want Vinta to stay visible.`
    : 'Hid Vinta for exclusive fullscreen. Switch the app to borderless fullscreen if you want Vinta to stay visible.'

  if (reason === 'presentation_mode') {
    message = appName
      ? `Hid Vinta while ${appName} is in presentation mode.`
      : 'Hid Vinta while presentation mode is active.'
  } else if (reason === 'unsupported_app') {
    message = appName
      ? `Hid Vinta because ${appName} doesn't support overlays.`
      : "Hid Vinta because the active app doesn't support overlays."
  } else if (reason === 'borderless_fullscreen') {
    message = appName
      ? `Hid Vinta while ${appName} is in borderless fullscreen. You can change this in Settings > Behavior.`
      : 'Hid Vinta while borderless fullscreen is active. You can change this in Settings > Behavior.'
  }

  const avatarIcon = appName ? await getNativeAppIcon(appName) : ''

  sendVintaNotification({
    title: 'Auto-Hidden',
    message,
    icon: avatarIcon || undefined,
    dedupeKey: `auto-hide:${appName || 'Vinta'}`,
    playSound: false
  })
}

/**
 * Sends a notice when Vinta relocates to another monitor to avoid fullscreen conflicts.
 */
export function notifyDisplayRelocated(params: {
  targetName: string
  appName?: string
}): void {
  const targetName = params.targetName || 'Secondary Display'
  const appName = params.appName?.trim() || ''

  sendVintaNotification({
    title: `Moved to ${targetName}`,
    message: appName
      ? `Moved to ${targetName} to stay clear of ${appName}.`
      : `Moved to ${targetName} to stay clear of fullscreen.`,
    dedupeKey: 'display-relocate'
  })
}

/**
 * Sends an in-app alert when a newer Vinta release is ready to install.
 */
export function notifyUpdateAvailable(latestVersion: string): void {
  sendVintaNotification({
    title: `Vinta ${latestVersion} is out`,
    message:
      'A new build is ready. Open Settings > Updates to download it and check what changed.',
    canOpen: true,
    isUpdate: true,
    playSound: false,
    emitNativeToast: false,
    dedupeKey: 'vinta-update-available'
  })
}

/**
 * Sends an in-app notice right after Vinta finishes updating to a new version.
 */
export function notifyPostUpdateInstalled(newVersion: string): void {
  sendVintaNotification({
    title: `Updated to Vinta ${newVersion}`,
    message:
      "You're on the latest build. Open Settings > Updates to see what's new.",
    canOpen: true,
    isUpdate: true,
    playSound: false,
    emitNativeToast: false,
    dedupeKey: 'vinta-post-update'
  })
}

/**
 * Sends a calendar reminder alert under the unified Vinta sender.
 */
export function notifyCalendarEvent(params: {
  title?: string
  message: string
  meetingUrl?: string
}): void {
  const fullMessage = params.meetingUrl
    ? `${params.message} • Join: ${params.meetingUrl}`
    : params.message
  sendVintaNotification({
    title: params.title?.trim() || 'Calendar Event',
    message: fullMessage,
    canOpen: Boolean(params.meetingUrl)
  })
}

/**
 * Sends a timer, alarm, or reminder alert under the unified Vinta sender.
 */
export function notifyTimerAlert(title: string, message: string): void {
  sendVintaNotification({ title, message })
}
