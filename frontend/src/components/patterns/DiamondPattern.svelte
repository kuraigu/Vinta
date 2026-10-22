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

  /**
   * Generates row indices for the diamond pattern grid.
   *
   * @returns Array of row indices.
   */
  function computeRowIndices(): number[] {
    return Array.from({ length: rows }, (_, i) => i)
  }

  /**
   * Generates column indices for the diamond pattern grid.
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
  class="vinta-diamond-pattern align-{align} size-{size} {className}"
  aria-hidden="true"
>
  {#each rowIndices as r (r)}
    <div class="diamond-row" class:offset={r % 2 !== 0}>
      {#each colIndices as d (d)}
        {@const isLight = (r * 3 + d) % 2 === 0}
        {@const variant =
          OPACITY_VARIANTS[(r * 5 + d * 3) % OPACITY_VARIANTS.length]}
        <span class="diamond-cell {isLight ? 'light' : 'dark'} {variant}"
        ></span>
      {/each}
    </div>
  {/each}
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

  .diamond-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: -1.5px 0;
    flex-shrink: 0;
  }

  .diamond-row.offset {
    transform: translateX(10px);
  }

  .diamond-cell {
    width: 12px;
    height: 12px;
    border-radius: 1.5px;
    transform: rotate(45deg);
    flex-shrink: 0;
  }

  .vinta-diamond-pattern.size-compact .diamond-row {
    gap: 6px;
    margin: -1px 0;
  }

  .vinta-diamond-pattern.size-compact .diamond-row.offset {
    transform: translateX(8px);
  }

  .vinta-diamond-pattern.size-compact .diamond-cell {
    width: 9.5px;
    height: 9.5px;
    border-radius: 1.2px;
  }

  .diamond-cell.dark {
    background: var(--vinta-agenda-diamond-dark, rgba(0, 0, 0, 0.24));
  }

  .diamond-cell.light {
    background: var(--vinta-agenda-diamond-light, rgba(255, 255, 255, 0.24));
  }

  .diamond-cell.v1 {
    opacity: 0.45;
  }

  .diamond-cell.v2 {
    opacity: 0.7;
  }

  .diamond-cell.v3 {
    opacity: 0.9;
  }

  .diamond-cell.v4 {
    opacity: 1;
  }
</style>
