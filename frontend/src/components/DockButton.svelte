<script lang="ts">
  import type { Snippet } from 'svelte'

  type Props = {
    onclick: (e: MouseEvent) => void
    tooltip?: string
    tooltipPosition?: 'top' | 'left' | 'right' | 'bottom'
    variant?: '' | 'danger' | 'warning' | 'success'
    shape?: 'rounded' | 'circle'
    active?: boolean
    badge?: number
    useCustomTooltip?: boolean
    children: Snippet
  }

  let {
    onclick,
    tooltip,
    tooltipPosition = 'top',
    variant = '',
    shape = 'rounded',
    active = false,
    badge = undefined,
    useCustomTooltip = true,
    children
  }: Props = $props()
</script>

<button
  class="dock-btn {variant} shape-{shape} tooltip-{tooltipPosition}"
  class:active
  onclick={(e) => {
    e.stopPropagation()
    if (onclick) onclick(e)
  }}
  data-tooltip={useCustomTooltip ? tooltip : ''}
  title={!useCustomTooltip ? tooltip : undefined}
  aria-label={tooltip}
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
    color: var(--c-text-secondary);
    width: 32px;
    height: 32px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.22, 1, 0.36, 1);
    position: relative;
  }

  .dock-btn.shape-circle {
    border-radius: 50%;
  }

  .dock-btn[data-tooltip]:not([data-tooltip=''])::after {
    content: attr(data-tooltip);
    position: absolute;
    bottom: calc(100% + 12px);
    left: 50%;
    transform: translateX(-50%) scale(0.9);
    background: var(--c-dock-popover-bg);
    color: var(--c-text-primary);
    padding: 6px 10px;
    border-radius: 8px;
    font-size: 0.75rem;
    font-weight: 500;
    white-space: nowrap;
    opacity: 0;
    pointer-events: none;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    box-shadow: var(--c-dock-popover-shadow);
    border: 1px solid var(--c-dock-popover-border);
    z-index: 100;
  }

  .dock-btn[data-tooltip]:not([data-tooltip='']):hover::after {
    opacity: 1;
    transform: translateX(-50%) scale(1);
  }

  /* Right-side tooltip */
  .dock-btn.tooltip-right[data-tooltip]:not([data-tooltip=''])::after {
    bottom: auto;
    top: 50%;
    left: calc(100% + 10px);
    transform: translateY(-50%) scale(0.9);
  }

  .dock-btn.tooltip-right[data-tooltip]:not([data-tooltip='']):hover::after {
    transform: translateY(-50%) scale(1);
  }

  /* Left-side tooltip */
  .dock-btn.tooltip-left[data-tooltip]:not([data-tooltip=''])::after {
    bottom: auto;
    top: 50%;
    left: auto;
    right: calc(100% + 10px);
    transform: translateY(-50%) scale(0.9);
  }

  .dock-btn.tooltip-left[data-tooltip]:not([data-tooltip='']):hover::after {
    transform: translateY(-50%) scale(1);
  }

  /* Bottom-side tooltip */
  .dock-btn.tooltip-bottom[data-tooltip]:not([data-tooltip=''])::after {
    bottom: auto;
    top: calc(100% + 8px);
    left: 50%;
    right: auto;
    transform: translateX(-50%) scale(0.9);
  }

  .dock-btn.tooltip-bottom[data-tooltip]:not([data-tooltip='']):hover::after {
    opacity: 1;
    transform: translateX(-50%) scale(1);
  }

  .dock-btn:hover {
    color: var(--c-text-primary);
    background: var(--c-hover, rgba(255, 255, 255, 0.06));
    border-color: rgba(255, 255, 255, 0.06);
    transform: scale(1.06);
  }

  .dock-btn.active {
    color: var(--c-text-primary, #ffffff);
    background: rgba(255, 255, 255, 0.13);
    border-color: var(--c-border, rgba(255, 255, 255, 0.16));
    box-shadow:
      0 2px 6px rgba(0, 0, 0, 0.25),
      inset 0 1px 0 rgba(255, 255, 255, 0.1);
  }

  .dock-btn.active:hover {
    background: rgba(255, 255, 255, 0.17);
    border-color: rgba(255, 255, 255, 0.2);
    box-shadow:
      0 3px 8px rgba(0, 0, 0, 0.28),
      inset 0 1px 0 rgba(255, 255, 255, 0.12);
  }

  /* Action button hover colors */
  .dock-btn.danger:hover {
    color: #ffffff;
    background: color-mix(in srgb, var(--c-danger, #ff3b30) 35%, transparent);
    border-color: color-mix(in srgb, var(--c-danger, #ff3b30) 60%, transparent);
    box-shadow: 0 2px 8px
      color-mix(in srgb, var(--c-danger, #ff3b30) 30%, transparent);
  }

  .dock-btn.warning:hover {
    color: #ffffff;
    background: color-mix(in srgb, var(--c-warning, #ffcc00) 35%, transparent);
    border-color: color-mix(
      in srgb,
      var(--c-warning, #ffcc00) 60%,
      transparent
    );
    box-shadow: 0 2px 8px
      color-mix(in srgb, var(--c-warning, #ffcc00) 30%, transparent);
  }

  .dock-btn.success:hover {
    color: #ffffff;
    background: color-mix(in srgb, var(--c-success, #34c759) 35%, transparent);
    border-color: color-mix(
      in srgb,
      var(--c-success, #34c759) 60%,
      transparent
    );
    box-shadow: 0 2px 8px
      color-mix(in srgb, var(--c-success, #34c759) 30%, transparent);
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
    background: var(--c-primary, #38bdf8);
    color: #ffffff;
    font-size: 9px;
    font-weight: 700;
    line-height: 12px;
    text-align: center;
    pointer-events: none;
    box-sizing: border-box;
    border: 1.5px solid var(--c-dock-popover-bg);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
  }
</style>
