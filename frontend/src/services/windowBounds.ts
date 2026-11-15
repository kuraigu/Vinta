/**
 * @fileoverview Synchronizes clickable screen areas with the Go backend.
 *
 * Measures the island capsule, satellites, and menus in screen pixels and forwards
 * the rectangles to the native backend so transparent desktop areas remain clickable.
 */

import {
  SetHitRects,
  TrimMemory,
  SetIsHidden
} from '../../bindings/Vinta/backend/app.js'
import { HitRect } from '../../bindings/Vinta/backend/models.js'
import { resyncAllSatellites } from '../stores/satellites'
import { dismissAllTooltips } from './tooltipBounds'

/**
 * Window region boundary dimensions and corner radius.
 */
export interface WindowBounds {
  w: number
  h: number
  r: number
}

/**
 * Coordinates dynamic island and satellite bounds synchronization with the native Go hit-testing subsystem.
 */
export class WindowBoundsManager {
  private boundsTransitionTimeout: number | undefined
  private activeIslandNode: HTMLElement | null = null
  private idleWatchdogTimer: number | undefined
  private lastHitRectsSignature: string = ''

  public currentOSBounds: WindowBounds = { w: 0, h: 0, r: 0 }
  public targetPendingBounds: WindowBounds = { w: 0, h: 0, r: 0 }

  /**
   * Dynamic getter for live horizontal offset.
   */
  public getLiveOffset: () => number = () => 0

  /**
   * Dynamic getter to check if dragging is currently in progress.
   */
  public getIsDragging: () => boolean = () => false

  /**
   * Measures physical client pixel bounds of the active island, satellites, menus, and folder windows.
   */
  private measureHitRects(
    islandNode?: HTMLElement | null,
    isHidden: boolean = false
  ): { hitRects: HitRect[]; hiddenSig?: string } {
    if (islandNode) {
      this.activeIslandNode = islandNode
    }
    const dpr = window.devicePixelRatio || 1

    if (isHidden) {
      const targetNode = islandNode || this.activeIslandNode
      if (targetNode) {
        const rect = targetNode.getBoundingClientRect()
        if (rect.width > 0 && rect.height > 0) {
          const l = Math.round(rect.left * dpr)
          const t = 0
          const w = Math.round(rect.width * dpr)
          const h = Math.round((rect.bottom + 2) * dpr)
          return {
            hitRects: [new HitRect({ left: l, top: t, width: w, height: h })],
            hiddenSig: `hidden:${l},${t},${w},${h}`
          }
        }
      }
      return { hitRects: [], hiddenSig: 'hidden:empty' }
    }

    const hitRects: HitRect[] = []

    const targetNode = islandNode || this.activeIslandNode

    if (targetNode) {
      const rect = targetNode.getBoundingClientRect()
      if (rect.width > 0 && rect.height > 0) {
        hitRects.push(
          new HitRect({
            left: Math.round(rect.left * dpr),
            top: 0,
            width: Math.round(rect.width * dpr),
            height: Math.round((rect.bottom + 4) * dpr)
          })
        )
      }
    }

    // Include all active satellite hosts (ControlsSatellite, MediaSatellite, etc.)
    if (typeof document !== 'undefined') {
      const satellites =
        document.querySelectorAll<HTMLElement>('.satellite-host')
      for (const sat of satellites) {
        const rect = sat.getBoundingClientRect()
        if (rect.width > 0 && rect.height > 0) {
          // Add a 4px perimeter margin and bridge the 12px gap to the main island
          // so clicks and fast cursor entry do not pass through to background apps.
          const pad = 4
          const l = Math.max(0, Math.round((rect.left - 12 - pad) * dpr))
          const t = Math.max(0, Math.round((rect.top - pad) * dpr))
          const r = Math.round((rect.right + pad) * dpr)
          const b = Math.round((rect.bottom + pad) * dpr)
          hitRects.push(
            new HitRect({
              left: l,
              top: t,
              width: r - l,
              height: b - t
            })
          )
        }
      }

      // Include active context menu if present
      const contextMenu = document.querySelector<HTMLElement>(
        '.ctx-menu, .context-menu'
      )
      if (contextMenu) {
        const rect = contextMenu.getBoundingClientRect()
        if (rect.width > 0 && rect.height > 0) {
          hitRects.push(
            new HitRect({
              left: Math.round(rect.left * dpr),
              top: Math.round(rect.top * dpr),
              width: Math.round(rect.width * dpr),
              height: Math.round(rect.height * dpr)
            })
          )
        }
      }

      // Include simulated showcase desktop folder windows if present and visible
      const folderWindows = document.querySelectorAll<HTMLElement>(
        '.desktop-folder-window'
      )
      for (const win of folderWindows) {
        const rect = win.getBoundingClientRect()
        if (rect.width > 0 && rect.height > 0) {
          hitRects.push(
            new HitRect({
              left: Math.round(rect.left * dpr),
              top: Math.round(rect.top * dpr),
              width: Math.round(rect.width * dpr),
              height: Math.round(rect.height * dpr)
            })
          )
        }
      }
    }

    return { hitRects }
  }

