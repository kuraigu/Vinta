<!--
  @component DiamondPattern
  @description Decorative background SVG grid of rotated diamond shapes.
-->
<script lang="ts">
  interface Props {
    align?: 'center' | 'right' | 'full'
    size?: 'default' | 'compact'
    rows?: number
    cols?: number
    class?: string
  }

  let {
    align = 'center',
    size = 'default',
    rows = 12,
    cols = 30,
    class: className = ''
  }: Props = $props()

  const OPACITY_VARIANTS = ['v1', 'v3', 'v2', 'v4', 'v1', 'v4', 'v3', 'v2']

  const OPACITY_MAP: Record<string, number> = {
    v1: 0.45,
    v2: 0.7,
    v3: 0.9,
    v4: 1
  }

  interface DiamondRect {
    x: number
    y: number
    cx: number
    cy: number
    isLight: boolean
    opacity: number
  }

  interface DiamondGridGeometry {
    width: number
    height: number
    cellSize: number
    radius: number
    rects: DiamondRect[]
  }

  /**
   * Computes deterministic SVG coordinates for all diamonds in the grid.
   *
   * @returns SVG canvas dimensions and positioned diamond descriptors.
   */
  function computeDiamondGrid(): DiamondGridGeometry {
    const isCompact = size === 'compact'
    const cellSize = isCompact ? 9.5 : 12
    const colGap = isCompact ? 6 : 8
    const rowStep = isCompact ? 7.5 : 9
    const rowOffset = isCompact ? 8 : 10
    const radius = isCompact ? 1.2 : 1.5

    const rowWidth = cols > 0 ? cols * cellSize + (cols - 1) * colGap : 0
    const svgWidth = rowWidth + rowOffset * 2 + 8
    const svgHeight = rows > 0 ? (rows - 1) * rowStep + cellSize + 8 : 0
    const startXBase = (svgWidth - rowWidth) / 2
    const startY = 4

    const rects: DiamondRect[] = []
    for (let r = 0; r < rows; r++) {
      const offsetX = r % 2 !== 0 ? rowOffset : 0
      const y = startY + r * rowStep
      const cy = y + cellSize / 2
      for (let d = 0; d < cols; d++) {
        const x = startXBase + offsetX + d * (cellSize + colGap)
        const cx = x + cellSize / 2
        const isLight = (r * 3 + d) % 2 === 0
        const variant =
          OPACITY_VARIANTS[(r * 5 + d * 3) % OPACITY_VARIANTS.length]
        const opacity = OPACITY_MAP[variant] ?? 0.7
        rects.push({ x, y, cx, cy, isLight, opacity })
      }
    }

    return {
      width: svgWidth,
      height: svgHeight,
      cellSize,
      radius,
      rects
    }
  }

  let grid = $derived.by(computeDiamondGrid)
</script>

<div
  class="vinta-diamond-pattern align-{align} size-{size} {className}"
  aria-hidden="true"
>
  <svg
    class="diamond-pattern-svg"
    width={grid.width}
    height={grid.height}
    viewBox="0 0 {grid.width} {grid.height}"
  >
    {#each grid.rects as rect, i (i)}
      <rect
        x={rect.x}
        y={rect.y}
        width={grid.cellSize}
        height={grid.cellSize}
        rx={grid.radius}
        ry={grid.radius}
        opacity={rect.opacity}
        transform="rotate(45 {rect.cx} {rect.cy})"
        class={rect.isLight ? 'diamond-light' : 'diamond-dark'}
      />
    {/each}
  </svg>
</div>

<style>
  .vinta-diamond-pattern {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    overflow: hidden;
    pointer-events: none;
    z-index: 0;
  }

  .diamond-pattern-svg {
    flex-shrink: 0;
    overflow: visible;
  }

  .vinta-diamond-pattern.align-center {
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

  .vinta-diamond-pattern.align-right {
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

  .diamond-dark {
    fill: var(--vinta-agenda-diamond-dark, rgba(0, 0, 0, 0.24));
  }

  .diamond-light {
    fill: var(--vinta-agenda-diamond-light, rgba(255, 255, 255, 0.24));
  }
</style>
