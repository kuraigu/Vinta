<!--
  @component DynamicEqualizer
  @description Animated sound wave visualizer bars that bounce while media is actively playing.
-->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { isIslandHiddenStore } from '../store'

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

  // Cached dimension and theme metrics to avoid layout/style thrashing during rAF
  let cssWidth = 0
  let cssHeight = 0
  let dpr = 1
  let cachedPrimaryColor = '#ff6b00'
  let cachedPurpleColor = '#c084fc'
  let cachedCustomColor = ''
  let cachedGradient: CanvasGradient | null = null

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

  function refreshThemeColors() {
    if (!canvasRef) return
    const style = getComputedStyle(canvasRef)
    cachedPrimaryColor =
      style.getPropertyValue('--vinta-accent-primary').trim() || '#ff6b00'
    cachedPurpleColor =
      style.getPropertyValue('--vinta-accent-purple').trim() || '#c084fc'
    if (color && color.startsWith('var(')) {
      const varName = color.replace(/^var\(\s*([^,)]+).*/, '$1').trim()
      cachedCustomColor = style.getPropertyValue(varName).trim() || '#00f0ff'
    } else {
      cachedCustomColor = color || ''
    }
    cachedGradient = null
  }

  function handleColorChangeEffect() {
    void color
    refreshThemeColors()
    if (!isPlaying) {
      renderStatic()
    }
  }

  $effect(handleColorChangeEffect)

  function updateCanvasSize() {
    if (!canvasRef) return
    const rect = canvasRef.getBoundingClientRect()
    if (cssWidth !== rect.width) {
      cachedGradient = null
    }
    cssWidth = rect.width
    cssHeight = rect.height
    dpr = window.devicePixelRatio || 1

    const targetW = Math.round(cssWidth * dpr)
    const targetH = Math.round(cssHeight * dpr)

    if (canvasRef.width !== targetW || canvasRef.height !== targetH) {
      canvasRef.width = targetW
      canvasRef.height = targetH
    }
    refreshThemeColors()
  }

  let animTimer = 0
  const FRAME_INTERVAL_MS = 45 // ~22 FPS to match 20Hz WASAPI peak updates

  function stopLoop() {
    if (animTimer) {
      clearTimeout(animTimer)
      animTimer = 0
    }
    if (animFrame) {
      cancelAnimationFrame(animFrame)
      animFrame = 0
    }
  }

  function handlePlayStateEffect() {
    if (isPlaying && !$isIslandHiddenStore && cssWidth > 0 && cssHeight > 0) {
      if (!animFrame && !animTimer) {
        animFrame = requestAnimationFrame(renderLoop)
      }
    } else {
      stopLoop()
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
    animFrame = 0
    if (
      !canvasRef ||
      cssWidth === 0 ||
      cssHeight === 0 ||
      $isIslandHiddenStore ||
      document.hidden
    ) {
      stopLoop()
      return
    }
    drawFrame(time)
    if (isPlaying && !$isIslandHiddenStore) {
      animTimer = window.setTimeout(() => {
        animTimer = 0
        if (
          isPlaying &&
          !$isIslandHiddenStore &&
          !document.hidden &&
          cssWidth > 0 &&
          cssHeight > 0 &&
          !animFrame
        ) {
          animFrame = requestAnimationFrame(renderLoop)
        }
      }, FRAME_INTERVAL_MS)
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
    if (cachedCustomColor) {
      activeFill = cachedCustomColor
    } else {
      if (!cachedGradient) {
        const g = ctx.createLinearGradient(0, 0, width, 0)
        g.addColorStop(0, cachedPrimaryColor)
        g.addColorStop(1, cachedPurpleColor)
        cachedGradient = g
      }
      activeFill = cachedGradient
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

    // Draw all unplayed bars with clean, lighter translucent white for clear visibility across dark backdrop
    ctx.fillStyle = 'rgba(255, 255, 255, 0.42)'
    if (typeof ctx.roundRect === 'function') {
      ctx.beginPath()
      for (let i = 0; i < count; i++) {
        const b = barData[i]
        ctx.roundRect(b.x, b.y, b.w, b.h, b.r)
      }
      ctx.fill()
    } else {
      for (let i = 0; i < count; i++) {
        const b = barData[i]
        ctx.fillRect(b.x, b.y, b.w, b.h)
      }
    }

    // Draw played bars with precision clipping at progressPx
    if (progressPx > 0) {
      ctx.save()
      ctx.beginPath()
      ctx.rect(0, 0, progressPx, height)
      ctx.clip()

      ctx.fillStyle = activeFill
      if (typeof ctx.roundRect === 'function') {
        ctx.beginPath()
        for (let i = 0; i < count; i++) {
          const b = barData[i]
          if (b.x > progressPx) break
          ctx.roundRect(b.x, b.y, b.w, b.h, b.r)
        }
        ctx.fill()
      } else {
        for (let i = 0; i < count; i++) {
          const b = barData[i]
          if (b.x > progressPx) break
          ctx.fillRect(b.x, b.y, b.w, b.h)
        }
      }
      ctx.restore()
    }
  }

  function handleVisibilityChange() {
    if (
      !document.hidden &&
      isPlaying &&
      !$isIslandHiddenStore &&
      cssWidth > 0 &&
      cssHeight > 0 &&
      !animFrame &&
      !animTimer
    ) {
      animFrame = requestAnimationFrame(renderLoop)
    }
  }

  onMount(() => {
    updateCanvasSize()

    if (canvasRef && typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => {
        updateCanvasSize()
        if (cssWidth <= 0 || cssHeight <= 0) {
          stopLoop()
          return
        }
        if (isPlaying && !$isIslandHiddenStore) {
          if (!animFrame && !animTimer) {
            animFrame = requestAnimationFrame(renderLoop)
          }
        } else {
          renderStatic()
        }
      })
      resizeObserver.observe(canvasRef)
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    if (isPlaying && !$isIslandHiddenStore && cssWidth > 0 && cssHeight > 0) {
      if (!animFrame && !animTimer) {
        animFrame = requestAnimationFrame(renderLoop)
      }
    } else {
      renderStatic()
    }
  })

  onDestroy(() => {
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    stopLoop()
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
    contain: strict;
  }

  canvas.dynamic-equalizer.minimal {
    width: 18px;
    height: 12px;
  }
</style>
