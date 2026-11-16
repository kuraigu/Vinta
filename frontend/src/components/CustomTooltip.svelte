<!--
  @component CustomTooltip
  @description Floating tooltip container. Positions itself relative to its trigger
  element and calls dispatchTooltipBounds() so Windows expands the window hit-test
  region, preventing the tooltip from getting clipped by OS window borders.
-->
<script lang="ts">
  import { onDestroy, type Snippet } from 'svelte'
  import { fade } from 'svelte/transition'
  import {
    activeTooltipStore,
    showTooltip,
    hideTooltip,
    type TooltipPosition
  } from '../stores/tooltip'
  import {
    dispatchTooltipBounds,
    dismissTooltipBounds,
    dismissAllTooltips
  } from '../services/tooltipBounds'

  interface Props {
    text?: string
    subtitle?: string
    position?: TooltipPosition
    align?: 'left' | 'center' | 'right'
    children?: Snippet
  }

  let {
    text = undefined,
    subtitle = undefined,
    position = 'bottom',
    align = 'center',
    children = undefined
  }: Props = $props()

  let triggerNode = $state<HTMLElement>()
  let tooltipNode = $state<HTMLDivElement>()
  let lastDispatchedId = $state('')
  let resolvedPosition = $state<TooltipPosition>('bottom')
  let clampOffset = $state<{ x: number; y: number }>({ x: 0, y: 0 })

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
    const pos = resolvedPosition
    let top = 0
    let transform = ''
    let left = 0

    if (pos === 'left') {
      left = target.left - margin
      top = target.top + target.height / 2
      transform = 'translate(-100%, -50%)'
    } else if (pos === 'right') {
      left = target.right + margin
      top = target.top + target.height / 2
      transform = 'translateY(-50%)'
    } else {
      if (pos === 'top') {
        top = target.top - margin
        transform = 'translate(-50%, -100%)'
      } else {
        top = target.bottom + margin
        transform = 'translateX(-50%)'
      }

      if (tip.align === 'left') {
        left = target.left
        transform = pos === 'top' ? 'translateY(-100%)' : 'none'
      } else if (tip.align === 'right') {
        left = target.right
        transform =
          pos === 'top' ? 'translate(-100%, -100%)' : 'translateX(-100%)'
      } else {
        left = target.left + target.width / 2
      }
    }

    const uiScale =
      (typeof document !== 'undefined' &&
        parseFloat(
          document.documentElement.style.getPropertyValue('--vinta-ui-scale')
        )) ||
      1
    const adjLeft = Math.round((left + clampOffset.x) / uiScale)
    const adjTop = Math.round((top + clampOffset.y) / uiScale)
    return `left: ${adjLeft}px; top: ${adjTop}px; transform: ${transform};`
  }

  let computedStyle = $derived.by(computeStyle)

  let lastMeasuredId = ''

  /**
   * Synchronizes the tooltip's rendered screen bounding box with native Win32 window regions,
   * flipping or clamping if the tooltip would clip past the viewport edges.
   */
  function syncBoundsEffect() {
    const tip = $activeTooltipStore
    const isVisible = tip.visible
    const id = tip.id

    if (isVisible && tooltipNode && id) {
      if (lastMeasuredId !== id) {
        lastMeasuredId = id
        resolvedPosition = tip.position
        clampOffset = { x: 0, y: 0 }
        requestAnimationFrame(() => {
          if (!tooltipNode || $activeTooltipStore.id !== id) return
          const rect = tooltipNode.getBoundingClientRect()
          if (rect.width <= 0 || rect.height <= 0) return

          const pad = 8
          const maxW = typeof window !== 'undefined' ? window.innerWidth - pad : 1912
          const maxH = typeof window !== 'undefined' ? window.innerHeight - pad : 1072
          let nextPos = tip.position
          if (nextPos === 'right' && rect.right > maxW && (tip.targetRect?.left ?? 0) - 8 - rect.width >= pad) {
            nextPos = 'left'
          } else if (nextPos === 'left' && rect.left < pad && (tip.targetRect?.right ?? 0) + 8 + rect.width <= maxW) {
            nextPos = 'right'
          }
          if (nextPos !== resolvedPosition) {
            resolvedPosition = nextPos
            return
          }

          let dx = 0
          let dy = 0
          if (rect.left < pad) dx = pad - rect.left
          else if (rect.right > maxW) dx = maxW - rect.right
          if (rect.top < pad) dy = pad - rect.top
          else if (rect.bottom > maxH) dy = maxH - rect.bottom
          if (dx !== 0 || dy !== 0) {
            clampOffset = { x: dx, y: dy }
            return
          }

          if (lastDispatchedId && lastDispatchedId !== id) {
            dismissTooltipBounds(lastDispatchedId)
          }
          dispatchTooltipBounds(id, rect, 7)
          lastDispatchedId = id
        })
        return
      }

      if (lastDispatchedId && lastDispatchedId !== id) {
        dismissTooltipBounds(lastDispatchedId)
      }
      const rect = tooltipNode.getBoundingClientRect()
      if (rect.width > 0 && rect.height > 0) {
        dispatchTooltipBounds(id, rect, 7)
        lastDispatchedId = id
      }
    } else if (!isVisible && lastDispatchedId) {
      lastMeasuredId = ''
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
    dismissAllTooltips()
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
        showTooltip({ text, subtitle, position, align }, triggerNode)
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
    class="custom-tooltip align-{$activeTooltipStore.align} pos-{resolvedPosition}"
    style={computedStyle}
    transition:fade={{ duration: 120 }}
    use:portal
  >
    <span class="tooltip-text">{$activeTooltipStore.text}</span>
    {#if $activeTooltipStore.subtitle}
      <span class="tooltip-subtitle">{$activeTooltipStore.subtitle}</span>
    {/if}
  </div>
{/if}

<style src="./styles/CustomTooltip.css"></style>
