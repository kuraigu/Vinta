<script lang="ts">
  import { onDestroy, type Snippet } from 'svelte'
  import { fade } from 'svelte/transition'
  import {
    activeTooltipStore,
    showTooltip,
    hideTooltip,
    type TooltipOptions
  } from '../stores/tooltip'
  import {
    dispatchTooltipBounds,
    dismissTooltipBounds
  } from '../services/tooltipBounds'

  let {
    text,
    position = 'bottom',
    align = 'center',
    children
  } = $props<{
    text?: string
    position?: 'top' | 'bottom'
    align?: 'left' | 'center' | 'right'
    children?: Snippet
  }>()

  let triggerNode = $state<HTMLElement>()
  let tooltipNode = $state<HTMLDivElement>()
  let lastDispatchedId = $state('')

  /**
   * Action portaling the tooltip element into document.body to escape transformed containers.
   *
   * @param node - The tooltip DOM element to portal.
   */
  function portal(node: HTMLElement) {
    document.body.appendChild(node)
    return {
      destroy() {
        if (node.parentElement) {
          node.parentElement.removeChild(node)
        }
      }
    }
  }

  /**
   * Calculates viewport coordinates and CSS transform for the active tooltip.
   */
  function computeStyle(): string {
    const tip = $activeTooltipStore
    if (!tip.targetRect) return 'display: none;'

    const target = tip.targetRect
    const margin = 8
    let top = 0
    let transform = ''
    let left = 0

    if (tip.position === 'top') {
      top = target.top - margin
      transform = 'translate(-50%, -100%)'
    } else {
      top = target.bottom + margin
      transform = 'translateX(-50%)'
    }

    if (tip.align === 'left') {
      left = target.left
      transform = tip.position === 'top' ? 'translateY(-100%)' : 'none'
    } else if (tip.align === 'right') {
      left = target.right
      transform =
        tip.position === 'top' ? 'translate(-100%, -100%)' : 'translateX(-100%)'
    } else {
      left = target.left + target.width / 2
    }

    return `left: ${Math.round(left)}px; top: ${Math.round(top)}px; transform: ${transform};`
  }

  let computedStyle = $derived.by(computeStyle)

  /**
   * Synchronizes the tooltip's rendered screen bounding box with native Win32 window regions.
   */
  function syncBoundsEffect() {
    const isVisible = $activeTooltipStore.visible
    const id = $activeTooltipStore.id

    if (isVisible && tooltipNode && id) {
      const rect = tooltipNode.getBoundingClientRect()
      if (rect.width > 0 && rect.height > 0) {
        dispatchTooltipBounds(id, rect, 7)
        lastDispatchedId = id
      }
    } else if (!isVisible && lastDispatchedId) {
      dismissTooltipBounds(lastDispatchedId)
      lastDispatchedId = ''
    }
  }

  $effect(syncBoundsEffect)

  onDestroy(() => {
    if (lastDispatchedId) {
      dismissTooltipBounds(lastDispatchedId)
      lastDispatchedId = ''
    }
  })
</script>

{#if children}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    bind:this={triggerNode}
    class="custom-tooltip-trigger"
    role="presentation"
    onpointerenter={() => {
      if (text && triggerNode) {
        showTooltip({ text, position, align }, triggerNode)
      }
    }}
    onpointerleave={hideTooltip}
    onclick={hideTooltip}
  >
    {@render children()}
  </div>
{:else if $activeTooltipStore.visible && $activeTooltipStore.text}
  <div
    bind:this={tooltipNode}
    class="custom-tooltip align-{$activeTooltipStore.align} pos-{$activeTooltipStore.position}"
    style={computedStyle}
    transition:fade={{ duration: 120 }}
    use:portal
  >
    {$activeTooltipStore.text}
  </div>
{/if}

<style src="./styles/CustomTooltip.css"></style>
