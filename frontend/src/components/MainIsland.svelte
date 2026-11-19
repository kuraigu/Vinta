<!--
  @component MainIsland
  @description Core dynamic island capsule. Handles expansion and collapse animations,
  drag-to-reposition mechanics, island auto-hide timers, backdrop blur effects, and
  switching between the compact glance view and expanded full application pages.
-->
<script lang="ts">
  import { getContext, onMount, type Component } from 'svelte'
  import { fade } from 'svelte/transition'
  import { Events } from '$lib/events'
  import Clock from './Clock.svelte'
  import Dock from './Dock.svelte'
  import FileDropIndicator from './FileDropIndicator.svelte'
  import HomePage from '../pages/HomePage.svelte'
  import MediaPage from '../pages/MediaPage.svelte'
  import TimerPage from '../pages/TimerPage.svelte'
  import FileShelfPage, {
    config as fileShelfConfig
  } from '../pages/FileShelfPage.svelte'
  import NotificationsOverlayPage from '../pages/overlays/NotificationsOverlayPage.svelte'
  import AppsOverlayPage from '../pages/overlays/AppsOverlayPage.svelte'
  import FeedbackOverlayPage from '../pages/overlays/FeedbackOverlayPage.svelte'
  import {
    autoHideDisplayStore,
    type OverlayPageId,
    allowDragPositionStore,
    autoHideDelayStore,
    pageVisibilityStore,
    animScaleStore,
    ISLAND_WIDTH_TIERS,
    type IslandConfig
  } from '../store'
  import type { IslandGestureController } from '../services/islandGestures'
  import type { NavFn } from '../utils/navigation'
  import PageDeck from './PageDeck.svelte'

  /**
   * Component properties for MainIsland.
   */
  interface Props {
    activePage: Component | null
    expanded?: boolean
    isHidden?: boolean
    isQuitting?: boolean
    isTransitionGlow?: boolean
    isBlinking?: boolean
    isPageDormant?: boolean
    isDockVisible?: boolean
    isTimeoutActive?: boolean
    shouldCurrentPageAutoHide?: boolean
    isMediaActive?: boolean
    isTimerActive?: boolean
    hideTrigger?: number
    islandConfig: IslandConfig
    currentMinimalW: number
    currentExpandedH?: number
    activeOverlay?: OverlayPageId
    onCloseOverlay?: () => void
    hiddenBounds: { w: number; h: number; r: number }
    activeOpacity?: number
    glowOpacity?: number
    islandAnim?: string
    activeMonitorWidth?: number
    gestures: IslandGestureController
    virtualMinOffset?: number
    virtualMaxOffset?: number
    monitorSnapOffsets?: number[]
    currentLiveOffsetX?: number
    wasExpandedBeforeHide?: boolean
    islandNode?: HTMLDivElement
    onStartDragRunway?: () => void
    onUpdateOffset?: (offset: number) => void
    onGlideStep?: (liveOffset: number) => void
    onStartInertia?: (
      startOffset: number,
      targetOffset: number,
      velocity: number
    ) => void
    onFinishDrag?: () => void
    onExpand?: () => void
    onShowWidget?: () => void
    onHideWidget?: () => void
    onMouseEnter?: () => void
    onMouseLeave?: () => void
    onWindowBlur?: () => void
    onRecoverBounds?: () => void
    onAnimationDone?: () => void
  }

  let {
    activePage,
    expanded = false,
    isHidden = false,
    isQuitting = false,
    isTransitionGlow = false,
    isBlinking = false,
    isPageDormant = false,
    isDockVisible = true,
    isTimeoutActive = false,
    shouldCurrentPageAutoHide = false,
    isMediaActive = false,
    isTimerActive = false,
    hideTrigger = 0,
    islandConfig,
    currentMinimalW,
    currentExpandedH,
    activeOverlay = null,
    onCloseOverlay = () => {},
    hiddenBounds,
    activeOpacity = 1,
    glowOpacity = 0.6,
    islandAnim = '0.32s cubic-bezier(0.22, 1, 0.36, 1)',
    activeMonitorWidth = 1920,
    gestures,
    virtualMinOffset = -960,
    virtualMaxOffset = 960,
    monitorSnapOffsets = [0],
    currentLiveOffsetX = 0,
    wasExpandedBeforeHide = false,
    islandNode = $bindable(),
    onStartDragRunway,
    onUpdateOffset,
    onGlideStep,
    onStartInertia,
    onFinishDrag,
    onExpand,
    onShowWidget,
    onHideWidget,
    onMouseEnter,
    onMouseLeave,
    onWindowBlur,
    onRecoverBounds,
    onAnimationDone
  }: Props = $props()

  let isMorphing = $state(false)
  let morphTimer: ReturnType<typeof setTimeout> | null = null
  let prevExpandedState = $state(false)
  let prevHiddenState = $state(false)
  let prevActivePage = $state<unknown>(undefined)
  let prevMinimalW = $state<number | undefined>(undefined)
  let prevExpandedH = $state<number | undefined>(undefined)

  $effect(function trackIslandMorphState() {
    const targetMinimalW = expanded ? undefined : currentMinimalW
    const targetExpandedH = expanded
      ? (currentExpandedH ?? islandConfig.expanded.h)
      : undefined
    const initialized = prevActivePage !== undefined
    const changed =
      expanded !== prevExpandedState ||
      isHidden !== prevHiddenState ||
      (initialized &&
        (activePage !== prevActivePage ||
          targetMinimalW !== prevMinimalW ||
          targetExpandedH !== prevExpandedH))

    prevExpandedState = expanded
    prevHiddenState = isHidden
    prevActivePage = activePage
    prevMinimalW = targetMinimalW
    prevExpandedH = targetExpandedH

    if (changed) {
      const scale = $animScaleStore
      if (morphTimer) {
        clearTimeout(morphTimer)
        morphTimer = null
      }
      if (scale <= 0) {
        isMorphing = false
        return
      }
      isMorphing = true
      morphTimer = setTimeout(
        () => {
          isMorphing = false
          morphTimer = null
        },
        Math.round(650 * scale)
      )
    }
  })

  let targetViewW = $derived(
    Math.min(
      expanded
        ? activeOverlay !== null
          ? Math.max(islandConfig.expanded.w, ISLAND_WIDTH_TIERS.standard)
          : islandConfig.expanded.w
        : currentMinimalW,
      activeMonitorWidth ? activeMonitorWidth - 24 : 1920
    )
  )

  let targetViewH = $derived(
    expanded
      ? (currentExpandedH ?? islandConfig.expanded.h)
      : islandConfig.minimal.h
  )

  function handleTransitionEnd(e: TransitionEvent): void {
    if (e.target !== islandNode) return
    if (
      e.propertyName === 'width' ||
      e.propertyName === 'height' ||
      e.propertyName === 'border-radius' ||
      e.propertyName === 'border-bottom-left-radius' ||
      e.propertyName === 'border-bottom-right-radius' ||
      e.propertyName === 'transform'
    ) {
      if (morphTimer) {
        clearTimeout(morphTimer)
        morphTimer = null
      }
      isMorphing = false
      onAnimationDone?.()
      onRecoverBounds?.()
    }
  }

  function handlePointerDown(e: PointerEvent) {
    if (gestures.isInertia) {
      gestures.interruptInertia()
      onFinishDrag?.()
    }

    if (!gestures.canInitiateDrag(e, $allowDragPositionStore, expanded)) return

    gestures.startDrag(
      e,
      currentLiveOffsetX,
      () => onStartDragRunway?.(),
      () => onFinishDrag?.()
    )

    const handleGlobalPointerMove = (ev: PointerEvent) => {
      handlePointerMove(ev)
    }

    const handleGlobalPointerUp = (ev: PointerEvent) => {
      window.removeEventListener('pointermove', handleGlobalPointerMove)
      window.removeEventListener('pointerup', handleGlobalPointerUp)
      window.removeEventListener('pointercancel', handleGlobalPointerUp)
      handlePointerUp(ev)
    }

    window.addEventListener('pointermove', handleGlobalPointerMove)
    window.addEventListener('pointerup', handleGlobalPointerUp)
    window.addEventListener('pointercancel', handleGlobalPointerUp)
  }

  function handlePointerMove(e: PointerEvent) {
    gestures.handlePointerMove(
      e,
      virtualMinOffset,
      virtualMaxOffset,
      monitorSnapOffsets,
      (nextOffset) => onUpdateOffset?.(nextOffset),
      (upEv) => handlePointerUp(upEv)
    )
  }

  function handlePointerUp(e: PointerEvent) {
    if (
      !gestures.isDragging &&
      !gestures.isTrackRegionActive &&
      !gestures.isInertia
    )
      return
    gestures.handlePointerUp(
      e,
      currentLiveOffsetX,
      virtualMinOffset,
      virtualMaxOffset,
      monitorSnapOffsets,
      (nextOffset) => onUpdateOffset?.(nextOffset),
      () => onFinishDrag?.(),
      (start, target, vel) => onStartInertia?.(start, target, vel),
      (live) => onGlideStep?.(live)
    )
  }

  function handleContextMenu(e: MouseEvent) {
    if (gestures.isDragging || gestures.isInertia) {
      if (gestures.isInertia) {
        gestures.interruptInertia()
        onFinishDrag?.()
      }
      e.preventDefault()
      e.stopPropagation()
      return
    }
    if (expanded) return

    e.preventDefault()
    e.stopPropagation()
    if (isHidden) {
      onShowWidget?.()
    } else {
      onHideWidget?.()
    }
  }

  function handleIslandClick(e: MouseEvent) {
    if (gestures.dragMovedDistance > 4) {
      e.preventDefault()
      e.stopPropagation()
      return
    }
    if (isHidden) {
      onShowWidget?.()
      return
    }
    if (!expanded) {
      onExpand?.()
    }
  }

  function handleDragHandleDblClick(e: MouseEvent) {
    e.stopPropagation()
    onRecoverBounds?.()
  }

  const navigate = getContext<NavFn | undefined>('navigate')

  const FILE_HOLD_DURATION_MS = 950
  let isHoldingFilesOverIsland = $state(false)
  let fileHoldProgress = $state(0)
  let fileHoldInterval: ReturnType<typeof setInterval> | null = null
  let fileHoldStartTime = 0
  let lastFileDragOverTime = 0

  function stopFileHoldTimer(): void {
    if (fileHoldInterval) {
      clearInterval(fileHoldInterval)
      fileHoldInterval = null
    }
    isHoldingFilesOverIsland = false
    fileHoldProgress = 0
    fileHoldStartTime = 0
    lastFileDragOverTime = 0
  }

  function startOrRefreshFileHoldTimer(): void {
    if ($pageVisibilityStore.fileshelf === false) return
    onMouseEnter?.()
    if (isHidden) {
      onShowWidget?.()
    }
    if (expanded && activePage === FileShelfPage) {
      stopFileHoldTimer()
      return
    }

    lastFileDragOverTime = Date.now()
    if (fileHoldInterval) return

    fileHoldStartTime = Date.now()
    isHoldingFilesOverIsland = true
    fileHoldProgress = 5

    fileHoldInterval = setInterval(() => {
      const now = Date.now()
      if (now - lastFileDragOverTime > 240) {
        stopFileHoldTimer()
        return
      }
      const elapsed = now - fileHoldStartTime
      const pct = Math.min(100, (elapsed / FILE_HOLD_DURATION_MS) * 100)
      fileHoldProgress = pct
      if (elapsed >= FILE_HOLD_DURATION_MS) {
        stopFileHoldTimer()
        navigate?.(FileShelfPage, fileShelfConfig, true)
      }
    }, 40)
  }

  function handleIslandDragEnter(e: DragEvent): void {
    if (!e.dataTransfer?.types?.includes('Files')) return
    startOrRefreshFileHoldTimer()
  }

  function handleIslandDragOver(e: DragEvent): void {
    if (!e.dataTransfer?.types?.includes('Files')) return
    startOrRefreshFileHoldTimer()
  }

  function handleIslandDragLeave(e: DragEvent): void {
    if (!e.dataTransfer?.types?.includes('Files')) return
    const nextTarget = e.relatedTarget as Node | null
    if (!nextTarget || !islandNode?.contains(nextTarget)) {
      stopFileHoldTimer()
      onMouseLeave?.()
    }
  }

  function handleIslandDrop(): void {
    stopFileHoldTimer()
  }

  onMount(() => {
    const handleDocumentPointerLeave = (e: PointerEvent) => {
      if (!e.relatedTarget && onMouseLeave) {
        onMouseLeave()
      }
    }

    const handleWindowBlur = () => {
      const wasHoldingFiles = isHoldingFilesOverIsland
      stopFileHoldTimer()
      if (onMouseLeave) {
        onMouseLeave()
      }
      if (!wasHoldingFiles) {
        onWindowBlur?.()
      }
    }

    const handleDevFileHold = (ev: any) => {
      const detail = (ev?.data ?? ev) as
        { active?: boolean; progress?: number } | undefined
      isHoldingFilesOverIsland = !!detail?.active
      fileHoldProgress = detail?.progress ?? 0
    }

    document.documentElement.addEventListener(
      'pointerleave',
      handleDocumentPointerLeave
    )
    window.addEventListener('blur', handleWindowBlur)
    const unsubFileHold = Events.OnGlobal(
      'devtest:file-hold',
      handleDevFileHold
    )

    return () => {
      stopFileHoldTimer()
      document.documentElement.removeEventListener(
        'pointerleave',
        handleDocumentPointerLeave
      )
      window.removeEventListener('blur', handleWindowBlur)
      unsubFileHold()
    }
  })
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  bind:this={islandNode}
  data-vinta="main-island"
  data-file-drop-target
  class="island {expanded ? 'expanded' : ''} {isTransitionGlow
    ? 'is-transition-glow'
    : ''} {isMorphing ? 'is-transitioning' : ''} {isHidden
    ? 'is-hidden'
    : ''} {isQuitting ? 'is-quitting' : ''}"
  onmouseenter={onMouseEnter}
  onmouseleave={onMouseLeave}
  onpointerdown={handlePointerDown}
  oncontextmenu={handleContextMenu}
  onclick={handleIslandClick}
  ondragenter={handleIslandDragEnter}
  ondragover={handleIslandDragOver}
  ondragleave={handleIslandDragLeave}
  ondrop={handleIslandDrop}
  ontransitionend={handleTransitionEnd}
  ontransitioncancel={handleTransitionEnd}
  style="
    opacity: {isQuitting ? 0 : activeOpacity};
    filter: {isQuitting ? 'blur(12px)' : 'none'};
    width: {isHidden ? hiddenBounds.w : targetViewW}px;
    height: {isHidden ? hiddenBounds.h : targetViewH}px;
    --island-target-w: {targetViewW}px;
    --island-target-h: {targetViewH}px;
    --island-r: {isHidden
    ? hiddenBounds.r
    : expanded
      ? islandConfig.expanded.r
      : islandConfig.minimal.r}px;
    border-radius: 0 0 var(--island-r) var(--island-r);
    --island-border-color: {isHidden
    ? 'var(--vinta-border)'
    : islandConfig.borderColor || 'var(--vinta-island-border-idle)'};
    --island-glow-color: {isHidden
    ? 'transparent'
    : islandConfig.glowColor && islandConfig.glowColor !== 'transparent'
      ? islandConfig.glowColor
      : 'var(--vinta-island-glow-idle)'};
    --island-hidden-glow: {islandConfig.glowColor &&
  islandConfig.glowColor !== 'transparent'
    ? islandConfig.glowColor
    : 'rgba(255, 255, 255, 0.6)'};
    --island-glow-opacity: {glowOpacity};
    --island-anim: {islandAnim};
  "
