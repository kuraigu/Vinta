<svelte:options />

<script module lang="ts">
  /**
   * Represents a selectable tab definition.
   */
  export interface TabItem<K extends string = string> {
    id: K
    label: string
    icon?: any
    tooltip?: string
  }
</script>

<script lang="ts" generics="T extends string = string">
  import { islandTooltip } from '../services/islandTooltip'

  let {
    tabs,
    value = $bindable(),
    onchange,
    colorVariant = 'subtle',
    ariaLabel = 'Navigation tabs'
  }: {
    tabs: TabItem<T>[]
    value: T
    onchange?: (value: T) => void
    colorVariant?: 'subtle' | 'primary' | 'accent' | 'amber' | 'emerald'
    ariaLabel?: string
  } = $props()

  /**
   * Dispatches the tab selection event and updates the bound value.
   *
   * @param id The selected tab identifier.
   */
  function handleSelect(id: T): void {
    value = id
    if (onchange) {
      onchange(id)
    }
  }
</script>

<div
  class="tabs-nav variant-{colorVariant}"
  role="tablist"
  aria-label={ariaLabel}
>
  {#each tabs as tab (tab.id)}
    <button
      class="tab-btn"
      class:active={value === tab.id}
      onclick={() => handleSelect(tab.id)}
      role="tab"
      aria-selected={value === tab.id}
      use:islandTooltip={tab.tooltip || ''}
    >
      {#if tab.icon}
        <tab.icon size={12} />
      {/if}
      <span>{tab.label}</span>
    </button>
  {/each}
</div>

<style>
  .tabs-nav {
    display: flex;
    align-items: center;
    background: var(--c-item-bg, rgba(255, 255, 255, 0.04));
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.08));
    padding: 3px;
    border-radius: 12px;
    gap: 2px;
    flex-shrink: 0;
  }

  .tab-btn {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 4px 9px;
    border-radius: 9px;
    border: none;
    background: transparent;
    color: var(--c-text-secondary, #b0b0b0);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.18s cubic-bezier(0.2, 0.8, 0.2, 1);
    outline: none;
    white-space: nowrap;
    user-select: none;
  }

  .tab-btn:hover {
    color: var(--c-text-primary, #ffffff);
  }

  .tab-btn.active {
    background: var(--c-hover, rgba(255, 255, 255, 0.12));
    color: var(--c-text-primary, #ffffff);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
  }

  .tabs-nav.variant-primary .tab-btn.active {
    background: var(--c-primary, #38bdf8);
    color: var(--c-primary-text, #ffffff);
    font-weight: 700;
    box-shadow: 0 2px 8px rgba(56, 189, 248, 0.35);
  }

  .tabs-nav.variant-amber .tab-btn.active {
    background: #f59e0b;
    color: #ffffff;
    font-weight: 700;
    box-shadow: 0 2px 8px rgba(245, 158, 11, 0.35);
  }

  .tabs-nav.variant-emerald .tab-btn.active {
    background: #10b981;
    color: #ffffff;
    font-weight: 700;
    box-shadow: 0 2px 8px rgba(16, 185, 129, 0.35);
  }

  .tabs-nav.variant-accent .tab-btn.active {
    background: var(--c-accent, var(--c-primary, #38bdf8));
    color: #ffffff;
    font-weight: 700;
  }
</style>
