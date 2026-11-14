<!--
  @component BakedCanvasLayer
  @description Reusable offscreen-baked canvas layer for ambient infinite animations.
  Bakes static artwork/gradients once per resize or key change and runs a capped-FPS
  loop that pauses automatically when hidden or scrolled off-screen.
-->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { isIslandHiddenStore } from '../store'
  import {
    createBakedLoop,
    type BakedLoopController
  } from '../services/bakedAnimation'

  interface Props {
    fps?: number
    paused?: boolean
    bakeKey?: unknown
    class?: string
    onBake?: (
      ctx: CanvasRenderingContext2D,
      width: number,
      height: number,
      dpr: number
    ) => void | Promise<void>
    onDraw: (
      ctx: CanvasRenderingContext2D,
      width: number,
      height: number,
      elapsedSec: number,
      bakedCanvas: HTMLCanvasElement | null
    ) => void
  }

  let {
    fps = 18,
    paused = false,
    bakeKey = undefined,
    class: className = '',
    onBake,
    onDraw
  }: Props = $props()

  let canvasEl = $state<HTMLCanvasElement | null>(null)
  let bakedCanvas: HTMLCanvasElement | null = null
  let cssWidth = 0
  let cssHeight = 0
  let dpr = 1
  let isVisibleInViewport = true
  let resizeObserver: ResizeObserver | null = null
  let intersectionObserver: IntersectionObserver | null = null
  let loop: BakedLoopController | null = null
  let bakeSeq = 0

  async function runBake() {
    if (!canvasEl || cssWidth <= 0 || cssHeight <= 0) return
    const seq = ++bakeSeq
    if (!onBake) {
      bakedCanvas = null
      drawNow(0)
      return
    }
    const offscreen = document.createElement('canvas')
    offscreen.width = Math.max(1, Math.round(cssWidth * dpr))
    offscreen.height = Math.max(1, Math.round(cssHeight * dpr))
    const offCtx = offscreen.getContext('2d')
    if (!offCtx) return
    offCtx.setTransform(dpr, 0, 0, dpr, 0, 0)
    await onBake(offCtx, cssWidth, cssHeight, dpr)
    if (seq !== bakeSeq) return
    bakedCanvas = offscreen
    drawNow(0)
  }

  function drawNow(elapsedSec: number) {
    if (!canvasEl || cssWidth <= 0 || cssHeight <= 0) return
    const ctx = canvasEl.getContext('2d')
    if (!ctx) return
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    onDraw(ctx, cssWidth, cssHeight, elapsedSec, bakedCanvas)
  }

  function updateSize() {
    if (!canvasEl) return
    const rect = canvasEl.getBoundingClientRect()
    if (rect.width <= 0 || rect.height <= 0) return
    const nextDpr = window.devicePixelRatio || 1
    const changed =
      Math.round(rect.width) !== Math.round(cssWidth) ||
      Math.round(rect.height) !== Math.round(cssHeight) ||
      nextDpr !== dpr
    if (!changed) return

    cssWidth = rect.width
    cssHeight = rect.height
    dpr = nextDpr
    canvasEl.width = Math.max(1, Math.round(cssWidth * dpr))
    canvasEl.height = Math.max(1, Math.round(cssHeight * dpr))
    void runBake()
  }

  function syncLoopEffect() {
    void paused
    void $isIslandHiddenStore
    loop?.sync()
  }
  $effect(syncLoopEffect)

  function rebakeOnKeyEffect() {
    void bakeKey
    if (cssWidth > 0 && cssHeight > 0) {
      void runBake()
    }
  }
  $effect(rebakeOnKeyEffect)

  onMount(() => {
    updateSize()

    if (canvasEl && typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => {
        updateSize()
      })
      resizeObserver.observe(canvasEl)
    }

    if (canvasEl && typeof IntersectionObserver !== 'undefined') {
      intersectionObserver = new IntersectionObserver(
        (entries) => {
          isVisibleInViewport = entries[0]?.isIntersecting ?? true
          loop?.sync()
        },
        { threshold: 0.01 }
      )
      intersectionObserver.observe(canvasEl)
    }

    loop = createBakedLoop({
      fps,
      isActive: () => !paused && !$isIslandHiddenStore && isVisibleInViewport,
      onFrame: (elapsedSec) => drawNow(elapsedSec)
    })
    loop.sync()
  })

  onDestroy(() => {
    loop?.stop()
    loop = null
    resizeObserver?.disconnect()
    resizeObserver = null
    intersectionObserver?.disconnect()
    intersectionObserver = null
    if (canvasEl) {
      canvasEl.width = 0
      canvasEl.height = 0
    }
    bakedCanvas = null
  })
</script>

<canvas bind:this={canvasEl} class="baked-canvas-layer {className}" aria-hidden="true"></canvas>

<style>
  canvas.baked-canvas-layer {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    display: block;
    contain: layout style paint;
  }
</style>
