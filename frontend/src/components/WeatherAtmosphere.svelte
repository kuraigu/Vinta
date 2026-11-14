<!--
  @component WeatherAtmosphere
  @description Shared offscreen-baked atmospheric background waves and particle effects
  for WeatherGlance and Weather hero surfaces, throttled via BakedCanvasLayer.
-->
<script lang="ts">
  import CapsulePattern from './patterns/CapsulePattern.svelte'
  import BakedCanvasLayer from './BakedCanvasLayer.svelte'
  import type { WeatherConditionType } from '../illustrations/weather/WeatherArt.svelte'

  interface Props {
    condition: WeatherConditionType
    seamless?: boolean
  }

  let { condition = 'cloudy', seamless = true }: Props = $props()

  interface WavePalette {
    back: [string, number]
    mid: [string, number]
    front: [string, number]
  }

  const WAVE_PALETTES: Record<WeatherConditionType, WavePalette> = {
    sunny: {
      back: ['#ea580c', 0.35],
      mid: ['#f97316', 0.42],
      front: ['#fb923c', 0.52]
    },
    night: {
      back: ['#1e1b4b', 0.45],
      mid: ['#312e81', 0.5],
      front: ['#3730a3', 0.4]
    },
    rain: {
      back: ['#0f172a', 0.5],
      mid: ['#1e3a8a', 0.45],
      front: ['#1d4ed8', 0.38]
    },
    thunder: {
      back: ['#030712', 0.6],
      mid: ['#1e1b4b', 0.55],
      front: ['#312e81', 0.48]
    },
    snow: {
      back: ['#1e3a8a', 0.45],
      mid: ['#2563eb', 0.4],
      front: ['#60a5fa', 0.35]
    },
    cloudy: {
      back: ['#0369a1', 0.35],
      mid: ['#0284c7', 0.42],
      front: ['#38bdf8', 0.4]
    },
    'cloudy-night': {
      back: ['#0f172a', 0.5],
      mid: ['#1e1b4b', 0.45],
      front: ['#2563eb', 0.35]
    },
    fog: {
      back: ['#1e293b', 0.5],
      mid: ['#334155', 0.45],
      front: ['#64748b', 0.35]
    }
  }

  const WAVE_BACK = new Path2D('M -50 70 Q 70 35 190 65 T 450 50 L 450 130 L -50 130 Z')
  const WAVE_MID = new Path2D('M -50 82 Q 90 55 220 80 T 450 68 L 450 130 L -50 130 Z')
  const WAVE_FRONT = new Path2D('M -50 94 Q 110 75 250 92 T 450 85 L 450 130 L -50 130 Z')

  let haloSprite: HTMLCanvasElement | null = null
  let cloudSprite: HTMLCanvasElement | null = null

  function createRadialSprite(
    size: number,
    dpr: number,
    stops: [number, string][]
  ): HTMLCanvasElement {
    const c = document.createElement('canvas')
    c.width = Math.round(size * dpr)
    c.height = Math.round(size * dpr)
    const sCtx = c.getContext('2d')
    if (sCtx) {
      sCtx.setTransform(dpr, 0, 0, dpr, 0, 0)
      const r = size / 2
      const grad = sCtx.createRadialGradient(r, r, 0, r, r, r)
      for (const [offset, color] of stops) {
        grad.addColorStop(offset, color)
      }
      sCtx.fillStyle = grad
      sCtx.beginPath()
      sCtx.arc(r, r, r, 0, Math.PI * 2)
      sCtx.fill()
    }
    return c
  }

  function bakeAtmosphereSprites(
    _ctx: CanvasRenderingContext2D,
    _w: number,
    _h: number,
    dpr: number
  ) {
    const isNight = condition === 'night'
    const rgb = isNight ? '199, 210, 254' : '254, 240, 138'
    haloSprite = createRadialSprite(260, dpr, [
      [0, `rgba(${rgb}, 0.18)`],
      [0.65, `rgba(${rgb}, 0.06)`],
      [1, `rgba(${rgb}, 0)`]
    ])
    cloudSprite = createRadialSprite(128, dpr, [
      [0, 'rgba(255, 255, 255, 0.10)'],
      [1, 'rgba(255, 255, 255, 0)']
    ])
  }

  function drawWaveLayer(
    ctx: CanvasRenderingContext2D,
    w: number,
    h: number,
    path: Path2D,
    color: string,
    alpha: number,
    dx: number,
    dy: number,
    scaleY: number
  ) {
    ctx.save()
    ctx.globalAlpha = alpha
    ctx.fillStyle = color
    ctx.translate(dx, h + dy)
    ctx.scale(w / 400, (h / 120) * scaleY)
    ctx.translate(0, -120)
    ctx.fill(path)
    ctx.restore()
  }

  function drawHalos(
    ctx: CanvasRenderingContext2D,
    w: number,
    t: number,
    isNight: boolean
  ) {
    const cx = w - 25
    const cy = 20
    const radii = [65, 97.5, 130]
    const delays = [0, 1.8, 3.6]
    const rgb = isNight ? '199, 210, 254' : '254, 240, 138'

    for (let i = 0; i < radii.length; i++) {
      const phase = ((t + delays[i]) % 5.5) / 5.5
      const wave = 0.5 - 0.5 * Math.cos(phase * Math.PI * 2)
      const r = radii[i] * (0.85 + 0.2 * wave)
      const alpha = (isNight ? 0.65 : 1.0) * (0.35 + 0.65 * wave)

      if (haloSprite) {
        ctx.save()
        ctx.globalAlpha = alpha
        ctx.drawImage(haloSprite, cx - r, cy - r, r * 2, r * 2)
        ctx.restore()
      }

      ctx.strokeStyle = `rgba(${rgb}, ${((isNight ? 0.12 : 0.18) * 1.4 * (0.35 + 0.65 * wave)).toFixed(3)})`
      ctx.lineWidth = 1
      ctx.beginPath()
      ctx.arc(cx, cy, r, 0, Math.PI * 2)
      ctx.stroke()
    }
  }

  function drawAtmosphere(
    ctx: CanvasRenderingContext2D,
    w: number,
    h: number,
    t: number
  ) {
    ctx.clearRect(0, 0, w, h)
    const palette = WAVE_PALETTES[condition] ?? WAVE_PALETTES.cloudy

    const bWave = Math.sin((t / 11) * Math.PI * 2)
    const mWave = Math.sin((t / 8.5) * Math.PI * 2 + 1.2)
    const fWave = Math.sin((t / 6.5) * Math.PI * 2 + 2.4)

    drawWaveLayer(ctx, w, h, WAVE_BACK, palette.back[0], palette.back[1], bWave * 18, bWave * -3, 1 + bWave * 0.04)
    drawWaveLayer(ctx, w, h, WAVE_MID, palette.mid[0], palette.mid[1], mWave * -18, mWave * 4, 1 + mWave * 0.04)
    drawWaveLayer(ctx, w, h, WAVE_FRONT, palette.front[0], palette.front[1], fWave * 15, fWave * -4, 1 + fWave * 0.05)

    if (condition === 'sunny' || condition === 'night') {
      drawHalos(ctx, w, t, condition === 'night')
    }

    if (condition === 'night') {
      const stars = [
        [0.18, 0.22, 1.4, 0],
        [0.34, 0.14, 1.1, 0.7],
        [0.52, 0.28, 1.5, 1.4],
        [0.68, 0.16, 1.2, 2.1],
        [0.82, 0.34, 1.3, 0.4],
        [0.25, 0.45, 1.0, 1.8]
      ]
      for (const [rx, ry, r, ph] of stars) {
        const a = 0.25 + 0.65 * (0.5 + 0.5 * Math.sin(((t + ph) / 2.5) * Math.PI * 2))
        ctx.fillStyle = `rgba(224, 231, 255, ${a.toFixed(3)})`
        ctx.beginPath()
        ctx.arc(w * rx, h * ry, r, 0, Math.PI * 2)
        ctx.fill()
      }
    } else if (condition === 'rain' || condition === 'thunder') {
      ctx.strokeStyle = 'rgba(186, 230, 253, 0.35)'
      ctx.lineWidth = 1.2
      ctx.lineCap = 'round'
      for (let i = 0; i < 12; i++) {
        const p = ((t * 1.15 + i * 0.19) % 1)
        const x = w * (0.08 + ((i * 0.083) % 0.88)) - p * 10
        const y = -16 + p * (h + 32)
        ctx.beginPath()
        ctx.moveTo(x, y)
        ctx.lineTo(x - 3, y + 12)
        ctx.stroke()
      }
    } else if (condition === 'snow') {
      for (let i = 0; i < 7; i++) {
        const p = ((t * 0.28 + i * 0.17) % 1)
        const x = w * (0.1 + ((i * 0.13) % 0.82)) + Math.sin(t * 1.6 + i) * 6
        const y = -10 + p * (h + 20)
        ctx.fillStyle = 'rgba(241, 245, 249, 0.65)'
        ctx.beginPath()
        ctx.arc(x, y, i % 2 === 0 ? 2.2 : 1.5, 0, Math.PI * 2)
        ctx.fill()
      }
    } else if (condition === 'cloudy' || condition === 'cloudy-night' || condition === 'fog') {
      if (cloudSprite) {
        for (let i = 0; i < 4; i++) {
          const drift = Math.sin((t / 12) * Math.PI * 2 + i * 1.7) * 18
          const cx = w * (0.2 + i * 0.22) + drift
          const cy = h * (0.28 + (i % 2) * 0.2)
          const r = 46 + (i % 2) * 18
          ctx.drawImage(cloudSprite, cx - r, cy - r, r * 2, r * 2)
        }
      }
    }
  }
</script>

<div
  class="weather-card-aura-bg {condition}"
  class:is-seamless={seamless}
  aria-hidden="true"
>
  <div class="weather-atmosphere-canvas {condition}">
    <BakedCanvasLayer
      fps={12}
      bakeKey={condition}
      onBake={bakeAtmosphereSprites}
      onDraw={drawAtmosphere}
    />
    <CapsulePattern align="right" rows={16} cols={20} />
  </div>

  {#if seamless}
    <div class="weather-seamless-scrim"></div>
  {/if}
</div>

<style src="./styles/WeatherAtmosphere.css"></style>
