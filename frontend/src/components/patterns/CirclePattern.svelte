<!--
  @component CirclePattern
  @description Decorative background canvas grid of concentric circular dots.
-->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { themeStore } from '../../store'
  import {
    drawPatternToCanvas,
    type PatternAlign,
    type PatternSize
  } from './patternCanvas'

  interface Props {
    align?: PatternAlign
    size?: PatternSize
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

  let canvasEl = $state<HTMLCanvasElement | null>(null)
  let resizeObserver: ResizeObserver | null = null

  function redrawPattern() {
    if (!canvasEl) return
    drawPatternToCanvas(canvasEl, {
      kind: 'circle',
      align,
      size,
      rows,
      cols,
      theme: $themeStore
    })
  }

  $effect(function syncPatternEffect() {
    void align
    void size
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
  class="vinta-circle-pattern {className}"
  aria-hidden="true"
></canvas>

<style>
  .vinta-circle-pattern {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    z-index: 0;
  }
</style>
