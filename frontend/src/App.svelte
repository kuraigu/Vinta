<script lang="ts">
  import { onMount, setContext, type Component } from 'svelte'
  import { FocusWindow, Quit, GetDisplayLayout } from '../bindings/Vinta/backend/app.js'
  import { Events } from '@wailsio/runtime'

  import NotificationPage, {
    config as notifConfig
  } from './pages/NotificationPage.svelte'
  import HomePage, { config as homeConfig } from './pages/HomePage.svelte'
  import MediaPage, { config as mediaConfig } from './pages/MediaPage.svelte'
  import NotesPage, { config as notesConfig } from './pages/NotesPage.svelte'
  import AppsPage, { config as appsConfig } from './pages/AppsPage.svelte'
  import SettingsPage, {
    config as settingsConfig
  } from './pages/SettingsPage.svelte'
  import CalculatorPage, {
    config as calcConfig
  } from './pages/CalculatorPage.svelte'

  import MainIsland from './components/MainIsland.svelte'
  import ControlsSatellite from './components/ControlsSatellite.svelte'
  import CustomTooltip from './components/CustomTooltip.svelte'
  import { getFallbackPage } from './utils/navigation'
  import { initIslandEvents } from './services/islandEvents'
  import { WindowBoundsManager } from './services/windowBounds'
  import {
    IslandGestureController,
    computeAllowedOffsets,
    computeMonitorSnapOffsets
  } from './services/islandGestures'
  import { playNotifChime } from './services/audioService'
  import {
    satellitesStore,
    getActiveSatelliteFootprint
  } from './stores/satellites'
  import {
    themeStore,
    autoHideStore,
    homeAutoHideStore,
    mediaAutoHideStore,
    autoHideDelayStore,
    autoHideDisplayStore,
    wakeMethodStore,
    opacityMinimalStore,
    opacityExpandedStore,
    opacityHiddenStore,
    hiddenPillOffsetXStore,
    notifSoundStore,
    notifVolumeStore,
    glowIntensityStore,
    globalMediaStore,
    type IslandConfig,
    homeWidgetsVisibilityStore,
    calculateHomeMinimalWidth,
    batteryDataStore,
    wirelessDataStore,
    weatherLocationStore,
    weatherUnitStore,
    timezoneStore,
    simulateOfflineStore,
    isSimulatingBatteryStore,
    isSimulatingWirelessStore,
    devPinnedStore,
    activePageIdStore,
    type BatteryData,
    type WirelessData,
    type WeatherLocation,
    type TemperatureUnit
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

  // Subsystem Managers
  const bounds = new WindowBoundsManager()
  const gestures = new IslandGestureController()

  // Dynamic Routing
  let activePage = $state<Component>(HomePage)
  let islandConfig = $state<IslandConfig>(homeConfig)

  let expanded = $state(false)
  let isIslandTransitioning = $state(false)
  let isDockVisible = $state(false)
  let islandTransitionTimeout: number | undefined

  let previousExpanded: boolean | undefined = undefined

  function handleExpandedTransitionEffect() {
    const isExp = expanded
    if (previousExpanded === undefined) {
      previousExpanded = isExp
      return
    }
    if (previousExpanded === isExp) {
      return
    }
    previousExpanded = isExp

    clearTimeout(islandTransitionTimeout)
    isIslandTransitioning = true
    isDockVisible = false
    const transitionDuration = isExp ? 340 : 280

    islandTransitionTimeout = window.setTimeout(() => {
      isIslandTransitioning = false
      if (isExp) {
        isDockVisible = true
      }
    }, transitionDuration)
  }
  $effect(handleExpandedTransitionEffect)

  // Peripheral Hover & Auto-hide State
  let isHovering = $state(false)
  let isSatelliteHovering = $state(false)
  const HOVER_GRACE_MS = 250
  let hoverGraceTimeout: number | undefined
  let hoverShowTimeout: number | undefined

  function clearHoverGrace() {
    if (hoverGraceTimeout !== undefined) {
      clearTimeout(hoverGraceTimeout)
      hoverGraceTimeout = undefined
    }
    if (hoverShowTimeout !== undefined) {
      clearTimeout(hoverShowTimeout)
      hoverShowTimeout = undefined
    }
  }

  // Island Visibility & Timing
  let isHidden = $state(false)
  let isBlinking = $state(false)
  let isPageDormant = $state(false)
  let isQuitting = $state(false)
  let isTransitionGlow = $state(false)
  let hideTrigger = $state(0)
  let dormantUnmountTimeout: number | undefined
  let memoryTrimTimeout: number | undefined
  let glowTimeout: number | undefined

  // Multi-Monitor Geometry State
  let displayLayout = $state({
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
    monitors: []
  })

  let activeMonitor = $derived(
    displayLayout.currentMonitor || {
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

  let extraDropdownW = $state(0)
  let extraDropdownH = $state(0)
  let islandAnim = $state('0.32s cubic-bezier(0.2, 0.8, 0.2, 1)')

  let islandContainerNode = $state<HTMLDivElement | null>(null)
  let islandNode = $state<HTMLDivElement | null>(null)

  let isControlsDrawerOpen = $state(false)

  function handleControlsDrawerChange(isOpen: boolean) {
    if (isControlsDrawerOpen === isOpen) return
    isControlsDrawerOpen = isOpen
    if (isOpen) {
      clearTimeout(hideTimeout)
      isTimeoutActive = false
    } else {
      hideTrigger++
    }
  }

  let isTimeoutActive = $state(false)
  let hideTimeout: number | undefined

  let isMediaActive = $derived.by(() => {
    if (!$globalMediaStore.isAvailable) return false
    if ($globalMediaStore.isPlaying) return true
    if (Array.isArray($globalMediaStore.sessions)) {
      return $globalMediaStore.sessions.some((s) => Boolean(s.isPlaying))
    }
    return false
  })

  let shouldCurrentPageAutoHide = $derived.by(() => {
    if (activePage === HomePage) {
      return $homeAutoHideStore
    }
    if (activePage === MediaPage) {
      if (!isMediaActive) return true
      return $mediaAutoHideStore
    }
    return $autoHideStore
  })

  function triggerGlow() {
    clearTimeout(glowTimeout)
    isTransitionGlow = true
    glowTimeout = window.setTimeout(() => {
      isTransitionGlow = false
    }, 3000)
  }

  function triggerBlink() {
    isBlinking = true
    window.setTimeout(() => {
      isBlinking = false
    }, 150)
  }

  function resetHideTimer() {
    clearTimeout(hideTimeout)
    isTimeoutActive = false

    if (
      $devPinnedStore ||
      isHovering ||
      isSatelliteHovering ||
      isControlsDrawerOpen ||
      isHidden ||
      expanded ||
      !shouldCurrentPageAutoHide
    ) {
      return
    }

    const delaySeconds = $autoHideDelayStore || 5
    isTimeoutActive = true
    hideTimeout = window.setTimeout(() => {
      isTimeoutActive = false
      if (activePage === MediaPage && !isMediaActive) {
        navigateTo(HomePage, homeConfig, false)
      } else {
        hideWidget()
      }
    }, delaySeconds * 1000)
  }

  function handleAutoHideSync() {
    void activePage
    void hideTrigger
    void isHovering
    void isSatelliteHovering
    void isHidden
    void expanded
    void shouldCurrentPageAutoHide
    void isMediaActive
    void $devPinnedStore
    void $autoHideDelayStore
    resetHideTimer()
  }
  $effect(handleAutoHideSync)

  // Derived Dimensions & Geometric Bounds
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

  let virtualMinOffset = $derived.by(() => {
    const halfW = currentIslandW / 2
    return Math.floor(12 + halfW - displayLayout.primaryCenter)
  })

  let virtualMaxOffset = $derived.by(() => {
    const halfW = currentIslandW / 2
    return Math.floor(
      displayLayout.virtualWidth - 12 - halfW - displayLayout.primaryCenter
    )
  })

  let monitorSnapOffsets = $derived.by(() =>
    computeMonitorSnapOffsets(
      displayLayout.monitors,
      displayLayout.virtualLeft,
      displayLayout.primaryCenter
    )
  )

  let activeLeftSatelliteFootprint = $derived.by(() =>
    getActiveSatelliteFootprint('left', $satellitesStore)
  )
  let activeRightSatelliteFootprint = $derived.by(() =>
    getActiveSatelliteFootprint('right', $satellitesStore)
  )

  let allowedOffsets = $derived.by(() =>
    computeAllowedOffsets(
      currentIslandW,
      activeMonitor,
      displayLayout.virtualLeft + displayLayout.primaryCenter,
      activeLeftSatelliteFootprint,
      activeRightSatelliteFootprint
    )
  )

  let currentLiveOffsetX = $state(0)

  function handleSmartEdgeOffsetAdjustment() {
    if (!gestures.isDragging && !isControlsDrawerOpen) {
      if (currentLiveOffsetX > allowedOffsets.max) {
        currentLiveOffsetX = allowedOffsets.max
        $hiddenPillOffsetXStore = currentLiveOffsetX
      } else if (currentLiveOffsetX < allowedOffsets.min) {
        currentLiveOffsetX = allowedOffsets.min
        $hiddenPillOffsetXStore = currentLiveOffsetX
      }
    }
  }
  $effect(handleSmartEdgeOffsetAdjustment)

  function syncStoreOffsetToLive(val: number) {
    if (!gestures.isDragging) {
      currentLiveOffsetX = val
    }
  }
  function handleHiddenPillOffsetSync() {
    syncStoreOffsetToLive($hiddenPillOffsetXStore)
  }
  $effect(handleHiddenPillOffsetSync)

  let clampedOffsetX = $derived.by(() => {
    if (Math.abs(currentLiveOffsetX) < 3) return 0
    if (gestures.isDragging) {
      return currentLiveOffsetX
    }
    return Math.max(
      allowedOffsets.min,
      Math.min(allowedOffsets.max, currentLiveOffsetX)
    )
  })

  bounds.getLiveOffset = () => clampedOffsetX
  bounds.getIsDragging = () => gestures.isDragging

  let glowOpacity = $derived(
    $glowIntensityStore === 'disabled'
      ? 0
      : $glowIntensityStore === 'vibrant'
        ? 1.0
        : 0.6
  )

  let activeOpacity = $derived(
    isHidden
      ? $opacityHiddenStore ?? 1.0
      : expanded
        ? $opacityExpandedStore ?? 1.0
        : $opacityMinimalStore ?? 1.0
  )

  function getBoundsForState(
    config: IslandConfig = islandConfig,
    isExp: boolean = expanded,
    isHid: boolean = isHidden
  ) {
    const isTargetHome =
      config === homeConfig || (config === islandConfig && activePage === HomePage)
    const isTargetMedia =
      config === mediaConfig || (config === islandConfig && activePage === MediaPage)

    const effectiveConfig =
      !isExp && !isHid && !isTargetHome && !isTargetMedia
        ? getFallbackPage().config
        : config

    const isEffectiveHome =
      effectiveConfig === homeConfig ||
      (effectiveConfig === islandConfig && activePage === HomePage)

    const baseW = isHid
      ? hiddenBounds.w
      : isExp
        ? effectiveConfig.expanded.w
        : isEffectiveHome
          ? calculateHomeMinimalWidth($homeWidgetsVisibilityStore, hasBattery)
          : effectiveConfig.minimal.w

    const baseH = isHid
      ? hiddenBounds.h
      : isExp
        ? effectiveConfig.expanded.h
        : effectiveConfig.minimal.h

    const maxMonW = activeMonitor?.width ? activeMonitor.width - 24 : 1920
    const boundedBaseW = isHid ? baseW : Math.min(baseW, maxMonW)
    const w = boundedBaseW + extraDropdownW
    const h = baseH + extraDropdownH
    const r = isHid
      ? hiddenBounds.r
      : isExp
        ? effectiveConfig.expanded.r
        : effectiveConfig.minimal.r

    return { w, h, r }
  }

  let wasExpandedBeforeHide = $state(false)

  async function navigateTo(
    page: Component,
    pageConfig: IslandConfig,
    expand: boolean = false
  ) {
    const nextBounds = getBoundsForState(pageConfig, expand, false)
    if (
      nextBounds.w > bounds.currentOSBounds.w ||
      nextBounds.h > bounds.currentOSBounds.h
    ) {
      await bounds.prepareExpansion(nextBounds, clampedOffsetX)
    }

    if (isHidden) {
      clearTimeout(dormantUnmountTimeout)
      dormantUnmountTimeout = undefined
      clearTimeout(memoryTrimTimeout)
      memoryTrimTimeout = undefined
      isPageDormant = false
      wasExpandedBeforeHide = false
      finishDragGesture()
      isHidden = false
      bounds.notifyHidden(false)
    }
    if (activePage !== page && expanded === expand) {
      triggerBlink()
    }
    islandAnim = expand
      ? '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'
      : '0.22s cubic-bezier(0.2, 0.8, 0.2, 1)'
    activePage = page
    islandConfig = pageConfig
    expanded = expand
    hideTrigger++

    let pageId = 'home'
    if (page === HomePage) pageId = 'home'
    else if (page === MediaPage) pageId = 'media'
    else if (page === NotificationPage) pageId = 'notifications'
    else if (page === NotesPage) pageId = 'notes'
    else if (page === AppsPage) pageId = 'apps'
    else if (page === SettingsPage) pageId = 'settings'
    else if (page === CalculatorPage) pageId = 'calculator'

    $activePageIdStore = pageId
    try {
      Events.Emit('devtest-sync-page', { pageId })
    } catch {}

    bounds.syncBounds(nextBounds, clampedOffsetX, islandNode, false, false)

    triggerGlow()
  }

  setContext('navigate', navigateTo)

  function showWidget() {
    clearHoverGrace()
    clearTimeout(dormantUnmountTimeout)
    dormantUnmountTimeout = undefined
    clearTimeout(memoryTrimTimeout)
    memoryTrimTimeout = undefined
    isPageDormant = false

    if (!isHidden) {
      resetHideTimer()
      return
    }

    finishDragGesture()
    wasExpandedBeforeHide = false
    expanded = false

    const fallback = getFallbackPage()
    if (activePage !== fallback.page) {
      activePage = fallback.page as Component
      islandConfig = fallback.config
      $activePageIdStore = fallback.page === HomePage ? 'home' : 'media'
    }

    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'
    isHidden = false
    bounds.notifyHidden(false)
    const nextBounds = getBoundsForState(islandConfig, false, false)
    bounds.syncBounds(nextBounds, clampedOffsetX, islandNode, false, false)
    resetHideTimer()
  }

  function hideWidget() {
    if (isHidden) return
    clearHoverGrace()
    clearTimeout(hideTimeout)
    isTimeoutActive = false
    isControlsDrawerOpen = false
    finishDragGesture()

    wasExpandedBeforeHide = expanded
    isDockVisible = false
    const duration = wasExpandedBeforeHide ? 300 : 220
    islandAnim = wasExpandedBeforeHide
      ? '0.30s cubic-bezier(0.2, 0.8, 0.2, 1)'
      : '0.22s cubic-bezier(0.2, 0.8, 0.2, 1)'

    isHidden = true
    const nextBounds = getBoundsForState(islandConfig, false, true)
    bounds.syncBounds(nextBounds, clampedOffsetX, islandNode, false, true)

    dormantUnmountTimeout = window.setTimeout(() => {
      if (isHidden && !isHovering) {
        isPageDormant = true
        expanded = false
        const fallback = getFallbackPage()
        if (activePage !== fallback.page) {
          activePage = fallback.page as Component
          islandConfig = fallback.config
          $activePageIdStore = fallback.page === HomePage ? 'home' : 'media'
        }
        bounds.notifyHidden(true)
      }
    }, duration + 50)

    memoryTrimTimeout = window.setTimeout(() => {
      if (isHidden && isPageDormant) {
        bounds.trimMemory()
      }
    }, 1500)
  }

  setContext('hidePill', hideWidget)

  function syncOSBounds() {
    if (gestures.isDragging || gestures.isTrackRegionActive) return
    const target = getBoundsForState()
    bounds.syncBounds(
      target,
      clampedOffsetX,
      islandNode,
      gestures.isDragging,
      isHidden
    )
  }

  $effect(syncOSBounds)

  // Drag Gesture Helpers
  function finishDragGesture() {
    gestures.finishDrag(undefined, () => {
      const resting = Math.max(
        allowedOffsets.min,
        Math.min(allowedOffsets.max, currentLiveOffsetX)
      )
      currentLiveOffsetX = resting
      $hiddenPillOffsetXStore = resting
      const target = getBoundsForState()
      bounds.syncBounds(target, resting, islandNode, false, isHidden)
    })
  }

  async function handleExpand() {
    if (expanded) return
    if (gestures.isDragging) finishDragGesture()
    if (isHidden) {
      isHidden = false
      bounds.notifyHidden(false)
    }
    const nextBounds = getBoundsForState(islandConfig, true, false)
    await bounds.prepareExpansion(nextBounds, clampedOffsetX)
    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)'
    expanded = true
    isControlsDrawerOpen = false
    bounds.syncBounds(nextBounds, clampedOffsetX, islandNode, false, false)
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

  function handleSatelliteMouseEnter() {
    clearHoverGrace()
    isSatelliteHovering = true
    isHovering = false
    clearTimeout(hideTimeout)
  }

  function handleSatelliteMouseLeave() {
    clearHoverGrace()
    hoverGraceTimeout = window.setTimeout(() => {
      isSatelliteHovering = false
      isHovering = false
      hideTrigger++
    }, HOVER_GRACE_MS)
  }

  onMount(() => {
    const disableGlobalContextMenu = (e: MouseEvent) => e.preventDefault()
    window.addEventListener('contextmenu', disableGlobalContextMenu)

    $themeStore = 'dark'
    document.documentElement.setAttribute('data-theme', 'dark')

    $hiddenPillOffsetXStore = 0
    currentLiveOffsetX = 0
    const b = getBoundsForState()
    bounds.syncBounds(b, 0, islandNode, false, isHidden)

    const handleOverlayChange = (e: Event) => {
      const detail = (e as CustomEvent).detail as
        | { open: boolean; x?: number; y?: number; w?: number; h?: number }
        | undefined
      if (!detail?.open) {
        extraDropdownH = 0
        extraDropdownW = 0
        return
      }
      const { x = 0, y = 0, w = 0, h = 0 } = detail
      const margin = 8
      extraDropdownW = Math.max(0, x + w - window.innerWidth + margin)
      extraDropdownH = Math.max(0, y + h - window.innerHeight + margin)
    }

    const handleQuit = () => {
      if (isQuitting) return
      isQuitting = true
      if ($notifSoundStore) {
        playNotifChime($notifVolumeStore * 0.4)
      }
      window.setTimeout(() => {
        Quit().catch(() => {})
      }, 360)
    }

    window.addEventListener('vinta-overlay', handleOverlayChange)
    window.addEventListener('vinta-quit', handleQuit)

    const refreshDisplay = async () => {
      try {
        const layout = await GetDisplayLayout()
        if (layout && layout.virtualWidth > 0) {
          displayLayout = layout
        }
      } catch {}
    }
    refreshDisplay()

    const handleResize = () => {
      refreshDisplay()
    }

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
      if (gestures.isDragging) {
        finishDragGesture()
      }
    }

    const handleWindowKeydownFailsafe = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && gestures.isDragging) {
        finishDragGesture()
      }
    }

    window.addEventListener('resize', handleResize)
    window.addEventListener('blur', handleGlobalDragFailsafe)
    window.addEventListener('pointercancel', handleGlobalDragFailsafe)
    window.addEventListener('keydown', handleWindowKeydownFailsafe)

    const cleanupEvents = initIslandEvents({
      showWidget,
      hideWidget,
      navigateTo,
      triggerGlow,
      getIsHidden: () => isHidden,
      getIsExpanded: () => expanded,
      getActivePage: () => activePage,
      getIsPageDormant: () => isPageDormant,
      onHideTrigger: () => {
        hideTrigger++
      },
      onMonitorChanged: (mon) => {
        displayLayout = {
          ...displayLayout,
          currentMonitor: mon
        }
      },
      onRefreshDisplay: refreshDisplay
    })

    const unsubDevTestNavigate = Events.On(
      'devtest-navigate',
      (ev: { data: unknown }) => {
        const data = ev?.data as { mode?: string; expand?: boolean }
        if (!data?.mode) return

        const pageMappings: Record<
          string,
          { page: Component; config: IslandConfig }
        > = {
          home: { page: HomePage, config: homeConfig },
          media: { page: MediaPage, config: mediaConfig },
          notifications: { page: NotificationPage, config: notifConfig },
          notes: { page: NotesPage, config: notesConfig },
          apps: { page: AppsPage, config: appsConfig },
          settings: { page: SettingsPage, config: settingsConfig },
          calculator: { page: CalculatorPage, config: calcConfig }
        }

        const match = pageMappings[data.mode]
        if (match) {
          $devPinnedStore = true
          navigateTo(match.page, match.config, Boolean(data.expand))
        }
      }
    )

    const unsubDevTestPin = Events.On('devtest-pin', (ev: { data: unknown }) => {
      const data = ev?.data as { pinned?: boolean }
      if (typeof data?.pinned === 'boolean') {
        $devPinnedStore = data.pinned
      }
    })

    const unsubDevTestTheme = Events.On('devtest-theme', (ev: { data: unknown }) => {
      const data = ev?.data as { theme?: 'dark' | 'light' }
      if (data?.theme === 'dark' || data?.theme === 'light') {
        $themeStore = data.theme
        document.documentElement.setAttribute('data-theme', data.theme)
      }
    })

    const unsubDevTestBounds = Events.On('devtest-bounds-overlay', (ev: { data: unknown }) => {
      const data = ev?.data as { show?: boolean }
      if (typeof data?.show === 'boolean') {
        $debugBoundsOverlay = data.show
        if (data.show) {
          document.body.classList.add('debug-bounds-overlay')
        } else {
          document.body.classList.remove('debug-bounds-overlay')
        }
      }
    })

    const unsubDevTestLogs = Events.On('devtest-logs', (ev: { data: unknown }) => {
      const data = ev?.data as { enabled?: boolean }
      if (typeof data?.enabled === 'boolean') {
        $debugLogs = data.enabled
      }
    })

    const unsubDevTestBattery = Events.On('devtest-battery', (ev: { data: unknown }) => {
      const data = ev?.data as { isSimulating?: boolean; data?: BatteryData }
      if (typeof data?.isSimulating === 'boolean') {
        isSimulatingBatteryStore.set(data.isSimulating)
        if (data.data) {
          batteryDataStore.set(data.data)
        }
      }
    })

    const unsubDevTestWireless = Events.On('devtest-wireless', (ev: { data: unknown }) => {
      const data = ev?.data as { isSimulating?: boolean; data?: WirelessData }
      if (typeof data?.isSimulating === 'boolean') {
        isSimulatingWirelessStore.set(data.isSimulating)
        if (data.data) {
          wirelessDataStore.set(data.data)
        }
      }
    })

    const unsubDevTestWeather = Events.On('devtest-weather', (ev: { data: unknown }) => {
      const data = ev?.data as {
        location?: WeatherLocation
        unit?: TemperatureUnit
        timezone?: string
        simulateOffline?: boolean
      }
      if (data?.location) weatherLocationStore.set(data.location)
      if (data?.unit) weatherUnitStore.set(data.unit)
      if (data?.timezone) timezoneStore.set(data.timezone)
      if (typeof data?.simulateOffline === 'boolean') simulateOfflineStore.set(data.simulateOffline)
    })

    return () => {
      finishDragGesture()
      cleanupEvents()
      bounds.cleanup()
      gestures.cleanup()
      unsubDevTestNavigate()
      unsubDevTestPin()
      unsubDevTestTheme()
      unsubDevTestBounds()
      unsubDevTestLogs()
      unsubDevTestBattery()
      unsubDevTestWireless()
      unsubDevTestWeather()
      clearHoverGrace()
      clearTimeout(dormantUnmountTimeout)
      clearTimeout(memoryTrimTimeout)
      clearTimeout(glowTimeout)
      clearTimeout(hideTimeout)
      clearTimeout(islandTransitionTimeout)
      window.removeEventListener('contextmenu', disableGlobalContextMenu)
      window.removeEventListener('vinta-overlay', handleOverlayChange)
      window.removeEventListener('vinta-quit', handleQuit)
      window.removeEventListener('resize', handleResize)
      window.removeEventListener('blur', handleGlobalDragFailsafe)
      window.removeEventListener('pointercancel', handleGlobalDragFailsafe)
      window.removeEventListener('keydown', handleWindowKeydownFailsafe)
    }
  })