  /**
   * Sends hit rectangles to the Go backend if the serialized signature changed.
   */
  private dispatchHitRects(
    hitRects: HitRect[],
    force: boolean = false,
    explicitSig?: string
  ): void {
    const sig =
      explicitSig ??
      hitRects.map((r) => `${r.left},${r.top},${r.width},${r.height}`).join(';')
    if (!force && sig === this.lastHitRectsSignature) {
      return
    }
    this.lastHitRectsSignature = sig
    SetHitRects(hitRects).catch(() => {})
  }

  /**
   * Synchronizes physical client pixel bounds of the active island and any satellite hosts to the Go backend.
   */
  public syncPillBounds(
    islandNode?: HTMLElement | null,
    isHidden: boolean = false,
    force: boolean = false
  ): void {
    const { hitRects, hiddenSig } = this.measureHitRects(islandNode, isHidden)
    this.dispatchHitRects(hitRects, force, hiddenSig)
  }

  private liveTrackingRaf: number | undefined
  private transitionTimer: number | undefined
  private transitionCleanup: (() => void) | undefined
  private baselineHitRects: HitRect[] | null = null
  private baselineOffsetX: number = 0
  private baselineDpr: number = 1
  private lastSyncedDragOffset: number = Number.NaN
  private lastDragSyncTime: number = 0
  private baselineInvalidated: boolean = false
  private isDragBaselineActive: boolean = false
  private resyncListenerAttached: boolean = false

  private handleResyncBounds = () => {
    if (this.isDragBaselineActive || this.liveTrackingRaf !== undefined) {
      this.baselineInvalidated = true
    }
  }

  /**
   * Captures a single DOM baseline snapshot at drag or glide start so subsequent
   * movement frames shift hit rects arithmetically with zero layout reflows.
   */
  public captureDragBaseline(
    islandNode?: HTMLElement | null,
    explicitOffset?: number
  ): void {
    if (islandNode) this.activeIslandNode = islandNode
    if (typeof window !== 'undefined' && !this.resyncListenerAttached) {
      window.addEventListener('vinta-resync-bounds', this.handleResyncBounds)
      this.resyncListenerAttached = true
    }
    const { hitRects } = this.measureHitRects(this.activeIslandNode, false)
    this.baselineHitRects = hitRects
    this.baselineOffsetX = explicitOffset ?? this.getLiveOffset()
    this.baselineDpr = window.devicePixelRatio || 1
    this.lastSyncedDragOffset = this.baselineOffsetX
    this.lastDragSyncTime = performance.now()
    this.baselineInvalidated = false
    this.isDragBaselineActive = true
    this.dispatchHitRects(hitRects)
  }

  /**
   * Event-driven drag/glide hit-rect update called from pointermove/glide callbacks.
   */
  public syncDragOffset(liveOffset: number): void {
    if (this.baselineInvalidated || !this.baselineHitRects) {
      this.captureDragBaseline(this.activeIslandNode, liveOffset)
      return
    }
    const now = performance.now()
    if (
      now - this.lastDragSyncTime < 32 ||
      Math.abs(liveOffset - this.lastSyncedDragOffset) < 0.5
    ) {
      return
    }
    this.lastDragSyncTime = now
    this.lastSyncedDragOffset = liveOffset
    const deltaPx = Math.round(
      (liveOffset - this.baselineOffsetX) * this.baselineDpr
    )
    const shifted = this.baselineHitRects.map(
      (r) =>
        new HitRect({
          left: Math.max(0, r.left + deltaPx),
          top: r.top,
          width: r.width,
          height: r.height
        })
    )
    this.dispatchHitRects(shifted)
  }

