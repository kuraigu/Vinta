import {
  SetWidgetBounds,
  TrimMemory,
  SetIsHidden
} from '../../bindings/Vinta/backend/app.js'
import { get } from 'svelte/store'
import { debugLogs } from '../devtest/debugConfig'
import { resyncAllSatellites } from '../stores/satellites'

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
  private recoveryDebounceTimer?: number
  private transitionEndDebounceTimer?: number

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

    const ipcCall = Promise.resolve(
      SetWidgetBounds(computedW, computedH, computedR, scale, offset)
    )
    const timeoutCall = new Promise<void>((_, reject) => {
      window.setTimeout(
        () => reject(new Error('SetWidgetBounds timeout')),
        1800
      )
    })

    return Promise.race([ipcCall, timeoutCall])
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

    this.recoverRestingBounds(toApply, offset, islandNode)
  }

  /**
   * Resets active shrinking timeout, disconnects observers, and removes lingering transition listeners.
   */
  public cancelPendingTransition(): void {
    clearTimeout(this.recoveryDebounceTimer)
    this.recoveryDebounceTimer = undefined
    clearTimeout(this.transitionEndDebounceTimer)
    this.transitionEndDebounceTimer = undefined
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
    explicitOffset?: number,
    islandNode?: HTMLElement | null
  ): Promise<void> {
    const currW = this.currentOSBounds.w
    const currH = this.currentOSBounds.h

    const isGrowing = target.w > currW || target.h > currH
    if (!isGrowing) {
      return
    }

    let envelopeW = Math.max(currW, target.w)
    let envelopeH = Math.max(currH, target.h)
    if (islandNode) {
      const rect = islandNode.getBoundingClientRect()
      if (rect.width > envelopeW) envelopeW = Math.round(rect.width)
      if (rect.height > envelopeH) envelopeH = Math.round(rect.height)
    }
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

    this.scheduleAnimationRecovery(
      target,
      offset,
      islandNode || null,
      false,
      380
    )
  }

  /**
   * Smoothly settles bounds when leaving horizontal dragging runway mode.
   * Keeps the wide corridor clipping envelope active while the island finishes translating
   * to its resting offset, preventing Win32 clipping during slower or frame-dropped animations.
   */
  public settleRunwayBounds(
    target: WindowBounds,
    offset: number,
    wrapperNode?: HTMLElement | null,
    islandNode?: HTMLElement | null,
    onSettled?: () => void
  ): void {
    this.cancelPendingTransition()

    let hasSettled = false
    const commitSettled = () => {
      if (hasSettled) return
      hasSettled = true
      if (wrapperNode) {
        wrapperNode.removeEventListener('transitionend', onTransitionEnd)
        wrapperNode.removeEventListener('transitioncancel', onTransitionEnd)
      }
      this.cancelPendingTransition()
      const liveOffset = this.getLiveOffset()
      const finalOffset = Number.isFinite(liveOffset) ? liveOffset : offset
      this.recoverRestingBounds(target, finalOffset, islandNode)
      onSettled?.()
    }

    if (!wrapperNode) {
      commitSettled()
      return
    }

    const onTransitionEnd = (e: TransitionEvent) => {
      if (
        e.target === wrapperNode &&
        (e.propertyName === 'transform' || e.propertyName === 'left')
      ) {
        commitSettled()
      }
    }

    this.activeIslandNode = wrapperNode
    this.activeTransitionEndHandler = onTransitionEnd
    wrapperNode.addEventListener('transitionend', onTransitionEnd)
    wrapperNode.addEventListener('transitioncancel', onTransitionEnd)

    // Fast-settle fallback timer (160ms for smooth snap transition completion)
    this.boundsTransitionTimeout = window.setTimeout(commitSettled, 160)
  }

  /**
   * Forcibly recovers the exact resting window clipping region to the target bounds.
   * Cancels lingering transitions, updates OS window bounds, re-syncs all satellites,
   * and runs a follow-up cache-bust check to guarantee the window region strictly matches
   * the visible island without blocking external mouse clicks.
   */
  public async recoverRestingBounds(
    target: WindowBounds,
    offset: number,
    islandNode?: HTMLElement | null,
    _isHidden: boolean = false
  ): Promise<void> {
    if (this.getIsDragging()) return
    this.cancelPendingTransition()

    clearTimeout(this.recoveryDebounceTimer)
    const delay = this.lastDispatchedBounds?.w === -1 ? 0 : 25
    this.recoveryDebounceTimer = window.setTimeout(async () => {
      if (this.getIsDragging()) return
      this.isBoundsIpcPending = false
      this.pendingBoundsArgs = null
      this.lastDispatchedBounds = null

      let finalW = Math.round(target.w)
      let finalH = Math.round(target.h)
      const finalR = Math.round(target.r)
      let needsFollowUp = false

      if (islandNode) {
        const rect = islandNode.getBoundingClientRect()
        if (rect.width > 0 && rect.height > 0) {
          const domW = Math.round(rect.width)
          const domH = Math.round(rect.height)
          if (domW > finalW) {
            finalW = domW
            needsFollowUp = true
          }
          if (domH > finalH) {
            finalH = domH
            needsFollowUp = true
          }
        }
      }

      this.currentOSBounds = { w: finalW, h: finalH, r: finalR }
      const liveOffset = this.getLiveOffset()
      const effectiveOffset = Number.isFinite(liveOffset) ? liveOffset : offset
      await this.updateBounds(finalW, finalH, finalR, effectiveOffset)
      resyncAllSatellites()

      // Post-animation secondary verification to bust Windows DWM cache under lag
      window.setTimeout(() => {
        if (this.getIsDragging()) return
        resyncAllSatellites()
        const currentLive = this.getLiveOffset()
        const followUpOffset = Number.isFinite(currentLive)
          ? currentLive
          : effectiveOffset
        // Pass -finalR to bust native cache in Go and guarantee SetWindowRgn
        this.updateBounds(finalW, finalH, -finalR, followUpOffset)
      }, 80)

      if (needsFollowUp) {
        window.setTimeout(() => {
          if (this.getIsDragging()) return
          this.recoverRestingBounds(target, offset, islandNode, _isHidden)
        }, 120)
      }
    }, delay)
  }

  /**
   * Coordinates animation transitions and guarantees that once the CSS transition completes,
   * the exact resting bounds are recovered.
   */
  public scheduleAnimationRecovery(
    target: WindowBounds,
    offset: number,
    islandNode: HTMLElement | null,
    isHidden: boolean = false,
    durationMs: number = 380
  ): void {
    this.cancelPendingTransition()
    this.targetPendingBounds = { ...target }

    let hasRecovered = false
    let safetyAttempts = 0
    const commitRecovery = () => {
      if (hasRecovered) return
      clearTimeout(this.transitionEndDebounceTimer)
      this.transitionEndDebounceTimer = undefined

      if (islandNode && safetyAttempts < 8) {
        const rect = islandNode.getBoundingClientRect()
        const isStillOversized =
          rect.width - target.w > 4 || rect.height - target.h > 4
        // Prevent premature clipping during frame drops if element is still actively shrinking
        if (isStillOversized) {
          safetyAttempts++
          this.boundsTransitionTimeout = window.setTimeout(commitRecovery, 80)
          return
        }
      }

      hasRecovered = true
      this.cancelPendingTransition()
      const liveOffset = this.getLiveOffset()
      const finalOffset = Number.isFinite(liveOffset) ? liveOffset : offset
      this.recoverRestingBounds(target, finalOffset, islandNode, isHidden)
    }

    const triggerDebouncedCommit = () => {
      clearTimeout(this.transitionEndDebounceTimer)
      this.transitionEndDebounceTimer = window.setTimeout(() => {
        commitRecovery()
      }, 40)
    }

    if (islandNode) {
      const onTransitionEnd = (e: TransitionEvent) => {
        if (e.target !== islandNode) return
        if (
          e.propertyName === 'width' ||
          e.propertyName === 'height' ||
          e.propertyName === 'border-radius' ||
          e.propertyName === 'border-bottom-left-radius' ||
          e.propertyName === 'border-bottom-right-radius' ||
          e.propertyName === 'transform'
        ) {
          triggerDebouncedCommit()
        }
      }

      this.activeIslandNode = islandNode
      this.activeTransitionEndHandler = onTransitionEnd
      islandNode.addEventListener('transitionend', onTransitionEnd)
      islandNode.addEventListener('transitioncancel', onTransitionEnd)

      if (typeof ResizeObserver !== 'undefined') {
        this.resizeObserver = new ResizeObserver((entries) => {
          for (const entry of entries) {
            const { width, height } = entry.contentRect
            if (
              Math.abs(width - target.w) <= 2 &&
              Math.abs(height - target.h) <= 2
            ) {
              triggerDebouncedCommit()
            }
          }
        })
        this.resizeObserver.observe(islandNode)
      }
    }

    // Safety fallback timer to guarantee recovery under frame drops or lag
    this.boundsTransitionTimeout = window.setTimeout(commitRecovery, durationMs)
  }

  /**
   * Recovers and re-synchronizes the OS window clipping region.
   * Clears stuck IPC queues, cancels any stale transitions, and forcibly commits
   * the true resting bounds, live offset, and all active satellites to the Win32 window manager.
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

    // Settle bounds safely when leaving dragging runway
    if (this.lastDispatchedBounds?.w === -1) {
      const parentWrapper = islandNode?.parentElement
      if (parentWrapper) {
        this.settleRunwayBounds(target, offset, parentWrapper, islandNode)
        return
      }
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

    // Growing: expand Win32 window clipping envelope upfront to avoid edge clipping,
    // and schedule post-animation recovery so bounds snap back to the exact target when done.
    if (isGrowing) {
      this.cancelPendingTransition()
      let immediateW = Math.max(currW, target.w)
      let immediateH = Math.max(currH, target.h)
      if (islandNode) {
        const rect = islandNode.getBoundingClientRect()
        if (rect.width > immediateW) immediateW = Math.round(rect.width)
        if (rect.height > immediateH) immediateH = Math.round(rect.height)
      }

      this.applyBounds(immediateW, immediateH, target.r, offset)
      this.scheduleAnimationRecovery(target, offset, islandNode, isHidden, 380)
      return
    }

    // Shrinking: defer clipping region commit until CSS transition completes
    // so animated elements are not clipped mid-flight, then recover exact resting bounds.
    if (isShrinking) {
      const isNodeAtTarget =
        islandNode &&
        Math.abs(islandNode.offsetWidth - target.w) <= 2 &&
        Math.abs(islandNode.offsetHeight - target.h) <= 2

      if (isNodeAtTarget || !islandNode) {
        this.recoverRestingBounds(target, offset, islandNode, isHidden)
        return
      }

      this.scheduleAnimationRecovery(target, offset, islandNode, isHidden, 380)
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

  private idleWatchdogTimer?: number

  /**
   * Initializes a background idle watchdog that periodically reconciles true DOM geometry
   * against active OS bounds when the interface is stationary.
   */
  public startIdleReconciler(
    getIslandNode: () => HTMLElement | null,
    getTargetBounds: () => WindowBounds
  ): () => void {
    if (typeof window === 'undefined') return () => {}

    const runIdleCheck = () => {
      if (
        !this.getIsDragging() &&
        this.boundsTransitionTimeout === undefined &&
        !this.isBoundsIpcPending
      ) {
        const node = getIslandNode()
        if (node) {
          const rect = node.getBoundingClientRect()
          if (rect.width > 0 && rect.height > 0) {
            const domW = Math.round(rect.width)
            const domH = Math.round(rect.height)
            const diffW = Math.abs(domW - this.currentOSBounds.w)
            const diffH = Math.abs(domH - this.currentOSBounds.h)
            const currentLiveOffset = this.getLiveOffset()
            const diffOffset = this.lastDispatchedBounds
              ? Math.abs(currentLiveOffset - this.lastDispatchedBounds.offset)
              : 0
            if (diffW > 4 || diffH > 4 || diffOffset > 2) {
              const target = getTargetBounds()
              this.recoverRestingBounds(target, currentLiveOffset, node)
            }
          }
        }
      }
      this.idleWatchdogTimer = window.setTimeout(runIdleCheck, 400)
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
    this.pendingBoundsResolver = null
    this.pendingBoundsArgs = null
  }
}
