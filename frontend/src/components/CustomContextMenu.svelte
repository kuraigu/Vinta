<!--
  @component CustomContextMenu
  @description Floating context menu positioned at cursor coordinates. Automatically flips
  orientation near viewport edges and closes on outside clicks or Escape key.
-->
<script lang="ts">
  import type { ComponentType } from 'svelte'
  import { slide } from 'svelte/transition'
  import { Events } from '$lib/events'

  export interface ContextMenuItem {
    label: string
    action: () => void
    icon?: ComponentType<any>
    danger?: boolean
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

      if (finalX + naturalW > bounds.right - margin) {
        finalX = Math.max(
          bounds.left + margin,
          bounds.right - naturalW - margin
        )
      }

      if (finalY + naturalH > bounds.bottom - margin) {
        finalY = Math.max(
          bounds.top + margin,
          bounds.bottom - naturalH - margin
        )
      }

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
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onclose()
    }
    setTimeout(() => {
      document.addEventListener('mousedown', handle)
      document.addEventListener('keydown', handleKey)
    }, 0)
    return {
      destroy() {
        document.removeEventListener('mousedown', handle)
        document.removeEventListener('keydown', handleKey)
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
  {#each items as item, idx (item.divider ? `div-${idx}` : item.label)}
    {#if item.divider}
      <div class="ctx-divider"></div>
    {:else}
      <button
        type="button"
        class="ctx-item"
        class:disabled={item.disabled}
        class:danger={item.danger}
        role="menuitem"
        onmousedown={(e) => {
          e.preventDefault()
          handleItemClick(item)
        }}
        tabindex="-1"
      >
        {#if item.icon}
          {@const IconComp = item.icon}
          <IconComp size={13} class="ctx-item-icon" />
        {/if}
        <span>{item.label}</span>
      </button>
    {/if}
  {/each}
</div>

<style src="./styles/CustomContextMenu.css"></style>
