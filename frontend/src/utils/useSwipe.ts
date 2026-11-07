/**
 * @fileoverview Touch and mouse drag scroll action with momentum inertia.
 *
 * Allows smooth dragging and flick scrolling across feed lists while ignoring
 * clicks on sliders, buttons, and text inputs.
 */

/** Configuration options for the swipe and drag-scroll action. */
export type SwipeOptions = {
  axis?: 'x' | 'y' | 'both'
  enableSwipe?: boolean
  threshold?: number
  dragThreshold?: number
  ignoreSelector?: string
  onSwipeLeft?: () => void
  onSwipeRight?: () => void
  onSwipeUp?: () => void
  onSwipeDown?: () => void
  onDragStart?: () => void
  onDragMove?: (offset: { x: number; y: number }) => void
  onDragEnd?: () => void
}

interface ExtendedPointerEvent extends PointerEvent {
  _swipeHandledAxis?: 'x' | 'y' | 'both'
}

/**
 * Resolves the primary scrollable container.
 * Prefers the node itself, or checks ancestors and target if needed.
 */
function findScrollableContainer(
  node: HTMLElement,
  target: HTMLElement | null
): HTMLElement {
  // If the node itself is configured for scrolling, always use it
  if (
    node.classList.contains('scroll-view') ||
    node.scrollHeight > node.clientHeight
  ) {
    return node
  }

  const nodeStyle = window.getComputedStyle(node)
  if (nodeStyle.overflowY === 'auto' || nodeStyle.overflowY === 'scroll') {
    return node
  }

  let current = target
  while (current && current !== document.body) {
    const style = window.getComputedStyle(current)
    const isScrollable =
      (style.overflowY === 'auto' ||
        style.overflowY === 'scroll' ||
        style.overflow === 'auto' ||
        style.overflow === 'scroll') &&
      current.scrollHeight > current.clientHeight

    if (isScrollable) {
      return current
    }
    current = current.parentElement
  }
  return node
}

/**
 * Svelte action providing fluid touch and drag scrolling with momentum inertia
 * while protecting nested clickable targets.
 */
