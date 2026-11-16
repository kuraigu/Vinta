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
  class="dock-btn {variant} shape-{shape}"
  class:active
  {disabled}
  onclick={(e) => {
    e.stopPropagation()
    if (!disabled && onclick) onclick(e)
  }}
  use:islandTooltip={disabled ? { text: '' } : tooltipOpts}
  title={!useCustomTooltip && !disabled ? tooltip : undefined}
  aria-label={disabled ? undefined : tooltip}
>
  {@render children()}
  {#if badge !== undefined && badge > 0}
    <span class="badge">{badge > 99 ? '99+' : badge}</span>
  {/if}
</button>

<style>
  .dock-btn {
    background: transparent;
    border: 1px solid transparent;
    color: rgba(255, 255, 255, 0.85);
    width: 32px;
    height: 32px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition:
      transform 0.18s cubic-bezier(0.16, 1, 0.3, 1),
      background-color 0.15s ease,
      box-shadow 0.18s ease,
      color 0.15s ease;
    position: relative;
  }

  .dock-btn:focus-visible {
    outline: 2px solid var(--vinta-accent-primary, #38bdf8);
    outline-offset: 1px;
    background: rgba(255, 255, 255, 0.15);
  }

  .dock-btn.shape-circle {
    border-radius: 50%;
  }

  .dock-btn:disabled {
    opacity: 0.28;
    cursor: not-allowed;
    pointer-events: none;
    transform: none !important;
    filter: none !important;
  }

  .dock-btn:hover {
    color: #ffffff;
    background: rgba(255, 255, 255, 0.12);
    border-color: transparent;
    transform: scale(1.08);
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.35);
  }

  .dock-btn.active {
    color: #ffffff;
    background: rgba(255, 255, 255, 0.1);
    border: none;
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
    filter: none;
  }

  .dock-btn.active::before {
    content: '';
    position: absolute;
    bottom: -3px;
    left: 50%;
    transform: translateX(-50%);
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background: var(--vinta-accent-primary, #38bdf8);
    box-shadow: 0 0 6px rgba(56, 189, 248, 0.8);
    pointer-events: none;
  }

  .dock-btn.active:hover {
    background: rgba(255, 255, 255, 0.16);
    border-color: transparent;
    box-shadow: 0 3px 10px rgba(0, 0, 0, 0.3);
  }

  /* Action button hover colors */
  .dock-btn.danger:hover {
    color: #ffffff;
    background: color-mix(
      in srgb,
      var(--vinta-accent-danger, #ff3b30) 35%,
      transparent
    );
    border-color: color-mix(
      in srgb,
      var(--vinta-accent-danger, #ff3b30) 60%,
      transparent
    );
    box-shadow: 0 2px 8px
      color-mix(in srgb, var(--vinta-accent-danger, #ff3b30) 30%, transparent);
  }

  .dock-btn.warning:hover {
    color: #ffffff;
    background: color-mix(
      in srgb,
      var(--vinta-accent-warning, #ffcc00) 35%,
      transparent
    );
    border-color: color-mix(
      in srgb,
      var(--vinta-accent-warning, #ffcc00) 60%,
      transparent
    );
    box-shadow: 0 2px 8px
      color-mix(in srgb, var(--vinta-accent-warning, #ffcc00) 30%, transparent);
  }

  .dock-btn.success:hover {
    color: #ffffff;
    background: color-mix(
      in srgb,
      var(--vinta-accent-success, #34c759) 35%,
      transparent
    );
    border-color: color-mix(
      in srgb,
      var(--vinta-accent-success, #34c759) 60%,
      transparent
    );
    box-shadow: 0 2px 8px
      color-mix(in srgb, var(--vinta-accent-success, #34c759) 30%, transparent);
  }

  /* Notification badge */
  .badge {
    position: absolute;
    top: 1px;
    right: 1px;
    min-width: 14px;
    height: 14px;
    padding: 0 3px;
    border-radius: 7px;
    background: var(--vinta-accent-primary, #38bdf8);
    color: var(--vinta-text-on-accent, #ffffff);
    font-size: 9px;
    font-weight: 700;
    line-height: 12px;
    text-align: center;
    pointer-events: none;
    box-sizing: border-box;
    border: 1.5px solid var(--vinta-border, rgba(255, 255, 255, 0.25));
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.4);
  }
</style>
