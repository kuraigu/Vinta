<!--
  @component ControlsSatellite
  @description Floating control capsule anchored alongside the island. Provides quick
  sliders for display brightness and system volume, plus media playback buttons without
  needing to open the full media or action center pages.
-->
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
    onmouseleave
  }: Props = $props()

  const EXPANDED_RADII: FourPartRadii = [12, 12, 12, 12]
  const EXPANDED_GAPS: FourPartGaps = [0, 6, 0, 6]

  const getInitialExpanded = () => expanded
  let layoutExpanded = $state(getInitialExpanded())
  let isTransitioning = $state(false)
  let transitionTimer: number | undefined
  let settleTimer: number | undefined

  /**
   * Synchronizes satellite layout across island expansion and collapse cycles.
   * Delays mounting during island expansion (290ms) until the island finishes
   * animating, avoiding any flash of the controls during expansion, and unmounts
   * immediately on collapse.
   */
  function syncLayoutExpanded() {
    if (layoutExpanded === expanded) {
      return
    }

    clearTimeout(transitionTimer)
    clearTimeout(settleTimer)

    if (!expanded) {
      layoutExpanded = false
      isTransitioning = false
      return
    }

    isTransitioning = true
    const islandDuration = 290

    transitionTimer = window.setTimeout(() => {
      layoutExpanded = true
      settleTimer = window.setTimeout(() => {
        isTransitioning = false
        if (typeof window !== 'undefined') {
          window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
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
    const centerOffsetY = Math.max(0, Math.round((expandedHeight - 88) / 2))
    return {
      width: 32,
      height: 88,
      radius: 12,
      radii: EXPANDED_RADII,
      gap: 6,
      gaps: EXPANDED_GAPS,
      offsetY: centerOffsetY,
      tooltipPad: 0,
      tooltipSide: 'top' as const
    }
  }

  let satelliteBounds = $derived.by(computeSatelliteBounds)
  const isControlsActive = $derived(active && expanded && layoutExpanded)
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
  active={isControlsActive && !isTransitioning}
  animateIn={true}
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
    <FloatingControls {activePage} {expandedTooltipSide} />
  </div>
</SatelliteHost>
