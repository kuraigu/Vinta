/**
 * @fileoverview Svelte action for attaching tooltips to island buttons.
 *
 * Shows the custom tooltip component on hover and hides it when clicked or scrolled.
 */

import {
  showTooltip,
  hideTooltip,
  type TooltipOptions
} from '../stores/tooltip'

export type IslandTooltipOptions = TooltipOptions

let mountedTooltipCount = 0

function handleGlobalScroll() {
  hideTooltip()
}

/**
 * Attaches a tooltip to an element on pointer hover.
 */
export function islandTooltip(
  node: HTMLElement,
  options: string | IslandTooltipOptions
) {
  let opts = options

  function handleEnter() {
    showTooltip(opts, node)
  }

  function handleLeave() {
    hideTooltip()
  }

  node.addEventListener('pointerenter', handleEnter)
  node.addEventListener('pointerleave', handleLeave)
  node.addEventListener('click', handleLeave)
  if (mountedTooltipCount === 0) {
    window.addEventListener('scroll', handleGlobalScroll, {
      capture: true,
      passive: true
    })
  }
  mountedTooltipCount += 1

  return {
    update(newOptions: string | IslandTooltipOptions) {
      opts = newOptions
    },
    destroy() {
      handleLeave()
      node.removeEventListener('pointerenter', handleEnter)
      node.removeEventListener('pointerleave', handleLeave)
      node.removeEventListener('click', handleLeave)
      mountedTooltipCount = Math.max(0, mountedTooltipCount - 1)
      if (mountedTooltipCount === 0) {
        window.removeEventListener('scroll', handleGlobalScroll, {
          capture: true
        })
      }
    }
  }
}
