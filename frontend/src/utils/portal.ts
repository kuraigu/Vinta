/**
 * @fileoverview Shared DOM portal, UI scale resolution, and viewport/island clamping helpers.
 *
 * Used by CustomTooltip, CustomSelect, and CustomContextMenu so floating popovers
 * escape overflow containers and stay aligned with Vinta's scaled WebView2 surface.
 */

import { Events } from '$lib/events'
import { getCachedUiScale } from '../stores/tooltip'

export interface RectBounds {
  left: number
  top: number
  right: number
  bottom: number
}

export interface PortalOptions {
  target?: HTMLElement
  resyncBounds?: boolean
  onMount?: (node: HTMLElement) => void
  onDestroy?: () => void
}

/**
 * Returns the effective Vinta UI scale factor from the shared tooltip cache,
 * falling back to `--vinta-ui-scale` on `document.documentElement`.
 */
export function getUiScale(): number {
  const cached = getCachedUiScale()
  if (Number.isFinite(cached) && cached > 0) {
    return cached
  }
  if (typeof document !== 'undefined') {
    const raw = parseFloat(
      document.documentElement.style.getPropertyValue('--vinta-ui-scale')
    )
    if (Number.isFinite(raw) && raw > 0) {
      return raw
    }
  }
  return 1
}

/**
 * Returns the bounding box of the `.island` container when large enough to hold the popover,
 * or falls back to the browser viewport (e.g. when the island is minimized or hidden).
 */
export function getIslandOrViewportBounds(
  minWidth: number = 0,
  minHeight: number = 0
): RectBounds {
  if (typeof document !== 'undefined') {
    const islandEl = document.querySelector('.island')
    if (islandEl) {
      const r = islandEl.getBoundingClientRect()
      if (r.width >= minWidth + 16 && r.height >= minHeight + 16) {
        return { left: r.left, top: r.top, right: r.right, bottom: r.bottom }
      }
    }
  }
  if (typeof window !== 'undefined') {
    return {
      left: 0,
      top: 0,
      right: window.innerWidth,
      bottom: window.innerHeight
    }
  }
  return { left: 0, top: 0, right: 800, bottom: 600 }
}

/**
 * Clamps a floating box (`width` × `height`) at `(x, y)` cleanly within `bounds`.
 */
export function clampToBounds(
  x: number,
  y: number,
  width: number,
  height: number,
  bounds: RectBounds = getIslandOrViewportBounds(width, height),
  margin: number = 8
): { x: number; y: number } {
  let finalX = x
  let finalY = y

  if (finalX + width > bounds.right - margin) {
    finalX = Math.max(bounds.left + margin, bounds.right - width - margin)
  }
  if (finalY + height > bounds.bottom - margin) {
    finalY = Math.max(bounds.top + margin, bounds.bottom - height - margin)
  }

  return {
    x: Math.max(bounds.left + margin, finalX),
    y: Math.max(bounds.top + margin, finalY)
  }
}

/**
 * Svelte action mounting a floating element directly on `document.body` (or `target`)
 * and cleaning it up on unmount with optional Win32 hit-test resynchronization.
 */
export function portal(node: HTMLElement, options: PortalOptions = {}) {
  const host = options.target ?? document.body
  host.appendChild(node)

  if (options.onMount) {
    options.onMount(node)
  }
  if (options.resyncBounds) {
    Events.Emit('vinta:resync-bounds')
  }

  return {
    destroy() {
      options.onDestroy?.()
      if (node.parentElement) {
        node.parentElement.removeChild(node)
      }
      if (options.resyncBounds) {
        Events.Emit('vinta:resync-bounds')
      }
    }
  }
}