  /**
   * Starts timed live tracking when durationMs is provided, or captures an event-driven
   * drag baseline when durationMs is omitted.
   */
  public startLiveTracking(
    islandNode?: HTMLElement | null,
    durationMs?: number,
    onSettled?: () => void
  ): void {
    if (islandNode) this.activeIslandNode = islandNode
    this.stopLiveTracking()

    if (durationMs === undefined) {
      this.captureDragBaseline(this.activeIslandNode)
      return
    }

    const startTime = performance.now()
    let lastSyncTime = startTime
    const THROTTLE_MS = 32

    const trackStep = (now: number) => {
      if (now - lastSyncTime >= THROTTLE_MS) {
        lastSyncTime = now
        this.syncPillBounds(this.activeIslandNode)
      }
      if (now - startTime >= durationMs) {
        this.stopLiveTracking()
        this.syncPillBounds(this.activeIslandNode)
        onSettled?.()
        return
      }
      this.liveTrackingRaf = requestAnimationFrame(trackStep)
    }

    this.syncPillBounds(this.activeIslandNode)
    this.liveTrackingRaf = requestAnimationFrame(trackStep)
  }

  /**
   * Stops active live tracking loop and clears any drag baseline.
   */
  public stopLiveTracking(): void {
    if (this.liveTrackingRaf !== undefined) {
      cancelAnimationFrame(this.liveTrackingRaf)
      this.liveTrackingRaf = undefined
    }
    this.isDragBaselineActive = false
    this.baselineHitRects = null
    this.baselineInvalidated = false
    this.lastSyncedDragOffset = Number.NaN
  }

  /**
   * Pre-commits a union bounding rectangle covering both current and target dimensions
   * at transition start, then settles via transitionend (with a safety fallback timeout).
   */
  public startTrackingTransition(
    islandNode?: HTMLElement | null,
    durationMs: number = 340,
    isHidden: boolean = false,
    onSettled?: () => void
  ): void {
    if (islandNode) this.activeIslandNode = islandNode
    this.stopTrackingTransition()

    const node = this.activeIslandNode
    if (node && !isHidden) {
      const dpr = window.devicePixelRatio || 1
      const rect = node.getBoundingClientRect()
      if (rect.width > 0 && rect.height > 0) {
        const centerX = rect.left + rect.width / 2
        const unionW = Math.max(
          rect.width,
          this.currentOSBounds.w || rect.width
        )
        const unionH = Math.max(
          rect.bottom + 4,
          (this.currentOSBounds.h || rect.height) + 4
        )
        const l = Math.max(0, Math.round((centerX - unionW / 2) * dpr))
        const w = Math.round(unionW * dpr)
        const h = Math.round(unionH * dpr)
        // Include active satellites, context menus, and folder windows alongside the union rect
        // so satellite buttons and popovers remain clickable during island morphs.
        const { hitRects } = this.measureHitRects(node, false)
        const unionRect = new HitRect({ left: l, top: 0, width: w, height: h })
        const combined =
          hitRects.length > 0 ? [unionRect, ...hitRects.slice(1)] : [unionRect]
        const sig = `transition:${combined.map((r) => `${r.left},${r.top},${r.width},${r.height}`).join(';')}`
        if (sig !== this.lastHitRectsSignature) {
          this.lastHitRectsSignature = sig
          SetHitRects(combined).catch(() => {})
        }
      }
    }

    let settled = false
    const finishOnce = () => {
      if (settled) return
      settled = true
      this.transitionCleanup?.()
      this.transitionCleanup = undefined
      if (this.transitionTimer !== undefined) {
        clearTimeout(this.transitionTimer)
        this.transitionTimer = undefined
      }
      this.syncPillBounds(this.activeIslandNode, isHidden, true)
      resyncAllSatellites()
      dismissAllTooltips()
      onSettled?.()
    }

    // Do not settle on the first width/transform transitionend event: width and height can finish
    // on different frames, which would commit clipped resting bounds and unblock ResizeObserver early.
    this.transitionTimer = window.setTimeout(finishOnce, durationMs)
  }

