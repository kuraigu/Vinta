export type SwipeOptions = {
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
  _swipeHandled?: boolean
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
  let isVerticalScroll = false
  let scrollTarget: HTMLElement | null = null
  let startX = 0
  let startY = 0
  let lastY = 0

  let velocityY = 0
  let lastTime = 0
  let inertiaFrame = 0

  function handlePointerDown(e: PointerEvent) {
    const extEvent = e as ExtendedPointerEvent
    if (extEvent._swipeHandled) return
    if (opts.enableSwipe === false) return
    if (
      opts.ignoreSelector &&
      (e.target as HTMLElement).closest(opts.ignoreSelector)
    )
      return
    if (e.button !== 0 && e.pointerType === 'mouse') return

    extEvent._swipeHandled = true

    cancelAnimationFrame(inertiaFrame)
    isDragging = true
    hasMovedFar = false
    isVerticalScroll = false
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
    const dragThresh = opts.dragThreshold ?? 8
    const now = performance.now()
    const dt = Math.max(1, now - lastTime)

    if (
      !hasMovedFar &&
      (Math.abs(dx) > dragThresh || Math.abs(dy) > dragThresh)
    ) {
      hasMovedFar = true

      // Unconditionally capture pointer so we don't trigger hover states on other elements
      try {
        node.setPointerCapture(e.pointerId)
      } catch {
        // Ignore pointer capture errors if already released
      }

      if (Math.abs(dy) > Math.abs(dx)) {
        isVerticalScroll = true
        scrollTarget = findScrollableParent(e.target as HTMLElement)
      } else {
        isVerticalScroll = false
      }
    }

    if (isVerticalScroll) {
      if (scrollTarget) {
        const stepY = e.clientY - lastY
        scrollTarget.scrollTop -= stepY
        velocityY = stepY / dt
      }
      lastY = e.clientY
      lastTime = now
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
    const thresh = opts.threshold ?? 40

    if (Math.abs(dx) > Math.abs(dy)) {
      if (dx < -thresh) {
        opts.onSwipeLeft?.()
      } else if (dx > thresh) {
        opts.onSwipeRight?.()
      }
    } else {
      if (dy < -thresh) {
        opts.onSwipeUp?.()
      } else if (dy > thresh) {
        opts.onSwipeDown?.()
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
    scrollTarget = null
  }

  function handleClickCapture(e: MouseEvent) {
    if (hasMovedFar) {
      e.stopPropagation()
      e.preventDefault()
      hasMovedFar = false
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
      node.removeEventListener('pointerdown', handlePointerDown)
      node.removeEventListener('pointermove', handlePointerMove)
      node.removeEventListener('pointerup', handlePointerUp)
      node.removeEventListener('pointercancel', handlePointerUp)
      node.removeEventListener('click', handleClickCapture, true)
    }
  }
}
