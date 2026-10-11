import {
  SetTooltipBounds,
  ClearTooltipBounds
} from '../../bindings/Vinta/backend/app.js'

/**
 * Tooltip bounds synchronization service.
 * Dispatches active tooltip bounding boxes to the native layer so Win32 SetWindowRgn
 * includes the tooltip area, preventing OS-level window clipping when tooltips extend
 * beyond the main island pill.
 */

/**
 * Dispatches an active tooltip's viewport bounding box.
 *
 * @param id - Unique identifier for the tooltip.
 * @param rect - Viewport bounding box of the tooltip element.
 * @param radius - Corner radius in pixels.
 */
export function dispatchTooltipBounds(
  id: string,
  rect: DOMRect | { left: number; top: number; width: number; height: number },
  radius: number = 7
): void {
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
  if (typeof ClearTooltipBounds === 'function') {
    ClearTooltipBounds(id).catch(() => {})
  }
}
