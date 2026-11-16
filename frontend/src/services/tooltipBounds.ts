/**
 * @fileoverview Tooltip bounds helpers.
 *
 * CustomTooltip renders with `pointer-events: none` inside an unclipped overlay,
 * so native Win32 tooltip region IPC updates are unnecessary no-ops.
 */

/**
 * No-op retained for API compatibility; tooltips are pointer-transparent.
 */
export function dispatchTooltipBounds(
  _id: string,
  _rect: DOMRect | { left: number; top: number; width: number; height: number },
  _radius: number = 7
): void {}

/**
 * No-op retained for API compatibility.
 */
export function dismissTooltipBounds(_id: string): void {}

/**
 * No-op retained for API compatibility.
 */
export function dismissAllTooltips(): void {}
