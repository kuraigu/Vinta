<script lang="ts">
  import { onMount, setContext, type Component } from 'svelte'
  import { get } from 'svelte/store'
  import { fade } from 'svelte/transition'
  import {
    SetWidgetBounds,
    IsTestMode,
    FocusWindow,
    Quit,
    GetBatteryStatus
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
    notifAutoHideStore,
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
    devPinnedStore,
    devTestingStore,
    activePageIdStore,
    type BatteryData
  } from './store'
  import { preloadIcons, findAppIconConflicts } from './appIcons'
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
    if (activePage === NotificationPage) return $notifAutoHideStore
    return $autoHideStore
  })

  let screenWidth = $state(
    typeof window !== 'undefined'
      ? window.screen?.width || window.innerWidth || 1920
      : 1920
  )

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
    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'
    if (inertiaAnimFrame) {
      cancelAnimationFrame(inertiaAnimFrame)
      inertiaAnimFrame = undefined
    }
    clearTimeout(inertiaSafetyTimeout)
    isInFlight = false
    isDraggingHidden = false
    isHidden = false
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
      boundsTransitionTimeout = window.setTimeout(() => {
        if (targetPendingBounds.w > 0 && targetPendingBounds.h > 0) {
          applyBounds(
            targetPendingBounds.w,
            targetPendingBounds.h,
            targetPendingBounds.r
          )
          targetPendingBounds = { w: 0, h: 0, r: 0 }
        }
      }, 25)
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

  let maxAllowedOffsetForCurrentW = $derived(
    Math.max(0, Math.floor((screenWidth - currentIslandW) / 2 - 12))
  )

  let maxDevPillOffset = $derived(
    Math.max(0, Math.floor((screenWidth - 430) / 2 - 12))
  )

  let currentLiveOffsetX = $state(0)

  // Smart offset safety: smoothly clamp main island back into bounds if content expands near screen edge
  function handleSmartEdgeOffsetAdjustment() {
    const maxOffset = maxAllowedOffsetForCurrentW
    if (
      !isDraggingHidden &&
      !isInFlight &&
      Math.abs(currentLiveOffsetX) > maxOffset
    ) {
      currentLiveOffsetX = Math.sign(currentLiveOffsetX) * maxOffset
      $hiddenPillOffsetXStore = currentLiveOffsetX
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
      -maxAllowedOffsetForCurrentW,
      Math.min(maxAllowedOffsetForCurrentW, currentLiveOffsetX)
    )
  })

  let devClampedOffsetX = $derived.by(() => {
    if (Math.abs(clampedOffsetX) < 3) return 0
    return Math.max(
      -maxDevPillOffset,
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
      isHidden = false
      stopInertia()
    }
    if (activePage !== page && expanded === expand) {
      triggerBlink()
    }
    if (expanded !== expand) {
      islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'
    }
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
        applyBounds(w, h, r)
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

    // Preload SVG caches in the background after initial UI has rendered and settled
    if (typeof window.requestIdleCallback === 'function') {
      window.requestIdleCallback(() => preloadIcons(), { timeout: 3000 })
    } else {
      window.setTimeout(() => preloadIcons(), 2000)
    }

    // Check if launched via `scripts.bat test`
    IsTestMode().then((testMode) => {
      isTestMode = testMode
      if (isTestMode) {
        const conflicts = findAppIconConflicts()
        if (conflicts.length > 0) {
          LogFatal('APP ICON CONFLICTS FOUND:\n' + conflicts.join('\n'))
        } else {
          LogInfo('APP ICON TEST PASSED: No conflicts found.')
        }
      }
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
      if (data && typeof data.peak === 'number') {
        audioPeakStore.set(data.peak as number)
      }
    })

    EventsOn(
      'notification',
      (payload: {
        appName: string
        icon: string
        title: string
        message: string
        heroImage: string
      }) => {
        const appName = payload.appName || ''
        const title = payload.title || ''
        const message = payload.message || ''

        const newNotif = {
          id: Date.now() + Math.floor(Math.random() * 1000),
          appName,
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

    const handleResize = () => {
      screenWidth = window.screen?.width || window.innerWidth || 1920
    }
    const handleGlobalBlur = () => {
      if (isDraggingHidden) {
        isDraggingHidden = false
        stopInertia()
      }
    }
    window.addEventListener('resize', handleResize)
    window.addEventListener('blur', handleGlobalBlur)

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

    return () => {
      stopInertia()
      window.clearInterval(batteryInterval)
      window.removeEventListener('contextmenu', disableGlobalContextMenu)
      window.removeEventListener('vinta-overlay', handleOverlayChange)
      window.removeEventListener('vinta-quit', handleQuit)
      window.removeEventListener('resize', handleResize)
      window.removeEventListener('blur', handleGlobalBlur)
    }
  })

  /**
   * Synchronizes OS window bounds when island state changes.
   * In-flight drags skip this to prevent fighting the full-width track region,
   * but the drag handlers call stopInertia() before any state changes.
   */
  function syncOSBoundsOnStateChange() {
    if (isInFlight) return
    const target = getBoundsForState()

    // When shrinking (e.g. collapsing/minimizing from expanded, or hiding), hold window bounds until CSS transition finishes
    if (currentOSBounds.h > target.h || currentOSBounds.w > target.w) {
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
      }, 380)
    } else {
      clearTimeout(boundsTransitionTimeout)
      targetPendingBounds = { w: 0, h: 0, r: 0 }
      applyBounds(target.w, target.h, target.r)
    }
  }

  $effect(syncOSBoundsOnStateChange)

  function stopInertia() {
    if (inertiaAnimFrame) {
      cancelAnimationFrame(inertiaAnimFrame)
      inertiaAnimFrame = undefined
    }
    clearTimeout(inertiaSafetyTimeout)
    inertiaSafetyTimeout = undefined
    isInFlight = false
    dragHistory = []

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
    isDraggingHidden = true
    hasPointerCaptured = false
    dragMovedDistance = 0
    clearTimeout(dragResetTimeout)
    hiddenDragStartX = e.screenX
    hiddenDragStartOffset = currentLiveOffsetX
    dragHistory = [{ x: e.screenX, time: performance.now() }]
  }

  function handleHiddenPointerMove(e: PointerEvent) {
    if (!isDraggingHidden) return
    // If no mouse buttons are pressed (e.g. released outside/interrupted), cancel drag
    if (e.buttons === 0) {
      handleHiddenPointerUp(e)
      return
    }

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
      // Hard clamp: No rubberband past the edge
      currentLiveOffsetX = Math.max(
        -maxAllowedOffsetForCurrentW,
        Math.min(maxAllowedOffsetForCurrentW, rawOffset)
      )

      // Keep OS click region tightly aligned around current drag position
      const { w, h, r } = getBoundsForState()
      updateBounds(w, h, r, currentLiveOffsetX)
    }
  }

  function handleHiddenPointerUp(e: PointerEvent) {
    if (!isDraggingHidden) return
    isDraggingHidden = false
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

        const { w: islandW, h, r } = getBoundsForState()
        const friction = 0.0035

        // Initial glide corridor with compact lead comfort padding
        const initialLead = Math.max(-20, Math.min(20, velocity * 15))
        updateBounds(
          islandW + Math.abs(initialLead) + 24,
          h,
          r,
          currentLiveOffsetX + initialLead / 2
        )

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
            if (currentLiveOffsetX <= -maxAllowedOffsetForCurrentW) {
              currentLiveOffsetX = -maxAllowedOffsetForCurrentW
              stopInertia()
              return
            }

            // Snap to rest when velocity decays
            if (Math.abs(velocity) < 0.035) {
              stopInertia()
              return
            }

            // Dynamic corridor: island width + lead buffer + compact comfort width padding to prevent clipping
            const lead = Math.max(-20, Math.min(20, velocity * 15))
            const dynamicW = islandW + Math.abs(lead) + 24
            const dynamicOffset = currentLiveOffsetX + lead / 2
            updateBounds(dynamicW, h, r, dynamicOffset)
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
      {#key activePage}
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

      {#if expanded}
        <div in:fade={{ duration: 200, delay: 150 }}>
          <Dock {activePage} />
        </div>
      {/if}
    </div>

    <!-- Shrinking Timeout Bar -->
    {#if isTimeoutActive && shouldCurrentPageAutoHide && !expanded && !isHidden}
      {#key hideTrigger}
        <div
          class="timeout-bar"
          style="animation-duration: {$autoHideDelayStore || 10}s;"
        ></div>
      {/key}
    {/if}
  </div>

  {#if isTestMode}
    <div
      style="
        transform: translateX({devClampedOffsetX}px);
        transition: {isDraggingHidden || isInFlight ? 'none' : islandAnim};
      "
    >
      <DevPill />
    </div>
  {/if}
</div>

<style>
  /* Remove translateZ from root to prevent transparent viewport memory bloat */
  :global(html) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    height: 100vh;
    overflow: hidden;
    user-select: none;
    -webkit-user-select: none;
    font-family:
      'Inter',
      system-ui,
      -apple-system,
      sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  .island-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 15px;
    position: relative;
  }

  .island {
    --island-border-color: transparent;
    --island-glow-color: transparent;
    --island-border-width: 1.5px;
    --island-glow-width: 1.5px;
    --ambient-glow-ring:
      inset 0 0 0 var(--island-glow-width) var(--island-glow-color),
      inset 0 0 0 var(--island-glow-width) var(--island-glow-color);
    --island-hover-border:
      inset 0 0 0 var(--island-border-width) var(--island-border-color),
      inset 0 0 0 var(--island-border-width) var(--island-border-color);
    --island-hover-glow: 0 0 8px var(--island-border-color);

    margin-top: 5px;
    position: relative;
    background: var(--c-bg);

    background-clip: padding-box;
    -webkit-background-clip: padding-box;

    border: none;
    box-shadow: var(--island-box-shadow);

    transition:
      filter 0.4s ease,
      box-shadow 0.4s ease,
      --island-border-color 0.3s ease,
      width var(--island-anim),
      height var(--island-anim),
      border-radius var(--island-anim);

    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--c-text-primary);
    pointer-events: auto;

    transform: translateZ(0);
    -webkit-transform: translateZ(0);
  }

  .island {
    will-change: width, height, border-radius;

    contain: layout style;

    font-variant-numeric: tabular-nums;
  }

  .timeout-bar {
    position: absolute;
    bottom: 2px;
    left: 50%;
    transform: translateX(-50%);
    height: 3px;
    border-radius: 1.5px;
    background-color: var(--island-hidden-glow);
    animation: shrinkBar 10s linear forwards;
    opacity: 0.7;
    z-index: 20;
    pointer-events: none;
  }

  @keyframes shrinkBar {
    0% {
      width: 40%;
    }
    100% {
      width: 0%;
    }
  }

  .island.is-hidden {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
    border: 1px solid var(--c-border);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    background-color: var(--c-bg);
  }

  .island-drag-handle {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 12px;
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 120;
    cursor: grab;
    pointer-events: auto;
  }

  .island-drag-handle:active {
    cursor: grabbing;
  }

  .drag-pill-bar {
    width: 28px;
    height: 3px;
    border-radius: 9999px;
    background: var(--c-text-primary, #ffffff);
    opacity: 0.18;
    transition:
      opacity 0.2s ease,
      width 0.2s ease;
  }

  .island-drag-handle:hover .drag-pill-bar {
    width: 36px;
    opacity: 0.45;
    background: var(--c-text-primary, #ffffff);
  }

  .island-drag-handle:active .drag-pill-bar {
    opacity: 0.6;
  }

  .hidden-widget-pill {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    padding: 0 4px;
    box-sizing: border-box;
    z-index: 10;
    color: var(--c-text-primary);
    overflow: hidden;
  }

  .hidden-widget-pill :global(.clock-container) {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .hidden-widget-pill :global(.clock-container.minimal .time) {
    font-size: 0.72rem;
    font-weight: 700;
    line-height: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    margin: 0;
  }

  .hidden-widget-pill :global(.ampm-minimal) {
    font-size: 0.58rem;
    font-weight: 700;
    opacity: 0.85;
    margin-left: 2px;
    line-height: 1;
  }

  .hidden-widget-pill :global(.weather-minimal) {
    font-size: 0.72rem;
    gap: 0.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
  }

  .hidden-dash-mini {
    position: absolute;
    bottom: 1px;
    left: 50%;
    transform: translateX(-50%);
    width: 14px;
    height: 1.5px;
    border-radius: 1px;
    background-color: var(--island-hidden-glow);
    opacity: 0.7;
  }

  .hidden-dash {
    position: absolute;
    bottom: 2px;
    left: 50%;
    transform: translateX(-50%);
    width: 20px;
    height: 2px;
    border-radius: 2px;
    background-color: var(--island-hidden-glow);
    box-shadow: 0 0 4px var(--island-hidden-glow);
    z-index: 10;
  }

  @property --angle {
    syntax: '<angle>';
    initial-value: 0deg;
    inherits: false;
  }

  @keyframes orbit {
    from {
      --angle: 0deg;
    }
    to {
      --angle: 360deg;
    }
  }

  @keyframes fade-glow {
    0% {
      opacity: 0;
    }
    10% {
      opacity: var(--island-glow-opacity, 1);
    }
    75% {
      opacity: var(--island-glow-opacity, 1);
    }
    100% {
      opacity: 0;
    }
  }

  .island::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    box-shadow: var(--ambient-glow-ring);
    -webkit-mask: conic-gradient(
      from var(--angle),
      rgba(0, 0, 0, 0) 0%,
      rgba(0, 0, 0, 1) 50%,
      rgba(0, 0, 0, 0) 100%
    );
    mask: conic-gradient(
      from var(--angle),
      rgba(0, 0, 0, 0) 0%,
      rgba(0, 0, 0, 1) 50%,
      rgba(0, 0, 0, 0) 100%
    );
    pointer-events: none;
    opacity: 0;
    filter: drop-shadow(0 0 5px var(--island-glow-color));
    z-index: 10;
  }

  .island.is-transition-glow::before {
    animation:
      orbit 3s linear forwards,
      fade-glow 3s ease-in-out forwards;
  }

  .island:not(.expanded) {
    cursor: pointer;
  }

  .island:not(.expanded):not(.is-hidden):hover::before {
    opacity: 0 !important;
    animation: none !important;
  }

  .island:not(.expanded):not(.is-hidden):hover {
    box-shadow:
      var(--island-hover-border), var(--island-hover-glow),
      var(--island-hover-drop-shadow);
  }

  .island.is-hidden:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
  }

  .content {
    display: grid;
    place-items: center;
    width: 100%;
    height: 100%;
    border-radius: inherit;
    overflow: hidden;
    position: relative;
    transition: opacity 0.22s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  .view {
    grid-area: 1 / 1;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    transition:
      width var(--island-anim),
      height var(--island-anim);
  }

  .island.is-quitting {
    transform: scale(0.6) translateY(-20px) !important;
    opacity: 0 !important;
    filter: blur(12px) !important;
    transition:
      transform 0.35s cubic-bezier(0.16, 1, 0.3, 1),
      opacity 0.35s ease,
      filter 0.35s ease !important;
  }
</style>
