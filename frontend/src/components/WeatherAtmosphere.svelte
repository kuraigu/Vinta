<!--
  @component WeatherAtmosphere
  @description GPU-composited atmospheric background waves and particle effects
  for WeatherGlance and Weather hero surfaces, pausing automatically when hidden,
  collapsed, or scrolled off-screen.
-->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { isIslandHiddenStore } from '../store'
  import CapsulePattern from './patterns/CapsulePattern.svelte'
  import type { WeatherConditionType } from '../illustrations/weather/WeatherArt.svelte'

  interface Props {
    condition: WeatherConditionType
    seamless?: boolean
    paused?: boolean
  }

  let {
    condition = 'cloudy',
    seamless = true,
    paused = false
  }: Props = $props()

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

  const WAVE_BACK_D = 'M -50 70 Q 70 35 190 65 T 450 50 L 450 130 L -50 130 Z'
  const WAVE_MID_D = 'M -50 82 Q 90 55 220 80 T 450 68 L 450 130 L -50 130 Z'
  const WAVE_FRONT_D = 'M -50 94 Q 110 75 250 92 T 450 85 L 450 130 L -50 130 Z'

  const HALO_RINGS = [
    { size: 130, delay: '0s' },
    { size: 195, delay: '-1.8s' },
    { size: 260, delay: '-3.6s' }
  ]

  const NIGHT_STARS = [
    { left: '18%', top: '22%', size: 2.8, delay: '0s' },
    { left: '34%', top: '14%', size: 2.2, delay: '-0.7s' },
    { left: '52%', top: '28%', size: 3.0, delay: '-1.4s' },
    { left: '68%', top: '16%', size: 2.4, delay: '-2.1s' },
    { left: '82%', top: '34%', size: 2.6, delay: '-0.4s' },
    { left: '25%', top: '45%', size: 2.0, delay: '-1.8s' }
  ]

  const RAIN_DROPS = Array.from({ length: 12 }, (_, i) => ({
    left: `${((0.08 + ((i * 0.083) % 0.88)) * 100).toFixed(2)}%`,
    delay: `${(-((i * 0.19) % 1) / 1.15).toFixed(3)}s`
  }))

  const SNOW_FLAKES = Array.from({ length: 7 }, (_, i) => ({
    left: `${((0.1 + ((i * 0.13) % 0.82)) * 100).toFixed(2)}%`,
    size: i % 2 === 0 ? 4.4 : 3.0,
    delay: `${(-((i * 0.17) % 1) / 0.28).toFixed(3)}s`
  }))

  const CLOUD_PUFFS = Array.from({ length: 4 }, (_, i) => ({
    left: `${20 + i * 22}%`,
    top: `${28 + (i % 2) * 20}%`,
    size: (46 + (i % 2) * 18) * 2,
    delay: `${(-(i * 1.7 * (12 / (Math.PI * 2)))).toFixed(2)}s`
  }))

  let palette = $derived(WAVE_PALETTES[condition] ?? WAVE_PALETTES.cloudy)
  let isNightHalo = $derived(condition === 'night')

  let rootEl = $state<HTMLDivElement | null>(null)
  let isInView = $state(true)
  let observer: IntersectionObserver | null = null

  onMount(() => {
    if (rootEl && typeof IntersectionObserver !== 'undefined') {
      observer = new IntersectionObserver(
        (entries) => {
          isInView = entries[0]?.isIntersecting ?? true
        },
        { threshold: 0.01 }
      )
      observer.observe(rootEl)
    }
  })

  onDestroy(() => {
    observer?.disconnect()
    observer = null
  })
</script>

<div
  bind:this={rootEl}
  class="weather-card-aura-bg {condition}"
  class:is-seamless={seamless}
  class:is-paused={paused || !isInView || $isIslandHiddenStore}
  aria-hidden="true"
>
  <div class="weather-atmosphere-canvas {condition}">
    <svg
      class="weather-wave-svg wave-back"
      viewBox="0 0 400 120"
      preserveAspectRatio="none"
    >
      <path d={WAVE_BACK_D} fill={palette.back[0]} opacity={palette.back[1]} />
    </svg>
    <svg
      class="weather-wave-svg wave-mid"
      viewBox="0 0 400 120"
      preserveAspectRatio="none"
    >
      <path d={WAVE_MID_D} fill={palette.mid[0]} opacity={palette.mid[1]} />
    </svg>
    <svg
      class="weather-wave-svg wave-front"
      viewBox="0 0 400 120"
      preserveAspectRatio="none"
    >
      <path
        d={WAVE_FRONT_D}
        fill={palette.front[0]}
        opacity={palette.front[1]}
      />
    </svg>

    {#if condition === 'sunny' || condition === 'night'}
      {#each HALO_RINGS as ring, i (i)}
        <div
          class="weather-halo-ring"
          class:is-night={isNightHalo}
          style="width: {ring.size}px; height: {ring.size}px; animation-delay: {ring.delay};"
        ></div>
      {/each}
    {/if}

    {#if condition === 'night'}
      {#each NIGHT_STARS as star, i (i)}
        <div
          class="weather-star"
          style="left: {star.left}; top: {star.top}; width: {star.size}px; height: {star.size}px; animation-delay: {star.delay};"
        ></div>
      {/each}
    {:else if condition === 'rain' || condition === 'thunder'}
      {#each RAIN_DROPS as drop, i (i)}
        <div
          class="weather-rain-track"
          style="left: {drop.left}; animation-delay: {drop.delay};"
        >
          <div class="weather-rain-drop"></div>
        </div>
      {/each}
    {:else if condition === 'snow'}
      {#each SNOW_FLAKES as flake, i (i)}
        <div
          class="weather-snow-track"
          style="left: {flake.left}; animation-delay: {flake.delay};"
        >
          <div
            class="weather-snow-flake"
            style="width: {flake.size}px; height: {flake.size}px;"
          ></div>
        </div>
      {/each}
    {:else if condition === 'cloudy' || condition === 'cloudy-night' || condition === 'fog'}
      {#each CLOUD_PUFFS as puff, i (i)}
        <div
          class="weather-cloud-puff"
          style="left: {puff.left}; top: {puff.top}; width: {puff.size}px; height: {puff.size}px; animation-delay: {puff.delay};"
        ></div>
      {/each}
    {/if}

    <div class="weather-capsule-layer">
      <CapsulePattern align="right" rows={16} cols={20} />
    </div>
  </div>

  {#if seamless}
    <div class="weather-seamless-scrim"></div>
  {/if}
</div>

<style src="./styles/WeatherAtmosphere.css"></style>
