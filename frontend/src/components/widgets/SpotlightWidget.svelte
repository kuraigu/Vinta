<!--
  @component SpotlightWidget
  @description Anchored Home spotlight carousel widget owning Home-specific discovery items,
  random initial rotation, and link/navigation handlers.
-->
<script lang="ts">
  import { getContext } from 'svelte'
  import { Heart, Star, Palette, MessageSquareHeart } from 'lucide-svelte'
  import { Browser } from '@wailsio/runtime'
  import SpotlightCarousel, {
    type SpotlightItem
  } from '../SpotlightCarousel.svelte'
  import SettingsPage, {
    config as settingsConfig
  } from '../../pages/SettingsPage.svelte'
  import { bakeImageToCanvas } from '../../services/bakedAnimation'
  import { reportFeedback } from '../../services/feedbackService'
  import { OpenURL } from '../../../bindings/Vinta/backend/app.js'
  import {
    activeSettingsTabStore,
    tStore,
    type IslandConfig
  } from '../../store'

  import enohotelLogoSvg from '../../illustrations/enohotel-daniel-belda/simple_logo.svg'
  import enohotelBgSvg from '../../illustrations/enohotel-daniel-belda/enohotel-card-bg.svg'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let enohotelSunSprite: HTMLCanvasElement | null = null

  async function bakeEnohotelBg(
    ctx: CanvasRenderingContext2D,
    w: number,
    h: number,
    dpr: number
  ) {
    const imgCanvas = await bakeImageToCanvas(enohotelBgSvg, w, h, dpr)
    if (imgCanvas) {
      ctx.drawImage(imgCanvas, 0, 0, w, h)
    }
    const sunSize = 64
    const sunCanvas = document.createElement('canvas')
    sunCanvas.width = Math.round(sunSize * dpr)
    sunCanvas.height = Math.round(sunSize * dpr)
    const sCtx = sunCanvas.getContext('2d')
    if (sCtx) {
      sCtx.setTransform(dpr, 0, 0, dpr, 0, 0)
      const r = sunSize / 2
      const grad = sCtx.createRadialGradient(r, r, 0, r, r, r)
      grad.addColorStop(0, 'rgba(255, 238, 173, 0.95)')
      grad.addColorStop(0.55, 'rgba(255, 224, 130, 0.45)')
      grad.addColorStop(1, 'rgba(184, 110, 51, 0)')
      sCtx.fillStyle = grad
      sCtx.beginPath()
      sCtx.arc(r, r, r, 0, Math.PI * 2)
      sCtx.fill()
      enohotelSunSprite = sunCanvas
    }
  }

  function drawEnohotelAmbient(
    ctx: CanvasRenderingContext2D,
    w: number,
    h: number,
    t: number,
    baked: HTMLCanvasElement | null
  ) {
    ctx.clearRect(0, 0, w, h)
    if (baked) {
      ctx.drawImage(baked, 0, 0, w, h)
    }
    const s = Math.max(w / 320, h / 84)
    const ox = (w - 320 * s) / 2
    const oy = (h - 84 * s) / 2

    const sunWave = 0.5 - 0.5 * Math.cos((t / 3.8) * Math.PI * 2)
    if (enohotelSunSprite) {
      const sunR = 38 * s * (1 + 0.14 * sunWave)
      ctx.save()
      ctx.globalCompositeOperation = 'screen'
      ctx.globalAlpha = 0.22 + 0.68 * sunWave
      ctx.drawImage(
        enohotelSunSprite,
        ox + 273 * s - sunR,
        oy + 39 * s - sunR,
        sunR * 2,
        sunR * 2
      )
      const floorR = 34 * s * (1 + 0.1 * sunWave)
      ctx.globalAlpha = 0.12 + 0.35 * sunWave
      ctx.drawImage(
        enohotelSunSprite,
        ox + 258 * s - floorR,
        oy + 70 * s - floorR * 0.6,
        floorR * 2,
        floorR * 1.2
      )
      ctx.restore()
    }

    const ringWaveA = 0.5 - 0.5 * Math.cos((t / 3.4) * Math.PI * 2)
    const ringWaveB = 0.5 - 0.5 * Math.cos(((t + 1.2) / 4.2) * Math.PI * 2)
    const glintAngle =
      Math.PI + 0.12 + 0.36 * (0.5 - 0.5 * Math.cos((t / 5.0) * Math.PI * 2))
    ctx.lineCap = 'round'

    ctx.strokeStyle = `rgba(235, 190, 95, ${(0.18 + 0.72 * ringWaveA).toFixed(3)})`
    ctx.lineWidth = 1.6 * s
    ctx.beginPath()
    ctx.arc(
      ox + 314 * s,
      oy + 86 * s,
      68 * s * (1 + 0.025 * ringWaveA),
      Math.PI,
      Math.PI * 1.65
    )
    ctx.stroke()

    ctx.strokeStyle = `rgba(255, 226, 148, ${(0.4 + 0.55 * ringWaveA).toFixed(3)})`
    ctx.lineWidth = 2.0 * s
    ctx.beginPath()
    ctx.arc(
      ox + 314 * s,
      oy + 86 * s,
      58 * s * (1 + 0.025 * ringWaveB),
      glintAngle,
      glintAngle + 0.28
    )
    ctx.stroke()

    ctx.strokeStyle = `rgba(235, 190, 95, ${(0.2 + 0.68 * ringWaveB).toFixed(3)})`
    ctx.lineWidth = 1.6 * s
    ctx.beginPath()
    ctx.arc(
      ox + 314 * s,
      oy + 86 * s,
      58 * s * (1 + 0.025 * ringWaveB),
      Math.PI,
      Math.PI * 1.65
    )
    ctx.stroke()

    const motes = [
      { x: 242, y: 28, r: 1.4, dur: 3.6, phase: 0 },
      { x: 266, y: 20, r: 1.7, dur: 4.4, phase: 1.1 },
      { x: 284, y: 24, r: 1.3, dur: 3.8, phase: 0.6 },
      { x: 222, y: 36, r: 1.3, dur: 4.6, phase: 2.1 },
      { x: 254, y: 42, r: 1.5, dur: 3.9, phase: 1.7 },
      { x: 274, y: 34, r: 1.4, dur: 4.2, phase: 2.9 }
    ]
    for (const m of motes) {
      const p = ((t + m.phase) % m.dur) / m.dur
      const alpha = p < 0.3 ? (p / 0.3) * 0.95 : (1 - (p - 0.3) / 0.7) * 0.95
      const dy = 6 - 22 * p
      const dx = -6 * p
      const mx = ox + (m.x + dx) * s
      const my = oy + (m.y + dy) * s
      if (enohotelSunSprite && alpha > 0.05) {
        const glowR = m.r * 3.2 * s
        ctx.globalAlpha = alpha * 0.55
        ctx.drawImage(
          enohotelSunSprite,
          mx - glowR,
          my - glowR,
          glowR * 2,
          glowR * 2
        )
      }
      ctx.globalAlpha = 1
      ctx.fillStyle = `rgba(255, 232, 165, ${Math.max(0, alpha).toFixed(3)})`
      ctx.beginPath()
      ctx.arc(mx, my, m.r * s, 0, Math.PI * 2)
      ctx.fill()
    }
  }

  let spotlightItems: SpotlightItem[] = $derived([
    {
      id: 'support',
      title: $tStore('spotlight.support_title'),
      description: $tStore('spotlight.support_desc'),
      iconType: 'component',
      iconComponent: Heart,
      badgeColorName: 'rose',
      url: 'https://ko-fi.com/dippedinsoju'
    },
    {
      id: 'github',
      title: $tStore('spotlight.github_title'),
      description: $tStore('spotlight.github_desc'),
      iconType: 'component',
      iconComponent: Star,
      badgeColorName: 'amber',
      url: 'https://github.com/kuraigu/Vinta'
    },
    {
      id: 'enohotel',
      title: $tStore('spotlight.enohotel_title'),
      description: $tStore('spotlight.enohotel_desc'),
      iconType: 'image',
      iconSrc: enohotelLogoSvg,
      badgeColorName: 'amber',
      ambient: {
        fps: 14,
        onBake: bakeEnohotelBg,
        onDraw: drawEnohotelAmbient
      },
      url: 'https://enohoteldanielbelda.com/'
    },
    {
      id: 'appearance',
      title: $tStore('spotlight.appearance_title'),
      description: $tStore('spotlight.appearance_desc'),
      iconType: 'component',
      iconComponent: Palette,
      badgeColorName: 'cyan',
      internalTab: 'appearance'
    },
    {
      id: 'feedback',
      title: $tStore('spotlight.feedback_title'),
      description: $tStore('spotlight.feedback_desc'),
      iconType: 'component',
      iconComponent: MessageSquareHeart,
      badgeColorName: 'emerald',
      action: () => void reportFeedback()
    }
  ])

  const initialIndex = Math.floor(Math.random() * 5)

  async function handleSelect(item: SpotlightItem) {
    if (item.action) {
      await item.action()
      return
    }

    if (item.url) {
      try {
        await Browser.OpenURL(item.url)
        return
      } catch {
        // Fallback
      }

      try {
        if (typeof OpenURL === 'function') {
          await OpenURL(item.url)
          return
        }
      } catch {
        // Fallback
      }

      window.open(item.url, '_blank')
      return
    }

    if (item.internalTab && navigate) {
      $activeSettingsTabStore = item.internalTab
      navigate(SettingsPage, settingsConfig, true)
    }
  }
</script>

<div data-vinta-widget="spotlight">
  <SpotlightCarousel
    items={spotlightItems}
    {initialIndex}
    onSelect={handleSelect}
  />
</div>