  /**
   * Stops any active transition timer, listener, or tracking frame.
   */
  public stopTrackingTransition(): void {
    this.stopLiveTracking()
    this.transitionCleanup?.()
    this.transitionCleanup = undefined
    if (this.transitionTimer !== undefined) {
      clearTimeout(this.transitionTimer)
      this.transitionTimer = undefined
    }
  }

  /**
   * Synchronizes bounds during inertia gliding using live coordinate tracking.
   * Tracks the live position of the gliding island frame-by-frame so the vacated path
   * remains immediately click-through to background applications.
   */
  public syncGlideCorridor(
    islandNode: HTMLElement | null,
    _isHidden: boolean,
    _startOffset: number,
    _targetOffset: number
  ): void {
    if (islandNode) this.activeIslandNode = islandNode
    this.startLiveTracking(islandNode)
  }

  /**
   * Dispatches boundary updates with dynamic hit-testing bounds.
   */
  public updateBounds(
    w: number,
    h: number,
    r: number,
    _explicitOffset?: number
  ): Promise<void> {
    this.currentOSBounds = { w, h, r }
    this.syncPillBounds(this.activeIslandNode)
    return Promise.resolve()
  }

  /**
   * Applies target bounds immediately and tracks current state.
   */
  public applyBounds(
    w: number,
    h: number,
    r: number,
    explicitOffset?: number
  ): Promise<void> {
    this.cancelPendingTransition()
    return this.updateBounds(w, h, r, explicitOffset)
  }

  public tryCommitShrinkingBounds(
    islandNode?: HTMLElement | null,
    _explicitOffset?: number,
    _isDragging: boolean = false
  ): void {
    this.syncPillBounds(islandNode)
  }

  /**
   * Resets active shrinking state.
   */
  public cancelPendingTransition(): void {
    this.targetPendingBounds = { w: 0, h: 0, r: 0 }
    this.stopTrackingTransition()
  }

  /**
   * Pre-expansion hook to continuously track expanding bounds in real time.
   */
  public prepareExpansion(
    target: WindowBounds,
    _explicitOffset?: number,
    islandNode?: HTMLElement | null
  ): Promise<void> {
    if (islandNode) this.activeIslandNode = islandNode
    this.currentOSBounds = { ...target }
    this.startTrackingTransition(islandNode, 360, false)
    return Promise.resolve()
  }

  /**
   * Settles bounds when leaving horizontal dragging mode.
   */
  public settleRunwayBounds(
    target: WindowBounds,
    _offset: number,
    _wrapperNode?: HTMLElement | null,
    islandNode?: HTMLElement | null,
    onSettled?: () => void
  ): void {
    if (islandNode) this.activeIslandNode = islandNode
    this.currentOSBounds = { ...target }
    this.startTrackingTransition(islandNode, 360, false, onSettled)
  }

  /**
   * Recovers resting bounds and updates native hit testing.
   */
  public recoverRestingBounds(
    target: WindowBounds,
    _offset: number,
    islandNode?: HTMLElement | null,
    isHidden: boolean = false
  ): Promise<void> {
    if (islandNode) this.activeIslandNode = islandNode
    this.currentOSBounds = { ...target }
    this.stopTrackingTransition()
    this.syncPillBounds(islandNode || this.activeIslandNode, isHidden)
    resyncAllSatellites()
    dismissAllTooltips()
    return Promise.resolve()
  }

  /**
   * Compatibility adapter delegating directly to syncPillBounds.
   */
  public scheduleAnimationRecovery(
    _target: WindowBounds,
    _offset: number,
    islandNode: HTMLElement | null,
    isHidden: boolean = false
  ): void {
    this.syncPillBounds(islandNode, isHidden)
  }

  /**
   * Recovers bounds forcibly.
   */
  public async recoverBounds(
    target: WindowBounds,
    offset: number,
    islandNode?: HTMLElement | null,
    _force: boolean = true
  ): Promise<void> {
    return this.recoverRestingBounds(target, offset, islandNode, false)
  }

  /**
   * Consolidated window bounds synchronizer. Tracks the island across any state transition.
   */
  public syncBounds(
    target: WindowBounds,
    _offset: number,
    islandNode: HTMLElement | null,
    _isDragging: boolean = false,
    isHidden: boolean = false
  ): void {
    if (islandNode) this.activeIslandNode = islandNode
    this.currentOSBounds = { ...target }
    this.syncPillBounds(islandNode, isHidden)
  }

