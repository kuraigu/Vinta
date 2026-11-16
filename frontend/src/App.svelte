<!--
  @component App
  @description Root application container. Initializes theme variables, loads user
  settings, binds global keyboard shortcuts (Esc to collapse, hotkeys to toggle pages),
  and coordinates layout between the main island capsule and floating satellite widgets.
-->
<script lang="ts">
  import { onMount, setContext, untrack, type Component } from 'svelte'
  import { get } from 'svelte/store'
  import {
    FocusWindow,
    Quit,
    GetDisplayLayout,
    SendMediaControl,
    SaveNotifications,
    RequestMediaUpdate,
    SetUIScale
  } from '../bindings/Vinta/backend/app.js'
  import type { DisplayLayout } from '../bindings/Vinta/backend/models.js'
  import { Events } from '$lib/events'

  // Pages & Configs
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
  import TimerPage, { config as timerConfig } from './pages/TimerPage.svelte'
  import CalendarPage, {
    config as calendarConfig
  } from './pages/CalendarPage.svelte'
  import WeatherPage, {
    config as weatherConfig
  } from './pages/WeatherPage.svelte'
  import FileShelfPage, {
    config as fileShelfConfig
  } from './pages/FileShelfPage.svelte'
  import OnboardingPage, {
    config as onboardingConfig
  } from './pages/OnboardingPage.svelte'

  // Components & Overlays
  import MainIsland from './components/MainIsland.svelte'
  import ControlsSatellite from './components/ControlsSatellite.svelte'
  import NotificationSatellite from './components/NotificationSatellite.svelte'
  import CustomTooltip from './components/CustomTooltip.svelte'
  import VirtualCursor from './devtest/VirtualCursor.svelte'

  // Services & Utils
  import { getFallbackPage } from './utils/navigation'
  import { initIslandEvents } from './services/islandEvents'
  import { WindowBoundsManager } from './services/windowBounds'
  import {
    IslandGestureController,
    computeAllowedOffsets,
    computeMonitorSnapOffsets,
    getMonitorForOffset
  } from './services/islandGestures'
  import {
    playNotifChime,
    startAlarmRingingChime,
    stopAlarmRingingChime
  } from './services/audioService'
  import { applyThemeToDOM, refreshStylesheetsAndTheme } from './themes/engine'

  // Stores
  import { fileShelfStore, clearFileShelf } from './stores/fileShelf'
  import {
    addAlarm,
    cleanShowcaseAlarms,
    dismissRingingAlarm
  } from './stores/timerApp'
  import { calendarStore } from './stores/calendarStore'
  import {
    satellitesStore,
    getActiveSatelliteFootprint,
    resyncAllSatellites
  } from './stores/satellites'
  import { hideTooltip, setCachedUiScale } from './stores/tooltip'
  import {
    debugBoundsOverlay,
    debugLogs,
    debugClassInspector
  } from './devtest/debugConfig'
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
    uiScaleStore,
    monitorUiScalesStore,
    animScaleStore,
    getMonitorScaleKey,
    getMonitorBasePhysicalScale,
    getMaxScaleForMonitor,
    hiddenPillOffsetXStore,
    isIslandHiddenStore,
    notifSoundStore,
    notifVolumeStore,
    glowIntensityStore,
    globalMediaStore,
    type IslandConfig,
    homeWidgetsVisibilityStore,
    hasCompletedOnboardingStore,
    syncSettingsToDisk,
    homeMeasuredHeightStore,
    calculateHomeMinimalWidth,
    calculateHomeExpandedHeight,
    notifData,
    updateInfoStore,
    batteryDataStore,
    wirelessDataStore,
    weatherLocationStore,
    weatherUnitStore,
    timezoneStore,
    simulateOfflineStore,
    simulateWeatherStore,
    isSimulatingBatteryStore,
    isSimulatingWirelessStore,
    devPinnedStore,
    activePageIdStore,
    activeOverlayStore,
    notificationsOverlayTabStore,
    type OverlayPageId,
    type NotificationsOverlayTab,
    isMediaDrawerOpenStore,
    timerAppStore,
    isAnyTimerRunning,
    type BatteryData,
    type WirelessData,
    type WeatherLocation,
    type TemperatureUnit,
    displayLayoutStore,
    alarmVolumeStore,
    alarmSoundPresetStore,
    simulatedClockStore
  } from './store'

  // Subsystem Managers & Static Maps
  const bounds = new WindowBoundsManager()
  const gestures = new IslandGestureController()
  const HOVER_GRACE_MS = 250

  const PAGE_MAPPINGS: Record<
    string,
    { page: Component; config: IslandConfig }
  > = {
    home: { page: HomePage, config: homeConfig },
    media: { page: MediaPage, config: mediaConfig },
    notifications: { page: NotificationPage, config: notifConfig },
    notes: { page: NotesPage, config: notesConfig },
    apps: { page: AppsPage, config: appsConfig },
    settings: { page: SettingsPage, config: settingsConfig },
    calculator: { page: CalculatorPage, config: calcConfig },
    timer: { page: TimerPage, config: timerConfig },
    calendar: { page: CalendarPage, config: calendarConfig },
    weather: { page: WeatherPage, config: weatherConfig },
    fileshelf: { page: FileShelfPage, config: fileShelfConfig }
  }

  const PAGE_ID_BY_COMPONENT = new Map<Component, string>([
    [HomePage, 'home'],
    [MediaPage, 'media'],
    [NotificationPage, 'notifications'],
    [NotesPage, 'notes'],
    [AppsPage, 'apps'],
    [SettingsPage, 'settings'],
    [CalculatorPage, 'calculator'],
    [TimerPage, 'timer'],
    [CalendarPage, 'calendar'],
    [WeatherPage, 'weather'],
    [FileShelfPage, 'fileshelf']
  ])

  // DOM References
  let islandContainerNode = $state<HTMLDivElement | null>(null)
  let islandWrapperNode = $state<HTMLDivElement | null>(null)
  let islandNode = $state<HTMLDivElement | null>(null)

  // Core Island State & Dynamic Routing
  let activePage = $state<Component>(
    $hasCompletedOnboardingStore ? HomePage : OnboardingPage
  )
  let islandConfig = $state<IslandConfig>(
    $hasCompletedOnboardingStore ? homeConfig : onboardingConfig
  )
  let expanded = $state(!$hasCompletedOnboardingStore)
  let isDockVisible = $derived(expanded && activePage !== OnboardingPage)

  let isHidden = $state(false)
  let isBlinking = $state(false)
  let isPageDormant = $state(false)
  let isQuitting = $state(false)
  let isTransitionGlow = $state(false)
  let hideTrigger = $state(0)
  function formatIslandAnim(
    baseSec: number,
    easing: string = 'cubic-bezier(0.18, 0.9, 0.32, 1)'
  ): string {
    const scale = $animScaleStore
    if (scale <= 0) return '0s linear'
    return `${+(baseSec * scale).toFixed(3)}s ${easing}`
  }

  function scaleAnimMs(baseMs: number): number {
    const scale = $animScaleStore
    return scale <= 0 ? 0 : Math.round(baseMs * scale)
  }

  let islandAnim = $state('0.65s cubic-bezier(0.18, 0.9, 0.32, 1)')

  let wasExpandedBeforeHide = $state(false)
  let wasHiddenBeforeNotif = $state(false)
  let previousExpanded: boolean | undefined = undefined

  // Timers
  let glowTimeout: number | undefined
  let hideTimeout: number | undefined
  let hoverGraceTimeout: number | undefined
  let hoverShowTimeout: number | undefined

  interface NavHistoryItem {
    page: Component
    config: IslandConfig
  }
  let navHistory: NavHistoryItem[] = []

  // Memoized booleans so 1-second timer ticks don't re-trigger auto-hide or MainIsland effects
  let isTimerActive = $derived(isAnyTimerRunning($timerAppStore))
  let ringingAlarm = $derived($timerAppStore.ringingAlarm)

  let isMediaActive = $derived.by(() => {
    if (!$globalMediaStore.isAvailable) return false
    if ($globalMediaStore.isPlaying) return true
    if (Array.isArray($globalMediaStore.sessions)) {
      return $globalMediaStore.sessions.some((s) => Boolean(s.isPlaying))
    }
    return false
  })

  // Multi-Monitor Geometry & Derived Bounds
  let displayLayout = $state<DisplayLayout>({
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
      isPortrait: false,
      dpiScale: 1.0
    },
    monitors: []
  })

  let currentLiveOffsetX = $state(0)
  let extraDropdownW = $state(0)
  let extraDropdownH = $state(0)

  let cachedWindowDpr = $state(
    typeof window !== 'undefined' && window.devicePixelRatio > 0
      ? window.devicePixelRatio
      : 1.0
  )

  function getWindowDpr(): number {
    return cachedWindowDpr
  }

  let activeMonitor = $derived.by(() => {
    const baseCenter = displayLayout.virtualLeft + displayLayout.primaryCenter
    const resolved = getMonitorForOffset(
      currentLiveOffsetX,
      displayLayout.monitors,
      baseCenter,
      getWindowDpr()
    )
    return (
      resolved ||
      displayLayout.currentMonitor || {
        id: 1,
        name: 'Primary Display',
        left: 0,
        top: 0,
        width: 1920,
        height: 1080,
        isPrimary: true,
        isPortrait: false,
        dpiScale: 1.0
      }
    )
  })

  let monitorCustomScale = $derived.by(() => {
    const key = getMonitorScaleKey(activeMonitor)
    const perMonitor = $monitorUiScalesStore?.[key]
    if (typeof perMonitor === 'number' && perMonitor >= 1.0) {
      return perMonitor
    }
    return activeMonitor?.isPrimary ? $uiScaleStore || 1.0 : 1.0
  })

  let maxScaleForActiveMonitor = $derived.by(() =>
    getMaxScaleForMonitor(
      activeMonitor?.width,
      activeMonitor?.height,
      activeMonitor?.dpiScale
    )
  )

  let effectiveUiScale = $derived.by(() => {
    const clampedCustom = Math.max(
      1.0,
      Math.min(maxScaleForActiveMonitor, monitorCustomScale)
    )
    const windowDpr = getWindowDpr()
    const basePhysical = getMonitorBasePhysicalScale(
      activeMonitor?.width,
      activeMonitor?.height,
      activeMonitor?.dpiScale
    )
    return clampedCustom * (basePhysical / windowDpr)
  })

  let hiddenBounds = $derived(
    $autoHideDisplayStore === 'clock'
      ? { w: 128, h: 28, r: 14 }
      : { w: 112, h: 12, r: 6 }
  )

  let hasBattery = $derived($batteryDataStore.hasBattery)

  let currentMinimalW = $derived(
    activePage === HomePage
      ? calculateHomeMinimalWidth($homeWidgetsVisibilityStore, hasBattery)
      : islandConfig.minimal.w
  )

  let hasActiveNotification = $derived(
    $notifData.some(
      (n) =>
        !n.isUpdate &&
        !n.title?.toLowerCase().includes('update available') &&
        !(
          (n.appName?.toLowerCase() === 'vinta' ||
            n.aumid?.toLowerCase().includes('vinta')) &&
          n.title?.toLowerCase().includes('update')
        )
    )
  )

  let hasActiveUpdateBanner = $derived(
    Boolean($updateInfoStore?.updateAvailable)
  )

  let currentExpandedH = $derived.by(() => {
    const baseH =
      activePage === HomePage
        ? $homeMeasuredHeightStore > 0
          ? Math.max(
              islandConfig.expanded.minH ?? 260,
              Math.min(
                islandConfig.expanded.maxH ?? 640,
                $homeMeasuredHeightStore
              )
            )
          : calculateHomeExpandedHeight(
              $homeWidgetsVisibilityStore,
              hasActiveNotification,
              hasActiveUpdateBanner,
              islandConfig.expanded.minH ?? 260,
              islandConfig.expanded.maxH ?? islandConfig.expanded.h
            )
        : activePage === MediaPage && $isMediaDrawerOpenStore
          ? 335
          : islandConfig.expanded.h
    return $activeOverlayStore !== null ? Math.max(baseH, 380) : baseH
  })

  let currentIslandW = $derived(
    isHidden
      ? hiddenBounds.w
      : Math.min(
          expanded
            ? $activeOverlayStore !== null
              ? Math.max(islandConfig.expanded.w, 440)
              : islandConfig.expanded.w
            : currentMinimalW,
          activeMonitor.width ? activeMonitor.width - 24 : 1920
        )
  )

  let scaledIslandW = $derived(currentIslandW * effectiveUiScale)

  let virtualMinOffset = $derived.by(() => {
    const halfW = scaledIslandW / 2
    const dpr = getWindowDpr()
    return Math.floor(12 + halfW - displayLayout.primaryCenter / dpr)
  })

  let virtualMaxOffset = $derived.by(() => {
    const halfW = scaledIslandW / 2
    const dpr = getWindowDpr()
    return Math.floor(
      displayLayout.virtualWidth / dpr -
        12 -
        halfW -
        displayLayout.primaryCenter / dpr
    )
  })

  let monitorSnapOffsets = $derived.by(() =>
    computeMonitorSnapOffsets(
      displayLayout.monitors,
      displayLayout.virtualLeft,
      displayLayout.primaryCenter,
      getWindowDpr()
    )
  )

  let activeLeftSatelliteFootprint = $derived.by(
    () =>
      getActiveSatelliteFootprint('left', $satellitesStore) * effectiveUiScale
  )
  let activeRightSatelliteFootprint = $derived.by(
    () =>
      getActiveSatelliteFootprint('right', $satellitesStore) * effectiveUiScale
  )

  let allowedOffsets = $derived.by(() =>
    computeAllowedOffsets(
      scaledIslandW,
      activeMonitor,
      displayLayout.virtualLeft + displayLayout.primaryCenter,
      activeLeftSatelliteFootprint,
      activeRightSatelliteFootprint,
      getWindowDpr()
    )
  )

  let clampedOffsetX = $derived.by(() => {
    if (Math.abs(currentLiveOffsetX) < 3) return 0
    if (gestures.isDragging || gestures.isInertia) {
      return currentLiveOffsetX
    }
    return Math.max(
      allowedOffsets.min,
      Math.min(allowedOffsets.max, currentLiveOffsetX)
    )
  })

  let glowOpacity = $derived(
    $glowIntensityStore === 'disabled'
      ? 0
      : $glowIntensityStore === 'vibrant'
        ? 1.0
        : 0.6
  )

  let activeOpacity = $derived(
    isHidden
      ? ($opacityHiddenStore ?? 1.0)
      : expanded
        ? ($opacityExpandedStore ?? 1.0)
        : ($opacityMinimalStore ?? 1.0)
  )

  function isGestureBusy(): boolean {
    return (
      gestures.isDragging ||
      gestures.isTrackRegionActive ||
      gestures.isSettling ||
      gestures.isInertia
    )
  }

  bounds.getLiveOffset = () => clampedOffsetX
  bounds.getIsDragging = isGestureBusy

  /**
   * Pre-calculates the target clamped horizontal offset for a specific island width,
   * taking monitor bounds and active satellite footprints into account.
   */
  function calculateTargetOffsetForWidth(
    width: number,
    offset: number = currentLiveOffsetX
  ): number {
    const range = computeAllowedOffsets(
      width * effectiveUiScale,
      activeMonitor,
      displayLayout.virtualLeft + displayLayout.primaryCenter,
      activeLeftSatelliteFootprint,
      activeRightSatelliteFootprint,
      getWindowDpr()
    )
    if (Math.abs(offset) < 3) return 0
    return Math.max(range.min, Math.min(range.max, offset))
  }

  function getBoundsForState(
    config: IslandConfig = islandConfig,
    isExp: boolean = expanded,
    isHid: boolean = isHidden
  ) {
    const isTargetHome =
      config === homeConfig ||
      (config === islandConfig && activePage === HomePage)
    const isTargetMedia =
      config === mediaConfig ||
      (config === islandConfig && activePage === MediaPage)
    const isTargetNotif =
      config === notifConfig ||
      (config === islandConfig && activePage === NotificationPage)

    const effectiveConfig =
      !isExp && !isHid && !isTargetHome && !isTargetMedia && !isTargetNotif
        ? getFallbackPage().config
        : config

    const isEffectiveHome =
      effectiveConfig === homeConfig ||
      (effectiveConfig === islandConfig && activePage === HomePage)

    const rawBaseW = isHid
      ? hiddenBounds.w
      : isExp
        ? effectiveConfig.expanded.w
        : isEffectiveHome
          ? calculateHomeMinimalWidth($homeWidgetsVisibilityStore, hasBattery)
          : effectiveConfig.minimal.w

    const baseW =
      !isHid && isExp && $activeOverlayStore !== null
        ? Math.max(rawBaseW, 440)
        : rawBaseW

    const rawBaseH = isHid
      ? hiddenBounds.h
      : isExp
        ? isEffectiveHome
          ? $homeMeasuredHeightStore > 0
            ? Math.max(
                effectiveConfig.expanded.minH ?? 260,
                Math.min(
                  effectiveConfig.expanded.maxH ?? 640,
                  $homeMeasuredHeightStore
                )
              )
            : calculateHomeExpandedHeight(
                $homeWidgetsVisibilityStore,
                hasActiveNotification,
                hasActiveUpdateBanner,
                effectiveConfig.expanded.minH ?? 260,
                effectiveConfig.expanded.maxH ?? effectiveConfig.expanded.h
              )
          : effectiveConfig.expanded.h
        : effectiveConfig.minimal.h

    const baseH =
      !isHid && isExp && $activeOverlayStore !== null
        ? Math.max(rawBaseH, 380)
        : rawBaseH

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

  // Hover, Auto-Hide & Visual Feedback
  let isHovering = $state(false)
  let isSatelliteHovering = $state(false)
  let isTimeoutActive = $state(false)

  /**
   * Evaluates whether the current page should engage auto-hide countdown.
   * Notifications always auto-dismiss, while Home and Media follow user settings.
   */
  let shouldCurrentPageAutoHide = $derived.by(() => {
    if (activePage === NotificationPage) return true
    if (activePage === HomePage) return $homeAutoHideStore
    if (activePage === MediaPage) {
      return !isMediaActive ? true : $mediaAutoHideStore
    }
    if (activePage === TimerPage) {
      return !isTimerActive
    }
    return $autoHideStore
  })

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
      if (activePage === NotificationPage) {
        if (wasHiddenBeforeNotif) {
          wasHiddenBeforeNotif = false
          hideWidget()
        } else {
          const fallback = getFallbackPage()
          navigateTo(fallback.page as Component, fallback.config, false)
        }
      } else if (activePage === TimerPage && !isTimerActive) {
        const fallback = getFallbackPage()
        navigateTo(fallback.page as Component, fallback.config, false)
      } else if (activePage === MediaPage && !isMediaActive) {
        navigateTo(HomePage, homeConfig, false)
      } else {
        hideWidget()
      }
    }, delaySeconds * 1000)
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

  let collapseOnOverlayClose = false

  function openOverlay(
    overlay: OverlayPageId,
    options?: { tab?: NotificationsOverlayTab; collapseOnClose?: boolean }
  ) {
    if (!overlay) {
      closeOverlay()
      return
    }
    if (options?.tab) {
      $notificationsOverlayTabStore = options.tab
    }
    collapseOnOverlayClose = Boolean(options?.collapseOnClose)
    if (activePage === NotificationPage) {
      const fallback = getFallbackPage()
      activePage = fallback.page as Component
      islandConfig = fallback.config
      $activePageIdStore = PAGE_ID_BY_COMPONENT.get(activePage) ?? 'home'
    }
    if (isHidden) {
      isHidden = false
      bounds.notifyHidden(false)
    }
    isPageDormant = false
    clearTimeout(hideTimeout)
    isTimeoutActive = false
    $activeOverlayStore = overlay
    expanded = true
    const nextBounds = getBoundsForState(islandConfig, true, false)
    const targetOffset = calculateTargetOffsetForWidth(
      nextBounds.w,
      currentLiveOffsetX
    )
    if (targetOffset !== currentLiveOffsetX) {
      currentLiveOffsetX = targetOffset
      $hiddenPillOffsetXStore = targetOffset
    }
    bounds.currentOSBounds = { ...nextBounds }
    islandAnim = formatIslandAnim(0.65)
    bounds.startTrackingTransition(islandNode, scaleAnimMs(660), false)
    triggerGlow()
  }

  function closeOverlay() {
    if ($activeOverlayStore === null) return
    $activeOverlayStore = null
    if (collapseOnOverlayClose) {
      collapseOnOverlayClose = false
      expanded = false
      const nextBounds = getBoundsForState(islandConfig, false, false)
      const targetOffset = calculateTargetOffsetForWidth(
        nextBounds.w,
        currentLiveOffsetX
      )
      if (targetOffset !== currentLiveOffsetX) {
        currentLiveOffsetX = targetOffset
        $hiddenPillOffsetXStore = targetOffset
      }
      bounds.currentOSBounds = { ...nextBounds }
      islandAnim = formatIslandAnim(0.54)
      bounds.startTrackingTransition(islandNode, scaleAnimMs(550), false)
      resetHideTimer()
      return
    }
    const nextBounds = getBoundsForState(islandConfig, expanded, false)
    bounds.currentOSBounds = { ...nextBounds }
    islandAnim = formatIslandAnim(0.56)
    bounds.startTrackingTransition(islandNode, scaleAnimMs(570), false)
  }

  // Navigation, Show/Hide & Drag Gestures
  function navigateTo(
    page: Component,
    pageConfig: IslandConfig,
    expand: boolean = false,
    isBackNav: boolean = false
  ) {
    if (expand && page === NotificationPage) {
      openOverlay('notifications', {
        tab: 'notifications',
        collapseOnClose: !expanded && activePage === NotificationPage
      })
      return
    }
    if (expand && page === AppsPage) {
      openOverlay('apps')
      return
    }
    if ($activeOverlayStore !== null) {
      $activeOverlayStore = null
      collapseOnOverlayClose = false
    }
    if (page === HomePage && activePage === OnboardingPage) {
      hasCompletedOnboardingStore.set(true)
      try {
        localStorage.setItem(
          'vinta_has_completed_onboarding',
          JSON.stringify(true)
        )
      } catch {}
      syncSettingsToDisk(true)
    } else if (
      !$hasCompletedOnboardingStore &&
      page !== OnboardingPage &&
      page !== HomePage
    ) {
      return
    }

    if (!isBackNav && activePage && activePage !== page && expanded) {
      navHistory.push({ page: activePage as Component, config: islandConfig })
      if (navHistory.length > 20) {
        navHistory.shift()
      }
    }

    gestures.finishSettling()
    if (!expand && (gestures.isDragging || gestures.isInertia)) {
      finishDragGesture()
    }

    if (page !== NotificationPage) {
      wasHiddenBeforeNotif = false
    }

    bounds.stopTrackingTransition()
    bounds.cancelScheduledTrimMemory()
    isPageDormant = false

    if (isHidden) {
      wasExpandedBeforeHide = false
      finishDragGesture()
      isHidden = false
      bounds.notifyHidden(false)
    }
    if (activePage !== page && expanded === expand) {
      triggerBlink()
    }
    islandAnim = formatIslandAnim(expand ? 0.65 : 0.48)
    activePage = page
    islandConfig = pageConfig
    expanded = expand
    hideTrigger++

    const pageId = PAGE_ID_BY_COMPONENT.get(page) ?? 'home'
    $activePageIdStore = pageId
    try {
      Events.Emit('devtest-sync-page', { pageId })
    } catch {}

    const nextBounds = getBoundsForState(pageConfig, expand, false)
    const targetOffset = calculateTargetOffsetForWidth(
      nextBounds.w,
      currentLiveOffsetX
    )
    if (targetOffset !== currentLiveOffsetX) {
      currentLiveOffsetX = targetOffset
      $hiddenPillOffsetXStore = targetOffset
    }
    // Update currentOSBounds before startTrackingTransition so the transition union rect
    // expands immediately instead of leaving a click-through deadzone during the morph.
    bounds.currentOSBounds = { ...nextBounds }
    bounds.syncPillBounds(islandNode, false, true)
    bounds.startTrackingTransition(
      islandNode,
      scaleAnimMs(expand ? 660 : 500),
      false,
      () => {
        if (!isGestureBusy()) {
          bounds.recoverRestingBounds(
            getBoundsForState(),
            clampedOffsetX,
            islandNode,
            isHidden
          )
          bounds.syncPillBounds(islandNode, isHidden, true)
        }
        Events.Emit('vinta:state-change', {
          expanded: expand,
          isHidden,
          page: pageId
        })
      }
    )

    triggerGlow()
  }

  function goBack(): void {
    while (navHistory.length > 0) {
      const prev = navHistory.pop()
      if (prev && prev.page !== activePage) {
        navigateTo(prev.page, prev.config, true, true)
        return
      }
    }
    navigateTo(AppsPage as Component, appsConfig, true, true)
  }

  function showWidget() {
    clearHoverGrace()
    bounds.stopTrackingTransition()
    bounds.cancelScheduledTrimMemory()
    isPageDormant = false

    if (!isHidden) {
      resetHideTimer()
      return
    }

    if (typeof RequestMediaUpdate === 'function') {
      RequestMediaUpdate().catch(() => {})
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

    islandAnim = formatIslandAnim(0.65)
    isHidden = false
    bounds.notifyHidden(false)
    const nextBounds = getBoundsForState(islandConfig, false, false)
    const targetOffset = calculateTargetOffsetForWidth(
      nextBounds.w,
      currentLiveOffsetX
    )
    if (targetOffset !== currentLiveOffsetX) {
      currentLiveOffsetX = targetOffset
      $hiddenPillOffsetXStore = targetOffset
    }
    bounds.startTrackingTransition(islandNode, scaleAnimMs(660), false)
    resetHideTimer()
  }

  function hideWidget() {
    if (
      isHidden ||
      !$hasCompletedOnboardingStore ||
      activePage === OnboardingPage
    ) {
      return
    }
    wasHiddenBeforeNotif = false
    clearHoverGrace()
    clearTimeout(hideTimeout)
    isTimeoutActive = false
    isHovering = false
    finishDragGesture()

    wasExpandedBeforeHide = expanded
    const duration = wasExpandedBeforeHide ? 560 : 440
    islandAnim = formatIslandAnim(wasExpandedBeforeHide ? 0.56 : 0.44)

    isHidden = true
    bounds.startTrackingTransition(
      islandNode,
      scaleAnimMs(duration + 50),
      true,
      () => {
      if (!isHidden) return
      isPageDormant = true
      expanded = false
      const fallback = getFallbackPage()
      if (activePage !== fallback.page) {
        activePage = fallback.page as Component
        islandConfig = fallback.config
        $activePageIdStore = fallback.page === HomePage ? 'home' : 'media'
      }
      bounds.notifyHidden(true)
      bounds.recoverRestingBounds(
        getBoundsForState(islandConfig, false, true),
        clampedOffsetX,
        islandNode,
        true
      )
      Events.Emit('vinta:state-change', {
        expanded: false,
        isHidden: true,
        dormant: true
      })
    })
  }

  function finishDragGesture() {
    const wasInRunway = isGestureBusy()
    gestures.finishDrag(undefined, () => {
      const baseCenter = displayLayout.virtualLeft + displayLayout.primaryCenter
      const dpr = getWindowDpr()
      const targetMon =
        getMonitorForOffset(
          currentLiveOffsetX,
          displayLayout.monitors,
          baseCenter,
          dpr
        ) || activeMonitor
      const monAllowed = computeAllowedOffsets(
        currentIslandW * effectiveUiScale,
        targetMon,
        baseCenter,
        activeLeftSatelliteFootprint,
        activeRightSatelliteFootprint,
        dpr
      )
      const resting = Math.max(
        monAllowed.min,
        Math.min(monAllowed.max, currentLiveOffsetX)
      )
      const delta = Math.abs(resting - currentLiveOffsetX)
      const needsSettle = delta > 2
      currentLiveOffsetX = resting
      $hiddenPillOffsetXStore = resting
      if (displayLayout.currentMonitor?.id !== targetMon.id) {
        displayLayout = {
          ...displayLayout,
          currentMonitor: targetMon
        }
        $displayLayoutStore = displayLayout
      }
      Events.Emit('cancel-fullscreen-relocate')
      const target = getBoundsForState()

      if (wasInRunway) {
        if (!needsSettle) {
          gestures.finishSettling()
          bounds.cancelPendingTransition()
          bounds.recoverRestingBounds(target, resting, islandNode, isHidden)
        } else {
          gestures.isSettling = true
          gestures.isTrackRegionActive = true
          bounds.settleRunwayBounds(
            target,
            resting,
            islandWrapperNode,
            islandNode,
            () => {
              gestures.finishSettling()
              bounds.recoverRestingBounds(target, resting, islandNode, isHidden)
            }
          )
        }
      } else {
        bounds.syncBounds(target, resting, islandNode, false, isHidden)
      }
    })
  }

  async function handleExpand() {
    if (expanded) return
    if (gestures.isDragging || gestures.isInertia) finishDragGesture()
    if (activePage === NotificationPage) {
      openOverlay('notifications', {
        tab: 'notifications',
        collapseOnClose: true
      })
      return
    }
    if (isHidden) {
      isHidden = false
      bounds.notifyHidden(false)
    }
    const nextBounds = getBoundsForState(islandConfig, true, false)
    const targetOffset = calculateTargetOffsetForWidth(
      nextBounds.w,
      currentLiveOffsetX
    )
    if (targetOffset !== currentLiveOffsetX) {
      currentLiveOffsetX = targetOffset
      $hiddenPillOffsetXStore = targetOffset
    }
    bounds.cancelScheduledTrimMemory()
    bounds.currentOSBounds = { ...nextBounds }
    islandAnim = formatIslandAnim(0.65)
    bounds.startTrackingTransition(islandNode, scaleAnimMs(660), false)
    expanded = true
  }

  function recoverBoundsIfIdle() {
    if (!isGestureBusy()) {
      bounds.recoverRestingBounds(
        getBoundsForState(),
        clampedOffsetX,
        islandNode,
        isHidden
      )
    }
  }

  setContext('navigate', navigateTo)
  setContext('goBack', goBack)
  setContext('hidePill', hideWidget)
  setContext('openOverlay', openOverlay)
  setContext('closeOverlay', closeOverlay)

  // Reactive Effects
  $effect(function handleDebugBoundsOverlayEffect() {
    const enabled = $debugBoundsOverlay
    if (typeof document !== 'undefined') {
      document.body.classList.toggle('debug-bounds-overlay', enabled)
    }
  })

  $effect(function handleOnboardingSync() {
    if (!$hasCompletedOnboardingStore) {
      if (activePage !== OnboardingPage) {
        activePage = OnboardingPage
        islandConfig = onboardingConfig
        expanded = true
        isPageDormant = false
        isHidden = false
      }
    } else if (activePage === OnboardingPage) {
      activePage = HomePage
      islandConfig = homeConfig
      expanded = false
      $activePageIdStore = 'home'
    }
  })

  $effect(function handleExpandedTransitionEffect() {
    const isExp = expanded
    if (previousExpanded === undefined) {
      previousExpanded = isExp
      return
    }
    if (previousExpanded === isExp) return
    previousExpanded = isExp

    hideTooltip()
    gestures.finishSettling()

    if (isExp) {
      bounds.cancelScheduledTrimMemory()
    }
    bounds.currentOSBounds = { ...getBoundsForState() }
    bounds.syncPillBounds(islandNode, isHidden, true)
    bounds.startTrackingTransition(
      islandNode,
      scaleAnimMs(isExp ? 660 : 500),
      false,
      () => {
      if (!isGestureBusy()) {
        bounds.recoverRestingBounds(
          getBoundsForState(),
          clampedOffsetX,
          islandNode,
          isHidden
        )
        bounds.syncPillBounds(islandNode, isHidden, true)
      }
      Events.Emit('vinta:state-change', {
        expanded: isExp,
        isHidden,
        page: $activePageIdStore
      })
    })
  })

  $effect(function handleStateChangeTrim() {
    void activePage
    const exp = expanded
    const hidden = isHidden
    const dormant = isPageDormant
    untrack(() => {
      Events.Emit('vinta:state-change', {
        expanded: exp,
        isHidden: hidden,
        dormant
      })
    })
  })

  $effect(function handleAutoHideSync() {
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
    untrack(() => resetHideTimer())
  })

  $effect(function handleSmartEdgeOffsetAdjustment() {
    if (!gestures.isDragging && !gestures.isInertia) {
      let adjusted = false
      if (currentLiveOffsetX > allowedOffsets.max) {
        currentLiveOffsetX = allowedOffsets.max
        $hiddenPillOffsetXStore = currentLiveOffsetX
        adjusted = true
      } else if (currentLiveOffsetX < allowedOffsets.min) {
        currentLiveOffsetX = allowedOffsets.min
        $hiddenPillOffsetXStore = currentLiveOffsetX
        adjusted = true
      }
      if (adjusted && !gestures.isTrackRegionActive && !gestures.isSettling) {
        bounds.syncBounds(
          getBoundsForState(),
          currentLiveOffsetX,
          islandNode,
          false,
          isHidden
        )
      }
    }
  })

  $effect(function handleHiddenPillOffsetSync() {
    const val = $hiddenPillOffsetXStore
    if (!gestures.isDragging && !gestures.isInertia) {
      currentLiveOffsetX = val
    }
  })

  $effect(function handleIsHiddenSyncEffect() {
    $isIslandHiddenStore = isHidden
  })

  let wasRingingAlarm = false

  $effect(function handleRingingAlarmNavigation() {
    if (ringingAlarm) {
      wasRingingAlarm = true
      if (isHidden) showWidget()
      if (activePage !== TimerPage) {
        navigateTo(TimerPage, timerConfig, false)
      }
    } else if (wasRingingAlarm) {
      wasRingingAlarm = false
      if (!expanded && activePage === TimerPage && !isTimerActive) {
        const fallback = getFallbackPage()
        navigateTo(fallback.page as Component, fallback.config, false)
      }
    }
  })

  $effect(function syncOSBounds() {
    if (isGestureBusy()) return
    const target = getBoundsForState()
    bounds.syncBounds(target, clampedOffsetX, islandNode, false, isHidden)
  })

  $effect(function handleActiveThemeChange() {
    const currentTheme = $themeStore || 'dark-gold'
    applyThemeToDOM(currentTheme)
  })

  $effect(function handleUiScaleSyncEffect() {
    const scale = effectiveUiScale
    setCachedUiScale(scale)
    if (typeof document !== 'undefined') {
      document.documentElement.style.setProperty(
        '--vinta-ui-scale',
        String(scale)
      )
    }
    if (typeof SetUIScale === 'function') {
      void SetUIScale(scale).catch(() => {})
    }
    const rafId = requestAnimationFrame(() => {
      resyncAllSatellites()
      bounds.syncPillBounds(islandNode, isHidden, true)
    })
    return () => cancelAnimationFrame(rafId)
  })

  $effect(function handleAnimScaleSyncEffect() {
    if (typeof document !== 'undefined') {
      document.documentElement.style.setProperty(
        '--vinta-anim-scale',
        String($animScaleStore)
      )
    }
  })

  // Wails Event Subscriptions (Top-Level for Auto-Unsubscribe)
  function unwrapEventData<T = any>(ev: { data?: unknown } | any): T {
    const raw =
      ev?.data !== undefined
        ? Array.isArray(ev.data)
          ? ev.data[0]
          : ev.data
        : ev
    return (Array.isArray(raw) ? raw[0] : (raw?.data ?? raw)) as T
  }

  // DevTest: Navigation, Theme & Debug Overlays
  Events.On('devtest-navigate', (ev) => {
    const data = unwrapEventData<{ mode?: string; expand?: boolean }>(ev)
    if (!data?.mode) return

    const expand = Boolean(data.expand)
    if (
      !expand &&
      (data.mode === 'fileshelf' ||
        data.mode === 'notes' ||
        data.mode === 'apps' ||
        data.mode === 'settings')
    ) {
      const fallback = getFallbackPage()
      $devPinnedStore = false
      navigateTo(fallback.page as Component, fallback.config, false)
      return
    }

    const match = PAGE_MAPPINGS[data.mode]
    if (match) {
      $hasCompletedOnboardingStore = true
      hasCompletedOnboardingStore.set(true)
      $devPinnedStore = true
      navigateTo(match.page, match.config, expand)
    }
  })

  Events.On('devtest-pin', (ev) => {
    const data = unwrapEventData<{ pinned?: boolean }>(ev)
    if (typeof data?.pinned === 'boolean') {
      $devPinnedStore = data.pinned
    }
  })

  Events.On('devtest-theme', (ev) => {
    const data = ev?.data as { theme?: string }
    if (data?.theme) {
      $themeStore = data.theme
    }
  })

  Events.On('devtest-reload-css', async (ev) => {
    const data = ev?.data as {
      theme?: string
      mode?: string
      expand?: boolean
      pinned?: boolean
    }
    const targetTheme = data?.theme || $themeStore || 'dark-gold'
    await refreshStylesheetsAndTheme(targetTheme)
    try {
      sessionStorage.setItem(
        'vinta_devtest_reload_state',
        JSON.stringify({
          mode: data?.mode || $activePageIdStore || 'home',
          expand: typeof data?.expand === 'boolean' ? data.expand : expanded,
          pinned:
            typeof data?.pinned === 'boolean' ? data.pinned : $devPinnedStore
        })
      )
    } catch {}
    window.location.reload()
  })

  Events.On('devtest-bounds-overlay', (ev) => {
    const data = ev?.data as { show?: boolean }
    if (typeof data?.show === 'boolean') {
      $debugBoundsOverlay = data.show
      document.body.classList.toggle('debug-bounds-overlay', data.show)
    }
  })

  Events.On('devtest-logs', (ev) => {
    const data = ev?.data as { enabled?: boolean }
    if (typeof data?.enabled === 'boolean') {
      $debugLogs = data.enabled
    }
  })

  Events.On('devtest-class-inspector', (ev) => {
    const data = ev?.data as { enabled?: boolean }
    if (typeof data?.enabled === 'boolean') {
      $debugClassInspector = data.enabled
    }
  })

  // DevTest: Hardware, Weather & Clock Simulation
  Events.On('devtest-battery', (ev) => {
    const data = ev?.data as { isSimulating?: boolean; data?: BatteryData }
    if (typeof data?.isSimulating === 'boolean') {
      isSimulatingBatteryStore.set(data.isSimulating)
      if (data.data) batteryDataStore.set(data.data)
    }
  })

  Events.On('devtest-wireless', (ev) => {
    const data = ev?.data as { isSimulating?: boolean; data?: WirelessData }
    if (typeof data?.isSimulating === 'boolean') {
      isSimulatingWirelessStore.set(data.isSimulating)
      if (data.data) wirelessDataStore.set(data.data)
    }
  })

  Events.On('devtest-weather', (ev) => {
    const data = ev?.data as {
      location?: WeatherLocation
      unit?: TemperatureUnit
      timezone?: string
      simulateOffline?: boolean
    }
    if (data?.location) weatherLocationStore.set(data.location)
    if (data?.unit) weatherUnitStore.set(data.unit)
    if (data?.timezone) timezoneStore.set(data.timezone)
    if (typeof data?.simulateOffline === 'boolean') {
      simulateOfflineStore.set(data.simulateOffline)
    }
  })

  Events.On('devtest-weather-mock', (ev) => {
    const data = ev?.data as { mock?: any }
    simulateWeatherStore.set(data?.mock ?? null)
  })

  Events.On('devtest-clock', (ev) => {
    const data = unwrapEventData<{ time?: number | string }>(ev)
    if (data?.time) {
      const d = new Date(data.time)
      if (!isNaN(d.getTime())) {
        simulatedClockStore.set(d)
        return
      }
    }
    simulatedClockStore.set(null)
  })

  // DevTest: Media, Timer, FileShelf & Notifications
  Events.On('devtest-trigger-ringing-alarm', (ev) => {
    const data = unwrapEventData<{ label?: string; time?: string }>(ev)
    try {
      SendMediaControl('pause')
    } catch {}
    globalMediaStore.update((m) => ({ ...m, isPlaying: false }))
    timerAppStore.update((s) => ({
      ...s,
      ringingAlarm: {
        id: 'showcase-alarm',
        alarmId: 'mock_alarm_1',
        label: data?.label || '16:00 Launch Review',
        time: data?.time || '16:00',
        days: [],
        repeat: 'once',
        triggeredAt: Date.now()
      }
    }))
    startAlarmRingingChime(get(alarmVolumeStore), get(alarmSoundPresetStore))
  })

  Events.On('devtest-dismiss-alarm', () => {
    stopAlarmRingingChime()
    dismissRingingAlarm()
  })

  Events.On('devtest-timer-state', (ev) => {
    const data = unwrapEventData<any>(ev) || {}
    if (data.activeTab) {
      timerAppStore.update((s) => ({ ...s, activeTab: data.activeTab }))
    }
    if (data.newAlarm) {
      addAlarm(
        data.newAlarm.time,
        data.newAlarm.label,
        data.newAlarm.repeat || 'once',
        data.newAlarm.days || []
      )
    }
  })

  Events.On('devtest-media-state', (ev) => {
    const data = unwrapEventData<any>(ev) || {}
    globalMediaStore.update((m) => ({ ...m, ...data }))
  })

  Events.On('devtest-mock-fileshelf', (ev) => {
    const data = unwrapEventData<{ singleFile?: boolean }>(ev) || {}
    const now = Date.now()
    const pdfItem = {
      path: 'C:\\Projects\\proposal_final.pdf',
      name: 'proposal_final.pdf',
      size: 2450000,
      extension: 'pdf',
      isDir: false,
      category: 'document' as const,
      addedAt: now
    }
    const items = data.singleFile
      ? [pdfItem]
      : [
          pdfItem,
          {
            path: 'C:\\Projects\\hero_render_4k.png',
            name: 'hero_render_4k.png',
            size: 8900000,
            extension: 'png',
            isDir: false,
            category: 'image' as const,
            addedAt: now
          }
        ]
    fileShelfStore.update((s) => ({ ...s, items }))
  })

  Events.On('devtest-clear-fileshelf', () => {
    clearFileShelf()
  })

  Events.On('devtest-file-drop-hold', (ev) => {
    const data = ev?.data as { active?: boolean; progress?: number }
    Events.Emit('devtest:file-hold', data)
  })

  Events.On('devtest-clear-notifications', () => {
    notifData.set([])
    try {
      localStorage.removeItem('vinta_notifications')
    } catch {}
    if (typeof SaveNotifications === 'function') {
      SaveNotifications('[]').catch(() => {})
    }
  })

  Events.On('devtest-cleanup', () => {
    simulatedClockStore.set(null)
    cleanShowcaseAlarms()
    calendarStore.update((s) => ({
      ...s,
      events: s.events.filter(
        (e) =>
          !e.title.toLowerCase().includes('showcase') &&
          !e.title.toLowerCase().includes('launch') &&
          !e.title.toLowerCase().includes('vinta') &&
          !e.title.toLowerCase().includes('demo')
      )
    }))
    timerAppStore.update((s) => ({
      ...s,
      ringingAlarm: null,
      focus: {
        ...s.focus,
        isRunning: false,
        timeRemaining: 0
      }
    }))
    fileShelfStore.update((s) => ({ ...s, items: [] }))
    try {
      localStorage.removeItem('vinta_file_shelf_paths')
    } catch {}
    notifData.set([])
    try {
      localStorage.removeItem('vinta_notifications')
    } catch {}
    if (typeof SaveNotifications === 'function') {
      SaveNotifications('[]').catch(() => {})
    }
    stopAlarmRingingChime()
  })

  // Mount & Window Lifecycle
  onMount(() => {
    const disableGlobalContextMenu = (e: MouseEvent) => e.preventDefault()
    const disableDefaultBrowserDrag = (e: DragEvent) => {
      const target = e.target as HTMLElement | null
      if (!target || target.closest('[draggable="true"]')) return
      if (
        target.tagName === 'IMG' ||
        target.tagName === 'SVG' ||
        target.tagName === 'A' ||
        target.closest('img, svg, a')
      ) {
        e.preventDefault()
      }
    }
    const disableMiddleClickScroll = (e: MouseEvent) => {
      if (e.button !== 1) return
      const target = e.target as HTMLElement | null
      if (target?.closest('[data-allow-middle-scroll]')) return
      e.preventDefault()
    }
    const disableBrowserShortcuts = (e: KeyboardEvent) => {
      if (e.defaultPrevented) return
      const target = e.target as HTMLElement | null
      if (target?.closest('[data-allow-browser-shortcut]')) return
      const mod = e.ctrlKey || e.metaKey
      const key = e.key.toLowerCase()
      if (
        e.key === 'F5' ||
        e.key === 'F3' ||
        (mod &&
          (key === 'p' ||
            key === 's' ||
            key === 'u' ||
            key === 'r' ||
            key === 'g' ||
            key === '=' ||
            key === '+' ||
            key === '-' ||
            key === '0'))
      ) {
        e.preventDefault()
      }
    }
    window.addEventListener('contextmenu', disableGlobalContextMenu)
    window.addEventListener('dragstart', disableDefaultBrowserDrag, true)
    window.addEventListener('mousedown', disableMiddleClickScroll, true)
    window.addEventListener('keydown', disableBrowserShortcuts)

    $hiddenPillOffsetXStore = 0
    currentLiveOffsetX = 0
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        const b = getBoundsForState()
        bounds.syncBounds(b, 0, islandNode, false, isHidden)
        bounds.syncPillBounds(islandNode, isHidden, true)
      })
    })

    const initSyncTimer1 = window.setTimeout(() => {
      bounds.syncPillBounds(islandNode, isHidden, true)
      resyncAllSatellites()
    }, 120)

    const initSyncTimer2 = window.setTimeout(() => {
      bounds.syncPillBounds(islandNode, isHidden, true)
      resyncAllSatellites()
    }, 320)

    const handleOverlayChange = (ev: unknown) => {
      const detail = unwrapEventData<
        | { open: boolean; x?: number; y?: number; w?: number; h?: number }
        | undefined
      >(ev)
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
      }, 120)
    }

    const handleResyncBounds = () => {
      bounds.syncPillBounds(islandNode, isHidden)
      resyncAllSatellites()
    }

    const refreshDisplay = async () => {
      try {
        const layout = await GetDisplayLayout()
        if (layout && layout.virtualWidth > 0) {
          displayLayout = layout
          $displayLayoutStore = layout
          const clamped = Math.max(
            allowedOffsets.min,
            Math.min(allowedOffsets.max, currentLiveOffsetX)
          )
          if (clamped !== currentLiveOffsetX) {
            currentLiveOffsetX = clamped
            $hiddenPillOffsetXStore = clamped
          }
          bounds.syncBounds(
            getBoundsForState(),
            currentLiveOffsetX,
            islandNode,
            false,
            isHidden
          )
          resyncAllSatellites()
        }
      } catch {}
    }
    refreshDisplay()

    const handleResize = () => {
      if (typeof window !== 'undefined' && window.devicePixelRatio > 0) {
        cachedWindowDpr = window.devicePixelRatio
      }
      refreshDisplay()
      bounds.syncPillBounds(islandNode, isHidden)
      resyncAllSatellites()
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
        if (gestures.isDragging) finishDragGesture()
        return
      }
      if (e?.type === 'pointercancel') {
        if (e instanceof PointerEvent && e.buttons === 1) return
        if (gestures.isDragging) finishDragGesture()
        return
      }
      if (gestures.isDragging) finishDragGesture()
    }

    const handleWindowKeydownFailsafe = (e: KeyboardEvent) => {
      if (e.defaultPrevented) return
      if (e.key === 'Escape') {
        if (!$hasCompletedOnboardingStore || activePage === OnboardingPage) {
          return
        }
        if (gestures.isDragging || gestures.isInertia) {
          finishDragGesture()
          return
        }
        if ($activeOverlayStore !== null) {
          closeOverlay()
          return
        }
        if (expanded && !isHidden) {
          const fallback = getFallbackPage()
          if (activePage !== fallback.page) {
            navigateTo(fallback.page as Component, fallback.config, true)
          } else {
            expanded = false
            const nextBounds = getBoundsForState(islandConfig, false, false)
            const targetOffset = calculateTargetOffsetForWidth(
              nextBounds.w,
              currentLiveOffsetX
            )
            if (targetOffset !== currentLiveOffsetX) {
              currentLiveOffsetX = targetOffset
              $hiddenPillOffsetXStore = targetOffset
            }
            bounds.syncBounds(
              nextBounds,
              targetOffset,
              islandNode,
              false,
              false
            )
            resetHideTimer()
          }
        }
      }
    }

    const handleWindowFocus = () => {
      recoverBoundsIfIdle()
    }

    const handleHotkeyToggle = () => {
      if (!$hasCompletedOnboardingStore || activePage === OnboardingPage) {
        return
      }
      if (expanded && !isHidden) {
        const fallback = getFallbackPage()
        if (activePage !== fallback.page) {
          activePage = fallback.page as Component
          islandConfig = fallback.config
          $activePageIdStore = fallback.page === HomePage ? 'home' : 'media'
        }
        hideWidget()
      } else {
        const fallback = getFallbackPage()
        navigateTo(fallback.page as Component, fallback.config, true)
      }
    }

    const handleQuickNotesHotkey = () => {
      if (!$hasCompletedOnboardingStore || activePage === OnboardingPage) {
        return
      }
      if (activePage === NotesPage && expanded && !isHidden) {
        const fallback = getFallbackPage()
        navigateTo(fallback.page as Component, fallback.config, false)
      } else {
        navigateTo(NotesPage, notesConfig, true)
      }
    }

    const handleModeSwitchHotkey = () => {
      if (!$hasCompletedOnboardingStore || activePage === OnboardingPage) {
        return
      }
      if (isHidden) {
        showWidget()
        return
      }
      if (expanded) {
        expanded = false
        const fallback = getFallbackPage()
        if (activePage !== fallback.page) {
          activePage = fallback.page as Component
          islandConfig = fallback.config
          $activePageIdStore = fallback.page === HomePage ? 'home' : 'media'
        }
        const nextBounds = getBoundsForState(islandConfig, false, false)
        const targetOffset = calculateTargetOffsetForWidth(
          nextBounds.w,
          currentLiveOffsetX
        )
        if (targetOffset !== currentLiveOffsetX) {
          currentLiveOffsetX = targetOffset
          $hiddenPillOffsetXStore = targetOffset
        }
        bounds.syncBounds(nextBounds, targetOffset, islandNode, false, false)
        resetHideTimer()
      } else {
        const fallback = getFallbackPage()
        navigateTo(fallback.page as Component, fallback.config, true)
      }
    }

    const unsubOverlay = Events.OnGlobal('vinta:overlay', handleOverlayChange)
    const unsubQuit = Events.OnGlobal('vinta:quit', handleQuit)
    const unsubResyncBounds = Events.OnGlobal(
      'vinta:resync-bounds',
      handleResyncBounds
    )
    window.addEventListener('resize', handleResize)
    window.addEventListener('blur', handleGlobalDragFailsafe)
    window.addEventListener('pointercancel', handleGlobalDragFailsafe)
    window.addEventListener('keydown', handleWindowKeydownFailsafe)
    window.addEventListener('focus', handleWindowFocus)

    const stopIdleReconciler = bounds.startIdleReconciler(
      () => islandNode,
      () => getBoundsForState(),
      () => isHidden
    )

    // Event-driven working-set trim via $lib/events instead of a periodic polling watchdog.
    const unsubStateTrim = Events.OnGlobal('vinta:state-change', () => {
      if (expanded && !isHidden) {
        bounds.cancelScheduledTrimMemory()
      } else {
        bounds.scheduleTrimMemory(isHidden ? 800 : 900)
      }
    })

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
      onRefreshDisplay: refreshDisplay,
      onRelocateOffset: (offset) => {
        currentLiveOffsetX = offset
        $hiddenPillOffsetXStore = offset
        const baseCenter =
          displayLayout.virtualLeft + displayLayout.primaryCenter
        const targetMon = getMonitorForOffset(
          offset,
          displayLayout.monitors,
          baseCenter,
          getWindowDpr()
        )
        if (targetMon && displayLayout.currentMonitor?.id !== targetMon.id) {
          displayLayout = {
            ...displayLayout,
            currentMonitor: targetMon
          }
          $displayLayoutStore = displayLayout
        }
        bounds.syncBounds(
          getBoundsForState(),
          offset,
          islandNode,
          false,
          isHidden
        )
        bounds.startLiveTracking(islandNode, 400)
      },
      onNotificationWake: () => {
        wasHiddenBeforeNotif = isHidden
      },
      onHotkeyToggle: handleHotkeyToggle,
      onQuickNotesHotkey: handleQuickNotesHotkey,
      onModeSwitchHotkey: handleModeSwitchHotkey
    })

    try {
      const savedDevReload = sessionStorage.getItem(
        'vinta_devtest_reload_state'
      )
      if (savedDevReload) {
        sessionStorage.removeItem('vinta_devtest_reload_state')
        const parsed = JSON.parse(savedDevReload) as {
          mode?: string
          expand?: boolean
          pinned?: boolean
        }
        if (typeof parsed.pinned === 'boolean') {
          $devPinnedStore = parsed.pinned
        }
        if (parsed.mode && PAGE_MAPPINGS[parsed.mode]) {
          const match = PAGE_MAPPINGS[parsed.mode]
          navigateTo(match.page, match.config, Boolean(parsed.expand))
        }
      }
    } catch {}

    return () => {
      clearTimeout(initSyncTimer1)
      clearTimeout(initSyncTimer2)
      stopIdleReconciler()
      unsubStateTrim()
      finishDragGesture()
      cleanupEvents()
      bounds.cleanup()
      gestures.cleanup()
      simulatedClockStore.set(null)
      clearHoverGrace()
      clearTimeout(glowTimeout)
      clearTimeout(hideTimeout)
      window.removeEventListener('contextmenu', disableGlobalContextMenu)
      window.removeEventListener('dragstart', disableDefaultBrowserDrag, true)
      window.removeEventListener('mousedown', disableMiddleClickScroll, true)
      window.removeEventListener('keydown', disableBrowserShortcuts)
      unsubOverlay()
      unsubQuit()
      unsubResyncBounds()
      window.removeEventListener('resize', handleResize)
      window.removeEventListener('blur', handleGlobalDragFailsafe)
      window.removeEventListener('pointercancel', handleGlobalDragFailsafe)
      window.removeEventListener('keydown', handleWindowKeydownFailsafe)
      window.removeEventListener('focus', handleWindowFocus)
    }
  })
