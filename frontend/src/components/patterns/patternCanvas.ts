/**
 * @fileoverview Shared offscreen-cached canvas renderer for decorative background patterns.
 *
 * Replaces hundreds of DOM/SVG nodes and CSS mask-image passes with a single
 * cached offscreen bitmap per unique (kind, align, size, rows, cols, theme, dpr).
 */

export type PatternKind = 'capsule' | 'circle' | 'diamond'
export type PatternAlign = 'center' | 'right' | 'full'
export type PatternSize = 'default' | 'compact'

export interface PatternDrawOptions {
  kind: PatternKind
  align: PatternAlign
  size?: PatternSize
  rows: number
  cols: number
  theme: string
}

const OPACITY_VARIANTS_STANDARD = [0.4, 0.85, 0.65, 1, 0.4, 1, 0.85, 0.65]
const OPACITY_VARIANTS_DIAMOND = [0.45, 0.9, 0.7, 1, 0.45, 1, 0.9, 0.7]
const CAPSULE_WIDTHS = [14, 22, 30, 22, 14, 30]
const SKEW_TAN = Math.tan((-18 * Math.PI) / 180)
const SQRT1_2 = Math.SQRT1_2

const patternCache = new Map<string, HTMLCanvasElement>()
let lastCachedTheme = ''

function getPatternDimensions(opts: PatternDrawOptions): {
  width: number
  height: number
} {
  const { kind, size = 'default', rows, cols } = opts
  const isCompact = size === 'compact'

  if (kind === 'capsule') {
    const cellHeight = 9
    const rowGap = 5
    const colGap = 6
    let maxRowWidth = 0
    for (let r = 0; r < rows; r++) {
      let rw = 0
      for (let c = 0; c < cols; c++) {
        const w = CAPSULE_WIDTHS[(r * 2 + c * 5) % CAPSULE_WIDTHS.length] ?? 22
        rw += w + (c > 0 ? colGap : 0)
      }
      if (rw > maxRowWidth) maxRowWidth = rw
    }
    return {
      width: maxRowWidth + 32,
      height: rows > 0 ? rows * cellHeight + (rows - 1) * rowGap : 0
    }
  }

  if (kind === 'circle') {
    const diameter = isCompact ? 7.5 : 10
    const colGap = isCompact ? 6 : 8
    const rowGap = isCompact ? 4 : 6
    const offset = isCompact ? 7 : 10
    const rowWidth = cols > 0 ? cols * diameter + (cols - 1) * colGap : 0
    return {
      width: rowWidth + offset * 2 + 8,
      height: rows > 0 ? rows * diameter + (rows - 1) * rowGap + 8 : 0
    }
  }

  const cellSize = isCompact ? 9.5 : 12
  const colGap = isCompact ? 6 : 8
  const rowStep = isCompact ? 7.5 : 9
  const rowOffset = isCompact ? 8 : 10
  const rowWidth = cols > 0 ? cols * cellSize + (cols - 1) * colGap : 0
  return {
    width: rowWidth + rowOffset * 2 + 8,
    height: rows > 0 ? (rows - 1) * rowStep + cellSize + 8 : 0
  }
}

function drawRoundedRect(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  r: number
) {
  if (typeof ctx.roundRect === 'function') {
    ctx.beginPath()
    ctx.roundRect(x, y, w, h, r)
    ctx.fill()
    return
  }
  const rad = Math.min(r, w / 2, h / 2)
  ctx.beginPath()
  ctx.moveTo(x + rad, y)
  ctx.lineTo(x + w - rad, y)
  ctx.quadraticCurveTo(x + w, y, x + w, y + rad)
  ctx.lineTo(x + w, y + h - rad)
  ctx.quadraticCurveTo(x + w, y + h, x + w - rad, y + h)
  ctx.lineTo(x + rad, y + h)
  ctx.quadraticCurveTo(x, y + h, x, y + h - rad)
  ctx.lineTo(x, y + rad)
  ctx.quadraticCurveTo(x, y, x + rad, y)
  ctx.closePath()
  ctx.fill()
}

