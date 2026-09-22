<script lang="ts">
  import { onDestroy, type Snippet } from 'svelte'
  import { fly, scale } from 'svelte/transition'
  import { cubicOut } from 'svelte/easing'
  import {
    registerSatellite,
    unregisterSatellite,
    type SatelliteAnchor,
    type TooltipSide
  } from '../stores/satellites'

  /**
   * Configuration properties for a declarative floating satellite host.
   */
  type Props = {
    /** Unique identifier for the satellite. */
    id: string
    /** Anchor flank relative to the main Dynamic Island. */
    anchor?: SatelliteAnchor
    /** Visual width of the satellite container in pixels. */
    width: number
    /** Visual height of the satellite container in pixels. */
    height: number
    /** Corner radius for Win32 hit-test region generation. */
    radius?: number
    /** Spacing in pixels between island edge and satellite. */
    gap?: number
    /** Additional horizontal coordinate offset for stacking. */
    offsetX?: number
    /** Additional vertical coordinate offset for stacking. */
    offsetY?: number
    /** Permanent tooltip region padding in pixels. */
    tooltipPad?: number
    /** Direction the tooltip extends from the satellite. */
    tooltipSide?: TooltipSide
    /** Whether the satellite is currently active and clickable. */
    active: boolean
    /** Slotted child content to render inside the satellite host. */
    children?: Snippet
    /** Optional mouse enter callback for peripheral hover tracking. */
    onmouseenter?: (e: MouseEvent) => void
    /** Optional mouse leave callback for peripheral hover tracking. */
    onmouseleave?: (e: MouseEvent) => void
  }

  let {
    id,
    anchor = 'right',
    width,
    height,
    radius = 20,
    gap = 12,
    offsetX = 0,
    offsetY = 0,
    tooltipPad = 0,
    tooltipSide,
    active,
    children,
    onmouseenter,
    onmouseleave
  }: Props = $props()

  /**
   * Registers or updates satellite geometry and synchronizes bounds with the native backend.
   */
  function syncSatelliteBoundsEffect() {
    registerSatellite({
      id,
      anchor,
      width,
      height,
      radius,
      gap,
      offsetX,
      offsetY,
      tooltipPad,
      tooltipSide,
      active
    })
  }

  $effect(syncSatelliteBoundsEffect)

  onDestroy(() => {
    unregisterSatellite(id)
  })

  /**
   * Computes directional offset vectors for entry and exit transitions
   * based on the active anchor flank.
   */
  function computeFlyParams() {
    switch (anchor) {
      case 'left':
        return {
          inX: 16,
          outX: 12,
          inY: 0,
          outY: 0
        }
      case 'bottom':
        return {
          inX: 0,
          outX: 0,
          inY: -16,
          outY: -12
        }
      case 'top':
        return {
          inX: 0,
          outX: 0,
          inY: 16,
          outY: 12
        }
      case 'right':
      default:
        return {
          inX: -16,
          outX: -12,
          inY: 0,
          outY: 0
        }
    }
  }

  let flyParams = $derived.by(computeFlyParams)
</script>

{#if active}
  <div
    class="satellite-host"
    data-anchor={anchor}
    role="none"
    style="
      --sat-gap: {gap}px;
      --sat-offset-x: {offsetX}px;
      --sat-offset-y: {offsetY}px;
    "
    {onmouseenter}
    {onmouseleave}
    in:fly={{
      x: flyParams.inX,
      y: flyParams.inY,
      duration: 200,
      delay: 40,
      easing: cubicOut
    }}
    out:scale={{
      duration: 180,
      start: 0.6,
      easing: cubicOut
    }}
  >
    {#if children}
      {@render children()}
    {/if}
  </div>
{/if}

<style>
  .satellite-host {
    position: absolute;
    z-index: 110;
    pointer-events: auto;
    cursor: default;
    transition:
      top 0.26s cubic-bezier(0.2, 0.8, 0.2, 1),
      left 0.26s cubic-bezier(0.2, 0.8, 0.2, 1),
      right 0.26s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  /* Bridge across the gap to ensure continuous pointer capture during hover transition */
  .satellite-host::before {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    pointer-events: auto;
  }

  .satellite-host[data-anchor='right']::before {
    right: 100%;
    width: calc(var(--sat-gap) + 4px);
  }

  .satellite-host[data-anchor='left']::before {
    left: 100%;
    width: calc(var(--sat-gap) + 4px);
  }

  .satellite-host[data-anchor='bottom']::before {
    bottom: 100%;
    left: 0;
    right: 0;
    height: calc(var(--sat-gap) + 4px);
  }

  .satellite-host[data-anchor='top']::before {
    top: 100%;
    left: 0;
    right: 0;
    height: calc(var(--sat-gap) + 4px);
  }

  .satellite-host[data-anchor='right'] {
    left: calc(100% + var(--sat-gap) + var(--sat-offset-x));
    top: var(--sat-offset-y);
  }

  .satellite-host[data-anchor='left'] {
    right: calc(100% + var(--sat-gap) + var(--sat-offset-x));
    top: var(--sat-offset-y);
  }

  .satellite-host[data-anchor='bottom'] {
    top: calc(100% + var(--sat-gap) + var(--sat-offset-y));
    left: calc(50% + var(--sat-offset-x));
    transform: translateX(-50%);
  }

  .satellite-host[data-anchor='top'] {
    bottom: calc(100% + var(--sat-gap) + var(--sat-offset-y));
    left: calc(50% + var(--sat-offset-x));
    transform: translateX(-50%);
  }
</style>
