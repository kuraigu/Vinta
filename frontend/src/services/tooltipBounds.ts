/**
 * Tooltip bounds synchronization service.
 * Visual tooltips have pointer-events disabled and do not need Win32 click regions.
 */

/**
 * Dispatches an active tooltip's viewport bounding box.
 * Kept as a no-op so visual hover overlays do not manipulate native window regions.
 *
 * @param _id - Unique identifier for the tooltip.
 * @param _rect - Viewport bounding box of the tooltip element.
 * @param _radius - Optional corner radius in pixels.
 */
export function dispatchTooltipBounds(
  _id: string,
  _rect: DOMRect | { left: number; top: number; width: number; height: number },
  _radius: number = 7
): void {
  // No-op: tooltips are non-interactive visual overlays and do not require Win32 window regions.
}

/**
 * Removes an active tooltip's bounding region.
 * Kept as a no-op so visual hover overlays do not manipulate native window regions.
 *
 * @param _id - Unique identifier for the tooltip.
 */
export function dismissTooltipBounds(_id: string): void {
  // No-op: tooltips are non-interactive visual overlays and do not require Win32 window regions.
}
