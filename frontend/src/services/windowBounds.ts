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
   * Synchronizes physical client pixel bounds of the active island and any satellite hosts to the Go backend.
   */
  public syncPillBounds(
    islandNode?: HTMLElement | null,
    isHidden: boolean = false,
    force: boolean = false
  ): void {
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
          const sig = `hidden:${l},${t},${w},${h}`
          if (sig !== this.lastHitRectsSignature) {
            this.lastHitRectsSignature = sig
            SetHitRects([
              new HitRect({
                left: l,
                top: t,
                width: w,
                height: h
              })
            ]).catch(() => {})
          }
          return
        }
      }
      if (this.lastHitRectsSignature !== 'hidden:empty') {
        this.lastHitRectsSignature = 'hidden:empty'
        SetHitRects([]).catch(() => {})
      }
      return
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

    const sig = hitRects
      .map((r) => `${r.left},${r.top},${r.width},${r.height}`)
      .join(';')
    if (!force && sig === this.lastHitRectsSignature) {
      return
    }
    this.lastHitRectsSignature = sig
    SetHitRects(hitRects).catch(() => {})
  }

  private liveTrackingRaf: number | undefined
  private transitionTimer: number | undefined

  /**
   * Starts a live requestAnimationFrame tracking loop that reports physical client coordinates
   * of the island to Go during mounting, resizing, dragging, transitions, or gliding.
   */
  public startLiveTracking(
    islandNode?: HTMLElement | null,
    durationMs?: number,
    onSettled?: () => void
  ): void {
    if (islandNode) this.activeIslandNode = islandNode
    this.stopLiveTracking()

    const startTime = performance.now()
    let lastSyncTime = startTime
    const THROTTLE_MS = 32

    const trackStep = (now: number) => {
      if (now - lastSyncTime >= THROTTLE_MS) {
        lastSyncTime = now
        this.syncPillBounds(this.activeIslandNode)
      }
      if (durationMs !== undefined && now - startTime >= durationMs) {
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
   * Stops active live requestAnimationFrame tracking loop.
   */
  public stopLiveTracking(): void {
    if (this.liveTrackingRaf !== undefined) {
      cancelAnimationFrame(this.liveTrackingRaf)
      this.liveTrackingRaf = undefined
    }
  }

  /**
   * Pre-commits a union bounding rectangle covering both current and target dimensions
   * at transition start, then schedules a single DOM measurement when the transition settles.
   * Eliminates 30Hz synchronous layout reflows during expand/collapse morphs.
   */
  public startTrackingTransition(
    islandNode?: HTMLElement | null,
    durationMs: number = 340,
    isHidden: boolean = false
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
        const sig = `transition:${l},0,${w},${h}`
        if (sig !== this.lastHitRectsSignature) {
          this.lastHitRectsSignature = sig
          SetHitRects([
            new HitRect({
              left: l,
              top: 0,
              width: w,
              height: h
            })
          ]).catch(() => {})
        }
      }
    }

    this.transitionTimer = window.setTimeout(() => {
      this.transitionTimer = undefined
      this.syncPillBounds(this.activeIslandNode, isHidden, true)
      resyncAllSatellites()
      dismissAllTooltips()
    }, durationMs)
  }

  /**
   * Stops any active transition timer or tracking frame.
   */
  public stopTrackingTransition(): void {
    this.stopLiveTracking()
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
    this.startTrackingTransition(islandNode, 360, false)
    setTimeout(() => {
      onSettled?.()
    }, 360)
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

  /**
   * Compatibility adapter delegating to consolidated syncBounds.
   */
  public syncOnStateChange(
    target: WindowBounds,
    _wasExpanded: boolean,
    _expanded: boolean,
    _wasHidden: boolean,
    isHidden: boolean,
    isDragging: boolean,
    offset: number,
    islandNode: HTMLElement | null
  ): void {
    this.syncBounds(target, offset, islandNode, isDragging, isHidden)
  }

  /**
   * Trims unused OS working set memory.
   */
  public trimMemory(): void {
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
   * Initializes a background idle reconciler that periodically reconciles true DOM geometry.
   */
  public startIdleReconciler(
    getIslandNode: () => HTMLElement | null,
    getTargetBounds: () => WindowBounds,
    getIsHidden: () => boolean = () => false
  ): () => void {
    if (typeof window === 'undefined') return () => {}

    const runIdleCheck = () => {
      if (!this.getIsDragging() && !getIsHidden()) {
        const node = getIslandNode()
        if (node) {
          this.activeIslandNode = node
          const target = getTargetBounds()
          this.currentOSBounds = { ...target }
          this.syncPillBounds(node, false)
        }
      }
      this.idleWatchdogTimer = window.setTimeout(runIdleCheck, 500)
    }

    this.idleWatchdogTimer = window.setTimeout(runIdleCheck, 500)

    return () => {
      clearTimeout(this.idleWatchdogTimer)
    }
  }

  /**
   * Cleans up all pending timers and callbacks.
   */
  public cleanup(): void {
    this.cancelPendingTransition()
    clearTimeout(this.idleWatchdogTimer)
  }
}
