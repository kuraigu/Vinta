<script lang="ts">
  interface Props {
    align?: 'center' | 'right' | 'full'
    size?: 'default' | 'compact'
    rows?: number
    cols?: number
    class?: string
  }

  let {
    align = 'right',
    size = 'default',
    rows = 10,
    cols = 20,
    class: className = ''
  }: Props = $props()

  const OPACITY_VARIANTS = ['v1', 'v3', 'v2', 'v4', 'v1', 'v4', 'v3', 'v2']

  /**
   * Generates row indices for the circle pattern grid.
   *
   * @returns Array of row indices.
   */
  function computeRowIndices(): number[] {
    return Array.from({ length: rows }, (_, i) => i)
  }

  /**
   * Generates column indices for the circle pattern grid.
   *
   * @returns Array of column indices.
   */
  function computeColIndices(): number[] {
    return Array.from({ length: cols }, (_, i) => i)
  }

  let rowIndices = $derived.by(computeRowIndices)
  let colIndices = $derived.by(computeColIndices)
</script>

<div
  class="vinta-circle-pattern align-{align} size-{size} {className}"
  aria-hidden="true"
>
  {#each rowIndices as r (r)}
    <div class="circle-row" class:offset={r % 2 !== 0}>
      {#each colIndices as c (c)}
        {@const isLight = (r * 3 + c) % 2 === 0}
        {@const variant =
          OPACITY_VARIANTS[(r * 5 + c * 3) % OPACITY_VARIANTS.length]}
        <span class="circle-cell {isLight ? 'light' : 'dark'} {variant}"></span>
      {/each}
    </div>
  {/each}
</div>

<style>
  .vinta-circle-pattern {
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

  .vinta-circle-pattern.align-center {
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

  .vinta-circle-pattern.align-right {
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

  .circle-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 3px 0;
    flex-shrink: 0;
  }

  .circle-row.offset {
    transform: translateX(10px);
  }

  .circle-cell {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .vinta-circle-pattern.size-compact .circle-row {
    gap: 6px;
    margin: 2px 0;
  }

  .vinta-circle-pattern.size-compact .circle-row.offset {
    transform: translateX(7px);
  }

  .vinta-circle-pattern.size-compact .circle-cell {
    width: 7.5px;
    height: 7.5px;
  }

  .circle-cell.dark {
    background: var(--vinta-agenda-diamond-dark, rgba(0, 0, 0, 0.22));
  }

  .circle-cell.light {
    background: var(--vinta-agenda-diamond-light, rgba(255, 255, 255, 0.22));
  }

  .circle-cell.v1 {
    opacity: 0.4;
  }

  .circle-cell.v2 {
    opacity: 0.65;
  }

  .circle-cell.v3 {
    opacity: 0.85;
  }

  .circle-cell.v4 {
    opacity: 1;
  }
</style>
