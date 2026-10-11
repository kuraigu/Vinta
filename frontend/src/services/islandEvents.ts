import type { Component } from 'svelte'
import { get } from 'svelte/store'
import { Events } from '@wailsio/runtime'
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
import { playNotifChime } from './audioService'
import {
  themeStore,
  globalMediaStore,
  audioPeakStore,
  notifData,
  notificationPermissionDenied,
  notifSoundStore,
  notifVolumeStore,
  dndStore,
  notifAutoWakeStore,
  pageVisibilityStore,
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
  type WirelessData
} from '../store'
import type { MonitorInfo } from '../../bindings/Vinta/backend/models.js'

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
  onNotificationWake?: () => void
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
    const unsub = Events.On(eventName, (ev: { data: T }) => {
      callback(ev?.data)
    })
    unsubs.push(unsub)
    return unsub
  }

  // System Tray & Theme toggles
  onEvent('toggle-theme', () => {
    themeStore.update((t) => (t === 'dark' ? 'light' : 'dark'))
  })

  onEvent('show-pill', () => {
    callbacks.showWidget()
  })

  // Fullscreen / Game Mode Handling
  let wasHiddenBeforeGameMode = false
  onEvent('game-mode-enter', (reason: unknown) => {
    if (callbacks.getIsHidden()) {
      wasHiddenBeforeGameMode = true
    } else {
      wasHiddenBeforeGameMode = false
      callbacks.hideWidget()
    }

    let reasonText = 'exclusive fullscreen application'
    if (typeof reason === 'string') {
      if (reason === 'presentation_mode') {
        reasonText = 'presentation mode'
      } else if (reason === 'borderless_fullscreen') {
        reasonText = 'borderless fullscreen application'
      }
    }

    const newNotif = {
      id: Date.now(),
      appName: 'Vinta',
      title: 'Auto-Hidden for Fullscreen',
      message: `Vinta was automatically hidden to prevent overlay conflicts during ${reasonText}.`,
      timestamp: Date.now()
    }

    notifData.update((items) => [newNotif, ...items].slice(0, 99))
  })

  onEvent('game-mode-exit', () => {
    if (!wasHiddenBeforeGameMode && callbacks.getIsHidden()) {
      callbacks.showWidget()
    }
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
    if (callbacks.getIsPageDormant() && !isIncomingPlaying) return

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
      get(pageVisibilityStore).media
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
      callbacks.getActivePage() !== MediaPage
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
    }) => {
      const appName = payload.appName || ''
      const aumid = payload.aumid || ''
      const title = payload.title || ''
      const message = payload.message || ''

      const newNotif = {
        id: Date.now() + Math.floor(Math.random() * 1000),
        nativeId: payload.id,
        appName,
        aumid,
        icon: payload.icon,
        appLogo: payload.appLogo || '',
        title,
        message,
        heroImage: payload.heroImage,
        timestamp: Date.now()
      }

      notifData.update((n) => [newNotif, ...n].slice(0, 99))

      const isNotifEnabled = get(pageVisibilityStore).notifications
      const mediaIsActive =
        callbacks.getActivePage() === MediaPage &&
        get(globalMediaStore).isPlaying

      if (
        !get(dndStore) &&
        get(notifAutoWakeStore) &&
        isNotifEnabled &&
        !callbacks.getIsExpanded() &&
        !mediaIsActive
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
            it.appName === payload.appName
          if (matchesId || matchesContent) {
            updated = true
            return {
              ...it,
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

  // Immediate Media State Query
  if (typeof RequestMediaUpdate === 'function') {
    RequestMediaUpdate()
  }

  // Periodic Hardware Status Polling
  const updateBattery = () => {
    if (get(isSimulatingBatteryStore)) return
    if (typeof GetBatteryStatus === 'function') {
      GetBatteryStatus()
        .then((status) => {
          if (status && !get(isSimulatingBatteryStore)) {
            batteryDataStore.set(status as BatteryData)
          }
        })
        .catch(() => {})
    }
  }
  updateBattery()
  const batteryInterval = window.setInterval(updateBattery, 8000)

  const updateWireless = () => {
    if (get(isSimulatingWirelessStore)) return
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
  const wirelessInterval = window.setInterval(updateWireless, 10000)

  return () => {
    window.clearInterval(batteryInterval)
    window.clearInterval(wirelessInterval)
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
