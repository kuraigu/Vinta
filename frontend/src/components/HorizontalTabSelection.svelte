<!--
  @component HorizontalTabSelection
  @description Segmented pill tab switcher with a compositor-only sliding indicator,
  roving keyboard focus, wheel scrolling, and scroll-driven edge fades.
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
    badge?: number | string
  }
</script>

<script lang="ts" generics="T extends string = string">
  import { onMount } from 'svelte'
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

  let tablistEl = $state<HTMLDivElement | null>(null)
  let tabEls = $state<(HTMLButtonElement | null)[]>([])
  let indicatorX = $state(0)
  let indicatorW = $state(0)
  let hasIndicator = $state(false)

  /**
   * Measures the active tab element to position the compositor-only pill indicator.
   */
  function updateIndicator(): void {
    const activeIdx = tabs.findIndex((t) => t.id === value)
    const activeBtn = activeIdx >= 0 ? tabEls[activeIdx] : null
    if (!activeBtn) {
      hasIndicator = false
      return
    }
    indicatorX = activeBtn.offsetLeft
    indicatorW = activeBtn.offsetWidth
    hasIndicator = indicatorW > 0
  }

  $effect(function syncActiveTabPosition() {
    void value
    void tabs
    const raf = requestAnimationFrame(() => {
      updateIndicator()
      const activeIdx = tabs.findIndex((t) => t.id === value)
      const activeBtn = activeIdx >= 0 ? tabEls[activeIdx] : null
      if (activeBtn) {
        activeBtn.scrollIntoView({
          block: 'nearest',
          inline: 'nearest',
          container: 'nearest'
        } as Parameters<HTMLElement['scrollIntoView']>[0])
      }
    })
    return () => cancelAnimationFrame(raf)
  })

  onMount(() => {
    if (!tablistEl || typeof ResizeObserver === 'undefined') return
    const ro = new ResizeObserver(() => {
      updateIndicator()
    })
    ro.observe(tablistEl)
    for (const el of tabEls) {
      if (el) ro.observe(el)
    }
    return () => ro.disconnect()
  })

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

  /**
   * Implements roving tabindex arrow, Home, and End navigation across the tablist.
   */
  function handleKeyDown(e: KeyboardEvent, index: number): void {
    if (tabs.length === 0) return
    const isRtl =
      tablistEl !== null && getComputedStyle(tablistEl).direction === 'rtl'
    let nextIdx = -1

    if (e.key === 'ArrowRight') {
      nextIdx = isRtl
        ? (index - 1 + tabs.length) % tabs.length
        : (index + 1) % tabs.length
    } else if (e.key === 'ArrowLeft') {
      nextIdx = isRtl
        ? (index + 1) % tabs.length
        : (index - 1 + tabs.length) % tabs.length
    } else if (e.key === 'Home') {
      nextIdx = 0
    } else if (e.key === 'End') {
      nextIdx = tabs.length - 1
    }

    if (nextIdx >= 0) {
      e.preventDefault()
      handleSelect(tabs[nextIdx].id)
      tabEls[nextIdx]?.focus()
    }
  }

  /**
   * Translates vertical mouse-wheel scrolling into horizontal tablist scrolling on overflow.
   */
  function handleWheel(e: WheelEvent): void {
    if (!tablistEl || tablistEl.scrollWidth <= tablistEl.clientWidth) return
    if (Math.abs(e.deltaY) > Math.abs(e.deltaX)) {
      e.preventDefault()
      tablistEl.scrollLeft += e.deltaY
    }
  }
</script>

<div
  bind:this={tablistEl}
  class="tabs-nav variant-{colorVariant}"
  role="tablist"
  aria-label={ariaLabel}
  onwheel={handleWheel}
>
  <div
    class="tab-indicator"
    class:visible={hasIndicator}
    aria-hidden="true"
    style="transform: translate3d({indicatorX}px, 0, 0) scaleX({indicatorW});"
  ></div>

  {#each tabs as tab, i (tab.id)}
    {@const isActive = value === tab.id}
    <button
      bind:this={tabEls[i]}
      type="button"
      class="tab-btn"
      class:active={isActive}
      tabindex={isActive ? 0 : -1}
      onclick={() => handleSelect(tab.id)}
      onkeydown={(e) => handleKeyDown(e, i)}
      role="tab"
      aria-selected={isActive}
      use:islandTooltip={tab.tooltip || ''}
    >
      {#if tab.icon}
        <tab.icon size={12} />
      {/if}
      <span>{tab.label}</span>
      {#if tab.badge !== undefined && tab.badge !== '' && tab.badge !== 0}
        <span class="tab-badge">{tab.badge}</span>
      {/if}
    </button>
  {/each}
</div>

<style src="./styles/HorizontalTabSelection.css"></style>
