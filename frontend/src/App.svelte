<script lang="ts">
  import { onMount, setContext, type Component } from 'svelte'
  import { get } from 'svelte/store'
  import { fade } from 'svelte/transition'
  import {
    SetWidgetBounds,
    SetDragCapture,
    IsTestMode,
    FocusWindow,
    Quit,
    GetBatteryStatus,
    GetWirelessStatus,
    CheckPostUpdateInstalled,
    TrimMemory,
    SetIsHidden,
    GetDisplayLayout
  } from '../wailsjs/go/backend/App.js'
  import { EventsOn, LogFatal, LogInfo } from '../wailsjs/runtime/runtime.js'
  import NotificationPage, {
    config as notifConfig
  } from './pages/NotificationPage.svelte'
  import HomePage, { config as homeConfig } from './pages/HomePage.svelte'
  import MediaPage, { config as mediaConfig } from './pages/MediaPage.svelte'
  import NotesPage from './pages/NotesPage.svelte'
  import AppsPage from './pages/AppsPage.svelte'
  import SettingsPage from './pages/SettingsPage.svelte'
  import Dock from './components/Dock.svelte'
  import Clock from './components/Clock.svelte'
  import DevPill from './devtest/DevPill.svelte'
  import { AppControls } from './utils/navigation'
  import { playNotifChime } from './services/audioService'
  import {
    themeStore,
    shapeStore,
    autoHideStore,
    homeAutoHideStore,
    mediaAutoHideStore,
    autoHideDelayStore,
    autoHideDisplayStore,
    notifAutoWakeStore,
    globalMediaStore,
    audioPeakStore,
    type MediaData,
    notifData,
    notificationPermissionDenied,
    wakeMethodStore,
    dndStore,
    opacityMinimalStore,
    opacityExpandedStore,
    opacityHiddenStore,
    hiddenPillOffsetXStore,
    allowDragPositionStore,
    notifSoundStore,
    notifVolumeStore,
    glowIntensityStore,
    type IslandConfig,
    homeWidgetsVisibilityStore,
    pageVisibilityStore,
    calculateHomeMinimalWidth,
    batteryDataStore,
    isSimulatingBatteryStore,
    wirelessDataStore,
    isSimulatingWirelessStore,
    devPinnedStore,
    devTestingStore,
    activePageIdStore,
    updateInfoStore,
    autoCheckUpdatesStore,
    type UpdateCheckResult,
    type BatteryData,
    type WirelessData
  } from './store'
  import { debugBoundsOverlay, debugLogs } from './devtest/debugConfig'

  function syncDebugBoundsOverlay(enabled: boolean) {
    if (typeof document !== 'undefined') {
      if (enabled) {
        document.body.classList.add('debug-bounds-overlay')
      } else {
        document.body.classList.remove('debug-bounds-overlay')
      }
    }
  }
  $effect(() => {
    syncDebugBoundsOverlay($debugBoundsOverlay)
  })

  // Dynamic Routing: The active app that the island is displaying
  let activePage = $state<Component>(HomePage)
  let islandConfig = $state<IslandConfig>(homeConfig)

  let expanded = $state(false)
  let isBlinking = $state(false)

  function triggerBlink() {
    isBlinking = true
    window.setTimeout(() => {
      isBlinking = false
    }, 100)
  }

  let isHidden = $state(false)

  let activeOpacity = $derived(
    isHidden
      ? $opacityHiddenStore
      : expanded
        ? $opacityExpandedStore
        : $opacityMinimalStore
  )
  let isHovering = $state(false)
  let hideTimeout: number
  let hoverShowTimeout: number

  let isDraggingHidden = $state(false)
  let isInFlight = $state(false)
  let hiddenDragStartX = 0
  let hiddenDragStartOffset = 0
  let dragMovedDistance = 0
  let dragResetTimeout: number | undefined
  let inertiaSafetyTimeout: number | undefined
  let inertiaAnimFrame: number | undefined
  let dragHistory: Array<{ x: number; time: number }> = []

  let islandNode = $state<HTMLElement>()
  let devPillContainerNode = $state<HTMLElement>()
  let isTestMode = $state(false)

  let isTransitionGlow = $state(true)
  let glowTimeout: number

  let islandAnim = $state('0.32s cubic-bezier(0.2, 0.8, 0.2, 1)')

  function triggerGlow() {
    isTransitionGlow = true
    clearTimeout(glowTimeout)
    glowTimeout = window.setTimeout(() => {
      isTransitionGlow = false
    }, 3000)
  }

  let isTimeoutActive = $state(false)

  // Per-page auto-hide: each page can opt in/out independently
  let shouldCurrentPageAutoHide = $derived.by(() => {
    if (activePage === HomePage) return $homeAutoHideStore
    if (activePage === MediaPage) return $mediaAutoHideStore
    return $autoHideStore
  })

  let displayLayout = $state({
    virtualLeft: 0,
    virtualTop: 0,
    virtualWidth: 1920,
    virtualHeight: 1080,
    primaryCenter: 960,
    monitors: [
      {
        id: 1,
        name: 'Primary Display',
        left: 0,
        top: 0,
        width: 1920,
        height: 1080,
        isPrimary: true
      }
    ]
  })

  let hideTrigger = $state(0)

  function handleAutoHideAndRevert() {
    void hideTrigger
    clearTimeout(hideTimeout)
    isTimeoutActive = false

    if ($devPinnedStore) return

    if (!expanded && !isHidden && !isHovering) {
      if (
        shouldCurrentPageAutoHide ||
        (activePage !== HomePage && activePage !== MediaPage)
      ) {
        isTimeoutActive = true
        const delayMs = ($autoHideDelayStore || 5) * 1000
        hideTimeout = window.setTimeout(() => {
          isTimeoutActive = false
          if (shouldCurrentPageAutoHide) {
            hideWidget()
          }
          if (activePage !== HomePage && activePage !== MediaPage) {
            const fallback = AppControls.getFallbackPage()
            navigateTo(fallback.page as Component, fallback.config, false)
          }
        }, delayMs)
      }
    }
  }

  $effect(handleAutoHideAndRevert)

  let isPageDormant = $state(false)
  let dormantUnmountTimeout: number | undefined
  let memoryTrimTimeout: number | undefined

  // Manage OS bounds when hiding/unhiding
  function hideWidget() {
    if (isHidden) return
    islandAnim = '0.35s cubic-bezier(0.2, 0.8, 0.2, 1)'
    if (inertiaAnimFrame) {
      cancelAnimationFrame(inertiaAnimFrame)
      inertiaAnimFrame = undefined
    }
    clearTimeout(inertiaSafetyTimeout)
    isInFlight = false
    isDraggingHidden = false
    isHidden = true
    expanded = false
    notifyOSHidden(true)

    // Multi-tier resource management:
    // Allow the 0.22s CSS opacity fade-out to finish, then unmount the active page
    // to release DOM elements, images, canvases, and texture caches.
    clearTimeout(dormantUnmountTimeout)
    dormantUnmountTimeout = window.setTimeout(() => {
      if (isHidden) {
        isPageDormant = true
      }
    }, 240)

    // After 3 seconds in hidden mode, instruct the OS to reclaim unused physical memory
    clearTimeout(memoryTrimTimeout)
    memoryTrimTimeout = window.setTimeout(() => {
      if (isHidden) {
        trimOSMemory()
      }
    }, 3000)

    if (activePage !== HomePage && activePage !== MediaPage) {
      const fallback = AppControls.getFallbackPage()
      activePage = fallback.page as Component
      islandConfig = fallback.config
    }
  }

  // Global reactive fallback sync effect when island is minimized
  let mediaIdleTimeout: number | undefined

  function handlePageVisibilityFallback() {
    if ($devPinnedStore || $devTestingStore) return
    const pageVis = $pageVisibilityStore
    if (!pageVis.media && activePage === MediaPage) {
      navigateTo(HomePage, homeConfig, false)
    }
    if (!pageVis.notifications && activePage === NotificationPage) {
      navigateTo(HomePage, homeConfig, false)
    }
    if (!pageVis.notes && activePage === NotesPage) {
      navigateTo(HomePage, homeConfig, false)
    }
  }

  $effect(handlePageVisibilityFallback)

  function handleMediaIdleTransition() {
    if ($devPinnedStore || $devTestingStore) return
    const mediaState = $globalMediaStore
    const isMediaEnabled = $pageVisibilityStore.media
    const isMediaConnected = mediaState.isAvailable && isMediaEnabled

    if (!expanded && !isHidden) {
      if (isMediaConnected) {
        if (mediaIdleTimeout) {
          clearTimeout(mediaIdleTimeout)
          mediaIdleTimeout = undefined
        }
      } else {
        // Media has disconnected, stopped, or disabled - transition to HomePage
        if (activePage === MediaPage && !mediaIdleTimeout) {
          mediaIdleTimeout = window.setTimeout(() => {
            mediaIdleTimeout = undefined
            const currentMedia = get(globalMediaStore)
            const currentMediaEnabled = get(pageVisibilityStore).media
            if (
              (!currentMedia.isAvailable || !currentMediaEnabled) &&
              !expanded &&
              !isHidden &&
              activePage === MediaPage
            ) {
              navigateTo(HomePage, homeConfig, false)
            }
          }, 3000)
        }
      }
    } else {
      if (mediaIdleTimeout) {
        clearTimeout(mediaIdleTimeout)
        mediaIdleTimeout = undefined
      }
    }
  }

  $effect(handleMediaIdleTransition)

  function showWidget() {
    if (!isHidden) return
    clearTimeout(dormantUnmountTimeout)
    dormantUnmountTimeout = undefined
    clearTimeout(memoryTrimTimeout)
    memoryTrimTimeout = undefined
    isPageDormant = false

    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'
    if (inertiaAnimFrame) {
      cancelAnimationFrame(inertiaAnimFrame)
      inertiaAnimFrame = undefined
    }
    clearTimeout(inertiaSafetyTimeout)
    isInFlight = false
    isDraggingHidden = false
    isHidden = false
    notifyOSHidden(false)
    const { w, h, r } = getBoundsForState(islandConfig, expanded, false)
    applyBounds(w, h, r)
    if (
      !$devTestingStore &&
      $globalMediaStore.isPlaying &&
      $pageVisibilityStore.media &&
      activePage === HomePage &&
      !expanded
    ) {
      navigateTo(MediaPage, mediaConfig, false)
    }
    triggerGlow()
  }

  // Expose manual hide for the dock
  setContext('hidePill', () => {
    hideWidget()
  })

  // Expose manual wake up for components
  setContext('wakeupPill', showWidget)

  let isBoundsIpcPending = false
  let pendingBoundsArgs: {
    w: number
    h: number
    r: number
    offset?: number
  } | null = null

  function updateBounds(
    w: number,
    h: number,
    r: number,
    explicitOffset?: number
  ) {
    const computedW = Math.round(w)
    const computedH = Math.round(h)
    const computedR = Math.round($shapeStore === 'box' ? 12 : r)
    const scale = window.devicePixelRatio || 1
    const rawOffset =
      explicitOffset !== undefined
        ? explicitOffset
        : isNaN(clampedOffsetX)
          ? 0
          : clampedOffsetX
    const offset = Math.round(rawOffset)

    if (isBoundsIpcPending) {
      pendingBoundsArgs = { w: computedW, h: computedH, r: computedR, offset }
      return
    }

    if (get(debugLogs)) {
      console.log('[Vinta:Bounds]', {
        w: computedW,
        h: computedH,
        r: computedR,
        scale,
        offset
      })
    }

    isBoundsIpcPending = true
    Promise.resolve(
      SetWidgetBounds(computedW, computedH, computedR, scale, offset)
    )
      .catch((err) => {
        if (get(debugLogs)) {
          console.error('[Vinta:BoundsError]', err)
        }
      })
      .finally(() => {
        isBoundsIpcPending = false
        if (pendingBoundsArgs) {
          const next = pendingBoundsArgs
          pendingBoundsArgs = null
          updateBounds(next.w, next.h, next.r, next.offset)
        }
      })
  }

  function setOSDragCapture(enable: boolean) {
    if (typeof SetDragCapture === 'function') {
      SetDragCapture(enable).catch(() => {})
    }
  }

  function trimOSMemory() {
    if (typeof TrimMemory === 'function') {
      TrimMemory().catch(() => {})
    }
  }

  function notifyOSHidden(hidden: boolean) {
    if (typeof SetIsHidden === 'function') {
      SetIsHidden(hidden).catch(() => {})
    }
  }

  // --- Centralized OS Bounds Manager ---
  let currentOSBounds = { w: 0, h: 0, r: 0 }
  let targetPendingBounds = { w: 0, h: 0, r: 0 }
  let boundsTransitionTimeout: number | undefined

  function applyBounds(w: number, h: number, r: number) {
    currentOSBounds = { w, h, r }
    updateBounds(w, h, r, clampedOffsetX)
  }

  function handleIslandTransitionEnd(e: TransitionEvent) {
    if (
      e.target === e.currentTarget &&
      (e.propertyName === 'width' || e.propertyName === 'height')
    ) {
      clearTimeout(boundsTransitionTimeout)
      if (targetPendingBounds.w > 0 && targetPendingBounds.h > 0) {
        applyBounds(
          targetPendingBounds.w,
          targetPendingBounds.h,
          targetPendingBounds.r
        )
        targetPendingBounds = { w: 0, h: 0, r: 0 }
      }
    }
  }

  let hiddenBounds = $derived(
    $autoHideDisplayStore === 'clock'
      ? { w: 105, h: 24, r: 12 }
      : { w: 60, h: 8, r: 4 }
  )

  let hasBattery = $derived($batteryDataStore.hasBattery)

  let currentMinimalW = $derived(
    activePage === HomePage
      ? calculateHomeMinimalWidth($homeWidgetsVisibilityStore, hasBattery)
      : islandConfig.minimal.w
  )

  let currentIslandW = $derived(
    isHidden
      ? hiddenBounds.w
      : expanded
        ? islandConfig.expanded.w
        : currentMinimalW
  )

  let minAllowedOffsetForCurrentW = $derived.by(() => {
    const halfW = currentIslandW / 2
    return Math.floor(12 + halfW - displayLayout.primaryCenter)
  })

  let maxAllowedOffsetForCurrentW = $derived.by(() => {
    const halfW = currentIslandW / 2
    return Math.floor(
      displayLayout.virtualWidth - 12 - halfW - displayLayout.primaryCenter
    )
  })

  let minDevPillOffset = $derived.by(() => {
    const halfW = 430 / 2
    return Math.floor(12 + halfW - displayLayout.primaryCenter)
  })

  let maxDevPillOffset = $derived.by(() => {
    const halfW = 430 / 2
    return Math.floor(
      displayLayout.virtualWidth - 12 - halfW - displayLayout.primaryCenter
    )
  })

  let currentLiveOffsetX = $state(0)

  // Smart offset safety: smoothly clamp main island back into bounds if content expands near screen edge
  function handleSmartEdgeOffsetAdjustment() {
    if (!isDraggingHidden && !isInFlight) {
      if (currentLiveOffsetX > maxAllowedOffsetForCurrentW) {
        currentLiveOffsetX = maxAllowedOffsetForCurrentW
        $hiddenPillOffsetXStore = currentLiveOffsetX
      } else if (currentLiveOffsetX < minAllowedOffsetForCurrentW) {
        currentLiveOffsetX = minAllowedOffsetForCurrentW
        $hiddenPillOffsetXStore = currentLiveOffsetX
      }
    }
  }

  $effect(handleSmartEdgeOffsetAdjustment)

  // Sync store updates into live offset when not actively in flight
  function syncStoreOffsetToLive(val: number) {
    if (!isInFlight && !isDraggingHidden) {
      currentLiveOffsetX = val
    }
  }
  function handleHiddenPillOffsetSync() {
    syncStoreOffsetToLive($hiddenPillOffsetXStore)
  }
  $effect(handleHiddenPillOffsetSync)

  let clampedOffsetX = $derived.by(() => {
    if (Math.abs(currentLiveOffsetX) < 3) return 0
    return Math.max(
      minAllowedOffsetForCurrentW,
      Math.min(maxAllowedOffsetForCurrentW, currentLiveOffsetX)
    )
  })

  let devClampedOffsetX = $derived.by(() => {
    if (Math.abs(clampedOffsetX) < 3) return 0
    return Math.max(
      minDevPillOffset,
      Math.min(maxDevPillOffset, clampedOffsetX)
    )
  })

  let glowOpacity = $derived(
    $glowIntensityStore === 'disabled'
      ? 0
      : $glowIntensityStore === 'vibrant'
        ? 1.0
        : 0.6
  )

  let extraDropdownH = $state(0)
  let extraDropdownW = $state(0)
  let isQuitting = $state(false)

  function getBoundsForState(
    config: IslandConfig = islandConfig,
    isExp: boolean = expanded,
    isHid: boolean = isHidden
  ) {
    const baseW = isHid
      ? hiddenBounds.w
      : isExp
        ? config.expanded.w
        : config === homeConfig || activePage === HomePage
          ? calculateHomeMinimalWidth($homeWidgetsVisibilityStore, hasBattery)
          : config.minimal.w

    const baseH = isHid
      ? hiddenBounds.h
      : isExp
        ? config.expanded.h
        : config.minimal.h

    const w = baseW + extraDropdownW
    const h = baseH + extraDropdownH
    const r = isHid
      ? hiddenBounds.r
      : $shapeStore === 'box'
        ? 12
        : isExp
          ? config.expanded.r
          : config.minimal.r

    return { w, h, r }
  }

  function navigateTo(
    page: Component,
    pageConfig: IslandConfig,
    expand: boolean = false
  ) {
    if (isHidden) {
      clearTimeout(dormantUnmountTimeout)
      dormantUnmountTimeout = undefined
      clearTimeout(memoryTrimTimeout)
      memoryTrimTimeout = undefined
      isPageDormant = false
      isHidden = false
      notifyOSHidden(false)
      stopInertia()
    }
    if (activePage !== page && expanded === expand) {
      triggerBlink()
    }
    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'
    activePage = page
    islandConfig = pageConfig

    if (page === HomePage) $activePageIdStore = 'home'
    else if (page === MediaPage) $activePageIdStore = 'media'
    else if (page === NotificationPage) $activePageIdStore = 'notification'
    else if (page === NotesPage) $activePageIdStore = 'notes'
    else if (page === AppsPage) $activePageIdStore = 'apps'
    else if (page === SettingsPage) $activePageIdStore = 'settings'

    if (expand) {
      const { w, h, r } = getBoundsForState(pageConfig, true, false)
      if (w > currentOSBounds.w || h > currentOSBounds.h) {
        applyBounds(
          Math.max(w, currentOSBounds.w),
          Math.max(h, currentOSBounds.h),
          r
        )
      }
      requestAnimationFrame(() => {
        expanded = true
      })
    } else {
      expanded = false
    }
    triggerGlow()
  }

  // Global navigate for dev pill / pages
  setContext('navigate', navigateTo)

  onMount(() => {
    // Disable native browser context menu globally across the widget
    const disableGlobalContextMenu = (e: MouseEvent) => e.preventDefault()
    window.addEventListener('contextmenu', disableGlobalContextMenu)

    // Apply initial theme immediately to avoid layout styling flashes
    $themeStore = 'dark'
    document.documentElement.setAttribute('data-theme', 'dark')

    // Check if launched via `scripts.bat test`
    IsTestMode().then((testMode) => {
      isTestMode = testMode
    })

    // Reset offset to 0 on launch
    $hiddenPillOffsetXStore = 0
    currentLiveOffsetX = 0
    const { w, h, r } = getBoundsForState()
    currentOSBounds = { w, h, r }
    updateBounds(w, h, r, 0)

    const handleOverlayChange = (e: Event) => {
      const detail = (e as CustomEvent).detail as
        | { open: boolean; x?: number; y?: number; w?: number; h?: number }
        | undefined
      if (!detail?.open) {
        extraDropdownH = 0
        extraDropdownW = 0
        return
      }
      // Compute how much the overlay rect extends beyond the viewport edges.
      // The island window expands by this overhang so nothing is clipped.
      const { x = 0, y = 0, w = 0, h = 0 } = detail
      const margin = 8
      extraDropdownW = Math.max(0, x + w - window.innerWidth + margin)
      extraDropdownH = Math.max(0, y + h - window.innerHeight + margin)
    }
    const handleQuit = () => {
      isQuitting = true
      if ($notifSoundStore) {
        playNotifChime($notifVolumeStore * 0.4)
      }
      window.setTimeout(() => {
        Quit()
      }, 350)
    }

    window.addEventListener('vinta-overlay', handleOverlayChange)
    window.addEventListener('vinta-quit', handleQuit)

    // Listen to systray events
    EventsOn('toggle-theme', () => {
      themeStore.update((t) => (t === 'dark' ? 'light' : 'dark'))
    })
    EventsOn('toggle-shape', () => {
      shapeStore.update((s) => (s === 'pill' ? 'box' : 'pill'))
    })
    EventsOn('show-pill', showWidget)

    let prevIsPlaying = false
    EventsOn('media-update', (data: Record<string, unknown>) => {
      if (data) {
        const isIncomingPlaying = Boolean(data.isPlaying)
        if (isPageDormant && !isIncomingPlaying) return

        globalMediaStore.update((m) => {
          const isLocked = Boolean(
            (m as unknown as Record<string, unknown>)._isPlayStateLocked
          )
          const isIncomingAvailable = data.isAvailable !== false
          const incomingTitle = (data.title as string) || ''
          const incomingArt = (data.albumArt as string) || ''
          const isSameTrack =
            incomingTitle === m.title && data.artist === m.artist
          const resolvedAlbumArt = incomingArt
            ? incomingArt
            : isSameTrack && m.albumArt
              ? m.albumArt
              : ''

          if (!isIncomingAvailable) {
            // Media closed or disconnected: preserve last played metadata in paused state
            // during the grace period before transitioning back to HomePage.
            return {
              ...m,
              isAvailable: false,
              isPlaying: false,
              _isPlayStateLocked: false
            }
          }

          return {
            ...(data as unknown as MediaData),
            albumArt: resolvedAlbumArt,
            isPlaying: isLocked ? m.isPlaying : Boolean(data.isPlaying),
            _isPlayStateLocked: isLocked
          }
        })
        const isPlaying = Boolean(data.isPlaying)
        if (
          isPlaying &&
          !prevIsPlaying &&
          !expanded &&
          activePage === HomePage &&
          get(pageVisibilityStore).media
        ) {
          navigateTo(MediaPage, mediaConfig, false)
        }
        prevIsPlaying = isPlaying
      }
    })

    EventsOn('audio-peak', (data: Record<string, unknown>) => {
      if (isHidden || isPageDormant || activePage !== MediaPage) return
      if (data && typeof data.peak === 'number') {
        audioPeakStore.set(data.peak as number)
      }
    })

    EventsOn(
      'notification',
      (payload: {
        appName: string
        aumid: string
        icon: string
        title: string
        message: string
        heroImage: string
      }) => {
        const appName = payload.appName || ''
        const aumid = payload.aumid || ''
        const title = payload.title || ''
        const message = payload.message || ''

        const newNotif = {
          id: Date.now() + Math.floor(Math.random() * 1000),
          appName,
          aumid,
          icon: payload.icon,
          title,
          message,
          heroImage: payload.heroImage,
          timestamp: Date.now()
        }

        notifData.update((n) => {
          return [newNotif, ...n].slice(0, 99)
        })

        // Auto-wake & navigate to notification page when notification arrives
        const isNotifEnabled = get(pageVisibilityStore).notifications
        const mediaIsActive =
          activePage === MediaPage && $globalMediaStore.isPlaying
        if (
          !$dndStore &&
          $notifAutoWakeStore &&
          isNotifEnabled &&
          !expanded &&
          !mediaIsActive
        ) {
          showWidget()
          navigateTo(NotificationPage, notifConfig, false)
          triggerGlow()
        }

        if ($notifSoundStore && !$dndStore && isNotifEnabled) {
          playNotifChime($notifVolumeStore)
        }
        if (isNotifEnabled) {
          showWidget()
          hideTrigger++
        }
      }
    )

    EventsOn('notification_permission_denied', () => {
      notificationPermissionDenied.set(true)
      if (get(pageVisibilityStore).notifications) {
        showWidget()
        navigateTo(NotificationPage, notifConfig, true)
      }
    })

    EventsOn('update:available', (data: UpdateCheckResult) => {
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
              canOpen: false
            },
            ...notifs
          ])
          showWidget()
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
                canOpen: false
              },
              ...notifs
            ])
            showWidget()
          }
        })
        .catch(() => {})
    }

    const refreshDisplay = async () => {
      try {
        const layout = await GetDisplayLayout()
        if (layout && layout.virtualWidth > 0) {
          displayLayout = layout
        }
      } catch {
        // Keep current layout
      }
    }
    refreshDisplay()

    const handleResize = () => {
      refreshDisplay()
    }
    // Global failsafe: only abort if focus is truly lost to another application.
    // Internal blur events caused by window activation or clicks are safely ignored.
    const handleGlobalDragFailsafe = (e?: Event) => {
      if (e?.type === 'blur') {
        if (
          typeof document !== 'undefined' &&
          !document.hidden &&
          document.hasFocus()
        ) {
          return
        }
      }
      if (isDraggingHidden) {
        handleDragFailsafe()
      }
    }
    const handleWindowKeydownFailsafe = (e: KeyboardEvent) => {
      if (
        e.key === 'Escape' &&
        (isDraggingHidden || isInFlight || isTrackRegionActive)
      ) {
        handleDragFailsafe()
      }
    }
    window.addEventListener('resize', handleResize)
    window.addEventListener('blur', handleGlobalDragFailsafe)
    window.addEventListener('pointercancel', handleGlobalDragFailsafe)
    window.addEventListener('keydown', handleWindowKeydownFailsafe)

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
      stopInertia()
      window.clearInterval(batteryInterval)
      window.clearInterval(wirelessInterval)
      window.removeEventListener('contextmenu', disableGlobalContextMenu)
      window.removeEventListener('vinta-overlay', handleOverlayChange)
      window.removeEventListener('vinta-quit', handleQuit)
      window.removeEventListener('resize', handleResize)
      window.removeEventListener('blur', handleGlobalDragFailsafe)
      window.removeEventListener('pointercancel', handleGlobalDragFailsafe)
      window.removeEventListener('keydown', handleWindowKeydownFailsafe)
    }
  })

  let prevExpandedState = false
  let prevHiddenState = false

  /**
   * Synchronizes OS window bounds when island state changes.
   * In-flight drags skip this to prevent fighting the full-width track region,
   * but the drag handlers call stopInertia() before any state changes.
   */
  let hasInitializedBounds = false

  function syncOSBoundsOnStateChange() {
    if (isDraggingHidden || isInFlight || isTrackRegionActive) return
    const target = getBoundsForState()

    if (!hasInitializedBounds) {
      hasInitializedBounds = true
      currentOSBounds = { w: target.w, h: target.h, r: target.r }
      prevExpandedState = expanded
      prevHiddenState = isHidden
      return
    }

    const wasExpanded = prevExpandedState
    const wasHidden = prevHiddenState
    prevExpandedState = expanded
    prevHiddenState = isHidden

    const isCollapsing = wasExpanded && !expanded
    const isHiding = !wasHidden && isHidden
    const isShrinking =
      isCollapsing ||
      isHiding ||
      currentOSBounds.w > target.w ||
      currentOSBounds.h > target.h
    const isGrowing =
      currentOSBounds.w < target.w || currentOSBounds.h < target.h

    if (isGrowing) {
      clearTimeout(boundsTransitionTimeout)
      targetPendingBounds = { w: 0, h: 0, r: 0 }
      const immediateW = Math.max(currentOSBounds.w, target.w)
      const immediateH = Math.max(currentOSBounds.h, target.h)
      applyBounds(immediateW, immediateH, target.r)
    }

    if (isShrinking) {
      targetPendingBounds = target
      clearTimeout(boundsTransitionTimeout)
      boundsTransitionTimeout = window.setTimeout(() => {
        if (targetPendingBounds.w > 0 && targetPendingBounds.h > 0) {
          applyBounds(
            targetPendingBounds.w,
            targetPendingBounds.h,
            targetPendingBounds.r
          )
          targetPendingBounds = { w: 0, h: 0, r: 0 }
        }
      }, 450)
    } else if (!isGrowing) {
      clearTimeout(boundsTransitionTimeout)
      targetPendingBounds = { w: 0, h: 0, r: 0 }
      applyBounds(target.w, target.h, target.r)
    }
  }

  $effect(syncOSBoundsOnStateChange)

  let isTrackRegionActive = false
  let dragWatchdogTimeout: number | undefined

  // Resets the drag inactivity watchdog timer. If pointer interaction stalls
  // during a drag gesture, the watchdog automatically restores tight resting bounds.
  function resetDragWatchdog() {
    clearTimeout(dragWatchdogTimeout)
    dragWatchdogTimeout = window.setTimeout(() => {
      if (isDraggingHidden || isInFlight || isTrackRegionActive) {
        handleDragFailsafe()
      }
    }, 6000)
  }

  // Failsafe handler to cleanly cancel any drag operation, release OS mouse capture,
  // and guarantee the Win32 window clipping region escapes 100% width mode.
  function handleDragFailsafe() {
    clearTimeout(dragWatchdogTimeout)
    dragWatchdogTimeout = undefined
    isDraggingHidden = false
    setOSDragCapture(false)
    stopInertia()
  }

  function stopInertia() {
    if (inertiaAnimFrame) {
      cancelAnimationFrame(inertiaAnimFrame)
      inertiaAnimFrame = undefined
    }
    clearTimeout(inertiaSafetyTimeout)
    inertiaSafetyTimeout = undefined
    clearTimeout(dragWatchdogTimeout)
    dragWatchdogTimeout = undefined
    isInFlight = false
    isTrackRegionActive = false
    dragHistory = []
    setOSDragCapture(false)

    // Persist final resting offset to store once motion is fully settled
    $hiddenPillOffsetXStore = clampedOffsetX

    // Commit final exact tight OS window bounds at resting position
    const { w, h, r } = getBoundsForState()
    updateBounds(w, h, r, clampedOffsetX)
    currentOSBounds = { w, h, r }
  }

  let hasPointerCaptured = false

  function handleHiddenPointerDown(e: PointerEvent) {
    if (e.button !== 0) return // Only primary (left) button can initiate dragging
    if (!$allowDragPositionStore) return

    const target = e.target as HTMLElement
    const isDragHandle = Boolean(
      target.closest('.island-drag-handle, .dock-drag-handle')
    )
    const isAltDrag = e.altKey

    // When expanded, only allow dragging via the top drag handle, dock handle, or Alt+drag
    if (expanded && !isDragHandle && !isAltDrag) return

    if (
      !isDragHandle &&
      target.closest(
        'button, input, select, textarea, a, .custom-select, .no-drag, [role="button"]'
      )
    ) {
      return
    }

    if (inertiaAnimFrame) {
      stopInertia()
    }
    setOSDragCapture(true)
    isDraggingHidden = true
    hasPointerCaptured = false
    dragMovedDistance = 0
    clearTimeout(dragResetTimeout)
    resetDragWatchdog()
    hiddenDragStartX = e.screenX
    hiddenDragStartOffset = currentLiveOffsetX
    dragHistory = [{ x: e.screenX, time: performance.now() }]
  }

  function handleHiddenPointerMove(e: PointerEvent) {
    if (!isDraggingHidden) return
    // If no mouse buttons are pressed (e.g. released outside/interrupted), cancel drag immediately
    if (e.buttons === 0) {
      handleHiddenPointerUp(e)
      return
    }

    resetDragWatchdog()

    const now = performance.now()
    dragHistory.push({ x: e.screenX, time: now })
    if (dragHistory.length > 25) {
      dragHistory.shift()
    }

    const dx = e.screenX - hiddenDragStartX
    dragMovedDistance = Math.abs(dx)
    if (dragMovedDistance > 6) {
      if (!hasPointerCaptured) {
        hasPointerCaptured = true
        try {
          ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
        } catch {
          // Ignore
        }
      }
      const rawOffset = hiddenDragStartOffset + dx
      let targetOffset = rawOffset

      // Magnetic snap when near any monitor center (within ±28px)
      if (displayLayout.monitors && displayLayout.monitors.length > 0) {
        for (const mon of displayLayout.monitors) {
          const monClientLeft = mon.left - displayLayout.virtualLeft
          const monCenter = monClientLeft + mon.width / 2
          const monSnapOffset = monCenter - displayLayout.primaryCenter
          if (Math.abs(rawOffset - monSnapOffset) <= 28) {
            targetOffset = monSnapOffset
            break
          }
        }
      }

      currentLiveOffsetX = Math.max(
        minAllowedOffsetForCurrentW,
        Math.min(maxAllowedOffsetForCurrentW, targetOffset)
      )

      // Dynamic OnDrag window bounding region management:
      // When the cursor is physically inside the horizontal bounds of the island,
      // expand the Win32 window clipping region to 100% width (-1) to provide unlimited
      // horizontal runway and prevent cursor outrunning.
      // When the cursor leaves the island bounds (e.g. edge clamped or fast flick),
      // dynamically update bounds to track and rubberband with the island.
      let isCursorInside: boolean
      if (islandNode) {
        const rect = islandNode.getBoundingClientRect()
        const isInsideIsland = e.clientX >= rect.left && e.clientX <= rect.right
        let isInsideDevPill = false
        if (devPillContainerNode) {
          const devRect = devPillContainerNode.getBoundingClientRect()
          isInsideDevPill =
            e.clientX >= devRect.left && e.clientX <= devRect.right
        }
        isCursorInside = isInsideIsland || isInsideDevPill
      } else {
        const islandCenterX = displayLayout.primaryCenter + clampedOffsetX
        const halfW = currentIslandW / 2
        isCursorInside =
          e.screenX >= islandCenterX - halfW &&
          e.screenX <= islandCenterX + halfW
      }

      if (isCursorInside) {
        if (!isTrackRegionActive) {
          isTrackRegionActive = true
          const { h, r } = getBoundsForState()
          updateBounds(-1, h, r, clampedOffsetX)
        }
      } else {
        if (isTrackRegionActive) {
          isTrackRegionActive = false
        }
        // Symmetrical motion margin: provides breathing room on both sides of the moving island
        // to prevent the leading edge from clipping due to rendering/IPC phase difference.
        const motionMargin = 40
        const { w, h, r } = getBoundsForState()
        updateBounds(w + motionMargin, h, r, clampedOffsetX)
      }
    }
  }

  function handleHiddenPointerUp(e: PointerEvent) {
    if (!isDraggingHidden) return
    isDraggingHidden = false
    clearTimeout(dragWatchdogTimeout)
    dragWatchdogTimeout = undefined
    setOSDragCapture(false)
    if (hasPointerCaptured) {
      hasPointerCaptured = false
      try {
        ;(e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId)
      } catch {
        // Ignore
      }
    }

    if (dragMovedDistance > 6) {
      const now = performance.now()
      dragHistory.push({ x: e.screenX, time: now })

      // Sample points within the last 150ms of release
      const recent = dragHistory.filter((p) => now - p.time <= 150)
      let velocity = 0

      if (recent.length >= 2) {
        const oldest = recent[0]
        const newest = recent[recent.length - 1]
        const dt = newest.time - oldest.time
        const dx = newest.x - oldest.x
        if (dt > 15) {
          velocity = dx / dt // true px/ms
        }
      }

      if (Math.abs(velocity) >= 0.08) {
        const maxVelocity = 2.2 // px/ms
        velocity = Math.max(-maxVelocity, Math.min(maxVelocity, velocity))
        isInFlight = true

        const friction = 0.0035

        clearTimeout(inertiaSafetyTimeout)
        inertiaSafetyTimeout = window.setTimeout(stopInertia, 1200)

        let lastFrameTime = performance.now()

        function runInertiaStep(currentTime: number) {
          const frameDt = Math.min(32, currentTime - lastFrameTime)
          lastFrameTime = currentTime

          if (frameDt > 0) {
            const decay = Math.exp(-friction * frameDt)
            velocity *= decay
            currentLiveOffsetX += velocity * frameDt

            // Edge reached: hard clamp with zero bounce and immediately stop
            if (currentLiveOffsetX >= maxAllowedOffsetForCurrentW) {
              currentLiveOffsetX = maxAllowedOffsetForCurrentW
              stopInertia()
              return
            }
            if (currentLiveOffsetX <= minAllowedOffsetForCurrentW) {
              currentLiveOffsetX = minAllowedOffsetForCurrentW
              stopInertia()
              return
            }

            // Snap to rest when velocity decays
            if (Math.abs(velocity) < 0.035) {
              stopInertia()
              return
            }

            // Dynamic OnDrag: expand bounds with velocity-based margin to prevent leading-edge clipping
            const motionMargin = Math.min(
              60,
              Math.round(Math.abs(velocity) * 20) + 30
            )
            const { w, h, r } = getBoundsForState()
            updateBounds(w + motionMargin, h, r, clampedOffsetX)
          }

          inertiaAnimFrame = requestAnimationFrame(runInertiaStep)
        }

        inertiaAnimFrame = requestAnimationFrame(runInertiaStep)
      } else {
        stopInertia()
      }

      clearTimeout(dragResetTimeout)
      dragResetTimeout = window.setTimeout(() => {
        dragMovedDistance = 0
      }, 150)
    } else {
      dragMovedDistance = 0
      stopInertia()
    }
  }

  function handleContextMenu(e: MouseEvent) {
    if (isDraggingHidden || dragMovedDistance > 6 || isInFlight) {
      isDraggingHidden = false
      dragMovedDistance = 0
      stopInertia()
      e.preventDefault()
      e.stopPropagation()
      return
    }
    if (expanded) return

    const target = e.target as HTMLElement
    if (
      target.closest(
        'button, input, select, textarea, a, .custom-select, [role="button"]'
      )
    ) {
      return
    }
    e.preventDefault()
    if (isHidden) {
      showWidget()
    } else {
      hideWidget()
    }
  }

  function handleExpand() {
    if (expanded) return
    // Cancel any active drag/inertia first so the OS region update isn't blocked
    if (isInFlight || isDraggingHidden) stopInertia()
    triggerBlink()
    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'

    // Immediately open expanded OS bounds synchronously before starting CSS expansion
    const { w, h, r } = getBoundsForState(islandConfig, true, isHidden)
    applyBounds(w, h, r)

    requestAnimationFrame(() => {
      expanded = true
    })
  }

  function handleIslandClick(e: MouseEvent) {
    if (dragMovedDistance > 6 || isInFlight) {
      dragMovedDistance = 0
      e.preventDefault()
      e.stopPropagation()
      return
    }
    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'

    if (isHidden) {
      showWidget()
      return
    }
    if (!expanded) {
      handleExpand()
    }
  }

  function handleMouseEnter() {
    isHovering = true
    clearTimeout(hideTimeout)
    if (isHidden && $wakeMethodStore === 'hover') {
      hoverShowTimeout = window.setTimeout(showWidget, 150)
    } else if (!isHidden) {
      showWidget()
    }
  }

  function handleMouseLeave() {
    isHovering = false
    clearTimeout(hoverShowTimeout)
    hideTrigger++
  }

  function handleFocusIn(e: FocusEvent) {
    const target = e.target as HTMLElement
    if (
      target &&
      (target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.isContentEditable)
    ) {
      FocusWindow().catch(() => {})
    }
  }

  function handleGlobalPointerDown() {
    FocusWindow().catch(() => {})
  }
