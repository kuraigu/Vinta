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

function findScrollableParent(el: HTMLElement | null): HTMLElement | null {
  let current = el
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
  return null
}

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

  let velocityY = 0
  let lastTime = 0
  let inertiaFrame = 0

  function handlePointerDown(e: PointerEvent) {
    const extEvent = e as ExtendedPointerEvent
    const requestedAxis = opts.axis ?? 'both'

    if (
      extEvent._swipeHandledAxis === 'both' ||
      extEvent._swipeHandledAxis === requestedAxis
    )
      return
    if (opts.enableSwipe === false) return
    const defaultIgnore =
      'select, option, input, textarea, range, .custom-select, .custom-select-trigger, .custom-select-options, .custom-select-option, .no-swipe'
    const fullIgnoreSelector = opts.ignoreSelector
      ? `${defaultIgnore}, ${opts.ignoreSelector}`
      : defaultIgnore

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
    scrollTarget = null
    startX = e.clientX
    startY = e.clientY
    lastY = e.clientY
    velocityY = 0
    lastTime = performance.now()
    opts.onDragStart?.()
  }

  function handlePointerMove(e: PointerEvent) {
    if (!isDragging) return
    const dx = e.clientX - startX
    const dy = e.clientY - startY
    const dragThresh = opts.dragThreshold ?? 6
    const requestedAxis = opts.axis ?? 'both'
    const now = performance.now()
    const dt = Math.max(1, now - lastTime)

    const stepY = e.clientY - lastY
    velocityY = stepY / dt
    lastY = e.clientY
    lastTime = now

    if (
      !hasMovedFar &&
      (Math.abs(dx) > dragThresh || Math.abs(dy) > dragThresh)
    ) {
      const isVertical = Math.abs(dy) > Math.abs(dx)

      if (requestedAxis === 'x' && isVertical) {
        isDragging = false
        opts.onDragEnd?.()
        return
      }
      if (requestedAxis === 'y' && !isVertical) {
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

      if (isVertical) {
        isVerticalScroll = true
        scrollTarget = findScrollableParent(e.target as HTMLElement)
      } else {
        isVerticalScroll = false
      }
    }

    if (!isLockedToAxis && hasMovedFar) {
      isLockedToAxis = true
    }

    if (!isLockedToAxis) return

    if (isVerticalScroll) {
      if (scrollTarget) {
        scrollTarget.scrollTop -= stepY
      }
      opts.onDragMove?.({ x: 0, y: dy })
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
      }, 300)
    }

    if (isVerticalScroll && scrollTarget && Math.abs(velocityY) > 0.1) {
      let v = velocityY
      const target = scrollTarget
      function loop() {
        if (Math.abs(v) < 0.05 || isDragging) return
        target.scrollTop -= v * 16
        v *= 0.92 // friction multiplier
        inertiaFrame = requestAnimationFrame(loop)
      }
      inertiaFrame = requestAnimationFrame(loop)
    }

    hasMovedFar = false
    isVerticalScroll = false
    isLockedToAxis = false
    scrollTarget = null
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