  private trimTimer: number | undefined

  /**
   * Compatibility adapter delegating to consolidated syncBounds and scheduling memory trims on idle/collapsed state changes.
   */
  public syncOnStateChange(
    target: WindowBounds,
    _wasExpanded: boolean,
    expanded: boolean,
    _wasHidden: boolean,
    isHidden: boolean,
    isDragging: boolean,
    offset: number,
    islandNode: HTMLElement | null
  ): void {
    this.syncBounds(target, offset, islandNode, isDragging, isHidden)
    if (expanded && !isHidden) {
      this.cancelScheduledTrimMemory()
    } else if (!isDragging) {
      this.scheduleTrimMemory(isHidden ? 800 : 900)
    }
  }

  /**
   * Cancels any pending scheduled memory trim.
   */
  public cancelScheduledTrimMemory(): void {
    if (this.trimTimer !== undefined) {
      clearTimeout(this.trimTimer)
      this.trimTimer = undefined
    }
  }

  /**
   * Schedules a debounced working-set memory trim after a state change settles.
   * Keep delay >= 800ms so rapid hover/re-expand does not page-fault freshly trimmed memory.
   */
  public scheduleTrimMemory(delayMs: number = 850): void {
    if (typeof window === 'undefined') return
    this.cancelScheduledTrimMemory()
    this.trimTimer = window.setTimeout(() => {
      this.trimTimer = undefined
      this.trimMemory()
    }, delayMs)
  }

  /**
   * Trims unused OS working set memory.
   */
  public trimMemory(): void {
    this.cancelScheduledTrimMemory()
    if (typeof TrimMemory === 'function') {
      TrimMemory().catch(() => {})
    }
  }

  /**
   * Notifies backend whether island is completely hidden.
   */
  public notifyHidden(hidden: boolean): void {
    if (typeof SetIsHidden === 'function') {
      SetIsHidden(hidden).catch(() => {})
    }
    this.syncPillBounds(this.activeIslandNode, hidden)
  }

  /**
   * Initializes an event-driven ResizeObserver on the island node so bounds reconcile
   * whenever DOM geometry changes without a periodic polling timer.
   */
  public startIdleReconciler(
    getIslandNode: () => HTMLElement | null,
    getTargetBounds: () => WindowBounds,
    getIsHidden: () => boolean = () => false
  ): () => void {
    if (typeof window === 'undefined') return () => {}

    let observer: ResizeObserver | undefined
    let mountRaf: number | undefined
    let reconcileRaf: number | undefined
    let observedNode: HTMLElement | null = null

    const reconcileNow = () => {
      if (
        !this.getIsDragging() &&
        !getIsHidden() &&
        this.transitionTimer === undefined
      ) {
        const node = getIslandNode()
        if (node) {
          this.activeIslandNode = node
          this.currentOSBounds = { ...getTargetBounds() }
          this.syncPillBounds(node, false, false)
        }
      }
    }

    const reconcile = () => {
      if (reconcileRaf !== undefined) return
      reconcileRaf = requestAnimationFrame(() => {
        reconcileRaf = undefined
        reconcileNow()
      })
    }

    const attachObserver = () => {
      const node = getIslandNode()
      if (!node) {
        mountRaf = requestAnimationFrame(attachObserver)
        return
      }
      if (node === observedNode) return
      observedNode = node
      reconcileNow()
      if (typeof ResizeObserver !== 'undefined') {
        observer?.disconnect()
        observer = new ResizeObserver(reconcile)
        observer.observe(node)
      }
    }

    attachObserver()

    return () => {
      if (mountRaf !== undefined) {
        cancelAnimationFrame(mountRaf)
        mountRaf = undefined
      }
      if (reconcileRaf !== undefined) {
        cancelAnimationFrame(reconcileRaf)
        reconcileRaf = undefined
      }
      observer?.disconnect()
      observer = undefined
      observedNode = null
    }
  }

  /**
   * Cleans up all pending timers and callbacks.
   */
  public cleanup(): void {
    this.cancelPendingTransition()
    this.cancelScheduledTrimMemory()
    clearTimeout(this.idleWatchdogTimer)
    if (this.resyncListenerAttached && typeof window !== 'undefined') {
      window.removeEventListener('vinta-resync-bounds', this.handleResyncBounds)
      this.resyncListenerAttached = false
    }
  }
}
