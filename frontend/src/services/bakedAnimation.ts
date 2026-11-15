/**
 * @fileoverview Reusable baked-animation utilities for ambient infinite loops.
 *
 * Pre-renders expensive SVG/gradient artwork onto offscreen bitmaps once and
 * drives updates through a frame-capped loop that pauses when hidden or off-screen.
 */

export interface BakedLoopOptions {
  /** Target frames per second (defaults to 18 FPS for low-overhead ambient motion). */
  fps?: number
  /** Predicate returning true when the animation should actively tick. */
  isActive?: () => boolean
  /** Callback invoked on each throttled frame with elapsed seconds. */
  onFrame: (elapsedSec: number) => void
}

export interface BakedLoopController {
  start: () => void
  stop: () => void
  sync: () => void
}

/**
 * Creates a throttled requestAnimationFrame loop that automatically pauses
 * when `document.hidden` is true or `isActive()` returns false.
 */
export function createBakedLoop(opts: BakedLoopOptions): BakedLoopController {
  const fps = Math.max(1, opts.fps ?? 18)
  const frameIntervalMs = 1000 / fps
  let rafId = 0
  let startTime = 0
  let lastFrameTime = 0
  let disposed = false

  function shouldRun(): boolean {
    if (disposed || document.hidden) return false
    return opts.isActive ? opts.isActive() : true
  }

  function tick(now: number) {
    rafId = 0
    if (!shouldRun()) return

    if (!startTime) startTime = now
    if (now - lastFrameTime >= frameIntervalMs) {
      lastFrameTime = now
      opts.onFrame((now - startTime) / 1000)
    }
    rafId = requestAnimationFrame(tick)
  }

  function start() {
    if (disposed || rafId || !shouldRun()) return
    lastFrameTime = 0
    rafId = requestAnimationFrame(tick)
  }

  function stop() {
    if (rafId) {
      cancelAnimationFrame(rafId)
      rafId = 0
    }
  }

  function sync() {
    if (shouldRun()) {
      start()
    } else {
      stop()
    }
  }

  function handleVisibilityChange() {
    sync()
  }

  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', handleVisibilityChange)
  }

  return {
    start,
    stop: () => {
      disposed = true
      stop()
      if (typeof document !== 'undefined') {
        document.removeEventListener('visibilitychange', handleVisibilityChange)
      }
    },
    sync
  }
}

/**
 * Rasterizes an image or SVG URL onto an offscreen canvas at the target DPR
 * so expensive SVG filters and gradients are evaluated only once.
 */
export function bakeImageToCanvas(
  src: string,
  width: number,
  height: number,
  dpr = 1
): Promise<HTMLCanvasElement | null> {
  return new Promise((resolve) => {
    if (!src || width <= 0 || height <= 0) {
      resolve(null)
      return
    }
    const img = new Image()
    img.decoding = 'async'
    img.onload = () => {
      const canvas = document.createElement('canvas')
      canvas.width = Math.max(1, Math.round(width * dpr))
      canvas.height = Math.max(1, Math.round(height * dpr))
      const ctx = canvas.getContext('2d')
      if (!ctx) {
        resolve(null)
        return
      }
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
      ctx.drawImage(img, 0, 0, width, height)
      resolve(canvas)
    }
    img.onerror = () => resolve(null)
    img.src = src
  })
}
