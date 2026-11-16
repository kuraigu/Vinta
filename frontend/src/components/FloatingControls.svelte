<!--
  @component FloatingControls
  @description Compact satellite toolbar for quick media skip, volume mute, and
  island expand triggers.
-->
<script lang="ts">
  import { getContext } from 'svelte'
  import { EyeOff, Minimize2, Power } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from './DockButton.svelte'
  import {
    hasCompletedOnboardingStore,
    wakeMethodStore,
    globalHotkeyStore,
    tStore,
    type IslandConfig
  } from '../store'
  import OnboardingPage from '../pages/OnboardingPage.svelte'

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

  let isOnboarding = $derived(
    !$hasCompletedOnboardingStore || activePage === OnboardingPage
  )

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
  <div class="button-wrapper">
    <DockButton
      shape="rounded"
      tooltip={isOnboarding ? '' : $tStore('controls.minimize')}
      tooltipPosition={expandedTooltipSide}
      disabled={isOnboarding}
      onclick={(e: MouseEvent) =>
        AppControls.doMinimize(navigate, e, activePage)}
    >
      <Minimize2 size={13.5} strokeWidth={2.2} />
    </DockButton>
  </div>

  <div class="button-wrapper">
    <DockButton
      shape="rounded"
      tooltip={isOnboarding ? '' : $tStore('controls.hide')}
      tooltipSubtitle={isOnboarding ? undefined : wakeHint}
      tooltipPosition={expandedTooltipSide}
      disabled={isOnboarding}
      onclick={(e: MouseEvent) => AppControls.doHide(navigate, hidePill, e)}
    >
      <EyeOff size={13.5} strokeWidth={2.2} />
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
      <Power size={13.5} strokeWidth={2.2} />
    </DockButton>
  </div>
</div>

<style src="./styles/FloatingControls.css"></style>