</script>

<div
  bind:this={islandContainerNode}
  class="island-container"
  role="application"
  onfocusin={(e: FocusEvent) => {
    const t = e.target as HTMLElement
    if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable)) {
      FocusWindow().catch(() => {})
    }
  }}
  onpointerdown={() => {
    FocusWindow().catch(() => {})
  }}
  style="position: absolute; left: {displayLayout.primaryCenter}px; top: {Math.max(
    0,
    activeMonitor.top - displayLayout.virtualTop
  )}px; transform: translateX(-50%);"
>
  <div
    class="island-wrapper"
    style="
      pointer-events: {isQuitting ? 'none' : 'auto'};
      transform: translateX({clampedOffsetX}px);
      transition: {gestures.isDragging ? 'none' : islandAnim};
    "
  >
    <MainIsland
      bind:islandNode
      {activePage}
      {expanded}
      {isHidden}
      {isQuitting}
      {isTransitionGlow}
      {isBlinking}
      {isPageDormant}
      {isDockVisible}
      {isTimeoutActive}
      {shouldCurrentPageAutoHide}
      {isMediaActive}
      {hideTrigger}
      {islandConfig}
      {currentMinimalW}
      {hiddenBounds}
      {activeOpacity}
      {glowOpacity}
      {islandAnim}
      activeMonitorWidth={activeMonitor.width}
      {gestures}
      {virtualMinOffset}
      {virtualMaxOffset}
      {monitorSnapOffsets}
      {currentLiveOffsetX}
      {wasExpandedBeforeHide}
      onStartDragRunway={() => {
        const { h, r } = getBoundsForState()
        bounds.updateBounds(-1, Math.max(h, 200), r, 0)
      }}
      onUpdateOffset={(next) => {
        currentLiveOffsetX = next
      }}
      onFinishDrag={finishDragGesture}
      onExpand={handleExpand}
      onShowWidget={showWidget}
      onHideWidget={hideWidget}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    />

    {#if !isQuitting && !isHidden}
      <ControlsSatellite
        {activePage}
        {expanded}
        expandedHeight={islandConfig.expanded.h}
        expandedWidth={islandConfig.expanded.w}
        activeMonitorWidth={activeMonitor.width}
        monitorLeft={activeMonitor.left}
        primaryCenter={displayLayout.primaryCenter}
        {clampedOffsetX}
        active={!isIslandTransitioning}
        onmouseenter={handleSatelliteMouseEnter}
        onmouseleave={handleSatelliteMouseLeave}
        onDrawerChange={handleControlsDrawerChange}
      />
    {/if}
  </div>
</div>

<CustomTooltip />

<style src="./App.css"></style>
