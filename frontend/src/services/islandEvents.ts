/**
 * @fileoverview Coordinates backend system events with the frontend UI.
 *
 * Listens for OS events from the Go backend (media updates, toast notifications,
 * display changes, battery/wireless status, hotkeys) and triggers UI updates.
 */

import type { Component } from 'svelte'
import { get } from 'svelte/store'
import { Events } from '$lib/events'
import appIcon from '@appicon'
import {
  GetBatteryStatus,
  GetWirelessStatus,
  CheckPostUpdateInstalled,
  RequestMediaUpdate
} from '../../bindings/Vinta/backend/app.js'
import HomePage from '../pages/HomePage.svelte'
import MediaPage, { config as mediaConfig } from '../pages/MediaPage.svelte'
import NotificationPage, {
  config as notifConfig
} from '../pages/NotificationPage.svelte'
import SettingsPage, {
  config as settingsConfig
} from '../pages/SettingsPage.svelte'
import FileShelfPage, {
  config as fileShelfConfig
} from '../pages/FileShelfPage.svelte'
import { initFileShelfStore, addFilesToShelf } from '../stores/fileShelf'
import { playNotifChime } from './audioService'
import {
  themeStore,
  globalMediaStore,
  audioPeakStore,
  notifData,
  MAX_NOTIFICATIONS,
  notificationPermissionDenied,
  suppressExternalNotifications,
  notifSoundStore,
  notifVolumeStore,
  dndStore,
  notifAutoWakeStore,
  notifSatelliteEnabledStore,
  pageVisibilityStore,
  homeWidgetsVisibilityStore,
  activeSettingsTabStore,
  updateInfoStore,
  autoCheckUpdatesStore,
  batteryDataStore,
  isSimulatingBatteryStore,
  wirelessDataStore,
  isSimulatingWirelessStore,
  type IslandConfig,
  type MediaData,
  type MediaSessionItem,
  type UpdateCheckResult,
  type BatteryData,
  type WirelessData,
  hiddenPillOffsetXStore,
  displayLayoutStore,
  hasCompletedOnboardingStore
} from '../store'
import type { MonitorInfo } from '../../bindings/Vinta/backend/models.js'
import { sendVintaNotification } from './vintaNotification.js'
import { showSatelliteNotification } from '../stores/notificationSatellite'
import { getDeveloperAnnouncements } from './announcementsService'

/**
 * Callbacks provided by the island orchestrator to handle system and runtime events.
 */
export interface IslandEventCallbacks {
  showWidget: () => void
  hideWidget: () => void
  navigateTo: (page: Component, config: IslandConfig, expand?: boolean) => void
  triggerGlow: () => void
  getIsHidden: () => boolean
  getIsExpanded: () => boolean
  getActivePage: () => Component
  getIsPageDormant: () => boolean
  onHideTrigger: () => void
  onMonitorChanged: (mon: MonitorInfo) => void
  onRefreshDisplay: () => void
  onRelocateOffset?: (newOffset: number) => void
  onNotificationWake?: () => void
  onHotkeyToggle?: () => void
  onQuickNotesHotkey?: () => void
  onModeSwitchHotkey?: () => void
}

/**
 * Subscribes to OS background events, media updates, notifications, and hardware status polling.
 *
 * @param callbacks Orchestration callbacks for navigation, visibility, and display refreshes.
 * @returns Teardown function to clean up all active event subscriptions and timers.
 */
