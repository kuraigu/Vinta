<script lang="ts">
  import { onMount, type Component } from 'svelte'
  import { fade } from 'svelte/transition'
  import Clock from './Clock.svelte'
  import Dock from './Dock.svelte'
  import HomePage from '../pages/HomePage.svelte'
  import MediaPage from '../pages/MediaPage.svelte'
  import TimerPage from '../pages/TimerPage.svelte'
  import {
    autoHideDisplayStore,
    allowDragPositionStore,
    autoHideDelayStore,
    type IslandConfig
  } from '../store'
  import type { IslandGestureController } from '../services/islandGestures'

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
    onFinishDrag?: () => void
    onExpand?: () => void
    onShowWidget?: () => void
    onHideWidget?: () => void
    onMouseEnter?: () => void
    onMouseLeave?: () => void
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
    hiddenBounds,
    activeOpacity = 1,
    glowOpacity = 0.6,
    islandAnim = '0.32s cubic-bezier(0.2, 0.8, 0.2, 1)',
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
    onFinishDrag,
    onExpand,
    onShowWidget,
    onHideWidget,
    onMouseEnter,
    onMouseLeave,
    onRecoverBounds,
    onAnimationDone
  }: Props = $props()

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
      onAnimationDone?.()
      onRecoverBounds?.()
    }
  }

  function handlePointerDown(e: PointerEvent) {
    if (gestures.isInertia) {
      gestures.interruptInertia()
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
      () => onFinishDrag?.()
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

  onMount(() => {
    const handleDocumentPointerLeave = (e: PointerEvent) => {
      if (!e.relatedTarget && onMouseLeave) {
        onMouseLeave()
      }
    }

    const handleWindowBlur = () => {
      if (onMouseLeave) {
        onMouseLeave()
      }
    }

    document.documentElement.addEventListener(
      'pointerleave',
      handleDocumentPointerLeave
    )
    window.addEventListener('blur', handleWindowBlur)

    return () => {
      document.documentElement.removeEventListener(
        'pointerleave',
        handleDocumentPointerLeave
      )
      window.removeEventListener('blur', handleWindowBlur)
    }
  })
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  bind:this={islandNode}
  data-vinta="main-island"
  class="island {expanded ? 'expanded' : ''} {isTransitionGlow
    ? 'is-transition-glow'
    : ''} {isHidden ? 'is-hidden' : ''} {isQuitting ? 'is-quitting' : ''}"
  onmouseenter={onMouseEnter}
  onmouseleave={onMouseLeave}
  onpointerdown={handlePointerDown}
  oncontextmenu={handleContextMenu}
  onclick={handleIslandClick}
  ontransitionend={handleTransitionEnd}
  ontransitioncancel={handleTransitionEnd}
  style="
    opacity: {isQuitting ? 0 : activeOpacity};
    filter: {isQuitting ? 'blur(12px)' : 'none'};
    width: {isHidden
    ? hiddenBounds.w
    : Math.min(
        expanded ? islandConfig.expanded.w : currentMinimalW,
        activeMonitorWidth ? activeMonitorWidth - 24 : 1920
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
>
  {#if isHidden}
    {#if $autoHideDisplayStore === 'clock'}
      <div
        class="hidden-widget-pill"
        transition:fade={{
          duration: 140,
          delay: wasExpandedBeforeHide ? 180 : 0
        }}
      >
        <Clock expanded={false} />
        <div class="hidden-dash-mini"></div>
      </div>
    {:else}
      <div
        class="hidden-dash"
        transition:fade={{
          duration: 140,
          delay: wasExpandedBeforeHide ? 180 : 0
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
      : '0.22s cubic-bezier(0.2, 0.8, 0.2, 1)'};"
  >
    {#if !isPageDormant}
      {#key [activePage, expanded]}
        {@const ActivePage = activePage}
        <div
          class="view"
          style="width: {Math.min(
            expanded ? islandConfig.expanded.w : currentMinimalW,
            activeMonitorWidth ? activeMonitorWidth - 24 : 1920
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
  {#if isTimeoutActive && (shouldCurrentPageAutoHide || (activePage !== HomePage && (activePage !== MediaPage || !isMediaActive) && (activePage !== TimerPage || !isTimerActive))) && !expanded && !isHidden}
    {#key hideTrigger}
      <div
        class="timeout-bar"
        style="animation-duration: {$autoHideDelayStore || 5}s;"
      ></div>
    {/key}
  {/if}
</div>

<style src="./styles/MainIsland.css"></style>
