<!--
  @component HorizontalTabSelection
  @description Segmented pill tab switcher with an animated active highlight bar.
-->
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

<style src="./styles/HorizontalTabSelection.css"></style>