function renderPatternShapes(
  ctx: CanvasRenderingContext2D,
  opts: PatternDrawOptions,
  width: number,
  darkColor: string,
  lightColor: string
) {
  const { kind, size = 'default', rows, cols } = opts
  const isCompact = size === 'compact'

  if (kind === 'capsule') {
    const cellHeight = 9
    const rowGap = 5
    const colGap = 6

    for (let r = 0; r < rows; r++) {
      let rw = 0
      const widths: number[] = []
      for (let c = 0; c < cols; c++) {
        const w = CAPSULE_WIDTHS[(r * 2 + c * 5) % CAPSULE_WIDTHS.length] ?? 22
        widths.push(w)
        rw += w + (c > 0 ? colGap : 0)
      }
      const offsetX = r % 2 !== 0 ? 12 : 0
      let curX = (width - rw) / 2 + offsetX
      const y = r * (cellHeight + rowGap)
      const cy = y + cellHeight / 2

      for (let c = 0; c < cols; c++) {
        const w = widths[c]
        const cx = curX + w / 2
        const isLight = (r * 3 + c) % 2 === 0
        const opacity =
          OPACITY_VARIANTS_STANDARD[
            (r * 5 + c * 3) % OPACITY_VARIANTS_STANDARD.length
          ] ?? 0.65

        ctx.save()
        ctx.globalAlpha = opacity
        ctx.fillStyle = isLight ? lightColor : darkColor
        ctx.translate(cx, cy)
        ctx.transform(1, 0, SKEW_TAN, 1, 0, 0)
        ctx.translate(-cx, -cy)
        drawRoundedRect(ctx, curX, y, w, cellHeight, 4.5)
        ctx.restore()

        curX += w + colGap
      }
    }
    return
  }

  if (kind === 'circle') {
    const diameter = isCompact ? 7.5 : 10
    const radius = diameter / 2
    const colGap = isCompact ? 6 : 8
    const rowGap = isCompact ? 4 : 6
    const offset = isCompact ? 7 : 10
    const rowWidth = cols > 0 ? cols * diameter + (cols - 1) * colGap : 0
    const startXBase = (width - rowWidth) / 2
    const startY = 4

    for (let r = 0; r < rows; r++) {
      const rowOffset = r % 2 !== 0 ? offset : 0
      const cy = startY + r * (diameter + rowGap) + radius
      for (let c = 0; c < cols; c++) {
        const cx = startXBase + rowOffset + c * (diameter + colGap) + radius
        const isLight = (r * 3 + c) % 2 === 0
        const opacity =
          OPACITY_VARIANTS_STANDARD[
            (r * 5 + c * 3) % OPACITY_VARIANTS_STANDARD.length
          ] ?? 0.65

        ctx.globalAlpha = opacity
        ctx.fillStyle = isLight ? lightColor : darkColor
        ctx.beginPath()
        ctx.arc(cx, cy, radius, 0, Math.PI * 2)
        ctx.fill()
      }
    }
    ctx.globalAlpha = 1
    return
  }

  const cellSize = isCompact ? 9.5 : 12
  const colGap = isCompact ? 6 : 8
  const rowStep = isCompact ? 7.5 : 9
  const rowOffset = isCompact ? 8 : 10
  const radius = isCompact ? 1.2 : 1.5
  const rowWidth = cols > 0 ? cols * cellSize + (cols - 1) * colGap : 0
  const startXBase = (width - rowWidth) / 2
  const startY = 4

  for (let r = 0; r < rows; r++) {
    const offsetX = r % 2 !== 0 ? rowOffset : 0
    const y = startY + r * rowStep
    const cy = y + cellSize / 2
    for (let d = 0; d < cols; d++) {
      const x = startXBase + offsetX + d * (cellSize + colGap)
      const cx = x + cellSize / 2
      const isLight = (r * 3 + d) % 2 === 0
      const opacity =
        OPACITY_VARIANTS_DIAMOND[
          (r * 5 + d * 3) % OPACITY_VARIANTS_DIAMOND.length
        ] ?? 0.7

      ctx.save()
      ctx.globalAlpha = opacity
      ctx.fillStyle = isLight ? lightColor : darkColor
      ctx.translate(cx, cy)
      ctx.transform(SQRT1_2, SQRT1_2, -SQRT1_2, SQRT1_2, 0, 0)
      ctx.translate(-cx, -cy)
      drawRoundedRect(ctx, x, y, cellSize, cellSize, radius)
      ctx.restore()
    }
  }
}

