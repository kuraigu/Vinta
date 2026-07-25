<script lang="ts">
  import { onMount, setContext, type Component } from 'svelte'
  import { fade } from 'svelte/transition'
  import {
    SetWidgetBounds,
    IsTestMode,
    FocusWindow,
    Quit
  } from '../wailsjs/go/backend/App.js'
  import { EventsOn, LogFatal, LogInfo } from '../wailsjs/runtime/runtime.js'
  import NotificationPage, {
    config as notifConfig
  } from './pages/NotificationPage.svelte'
  import HomePage, { config as homeConfig } from './pages/HomePage.svelte'
  import Dock from './components/Dock.svelte'
  import Clock from './components/Clock.svelte'
  import DevPill from './devtest/DevPill.svelte'
  import AppControls from './utils/AppControls.svelte'
  import { playNotifChime } from './services/audioService'
  import {
    themeStore,
    shapeStore,
    autoHideStore,
    autoHideDelayStore,
    autoHideDisplayStore,
    notifAutoWakeStore,
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
    compactScaleStore,
    type IslandConfig
  } from './store'

  $effect(() => {
    document.documentElement.setAttribute('data-theme', $themeStore)
  })

  // Dynamic Routing: The active app that the island is displaying
  let activePage = $state<Component>(HomePage)
  let islandConfig = $state<IslandConfig>(homeConfig)

  let expanded = $state(false)
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

  let islandNode = $state<HTMLElement>()
  let isTestMode = $state(false)

  let isTransitionGlow = $state(true)
  let glowTimeout: number

  let islandAnim = $state('0.4s cubic-bezier(0.22, 1, 0.36, 1)')

  function triggerGlow() {
    isTransitionGlow = true
    clearTimeout(glowTimeout)
    glowTimeout = window.setTimeout(() => {
      isTransitionGlow = false
    }, 3000)
  }

  $effect(() => {
    // Re-run this effect whenever expansion or page changes
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const _p = activePage

    triggerGlow()
  })

  let isTimeoutActive = $state(false)

  $effect(() => {
    // Auto-hide and auto-revert logic effect
    void hideTrigger
    clearTimeout(hideTimeout)
    isTimeoutActive = false

    if (!expanded && !isHidden && !isHovering) {
      // Set a timer if we need to auto-hide OR we need to revert back to HomePage
      if ($autoHideStore || activePage !== HomePage) {
        isTimeoutActive = true
        const delayMs = ($autoHideDelayStore || 10) * 1000
        hideTimeout = window.setTimeout(() => {
          isTimeoutActive = false
          if ($autoHideStore) {
            hideWidget()
          }
          if (activePage !== HomePage) {
            navigateTo(HomePage, homeConfig, false)
          }
        }, delayMs)
      }
    }
  })

  // Manage OS bounds when hiding/unhiding
  function hideWidget() {
    if (isHidden) return
    islandAnim = '0.5s cubic-bezier(0.16, 1, 0.3, 1)'
    isHidden = true
  }

  function showWidget() {
    if (!isHidden) return
    islandAnim = '0.5s cubic-bezier(0.16, 1, 0.3, 1)'
    isHidden = false
    triggerGlow()
  }

  // Expose manual hide for the dock
  setContext('hidePill', () => {
    hideWidget()
    if (expanded) {
      navigateTo(HomePage, homeConfig, false)
    }
  })

  // Expose manual wake up for components
  setContext('wakeupPill', showWidget)

  // Automatically wake up when new notifications arrive and play audio chime
  let prevNotifCount = $state(0)
  let hideTrigger = $state(0)
  notifData.subscribe((notifs) => {
    if (notifs.length > prevNotifCount) {
      if ($notifSoundStore && !$dndStore) {
        playNotifChime($notifVolumeStore)
      }
      showWidget()
      hideTrigger++
    }
    prevNotifCount = notifs.length
  })

  function updateBounds(w: number, h: number, r: number) {
    const computedR = $shapeStore === 'box' ? 12 : r
    const scale = window.devicePixelRatio || 1
    SetWidgetBounds(w, h, computedR, scale, $hiddenPillOffsetXStore)
  }

  // --- Centralized OS Bounds Manager ---
  let currentOSBounds = { w: 0, h: 0, r: 0 }
  let osBoundsTimeout: number

  function applyBounds(w: number, h: number, r: number) {
    if (w >= currentOSBounds.w && h >= currentOSBounds.h) {
      // Growing in all dimensions (or same). Safe to update OS region immediately.
      clearTimeout(osBoundsTimeout)
      updateBounds(w, h, r)
      currentOSBounds = { w, h, r }
    } else {
      // Shrinking in at least one dimension.
      // Create a temporary bounding box so the CSS animation isn't clipped
      const maxW = Math.max(currentOSBounds.w, w)
      const maxH = Math.max(currentOSBounds.h, h)
      const maxR = Math.max(currentOSBounds.r, r)
      updateBounds(maxW, maxH, maxR)

      clearTimeout(osBoundsTimeout)
      osBoundsTimeout = window.setTimeout(() => {
        updateBounds(w, h, r)
        currentOSBounds = { w, h, r }
      }, 450) // Wait slightly longer than 400ms CSS animation to be safe
    }
  }

  let hiddenBounds = $derived(
    $autoHideDisplayStore === 'clock'
      ? { w: 105, h: 24, r: 12 }
      : { w: 60, h: 8, r: 4 }
  )

  let compactW = $derived(
    $compactScaleStore === 'small'
      ? 320
      : $compactScaleStore === 'large'
        ? 400
        : islandConfig.minimal.w
  )

  let glowOpacity = $derived(
    $glowIntensityStore === 'disabled'
      ? 0
      : $glowIntensityStore === 'vibrant'
        ? 1.0
        : 0.6
  )

  let extraDropdownH = $state(0)
  let isQuitting = $state(false)

  onMount(() => {
    const handleDropdownToggle = (e: Event) => {
      const open = (e as CustomEvent).detail
      extraDropdownH = open ? 220 : 0
    }
    const handleQuit = () => {
      isQuitting = true
      if ($notifSoundStore) {
        playNotifChime($notifVolumeStore * 0.4)
      }
      setTimeout(() => {
        Quit()
      }, 350)
    }
    window.addEventListener('dev-dropdown-toggle', handleDropdownToggle)
    window.addEventListener('vinta-quit', handleQuit)
    return () => {
      window.removeEventListener('dev-dropdown-toggle', handleDropdownToggle)
      window.removeEventListener('vinta-quit', handleQuit)
    }
  })

  $effect(() => {
    if (!islandNode) return

    // Track $hiddenPillOffsetXStore so Win32 click-through region moves in sync during horizontal drag
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const _ox = $hiddenPillOffsetXStore

    const baseW = isHidden
      ? hiddenBounds.w
      : expanded
        ? islandConfig.expanded.w
        : compactW

    const baseH = isHidden
      ? hiddenBounds.h
      : expanded
        ? islandConfig.expanded.h
        : islandConfig.minimal.h

    const targetW = baseW
    const targetH = baseH + extraDropdownH

    // Always use box radius if shapeStore is box
    const r = isHidden
      ? hiddenBounds.r
      : $shapeStore === 'box'
        ? 12
        : expanded
          ? islandConfig.expanded.r
          : islandConfig.minimal.r

    applyBounds(targetW, targetH, r)
  })

  // Universal horizontal dragging across all island states
  let isDraggingHidden = $state(false)
  let hiddenDragStartX = 0
  let hiddenDragStartOffset = 0

  function handleHiddenPointerDown(e: PointerEvent) {
    if (!$allowDragPositionStore) return
    // Disable repositioning drag when expanded so page gestures and useSwipe work 100% cleanly
    if (expanded) return

    const target = e.target as HTMLElement
    if (
      target.closest(
        'button, input, select, textarea, a, .custom-select, .no-drag, [role="button"]'
      )
    ) {
      return
    }
    isDraggingHidden = true
    hiddenDragStartX = e.clientX
    hiddenDragStartOffset = $hiddenPillOffsetXStore
    try {
      ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
    } catch {
      // Ignore
    }
  }

  function handleHiddenPointerMove(e: PointerEvent) {
    if (!isDraggingHidden) return
    const dx = e.clientX - hiddenDragStartX
    $hiddenPillOffsetXStore = Math.max(
      -600,
      Math.min(600, hiddenDragStartOffset + dx)
    )
  }

  function handleHiddenPointerUp(e: PointerEvent) {
    if (!isDraggingHidden) return
    isDraggingHidden = false
    try {
      ;(e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId)
    } catch {
      // Ignore
    }
  }

  function navigateTo(
    page: Component,
    pageConfig: IslandConfig,
    expand: boolean = false
  ) {
    if (expanded !== expand) {
      islandAnim = '0.4s cubic-bezier(0.22, 1, 0.36, 1)'
    }
    activePage = page
    islandConfig = pageConfig
    expanded = expand
  }

  // Global navigate for dev pill / pages
  setContext('navigate', navigateTo)

  import { preloadIcons, findAppIconConflicts } from './appIcons'

  onMount(async () => {
    // Preload SVG caches in the background
    preloadIcons()

    // Check if launched via `scripts.bat test`
    isTestMode = await IsTestMode()
    if (isTestMode) {
      const conflicts = findAppIconConflicts()
      if (conflicts.length > 0) {
        LogFatal('APP ICON CONFLICTS FOUND:\n' + conflicts.join('\n'))
        // LogFatal quits the application automatically
      } else {
        LogInfo('APP ICON TEST PASSED: No conflicts found.')
        // In test mode we may optionally quit after all tests pass, but we'll leave it running for visual devtests unless told otherwise.
      }
    }

    currentOSBounds = {
      w: islandConfig.minimal.w,
      h: islandConfig.minimal.h,
      r: islandConfig.minimal.r
    }
    updateBounds(currentOSBounds.w, currentOSBounds.h, currentOSBounds.r)

    // Listen to systray events
    EventsOn('toggle-theme', () => {
      themeStore.update((t) => (t === 'dark' ? 'light' : 'dark'))
    })
    EventsOn('toggle-shape', () => {
      shapeStore.update((s) => (s === 'pill' ? 'box' : 'pill'))
    })
    EventsOn('show-pill', showWidget)
  })

  function handleExpand() {
    if (expanded) return
    FocusWindow()
    islandAnim = '0.4s cubic-bezier(0.22, 1, 0.36, 1)'
    expanded = true
  }

  function handleIslandClick() {
    FocusWindow()
    islandAnim = '0.4s cubic-bezier(0.22, 1, 0.36, 1)'

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
    FocusWindow()
    if (isHidden && $wakeMethodStore === 'hover') {
      hoverShowTimeout = window.setTimeout(showWidget, 150)
    } else if (!isHidden) {
      showWidget()
    }
  }

  function handleMouseLeave() {
    isHovering = false
    clearTimeout(hoverShowTimeout)
    if ($autoHideStore && !expanded && !isHidden) {
      const delayMs = ($autoHideDelayStore || 10) * 1000
      hideTimeout = window.setTimeout(hideWidget, delayMs)
    }
  }

  onMount(() => {
    EventsOn(
      'notification',
      (payload: {
        appName: string
        icon: string
        title: string
        message: string
        heroImage: string
      }) => {
        notifData.update((n) => {
          const newNotif = {
            id: Date.now(),
            appName: payload.appName,
            icon: payload.icon,
            title: payload.title,
            message: payload.message,
            heroImage: payload.heroImage,
            timestamp: Date.now()
          }

          n = [newNotif, ...n].slice(0, 99)

          if (!$dndStore && $notifAutoWakeStore && !expanded) {
            showWidget()
            navigateTo(NotificationPage, notifConfig, false)
            triggerGlow()
          }

          return n
        })
      }
    )

    EventsOn('notification_permission_denied', () => {
      notificationPermissionDenied.set(true)
      // Force navigate to Notification page to show the banner
      showWidget()
      navigateTo(NotificationPage, notifConfig, true)
    })
  })