>
  <FileDropIndicator
    active={isHoldingFilesOverIsland}
    progress={fileHoldProgress}
    isCompact={!expanded}
  />

  {#if isHidden}
    {#if $autoHideDisplayStore === 'clock'}
      <div
        class="hidden-widget-pill"
        transition:fade={{
          duration: Math.round(140 * $animScaleStore),
          delay: wasExpandedBeforeHide ? Math.round(180 * $animScaleStore) : 0
        }}
      >
        <Clock expanded={false} />
        <div class="hidden-dash-mini"></div>
      </div>
    {:else}
      <div
        class="hidden-dash"
        transition:fade={{
          duration: Math.round(140 * $animScaleStore),
          delay: wasExpandedBeforeHide ? Math.round(180 * $animScaleStore) : 0
        }}
      ></div>
    {/if}
  {/if}

  {#if !isHidden && expanded && $allowDragPositionStore}
    <div
      class="island-drag-handle"
      aria-label="Drag to reposition (or Alt + Drag)"
      ondblclick={handleDragHandleDblClick}
    >
      <div class="drag-pill-bar"></div>
    </div>
  {/if}

  <div
    class="content"
    style="opacity: {isHidden || isBlinking ? 0 : 1}; pointer-events: {isHidden
      ? 'none'
      : 'auto'}; transition: opacity {isBlinking
      ? '0s'
      : 'calc(0.42s * var(--vinta-anim-scale, 1)) cubic-bezier(0.18, 0.9, 0.32, 1)'};"
  >
    {#if !isPageDormant}
      <div
        class="view"
        style="width: {targetViewW}px; height: {targetViewH}px;"
      >
        <PageDeck activePage={activePage ?? undefined} {expanded} />
      </div>
    {/if}

    {#if expanded && !isPageDormant && activeOverlay === 'notifications'}
      <NotificationsOverlayPage onClose={onCloseOverlay} />
    {:else if expanded && !isPageDormant && activeOverlay === 'apps'}
      <AppsOverlayPage onClose={onCloseOverlay} />
    {:else if expanded && !isPageDormant && activeOverlay === 'feedback'}
      <FeedbackOverlayPage onClose={onCloseOverlay} />
    {/if}

    {#if !isPageDormant}
      <div
        class="dock-overlay-wrapper"
        class:is-visible={expanded && isDockVisible}
        aria-hidden={!(expanded && isDockVisible)}
      >
        <Dock {activePage} {islandConfig} />
      </div>
    {/if}
  </div>

  <!-- Shrinking Timeout Bar -->
  {#if isTimeoutActive && (shouldCurrentPageAutoHide || (activePage !== HomePage && (activePage !== MediaPage || !isMediaActive) && (activePage !== TimerPage || !isTimerActive))) && !expanded && !isHidden}
    {#key hideTrigger}
      <div
        class="timeout-bar"
        style="animation-duration: {$autoHideDelayStore || 5}s;"
      ></div>
    {/key}
  {/if}

  {#if isTransitionGlow}
    <div class="transition-glow-ring" aria-hidden="true"></div>
  {/if}
</div>

<style src="./styles/MainIsland.css"></style>
