<script lang="ts">
  import type { Component } from 'svelte'
  import { ChevronRight } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'

  interface Props {
    label: string
    icon?: Component
    tooltip?: string
    onclick?: (e: MouseEvent) => void
    class?: string
  }

  let {
    label,
    icon: IconComponent,
    tooltip,
    onclick,
    class: className = ''
  }: Props = $props()
</script>

<button
  type="button"
  class="action-capsule-btn {className}"
  {onclick}
  use:islandTooltip={tooltip || label}
  aria-label={label}
>
  {#if IconComponent}
    <IconComponent size={12} class="capsule-icon" />
  {/if}
  <span class="capsule-label">{label}</span>
  <ChevronRight size={13} class="capsule-arrow" />
</button>

<style>
  .action-capsule-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 3px 9px;
    border-radius: 999px;
    border: none;
    background: var(--vinta-collapse-btn-bg, rgba(255, 255, 255, 0.08));
    color: var(--vinta-text-secondary);
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    cursor: pointer;
    transition: all 0.2s ease;
    user-select: none;
    -webkit-user-select: none;
    line-height: 1.2;
    flex-shrink: 0;
  }

  .action-capsule-btn:hover {
    background: var(--vinta-collapse-btn-hover-bg, var(--vinta-hover));
    color: var(--vinta-text-primary);
  }

  .action-capsule-btn:active {
    transform: scale(0.96);
  }

  :global(.capsule-icon) {
    opacity: 0.85;
    flex-shrink: 0;
  }

  :global(.capsule-arrow) {
    opacity: 0.7;
    flex-shrink: 0;
    transition: transform 0.15s ease;
  }

  .action-capsule-btn:hover :global(.capsule-arrow) {
    opacity: 1;
    transform: translateX(1px);
  }
</style>
