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

function findScrollableParent(
  el: HTMLElement | null,
  stopNode: HTMLElement
): HTMLElement | null {
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
    if (current === stopNode) break
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

  function handlePointerDown(e: PointerEvent) {
    if (opts.enableSwipe === false) return
    if (
      opts.ignoreSelector &&
      (e.target as HTMLElement).closest(opts.ignoreSelector)
    )
      return
    if (e.button !== 0 && e.pointerType === 'mouse') return

    isDragging = true
    hasMovedFar = false
    isVerticalScroll = false
    scrollTarget = null
    startX = e.clientX
    startY = e.clientY
    lastY = e.clientY
    opts.onDragStart?.()
  }

  function handlePointerMove(e: PointerEvent) {
    if (!isDragging) return
    const dx = e.clientX - startX
    const dy = e.clientY - startY
    const dragThresh = opts.dragThreshold ?? 8

    if (
      !hasMovedFar &&
      (Math.abs(dx) > dragThresh || Math.abs(dy) > dragThresh)
    ) {
      hasMovedFar = true

      // Detect if drag direction is primarily vertical
      if (Math.abs(dy) > Math.abs(dx)) {
        isVerticalScroll = true
        scrollTarget = findScrollableParent(e.target as HTMLElement, node)
      } else {
        isVerticalScroll = false
        try {
          node.setPointerCapture(e.pointerId)
        } catch {
          // Ignore pointer capture errors if unsupported
        }
      }
    }

    if (isVerticalScroll) {
      if (scrollTarget) {
        const stepY = e.clientY - lastY
        scrollTarget.scrollTop -= stepY
      }
      lastY = e.clientY
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

    if (hasMovedFar && !isVerticalScroll) {
      try {
        node.releasePointerCapture(e.pointerId)
      } catch {
        // Ignore if capture already lost
      }
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
      node.removeEventListener('pointerdown', handlePointerDown)
      node.removeEventListener('pointermove', handlePointerMove)
      node.removeEventListener('pointerup', handlePointerUp)
      node.removeEventListener('pointercancel', handlePointerUp)
      node.removeEventListener('click', handleClickCapture, true)
    }
  }
}
