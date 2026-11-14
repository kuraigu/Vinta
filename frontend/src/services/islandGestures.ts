/**
 * @fileoverview Dragging and display snapping logic for the island pill.
 *
 * Keeps the island inside screen boundaries, snaps to display centers on multi-monitor
 * setups, and handles smooth deceleration when dragged and released.
 */

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
 * Resolves the physical display monitor corresponding to a horizontal offset.
 *
 * @param offset - Current horizontal live offset from primary display center.
 * @param monitors - List of connected physical monitors.
 * @param baseCenter - Virtual center coordinate (virtualLeft + primaryCenter).
 * @returns The monitor enclosing the coordinate, or the nearest monitor.
 */
export function getMonitorForOffset(
  offset: number,
  monitors: MonitorInfo[] | undefined,
  baseCenter: number
): MonitorInfo | undefined {
  if (!monitors || monitors.length === 0) {
    return undefined
  }
  const currentVirtualX = baseCenter + offset
  for (const mon of monitors) {
    if (currentVirtualX >= mon.left && currentVirtualX < mon.left + mon.width) {
      return mon
    }
  }

  let closestMon = monitors[0]
  let minDistance = Infinity
  for (const mon of monitors) {
    const monCenter = mon.left + mon.width / 2
    const dist = Math.abs(currentVirtualX - monCenter)
    if (dist < minDistance) {
      minDistance = dist
      closestMon = mon
    }
  }
  return closestMon
}

/**
 * Coordinates pointer dragging gestures, runway track region expansion,
 * and magnetic snapping across displays.
 */
export class IslandGestureController {
  public isDragging = false
  public isSettling = false
  public isTrackRegionActive = false
  public isInertia = false
  public dragMovedDistance = 0

  private dragStartX = 0
  private dragStartOffset = 0
  private hasPointerCaptured = false
  private dragResetTimeout: number | undefined
  private dragWatchdogTimeout: number | undefined
  private onExpandRunwayCallback?: () => void
  private hasExpandedRunway = false
  private dragHistory: Array<{ x: number; time: number }> = []
  private inertiaAnimFrame: number | undefined
  private inertiaSafetyTimeout: number | undefined
  private pendingMoveOffset: number | undefined
  private moveRafId: number | undefined

  private cancelPendingMove(): void {
    if (this.moveRafId !== undefined) {
      cancelAnimationFrame(this.moveRafId)
      this.moveRafId = undefined
    }
    this.pendingMoveOffset = undefined
  }

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

