/**
 * @fileoverview Shared global feedback timeout helpers for inline UI confirmations and status banners.
 *
 * Reads the user-configured global `feedbackTimeoutSecStore` so any settings tab or
 * component can automatically clear temporary status messages and restore its default layout.
 */

import { get } from 'svelte/store'
import { feedbackTimeoutSecStore } from '../stores/settings'

/**
 * Feedback severity used to scale how long an inline status message remains visible.
 */
export type FeedbackSeverity = 'success' | 'info' | 'warning' | 'error'

/**
 * Returns the active feedback timeout in milliseconds derived from `feedbackTimeoutSecStore`.
 * Warning and error messages stay visible 1.5x longer so users have time to read them.
 *
 * @param severity - Severity of the feedback message.
 * @returns Timeout duration in milliseconds.
 */
export function getFeedbackTimeoutMs(
  severity: FeedbackSeverity = 'success'
): number {
  const baseSec = Number(get(feedbackTimeoutSecStore))
  const clampedSec = Number.isFinite(baseSec) ? Math.max(0.5, baseSec) : 2.5
  const baseMs = Math.round(clampedSec * 1000)
  return severity === 'warning' || severity === 'error'
    ? Math.round(baseMs * 1.5)
    : baseMs
}

/**
 * Schedules automatic clearance of a keyed inline feedback message, cancelling any
 * previous pending timer for the same key.
 *
 * @param timers - Mutable map tracking active timer handles by key.
 * @param key - Identifier of the item or row displaying feedback.
 * @param onClear - Callback invoked when the timeout elapses.
 * @param severity - Severity used to compute the timeout duration.
 */
export function scheduleFeedbackClear(
  timers: Record<string, ReturnType<typeof setTimeout>>,
  key: string,
  onClear: () => void,
  severity: FeedbackSeverity = 'success'
): void {
  clearFeedbackTimer(timers, key)
  const delayMs = getFeedbackTimeoutMs(severity)
  timers[key] = setTimeout(() => {
    delete timers[key]
    onClear()
  }, delayMs)
}

/**
 * Cancels and removes any pending feedback clear timer for the given key.
 *
 * @param timers - Mutable map tracking active timer handles by key.
 * @param key - Identifier of the item or row whose timer should be cancelled.
 */
export function clearFeedbackTimer(
  timers: Record<string, ReturnType<typeof setTimeout>>,
  key: string
): void {
  if (timers[key]) {
    clearTimeout(timers[key])
    delete timers[key]
  }
}
