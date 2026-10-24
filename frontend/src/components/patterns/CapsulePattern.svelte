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

  /**
   * Generates row indices for the capsule mosaic pattern grid.
   *
   * @returns Array of row indices.
   */
  function computeRowIndices(): number[] {
    return Array.from({ length: rows }, (_, i) => i)
  }

  /**
   * Generates column indices for the capsule mosaic pattern grid.
   *
   * @returns Array of column indices.
   */
  function computeColIndices(): number[] {
    return Array.from({ length: cols }, (_, i) => i)
  }

  let rowIndices = $derived.by(computeRowIndices)
  let colIndices = $derived.by(computeColIndices)
</script>

<div class="vinta-capsule-pattern align-{align} {className}" aria-hidden="true">
  {#each rowIndices as r (r)}
    <div class="capsule-row" class:offset={r % 2 !== 0}>
      {#each colIndices as c (c)}
        {@const isLight = (r * 3 + c) % 2 === 0}
        {@const variant =
          OPACITY_VARIANTS[(r * 5 + c * 3) % OPACITY_VARIANTS.length]}
        {@const widthVar =
          WIDTH_VARIANTS[(r * 2 + c * 5) % WIDTH_VARIANTS.length]}
        <span
          class="capsule-cell {isLight ? 'light' : 'dark'} {variant} {widthVar}"
        ></span>
      {/each}
    </div>
  {/each}
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
    gap: 5px;
    overflow: hidden;
    pointer-events: none;
    z-index: 0;
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

  .capsule-row {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  .capsule-row.offset {
    transform: translateX(12px);
  }

  .capsule-cell {
    height: 9px;
    border-radius: 9999px;
    transform: skewX(-18deg);
    flex-shrink: 0;
  }

  .capsule-cell.w-sm {
    width: 14px;
  }

  .capsule-cell.w-md {
    width: 22px;
  }

  .capsule-cell.w-lg {
    width: 30px;
  }

  .capsule-cell.dark {
    background: var(--vinta-agenda-diamond-dark, rgba(0, 0, 0, 0.22));
  }

  .capsule-cell.light {
    background: var(--vinta-agenda-diamond-light, rgba(255, 255, 255, 0.22));
  }

  .capsule-cell.v1 {
    opacity: 0.4;
  }

  .capsule-cell.v2 {
    opacity: 0.65;
  }

  .capsule-cell.v3 {
    opacity: 0.85;
  }

  .capsule-cell.v4 {
    opacity: 1;
  }
</style>