</script>

<div
  bind:this={islandContainerNode}
  class="island-container"
  role="application"
  onfocusin={(e: FocusEvent) => {
    const t = e.target as HTMLElement
    if (
      t &&
      (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable)
    ) {
      FocusWindow().catch(() => {})
    }
  }}
  onpointerdown={() => {
    FocusWindow().catch(() => {})
  }}
  style="position: absolute; left: {displayLayout.primaryCenter /
    getWindowDpr()}px; top: {Math.max(
    0,
    (activeMonitor.top - displayLayout.virtualTop) / getWindowDpr()
  )}px; transform: translateX(-50%);"
>
  <div
    bind:this={islandWrapperNode}
    class="island-wrapper"
    ontransitionend={(e) => {
      if (e.target === islandWrapperNode) {
        recoverBoundsIfIdle()
      }
    }}
    ontransitioncancel={(e) => {
      if (e.target === islandWrapperNode) {
        recoverBoundsIfIdle()
      }
    }}
    style="
      pointer-events: {isQuitting ? 'none' : 'auto'};
      transform: translateX({clampedOffsetX}px);
      transition: {gestures.isDragging || gestures.isInertia
      ? 'none'
      : islandAnim};
    "
  >
    <div
      class="island-scale-box"
      style="--vinta-ui-scale: {effectiveUiScale};"
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
      {isTimerActive}
      {hideTrigger}
      {islandConfig}
      {currentMinimalW}
      {currentExpandedH}
      activeOverlay={$activeOverlayStore}
      onCloseOverlay={closeOverlay}
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
        bounds.currentOSBounds = { w: -1, h: Math.max(h, 200), r }
        bounds.captureDragBaseline(islandNode, currentLiveOffsetX)
      }}
      onStartInertia={(start, target) => {
        bounds.syncGlideCorridor(islandNode, isHidden, start, target)
      }}
      onGlideStep={(live) => {
        currentLiveOffsetX = live
        // Apply transform directly during drag/glide to avoid Svelte microtask frame jitter.
        if (islandWrapperNode) {
          islandWrapperNode.style.transform = `translateX(${live}px)`
        }
        bounds.syncDragOffset(live)
      }}
      onUpdateOffset={(next) => {
        currentLiveOffsetX = next
        if (islandWrapperNode) {
          islandWrapperNode.style.transform = `translateX(${next}px)`
        }
        bounds.syncDragOffset(next)
      }}
      onFinishDrag={finishDragGesture}
      onExpand={handleExpand}
      onShowWidget={showWidget}
      onHideWidget={hideWidget}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      onRecoverBounds={recoverBoundsIfIdle}
      onAnimationDone={recoverBoundsIfIdle}
    />

    {#if !isQuitting && !isHidden}
      <NotificationSatellite
        active={!isQuitting &&
          !isHidden &&
          activePage !== OnboardingPage &&
          $hasCompletedOnboardingStore}
        primaryCenter={displayLayout.primaryCenter / getWindowDpr()}
        {clampedOffsetX}
        monitorLeft={(activeMonitor.left - displayLayout.virtualLeft) /
          getWindowDpr()}
        activeMonitorWidth={activeMonitor.width / getWindowDpr()}
        islandWidth={currentIslandW}
        uiScale={effectiveUiScale}
      />
      <ControlsSatellite
        {activePage}
        {expanded}
        expandedHeight={currentExpandedH}
        expandedWidth={currentIslandW}
        activeMonitorWidth={activeMonitor.width / getWindowDpr()}
        monitorLeft={(activeMonitor.left - displayLayout.virtualLeft) /
          getWindowDpr()}
        primaryCenter={displayLayout.primaryCenter / getWindowDpr()}
        {clampedOffsetX}
        uiScale={effectiveUiScale}
        active={!isQuitting && !isHidden}
        onmouseenter={handleSatelliteMouseEnter}
        onmouseleave={handleSatelliteMouseLeave}
      />
    {/if}
    </div>
  </div>
</div>

<CustomTooltip />
{#if $debugClassInspector}
  {#await import('./devtest/ClassInspectorOverlay.svelte') then { default: ClassInspectorOverlay }}
    <ClassInspectorOverlay />
  {/await}
{/if}
<VirtualCursor
  primaryCenter={displayLayout.primaryCenter}
  {clampedOffsetX}
  monitorTop={Math.max(0, activeMonitor.top - displayLayout.virtualTop)}
  monitorWidth={activeMonitor.width}
/>

<style src="./App.css"></style>
