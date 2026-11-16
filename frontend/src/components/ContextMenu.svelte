<!--
  @component ContextMenu
  @description Floating context menu positioned at cursor coordinates. Automatically flips
  orientation near viewport edges and closes on outside clicks or Escape key.
-->
<script lang="ts">
  import { slide } from 'svelte/transition'
  import { Events } from '$lib/events'

  export interface ContextMenuItem {
    label: string
    action: () => void
    divider?: boolean
    disabled?: boolean
  }

  interface Props {
    x: number
    y: number
    items: ContextMenuItem[]
    onclose: () => void
  }

  let { x, y, items, onclose }: Props = $props()

  let adjustedX = $state(0)
  let adjustedY = $state(0)

  function syncCoordsFromProps() {
    adjustedX = x
    adjustedY = y
  }
  $effect(syncCoordsFromProps)

  function portal(node: HTMLElement) {
    document.body.appendChild(node)

    // Position dynamically inside the island container so it never clips
    // right or bottom and never causes out-of-bounds window expansion.
    requestAnimationFrame(() => {
      const margin = 8
      const naturalW = node.offsetWidth || node.scrollWidth || 150
      const naturalH = node.offsetHeight || node.scrollHeight || 160

      const islandEl = document.querySelector('.island')
      const bounds = islandEl
        ? islandEl.getBoundingClientRect()
        : {
            left: 0,
            top: 0,
            right: window.innerWidth,
            bottom: window.innerHeight
          }

      let finalX = x
      let finalY = y

      // If clipping right of island, flip/shift left
      if (finalX + naturalW > bounds.right - margin) {
        finalX = Math.max(
          bounds.left + margin,
          bounds.right - naturalW - margin
        )
      }

      // If clipping bottom of island, flip/shift up
      if (finalY + naturalH > bounds.bottom - margin) {
        finalY = Math.max(
          bounds.top + margin,
          bounds.bottom - naturalH - margin
        )
      }

      // Ensure clamped to top/left bounds
      finalX = Math.max(bounds.left + margin, finalX)
      finalY = Math.max(bounds.top + margin, finalY)

      adjustedX = finalX
      adjustedY = finalY
      Events.Emit('vinta:resync-bounds')
    })

    return {
      destroy() {
        if (node.parentElement) node.parentElement.removeChild(node)
        Events.Emit('vinta:resync-bounds')
      }
    }
  }

  function clickOutside(node: HTMLElement) {
    const handle = (e: MouseEvent) => {
      if (!node.contains(e.target as Node)) onclose()
    }
    // Delay so the triggering click doesn't immediately close it
    setTimeout(() => document.addEventListener('mousedown', handle), 0)
    return {
      destroy() {
        document.removeEventListener('mousedown', handle)
      }
    }
  }

  function handleItemClick(item: ContextMenuItem) {
    if (item.disabled) return
    onclose()
    item.action()
  }

  function stopPropagation(e: Event) {
    e.stopPropagation()
  }

  function getUiScale(): number {
    return (
      (typeof document !== 'undefined' &&
        parseFloat(
          document.documentElement.style.getPropertyValue('--vinta-ui-scale')
        )) ||
      1
    )
  }
</script>

<div
  use:portal
  use:clickOutside
  class="ctx-menu no-swipe"
  style="left: {adjustedX / getUiScale()}px; top: {adjustedY / getUiScale()}px;"
  transition:slide={{ duration: 100, axis: 'y' }}
  onmousedown={stopPropagation}
  role="menu"
  tabindex="-1"
>
  {#each items as item (item.label)}
    {#if item.divider}
      <div class="ctx-divider"></div>
    {:else}
      <button
        class="ctx-item"
        class:disabled={item.disabled}
        role="menuitem"
        onmousedown={(e) => {
          e.preventDefault()
          handleItemClick(item)
        }}
        tabindex="-1"
      >
        {item.label}
      </button>
    {/if}
  {/each}
</div>

<style>
  .ctx-menu {
    position: fixed;
    zoom: var(--vinta-ui-scale, 1);
    z-index: 999999;
    background: var(--vinta-dock-popover-bg, rgba(20, 22, 28, 0.96));
    border: 1px solid var(--vinta-dock-popover-border, rgba(255, 255, 255, 0.12));
    border-radius: 12px;
    padding: 5px;
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.06),
      var(--vinta-dock-popover-shadow, 0 10px 28px rgba(0, 0, 0, 0.52));
    min-width: 164px;
    pointer-events: auto;
    user-select: none;
  }

  .ctx-item {
    display: flex;
    align-items: center;
    width: 100%;
    min-height: 28px;
    background: none;
    border: none;
    padding: 5px 10px;
    text-align: left;
    font-size: 0.78rem;
    font-weight: 500;
    font-family: inherit;
    color: var(--vinta-text-secondary);
    cursor: pointer;
    border-radius: 7px;
    transition:
      background-color 0.12s ease,
      color 0.12s ease;
    pointer-events: auto;
  }

  .ctx-item:hover {
    background: var(--vinta-surface-hover, rgba(255, 255, 255, 0.1));
    color: var(--vinta-text-primary);
  }

  .ctx-item.disabled {
    opacity: 0.4;
    cursor: default;
    pointer-events: none;
  }

  .ctx-divider {
    height: 1px;
    background: var(--vinta-border, rgba(255, 255, 255, 0.08));
    margin: 4px 6px;
  }
</style>
