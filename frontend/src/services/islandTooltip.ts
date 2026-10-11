import {
  showTooltip,
  hideTooltip,
  type TooltipOptions
} from '../stores/tooltip'

export type IslandTooltipOptions = TooltipOptions

/**
 * Svelte action for attaching a custom styled tooltip to an island element.
 * Triggers the CustomTooltip singleton host upon hover.
 *
 * @param node - The DOM element triggering the tooltip on hover.
 * @param options - Text label or configuration object with alignment and position.
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
  window.addEventListener('scroll', handleLeave, { capture: true, passive: true })

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
