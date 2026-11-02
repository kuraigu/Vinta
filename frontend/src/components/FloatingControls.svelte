<script lang="ts">
  import { getContext } from 'svelte'
  import { EyeOff, Minimize2, Power } from 'lucide-svelte'
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
    expandedTooltipSide?: 'right' | 'left'
  }

  let { activePage, expandedTooltipSide = 'right' }: Props = $props()
</script>

<div
  class="floating-controls"
  role="toolbar"
  aria-label="Window controls"
  tabindex="-1"
>
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
</div>

<style src="./styles/FloatingControls.css"></style>
