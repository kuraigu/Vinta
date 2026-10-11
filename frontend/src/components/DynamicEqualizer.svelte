<script lang="ts">
  import { onMount, onDestroy } from 'svelte'

  interface Props {
    isPlaying?: boolean
    audioPeak?: number
    barCount?: number
    variant?: 'minimal' | 'full'
    progressPercent?: number
    color?: string
  }

  let {
    isPlaying = false,
    audioPeak = 0,
    barCount = 24,
    variant = 'minimal',
    progressPercent,
    color
  }: Props = $props()

  let canvasRef = $state<HTMLCanvasElement | null>(null)
  let animFrame: number = 0
  let resizeObserver: ResizeObserver | null = null

  // Cached dimension metrics to avoid getBoundingClientRect layout thrashing during rAF
  let cssWidth = 0
  let cssHeight = 0
  let dpr = 1

  function getMinHeight(): number {
    return variant === 'full' ? 0.28 : 0.25
  }

  let currentHeights: number[] = []

  function handleBarCountChangeEffect() {
    const count = barCount || 24
    const minH = getMinHeight()
    if (currentHeights.length !== count) {
      currentHeights = Array(count).fill(minH)
    }
  }

  $effect(handleBarCountChangeEffect)

  function updateCanvasSize() {
    if (!canvasRef) return
    const rect = canvasRef.getBoundingClientRect()
    cssWidth = rect.width
    cssHeight = rect.height
    dpr = window.devicePixelRatio || 1

    const targetW = Math.round(cssWidth * dpr)
    const targetH = Math.round(cssHeight * dpr)

    if (canvasRef.width !== targetW || canvasRef.height !== targetH) {
      canvasRef.width = targetW
      canvasRef.height = targetH
    }
  }

  function handlePlayStateEffect() {
    if (isPlaying) {
      if (!animFrame) {
        animFrame = requestAnimationFrame(renderLoop)
      }
    } else {
      if (animFrame) {
        cancelAnimationFrame(animFrame)
        animFrame = 0
      }
      renderStatic()
    }
  }

  $effect(handlePlayStateEffect)

  function handleProgressChangeEffect() {
    void progressPercent
    if (!isPlaying) {
      renderStatic()
    }
  }

  $effect(handleProgressChangeEffect)

  function renderLoop(time: number) {
    if (!canvasRef) return
    drawFrame(time)
    if (isPlaying) {
      animFrame = requestAnimationFrame(renderLoop)
    }
  }

  function renderStatic() {
    if (!canvasRef) return
    const count = barCount || 24
    const minH = getMinHeight()
    for (let i = 0; i < count; i++) {
      currentHeights[i] = minH
    }
    drawFrame(0)
  }

  function drawFrame(time: number) {
    const canvas = canvasRef
    if (!canvas || cssWidth === 0 || cssHeight === 0) return

    const ctx = canvas.getContext('2d', { alpha: true })
    if (!ctx) return

    const count = barCount || 24
    const width = cssWidth
    const height = cssHeight

    // Clear raw pixel buffer
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, width, height)

    const gap = variant === 'minimal' ? 1.5 : 3
    const totalGap = gap * (count - 1)
    const barWidth = Math.max(1, (width - totalGap) / count)
    const radius = Math.min(2, barWidth / 2)

    const t = time * 0.0045
    const peak = audioPeak || 0
    const amplitude = isPlaying
      ? Math.min(1.8, 0.45 + (peak > 0.001 ? peak * 2.2 : 0))
      : 0

    let activeFill: CanvasGradient | string
    if (color && !color.startsWith('var(')) {
      activeFill = color
    } else if (color && color.startsWith('var(') && canvasRef) {
      const varName = color.replace(/^var\(\s*([^,)]+).*/, '$1').trim()
      const computed = getComputedStyle(canvasRef)
        .getPropertyValue(varName)
        .trim()
      activeFill = computed || '#00f0ff'
    } else {
      const g = ctx.createLinearGradient(0, 0, width, 0)
      g.addColorStop(0, '#00f0ff')
      g.addColorStop(0.55, '#818cf8')
      g.addColorStop(1, '#c084fc')
      activeFill = g
    }

    const minH = getMinHeight()

    const progressPx =
      progressPercent !== undefined
        ? Math.max(0, Math.min(width, (progressPercent / 100) * width))
        : width

    const barData: { x: number; y: number; w: number; h: number; r: number }[] =
      []

    for (let i = 0; i < count; i++) {
      let targetH = minH
      if (isPlaying && amplitude > 0) {
        const wave1 = Math.sin(t * 2.8 + i * 0.45)
        const wave2 = Math.cos(t * 4.2 + i * 0.75) * 0.5
        const wave3 = Math.sin(t * 6.5 + i * 1.3) * 0.25
        const combined = Math.abs(wave1 + wave2 + wave3) / 1.75
        targetH = minH + combined * (1.0 - minH) * amplitude
        targetH = Math.max(minH, Math.min(1.0, targetH))
      }

      currentHeights[i] += (targetH - (currentHeights[i] || minH)) * 0.25
      const hVal = currentHeights[i]

      const minPixelH = variant === 'full' ? 5 : 3
      const barH = Math.max(minPixelH, height * hVal)
      const x = i * (barWidth + gap)
      const y = variant === 'full' ? (height - barH) / 2 : height - barH
      const barRadius = Math.min(radius, barH / 2)

      barData.push({ x, y, w: barWidth, h: barH, r: barRadius })
    }

    // 1. Draw all unplayed bars in translucent backdrop
    ctx.fillStyle = 'rgba(255, 255, 255, 0.2)'
    for (let i = 0; i < count; i++) {
      const b = barData[i]
      if (typeof ctx.roundRect === 'function') {
        ctx.beginPath()
        ctx.roundRect(b.x, b.y, b.w, b.h, b.r)
        ctx.fill()
      } else {
        ctx.fillRect(b.x, b.y, b.w, b.h)
      }
    }

    // 2. Draw played bars with precision clipping at progressPx
    if (progressPx > 0) {
      ctx.save()
      ctx.beginPath()
      ctx.rect(0, 0, progressPx, height)
      ctx.clip()

      ctx.fillStyle = activeFill
      for (let i = 0; i < count; i++) {
        const b = barData[i]
        if (b.x > progressPx) break
        if (typeof ctx.roundRect === 'function') {
          ctx.beginPath()
          ctx.roundRect(b.x, b.y, b.w, b.h, b.r)
          ctx.fill()
        } else {
          ctx.fillRect(b.x, b.y, b.w, b.h)
        }
      }
      ctx.restore()
    }
  }

  onMount(() => {
    updateCanvasSize()

    if (canvasRef && typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => {
        updateCanvasSize()
        if (!isPlaying) renderStatic()
      })
      resizeObserver.observe(canvasRef)
    }

    renderStatic()
  })

  onDestroy(() => {
    if (animFrame) {
      cancelAnimationFrame(animFrame)
      animFrame = 0
    }
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
    if (canvasRef) {
      canvasRef.width = 0
      canvasRef.height = 0
    }
  })
</script>

<canvas
  bind:this={canvasRef}
  class="dynamic-equalizer {variant || 'minimal'}"
  class:is-playing={isPlaying}
></canvas>

<style>
  canvas.dynamic-equalizer {
    display: block;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }

  canvas.dynamic-equalizer.minimal {
    width: 18px;
    height: 12px;
  }
</style>
