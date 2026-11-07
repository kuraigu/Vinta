/**
 * @fileoverview Expands the native window hit-test area to fit floating tooltips.
 *
 * Sends tooltip coordinates to the native backend so Windows doesn't clip tooltips
 * that stick out past the rounded island capsule.
 */

import {
  SetTooltipBounds,
  ClearTooltipBounds
} from '../../bindings/Vinta/backend/app.js'

/**
 * Tells the native window manager to expand the window shape around a visible tooltip.
 */
export function dispatchTooltipBounds(
  id: string,
  rect: DOMRect | { left: number; top: number; width: number; height: number },
  radius: number = 7
): void {
  if (
    typeof document !== 'undefined' &&
    document.body.classList.contains('devtest-window-body')
  ) {
    return
  }
  const scale = window.devicePixelRatio || 1
  if (typeof SetTooltipBounds === 'function') {
    SetTooltipBounds({
      id,
      x: Math.round(rect.left),
      y: Math.round(rect.top),
      width: Math.round(rect.width),
      height: Math.round(rect.height),
      radius,
      scale,
      active: true
    }).catch(() => {})
  }
}

/**
 * Removes an active tooltip's bounding region from the native window mask.
 *
 * @param id - Unique identifier for the tooltip.
 */
export function dismissTooltipBounds(id: string): void {
  if (
    typeof document !== 'undefined' &&
    document.body.classList.contains('devtest-window-body')
  ) {
    return
  }
  if (typeof ClearTooltipBounds === 'function') {
    ClearTooltipBounds(id).catch(() => {})
  }
}

/**
 * Removes all active floating tooltip regions from the native window mask.
 */
export function dismissAllTooltips(): void {
  if (
    typeof document !== 'undefined' &&
    document.body.classList.contains('devtest-window-body')
  ) {
    return
  }
  if (typeof ClearTooltipBounds === 'function') {
    ClearTooltipBounds('*').catch(() => {})
  }
}
