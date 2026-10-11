import {
  SetWidgetBounds,
  TrimMemory,
  SetIsHidden
} from '../../bindings/Vinta/backend/app.js'
import { get } from 'svelte/store'
import { debugLogs } from '../devtest/debugConfig'

/**
 * Window region boundary dimensions and corner radius.
 */
export interface WindowBounds {
  w: number
  h: number
  r: number
}

/**
 * Coordinates and schedules Win32 window clipping region updates,
 * IPC deduplication, and smooth shrinking animations.
 */
export class WindowBoundsManager {
  private isBoundsIpcPending = false
  private pendingBoundsArgs: {
    w: number
    h: number
    r: number
    offset: number
  } | null = null
  private lastDispatchedBounds: {
    w: number
    h: number
    r: number
    offset: number
    scale: number
  } | null = null
  private pendingBoundsResolver: (() => void) | null = null

  private boundsTransitionTimeout: number | undefined
  private activeTransitionEndHandler: ((e: TransitionEvent) => void) | null =
    null
  private activeIslandNode: HTMLElement | null = null
  private resizeObserver: ResizeObserver | null = null

  public currentOSBounds: WindowBounds = { w: 0, h: 0, r: 0 }
  public targetPendingBounds: WindowBounds = { w: 0, h: 0, r: 0 }

  /**
   * Dynamic getter for live horizontal offset to avoid stale closures or accidental zero resets.
   */
  public getLiveOffset: () => number = () => 0

  /**
   * Dynamic getter to check if dragging is currently in progress.
   */
  public getIsDragging: () => boolean = () => false

  /**
   * Dispatches Win32 boundary updates with queue deduplication.
   */
  public updateBounds(
    w: number,
    h: number,
    r: number,
    explicitOffset?: number
  ): Promise<void> {
    const computedW = Math.round(w)
    const computedH = Math.round(h)
    const computedR = Math.round(r)
    const scale = window.devicePixelRatio || 1
    const offset = Math.round(
      explicitOffset !== undefined ? explicitOffset : this.getLiveOffset()
    )

    const isCorridorUnchanged =
      computedW === -1 &&
      this.lastDispatchedBounds?.w === -1 &&
      this.lastDispatchedBounds.h === computedH &&
      this.lastDispatchedBounds.scale === scale

    const isTightUnchanged =
      this.lastDispatchedBounds !== null &&
      this.lastDispatchedBounds.w === computedW &&
      this.lastDispatchedBounds.h === computedH &&
      this.lastDispatchedBounds.r === computedR &&
      this.lastDispatchedBounds.offset === offset &&
      this.lastDispatchedBounds.scale === scale

    if (isCorridorUnchanged || isTightUnchanged) {
      return Promise.resolve()
    }

    if (this.isBoundsIpcPending) {
      this.pendingBoundsArgs = {
        w: computedW,
        h: computedH,
        r: computedR,
        offset
      }
      return new Promise<void>((resolve) => {
        const prevResolver = this.pendingBoundsResolver
        this.pendingBoundsResolver = () => {
          if (prevResolver) prevResolver()
          resolve()
        }
      })
    }

    if (get(debugLogs)) {
      console.log('[Vinta:Bounds]', {
        w: computedW,
        h: computedH,
        r: computedR,
        scale,
        offset
      })
    }

    this.isBoundsIpcPending = true

    return Promise.resolve(
      SetWidgetBounds(computedW, computedH, computedR, scale, offset)
    )
      .then(() => {
        this.lastDispatchedBounds = {
          w: computedW,
          h: computedH,
          r: computedR,
          offset,
          scale
        }
      })
      .catch((err) => {
        this.lastDispatchedBounds = null
        if (get(debugLogs)) {
          console.error('[Vinta:BoundsError]', err)
        }
      })
      .finally(() => {
        this.isBoundsIpcPending = false
        const resolver = this.pendingBoundsResolver
        this.pendingBoundsResolver = null
        if (this.pendingBoundsArgs) {
          const next = this.pendingBoundsArgs
          this.pendingBoundsArgs = null
          this.updateBounds(next.w, next.h, next.r, next.offset).then(() => {
            if (resolver) resolver()
          })
        } else if (resolver) {
          resolver()
        }
      })
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
    this.currentOSBounds = { w, h, r }
    return this.updateBounds(w, h, r, explicitOffset)
  }

