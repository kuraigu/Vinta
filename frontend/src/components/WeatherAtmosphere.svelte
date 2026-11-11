<!--
  @component WeatherAtmosphere
  @description Shared GPU-accelerated atmospheric background waves, particle effects,
  and theme-adaptive color-wash filter for WeatherGlance and Weather hero surfaces.
-->
<script lang="ts">
  import CapsulePattern from './patterns/CapsulePattern.svelte'
  import type { WeatherConditionType } from '../illustrations/weather/WeatherArt.svelte'

  interface Props {
    condition: WeatherConditionType
    seamless?: boolean
  }

  let { condition = 'cloudy', seamless = true }: Props = $props()
</script>

<div
  class="weather-card-aura-bg {condition}"
  class:is-seamless={seamless}
  aria-hidden="true"
>
  <div class="weather-atmosphere-canvas {condition}">
    <svg
      class="weather-waves-canvas"
      viewBox="0 0 400 120"
      preserveAspectRatio="none"
    >
      <path
        class="wave-contour wave-contour-back"
        d="M -50 70 Q 70 35 190 65 T 450 50 L 450 130 L -50 130 Z"
      />
      <path
        class="wave-contour wave-contour-mid"
        d="M -50 82 Q 90 55 220 80 T 450 68 L 450 130 L -50 130 Z"
      />
      <path
        class="wave-contour wave-contour-front"
        d="M -50 94 Q 110 75 250 92 T 450 85 L 450 130 L -50 130 Z"
      />
    </svg>

    <CapsulePattern align="right" rows={16} cols={20} />

    {#if condition === 'sunny'}
      <div class="weather-concentric-halos">
        <span class="halo h1"></span>
        <span class="halo h2"></span>
        <span class="halo h3"></span>
      </div>
    {:else if condition === 'night'}
      <div class="weather-concentric-halos night-halos">
        <span class="halo h1"></span>
        <span class="halo h2"></span>
        <span class="halo h3"></span>
      </div>
      <div class="weather-starfield">
        <span class="star st1"></span>
        <span class="star st2"></span>
        <span class="star st3"></span>
        <span class="star st4"></span>
        <span class="star st5"></span>
        <span class="star st6"></span>
      </div>
    {:else if condition === 'rain' || condition === 'thunder'}
      <div class="weather-rain-curtain">
        <span class="drop d1"></span>
        <span class="drop d2"></span>
        <span class="drop d3"></span>
        <span class="drop d4"></span>
        <span class="drop d5"></span>
        <span class="drop d6"></span>
        <span class="drop d7"></span>
        <span class="drop d8"></span>
        <span class="drop d9"></span>
        <span class="drop d10"></span>
        <span class="drop d11"></span>
        <span class="drop d12"></span>
      </div>
    {:else if condition === 'snow'}
      <div class="weather-snowfall">
        <span class="flake fl1 fl-hex"></span>
        <span class="flake fl2"></span>
        <span class="flake fl3 fl-hex"></span>
        <span class="flake fl4"></span>
        <span class="flake fl5 fl-hex"></span>
        <span class="flake fl6"></span>
        <span class="flake fl7"></span>
      </div>
    {:else if condition === 'cloudy' || condition === 'cloudy-night' || condition === 'fog'}
      <div class="weather-cloud-drifts">
        <span class="drift dr1"></span>
        <span class="drift dr2"></span>
        <span class="drift dr3"></span>
        <span class="drift dr4"></span>
      </div>
    {/if}
  </div>

  {#if seamless}
    <div class="weather-seamless-scrim"></div>
  {/if}
</div>

<style src="./styles/WeatherAtmosphere.css"></style>
