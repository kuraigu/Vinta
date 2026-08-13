<script lang="ts">
  import { slide } from 'svelte/transition'

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

  let adjustedX = $state(x)
  let adjustedY = $state(y)

  function notifyOverlay(open: boolean, rect?: { x: number; y: number; w: number; h: number }) {
    window.dispatchEvent(new CustomEvent('vinta-overlay', { detail: open ? { open, ...rect } : { open } }))
  }

  function portal(node: HTMLElement) {
    document.body.appendChild(node)

    // Position after the browser has painted so scrollWidth/scrollHeight
    // reflect the true content size (not 0 from the slide transition start).
    requestAnimationFrame(() => {
      const margin = 8
      const naturalW = node.scrollWidth
      const naturalH = node.scrollHeight

      let finalX = x
      let finalY = y
      if (finalX + naturalW > window.innerWidth - margin) {
        finalX = x - naturalW
      }
      if (finalY + naturalH > window.innerHeight - margin) {
        finalY = y - naturalH
      }
      finalX = Math.min(Math.max(margin, finalX), Math.max(margin, window.innerWidth - naturalW - margin))
      finalY = Math.min(Math.max(margin, finalY), Math.max(margin, window.innerHeight - naturalH - margin))

      adjustedX = finalX
      adjustedY = finalY

      notifyOverlay(true, { x: finalX, y: finalY, w: naturalW, h: naturalH })
    })

    return {
      destroy() {
        notifyOverlay(false)
        if (node.parentElement) node.parentElement.removeChild(node)
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
</script>

<div
  use:portal
  use:clickOutside
  class="ctx-menu no-swipe"
  style="left: {adjustedX}px; top: {adjustedY}px;"
  transition:slide={{ duration: 100, axis: 'y' }}
  onmousedown={stopPropagation}
  role="menu"
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
    z-index: 999999;
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    padding: 4px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
    min-width: 148px;
    pointer-events: auto;
    user-select: none;
  }

  .ctx-item {
    display: block;
    width: 100%;
    background: none;
    border: none;
    padding: 6px 12px;
    text-align: left;
    font-size: 12.5px;
    font-family: inherit;
    color: var(--c-text-secondary);
    cursor: pointer;
    border-radius: 5px;
    transition:
      background 0.1s,
      color 0.1s;
    pointer-events: auto;
  }

  .ctx-item:hover {
    background: var(--c-hover);
    color: var(--c-text-primary);
  }

  .ctx-item.disabled {
    opacity: 0.4;
    cursor: default;
    pointer-events: none;
  }

  .ctx-divider {
    height: 1px;
    background: var(--c-border);
    margin: 4px 6px;
  }
</style>
