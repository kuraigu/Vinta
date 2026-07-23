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

export function swipe(node: HTMLElement, options: SwipeOptions = {}) {
  let opts = options
  let isDragging = false
  let hasMovedFar = false
  let startX = 0
  let startY = 0

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
    startX = e.clientX
    startY = e.clientY
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
      try {
        node.setPointerCapture(e.pointerId)
      } catch {
        // Ignore pointer capture errors if unsupported
      }
    }

    opts.onDragMove?.({ x: dx, y: dy })
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
        // Ignore if capture already lost
      }
    }
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
