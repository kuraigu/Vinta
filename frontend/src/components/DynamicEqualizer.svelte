<script lang="ts">
  import { onMount, onDestroy } from 'svelte'

  interface Props {
    isPlaying?: boolean
    audioPeak?: number
    barCount?: number
    variant?: 'minimal' | 'full'
    progressPercent?: number
  }

  let {
    isPlaying = false,
    audioPeak = 0,
    barCount = 24,
    variant = 'minimal',
    progressPercent
  }: Props = $props()

  let canvasRef = $state<HTMLCanvasElement | null>(null)
  let animFrame: number = 0
  let resizeObserver: ResizeObserver | null = null

  // Cached dimension metrics to avoid getBoundingClientRect layout thrashing during rAF
  let cssWidth = 0
  let cssHeight = 0
  let dpr = 1

  const MIN_HEIGHT = 0.25 // 25% baseline height when quiet/paused
  let currentHeights: number[] = []

  $effect(() => {
    const count = barCount || 24
    if (currentHeights.length !== count) {
      currentHeights = Array(count).fill(MIN_HEIGHT)
    }
  })

  function updateCanvasSize() {
    if (!canvasRef) return
    dpr = window.devicePixelRatio || 1
    const w = canvasRef.clientWidth || (variant === 'minimal' ? 18 : 200)
    const h = canvasRef.clientHeight || (variant === 'minimal' ? 12 : 24)

    if (w > 0 && h > 0) {
      cssWidth = w
      cssHeight = h
      const targetW = Math.floor(w * dpr)
      const targetH = Math.floor(h * dpr)
      if (canvasRef.width !== targetW || canvasRef.height !== targetH) {
        canvasRef.width = targetW
        canvasRef.height = targetH
      }
    }
  }

  $effect(() => {
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
  })

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
    for (let i = 0; i < count; i++) {
      currentHeights[i] = MIN_HEIGHT
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
    const amplitude =
      isPlaying && peak > 0.008
        ? Math.min(1.5, Math.max(0.35, peak * 2.8))
        : 0

    // Linear Gradient for played audio bars
    let grad: CanvasGradient | string
    if (variant === 'minimal') {
      grad = '#34c759'
    } else {
      const g = ctx.createLinearGradient(0, height, 0, 0)
      g.addColorStop(0, '#00f0ff')
      g.addColorStop(0.5, '#a855f7')
      g.addColorStop(1, '#ff007f')
      grad = g
    }

    for (let i = 0; i < count; i++) {
      let targetH = MIN_HEIGHT
      if (isPlaying && amplitude > 0) {
        const wave1 = Math.sin(t * 2.8 + i * 0.45)
        const wave2 = Math.cos(t * 4.2 + i * 0.75) * 0.5
        const wave3 = Math.sin(t * 6.5 + i * 1.3) * 0.25
        const combined = Math.abs(wave1 + wave2 + wave3) / 1.75
        targetH = MIN_HEIGHT + combined * 0.75 * amplitude
        targetH = Math.max(MIN_HEIGHT, Math.min(1.0, targetH))
      }

      currentHeights[i] += (targetH - (currentHeights[i] || MIN_HEIGHT)) * 0.25
      const hVal = currentHeights[i]

      const barH = Math.max(3, height * hVal)
      const x = i * (barWidth + gap)
      const y = height - barH

      const barPercent = (i / Math.max(1, count - 1)) * 100
      const isPlayed =
        progressPercent === undefined || barPercent <= progressPercent

      ctx.fillStyle = isPlayed ? grad : 'rgba(255, 255, 255, 0.22)'

      if (typeof ctx.roundRect === 'function') {
        ctx.beginPath()
        ctx.roundRect(x, y, barWidth, barH, [radius, radius, 0, 0])
        ctx.fill()
      } else {
        ctx.fillRect(x, y, barWidth, barH)
      }
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
    }
    if (resizeObserver) {
      resizeObserver.disconnect()
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