  /**
   * Commits pending shrinking region bounds to the OS window manager.
   * Clears any active fallback timer and applies target dimensions immediately.
   */
  public tryCommitShrinkingBounds(
    islandNode?: HTMLElement | null,
    explicitOffset?: number,
    isDragging: boolean = false
  ): void {
    if (isDragging || this.getIsDragging()) return
    if (this.targetPendingBounds.w <= 0 && this.targetPendingBounds.h <= 0) {
      return
    }

    const offset =
      explicitOffset !== undefined ? explicitOffset : this.getLiveOffset()

    const toApply = { ...this.targetPendingBounds }
    this.cancelPendingTransition()

    if (toApply.w <= 0 || toApply.h <= 0) {
      return
    }

    // Safety guard against clipping under heavy load:
    // If islandNode has not reached target dimensions yet, retain a safe envelope
    // so Windows DWM does not clip the visible animated element.
    if (islandNode) {
      const liveW = islandNode.offsetWidth
      const liveH = islandNode.offsetHeight
      if (liveW > toApply.w + 6 || liveH > toApply.h + 6) {
        this.applyBounds(
          Math.max(toApply.w, liveW),
          Math.max(toApply.h, liveH),
          toApply.r,
          offset
        )
        return
      }
    }

    this.applyBounds(toApply.w, toApply.h, toApply.r, offset)
  }

  /**
   * Resets active shrinking timeout, disconnects observers, and removes lingering transition listeners.
   */
  public cancelPendingTransition(): void {
    clearTimeout(this.boundsTransitionTimeout)
    this.boundsTransitionTimeout = undefined
    this.targetPendingBounds = { w: 0, h: 0, r: 0 }
    if (this.resizeObserver) {
      this.resizeObserver.disconnect()
      this.resizeObserver = null
    }
    if (this.activeIslandNode && this.activeTransitionEndHandler) {
      this.activeIslandNode.removeEventListener(
        'transitionend',
        this.activeTransitionEndHandler
      )
      this.activeIslandNode.removeEventListener(
        'transitioncancel',
        this.activeTransitionEndHandler
      )
    }
    this.activeIslandNode = null
    this.activeTransitionEndHandler = null
  }

  /**
   * Pre-expands the Win32 window clipping region before an expanding CSS transition begins.
   * This guarantees that hardware DWM paint masks are already large enough before DOM elements
   * start growing, preventing any boundary clipping under heavy system load.
   */
  public async prepareExpansion(
    target: WindowBounds,
    explicitOffset?: number
  ): Promise<void> {
    const currW = this.currentOSBounds.w
    const currH = this.currentOSBounds.h

    const isGrowing = target.w > currW || target.h > currH
    if (!isGrowing) {
      return
    }

    const envelopeW = Math.max(currW, target.w)
    const envelopeH = Math.max(currH, target.h)
    const offset =
      explicitOffset !== undefined ? explicitOffset : this.getLiveOffset()

    this.cancelPendingTransition()
    this.currentOSBounds = { w: envelopeW, h: envelopeH, r: target.r }

    try {
      await Promise.race([
        this.updateBounds(envelopeW, envelopeH, target.r, offset),
        new Promise<void>((resolve) => window.setTimeout(resolve, 60))
      ])
    } catch {}
  }

