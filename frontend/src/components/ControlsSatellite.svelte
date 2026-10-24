<script lang="ts">
  import { onDestroy } from 'svelte'
  import SatelliteHost from './SatelliteHost.svelte'
  import FloatingControls from './FloatingControls.svelte'
  import type { FourPartRadii, FourPartGaps } from '../stores/satellites'

  /**
   * Component properties for ControlsSatellite.
   */
  interface Props {
    activePage: unknown
    expanded?: boolean
    expandedHeight?: number
    expandedWidth?: number
    activeMonitorWidth?: number
    monitorLeft?: number
    primaryCenter?: number
    clampedOffsetX?: number
    active?: boolean
    onmouseenter?: (e: MouseEvent) => void
    onmouseleave?: (e: MouseEvent) => void
    onDrawerChange?: (isOpen: boolean) => void
  }

  let {
    activePage,
    expanded = false,
    expandedHeight = 400,
    expandedWidth = 380,
    activeMonitorWidth = 1920,
    monitorLeft = 0,
    primaryCenter = 960,
    clampedOffsetX = 0,
    active = true,
    onmouseenter,
    onmouseleave,
    onDrawerChange
  }: Props = $props()

  let isDrawerOpen = $state(false)
  const getInitialExpanded = () => expanded
  let layoutExpanded = $state(getInitialExpanded())
  let isTransitioning = $state(false)
  let transitionTimer: number | undefined
  let settleTimer: number | undefined

  /**
   * Synchronizes satellite layout across island expansion and collapse cycles.
   * Smoothly fades out during the island size transition (110ms) to prevent
   * visual collisions and mismatched Win32 clipping regions between minimized
   * and expanded states, then re-enters at the resting position once the island
   * finishes animating.
   */
  function syncLayoutExpanded() {
    if (layoutExpanded === expanded) {
      return
    }

    clearTimeout(transitionTimer)
    clearTimeout(settleTimer)

    // Fade out immediately to avoid mid-flight clipping or collision artifacts
    isTransitioning = true

    const islandDuration = expanded ? 290 : 210

    transitionTimer = window.setTimeout(() => {
      layoutExpanded = expanded
      if (expanded) {
        isDrawerOpen = false
        onDrawerChange?.(false)
      }

      // Settle and fade in cleanly at the exact resting position
      settleTimer = window.setTimeout(() => {
        isTransitioning = false
        if (typeof window !== 'undefined') {
          window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
          window.setTimeout(() => {
            window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
          }, 250)
        }
      }, 30)
    }, islandDuration)
  }

  $effect(syncLayoutExpanded)

  onDestroy(() => {
    clearTimeout(transitionTimer)
    clearTimeout(settleTimer)
  })

  /**
   * Handles drawer state toggles and notifies external listeners.
   */
  function handleDrawerChange(isOpen: boolean) {
    isDrawerOpen = isOpen
    onDrawerChange?.(isOpen)
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
      requestAnimationFrame(() => {
        window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
      })
    }
  }

  /**
   * Automatically retracts the drawer if the main island enters expanded mode.
   */
  function handleAutoRetractDrawer() {
    if (layoutExpanded && isDrawerOpen) {
      isDrawerOpen = false
      onDrawerChange?.(false)
    }
  }

  $effect(handleAutoRetractDrawer)

  /**
   * Calculates which flank has more room to display the expanded tooltip.
   */
  function computeExpandedTooltipSide(): 'right' | 'left' {
    const islandCenterOnScreen = primaryCenter + clampedOffsetX - monitorLeft
    const halfExpandedW = expandedWidth / 2
    const dockGap = 6
    const dockWidth = 32
    const dockRightEdge =
      islandCenterOnScreen + halfExpandedW + dockGap + dockWidth
    const dockLeftEdge = dockRightEdge - dockWidth

    const rightClearance = activeMonitorWidth - dockRightEdge
    const leftClearance = dockLeftEdge
    return rightClearance >= leftClearance ? 'right' : 'left'
  }

  let expandedTooltipSide = $derived.by(computeExpandedTooltipSide)

  /**
   * Computes layout dimensions, 4-part gaps, and corner radii for the controls satellite.
   */
  function computeSatelliteBounds() {
    if (expanded && layoutExpanded) {
      const centerOffsetY = Math.max(0, Math.round((expandedHeight - 88) / 2))
      const expandedRadii: FourPartRadii = [12, 12, 12, 12]
      const expandedGaps: FourPartGaps = [0, 6, 0, 6]
      return {
        width: 32,
        height: 88,
        radius: 12,
        radii: expandedRadii,
        gap: 6,
        gaps: expandedGaps,
        offsetY: centerOffsetY,
        tooltipPad: 0,
        tooltipSide: 'top' as const
      }
    }

    const drawerRadii: FourPartRadii = [0, 0, 12, 12]
    const drawerGaps: FourPartGaps = [0, 6, 0, 6]

    if (isDrawerOpen) {
      return {
        width: 32,
        height: 116,
        radius: 0,
        radii: drawerRadii,
        gap: 6,
        gaps: drawerGaps,
        offsetY: 0,
        tooltipPad: 0,
        tooltipSide: 'top' as const
      }
    }

    return {
      width: 32,
      height: 24,
      radius: 0,
      radii: drawerRadii,
      gap: 6,
      gaps: drawerGaps,
      offsetY: 0,
      tooltipPad: 0,
      tooltipSide: 'top' as const
    }
  }

  let satelliteBounds = $derived.by(computeSatelliteBounds)
</script>

<SatelliteHost
  id="controls"
  anchor="right"
  width={satelliteBounds.width}
  height={satelliteBounds.height}
  radius={satelliteBounds.radii}
  gap={satelliteBounds.gaps}
  offsetY={satelliteBounds.offsetY}
  tooltipPad={satelliteBounds.tooltipPad}
  tooltipSide={satelliteBounds.tooltipSide}
  active={active && !isTransitioning}
  animateIn={!layoutExpanded}
  {onmouseenter}
  {onmouseleave}
>
  <div
    class="controls-transition-wrapper"
    style="
      width: 100%;
      height: 100%;
      pointer-events: {isTransitioning ? 'none' : 'auto'};
      opacity: {isTransitioning ? 0 : 1};
      transform: scale({isTransitioning ? 0.88 : 1});
      transition: opacity 0.1s ease, transform 0.1s ease;
    "
  >
    <FloatingControls
      {activePage}
      expanded={layoutExpanded}
      {expandedTooltipSide}
      onDrawerChange={handleDrawerChange}
    />
  </div>
</SatelliteHost>
