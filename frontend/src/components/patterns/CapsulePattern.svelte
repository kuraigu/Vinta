<!--
  @component CapsulePattern
  @description Decorative background SVG grid of subtle pill shapes used in themed card backgrounds.
-->
<script lang="ts">
  interface Props {
    align?: 'center' | 'right' | 'full'
    rows?: number
    cols?: number
    class?: string
  }

  let {
    align = 'right',
    rows = 12,
    cols = 20,
    class: className = ''
  }: Props = $props()

  const OPACITY_VARIANTS = ['v1', 'v3', 'v2', 'v4', 'v1', 'v4', 'v3', 'v2']
  const WIDTH_VARIANTS = ['w-sm', 'w-md', 'w-lg', 'w-md', 'w-sm', 'w-lg']

  const OPACITY_MAP: Record<string, number> = {
    v1: 0.4,
    v2: 0.65,
    v3: 0.85,
    v4: 1
  }
  const WIDTH_MAP: Record<string, number> = {
    'w-sm': 14,
    'w-md': 22,
    'w-lg': 30
  }

  interface CapsuleRect {
    x: number
    y: number
    w: number
    cx: number
    cy: number
    isLight: boolean
    opacity: number
  }

  interface CapsuleGridGeometry {
    width: number
    height: number
    rects: CapsuleRect[]
  }

  /**
   * Computes deterministic SVG coordinates for all capsules in the mosaic grid.
   *
   * @returns SVG canvas dimensions and positioned capsule descriptors.
   */
  function computeCapsuleGrid(): CapsuleGridGeometry {
    const cellHeight = 9
    const rowGap = 5
    const colGap = 6
    const rowWidths: number[] = []
    const rowCells: {
      w: number
      isLight: boolean
      opacity: number
    }[][] = []

    let maxRowWidth = 0
    for (let r = 0; r < rows; r++) {
      const cells: { w: number; isLight: boolean; opacity: number }[] = []
      let rw = 0
      for (let c = 0; c < cols; c++) {
        const isLight = (r * 3 + c) % 2 === 0
        const variant =
          OPACITY_VARIANTS[(r * 5 + c * 3) % OPACITY_VARIANTS.length]
        const widthVar = WIDTH_VARIANTS[(r * 2 + c * 5) % WIDTH_VARIANTS.length]
        const w = WIDTH_MAP[widthVar] ?? 22
        const opacity = OPACITY_MAP[variant] ?? 0.65
        cells.push({ w, isLight, opacity })
        rw += w + (c > 0 ? colGap : 0)
      }
      rowCells.push(cells)
      rowWidths.push(rw)
      if (rw > maxRowWidth) maxRowWidth = rw
    }

    const svgWidth = maxRowWidth + 32
    const svgHeight = rows > 0 ? rows * cellHeight + (rows - 1) * rowGap : 0
    const rects: CapsuleRect[] = []

    for (let r = 0; r < rows; r++) {
      const rw = rowWidths[r]
      const offsetX = r % 2 !== 0 ? 12 : 0
      let curX = (svgWidth - rw) / 2 + offsetX
      const y = r * (cellHeight + rowGap)
      const cy = y + cellHeight / 2

      for (const cell of rowCells[r]) {
        const cx = curX + cell.w / 2
        rects.push({
          x: curX,
          y,
          w: cell.w,
          cx,
          cy,
          isLight: cell.isLight,
          opacity: cell.opacity
        })
        curX += cell.w + colGap
      }
    }

    return { width: svgWidth, height: svgHeight, rects }
  }

  let grid = $derived.by(computeCapsuleGrid)
</script>

<div class="vinta-capsule-pattern align-{align} {className}" aria-hidden="true">
  <svg
    class="capsule-pattern-svg"
    width={grid.width}
    height={grid.height}
    viewBox="0 0 {grid.width} {grid.height}"
  >
    {#each grid.rects as rect, i (i)}
      <rect
        x={rect.x}
        y={rect.y}
        width={rect.w}
        height="9"
        rx="4.5"
        ry="4.5"
        opacity={rect.opacity}
        transform="translate({rect.cx} {rect.cy}) skewX(-18) translate({-rect.cx} {-rect.cy})"
        class={rect.isLight ? 'capsule-light' : 'capsule-dark'}
      />
    {/each}
  </svg>
</div>

<style>
  .vinta-capsule-pattern {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    overflow: hidden;
    pointer-events: none;
    z-index: 0;
  }

  .capsule-pattern-svg {
    flex-shrink: 0;
    overflow: visible;
  }

  .vinta-capsule-pattern.align-right {
    mask-image: linear-gradient(
      90deg,
      transparent 15%,
      rgba(0, 0, 0, 0.12) 38%,
      rgba(0, 0, 0, 0.5) 62%,
      rgba(0, 0, 0, 0.92) 86%,
      rgba(0, 0, 0, 0.75) 100%
    );
    -webkit-mask-image: linear-gradient(
      90deg,
      transparent 15%,
      rgba(0, 0, 0, 0.12) 38%,
      rgba(0, 0, 0, 0.5) 62%,
      rgba(0, 0, 0, 0.92) 86%,
      rgba(0, 0, 0, 0.75) 100%
    );
  }

  .vinta-capsule-pattern.align-center {
    mask-image: linear-gradient(
      90deg,
      rgba(0, 0, 0, 0.08) 0%,
      rgba(0, 0, 0, 0.55) 28%,
      rgba(0, 0, 0, 0.95) 55%,
      rgba(0, 0, 0, 0.35) 85%,
      rgba(0, 0, 0, 0.1) 100%
    );
    -webkit-mask-image: linear-gradient(
      90deg,
      rgba(0, 0, 0, 0.08) 0%,
      rgba(0, 0, 0, 0.55) 28%,
      rgba(0, 0, 0, 0.95) 55%,
      rgba(0, 0, 0, 0.35) 85%,
      rgba(0, 0, 0, 0.1) 100%
    );
  }

  .capsule-dark {
    fill: var(--vinta-agenda-diamond-dark, rgba(0, 0, 0, 0.22));
  }

  .capsule-light {
    fill: var(--vinta-agenda-diamond-light, rgba(255, 255, 255, 0.22));
  }
</style>
