<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { MapPin } from 'lucide-svelte'
  import {
    weatherLocationStore,
    weatherUnitStore,
    simulateOfflineStore,
    simulateWeatherStore,
    type IslandConfig
  } from '../store'
  import {
    fetchWeatherData,
    convertTemp,
    parseWmoCode,
    type WeatherData
  } from '../services/weatherService'
  import WeatherPage, {
    config as weatherPageConfig
  } from '../pages/WeatherPage.svelte'
  import CapsulePattern from './patterns/CapsulePattern.svelte'
  import WeatherArt, {
    type WeatherConditionType
  } from '../illustrations/weather/WeatherArt.svelte'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let {
    salutation = '',
    userName = '',
    timeText = '',
    amPmText = ''
  }: {
    salutation?: string
    userName?: string
    timeText?: string
    amPmText?: string
  } = $props()

  let weather = $state<WeatherData | null>(null)
  let loading = $state<boolean>(true)
  let error = $state<boolean>(false)
  let lastFetchKey = ''

  /**
   * Fetches weather metrics for the active location.
   */
  async function loadWeather(
    name: string,
    lat: number,
    lon: number,
    offline: boolean
  ) {
    loading = true
    error = false

    try {
      const data = await fetchWeatherData(name, lat, lon, offline)
      if (data) {
        weather = data
        error = false
      } else {
        error = true
      }
    } catch {
      error = true
    } finally {
      loading = false
    }
  }

  onMount(() => {
    function checkAndFetch() {
      const isOffline = Boolean($simulateOfflineStore)
      const loc = $weatherLocationStore
      const key = `${loc.lat.toFixed(4)}_${loc.lon.toFixed(4)}_${isOffline}`
      if (key === lastFetchKey) return
      lastFetchKey = key
      loadWeather(loc.name, loc.lat, loc.lon, isOffline)
    }

    checkAndFetch()

    const unsubLoc = weatherLocationStore.subscribe(checkAndFetch)
    const unsubOff = simulateOfflineStore.subscribe(checkAndFetch)
    const unsubMock = simulateWeatherStore.subscribe((mock) => {
      if (mock) {
        weather = mock
        loading = false
        error = false
      } else {
        lastFetchKey = ''
        checkAndFetch()
      }
    })

    return () => {
      unsubLoc()
      unsubOff()
      unsubMock()
    }
  })

  function computeDisplayTemp() {
    if (!weather) return { val: 0, symbol: '°', str: '--' }
    return convertTemp(weather.temp, $weatherUnitStore)
  }
  let displayTemp = $derived.by(computeDisplayTemp)

  function computeUnitLetter(): string {
    if ($weatherUnitStore === 'fahrenheit') return 'F'
    if ($weatherUnitStore === 'kelvin') return 'K'
    return 'C'
  }
  let unitLetter = $derived.by(computeUnitLetter)

  function openWeatherPage() {
    if (navigate) {
      navigate(WeatherPage, weatherPageConfig, true)
    }
  }

  function computeIsNight() {
    if (typeof weather?.isNight === 'boolean') {
      return weather.isNight
    }
    const hour = new Date().getHours()
    return hour < 6 || hour >= 18
  }
  let isNight = $derived.by(computeIsNight)

  function computeIllustrationType(): WeatherConditionType {
    if (!weather) return isNight ? 'cloudy-night' : 'cloudy'
    const code = weather.weatherCode
    if (code >= 95) return 'thunder'
    if ((code >= 51 && code <= 67) || (code >= 80 && code <= 82)) return 'rain'
    if (code >= 71 && code <= 77) return 'snow'
    if (code === 0) return isNight ? 'night' : 'sunny'
    if (code >= 1 && code <= 3) return isNight ? 'cloudy-night' : 'cloudy'
    if (code === 45 || code === 48) return 'fog'
    return isNight ? 'cloudy-night' : 'cloudy'
  }
  let illustrationType = $derived.by(computeIllustrationType)

  function computeBannerConditionTitle() {
    if (!weather) return isNight ? 'Clear Night' : 'Sunny'
    if (weather.condition) return weather.condition
    if (weather.weatherCode === 0) return isNight ? 'Clear Night' : 'Sunny'
    if (weather.weatherCode >= 1 && weather.weatherCode <= 3) {
      return isNight ? 'Cloudy Night' : 'Partly Cloudy'
    }
    return weather.condition
  }
  let bannerConditionTitle = $derived.by(computeBannerConditionTitle)

  function computeShortCityName() {
    const raw = $weatherLocationStore.name || 'Auto'
    return raw.split(',')[0].trim()
  }
  let shortCityName = $derived.by(computeShortCityName)
</script>

<div class="weather-glance-hero {illustrationType}">
  <!-- GPU-Accelerated Dynamic Ambient Background Landscape Waves & Atmospheric FX -->
  <div class="weather-card-aura-bg {illustrationType}" aria-hidden="true">
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

    {#if illustrationType === 'sunny'}
      <div class="weather-concentric-halos">
        <span class="halo h1"></span>
        <span class="halo h2"></span>
        <span class="halo h3"></span>
      </div>
    {:else if illustrationType === 'night'}
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
    {:else if illustrationType === 'rain' || illustrationType === 'thunder'}
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
    {:else if illustrationType === 'snow'}
      <div class="weather-snowfall">
        <span class="flake fl1 fl-hex"></span>
        <span class="flake fl2"></span>
        <span class="flake fl3 fl-hex"></span>
        <span class="flake fl4"></span>
        <span class="flake fl5 fl-hex"></span>
        <span class="flake fl6"></span>
        <span class="flake fl7"></span>
      </div>
    {:else if illustrationType === 'cloudy' || illustrationType === 'cloudy-night' || illustrationType === 'fog'}
      <div class="weather-cloud-drifts">
        <span class="drift dr1"></span>
        <span class="drift dr2"></span>
        <span class="drift dr3"></span>
        <span class="drift dr4"></span>
      </div>
    {/if}
  </div>

  <div class="home-hero-layout">
    <div class="home-hero-left">
      {#if salutation}
        <span class="hero-banner-salutation">{salutation},</span>
      {/if}
      {#if userName}
        <h1 class="hero-banner-username">{userName}</h1>
      {/if}
      <div class="hero-banner-info-stack">
        {#if weather}
          <div class="hero-banner-weather-row">
            <span class="hero-weather-temp">{displayTemp.val}°{unitLetter}</span>
            <span class="sub-divider">•</span>
            <span class="hero-weather-cond">{bannerConditionTitle}</span>
            <span class="sub-divider">•</span>
            <span class="hero-weather-loc">
              <MapPin size={10} class="loc-pin" />
              {$weatherLocationStore.name || shortCityName}
            </span>
          </div>
        {/if}

        {#if timeText}
          <div class="hero-banner-time-row">
            <span class="hero-banner-time">{timeText}</span>
            {#if amPmText}
              <span class="hero-banner-ampm">{amPmText}</span>
            {/if}
          </div>
        {/if}
      </div>
    </div>

    <div class="home-hero-art">
      <div class="weather-art-wrap">
        <div
          class="weather-art-link"
          role="link"
          tabindex="0"
          aria-label="Open weather page"
          onclick={(e) => {
            e.stopPropagation()
            openWeatherPage()
          }}
          onkeydown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault()
              openWeatherPage()
            }
          }}
        >
          <WeatherArt type={illustrationType} variant="hero" />
        </div>
      </div>
    </div>
  </div>
</div>

<style src="./styles/WeatherGlance.css"></style>
