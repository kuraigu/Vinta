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
    isHidden: boolean = false
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
          SetHitRects([
            new HitRect({
              left: Math.round(rect.left * dpr),
              top: Math.round(rect.top * dpr),
              width: Math.round(rect.width * dpr),
              height: Math.round(rect.height * dpr)
            })
          ]).catch(() => {})
          return
        }
      }
      SetHitRects([]).catch(() => {})
      return
    }

    const hitRects: HitRect[] = []

    const targetNode = islandNode || this.activeIslandNode
    const node = targetNode?.parentElement?.classList.contains('island-wrapper')
      ? targetNode.parentElement
      : targetNode

    if (node) {
      const rect = node.getBoundingClientRect()
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

    // Include all active satellite hosts (ControlsSatellite, MediaSatellite, etc.)
    if (typeof document !== 'undefined') {
      const satellites =
        document.querySelectorAll<HTMLElement>('.satellite-host')
      for (const sat of satellites) {
        const rect = sat.getBoundingClientRect()
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
    }

    SetHitRects(hitRects).catch(() => {})
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
  }

  /**
   * Pre-expansion hook to maintain compatibility with animation lifecycle.
   */
  public async prepareExpansion(
    _target: WindowBounds,
    _explicitOffset?: number,
    islandNode?: HTMLElement | null
  ): Promise<void> {
    this.syncPillBounds(islandNode, false)
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
    this.syncPillBounds(islandNode || this.activeIslandNode)
    if (typeof requestAnimationFrame !== 'undefined') {
      requestAnimationFrame(() => {
        this.syncPillBounds(islandNode || this.activeIslandNode)
      })
    }
    setTimeout(() => {
      onSettled?.()
      this.syncPillBounds(islandNode || this.activeIslandNode)
    }, 150)
    setTimeout(() => {
      onSettled?.()
      this.syncPillBounds(islandNode || this.activeIslandNode)
    }, 360)
  }

  /**
   * Recovers resting bounds and updates native hit testing.
   */
  public async recoverRestingBounds(
    target: WindowBounds,
    _offset: number,
    islandNode?: HTMLElement | null,
    isHidden: boolean = false
  ): Promise<void> {
    if (islandNode) this.activeIslandNode = islandNode
    this.currentOSBounds = { ...target }
    this.syncPillBounds(islandNode || this.activeIslandNode, isHidden)
    if (typeof requestAnimationFrame !== 'undefined') {
      requestAnimationFrame(() => {
        this.syncPillBounds(islandNode || this.activeIslandNode, isHidden)
      })
    }
    setTimeout(() => {
      this.syncPillBounds(islandNode || this.activeIslandNode, isHidden)
    }, 150)
    setTimeout(() => {
      this.syncPillBounds(islandNode || this.activeIslandNode, isHidden)
    }, 360)
    resyncAllSatellites()
    dismissAllTooltips()
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
   * Consolidated window bounds synchronizer.
   */
  public syncBounds(
    target: WindowBounds,
    _offset: number,
    islandNode: HTMLElement | null,
    _isDragging: boolean = false,
    isHidden: boolean = false
  ): void {
    this.activeIslandNode = islandNode
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
    getTargetBounds: () => WindowBounds
  ): () => void {
    if (typeof window === 'undefined') return () => {}

    const runIdleCheck = () => {
      if (!this.getIsDragging()) {
        const node = getIslandNode()
        if (node) {
          this.activeIslandNode = node
          const target = getTargetBounds()
          this.currentOSBounds = { ...target }
          this.syncPillBounds(node, false)
        }
      }
      this.idleWatchdogTimer = window.setTimeout(runIdleCheck, 400)
    }

    this.idleWatchdogTimer = window.setTimeout(runIdleCheck, 400)

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
