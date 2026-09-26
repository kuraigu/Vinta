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
  let exitSyncTimeout: number | undefined

  /**
   * Synchronizes layout expansion across transition cycles.
   * When active, layout is synced immediately. When inactive during island
   * transitions, layout update is deferred until the exit animation finishes (180ms),
   * ensuring that the target layout is fully ready before the satellite re-enters.
   */
  function syncLayoutExpanded() {
    clearTimeout(exitSyncTimeout)
    if (active) {
      layoutExpanded = expanded
      return
    }

    exitSyncTimeout = window.setTimeout(() => {
      layoutExpanded = expanded
    }, 180)
  }

  $effect(syncLayoutExpanded)

  onDestroy(() => {
    clearTimeout(exitSyncTimeout)
  })

  /**
   * Handles drawer state toggles and notifies external listeners.
   */
  function handleDrawerChange(isOpen: boolean) {
    isDrawerOpen = isOpen
    onDrawerChange?.(isOpen)
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
    const dockRightEdge = islandCenterOnScreen + halfExpandedW + dockGap + dockWidth
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
    if (layoutExpanded) {
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
  {clampedOffsetX}
  {active}
  animateIn={!layoutExpanded}
  {onmouseenter}
  {onmouseleave}
>
  <FloatingControls
    {activePage}
    expanded={layoutExpanded}
    {expandedTooltipSide}
    onDrawerChange={handleDrawerChange}
  />
</SatelliteHost>
