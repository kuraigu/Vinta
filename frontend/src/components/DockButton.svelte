<script lang="ts">
  import type { Snippet } from 'svelte'

  type Props = {
    onclick: (e: MouseEvent) => void
    tooltip: string
    variant?: '' | 'danger' | 'warning' | 'success'
    active?: boolean
    children: Snippet
  }

  let {
    onclick,
    tooltip,
    variant = '',
    active = false,
    children
  }: Props = $props()
</script>

<!-- Notice {onclick} replaces on:click={onclick} in Svelte 5 -->
<button
  class="dock-btn {variant}"
  class:active
  {onclick}
  data-tooltip={tooltip}
>
  {@render children()}
</button>

<style>
  .dock-btn {
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    width: 32px;
    height: 32px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
    position: relative;
  }

  .dock-btn[data-tooltip]:not([data-tooltip=''])::after {
    content: attr(data-tooltip);
    position: absolute;
    bottom: calc(100% + 12px);
    left: 50%;
    transform: translateX(-50%) scale(0.9);
    background: rgba(20, 20, 20, 0.95);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    color: #fff;
    padding: 6px 10px;
    border-radius: 8px;
    font-size: 0.75rem;
    font-weight: 500;
    white-space: nowrap;
    opacity: 0;
    pointer-events: none;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.1);
    z-index: 100;
  }

  :global([data-theme='light'])
    .dock-btn[data-tooltip]:not([data-tooltip=''])::after {
    background: rgba(255, 255, 255, 0.95);
    color: #000;
    border: 1px solid rgba(0, 0, 0, 0.1);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .dock-btn[data-tooltip]:not([data-tooltip='']):hover::after {
    opacity: 1;
    transform: translateX(-50%) scale(1);
  }

  .dock-btn:hover {
    color: var(--c-text-primary);
    background: var(--c-hover);
    transform: scale(1.1);
  }

  .dock-btn.active {
    color: var(--c-text-primary);
    background: var(--c-item-bg);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  /* Traffic light hover colors for actions */
  .dock-btn.danger:hover {
    color: var(--c-danger-hover-text);
    background: var(--c-danger-hover-bg);
  }

  .dock-btn.warning:hover {
    color: var(--c-warning-hover-text);
    background: var(--c-warning-hover-bg);
  }

  .dock-btn.success:hover {
    color: var(--c-success-hover-text);
    background: var(--c-success-hover-bg);
  }
</style>