  /**
   * Consolidated window bounds synchronizer.
   * Compares the target bounds against current OS bounds and determines
   * whether to expand immediately (to prevent clipping), defer shrinking
   * until CSS transitions finish, or apply directly.
   */
  public syncBounds(
    target: WindowBounds,
    offset: number,
    islandNode: HTMLElement | null,
    isDragging: boolean = false,
    isHidden: boolean = false
  ): void {
    if (isDragging || this.getIsDragging()) return

    // Restore tight hit-testing bounds immediately when leaving dragging runway
    if (this.lastDispatchedBounds?.w === -1) {
      this.cancelPendingTransition()
      this.applyBounds(target.w, target.h, target.r, offset)
      return
    }

    // Initial bounds application on startup
    if (this.currentOSBounds.w <= 0 && this.currentOSBounds.h <= 0) {
      this.cancelPendingTransition()
      this.applyBounds(target.w, target.h, target.r, offset)
      return
    }

    const currW = this.currentOSBounds.w
    const currH = this.currentOSBounds.h

    const isWidthGrowing = target.w > currW
    const isHeightGrowing = target.h > currH
    const isWidthShrinking = target.w < currW
    const isHeightShrinking = target.h < currH

    const isGrowing = isWidthGrowing || isHeightGrowing
    const isShrinking = isWidthShrinking || isHeightShrinking

    // Keep active transition timer running if already scheduled for the exact same target
    if (
      this.boundsTransitionTimeout !== undefined &&
      this.targetPendingBounds.w === target.w &&
      this.targetPendingBounds.h === target.h
    ) {
      this.targetPendingBounds.r = target.r
      return
    }

    // Growing: expand Win32 window clipping envelope upfront to avoid edge clipping
    if (isGrowing) {
      this.cancelPendingTransition()
      const immediateW = Math.max(currW, target.w)
      const immediateH = Math.max(currH, target.h)

      this.applyBounds(immediateW, immediateH, target.r, offset)
      return
    }

    // Shrinking: defer clipping region commit until CSS transition completes
    if (isShrinking) {
      const isNodeAtTarget =
        islandNode &&
        Math.abs(islandNode.offsetWidth - target.w) <= 3 &&
        Math.abs(islandNode.offsetHeight - target.h) <= 3

      if (isNodeAtTarget || !islandNode) {
        this.cancelPendingTransition()
        this.applyBounds(target.w, target.h, target.r, offset)
        return
      }

      this.cancelPendingTransition()
      this.targetPendingBounds = { ...target }

      let hasCommitted = false
      const commitOnce = () => {
        if (hasCommitted) return
        hasCommitted = true
        this.tryCommitShrinkingBounds(islandNode, offset, false)
      }

      const checkDimensions = () => {
        if (!islandNode) return
        const wDiff = Math.abs(islandNode.offsetWidth - target.w)
        const hDiff = Math.abs(islandNode.offsetHeight - target.h)
        if (wDiff <= 3 && hDiff <= 3) {
          commitOnce()
        }
      }

      const onTransitionEnd = (e: TransitionEvent) => {
        if (e.target !== islandNode) return
        checkDimensions()
      }

      this.activeIslandNode = islandNode
      this.activeTransitionEndHandler = onTransitionEnd
      islandNode.addEventListener('transitionend', onTransitionEnd)
      islandNode.addEventListener('transitioncancel', onTransitionEnd)

      if (typeof ResizeObserver !== 'undefined') {
        this.resizeObserver = new ResizeObserver(() => {
          checkDimensions()
        })
        this.resizeObserver.observe(islandNode)
      }

      this.boundsTransitionTimeout = window.setTimeout(commitOnce, 850)
      return
    }

    // Unchanged dimensions with offset or corner radius changes
    this.cancelPendingTransition()
    this.applyBounds(target.w, target.h, target.r, offset)
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
  }

  /**
   * Cleans up all pending timers and callbacks.
   */
  public cleanup(): void {
    this.cancelPendingTransition()
    this.pendingBoundsResolver = null
    this.pendingBoundsArgs = null
  }
}