    if (!isDragHandle) {
      if (
        target.closest(
          'button, input, select, textarea, a, .custom-select, .no-drag'
        )
      ) {
        return false
      }
      const roleBtn = target.closest('[role="button"]')
      if (
        roleBtn &&
        !roleBtn.classList.contains('notification-layout') &&
        !roleBtn.classList.contains('minimal-slot') &&
        !roleBtn.classList.contains('clickable-slot') &&
        !roleBtn.closest('.notification-layout, .banner-view-container')
      ) {
        return false
      }
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
    this.stopInertia()
    this.cancelPendingMove()
    this.isDragging = true
    this.isSettling = false
    this.hasPointerCaptured = false
    this.dragMovedDistance = 0
    this.hasExpandedRunway = true
    this.isTrackRegionActive = true
    this.onExpandRunwayCallback = onExpandRunway
    clearTimeout(this.dragResetTimeout)
    this.resetWatchdog(onWatchdogFailsafe)

    this.dragStartX = e.screenX
    this.dragStartOffset = currentOffset
    this.dragHistory = [{ x: e.screenX, time: performance.now() }]

    // Immediately apply the 100% full-width runway corridor before pointer motion begins
    onExpandRunway()
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

    const now = performance.now()
    this.dragHistory.push({ x: e.screenX, time: now })
    if (this.dragHistory.length > 35) {
      this.dragHistory.shift()
    }

    const dx = e.screenX - this.dragStartX
    this.dragMovedDistance = Math.max(this.dragMovedDistance, Math.abs(dx))

    if (this.dragMovedDistance > 2) {
      if (!this.hasExpandedRunway) {
        this.hasExpandedRunway = true
        this.isTrackRegionActive = true
        this.onExpandRunwayCallback?.()
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
      this.pendingMoveOffset = clamped
      if (this.moveRafId === undefined) {
        this.moveRafId = requestAnimationFrame(() => {
          this.moveRafId = undefined
          if (!this.isDragging || this.pendingMoveOffset === undefined) {
            return
          }
          const next = this.pendingMoveOffset
          this.pendingMoveOffset = undefined
          onUpdateOffset(next)
        })
      }
    }
  }

  /**
   * Handles pointer release, calculating release velocity and starting inertia glide if flicked.
   */
  public handlePointerUp(
    e: PointerEvent,
    currentOffset: number,
    virtualMin: number,
    virtualMax: number,
    _snapOffsets: number[],
    onUpdateOffset: (newOffset: number) => void,
    onComplete: () => void,
    onStartInertia?: (
      startOffset: number,
      targetOffset: number,
      velocity: number
    ) => void,
    onGlideStep?: (liveOffset: number) => void
  ): void {
    const pendingOffset = this.pendingMoveOffset
    this.cancelPendingMove()
    if (pendingOffset !== undefined) {
      currentOffset = pendingOffset
      onUpdateOffset(pendingOffset)
    }

    if (this.isInertia) {
      return
    }

    if (!this.isDragging && !this.isTrackRegionActive) {
      onComplete()
      return
    }

    this.isDragging = false

    const now = performance.now()
    this.dragHistory.push({ x: e.screenX, time: now })

    // Sample points within the last 150ms of release
    const recent = this.dragHistory.filter((p) => now - p.time <= 150)
    let velocity = 0

    if (recent.length >= 2) {
      const oldest = recent[0]
      const newest = recent[recent.length - 1]
      const dt = newest.time - oldest.time
      const dx = newest.x - oldest.x
      if (dt > 10) {
        velocity = dx / dt
      }
    }

    const maxVelocity = 2.5
    velocity = Math.max(-maxVelocity, Math.min(maxVelocity, velocity))

    if (Math.abs(velocity) >= 0.04 && this.dragMovedDistance > 3) {
      this.isInertia = true
      this.isTrackRegionActive = true

      const friction = 0.0028
      let lastFrameTime = performance.now()
      let liveOffset = currentOffset

      const stoppingDist = velocity / friction
      const targetOffset = Math.max(
        virtualMin,
        Math.min(virtualMax, currentOffset + stoppingDist)
      )
      onStartInertia?.(currentOffset, targetOffset, velocity)

      clearTimeout(this.inertiaSafetyTimeout)
      this.inertiaSafetyTimeout = window.setTimeout(() => {
        this.stopInertia()
        onComplete()
      }, 1500)

      const runInertiaStep = (currentTime: number) => {
        if (!this.isInertia) return

        const frameDt = Math.min(32, currentTime - lastFrameTime)
        lastFrameTime = currentTime

        if (frameDt > 0) {
          const decay = Math.exp(-friction * frameDt)
          velocity *= decay
          liveOffset += velocity * frameDt

          if (liveOffset <= virtualMin) {
            this.stopInertia()
            onUpdateOffset(virtualMin)
            onComplete()
            return
          }
          if (liveOffset >= virtualMax) {
            this.stopInertia()
            onUpdateOffset(virtualMax)
            onComplete()
            return
          }

          if (Math.abs(velocity) < 0.03) {
            this.stopInertia()
            onUpdateOffset(liveOffset)
            onComplete()
            return
          }

          if (onGlideStep) {
            onGlideStep(liveOffset)
          } else {
            onUpdateOffset(liveOffset)
          }
        }

        this.inertiaAnimFrame = requestAnimationFrame(runInertiaStep)
      }

      this.inertiaAnimFrame = requestAnimationFrame(runInertiaStep)
    } else {
      this.stopInertia()
      onComplete()
    }
  }

  /**
   * Cancels active inertia velocity loop and timers.
   */
  public stopInertia(): void {
    if (this.inertiaAnimFrame !== undefined) {
      cancelAnimationFrame(this.inertiaAnimFrame)
      this.inertiaAnimFrame = undefined
    }
    clearTimeout(this.inertiaSafetyTimeout)
    this.inertiaSafetyTimeout = undefined
    this.isInertia = false
    clearTimeout(this.dragResetTimeout)
    this.dragResetTimeout = window.setTimeout(() => {
      this.dragMovedDistance = 0
    }, 50)
  }

  /**
   * Interrupts an active inertia glide immediately, restoring gesture state.
   */
  public interruptInertia(): void {
    this.stopInertia()
    this.isTrackRegionActive = false
    this.hasExpandedRunway = false
    this.dragMovedDistance = 0
    clearTimeout(this.dragWatchdogTimeout)
    this.dragWatchdogTimeout = undefined
    clearTimeout(this.dragResetTimeout)
  }

  /**
   * Concludes the dragging gesture and triggers resting completion.
   */
  public finishDrag(e?: PointerEvent, onComplete?: () => void): void {
    this.stopInertia()
    clearTimeout(this.dragWatchdogTimeout)
    this.dragWatchdogTimeout = undefined
    this.isDragging = false
    this.isSettling = false
    this.isTrackRegionActive = false
    this.hasExpandedRunway = false
    this.onExpandRunwayCallback = undefined

    onComplete?.()

    clearTimeout(this.dragResetTimeout)
    this.dragResetTimeout = window.setTimeout(() => {
      this.dragMovedDistance = 0
    }, 50)
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
      if (this.isDragging || this.isSettling || this.isInertia) {
        this.stopInertia()
        onFailsafe()
      }
    }, 6000)
  }

  /**
   * Cleans up active gesture timers.
   */
  public cleanup(): void {
    this.stopInertia()
    clearTimeout(this.dragResetTimeout)
    clearTimeout(this.dragWatchdogTimeout)
    this.isDragging = false
    this.isSettling = false
    this.isTrackRegionActive = false
    this.hasExpandedRunway = false
    this.onExpandRunwayCallback = undefined
  }
}