</script>

<!-- PUT UTILS HERE -->
<AppControls />

<!-- ISLAND DIV HERE -->
<div class="island-container" role="application">
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
    ondblclick={(e) => {
      const target = e.target as HTMLElement
      if (
        !target.closest(
          'button, input, select, textarea, a, .custom-select, [role="button"]'
        )
      ) {
        $hiddenPillOffsetXStore = 0
      }
    }}
    onwheel={() => {
      FocusWindow()
    }}
    style="
      transform: translateX({$hiddenPillOffsetXStore}px);
      opacity: {activeOpacity};
      width: {isHidden
      ? hiddenBounds.w
      : expanded
        ? islandConfig.expanded.w
        : compactW}px;
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
  >
    {#if isHidden}
      {#if $autoHideDisplayStore === 'clock'}
        <div class="hidden-widget-pill" in:fade={{ duration: 200 }}>
          <Clock expanded={false} />
          <div class="hidden-dash-mini"></div>
        </div>
      {:else}
        <div class="hidden-dash" in:fade={{ duration: 200 }}></div>
      {/if}
    {/if}

    <div
      class="content"
      style="opacity: {isHidden ? 0 : 1}; display: {isHidden
        ? 'none'
        : 'grid'}; transition: opacity 0.2s;"
    >
      {#key activePage}
        {#if !expanded}
          {@const ActivePage = activePage}
          <div
            class="view"
            in:fade={{ duration: 150, delay: 250 }}
            out:fade={{ duration: 100 }}
            style="width: {islandConfig.minimal.w}px; height: {islandConfig
              .minimal.h}px;"
          >
            <ActivePage expanded={false} />
          </div>
        {:else}
          {@const ActivePage = activePage}
          <div
            class="view"
            in:fade={{ duration: 200, delay: 250 }}
            out:fade={{ duration: 100 }}
            style="width: {islandConfig.expanded.w}px; height: {islandConfig
              .expanded.h}px;"
          >
            <ActivePage expanded={true} />
            <div in:fade={{ duration: 200, delay: 350 }}>
              <Dock {activePage} />
            </div>
          </div>
        {/if}
      {/key}
    </div>

    <!-- Shrinking Timeout Bar -->
    {#if isTimeoutActive && activePage !== HomePage && !expanded && !isHidden}
      {#key hideTrigger}
        <div
          class="timeout-bar"
          style="animation-duration: {$autoHideDelayStore || 10}s;"
        ></div>
      {/key}
    {/if}
  </div>

  {#if isTestMode}
    <DevPill />
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
    gap: 15px; /* Matches the gap in backend/app.go */
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

    margin-top: 5px; /* Matches visualTop in backend/app.go */
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
    pointer-events: auto; /* Re-enable clicks on the island itself */

    transform: translateZ(0);
    -webkit-transform: translateZ(0);

    will-change: auto; /* Removed width/height to prevent texture reallocation lag */

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
    background-color: var(--c-bg); /* Maintain the native theme base */
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

    /* 
     * NOTE: We must inline the mask instead of using a CSS variable. 
     * Browsers have a bug where `@property` animations fail to update 
     * if nested inside an unregistered CSS variable applied to a mask.
     */
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
  }

  .island.is-transition-glow::before {
    animation:
      orbit 3s linear forwards,
      fade-glow 3s ease-in-out forwards;
  }

  .island:not(.expanded) {
    cursor: pointer;
  }

  .island:not(.expanded):hover::before {
    opacity: 0 !important; /* Hide glow when the solid hover ring activates */
    animation: none !important;
  }

  .island:not(.expanded):hover {
    box-shadow:
      var(--island-hover-border),
      var(--island-hover-glow),
      /* Neon glow effect */ var(--island-hover-drop-shadow);
  }

  .content {
    display: grid;
    place-items: center;
    width: 100%;
    height: 100%;
    border-radius: inherit;
    overflow: hidden;
    position: relative;
  }

  .view {
    grid-area: 1 / 1;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
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
