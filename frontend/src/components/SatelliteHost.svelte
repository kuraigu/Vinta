<!--
  @component SatelliteHost
  @description Hit-test wrapper for floating satellite widgets. Observes the bounding
  rectangle of child elements with a ResizeObserver and sends hit rects to the Go backend
  so Windows treats the floating pills as clickable without making the empty space between them intercept clicks.
-->
<script lang="ts">
  import { onDestroy, type Snippet } from 'svelte'
  import { fly, scale } from 'svelte/transition'
  import { cubicOut } from 'svelte/easing'
  import {
    registerSatellite,
    unregisterSatellite,
    type SatelliteAnchor,
    type TooltipSide,
    type FourPartRadii,
    type FourPartGaps
  } from '../stores/satellites'
  import { SetForceInteractive } from '../../bindings/Vinta/backend/app.js'

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
    /** Corner radius or 4-corner radii [tl, tr, br, bl] for hit-testing and styling. */
    radius?: number | FourPartRadii
    /** Spacing in pixels or directional gaps [top, right, bottom, left]. */
    gap?: number | FourPartGaps
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
    /** Whether to play the entry animation when mounting into view. */
    animateIn?: boolean
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
    animateIn = true,
    children,
    onmouseenter,
    onmouseleave
  }: Props = $props()

  let hostEl: HTMLElement | null = $state(null)

  /**
   * Synchronizes satellite bounds with the native Win32 window region.
   * For anchored satellites, coordinates are computed relative to the island
   * to keep the OS hit region stable and prevent subpixel animation jitter.
   */
  function syncSatelliteBounds() {
    if (!active) {
      unregisterSatellite(id)
      return
    }

    let x = 0
    let y = 0
    const w = width
    const h = height

    // For anchored satellites, coordinates are computed dynamically relative to the island in the backend
    // to track page size changes and transitions seamlessly. Only unanchored satellites pass direct coordinates.
    if (!anchor && hostEl) {
      const rect = hostEl.getBoundingClientRect()
      if (rect.width > 0 && rect.height > 0) {
        x = Math.round(rect.left)
        y = Math.round(rect.top)
      }
    }

    registerSatellite({
      id,
      anchor,
      x,
      y,
      width: w,
      height: h,
      radius,
      gap,
      offsetX,
      offsetY,
      tooltipPad,
      tooltipSide,
      active
    })

    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
    }
  }

  let lastActiveState: boolean | undefined

  /**
   * Synchronizes bounds whenever geometry properties or offset change.
   */
  function syncSatelliteBoundsEffect() {
    if (active) {
      void {
        width,
        height,
        gap,
        offsetX,
        offsetY,
        radius
      }
      syncSatelliteBounds()
      if (lastActiveState !== active) {
        lastActiveState = active
      }
    } else {
      if (lastActiveState !== false) {
        lastActiveState = false
        syncSatelliteBounds()
      }
    }
  }

  $effect(syncSatelliteBoundsEffect)

  let resizeObserver: ResizeObserver | undefined
  let transitionCleanup: (() => void) | undefined

  /**
   * Immediately synchronizes bounds when the DOM element mounts into the document.
   */
  function handleHostMount() {
    if (hostEl && active) {
      const node = hostEl
      syncSatelliteBounds()
      requestAnimationFrame(() => {
        syncSatelliteBounds()
        if (typeof window !== 'undefined') {
          window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
        }
      })

      transitionCleanup?.()
      const handleTransitionEnd = (e: TransitionEvent) => {
        if (e.target !== node) return
        syncSatelliteBounds()
        if (typeof window !== 'undefined') {
          window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
        }
      }
      node.addEventListener('transitionend', handleTransitionEnd)
      transitionCleanup = () => {
        node.removeEventListener('transitionend', handleTransitionEnd)
      }

      if (typeof ResizeObserver !== 'undefined') {
        resizeObserver?.disconnect()
        resizeObserver = new ResizeObserver(() => {
          syncSatelliteBounds()
          if (typeof window !== 'undefined') {
            window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
          }
        })
        resizeObserver.observe(node)
      }
    }
  }

  $effect(handleHostMount)

  onDestroy(() => {
    resizeObserver?.disconnect()
    transitionCleanup?.()
    transitionCleanup = undefined
    unregisterSatellite(id)
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
    }
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
  /**
   * Resolves 4-part directional spacing gaps.
   */
  function computeGaps() {
    if (Array.isArray(gap)) {
      return {
        top: gap[0],
        right: gap[1],
        bottom: gap[2],
        left: gap[3]
      }
    }
    return {
      top: gap,
      right: gap,
      bottom: gap,
      left: gap
    }
  }
  let gaps = $derived.by(computeGaps)

  /**
   * Resolves 4-corner radii [tl, tr, br, bl].
   */
  function computeRadii() {
    if (Array.isArray(radius)) {
      return {
        tl: radius[0],
        tr: radius[1],
        br: radius[2],
        bl: radius[3]
      }
    }
    return {
      tl: radius,
      tr: radius,
      br: radius,
      bl: radius
    }
  }
  let radii = $derived.by(computeRadii)
  /**
   * Directional entry transition that cleanly bypasses animations when animateIn is false.
   */
  function satelliteIn(
    node: Element,
    params: { enabled: boolean; x: number; y: number }
  ) {
    if (!params.enabled) {
      return { duration: 0 }
    }
    return fly(node, {
      x: params.x,
      y: params.y,
      duration: 200,
      delay: 40,
      easing: cubicOut
    })
  }

  function handleMouseEnter(e: MouseEvent) {
    if (typeof SetForceInteractive === 'function') {
      SetForceInteractive(true).catch(() => {})
    }
    onmouseenter?.(e)
  }

  function handleMouseLeave(e: MouseEvent) {
    if (typeof SetForceInteractive === 'function') {
      SetForceInteractive(false).catch(() => {})
    }
    onmouseleave?.(e)
  }

  onDestroy(() => {
    if (typeof SetForceInteractive === 'function') {
      SetForceInteractive(false).catch(() => {})
    }
  })

  $effect(() => {
    if (!active && typeof SetForceInteractive === 'function') {
      SetForceInteractive(false).catch(() => {})
    }
  })
</script>

{#if active}
  <div
    bind:this={hostEl}
    class="satellite-host"
    data-anchor={anchor}
    role="none"
    ontransitionrun={syncSatelliteBounds}
    ontransitionend={syncSatelliteBounds}
    style="
      width: {width}px;
      height: {height}px;
      --sat-gap: {gaps.left}px;
      --sat-gap-top: {gaps.top}px;
      --sat-gap-right: {gaps.right}px;
      --sat-gap-bottom: {gaps.bottom}px;
      --sat-gap-left: {gaps.left}px;
      --sat-radius-tl: {radii.tl}px;
      --sat-radius-tr: {radii.tr}px;
      --sat-radius-br: {radii.br}px;
      --sat-radius-bl: {radii.bl}px;
      --sat-offset-x: {offsetX}px;
      --sat-offset-y: {offsetY}px;
    "
    onmouseenter={handleMouseEnter}
    onmouseleave={handleMouseLeave}
    in:satelliteIn={{
      enabled: animateIn,
      x: flyParams.inX,
      y: flyParams.inY
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
    border-radius: var(--sat-radius-tl) var(--sat-radius-tr)
      var(--sat-radius-br) var(--sat-radius-bl);
    transition:
      height 0.22s cubic-bezier(0.16, 1, 0.3, 1),
      top 0.32s cubic-bezier(0.2, 0.8, 0.2, 1),
      left 0.32s cubic-bezier(0.2, 0.8, 0.2, 1),
      right 0.32s cubic-bezier(0.2, 0.8, 0.2, 1);
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
    width: calc(var(--sat-gap-left) + 4px);
  }

  .satellite-host[data-anchor='left']::before {
    left: 100%;
    width: calc(var(--sat-gap-right) + 4px);
  }

  .satellite-host[data-anchor='bottom']::before {
    bottom: 100%;
    left: 0;
    right: 0;
    height: calc(var(--sat-gap-top) + 4px);
  }

  .satellite-host[data-anchor='top']::before {
    top: 100%;
    left: 0;
    right: 0;
    height: calc(var(--sat-gap-bottom) + 4px);
  }

  .satellite-host[data-anchor='right'] {
    left: calc(100% + var(--sat-gap-left) + var(--sat-offset-x));
    top: calc(var(--sat-gap-top) + var(--sat-offset-y));
  }

  .satellite-host[data-anchor='left'] {
    right: calc(100% + var(--sat-gap-right) + var(--sat-offset-x));
    top: calc(var(--sat-gap-top) + var(--sat-offset-y));
  }

  .satellite-host[data-anchor='bottom'] {
    top: calc(100% + var(--sat-gap-top) + var(--sat-offset-y));
    left: calc(50% + var(--sat-offset-x));
    transform: translateX(-50%);
  }

  .satellite-host[data-anchor='top'] {
    bottom: calc(100% + var(--sat-gap-bottom) + var(--sat-offset-y));
    left: calc(50% + var(--sat-offset-x));
    transform: translateX(-50%);
  }
</style>
