<script lang="ts">
  import { getContext, onMount } from 'svelte'
  import { EyeOff, Minimize2, Power } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from './DockButton.svelte'
  import type { IslandConfig } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const hidePill = getContext<() => void>('hidePill')

  let { activePage }: { activePage: unknown } = $props()

  let containerNode = $state<HTMLElement>()
  let tooltipPosition = $state<'left' | 'right'>('right')

  /**
   * Adjusts tooltip popover direction depending on proximity to the right monitor boundary.
   */
  function updateTooltipPosition() {
    if (!containerNode) return
    const rect = containerNode.getBoundingClientRect()
    // Flip tooltip leftwards if near screen right boundary to prevent clipping
    tooltipPosition = window.innerWidth - rect.right < 130 ? 'left' : 'right'
  }

  onMount(() => {
    updateTooltipPosition()
    window.addEventListener('resize', updateTooltipPosition)
    return () => {
      window.removeEventListener('resize', updateTooltipPosition)
    }
  })
</script>

<div
  class="floating-controls"
  role="toolbar"
  aria-label="Window controls"
  tabindex="-1"
  bind:this={containerNode}
  onmouseenter={updateTooltipPosition}
>
  <DockButton
    shape="circle"
    variant="danger"
    tooltip="Quit Vinta"
    {tooltipPosition}
    onclick={(e: MouseEvent) => AppControls.doQuit(e)}
  >
    <Power size={18} />
  </DockButton>

  <DockButton
    shape="circle"
    variant="success"
    tooltip="Hide Widget"
    {tooltipPosition}
    onclick={(e: MouseEvent) => AppControls.doHide(navigate, hidePill, e)}
  >
    <EyeOff size={18} />
  </DockButton>

  <DockButton
    shape="circle"
    variant="warning"
    tooltip="Collapse Island"
    {tooltipPosition}
    onclick={(e: MouseEvent) => AppControls.doMinimize(navigate, e, activePage)}
  >
    <Minimize2 size={18} />
  </DockButton>
</div>

<style src="./styles/FloatingControls.css"></style>
