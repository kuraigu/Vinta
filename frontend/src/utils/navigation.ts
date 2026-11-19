/**
 * @fileoverview Island page navigation and default view resolver.
 *
 * Determines which page to show on wake (showing the timer if running, media player
 * if playing, or onboarding on first run) and provides navigation helper functions.
 */

import { Events } from '$lib/events'
import { syncSettingsToDisk } from '../store'
import HomePage, { config as homeConfig } from '../pages/HomePage.svelte'
import SettingsPage, {
  config as settingsConfig
} from '../pages/SettingsPage.svelte'
import MediaPage, { config as mediaConfig } from '../pages/MediaPage.svelte'
import Notification, {
  config as notifConfig
} from '../components/Notification.svelte'
import NotesPage, { config as notesConfig } from '../pages/NotesPage.svelte'
import CalculatorPage, {
  config as calculatorConfig
} from '../pages/CalculatorPage.svelte'
import TimerPage, { config as timerConfig } from '../pages/TimerPage.svelte'
import CalendarPage, {
  config as calendarConfig
} from '../pages/CalendarPage.svelte'
import FileShelfPage, {
  config as fileShelfConfig
} from '../pages/FileShelfPage.svelte'
import PerformancePage, {
  config as performanceConfig
} from '../pages/PerformancePage.svelte'
import OnboardingPage, {
  config as onboardingConfig
} from '../pages/OnboardingPage.svelte'
import {
  globalMediaStore,
  pageVisibilityStore,
  activeSettingsTabStore,
  timerAppStore,
  isAnyTimerRunning,
  hasCompletedOnboardingStore,
  type IslandConfig
} from '../store'
import { get } from 'svelte/store'

export type NavFn = (
  page: unknown,
  pageConfig: IslandConfig,
  expand?: boolean
) => void

export function getFallbackPage(): {
  page: unknown
  config: IslandConfig
} {
  if (!get(hasCompletedOnboardingStore)) {
    return { page: OnboardingPage, config: onboardingConfig }
  }

  const timerState = get(timerAppStore)
  const isTimerEnabled = get(pageVisibilityStore).timer !== false
  if (isAnyTimerRunning(timerState) && isTimerEnabled) {
    return { page: TimerPage, config: timerConfig }
  }

  const mediaState = get(globalMediaStore)
  const isMediaEnabled = get(pageVisibilityStore).media
  const isPlaying =
    Boolean(mediaState.isPlaying) ||
    (Array.isArray(mediaState.sessions) &&
      mediaState.sessions.some((s) => Boolean(s.isPlaying)))
  if (isPlaying && mediaState.isAvailable && isMediaEnabled) {
    return { page: MediaPage, config: mediaConfig }
  }
  return { page: HomePage, config: homeConfig }
}

export const AppControls = {
  getFallbackPage,

  goHome: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(HomePage, homeConfig, true)
  },

  goApps: (openAppsOverlay?: () => void, e?: Event) => {
    if (e) e.stopPropagation()
    openAppsOverlay?.()
  },

  goMedia: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(MediaPage, mediaConfig, true)
  },

  goNotif: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(Notification, notifConfig, true)
  },

  goSettings: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    if (get(activeSettingsTabStore) === 'about') {
      activeSettingsTabStore.set('appearance')
    }
    navigate(SettingsPage, settingsConfig, true)
  },

  goNotes: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(NotesPage, notesConfig, true)
  },

  goCalculator: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(CalculatorPage, calculatorConfig, true)
  },

  goTimer: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(TimerPage, timerConfig, true)
  },

  goFocus: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(TimerPage, timerConfig, true)
  },

  goCalendar: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(CalendarPage, calendarConfig, true)
  },

  goFileShelf: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(FileShelfPage, fileShelfConfig, true)
  },

  goPerformance: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(PerformancePage, performanceConfig, true)
  },

  goAbout: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    activeSettingsTabStore.set('about')
    navigate(SettingsPage, settingsConfig, true)
  },

  doMinimize: (navigate: NavFn, e?: Event, currentPage?: unknown) => {
    if (e) {
      e.stopPropagation()
      e.preventDefault()
    }
    if (currentPage === MediaPage) {
      navigate(MediaPage, mediaConfig, false)
      return
    }
    if (currentPage === TimerPage && isAnyTimerRunning(get(timerAppStore))) {
      navigate(TimerPage, timerConfig, false)
      return
    }
    const fallback = getFallbackPage()
    navigate(fallback.page, fallback.config, false)
  },

  doExpand: (navigate: NavFn, e?: Event, currentPage?: unknown) => {
    if (e) {
      e.stopPropagation()
      e.preventDefault()
    }
    if (currentPage === MediaPage) {
      navigate(MediaPage, mediaConfig, true)
      return
    }
    if (currentPage === TimerPage) {
      navigate(TimerPage, timerConfig, true)
      return
    }
    const fallback = getFallbackPage()
    navigate(fallback.page, fallback.config, true)
  },

  doHide: (_navigate: NavFn, hidePill: () => void, e?: Event) => {
    if (e) {
      e.stopPropagation()
      e.preventDefault()
    }
    hidePill()
  },

  doQuit: (e?: Event) => {
    if (e) e.stopPropagation()
    syncSettingsToDisk(true)
    Events.Emit('vinta:quit')
  }
}
