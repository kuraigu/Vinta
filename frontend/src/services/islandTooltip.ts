/**
 * Options for configuring an island element's custom tooltip.
 */
export interface IslandTooltipOptions {
  text: string
  align?: 'left' | 'center' | 'right'
  position?: 'top' | 'bottom'
}

/**
 * Svelte action for attaching a custom styled tooltip to an island element.
 * The tooltip is rendered as a fixed-position span appended to the node's
 * nearest positioned ancestor so it inherits CSS variable context.
 *
 * @param node - The DOM element triggering the tooltip on hover.
 * @param options - Text label or configuration object with alignment and position.
 */
export function islandTooltip(
  node: HTMLElement,
  options: string | IslandTooltipOptions
) {
  let opts = typeof options === 'string' ? { text: options } : options
  let tipEl: HTMLSpanElement | null = null

  function updatePosition() {
    if (!tipEl) return
    const nodeRect = node.getBoundingClientRect()
    const tipW = tipEl.offsetWidth
    const tipH = tipEl.offsetHeight

    let top = 0
    if (opts.position === 'bottom') {
      top = nodeRect.bottom + 8
    } else {
      top = nodeRect.top - tipH - 8
    }

    let left = 0
    if (opts.align === 'left') {
      left = nodeRect.left
    } else if (opts.align === 'right') {
      left = nodeRect.right - tipW
    } else {
      left = nodeRect.left + (nodeRect.width / 2) - (tipW / 2)
    }

    tipEl.style.left = `${Math.round(left)}px`
    tipEl.style.top = `${Math.round(top)}px`
  }

  function show() {
    if (!opts.text || !opts.text.trim()) return
    hide()

    tipEl = document.createElement('span')
    const alignClass = opts.align ? `align-${opts.align}` : 'align-center'
    const posClass = opts.position ? `pos-${opts.position}` : 'pos-top'
    tipEl.className = `island-custom-tooltip ${alignClass} ${posClass}`
    tipEl.textContent = opts.text

    tipEl.style.position = 'fixed'
    tipEl.style.left = '-9999px'
    tipEl.style.top = '-9999px'
    tipEl.style.zIndex = '99999'
    tipEl.style.pointerEvents = 'none'

    document.body.appendChild(tipEl)
    updatePosition()
  }

  function hide() {
    if (tipEl) {
      tipEl.remove()
      tipEl = null
    }
  }

  node.addEventListener('pointerenter', show)
  node.addEventListener('pointerleave', hide)
  node.addEventListener('click', hide)

  return {
    update(newOptions: string | IslandTooltipOptions) {
      opts = typeof newOptions === 'string' ? { text: newOptions } : newOptions
      if (!opts.text || !opts.text.trim()) {
        hide()
        return
      }
      if (tipEl) {
        tipEl.textContent = opts.text
        updatePosition()
      }
    },
    destroy() {
      hide()
      node.removeEventListener('pointerenter', show)
      node.removeEventListener('pointerleave', hide)
      node.removeEventListener('click', hide)
    }
  }
}
