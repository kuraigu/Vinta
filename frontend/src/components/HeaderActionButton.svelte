<!--
  @component HeaderActionButton
  @description Unified header action button primitive across Vinta pages and overlays.
  Supports three tiers: `ghost` (32x32 circular icon button matching NotificationsOverlayPage),
  `pill` (32px subtle labeled pill), and `hero` (32px accent primary CTA).
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import {
    islandTooltip,
    type IslandTooltipOptions
  } from '../services/islandTooltip'

  interface Props {
    variant?: 'ghost' | 'pill' | 'hero'
    state?: 'default' | 'active' | 'danger'
    disabled?: boolean
    ariaLabel?: string
    ariaPressed?: boolean
    tooltip?: string | IslandTooltipOptions
    class?: string
    onclick?: (e: MouseEvent) => void
    children?: Snippet
  }

  let {
    variant = 'ghost',
    state = 'default',
    disabled = false,
    ariaLabel,
    ariaPressed,
    tooltip,
    class: className = '',
    onclick,
    children
  }: Props = $props()
</script>

<button
  type="button"
  class="vinta-header-btn variant-{variant} state-{state} {className}"
  {disabled}
  aria-label={ariaLabel}
  aria-pressed={ariaPressed}
  use:islandTooltip={tooltip}
  {onclick}
>
  {@render children?.()}
</button>

<style src="./styles/HeaderActionButton.css"></style>
