/**
 * @fileoverview UI view state and developer debug flags.
 *
 * Tracks the current active page, whether the island is minimized into the pill,
 * media drawer visibility, and test simulation overrides (mock battery, offline mode).
 */

import { get, writable } from 'svelte/store'
import type { WeatherData } from '../services/weatherService'

export type OverlayPageId = 'notifications' | 'apps' | 'feedback' | null
export type FeedbackOverlayKind = 'bug' | 'feature'
export type AppsOverlayTab = 'island' | 'desktop'
export type NotificationsOverlayTab =
  | 'notifications'
  | 'announcements'
  | 'releases'

export interface FeedbackDraft {
  title: string
  describe: string
  extra: string
}

export const feedbackDraftsStore = writable<
  Record<FeedbackOverlayKind, FeedbackDraft>
>({
  bug: { title: '', describe: '', extra: '' },
  feature: { title: '', describe: '', extra: '' }
})

/**
 * Resets the persisted feedback draft fields for the specified issue kind.
 */
export function clearFeedbackDraft(kind: FeedbackOverlayKind): void {
  feedbackDraftsStore.update((drafts) => ({
    ...drafts,
    [kind]: { title: '', describe: '', extra: '' }
  }))
}

export const activeSettingsTabStore = writable<string>('appearance')
export const activePageIdStore = writable<string>('home')
export const activeOverlayStore = writable<OverlayPageId>(null)
export const appsOverlayTabStore = writable<AppsOverlayTab>('island')
export const notificationsOverlayTabStore =
  writable<NotificationsOverlayTab>('notifications')
export const feedbackOverlayKindStore = writable<FeedbackOverlayKind>('bug')
export const hiddenPillOffsetXStore = writable<number>(0)
export const isIslandHiddenStore = writable<boolean>(false)
export const overlayPinnedStore = writable<boolean>(false)
export const overlayPinOverrideStore = writable<{
  scope: string
  pinned: boolean
} | null>(null)
export const authSignInActiveStore = writable<boolean>(false)
export const isMediaDrawerOpenStore = writable<boolean>(false)

let activeAuthSignInCount = 0

/**
 * Returns the scope key for the currently visible page or overlay so pin state
 * is bound to that view and never carries over across page navigation.
 */
export function getCurrentPinScope(): string {
  const overlay = get(activeOverlayStore)
  if (overlay !== null) {
    return `overlay:${overlay}`
  }
  return `page:${get(activePageIdStore)}`
}

/**
 * Toggles pin state strictly for the currently active page or overlay.
 */
export function toggleCurrentOverlayPin(): void {
  const scope = getCurrentPinScope()
  const nextPinned = !get(overlayPinnedStore)
  overlayPinOverrideStore.set({ scope, pinned: nextPinned })
}

/**
 * Clears any manual per-page pin override.
 */
export function clearOverlayPinOverride(): void {
  if (get(overlayPinOverrideStore) !== null) {
    overlayPinOverrideStore.set(null)
  }
}

/**
 * Marks an external browser sign-in flow as active so the island temporarily stays pinned.
 */
export function beginAuthSignIn(): void {
  activeAuthSignInCount += 1
  if (activeAuthSignInCount === 1) {
    authSignInActiveStore.set(true)
  }
}

/**
 * Releases one active external browser sign-in flow, unpinning once all flows finish.
 */
export function endAuthSignIn(): void {
  activeAuthSignInCount = Math.max(0, activeAuthSignInCount - 1)
  if (activeAuthSignInCount === 0) {
    authSignInActiveStore.set(false)
  }
}

export const simulateOfflineStore = writable<boolean>(false)
export const simulateWeatherStore = writable<WeatherData | null>(null)
export const isSimulatingBatteryStore = writable<boolean>(false)
export const isSimulatingWirelessStore = writable<boolean>(false)
