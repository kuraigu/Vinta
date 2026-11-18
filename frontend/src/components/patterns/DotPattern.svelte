<!--
  @component DotPattern
  @description Decorative background SVG matrix of subtle dot patterns.
-->
<script lang="ts">
  interface Props {
    variant?: 'minimal-left' | 'corner-tr' | 'full'
    color?: string
    dotSize?: string
    gridSize?: string
    opacity?: number
    class?: string
  }

  let {
    variant = 'corner-tr',
    color,
    dotSize = '1.2px',
    gridSize,
    opacity,
    class: className = ''
  }: Props = $props()
</script>

<div
  class="vinta-dot-pattern variant-{variant} {className}"
  style:color
  style:--dot-size={dotSize}
  style:--grid-size={gridSize}
  style:opacity
  aria-hidden="true"
></div>

<style>
  .vinta-dot-pattern {
    position: absolute;
    pointer-events: none;
    background-image: radial-gradient(
      circle,
      currentColor var(--dot-size, 1.2px),
      transparent var(--dot-size, 1.2px)
    );
    z-index: 0;
    color: var(--vinta-primary, #38bdf8);
  }

  .vinta-dot-pattern.variant-minimal-left {
    inset: 0;
    background-size: var(--grid-size, 6px) var(--grid-size, 6px);
    opacity: 0.16;
    mask-image: radial-gradient(
      ellipse 80% 150% at 0% 50%,
      rgba(0, 0, 0, 0.95) 0%,
      rgba(0, 0, 0, 0.5) 45%,
      transparent 85%
    );
    -webkit-mask-image: radial-gradient(
      ellipse 80% 150% at 0% 50%,
      rgba(0, 0, 0, 0.95) 0%,
      rgba(0, 0, 0, 0.5) 45%,
      transparent 85%
    );
  }

  .vinta-dot-pattern.variant-corner-tr {
    top: 0;
    right: 0;
    width: 170px;
    height: 130px;
    background-size: var(--grid-size, 8px) var(--grid-size, 8px);
    opacity: 0.15;
    mask-image: radial-gradient(
      ellipse at top right,
      rgba(0, 0, 0, 1) 15%,
      transparent 75%
    );
    -webkit-mask-image: radial-gradient(
      ellipse at top right,
      rgba(0, 0, 0, 1) 15%,
      transparent 75%
    );
  }

  .vinta-dot-pattern.variant-full {
    inset: 0;
    background-size: var(--grid-size, 8px) var(--grid-size, 8px);
    opacity: 0.15;
  }
</style>
