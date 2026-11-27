<!--
  @component FloatingControls
  @description Compact satellite toolbar for quick media skip, volume mute, and
  island expand triggers.
-->
<script lang="ts">
  import { getContext } from 'svelte'
  import { EyeOff, Minimize2, Pin, Power } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from './DockButton.svelte'
  import {
    wakeMethodStore,
    globalHotkeyStore,
    clickOutsideMinimizeStore,
    overlayPinnedStore,
    toggleCurrentOverlayPin,
    tStore,
    type IslandConfig
  } from '../store'

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

  let wakeHint = $derived.by(() => {
    const hotkey = $globalHotkeyStore || 'Alt+Shift+V'
    return $wakeMethodStore === 'hover'
      ? $tStore('controls.wake_hover_hint', { hotkey })
      : $tStore('controls.wake_hotkey_hint', { hotkey })
  })
</script>

<div
  class="floating-controls"
  role="toolbar"
  aria-label={$tStore('controls.window_controls')}
  tabindex="-1"
>
  {#if $clickOutsideMinimizeStore}
    <div class="button-wrapper">
      <DockButton
        shape="rounded"
        active={$overlayPinnedStore}
        tooltip={$overlayPinnedStore
          ? $tStore('controls.unpin')
          : $tStore('controls.pin')}
        tooltipPosition={expandedTooltipSide}
        onclick={() => {
          toggleCurrentOverlayPin()
        }}
      >
        <Pin size={14} strokeWidth={2.2} />
      </DockButton>
    </div>
  {/if}

  <div class="button-wrapper">
    <DockButton
      shape="rounded"
      tooltip={$tStore('controls.minimize')}
      tooltipPosition={expandedTooltipSide}
      onclick={(e: MouseEvent) =>
        AppControls.doMinimize(navigate, e, activePage)}
    >
      <Minimize2 size={14} strokeWidth={2.2} />
    </DockButton>
  </div>

  <div class="button-wrapper">
    <DockButton
      shape="rounded"
      tooltip={$tStore('controls.hide')}
      tooltipSubtitle={wakeHint}
      tooltipPosition={expandedTooltipSide}
      onclick={(e: MouseEvent) => AppControls.doHide(navigate, hidePill, e)}
    >
      <EyeOff size={14} strokeWidth={2.2} />
    </DockButton>
  </div>

  <div class="controls-divider" aria-hidden="true"></div>

  <div class="button-wrapper">
    <DockButton
      shape="rounded"
      variant="danger"
      tooltip={$tStore('controls.quit')}
      tooltipPosition={expandedTooltipSide}
      onclick={(e: MouseEvent) => AppControls.doQuit(e)}
    >
      <Power size={14} strokeWidth={2.2} />
    </DockButton>
  </div>
</div>

<style src="./styles/FloatingControls.css"></style>
