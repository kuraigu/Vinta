/**
 * @fileoverview Floating notification capsule display and timer.
 *
 * Manages the floating notification pill that appears outside the island,
 * pauses the auto-dismiss timer on hover, and dismisses on click.
 */

import { writable, get } from 'svelte/store'
import type { AppNotification } from './system'
import { autoHideDelayStore, notifSatelliteEnabledStore } from './settings'

export const satelliteNotification = writable<AppNotification | null>(null)

let dismissTimer: number | null = null
let remainingTimeMs = 8000
let timerStartTime: number | null = null

function getDurationMs(): number {
  const configuredSeconds = get(autoHideDelayStore)
  if (typeof configuredSeconds === 'number' && configuredSeconds > 0) {
    return configuredSeconds * 1000
  }
  return 8000
}

/**
 * Automatically dismisses the active satellite notification capsule after timeout.
 */
function handleAutoDismiss() {
  satelliteNotification.set(null)
  dismissTimer = null
  timerStartTime = null
  remainingTimeMs = getDurationMs()
}

/**
 * Displays a new notification in the floating satellite surface.
 * Starts a countdown timer synced with notification auto-hide preference (or 8s default).
 */
export function showSatelliteNotification(notif: AppNotification) {
  if (!get(notifSatelliteEnabledStore)) return
  if (dismissTimer !== null) {
    window.clearTimeout(dismissTimer)
    dismissTimer = null
  }
  remainingTimeMs = getDurationMs()
  timerStartTime = Date.now()
  satelliteNotification.set(notif)
  dismissTimer = window.setTimeout(handleAutoDismiss, remainingTimeMs)
}

/**
 * Pauses the auto-dismiss timer when the user hovers over the satellite.
 */
export function pauseSatelliteTimer() {
  if (dismissTimer !== null && timerStartTime !== null) {
    window.clearTimeout(dismissTimer)
    dismissTimer = null
    const elapsed = Date.now() - timerStartTime
    remainingTimeMs = Math.max(1200, remainingTimeMs - elapsed)
  }
}

/**
 * Resumes the auto-dismiss timer when pointer hover ends.
 */
export function resumeSatelliteTimer() {
  if (dismissTimer === null) {
    timerStartTime = Date.now()
    dismissTimer = window.setTimeout(handleAutoDismiss, remainingTimeMs)
  }
}

/**
 * Dismisses the active satellite notification immediately.
 */
export function dismissSatelliteNotification() {
  if (dismissTimer !== null) {
    window.clearTimeout(dismissTimer)
    dismissTimer = null
  }
  satelliteNotification.set(null)
  timerStartTime = null
  remainingTimeMs = getDurationMs()
}

