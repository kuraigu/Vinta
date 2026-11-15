/**
 * @fileoverview Global tooltip state store.
 *
 * Controls which tooltip text is currently visible, its screen position, and alignment.
 */

import { writable } from 'svelte/store'
import { dismissAllTooltips } from '../services/tooltipBounds'

/**
 * Options for configuring a custom tooltip display.
 */
export interface TooltipOptions {
  text: string
  align?: 'left' | 'center' | 'right'
  position?: 'top' | 'bottom'
}

/**
 * Active tooltip state snapshot.
 */
export interface ActiveTooltipState {
  id: string
  text: string
  align: 'left' | 'center' | 'right'
  position: 'top' | 'bottom'
  targetRect: DOMRect | null
  visible: boolean
}

const initialTooltipState: ActiveTooltipState = {
  id: '',
  text: '',
  align: 'center',
  position: 'bottom',
  targetRect: null,
  visible: false
}

export const activeTooltipStore = writable<ActiveTooltipState>(initialTooltipState)
let isTooltipVisible = false

/**
 * Activates the tooltip for a given target element.
 *
 * @param options - Tooltip configuration or text label.
 * @param target - Target DOM element triggering the tooltip.
 */
export function showTooltip(
  options: string | TooltipOptions,
  target: HTMLElement
): void {
  const opts = typeof options === 'string' ? { text: options } : options
  if (!opts.text || !opts.text.trim()) return

  const rect = target.getBoundingClientRect()
  isTooltipVisible = true
  activeTooltipStore.set({
    id: 'tip-' + Math.random().toString(36).slice(2, 9),
    text: opts.text,
    align: opts.align || 'center',
    position: opts.position || 'bottom',
    targetRect: rect,
    visible: true
  })
}

/**
 * Dismisses any currently visible tooltip.
 */
export function hideTooltip(): void {
  if (!isTooltipVisible) return
  isTooltipVisible = false
  activeTooltipStore.update((state) => {
    if (!state.visible) return state
    return { ...state, visible: false }
  })
  dismissAllTooltips()
}
