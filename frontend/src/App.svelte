<script lang="ts">
  import { onMount, setContext, type Component } from 'svelte'
  import { get } from 'svelte/store'
  import { fade } from 'svelte/transition'
  import {
    SetWidgetBounds,
    IsTestMode,
    FocusWindow,
    Quit,
    GetBatteryStatus,
    GetWirelessStatus,
    CheckPostUpdateInstalled,
    TrimMemory,
    SetIsHidden,
    GetDisplayLayout,
    RequestMediaUpdate
  } from '../bindings/Vinta/backend/app.js'
  import { Events } from '@wailsio/runtime'

  const appEventUnsubs: (() => void)[] = []

  function EventsOn<T = unknown>(
    eventName: string,
    callback: (data: T) => void
  ): () => void {
    const unsub = Events.On(eventName, (ev: { data: T }) => {
      callback(ev?.data)
    })
    appEventUnsubs.push(unsub)
    return unsub
  }
  import NotificationPage, {
    config as notifConfig
  } from './pages/NotificationPage.svelte'
  import HomePage, { config as homeConfig } from './pages/HomePage.svelte'
  import MediaPage, { config as mediaConfig } from './pages/MediaPage.svelte'
  import NotesPage from './pages/NotesPage.svelte'
  import AppsPage from './pages/AppsPage.svelte'
  import SettingsPage, {
    config as settingsConfig
  } from './pages/SettingsPage.svelte'
  import CalculatorPage from './pages/CalculatorPage.svelte'
  import Dock from './components/Dock.svelte'
  import FloatingControls from './components/FloatingControls.svelte'
  import SatelliteHost from './components/SatelliteHost.svelte'
  import Clock from './components/Clock.svelte'
  import DevPill from './devtest/DevPill.svelte'
  import appIcon from '@appicon'
  import {
    satellitesStore,
    getActiveSatelliteFootprint
  } from './stores/satellites'
  import { AppControls } from './utils/navigation'
  import { playNotifChime } from './services/audioService'
  import {
    themeStore,
    autoHideStore,
    homeAutoHideStore,
    mediaAutoHideStore,
    autoHideDelayStore,
    autoHideDisplayStore,
    notifAutoWakeStore,
    globalMediaStore,
    audioPeakStore,
    type MediaData,
    type MediaSessionItem,
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
    activeSettingsTabStore,
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
  function handleDebugBoundsOverlayEffect() {
    syncDebugBoundsOverlay($debugBoundsOverlay)
  }
  $effect(handleDebugBoundsOverlayEffect)

  // Dynamic Routing: The active app that the island is displaying
  let activePage = $state<Component>(HomePage)
  let islandConfig = $state<IslandConfig>(homeConfig)

  let expanded = $state(false)
  let isIslandTransitioning = $state(false)
  let isDockVisible = $state(false)
  let islandTransitionTimeout: number | undefined

  /**
   * Tracks island expand and collapse animations to hide navigation and control docks
   * until geometry changes are fully completed.
   */
  function handleExpandedTransitionEffect() {
    const isExp = expanded
    clearTimeout(islandTransitionTimeout)
    isIslandTransitioning = true
    isDockVisible = false

    islandTransitionTimeout = window.setTimeout(() => {
      isIslandTransitioning = false
      if (isExp) {
        isDockVisible = true
      }
    }, 340)
  }

  $effect(handleExpandedTransitionEffect)
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
  let hiddenDragStartX = 0
  let hiddenDragStartOffset = 0
  let dragMovedDistance = 0
  let dragResetTimeout: number | undefined

  let islandNode = $state<HTMLElement>()
  let islandContainerNode = $state<HTMLElement>()
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

  interface MonitorInfo {
    id: number
    name: string
    left: number
    top: number
    width: number
    height: number
    isPrimary: boolean
    isPortrait?: boolean
  }

  interface DisplayLayoutInfo {
    virtualLeft: number
    virtualTop: number
    virtualWidth: number
    virtualHeight: number
    primaryCenter: number
    currentMonitor?: MonitorInfo
    monitors: MonitorInfo[]
  }

  let displayLayout = $state<DisplayLayoutInfo>({
    virtualLeft: 0,
    virtualTop: 0,
    virtualWidth: 1920,
    virtualHeight: 1080,
    primaryCenter: 960,
    currentMonitor: {
      id: 1,
      name: 'Primary Display',
      left: 0,
      top: 0,
      width: 1920,
      height: 1080,
      isPrimary: true,
      isPortrait: false
    },
    monitors: [
      {
        id: 1,
        name: 'Primary Display',
        left: 0,
        top: 0,
        width: 1920,
        height: 1080,
        isPrimary: true,
        isPortrait: false
      }
    ]
  })

  // Derive the active monitor based on island composite desktop coordinates
  function computeActiveMonitor() {
    const screenCenterX =
      displayLayout.virtualLeft +
      displayLayout.primaryCenter +
      currentLiveOffsetX
    if (displayLayout.monitors && displayLayout.monitors.length > 0) {
      for (const mon of displayLayout.monitors) {
        if (screenCenterX >= mon.left && screenCenterX < mon.left + mon.width) {
          return mon
        }
      }
    }
    return (
      displayLayout.currentMonitor ||
      displayLayout.monitors?.[0] || {
        id: 1,
        name: 'Primary Display',
        left: 0,
        top: 0,
        width: 1920,
        height: 1080,
        isPrimary: true,
        isPortrait: false
      }
    )
  }
  let activeMonitor = $derived.by(computeActiveMonitor)

  setContext('activeMonitor', () => activeMonitor)

  let hideTrigger = $state(0)
  let isSatelliteHovering = $state(false)
  let hoverGraceTimeout: number | undefined

  function clearHoverGrace() {
    if (hoverGraceTimeout) {
      clearTimeout(hoverGraceTimeout)
      hoverGraceTimeout = undefined
    }
  }

  function computeIsAnyHovered() {
    return isHovering || isSatelliteHovering
  }
  let isAnyHovered = $derived.by(computeIsAnyHovered)

  function handleSatelliteMouseEnter() {
    clearHoverGrace()
    isSatelliteHovering = true
    isHovering = false
    clearTimeout(hideTimeout)
    if (isHidden && $wakeMethodStore === 'hover') {
      hoverShowTimeout = window.setTimeout(showWidget, 150)
    } else if (!isHidden) {
      showWidget()
    }
  }

  const HOVER_GRACE_MS = 800

  function handleSatelliteMouseLeave() {
    clearHoverGrace()
    hoverGraceTimeout = window.setTimeout(() => {
      isHovering = false
      isSatelliteHovering = false
      clearTimeout(hoverShowTimeout)
      hideTrigger++
    }, HOVER_GRACE_MS)
  }

  function handleAutoHideAndRevert() {
    void hideTrigger
    clearTimeout(hideTimeout)
    isTimeoutActive = false

    if ($devPinnedStore) return

    if (!expanded && !isHidden && !isAnyHovered && !isControlsDrawerOpen) {
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
          } else if (activePage !== HomePage && activePage !== MediaPage) {
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

    const target = getBoundsForState(islandConfig, false, true)
    targetPendingBounds = target
    clearTimeout(boundsTransitionTimeout)
    boundsTransitionTimeout = window.setTimeout(tryCommitShrinkingBounds, 600)
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

  async function showWidget() {
    if (!isHidden) return
    clearTimeout(dormantUnmountTimeout)
    dormantUnmountTimeout = undefined
    clearTimeout(memoryTrimTimeout)
    memoryTrimTimeout = undefined
    isPageDormant = false

    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'
    isDraggingHidden = false

    // Pre-expand the OS bounds before toggling isHidden so the expanding CSS transition is never clipped
    const { w, h, r } = getBoundsForState(islandConfig, expanded, false)
    await applyBounds(
      Math.max(currentOSBounds.w, w),
      Math.max(currentOSBounds.h, h),
      r
    )

    isHidden = false
    notifyOSHidden(false)
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
  let lastDispatchedBounds: {
    w: number
    h: number
    r: number
    offset: number
    scale: number
  } | null = null

  let pendingBoundsResolver: (() => void) | null = null

  function updateBounds(
    w: number,
    h: number,
    r: number,
    explicitOffset?: number
  ): Promise<void> {
    const computedW = Math.round(w)
    const computedH = Math.round(h)
    const computedR = Math.round(r)
    const scale = window.devicePixelRatio || 1
    const rawOffset =
      explicitOffset !== undefined
        ? explicitOffset
        : isNaN(clampedOffsetX)
          ? 0
          : clampedOffsetX
    const offset = Math.round(rawOffset)

    const isCorridorUnchanged =
      computedW === -1 &&
      lastDispatchedBounds?.w === -1 &&
      lastDispatchedBounds.h === computedH &&
      lastDispatchedBounds.scale === scale

    const isTightUnchanged =
      lastDispatchedBounds !== null &&
      lastDispatchedBounds.w === computedW &&
      lastDispatchedBounds.h === computedH &&
      lastDispatchedBounds.r === computedR &&
      lastDispatchedBounds.offset === offset &&
      lastDispatchedBounds.scale === scale

    if (isCorridorUnchanged || isTightUnchanged) {
      return Promise.resolve()
    }

    if (isBoundsIpcPending) {
      pendingBoundsArgs = { w: computedW, h: computedH, r: computedR, offset }
      return new Promise<void>((resolve) => {
        const prevResolver = pendingBoundsResolver
        pendingBoundsResolver = () => {
          if (prevResolver) prevResolver()
          resolve()
        }
      })
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
    lastDispatchedBounds = {
      w: computedW,
      h: computedH,
      r: computedR,
      offset,
      scale
    }

    return Promise.resolve(
      SetWidgetBounds(computedW, computedH, computedR, scale, offset)
    )
      .catch((err) => {
        if (get(debugLogs)) {
          console.error('[Vinta:BoundsError]', err)
        }
      })
      .finally(() => {
        isBoundsIpcPending = false
        const resolver = pendingBoundsResolver
        pendingBoundsResolver = null
        if (pendingBoundsArgs) {
          const next = pendingBoundsArgs
          pendingBoundsArgs = null
          updateBounds(next.w, next.h, next.r, next.offset).then(() => {
            if (resolver) resolver()
          })
        } else if (resolver) {
          resolver()
        }
      })
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

  function applyBounds(w: number, h: number, r: number): Promise<void> {
    currentOSBounds = { w, h, r }
    return updateBounds(w, h, r, clampedOffsetX)
  }

  let shrinkingRetryCount = 0

  function tryCommitShrinkingBounds() {
    if (isDraggingHidden || isTrackRegionActive) return
    if (targetPendingBounds.w <= 0 || targetPendingBounds.h <= 0) {
      shrinkingRetryCount = 0
      return
    }

    // If islandNode is still visually animating down, do not shrink the OS region yet
    if (islandNode && shrinkingRetryCount < 4) {
      const currentVisualW = islandNode.offsetWidth
      const currentVisualH = islandNode.offsetHeight
      if (
        currentVisualW > targetPendingBounds.w + 4 ||
        currentVisualH > targetPendingBounds.h + 4
      ) {
        // Postpone check to let laggy or heavy-load animation finish smoothly
        shrinkingRetryCount++
        clearTimeout(boundsTransitionTimeout)
        boundsTransitionTimeout = window.setTimeout(
          tryCommitShrinkingBounds,
          100
        )
        return
      }
    }

    shrinkingRetryCount = 0
    clearTimeout(boundsTransitionTimeout)
    applyBounds(
      targetPendingBounds.w,
      targetPendingBounds.h,
      targetPendingBounds.r
    )
    targetPendingBounds = { w: 0, h: 0, r: 0 }
  }

  function handleIslandTransitionEnd(e: TransitionEvent) {
    if (e.target !== e.currentTarget) return
    if (e.propertyName !== 'width' && e.propertyName !== 'height') return

    // If we are shrinking, ensure the visual island has actually finished transitioning
    // to target dimensions before shrinking the OS region
    if (targetPendingBounds.w > 0 && targetPendingBounds.h > 0) {
      tryCommitShrinkingBounds()
    }
  }

  let hiddenBounds = $derived(
    $autoHideDisplayStore === 'clock'
      ? { w: 116, h: 28, r: 14 }
      : { w: 72, h: 12, r: 6 }
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
      : Math.min(
          expanded ? islandConfig.expanded.w : currentMinimalW,
          activeMonitor.width ? activeMonitor.width - 24 : 1920
        )
  )

  function computeVirtualMinOffset() {
    const halfW = currentIslandW / 2
    return Math.floor(12 + halfW - displayLayout.primaryCenter)
  }
  let virtualMinOffset = $derived.by(computeVirtualMinOffset)

  function computeVirtualMaxOffset() {
    const halfW = currentIslandW / 2
    return Math.floor(
      displayLayout.virtualWidth - 12 - halfW - displayLayout.primaryCenter
    )
  }
  let virtualMaxOffset = $derived.by(computeVirtualMaxOffset)

  function computeMonitorSnapOffsets() {
    if (!displayLayout.monitors || displayLayout.monitors.length === 0) {
      return [0]
    }
    return displayLayout.monitors.map((mon) => {
      const monClientLeft = mon.left - displayLayout.virtualLeft
      const monCenter = monClientLeft + mon.width / 2
      return Math.round(monCenter - displayLayout.primaryCenter)
    })
  }
  let monitorSnapOffsets = $derived.by(computeMonitorSnapOffsets)

  function computeActiveLeftSatelliteFootprint() {
    return getActiveSatelliteFootprint('left', $satellitesStore)
  }
  let activeLeftSatelliteFootprint = $derived.by(
    computeActiveLeftSatelliteFootprint
  )

  function computeActiveRightSatelliteFootprint() {
    return getActiveSatelliteFootprint('right', $satellitesStore)
  }
  let activeRightSatelliteFootprint = $derived.by(
    computeActiveRightSatelliteFootprint
  )

  let isControlsDrawerOpen = $state(false)

  /**
   * Tracks open/close state of the minimized horizontal control drawer.
   */
  function handleControlsDrawerChange(isOpen: boolean) {
    if (isControlsDrawerOpen === isOpen) return
    isControlsDrawerOpen = isOpen
    if (isOpen) {
      clearTimeout(hideTimeout)
      isTimeoutActive = false
    }
  }

  /**
   * Picks the tooltip side for the expanded vertical dock by comparing
   * available clearance on both flanks. Whichever side of the dock has
   * more room to render the tooltip label wins.
   */
  function computeExpandedTooltipSide(): 'right' | 'left' {
    const mon = activeMonitor
    const islandCenterOnScreen =
      displayLayout.primaryCenter + clampedOffsetX - mon.left
    const halfExpandedW = (islandConfig.expanded.w || 380) / 2
    const dockGap = 10
    const dockWidth = 32
    const dockRightEdge =
      islandCenterOnScreen + halfExpandedW + dockGap + dockWidth
    const dockLeftEdge = dockRightEdge - dockWidth

    const rightClearance = mon.width - dockRightEdge
    const leftClearance = dockLeftEdge
    return rightClearance >= leftClearance ? 'right' : 'left'
  }
  let expandedTooltipSide = $derived.by(computeExpandedTooltipSide)

  /**
   * Computes dynamic geometric bounds for the window controls satellite dock
   * across expanded and minimized states, expanding tooltip padding when hovered.
   */
  function computeControlsSatelliteBounds() {
    const isHovered = isSatelliteHovering
    if (expanded) {
      const expH = islandConfig.expanded.h || 400
      const centerOffsetY = Math.max(0, Math.round((expH - 88) / 2))
      return {
        width: 32,
        height: 88,
        radius: 16,
        gap: 10,
        offsetY: centerOffsetY,
        tooltipPad: isHovered ? 75 : 0,
        tooltipSide: expandedTooltipSide
      }
    }
    const minH = islandConfig.minimal.h || 36
    const centerOffsetY = Math.max(0, Math.round((minH - 32) / 2))
    if (isControlsDrawerOpen) {
      return {
        width: 32,
        height: 116,
        radius: 0,
        gap: 10,
        offsetY: centerOffsetY,
        tooltipPad: isHovered ? 45 : 0,
        tooltipSide: 'right' as const
      }
    }
    return {
      width: 32,
      height: 32,
      radius: 0,
      gap: 10,
      offsetY: centerOffsetY,
      tooltipPad: isHovered ? 45 : 0,
      tooltipSide: 'right' as const
    }
  }
  let controlsSatelliteBounds = $derived.by(computeControlsSatelliteBounds)

  function computeMinAllowedOffsetForCurrentW() {
    const halfW = currentIslandW / 2
    const mon = activeMonitor
    const baseCenter = displayLayout.virtualLeft + displayLayout.primaryCenter
    const leftMargin = 12 + activeLeftSatelliteFootprint
    const rightMargin = 12 + activeRightSatelliteFootprint
    const targetLeft = mon.left + leftMargin + halfW
    const targetRight = mon.left + mon.width - rightMargin - halfW
    if (targetLeft > targetRight) {
      return Math.floor(mon.left + mon.width / 2 - baseCenter)
    }
    return Math.floor(targetLeft - baseCenter)
  }
  let minAllowedOffsetForCurrentW = $derived.by(
    computeMinAllowedOffsetForCurrentW
  )

  function computeMaxAllowedOffsetForCurrentW() {
    const halfW = currentIslandW / 2
    const mon = activeMonitor
    const baseCenter = displayLayout.virtualLeft + displayLayout.primaryCenter
    const leftMargin = 12 + activeLeftSatelliteFootprint
    const rightMargin = 12 + activeRightSatelliteFootprint
    const targetLeft = mon.left + leftMargin + halfW
    const targetRight = mon.left + mon.width - rightMargin - halfW
    if (targetLeft > targetRight) {
      return Math.floor(mon.left + mon.width / 2 - baseCenter)
    }
    return Math.floor(targetRight - baseCenter)
  }
  let maxAllowedOffsetForCurrentW = $derived.by(
    computeMaxAllowedOffsetForCurrentW
  )

  let currentLiveOffsetX = $state(0)

  // Smart offset safety: smoothly clamp main island back into bounds if content expands near screen edge
  function handleSmartEdgeOffsetAdjustment() {
    if (!isDraggingHidden && !isControlsDrawerOpen) {
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

  // Sync store updates into live offset when not dragging
  function syncStoreOffsetToLive(val: number) {
    if (!isDraggingHidden) {
      currentLiveOffsetX = val
    }
  }
  function handleHiddenPillOffsetSync() {
    syncStoreOffsetToLive($hiddenPillOffsetXStore)
  }
  $effect(handleHiddenPillOffsetSync)

  function computeClampedOffsetX() {
    if (Math.abs(currentLiveOffsetX) < 3) return 0
    if (isDraggingHidden) {
      return currentLiveOffsetX
    }
    return Math.max(
      minAllowedOffsetForCurrentW,
      Math.min(maxAllowedOffsetForCurrentW, currentLiveOffsetX)
    )
  }
  let clampedOffsetX = $derived.by(computeClampedOffsetX)

  function computeMinDevPillOffset() {
    const halfW = 430 / 2
    const mon = activeMonitor
    const baseCenter = displayLayout.virtualLeft + displayLayout.primaryCenter
    const targetLeft = mon.left + 12 + halfW
    const targetRight = mon.left + mon.width - 12 - halfW
    if (targetLeft > targetRight) {
      return Math.floor(mon.left + mon.width / 2 - baseCenter)
    }
    return Math.floor(targetLeft - baseCenter)
  }
  let minDevPillOffset = $derived.by(computeMinDevPillOffset)

  function computeMaxDevPillOffset() {
    const halfW = 430 / 2
    const mon = activeMonitor
    const baseCenter = displayLayout.virtualLeft + displayLayout.primaryCenter
    const targetLeft = mon.left + 12 + halfW
    const targetRight = mon.left + mon.width - 12 - halfW
    if (targetLeft > targetRight) {
      return Math.floor(mon.left + mon.width / 2 - baseCenter)
    }
    return Math.floor(targetRight - baseCenter)
  }
  let maxDevPillOffset = $derived.by(computeMaxDevPillOffset)

  function computeDevClampedOffsetX() {
    if (Math.abs(clampedOffsetX) < 3) return 0
    return Math.max(
      minDevPillOffset,
      Math.min(maxDevPillOffset, clampedOffsetX)
    )
  }
  let devClampedOffsetX = $derived.by(computeDevClampedOffsetX)

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

    const maxMonW = activeMonitor?.width ? activeMonitor.width - 24 : 1920
    const boundedBaseW = isHid ? baseW : Math.min(baseW, maxMonW)
    const w = boundedBaseW + extraDropdownW
    const h = baseH + extraDropdownH
    const r = isHid
      ? hiddenBounds.r
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
      finishDrag()
      const bounds = getBoundsForState(pageConfig, expand, false)
      applyBounds(
        Math.max(bounds.w, currentOSBounds.w),
        Math.max(bounds.h, currentOSBounds.h),
        bounds.r
      ).then(() => {
        isHidden = false
        notifyOSHidden(false)
      })
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
    else if (page === CalculatorPage) $activePageIdStore = 'calculator'

    if (expand) {
      const { w, h, r } = getBoundsForState(pageConfig, true, false)
      if (w > currentOSBounds.w || h > currentOSBounds.h) {
        applyBounds(
          Math.max(w, currentOSBounds.w),
          Math.max(h, currentOSBounds.h),
          r
        ).then(() => {
          expanded = true
        })
      } else {
        expanded = true
      }
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
        Quit().catch(() => {})
      }, 200)
    }

    window.addEventListener('vinta-overlay', handleOverlayChange)
    window.addEventListener('vinta-quit', handleQuit)

    // Listen to systray events
    EventsOn('toggle-theme', () => {
      themeStore.update((t) => (t === 'dark' ? 'light' : 'dark'))
    })
    EventsOn('show-pill', showWidget)

    let wasHiddenBeforeGameMode = false
    EventsOn('game-mode-enter', (reason: unknown) => {
      if (isHidden) {
        wasHiddenBeforeGameMode = true
      } else {
        wasHiddenBeforeGameMode = false
        hideWidget()
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

    EventsOn('game-mode-exit', () => {
      if (!wasHiddenBeforeGameMode && isHidden) {
        showWidget()
      }
    })

    let prevIsPlaying = false
    EventsOn('media-update', (data: Record<string, unknown>) => {
      if (data) {
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
        if (isPageDormant && !isIncomingPlaying) return

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
              : isIncomingAvailable &&
                  incomingTitle === currentTitle &&
                  m.albumArt
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

    EventsOn(
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

    EventsOn('notification_permission_denied', () => {
      notificationPermissionDenied.set(true)
      if (get(pageVisibilityStore).notifications) {
        showWidget()
        navigateTo(NotificationPage, notifConfig, true)
      }
    })

    const handleNavigateSettings = (e: Event) => {
      const customEv = e as CustomEvent<{ tab?: string }>
      if (customEv.detail?.tab) {
        activeSettingsTabStore.set(customEv.detail.tab)
      }
      showWidget()
      navigateTo(SettingsPage, settingsConfig, true)
    }
    window.addEventListener('vinta-navigate-settings', handleNavigateSettings)

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
              canOpen: true,
              isUpdate: true,
              icon: appIcon,
              appLogo: appIcon
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
                canOpen: true,
                isUpdate: true,
                icon: appIcon,
                appLogo: appIcon
              },
              ...notifs
            ])
            showWidget()
          }
        })
        .catch(() => {})
    }

    EventsOn('monitor-changed', (mon: MonitorInfo) => {
      if (mon && mon.id) {
        displayLayout = {
          ...displayLayout,
          currentMonitor: mon
        }
      }
    })

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

    // Query active media state on startup so album art and playback state
    // are rendered immediately if media was playing before Vinta launched.
    if (typeof RequestMediaUpdate === 'function') {
      RequestMediaUpdate()
    }

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
      if (e.key === 'Escape' && isDraggingHidden) {
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
      finishDrag()
      window.clearInterval(batteryInterval)
      window.clearInterval(wirelessInterval)
      window.removeEventListener('contextmenu', disableGlobalContextMenu)
      window.removeEventListener('vinta-overlay', handleOverlayChange)
      window.removeEventListener('vinta-quit', handleQuit)
      window.removeEventListener(
        'vinta-navigate-settings',
        handleNavigateSettings
      )
      window.removeEventListener('resize', handleResize)
      window.removeEventListener('blur', handleGlobalDragFailsafe)
      window.removeEventListener('pointercancel', handleGlobalDragFailsafe)
      window.removeEventListener('keydown', handleWindowKeydownFailsafe)
      appEventUnsubs.forEach((unsub) => {
        try {
          unsub()
        } catch {
          // Ignore unsubscription error
        }
      })
      appEventUnsubs.length = 0
    }
  })

  let prevExpandedState = false
  let prevHiddenState = false

  /**
   * Synchronizes OS window bounds when island state changes.
   * In-flight drags skip this to prevent fighting the full-width track region,
   * but the drag handlers call finishDrag() before any state changes.
   */
  let hasInitializedBounds = false

  function syncOSBoundsOnStateChange() {
    if (isDraggingHidden || isTrackRegionActive) return
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
      boundsTransitionTimeout = window.setTimeout(tryCommitShrinkingBounds, 600)
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
      if (isDraggingHidden) {
        handleDragFailsafe()
      }
    }, 6000)
  }

  // Failsafe handler to cleanly cancel any drag operation.
  function handleDragFailsafe() {
    finishDrag()
  }

  function finishDrag() {
    clearTimeout(dragWatchdogTimeout)
    dragWatchdogTimeout = undefined
    isDraggingHidden = false
    isTrackRegionActive = false

    // Clamp current offset to active monitor once resting
    const restingOffset = Math.max(
      minAllowedOffsetForCurrentW,
      Math.min(maxAllowedOffsetForCurrentW, currentLiveOffsetX)
    )
    currentLiveOffsetX = restingOffset
    $hiddenPillOffsetXStore = restingOffset

    const { w, h, r } = getBoundsForState()
    updateBounds(w, h, r, restingOffset)
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

    isDraggingHidden = true
    hasPointerCaptured = false
    dragMovedDistance = 0
    clearTimeout(dragResetTimeout)
    resetDragWatchdog()
    hiddenDragStartX = e.screenX
    hiddenDragStartOffset = currentLiveOffsetX

    // While mouse is actively held down, expand Win32 clipping region to full width runway
    isTrackRegionActive = true
    const { h, r } = getBoundsForState()
    updateBounds(-1, Math.max(h, 200), r, 0)

    try {
      ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
      hasPointerCaptured = true
    } catch {
      // Ignore
    }
  }

  function handleHiddenPointerMove(e: PointerEvent) {
    if (!isDraggingHidden) return
    // If no mouse buttons are pressed (e.g. released outside/interrupted), cancel drag immediately
    if (e.buttons === 0) {
      handleHiddenPointerUp(e)
      return
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
      for (const snap of monitorSnapOffsets) {
        if (Math.abs(rawOffset - snap) <= 28) {
          targetOffset = snap
          break
        }
      }

      currentLiveOffsetX = Math.max(
        virtualMinOffset,
        Math.min(virtualMaxOffset, targetOffset)
      )
    }
  }

  function handleHiddenPointerUp(e: PointerEvent) {
    if (!isDraggingHidden) return
    finishDrag()
    if (hasPointerCaptured) {
      hasPointerCaptured = false
      try {
        ;(e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId)
      } catch {
        // Ignore
      }
    }

    clearTimeout(dragResetTimeout)
    dragResetTimeout = window.setTimeout(() => {
      dragMovedDistance = 0
    }, 150)
  }

  function handleContextMenu(e: MouseEvent) {
    if (isDraggingHidden || dragMovedDistance > 6) {
      dragMovedDistance = 0
      finishDrag()
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

  async function handleExpand() {
    if (expanded) return
    if (isDraggingHidden) finishDrag()
    clearTimeout(boundsTransitionTimeout)
    targetPendingBounds = { w: 0, h: 0, r: 0 }
    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'

    // Open expanded OS bounds before starting CSS expansion to eliminate clipping
    const { w, h, r } = getBoundsForState(islandConfig, true, isHidden)
    await applyBounds(
      Math.max(currentOSBounds.w, w),
      Math.max(currentOSBounds.h, h),
      r
    )

    expanded = true
    isControlsDrawerOpen = false
  }

  function handleIslandClick(e: MouseEvent) {
    if (dragMovedDistance > 6) {
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
    clearHoverGrace()
    isHovering = true
    isSatelliteHovering = false
    clearTimeout(hideTimeout)
    if (isHidden && $wakeMethodStore === 'hover') {
      hoverShowTimeout = window.setTimeout(showWidget, 150)
    } else if (!isHidden) {
      showWidget()
    }
  }

  function handleMouseLeave() {
    clearHoverGrace()
    hoverGraceTimeout = window.setTimeout(() => {
      isHovering = false
      isSatelliteHovering = false
      clearTimeout(hoverShowTimeout)
      hideTrigger++
    }, HOVER_GRACE_MS)
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
  bind:this={islandContainerNode}
  class="island-container"
  role="application"
  onfocusin={handleFocusIn}
  onpointerdown={handleGlobalPointerDown}
  style="position: absolute; left: {displayLayout.primaryCenter}px; top: {Math.max(
    0,
    activeMonitor.top - displayLayout.virtualTop
  )}px; transform: translateX(-50%);"
>
  <div
    class="island-wrapper {isQuitting ? 'is-quitting' : ''}"
    style="
      opacity: {isQuitting ? 0 : 1};
      pointer-events: {isQuitting ? 'none' : 'auto'};
      transform: translateX({clampedOffsetX}px) {isQuitting
      ? 'scale(0.6) translateY(-20px)'
      : ''};
      transition: {isDraggingHidden ? 'none' : islandAnim};
    "
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
      ontransitionend={handleIslandTransitionEnd}
      style="
        opacity: {isQuitting ? 0 : activeOpacity};
        filter: {isQuitting ? 'blur(12px)' : 'none'};
        width: {isHidden
        ? hiddenBounds.w
        : Math.min(
            expanded ? islandConfig.expanded.w : currentMinimalW,
            activeMonitor.width ? activeMonitor.width - 24 : 1920
          )}px;
        height: {isHidden
        ? hiddenBounds.h
        : expanded
          ? islandConfig.expanded.h
          : islandConfig.minimal.h}px;
        --island-r: {isHidden
        ? hiddenBounds.r
        : expanded
          ? islandConfig.expanded.r
          : islandConfig.minimal.r}px;
        border-radius: 0 0 var(--island-r) var(--island-r);
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
              style="width: {Math.min(
                expanded ? islandConfig.expanded.w : currentMinimalW,
                activeMonitor.width ? activeMonitor.width - 24 : 1920
              )}px; height: {expanded
                ? islandConfig.expanded.h
                : islandConfig.minimal.h}px;"
            >
              {#if ActivePage}
                <ActivePage {expanded} />
              {/if}
            </div>
          {/key}
        {/if}

        {#if expanded && isDockVisible}
          <div
            class="dock-overlay-wrapper"
            in:fade={{ duration: 160 }}
            out:fade={{ duration: 80 }}
          >
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

    {#if !isQuitting && !isIslandTransitioning && !isHidden}
      <SatelliteHost
        id="controls"
        anchor="right"
        width={controlsSatelliteBounds.width}
        height={controlsSatelliteBounds.height}
        radius={controlsSatelliteBounds.radius}
        gap={controlsSatelliteBounds.gap}
        offsetY={controlsSatelliteBounds.offsetY}
        tooltipPad={controlsSatelliteBounds.tooltipPad}
        tooltipSide={controlsSatelliteBounds.tooltipSide}
        active={true}
        onmouseenter={handleSatelliteMouseEnter}
        onmouseleave={handleSatelliteMouseLeave}
      >
        <FloatingControls
          {activePage}
          {expanded}
          {expandedTooltipSide}
          onDrawerChange={handleControlsDrawerChange}
        />
      </SatelliteHost>
    {/if}
  </div>

  {#if isTestMode}
    <div
      bind:this={devPillContainerNode}
      role="none"
      onmouseenter={handleSatelliteMouseEnter}
      onmouseleave={handleSatelliteMouseLeave}
      style="
        transform: translateX({devClampedOffsetX}px);
        transition: {isDraggingHidden ? 'none' : islandAnim};
      "
    >
      <DevPill />
    </div>
  {/if}
</div>

<style src="./App.css"></style>
