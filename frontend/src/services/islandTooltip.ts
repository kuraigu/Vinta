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
  window.addEventListener('scroll', handleLeave, {
    capture: true,
    passive: true
  })

  return {
    update(newOptions: string | IslandTooltipOptions) {
      opts = newOptions
    },
    destroy() {
      handleLeave()
      node.removeEventListener('pointerenter', handleEnter)
      node.removeEventListener('pointerleave', handleLeave)
      node.removeEventListener('click', handleLeave)
      window.removeEventListener('scroll', handleLeave, { capture: true })
    }
  }
}
