<script lang="ts">
  import { getContext } from 'svelte'
  import {
    EyeOff,
    Minimize2,
    Maximize2,
    Power,
    ChevronDown,
    ChevronUp
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
    onDrawerChange?: (isOpen: boolean) => void
  }

  let {
    activePage,
    expanded = false,
    expandedTooltipSide = 'right',
    onDrawerChange
  }: Props = $props()

  let isDrawerOpen = $state(false)
  let controlsEl = $state<HTMLDivElement | null>(null)

  function handleOpenDrawer(e?: MouseEvent) {
    if (e) e.stopPropagation()
    isDrawerOpen = true
    onDrawerChange?.(true)
  }

  function handleCloseDrawer(e?: MouseEvent) {
    if (e) e.stopPropagation()
    isDrawerOpen = false
    onDrawerChange?.(false)
  }

  /**
   * Automatically retracts drawer when the parent island expands.
   */
  function handleExpandedDrawerClose() {
    if (expanded && isDrawerOpen) {
      isDrawerOpen = false
      onDrawerChange?.(false)
    }
  }

  $effect(handleExpandedDrawerClose)
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
  bind:this={controlsEl}
  class="floating-controls"
  class:minimized={!expanded}
  class:folded={!expanded && !isDrawerOpen}
  role="toolbar"
  aria-label="Window controls"
  tabindex="-1"
  onclick={!expanded && !isDrawerOpen ? handleOpenDrawer : undefined}
  onfocusin={() => {
    if (!expanded && !isDrawerOpen) {
      handleOpenDrawer()
    }
  }}
  onfocusout={(e) => {
    if (!expanded && isDrawerOpen) {
      const nextTarget = e.relatedTarget as Node | null
      if (!controlsEl || !controlsEl.contains(nextTarget)) {
        handleCloseDrawer()
      }
    }
  }}
>
  {#if expanded}
    <!-- Expanded Card State: Vertical Dock attached to parent window -->
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
    <!-- Minimized Island: Top-edge Pull-down Tab -->
    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        tooltip="Window Controls"
        tooltipPosition="bottom"
        onclick={handleOpenDrawer}
      >
        <ChevronDown size={13.5} strokeWidth={2.2} />
      </DockButton>
    </div>
  {:else}
    <!-- Minimized Island: Dropped-down Vertical Drawer (Order: Retract, Expand, Hide, Quit) -->
    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        tooltip=""
        tooltipPosition="bottom"
        onclick={handleCloseDrawer}
      >
        <ChevronUp size={13.5} strokeWidth={2.2} />
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
