<!--
  @component CustomContextMenu
  @description Floating context menu positioned at cursor coordinates. Automatically flips
  orientation near viewport edges and closes on outside clicks or Escape key.
-->
<script lang="ts">
  import type { ComponentType } from 'svelte'
  import { slide } from 'svelte/transition'
  import { Events } from '$lib/events'
  import { portal, getUiScale, clampToBounds } from '../utils/portal'

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

  function mountMenuPortal(node: HTMLElement) {
    return portal(node, {
      onMount: () => {
        requestAnimationFrame(() => {
          const scale = getUiScale()
          const naturalW = (node.offsetWidth || node.scrollWidth || 164) * scale
          const naturalH =
            (node.offsetHeight || node.scrollHeight || 160) * scale
          const clamped = clampToBounds(x, y, naturalW, naturalH)
          adjustedX = clamped.x
          adjustedY = clamped.y
          requestAnimationFrame(() => {
            Events.Emit('vinta:resync-bounds')
          })
        })
      },
      resyncBounds: true
    })
  }

  function clickOutside(node: HTMLElement) {
    const handle = (e: MouseEvent) => {
      if (!node.contains(e.target as Node)) onclose()
    }
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onclose()
    }
    const handleBlur = () => {
      onclose()
    }
    setTimeout(() => {
      document.addEventListener('mousedown', handle)
      document.addEventListener('keydown', handleKey)
      window.addEventListener('blur', handleBlur)
    }, 0)
    return {
      destroy() {
        document.removeEventListener('mousedown', handle)
        document.removeEventListener('keydown', handleKey)
        window.removeEventListener('blur', handleBlur)
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
</script>

<div
  use:mountMenuPortal
  use:clickOutside
  class="ctx-menu no-swipe"
  style="left: {adjustedX / getUiScale()}px; top: {adjustedY / getUiScale()}px;"
  transition:slide={{ duration: 100, axis: 'y' }}
  onintroend={() => Events.Emit('vinta:resync-bounds')}
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
          <IconComp size={14} class="ctx-item-icon" />
        {/if}
        <span>{item.label}</span>
      </button>
    {/if}
  {/each}
</div>

<style src="./styles/CustomContextMenu.css"></style>