</script>

<!-- ISLAND DIV HERE -->
<div
  class="island-container"
  role="application"
  onfocusin={handleFocusIn}
  onpointerdown={handleGlobalPointerDown}
  style="position: absolute; left: {displayLayout.primaryCenter}px; transform: translateX(-50%);"
>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    bind:this={islandNode}
    class="island {expanded ? 'expanded' : ''} {isTransitionGlow
      ? 'is-transition-glow'
      : ''} {isHidden ? 'is-hidden' : ''} {isQuitting ? 'is-quitting' : ''}"
    onmouseenter={handleMouseEnter}
    onmouseleave={handleMouseLeave}
    onpointerdown={handleHiddenPointerDown}
    onpointermove={handleHiddenPointerMove}
    onpointerup={handleHiddenPointerUp}
    onpointercancel={handleHiddenPointerUp}
    oncontextmenu={handleContextMenu}
    style="
      transform: translateX({clampedOffsetX}px) {isQuitting
      ? 'scale(0.6) translateY(-20px)'
      : ''};
      opacity: {isQuitting ? 0 : activeOpacity};
      filter: {isQuitting ? 'blur(12px)' : 'none'};
      width: {isHidden
      ? hiddenBounds.w
      : expanded
        ? islandConfig.expanded.w
        : currentMinimalW}px;
      height: {isHidden
      ? hiddenBounds.h
      : expanded
        ? islandConfig.expanded.h
        : islandConfig.minimal.h}px;
      --island-r: {isHidden
      ? hiddenBounds.r
      : $shapeStore === 'box'
        ? 12
        : expanded
          ? islandConfig.expanded.r
          : islandConfig.minimal.r}px;
      border-radius: var(--island-r);
      --island-border-color: {isHidden
      ? 'var(--c-border)'
      : islandConfig.borderColor || 'transparent'};
      --island-glow-color: {isHidden
      ? 'transparent'
      : islandConfig.glowColor || 'transparent'};
      --island-hidden-glow: {islandConfig.glowColor ||
      'rgba(255, 255, 255, 0.6)'};
      --island-glow-opacity: {glowOpacity};
      --island-anim: {islandAnim};
    "
    onclick={handleIslandClick}
    ontransitionend={handleIslandTransitionEnd}
  >
    {#if isHidden}
      {#if $autoHideDisplayStore === 'clock'}
        <div class="hidden-widget-pill" transition:fade={{ duration: 180 }}>
          <Clock expanded={false} />
          <div class="hidden-dash-mini"></div>
        </div>
      {:else}
        <div class="hidden-dash" transition:fade={{ duration: 180 }}></div>
      {/if}
    {/if}

    {#if expanded && $allowDragPositionStore}
      <div
        class="island-drag-handle"
        title="Drag to reposition (or Alt + Drag)"
      >
        <div class="drag-pill-bar"></div>
      </div>
    {/if}

    <div
      class="content"
      style="opacity: {isHidden || isBlinking
        ? 0
        : 1}; pointer-events: {isHidden
        ? 'none'
        : 'auto'}; transition: opacity {isBlinking
        ? '0s'
        : '0.22s cubic-bezier(0.2, 0.8, 0.2, 1)'};"
    >
      {#if !isPageDormant}
        {#key [activePage, expanded]}
          {@const ActivePage = activePage}
          <div
            class="view"
            style="width: {expanded
              ? islandConfig.expanded.w
              : currentMinimalW}px; height: {expanded
              ? islandConfig.expanded.h
              : islandConfig.minimal.h}px;"
          >
            {#if ActivePage}
              <ActivePage {expanded} />
            {/if}
          </div>
        {/key}
      {/if}

      {#if expanded}
        <div in:fade={{ duration: 200, delay: 150 }}>
          <Dock {activePage} />
        </div>
      {/if}
    </div>

    <!-- Shrinking Timeout Bar -->
    {#if isTimeoutActive && (shouldCurrentPageAutoHide || (activePage !== HomePage && activePage !== MediaPage)) && !expanded && !isHidden}
      {#key hideTrigger}
        <div
          class="timeout-bar"
          style="animation-duration: {$autoHideDelayStore || 5}s;"
        ></div>
      {/key}
    {/if}
  </div>

  {#if isTestMode}
    <div
      bind:this={devPillContainerNode}
      style="
        transform: translateX({devClampedOffsetX}px);
        transition: {isDraggingHidden || isInFlight ? 'none' : islandAnim};
      "
    >
      <DevPill />
    </div>
  {/if}
</div>

<style src="./App.css"></style>
