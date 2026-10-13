import type { MonitorInfo } from '../../bindings/Vinta/backend/models.js'

/**
 * Bounds constraints for island placement.
 */
export interface AllowedOffsetRange {
  min: number
  max: number
}

/**
 * Calculates allowable coordinate bounds for the island on the active display monitor.
 */
export function computeAllowedOffsets(
  islandWidth: number,
  monitor: MonitorInfo,
  baseCenter: number,
  leftFootprint: number = 0,
  rightFootprint: number = 0
): AllowedOffsetRange {
  const halfW = islandWidth / 2
  const leftMargin = 12 + leftFootprint
  const rightMargin = 12 + rightFootprint
  const targetLeft = monitor.left + leftMargin + halfW
  const targetRight = monitor.left + monitor.width - rightMargin - halfW

  if (targetLeft > targetRight) {
    const center = Math.floor(monitor.left + monitor.width / 2 - baseCenter)
    return { min: center, max: center }
  }

  return {
    min: Math.floor(targetLeft - baseCenter),
    max: Math.floor(targetRight - baseCenter)
  }
}

/**
 * Computes magnetic snap positions for monitor centers.
 */
export function computeMonitorSnapOffsets(
  monitors: MonitorInfo[] | undefined,
  virtualLeft: number,
  primaryCenter: number
): number[] {
  if (!monitors || monitors.length === 0) {
    return [0]
  }
  return monitors.map((mon) => {
    const monClientLeft = mon.left - virtualLeft
    const monCenter = monClientLeft + mon.width / 2
    return Math.round(monCenter - primaryCenter)
  })
}

/**
 * Coordinates pointer dragging gestures, runway track region expansion,
 * and magnetic snapping across displays.
 */
export class IslandGestureController {
  public isDragging = false
  public isSettling = false
  public isTrackRegionActive = false
  public dragMovedDistance = 0

  private dragStartX = 0
  private dragStartOffset = 0
  private hasPointerCaptured = false
  private dragResetTimeout: number | undefined
  private dragWatchdogTimeout: number | undefined
  private onExpandRunwayCallback?: () => void
  private hasExpandedRunway = false

  /**
   * Evaluates whether an element or event can initiate dragging.
   */
  public canInitiateDrag(
    e: PointerEvent,
    allowDrag: boolean,
    expanded: boolean
  ): boolean {
    if (e.button !== 0 || !allowDrag) return false

    const target = e.target as HTMLElement
    const isDragHandle = Boolean(
      target.closest('.island-drag-handle, .dock-drag-handle')
    )
    const isAltDrag = e.altKey

    if (expanded && !isDragHandle && !isAltDrag) return false

    if (
      !isDragHandle &&
      target.closest(
        'button, input, select, textarea, a, .custom-select, .no-drag, [role="button"]'
      )
    ) {
      return false
    }

    return true
  }

  /**
   * Begins a dragging gesture and prepares movement tracking.
   */
  public startDrag(
    e: PointerEvent,
    currentOffset: number,
    onExpandRunway: () => void,
    onWatchdogFailsafe: () => void
  ): void {
    this.isDragging = true
    this.isSettling = false
    this.hasPointerCaptured = false
    this.dragMovedDistance = 0
    this.hasExpandedRunway = false
    this.isTrackRegionActive = false
    this.onExpandRunwayCallback = onExpandRunway
    clearTimeout(this.dragResetTimeout)
    this.resetWatchdog(onWatchdogFailsafe)

    this.dragStartX = e.screenX
    this.dragStartOffset = currentOffset

    try {
      ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
      this.hasPointerCaptured = true
    } catch {}
  }

  /**
   * Processes pointer motion, expands runway corridor on threshold, and snaps to magnetic centers.
   */
  public handlePointerMove(
    e: PointerEvent,
    virtualMin: number,
    virtualMax: number,
    snapOffsets: number[],
    onUpdateOffset: (newOffset: number) => void,
    onPointerUp: (e: PointerEvent) => void
  ): void {
    if (!this.isDragging) return
    if (e.buttons === 0) {
      onPointerUp(e)
      return
    }

    const dx = e.screenX - this.dragStartX
    this.dragMovedDistance = Math.abs(dx)

    if (this.dragMovedDistance > 2) {
      if (!this.hasExpandedRunway) {
        this.hasExpandedRunway = true
        this.isTrackRegionActive = true
        this.onExpandRunwayCallback?.()
      }

      if (!this.hasPointerCaptured) {
        this.hasPointerCaptured = true
        try {
          ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
        } catch {}
      }

      const rawOffset = this.dragStartOffset + dx
      let targetOffset = rawOffset

      for (const snap of snapOffsets) {
        if (Math.abs(rawOffset - snap) <= 28) {
          targetOffset = snap
          break
        }
      }

      const clamped = Math.max(virtualMin, Math.min(virtualMax, targetOffset))
      onUpdateOffset(clamped)
    }
  }

  /**
   * Concludes the dragging gesture and triggers resting completion.
   */
  public finishDrag(e?: PointerEvent, onComplete?: () => void): void {
    clearTimeout(this.dragWatchdogTimeout)
    this.dragWatchdogTimeout = undefined
    this.isDragging = false
    this.hasExpandedRunway = false
    this.onExpandRunwayCallback = undefined

    if (e && this.hasPointerCaptured) {
      this.hasPointerCaptured = false
      try {
        ;(e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId)
      } catch {}
    }

    onComplete?.()

    clearTimeout(this.dragResetTimeout)
    this.dragResetTimeout = window.setTimeout(() => {
      this.dragMovedDistance = 0
    }, 150)
  }

  /**
   * Concludes the settling phase after drag release and restores normal state tracking.
   */
  public finishSettling(): void {
    this.isSettling = false
    this.isTrackRegionActive = false
  }

  /**
   * Resets the drag inactivity watchdog timer.
   */
  public resetWatchdog(onFailsafe: () => void): void {
    clearTimeout(this.dragWatchdogTimeout)
    this.dragWatchdogTimeout = window.setTimeout(() => {
      if (this.isDragging || this.isSettling) {
        onFailsafe()
      }
    }, 6000)
  }

  /**
   * Cleans up active gesture timers.
   */
  public cleanup(): void {
    clearTimeout(this.dragResetTimeout)
    clearTimeout(this.dragWatchdogTimeout)
    this.isDragging = false
    this.isSettling = false
    this.isTrackRegionActive = false
    this.hasExpandedRunway = false
    this.onExpandRunwayCallback = undefined
  }
}
