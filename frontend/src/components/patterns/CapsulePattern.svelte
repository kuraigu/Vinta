<!--
  @component CapsulePattern
  @description Decorative background canvas mosaic of subtle pill shapes used in themed card backgrounds.
-->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { themeStore } from '../../store'
  import { drawPatternToCanvas, type PatternAlign } from './patternCanvas'

  interface Props {
    align?: PatternAlign
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

  let canvasEl = $state<HTMLCanvasElement | null>(null)
  let resizeObserver: ResizeObserver | null = null

  function redrawPattern() {
    if (!canvasEl) return
    drawPatternToCanvas(canvasEl, {
      kind: 'capsule',
      align,
      rows,
      cols,
      theme: $themeStore
    })
  }

  $effect(function syncPatternEffect() {
    void align
    void rows
    void cols
    void $themeStore
    redrawPattern()
  })

  onMount(() => {
    redrawPattern()
    if (canvasEl && typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(redrawPattern)
      resizeObserver.observe(canvasEl)
    }
  })

  onDestroy(() => {
    resizeObserver?.disconnect()
    resizeObserver = null
  })
</script>

<canvas
  bind:this={canvasEl}
  class="vinta-capsule-pattern {className}"
  aria-hidden="true"
></canvas>

<style>
  .vinta-capsule-pattern {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    z-index: 0;
  }
</style>