export function initIslandEvents(callbacks: IslandEventCallbacks): () => void {
  const unsubs: (() => void)[] = []

  function onEvent<T = unknown>(
    eventName: string,
    callback: (data: T) => void
  ): () => void {
    const unsub = Events.On(eventName, (ev: any) => {
      const data =
        ev?.data !== undefined
          ? Array.isArray(ev.data)
            ? ev.data[0]
            : ev.data
          : ev
      callback(data)
    })
    unsubs.push(unsub)
    return unsub
  }

  // Cross-window suppression of external notifications during showcase
  onEvent('devtest-suppress-notifications', (data: any) => {
    const suppress = typeof data?.suppress === 'boolean' ? data.suppress : true
    suppressExternalNotifications.set(suppress)
  })

  // System Tray & Theme toggles
  onEvent('toggle-theme', () => {
    themeStore.update((t) => (t === 'dark-gold' ? 'nord-frost' : 'dark-gold'))
  })

  onEvent('show-pill', () => {
    callbacks.showWidget()
  })

  onEvent('open-onboarding', () => {
    hasCompletedOnboardingStore.set(false)
    callbacks.showWidget()
  })

  // Pre-warm developer announcements cache so HomePage renders at stable height
  void getDeveloperAnnouncements()

  onEvent('hotkey-toggle-island', () => {
    if (callbacks.onHotkeyToggle) {
      callbacks.onHotkeyToggle()
    } else if (callbacks.getIsHidden()) {
      callbacks.showWidget()
    } else {
      callbacks.hideWidget()
    }
  })

  onEvent('hotkey-quick-notes', () => {
    callbacks.onQuickNotesHotkey?.()
  })

  onEvent('hotkey-mode-switch', () => {
    callbacks.onModeSwitchHotkey?.()
  })

  // Track last auto-hide notification timestamp per application to prevent spamming
  const lastAutoHideNotifPerApp = new Map<string, number>()
  const AUTO_HIDE_NOTIF_COOLDOWN_MS = 5 * 60 * 1000 // 5 minutes

  // Fullscreen / Game Mode Handling
  let wasHiddenBeforeGameMode = false
  onEvent('game-mode-enter', (payload: unknown) => {
    if (callbacks.getIsHidden()) {
      wasHiddenBeforeGameMode = true
    } else {
      wasHiddenBeforeGameMode = false
      callbacks.hideWidget()
    }

    let reason = ''
    let appName = ''
    if (payload && typeof payload === 'object') {
      const p = payload as Record<string, unknown>
      reason = String(p.reason || '')
      appName = String(p.appName || '')
    } else if (typeof payload === 'string') {
      reason = payload
    }

    const targetAppName = appName || 'Vinta'
    const now = Date.now()
    const lastSent = lastAutoHideNotifPerApp.get(targetAppName) || 0

    // Only dispatch notification if cooldown has elapsed for this application
    if (now - lastSent >= AUTO_HIDE_NOTIF_COOLDOWN_MS) {
      lastAutoHideNotifPerApp.set(targetAppName, now)

      let reasonText = 'an exclusive fullscreen application'
      if (reason === 'presentation_mode') {
        reasonText = 'presentation mode'
      } else if (reason === 'unsupported_app') {
        reasonText = 'an incompatible application'
      } else if (reason === 'borderless_fullscreen') {
        reasonText = 'a borderless fullscreen application'
      }

      const notifTitle = 'Auto-Hidden'
      const notifMessage = appName
        ? `${appName}: Hidden to prevent overlay conflicts with ${reasonText}.`
        : `Vinta was hidden to prevent overlay conflicts with ${reasonText}.`

      // Prune previous auto-hide notification for this specific application before dispatching
      notifData.update((items) =>
        items.filter(
          (item) =>
            !(item.title === notifTitle && item.appName === targetAppName)
        )
      )

      sendVintaNotification({
        title: notifTitle,
        message: notifMessage,
        subApp: appName || undefined,
        playSound: false
      })
    }
  })

  onEvent('game-mode-exit', () => {
    if (!wasHiddenBeforeGameMode && callbacks.getIsHidden()) {
      callbacks.showWidget()
    }
  })

  let lastRelocateTime = 0
  onEvent('game-mode-relocate', (payload: unknown) => {
    const p = payload as {
      appName?: string
      targetOffset?: number
      targetMonitor?: MonitorInfo
      originalMonitor?: MonitorInfo
      reason?: string
    }
    const appName = p?.appName || ''
    const targetName = p?.targetMonitor?.name || 'Secondary Display'
    const targetOffset = p?.targetOffset ?? 0

    if (callbacks.onRelocateOffset) {
      callbacks.onRelocateOffset(targetOffset)
    } else {
      hiddenPillOffsetXStore.set(targetOffset)
    }

    const now = Date.now()
    if (now - lastRelocateTime > 5000) {
      lastRelocateTime = now
      const notifTitle = `Moved to ${targetName}`
      const notifMessage = appName
        ? `${appName}: Moved to ${targetName} to prevent overlay conflicts.`
        : `Moved to ${targetName} to prevent overlay conflicts.`

      sendVintaNotification({
        title: notifTitle,
        message: notifMessage,
        subApp: 'Display Manager'
      })
    }
  })

  onEvent('game-mode-exit-relocate', (payload: unknown) => {
    const p = payload as {
      restoreOffset?: number
      restoredMonitor?: string
    }

    if (callbacks.onRelocateOffset) {
      callbacks.onRelocateOffset(p?.restoreOffset ?? 0)
    } else if (p?.restoreOffset !== undefined) {
      hiddenPillOffsetXStore.set(p.restoreOffset)
    }
  })

  onEvent('active-monitor-follow', (payload: unknown) => {
    const p = payload as {
      targetOffset?: number
      targetMonitor?: MonitorInfo
    }
    const targetOffset = p?.targetOffset ?? 0
    if (callbacks.onRelocateOffset) {
      callbacks.onRelocateOffset(targetOffset)
    } else {
      hiddenPillOffsetXStore.set(targetOffset)
    }
    if (p?.targetMonitor) {
      displayLayoutStore.update((layout) => ({
        ...layout,
        currentMonitor: p.targetMonitor as MonitorInfo
      }))
    }
  })

  onEvent('display-layout-changed', () => {
    callbacks.onRefreshDisplay?.()
  })

  // Media Playback Updates
  let prevIsPlaying = false
  onEvent('media-update', (data: Record<string, unknown>) => {
    if (!data) return

    const title = ((data.title as string) || '').trim()
    const artist = ((data.artist as string) || '').trim()
    const hasValidTrackInfo = title.length > 0 || artist.length > 0
    const hasPlaybackControls =
      Boolean(data.canPlay) ||
      Boolean(data.canPause) ||
      Boolean(data.canGoNext) ||
      Boolean(data.canGoPrev)

    const isIncomingAvailable =
      data.isAvailable !== false && hasValidTrackInfo && hasPlaybackControls
    const isIncomingPlaying = Boolean(data.isPlaying) && isIncomingAvailable
    if (
      callbacks.getIsPageDormant() &&
      !isIncomingPlaying &&
      !get(globalMediaStore).isPlaying
    ) {
      return
    }

    globalMediaStore.update((m) => {
      const isLocked = Boolean(
        (m as unknown as Record<string, unknown>)._isPlayStateLocked
      )
      const incomingTitle = ((data.title as string) || '').trim()
      const incomingArtist = ((data.artist as string) || '').trim()
      const incomingArt = (data.albumArt as string) || ''
      const incomingSourceApp = (data.sourceApp as string) || ''
      const currentTitle = (m.title || '').trim()
      const currentArtist = (m.artist || '').trim()

      const isSameTrack =
        incomingTitle === currentTitle &&
        (incomingArtist === currentArtist ||
          !incomingArtist ||
          !currentArtist) &&
        (!incomingSourceApp ||
          !m.sourceApp ||
          incomingSourceApp === m.sourceApp)

      const resolvedAlbumArt = incomingArt
        ? incomingArt
        : isSameTrack && m.albumArt
          ? m.albumArt
          : isIncomingAvailable && incomingTitle === currentTitle && m.albumArt
            ? m.albumArt
            : ''

      if (!isIncomingAvailable) {
        return {
          ...m,
          isAvailable: false,
          isPlaying: false,
          _isPlayStateLocked: false
        }
      }

      const filteredSessions = Array.isArray(data.sessions)
        ? (data.sessions as Array<Record<string, unknown>>).filter((s) => {
            const sTitle = ((s.title as string) || '').trim()
            const sArtist = ((s.artist as string) || '').trim()
            return sTitle.length > 0 || sArtist.length > 0
          })
        : []

      return {
        ...(data as unknown as MediaData),
        sessions: filteredSessions as unknown as MediaSessionItem[],
        albumArt: resolvedAlbumArt,
        isPlaying: isLocked ? m.isPlaying : Boolean(data.isPlaying),
        _isPlayStateLocked: isLocked
      }
    })

    const isPlaying = isIncomingPlaying
    if (
      isPlaying &&
      !prevIsPlaying &&
      !callbacks.getIsExpanded() &&
      callbacks.getActivePage() === HomePage &&
      get(pageVisibilityStore).media &&
      get(hasCompletedOnboardingStore)
    ) {
      callbacks.navigateTo(MediaPage, mediaConfig, false)
    } else if (!isPlaying && prevIsPlaying) {
      callbacks.onHideTrigger()
    }
    prevIsPlaying = isPlaying
  })

  // Audio Peak Visualizer Feed
  onEvent('audio-peak', (data: Record<string, unknown>) => {
    if (
      callbacks.getIsHidden() ||
      callbacks.getIsPageDormant() ||
      callbacks.getActivePage() !== MediaPage ||
      get(pageVisibilityStore).media === false
    ) {
      return
    }
    if (data && typeof data.peak === 'number') {
      audioPeakStore.set(data.peak as number)
    }
  })

  // System Notifications
  onEvent(
    'notification',
    (payload: {
      id?: number
      appName?: string
      aumid?: string
      icon?: string
      appLogo?: string
      title?: string
      message?: string
      heroImage?: string
      isShowcase?: boolean
    }) => {
      const p = (Array.isArray(payload) ? payload[0] : payload) || {}
      if (get(suppressExternalNotifications) && !p.isShowcase) {
        return
      }

      const appName = p.appName || ''
      const aumid = p.aumid || ''
      const title = p.title || ''
      const message = p.message || ''

      const newNotif = {
        id: Date.now() + Math.floor(Math.random() * 1000),
        nativeId: p.id,
        appName,
        aumid,
        icon: p.icon,
        appLogo: p.appLogo || '',
        title,
        message,
        heroImage: p.heroImage,
        timestamp: Date.now()
      }

      const isVintaSelf =
        aumid.toLowerCase().includes('freematrixstudios.vinta') ||
        appName.toLowerCase().startsWith('vinta')

      let isDuplicate = false
      notifData.update((items) => {
        if (payload.id) {
          const idx = items.findIndex((item) => item.nativeId === payload.id)
          if (idx !== -1) {
            isDuplicate = true
            return items
          }
        }
        if (isVintaSelf) {
          const idx = items.findIndex(
            (item) =>
              item.title === title &&
              item.appName.toLowerCase().startsWith('vinta')
          )
          if (idx !== -1) {
            isDuplicate = true
            // If already added by in-app dispatch, update its nativeId rather than duplicating
            if (payload.id && !items[idx].nativeId) {
              items[idx] = { ...items[idx], nativeId: payload.id }
            }
            return items
          }
        }
        return [newNotif, ...items].slice(0, MAX_NOTIFICATIONS)
      })

      if (isDuplicate) {
        return
      }

      const isNotifEnabled = get(pageVisibilityStore).notifications
      const currentPage = callbacks.getActivePage()
      const isExpanded = callbacks.getIsExpanded()
      const mediaIsActive =
        currentPage === MediaPage && get(globalMediaStore).isPlaying

      const shouldShowSatellite =
        !get(dndStore) &&
        isNotifEnabled &&
        get(notifSatelliteEnabledStore) &&
        currentPage !== HomePage &&
        currentPage !== NotificationPage

      if (shouldShowSatellite) {
        showSatelliteNotification(newNotif)
      } else if (
        !get(dndStore) &&
        get(notifAutoWakeStore) &&
        isNotifEnabled &&
        !isExpanded &&
        !mediaIsActive &&
        get(hasCompletedOnboardingStore)
      ) {
        callbacks.onNotificationWake?.()
        callbacks.showWidget()
        callbacks.navigateTo(NotificationPage, notifConfig, false)
        callbacks.triggerGlow()
      }

      if (get(notifSoundStore) && !get(dndStore) && isNotifEnabled) {
        playNotifChime(get(notifVolumeStore))
      }
      if (isNotifEnabled) {
        callbacks.showWidget()
        callbacks.onHideTrigger()
      }
    }
  )

  onEvent(
    'notification_update',
    (payload: {
      id?: number
      appName?: string
      aumid?: string
      icon?: string
      appLogo?: string
      title?: string
      message?: string
      heroImage?: string
    }) => {
      if (!payload) return
      notifData.update((items) => {
        let updated = false
        const next = items.map((it) => {
          const matchesId = payload.id && it.nativeId === payload.id
          const matchesContent =
            !it.nativeId &&
            it.title === payload.title &&
            (it.appName === payload.appName || it.aumid === payload.aumid)
          if (matchesId || matchesContent) {
            updated = true
            return {
              ...it,
              appName: payload.appName || it.appName,
              aumid: payload.aumid || it.aumid,
              icon: payload.icon || it.icon,
              appLogo: payload.appLogo || it.appLogo,
              heroImage:
                payload.heroImage !== undefined
                  ? payload.heroImage
                  : it.heroImage
            }
          }
          return it
        })
        return updated ? next : items
      })
    }
  )

  onEvent('notification_permission_denied', () => {
    notificationPermissionDenied.set(true)
    if (get(pageVisibilityStore).notifications) {
      callbacks.showWidget()
      callbacks.navigateTo(NotificationPage, notifConfig, true)
    }
  })

  // Navigation to Settings via Custom DOM Event
  const handleNavigateSettings = (e: Event) => {
    const customEv = e as CustomEvent<{ tab?: string }>
    if (customEv.detail?.tab) {
      activeSettingsTabStore.set(customEv.detail.tab)
    }
    callbacks.showWidget()
    callbacks.navigateTo(SettingsPage, settingsConfig, true)
  }
  window.addEventListener('vinta-navigate-settings', handleNavigateSettings)

  // Application Updates
  onEvent('update:available', (data: UpdateCheckResult) => {
    if (data && data.updateAvailable) {
      updateInfoStore.set(data)
      if (get(autoCheckUpdatesStore)) {
        const notifs = get(notifData)
        notifData.set([
          {
            id: Date.now(),
            appName: 'Vinta System',
            title: `Vinta Update Available (${data.latestVersion})`,
            message:
              'A new version of Vinta is ready for download. Open Settings > Updates to review release highlights.',
            timestamp: Date.now(),
            canOpen: true,
            isUpdate: true,
            icon: appIcon,
            appLogo: appIcon
          },
          ...notifs
        ])
        callbacks.showWidget()
      }
    }
  })

  if (typeof CheckPostUpdateInstalled === 'function') {
    CheckPostUpdateInstalled()
      .then((newVer: string) => {
        if (newVer) {
          const notifs = get(notifData)
          notifData.set([
            {
              id: Date.now(),
              appName: 'Vinta System',
              title: `Welcome to Vinta ${newVer}!`,
              message:
                'Vinta has been successfully updated. Check out the Updates tab in Settings to see what is new.',
              timestamp: Date.now(),
              canOpen: true,
              isUpdate: true,
              icon: appIcon,
              appLogo: appIcon
            },
            ...notifs
          ])
          callbacks.showWidget()
        }
      })
      .catch(() => {})
  }

  // Monitor Display Geometry
  onEvent('monitor-changed', (mon: MonitorInfo) => {
    if (mon && mon.id) {
      callbacks.onMonitorChanged(mon)
    }
  })

  // Native File Drop into File Holding Shelf
  void initFileShelfStore()
  onEvent('vinta-files-dropped', (rawPayload: unknown) => {
    const droppedPaths: string[] = Array.isArray(rawPayload)
      ? rawPayload
      : typeof rawPayload === 'string'
        ? [rawPayload]
        : []
    if (droppedPaths.length === 0) return
    if (get(pageVisibilityStore).fileshelf === false) return
    void addFilesToShelf(droppedPaths).then((added) => {
      if (added.length > 0) {
        callbacks.showWidget()
        callbacks.triggerGlow()
        callbacks.navigateTo(FileShelfPage, fileShelfConfig, true)
      }
    })
  })

  // Immediate Media State Query
  if (typeof RequestMediaUpdate === 'function') {
    RequestMediaUpdate()
  }

  // Hardware Status Sync: keep battery/wireless active when visible (even minimized),
  // because WS_EX_NOACTIVATE windows rarely receive focus events and minimal Home renders WiFi/battery.
  let batteryCapabilityDiscovered = false
  const updateBattery = (forceDiscovery = false) => {
    if (get(isSimulatingBatteryStore)) return
    if (callbacks.getIsHidden() && !forceDiscovery) return
    const hwVis = get(homeWidgetsVisibilityStore)
    const pageVis = get(pageVisibilityStore)
    if (
      !forceDiscovery &&
      batteryCapabilityDiscovered &&
      !hwVis.battery &&
      !hwVis.actionCenter &&
      !pageVis.actioncenter
    ) {
      return
    }
    if (typeof GetBatteryStatus === 'function') {
      GetBatteryStatus()
        .then((status) => {
          batteryCapabilityDiscovered = true
          if (status && !get(isSimulatingBatteryStore)) {
            batteryDataStore.set(status as BatteryData)
          }
        })
        .catch(() => {})
    }
  }
  updateBattery(true)

  let cleanupBatteryEvents: (() => void) | undefined
  const nav = navigator as Navigator & {
    getBattery?: () => Promise<{
      addEventListener: (type: string, listener: () => void) => void
      removeEventListener: (type: string, listener: () => void) => void
    }>
  }
  if (typeof nav.getBattery === 'function') {
    nav
      .getBattery()
      .then((bat) => {
        const onBatChange = () => updateBattery()
        bat.addEventListener('levelchange', onBatChange)
        bat.addEventListener('chargingchange', onBatChange)
        cleanupBatteryEvents = () => {
          bat.removeEventListener('levelchange', onBatChange)
          bat.removeEventListener('chargingchange', onBatChange)
        }
      })
      .catch(() => {})
  }
  // Keep a lightweight poll even with BatteryManager so Windows Battery Saver changes still surface.
  const batteryInterval = window.setInterval(() => {
    if (!callbacks.getIsHidden()) {
      updateBattery()
    }
  }, 30000)

  const updateWireless = () => {
    if (callbacks.getIsHidden() || get(isSimulatingWirelessStore)) return
    const hwVis = get(homeWidgetsVisibilityStore)
    const pageVis = get(pageVisibilityStore)
    if (!hwVis.wireless && !hwVis.actionCenter && !pageVis.actioncenter) return
    if (typeof GetWirelessStatus === 'function') {
      GetWirelessStatus()
        .then((status) => {
          if (status && !get(isSimulatingWirelessStore)) {
            wirelessDataStore.set(status as unknown as WirelessData)
          }
        })
        .catch(() => {})
    }
  }
  updateWireless()

  const handleHardwareFocusOrResync = () => {
    if (!callbacks.getIsHidden()) {
      updateBattery()
      updateWireless()
    }
  }
  window.addEventListener('focus', handleHardwareFocusOrResync)

  const wirelessInterval = window.setInterval(() => {
    if (!callbacks.getIsHidden()) {
      updateWireless()
    }
  }, 20000)

  return () => {
    window.clearInterval(batteryInterval)
    cleanupBatteryEvents?.()
    window.clearInterval(wirelessInterval)
    window.removeEventListener('focus', handleHardwareFocusOrResync)
    window.removeEventListener(
      'vinta-navigate-settings',
      handleNavigateSettings
    )
    unsubs.forEach((unsub) => {
      try {
        unsub()
      } catch {
        // Ignore unsubscription error
      }
    })
    unsubs.length = 0
  }
}
