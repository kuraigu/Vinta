<!--
  @component DockButton
  @description Individual launcher icon in the dock. Renders app icons or Lucide SVGs,
  shows tooltips on hover, and displays badge counts for unread notifications.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import { islandTooltip } from '../services/islandTooltip'

  type Props = {
    onclick: (e: MouseEvent) => void
    oncontextmenu?: (e: MouseEvent) => void
    tooltip?: string
    tooltipSubtitle?: string
    tooltipPosition?: 'top' | 'left' | 'right' | 'bottom'
    variant?: '' | 'danger' | 'warning' | 'success'
    shape?: 'rounded' | 'circle'
    active?: boolean
    disabled?: boolean
    badge?: number
    useCustomTooltip?: boolean
    children: Snippet
  }

  let {
    onclick,
    oncontextmenu = undefined,
    tooltip,
    tooltipSubtitle = undefined,
    tooltipPosition = 'top',
    variant = '',
    shape = 'rounded',
    active = false,
    disabled = false,
    badge = undefined,
    useCustomTooltip = true,
    children
  }: Props = $props()

  /**
   * Maps DockButton tooltip positions onto CustomTooltip position and alignment options.
   */
  function resolveTooltipOptions() {
    if (!useCustomTooltip || !tooltip) {
      return { text: '' }
    }
    return {
      text: tooltip,
      subtitle: tooltipSubtitle,
      position: tooltipPosition
    }
  }

  let tooltipOpts = $derived.by(resolveTooltipOptions)
</script>

<button
  type="button"
  class="dock-btn {variant} shape-{shape}"
  class:active
  {disabled}
  onclick={(e) => {
    e.stopPropagation()
    if (!disabled && onclick) onclick(e)
  }}
  {oncontextmenu}
  use:islandTooltip={disabled ? { text: '' } : tooltipOpts}
  title={!useCustomTooltip && !disabled ? tooltip : undefined}
  aria-label={disabled ? undefined : tooltip}
>
  {@render children()}
  {#if badge !== undefined && badge > 0}
    <span class="badge">{badge > 99 ? '99+' : badge}</span>
  {/if}
</button>

<style src="./styles/DockButton.css"></style>
