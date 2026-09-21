<script lang="ts">
  import { getContext, onDestroy } from 'svelte'
  import {
    EyeOff,
    Minimize2,
    Maximize2,
    Power,
    ChevronRight,
    ChevronLeft
  } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from './DockButton.svelte'
  import type { IslandConfig } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const hidePill = getContext<() => void>('hidePill')

  type Props = {
    activePage: unknown
    expanded?: boolean
    expandedTooltipSide?: 'right' | 'left'
    horizontalTooltipAlign?: 'left' | 'right'
    onDrawerChange?: (isOpen: boolean) => void
  }

  let {
    activePage,
    expanded = false,
    expandedTooltipSide = 'right',
    horizontalTooltipAlign = 'left',
    onDrawerChange
  }: Props = $props()

  let containerNode = $state<HTMLElement>()
  let isDrawerOpen = $state(false)
  let isTransitioning = $state(false)
  let transitionTimeout: number | undefined

  function triggerTransitionGuard(duration = 260) {
    clearTimeout(transitionTimeout)
    isTransitioning = true
    transitionTimeout = window.setTimeout(() => {
      isTransitioning = false
    }, duration)
  }

  function handleOpenDrawer(e: MouseEvent) {
    e.stopPropagation()
    isDrawerOpen = true
    triggerTransitionGuard(260)
    onDrawerChange?.(true)
  }

  function handleCloseDrawer(e: MouseEvent) {
    e.stopPropagation()
    isDrawerOpen = false
    triggerTransitionGuard(260)
    onDrawerChange?.(false)
  }

  /**
   * Resets the drawer state and activates transition clipping when expanded changes.
   */
  function handleExpandedStateSync() {
    const isExp = expanded
    triggerTransitionGuard(320)
    if (isExp && isDrawerOpen) {
      isDrawerOpen = false
      onDrawerChange?.(false)
    }
  }

  $effect(handleExpandedStateSync)

  onDestroy(() => {
    clearTimeout(transitionTimeout)
  })
</script>

<div
  class="floating-controls"
  class:horizontal={!expanded}
  class:folded={!expanded && !isDrawerOpen}
  class:transitioning={isTransitioning}
  class:align-right={horizontalTooltipAlign === 'right'}
  role="toolbar"
  aria-label="Window controls"
  tabindex="-1"
  bind:this={containerNode}
>
  {#if expanded}
    <!-- Expanded Card State: Vertical Dock (Order: Collapse, Hide, Quit) -->
    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        tooltip=""
        tooltipPosition={expandedTooltipSide}
        onclick={(e: MouseEvent) =>
          AppControls.doMinimize(navigate, e, activePage)}
      >
        <Minimize2 size={13.5} strokeWidth={2.2} />
      </DockButton>
    </div>

    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        tooltip=""
        tooltipPosition={expandedTooltipSide}
        onclick={(e: MouseEvent) => AppControls.doHide(navigate, hidePill, e)}
      >
        <EyeOff size={13.5} strokeWidth={2.2} />
      </DockButton>
    </div>

    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        variant="danger"
        tooltip=""
        tooltipPosition={expandedTooltipSide}
        onclick={(e: MouseEvent) => AppControls.doQuit(e)}
      >
        <Power size={13.5} strokeWidth={2.2} />
      </DockButton>
    </div>
  {:else if !isDrawerOpen}
    <!-- Minimized Island State: Folded Compact Toggle -->
    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        tooltip=""
        tooltipPosition="bottom"
        onclick={handleOpenDrawer}
      >
        <ChevronRight size={13.5} strokeWidth={2.2} />
      </DockButton>
    </div>
  {:else}
    <!-- Minimized Island State: Unfolded Horizontal Drawer (Order: Fold, Expand, Hide, Quit) -->
    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        tooltip=""
        tooltipPosition="bottom"
        onclick={handleCloseDrawer}
      >
        <ChevronLeft size={13.5} strokeWidth={2.2} />
      </DockButton>
    </div>

    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        tooltip=""
        tooltipPosition="bottom"
        onclick={(e: MouseEvent) =>
          AppControls.doExpand(navigate, e, activePage)}
      >
        <Maximize2 size={13.5} strokeWidth={2.2} />
      </DockButton>
    </div>

    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        tooltip=""
        tooltipPosition="bottom"
        onclick={(e: MouseEvent) => AppControls.doHide(navigate, hidePill, e)}
      >
        <EyeOff size={13.5} strokeWidth={2.2} />
      </DockButton>
    </div>

    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        variant="danger"
        tooltip=""
        tooltipPosition="bottom"
        onclick={(e: MouseEvent) => AppControls.doQuit(e)}
      >
        <Power size={13.5} strokeWidth={2.2} />
      </DockButton>
    </div>
  {/if}
</div>

<style src="./styles/FloatingControls.css"></style>