function applyAlignMask(
  ctx: CanvasRenderingContext2D,
  align: PatternAlign,
  width: number,
  height: number
) {
  if (align === 'full' || width <= 0 || height <= 0) return

  const grad = ctx.createLinearGradient(0, 0, width, 0)
  if (align === 'center') {
    grad.addColorStop(0, 'rgba(0, 0, 0, 0.08)')
    grad.addColorStop(0.28, 'rgba(0, 0, 0, 0.55)')
    grad.addColorStop(0.55, 'rgba(0, 0, 0, 0.95)')
    grad.addColorStop(0.85, 'rgba(0, 0, 0, 0.35)')
    grad.addColorStop(1, 'rgba(0, 0, 0, 0.1)')
  } else {
    grad.addColorStop(0, 'rgba(0, 0, 0, 0)')
    grad.addColorStop(0.15, 'rgba(0, 0, 0, 0)')
    grad.addColorStop(0.38, 'rgba(0, 0, 0, 0.12)')
    grad.addColorStop(0.62, 'rgba(0, 0, 0, 0.5)')
    grad.addColorStop(0.86, 'rgba(0, 0, 0, 0.92)')
    grad.addColorStop(1, 'rgba(0, 0, 0, 0.75)')
  }

  ctx.save()
  ctx.globalCompositeOperation = 'destination-in'
  ctx.fillStyle = grad
  ctx.fillRect(0, 0, width, height)
  ctx.restore()
}

/**
 * Draws a cached decorative pattern onto the target canvas, centering the
 * baked pattern bitmap and applying the alignment fade in-canvas.
 */
export function drawPatternToCanvas(
  canvas: HTMLCanvasElement,
  opts: PatternDrawOptions
) {
  const rect = canvas.getBoundingClientRect()
  if (rect.width <= 0 || rect.height <= 0) return

  if (opts.theme !== lastCachedTheme) {
    patternCache.clear()
    lastCachedTheme = opts.theme
  }

  const dpr = Math.min(window.devicePixelRatio || 1, 2)
  const targetW = Math.max(1, Math.round(rect.width * dpr))
  const targetH = Math.max(1, Math.round(rect.height * dpr))
  if (canvas.width !== targetW) canvas.width = targetW
  if (canvas.height !== targetH) canvas.height = targetH

  const size = opts.size ?? 'default'
  const cacheKey = `${opts.kind}_${size}_${opts.rows}_${opts.cols}_${opts.theme}_${dpr}`
  let baked = patternCache.get(cacheKey)
  const dims = getPatternDimensions(opts)

  if (!baked && dims.width > 0 && dims.height > 0) {
    const style = getComputedStyle(canvas)
    const fallbackDark =
      opts.kind === 'diamond' ? 'rgba(0, 0, 0, 0.24)' : 'rgba(0, 0, 0, 0.22)'
    const fallbackLight =
      opts.kind === 'diamond'
        ? 'rgba(255, 255, 255, 0.24)'
        : 'rgba(255, 255, 255, 0.22)'
    const darkColor =
      style.getPropertyValue('--vinta-agenda-diamond-dark').trim() ||
      fallbackDark
    const lightColor =
      style.getPropertyValue('--vinta-agenda-diamond-light').trim() ||
      fallbackLight

    baked = document.createElement('canvas')
    baked.width = Math.max(1, Math.round(dims.width * dpr))
    baked.height = Math.max(1, Math.round(dims.height * dpr))
    const bCtx = baked.getContext('2d')
    if (bCtx) {
      bCtx.setTransform(dpr, 0, 0, dpr, 0, 0)
      renderPatternShapes(bCtx, opts, dims.width, darkColor, lightColor)
      patternCache.set(cacheKey, baked)
    }
  }

  const ctx = canvas.getContext('2d')
  if (!ctx) return

  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  ctx.clearRect(0, 0, rect.width, rect.height)

  if (baked) {
    const dx = (rect.width - dims.width) / 2
    const dy = (rect.height - dims.height) / 2
    ctx.drawImage(baked, dx, dy, dims.width, dims.height)
    applyAlignMask(ctx, opts.align, rect.width, rect.height)
  }
}