export function swipe(node: HTMLElement, options: SwipeOptions = {}) {
  let opts = options
  let isDragging = false
  let hasMovedFar = false
  let preventNextClick = false
  let clearClickTimeout: ReturnType<typeof setTimeout> | null = null
  let isVerticalScroll = false
  let isLockedToAxis = false
  let scrollTarget: HTMLElement | null = null
  let startX = 0
  let startY = 0
  let lastY = 0
  let lastMoveTime = 0
  let velocityY = 0
  let inertiaFrame = 0

  function handlePointerDown(e: PointerEvent) {
    const extEvent = e as ExtendedPointerEvent
    const requestedAxis = opts.axis ?? 'both'

    if (
      extEvent._swipeHandledAxis === 'both' ||
      extEvent._swipeHandledAxis === requestedAxis
    ) {
      return
    }
    if (opts.enableSwipe === false) return

    // Do not initiate drag on interactive inputs or sliders
    const formInputs =
      'input, select, textarea, range, .custom-select-trigger, .custom-select-options, .custom-select-option, .no-swipe, .no-drag'
    const fullIgnoreSelector = opts.ignoreSelector
      ? `${formInputs}, ${opts.ignoreSelector}`
      : formInputs

    if ((e.target as HTMLElement).closest(fullIgnoreSelector)) return
    if (e.button !== 0 && e.pointerType === 'mouse') return

    cancelAnimationFrame(inertiaFrame)
    if (clearClickTimeout) {
      clearTimeout(clearClickTimeout)
      clearClickTimeout = null
    }

    isDragging = true
    hasMovedFar = false
    preventNextClick = false
    isVerticalScroll = false
    isLockedToAxis = false
    scrollTarget = findScrollableContainer(node, e.target as HTMLElement)
    startX = e.clientX
    startY = e.clientY
    lastY = e.clientY
    lastMoveTime = performance.now()
    velocityY = 0

    opts.onDragStart?.()
  }

  function handlePointerMove(e: PointerEvent) {
    if (!isDragging) return
    const now = performance.now()
    const dt = Math.max(1, now - lastMoveTime)
    lastMoveTime = now

    const dx = e.clientX - startX
    const dy = e.clientY - startY
    const stepY = e.clientY - lastY
    lastY = e.clientY

    // Track running velocity with exponential moving average
    const instantVelocity = stepY / dt
    velocityY = 0.7 * instantVelocity + 0.3 * velocityY

    const dragThresh = opts.dragThreshold ?? 6
    const requestedAxis = opts.axis ?? 'both'

    if (
      !hasMovedFar &&
      (Math.abs(dx) > dragThresh || Math.abs(dy) > dragThresh)
    ) {
      const isVertical = Math.abs(dy) > Math.abs(dx)

      if (requestedAxis === 'x' && isVertical) {
        // Disengage if user is scrolling vertically in an x-only swipe component
        isDragging = false
        opts.onDragEnd?.()
        return
      }
      if (requestedAxis === 'y' && !isVertical) {
        // Disengage if user is scrolling horizontally in a y-only swipe component
        isDragging = false
        opts.onDragEnd?.()
        return
      }

      hasMovedFar = true
      preventNextClick = true

      const extEvent = e as ExtendedPointerEvent
      extEvent._swipeHandledAxis = requestedAxis

      try {
        node.setPointerCapture(e.pointerId)
      } catch {
        // Ignore pointer capture errors
      }

      isVerticalScroll =
        requestedAxis === 'y' || (requestedAxis === 'both' && isVertical)
      if (isVerticalScroll && !scrollTarget) {
        scrollTarget = findScrollableContainer(node, e.target as HTMLElement)
      }
      isLockedToAxis = true
    }

    if (!isLockedToAxis) return

    if (isVerticalScroll && scrollTarget) {
      const isAtTop = scrollTarget.scrollTop <= 0 && stepY > 0
      const isAtBottom =
        scrollTarget.scrollTop >=
          scrollTarget.scrollHeight - scrollTarget.clientHeight && stepY < 0
      const effectiveStep = isAtTop || isAtBottom ? stepY * 0.35 : stepY

      scrollTarget.scrollTop -= effectiveStep
      opts.onDragMove?.({ x: 0, y: dy })
    } else if (requestedAxis === 'x') {
      opts.onDragMove?.({ x: dx, y: 0 })
    } else {
      opts.onDragMove?.({ x: dx, y: dy })
    }
  }

  function handlePointerUp(e: PointerEvent) {
    if (!isDragging) return
    isDragging = false

    const dx = e.clientX - startX
    const dy = e.clientY - startY
    const thresh = opts.threshold ?? 20
    const requestedAxis = opts.axis ?? 'both'

    const absDx = Math.abs(dx)
    const absDy = Math.abs(dy)

    if (absDx > absDy) {
      if (requestedAxis !== 'y') {
        if (dx < -thresh) {
          opts.onSwipeLeft?.()
        } else if (dx > thresh) {
          opts.onSwipeRight?.()
        }
      }
    } else {
      if (requestedAxis !== 'x') {
        if (dy < -thresh) {
          opts.onSwipeUp?.()
        } else if (dy > thresh) {
          opts.onSwipeDown?.()
        }
      }
    }

    opts.onDragEnd?.()

    if (hasMovedFar) {
      try {
        node.releasePointerCapture(e.pointerId)
      } catch {
        // Ignore pointer release errors
      }
    }

    if (preventNextClick) {
      clearClickTimeout = setTimeout(() => {
        preventNextClick = false
      }, 120)
    }

    // Launch momentum inertia if user flicked with noticeable velocity
    if (isVerticalScroll && scrollTarget && Math.abs(velocityY) > 0.12) {
      const target = scrollTarget
      let currentVelocity = velocityY
      let lastFrameTime = performance.now()

      const inertiaLoop = (frameNow: number) => {
        if (isDragging) return
        const frameDt = Math.min(32, frameNow - lastFrameTime)
        lastFrameTime = frameNow

        target.scrollTop -= currentVelocity * frameDt
        currentVelocity *= Math.pow(0.92, frameDt / 16.67)

        if (Math.abs(currentVelocity) > 0.04) {
          inertiaFrame = requestAnimationFrame(inertiaLoop)
        }
      }
      inertiaFrame = requestAnimationFrame(inertiaLoop)
    }

    hasMovedFar = false
    isVerticalScroll = false
    isLockedToAxis = false
    scrollTarget = null
    velocityY = 0
  }

  function handleClickCapture(e: MouseEvent) {
    if (preventNextClick) {
      e.stopPropagation()
      e.preventDefault()
      preventNextClick = false
      if (clearClickTimeout) {
        clearTimeout(clearClickTimeout)
        clearClickTimeout = null
      }
    }
  }

  node.addEventListener('pointerdown', handlePointerDown)
  node.addEventListener('pointermove', handlePointerMove)
  node.addEventListener('pointerup', handlePointerUp)
  node.addEventListener('pointercancel', handlePointerUp)
  node.addEventListener('click', handleClickCapture, true)

  return {
    update(newOptions: SwipeOptions) {
      opts = newOptions
    },
    destroy() {
      cancelAnimationFrame(inertiaFrame)
      if (clearClickTimeout) clearTimeout(clearClickTimeout)
      node.removeEventListener('pointerdown', handlePointerDown)
      node.removeEventListener('pointermove', handlePointerMove)
      node.removeEventListener('pointerup', handlePointerUp)
      node.removeEventListener('pointercancel', handlePointerUp)
      node.removeEventListener('click', handleClickCapture, true)
    }
  }
}
