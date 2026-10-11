<script lang="ts">
  import { onMount } from 'svelte'
  import { slide } from 'svelte/transition'
  import {
    Wind,
    Droplets,
    Compass,
    RefreshCw,
    WifiOff,
    MapPin,
    ChevronDown
  } from 'lucide-svelte'
  import {
    weatherLocationStore,
    weatherUnitStore,
    simulateOfflineStore
  } from '../store'
  import {
    fetchWeatherData,
    convertTemp,
    parseWmoCode,
    type WeatherData
  } from '../services/weatherService'

  let { expanded = false, simulateOffline = false } = $props<{
    expanded?: boolean
    simulateOffline?: boolean
  }>()

  let weather = $state<WeatherData | null>(null)
  let loading = $state<boolean>(true)
  let error = $state<boolean>(false)
  let isForecastOpen = $state(false)
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

  /**
   * Manually retries fetching weather data for the current location.
   */
  function handleRetry(e?: MouseEvent) {
    if (e) e.stopPropagation()
    const isOffline = Boolean($simulateOfflineStore || simulateOffline)
    const loc = $weatherLocationStore
    loadWeather(loc.name, loc.lat, loc.lon, isOffline)
  }

  onMount(() => {
    function checkAndFetch() {
      const isOffline = Boolean($simulateOfflineStore || simulateOffline)
      const loc = $weatherLocationStore
      const key = `${loc.lat.toFixed(4)}_${loc.lon.toFixed(4)}_${isOffline}`
      if (key === lastFetchKey) return
      lastFetchKey = key
      loadWeather(loc.name, loc.lat, loc.lon, isOffline)
    }

    checkAndFetch()

    const unsubLoc = weatherLocationStore.subscribe(checkAndFetch)
    const unsubOff = simulateOfflineStore.subscribe(checkAndFetch)

    return () => {
      unsubLoc()
      unsubOff()
    }
  })

  // Reactive temperature conversions
  function computeDisplayTemp() {
    if (!weather) return { val: 0, symbol: '°', str: '--' }
    return convertTemp(weather.temp, $weatherUnitStore)
  }
  let displayTemp = $derived.by(computeDisplayTemp)

  function computeDisplayHigh() {
    if (!weather) return { val: 0, symbol: '', str: '--' }
    return convertTemp(weather.high, $weatherUnitStore)
  }
  let displayHigh = $derived.by(computeDisplayHigh)

  function computeDisplayLow() {
    if (!weather) return { val: 0, symbol: '', str: '--' }
    return convertTemp(weather.low, $weatherUnitStore)
  }
  let displayLow = $derived.by(computeDisplayLow)

  function computeWeatherInfo() {
    if (!weather) return parseWmoCode(2)
    return parseWmoCode(weather.weatherCode)
  }
  let weatherInfo = $derived.by(computeWeatherInfo)

  function computeIsNight() {
    const hour = new Date().getHours()
    return hour < 6 || hour >= 18
  }
  let isNight = $derived.by(computeIsNight)

  function computeIllustrationType() {
    if (!weather) return isNight ? 'night' : 'sunny'
    const code = weather.weatherCode
    if (code >= 95) return 'thunder'
    if ((code >= 51 && code <= 67) || (code >= 80 && code <= 82)) return 'rain'
    if (code >= 71 && code <= 77) return 'snow'
    if (code === 0) return isNight ? 'night' : 'sunny'
    if (code >= 1 && code <= 3) return isNight ? 'night' : 'cloudy'
    if (code === 45 || code === 48) return 'fog'
    return isNight ? 'night' : 'sunny'
  }
  let illustrationType = $derived.by(computeIllustrationType)

  function computeBannerConditionTitle() {
    if (!weather) return isNight ? 'Clear Night' : 'Sunny'
    if (weather.weatherCode === 0) return isNight ? 'Clear Night' : 'Sunny'
    if (weather.weatherCode >= 1 && weather.weatherCode <= 3) {
      return isNight ? 'Cloudy Night' : 'Partly Cloudy'
    }
    return weather.condition
  }
  let bannerConditionTitle = $derived.by(computeBannerConditionTitle)

  function computeBannerDate() {
    try {
      const now = new Date()
      return now.toLocaleDateString('en-US', {
        weekday: 'long',
        day: 'numeric',
        month: 'long'
      })
    } catch {
      return ''
    }
  }
  let bannerDate = $derived.by(computeBannerDate)

  function computeShortCityName() {
    const raw = $weatherLocationStore.name || 'Auto'
    return raw.split(',')[0].trim()
  }
  let shortCityName = $derived.by(computeShortCityName)
</script>

{#if !expanded}
  <div
    class="weather-minimal"
    style="--weather-color: {weatherInfo.theme.color}"
  >
    {#if loading && !weather}
      <div class="minimal-skeleton"></div>
    {:else}
      {#if weather}
        {@const IconComp = weatherInfo.icon}
        <IconComp size={16} class="minimal-icon" />
        <span class="temp-minimal">{displayTemp.str}</span>
      {:else}
        <WifiOff size={15} class="offline-icon" />
        <span class="temp-minimal">--</span>
      {/if}
    {/if}
  </div>
{:else}
  <!-- Illustrated Weather Banner matching Screenshot -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="weather-banner-card {illustrationType}"
    class:forecast-expanded={isForecastOpen}
    onclick={() => {
      if (weather) isForecastOpen = !isForecastOpen
    }}
  >
    {#if loading && !weather}
      <div class="banner-loading">
        <RefreshCw size={22} class="spin-icon" />
        <span>Loading Weather...</span>
      </div>
    {:else if error && !weather}
      <div class="banner-error">
        <WifiOff size={26} class="offline-hero-icon" />
        <div class="error-text">
          <span class="error-title">Weather Offline</span>
          <span class="error-desc">Unable to fetch weather metrics</span>
        </div>
        <button class="retry-btn" onclick={handleRetry}>
          <RefreshCw size={12} /> Retry
        </button>
      </div>
    {:else if weather}
      <div class="banner-main-row">
        <!-- Left Metrics Block -->
        <div class="banner-left">
          <span class="condition-name">{bannerConditionTitle}</span>
          <div class="metrics-split">
            <span class="banner-temp"
              >{displayTemp.val}<span class="temp-degree">°</span></span
            >
            <div class="metrics-divider"></div>
            <div class="meta-stack">
              <span class="meta-date">{bannerDate}</span>
              <div class="meta-loc" title={$weatherLocationStore.name}>
                <MapPin size={11} class="loc-pin" />
                <span>{shortCityName}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Right Illustrated Vector Graphic -->
        <div class="banner-right">
          {#if illustrationType === 'sunny'}
            <svg
              class="weather-art-svg"
              viewBox="0 0 110 80"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <defs>
                <radialGradient id="sunOuterAura" cx="0.5" cy="0.5" r="0.5">
                  <stop offset="0%" stop-color="#ffffff" stop-opacity="0.25" />
                  <stop offset="60%" stop-color="#fef08a" stop-opacity="0.14" />
                  <stop offset="100%" stop-color="#fde047" stop-opacity="0" />
                </radialGradient>
                <radialGradient id="sunCore" cx="0.35" cy="0.35" r="0.6">
                  <stop offset="0%" stop-color="#fffbeb" />
                  <stop offset="45%" stop-color="#fde047" />
                  <stop offset="100%" stop-color="#f59e0b" />
                </radialGradient>
                <filter
                  id="sunBloom"
                  x="-30%"
                  y="-30%"
                  width="160%"
                  height="160%"
                >
                  <feDropShadow
                    dx="0"
                    dy="0"
                    stdDeviation="6"
                    flood-color="#f59e0b"
                    flood-opacity="0.5"
                  />
                </filter>
              </defs>
              <circle cx="78" cy="38" r="38" fill="url(#sunOuterAura)" />
              <circle cx="78" cy="38" r="26" fill="rgba(255, 255, 255, 0.15)" />
              <circle
                cx="78"
                cy="38"
                r="18"
                fill="url(#sunCore)"
                filter="url(#sunBloom)"
              />
            </svg>
          {:else if illustrationType === 'night'}
            <svg
              class="weather-art-svg"
              viewBox="0 0 110 80"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <defs>
                <radialGradient id="moonGlow" cx="0.5" cy="0.5" r="0.5">
                  <stop offset="0%" stop-color="#fef08a" stop-opacity="0.75" />
                  <stop offset="100%" stop-color="#fde047" stop-opacity="0" />
                </radialGradient>
                <linearGradient id="cloudGradNight" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#ffffff" stop-opacity="0.96" />
                  <stop
                    offset="100%"
                    stop-color="#e0e7ff"
                    stop-opacity="0.88"
                  />
                </linearGradient>
                <filter
                  id="cloudSoftShadow"
                  x="-15%"
                  y="-15%"
                  width="130%"
                  height="130%"
                >
                  <feDropShadow
                    dx="0"
                    dy="3"
                    stdDeviation="4"
                    flood-color="#1e1b4b"
                    flood-opacity="0.25"
                  />
                </filter>
              </defs>
              <!-- Star Sparkles -->
              <circle cx="24" cy="18" r="1.5" fill="#ffffff" opacity="0.8" />
              <circle cx="48" cy="12" r="1" fill="#ffffff" opacity="0.6" />
              <path
                d="M 38 26 Q 38 29 40 29 Q 38 29 38 32 Q 38 29 36 29 Q 38 29 38 26 Z"
                fill="#ffffff"
                opacity="0.85"
              />
              <!-- Crescent Moon -->
              <circle cx="78" cy="28" r="24" fill="url(#moonGlow)" />
              <path
                d="M 80 12 C 91 12 99 20 99 31 C 99 42 91 50 80 50 C 76 50 72 49 69 47 C 76 44 80 38 80 31 C 80 24 76 18 69 15 C 72 13 76 12 80 12 Z"
                fill="#fde047"
              />
              <!-- Fluffy Rounded Cloud -->
              <path
                d="M 34 54 A 10 10 0 0 1 34 34 A 13 13 0 0 1 54 26 A 14 14 0 0 1 76 36 A 10 10 0 0 1 76 54 Z"
                fill="url(#cloudGradNight)"
                filter="url(#cloudSoftShadow)"
              />
            </svg>
          {:else if illustrationType === 'rain' || illustrationType === 'thunder'}
            <svg
              class="weather-art-svg"
              viewBox="0 0 110 80"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <defs>
                <linearGradient id="rainCloudGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#f1f5f9" stop-opacity="0.95" />
                  <stop offset="100%" stop-color="#94a3b8" stop-opacity="0.9" />
                </linearGradient>
                <filter
                  id="rainCloudShadow"
                  x="-15%"
                  y="-15%"
                  width="130%"
                  height="130%"
                >
                  <feDropShadow
                    dx="0"
                    dy="3"
                    stdDeviation="4"
                    flood-color="#0f172a"
                    flood-opacity="0.25"
                  />
                </filter>
              </defs>
              <path
                d="M 34 46 A 9 9 0 0 1 34 28 A 12 12 0 0 1 52 20 A 13 13 0 0 1 72 30 A 9 9 0 0 1 72 46 Z"
                fill="url(#rainCloudGrad)"
                filter="url(#rainCloudShadow)"
              />
              {#if illustrationType === 'thunder'}
                <path
                  d="M 54 36 L 48 46 L 53 46 L 47 57 L 59 45 L 53 45 Z"
                  fill="#fde047"
                  filter="drop-shadow(0 0 4px #eab308)"
                />
              {/if}
              <line
                x1="36"
                y1="52"
                x2="31"
                y2="63"
                stroke="rgba(255,255,255,0.7)"
                stroke-width="2.5"
                stroke-linecap="round"
              />
              <line
                x1="48"
                y1="54"
                x2="43"
                y2="65"
                stroke="rgba(255,255,255,0.7)"
                stroke-width="2.5"
                stroke-linecap="round"
              />
              <line
                x1="60"
                y1="52"
                x2="55"
                y2="63"
                stroke="rgba(255,255,255,0.7)"
                stroke-width="2.5"
                stroke-linecap="round"
              />
              <line
                x1="72"
                y1="54"
                x2="67"
                y2="65"
                stroke="rgba(255,255,255,0.7)"
                stroke-width="2.5"
                stroke-linecap="round"
              />
            </svg>
          {:else}
            <!-- Default Cloud Art -->
            <svg
              class="weather-art-svg"
              viewBox="0 0 110 80"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <defs>
                <linearGradient id="cloudGradDay" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#ffffff" stop-opacity="0.98" />
                  <stop offset="100%" stop-color="#e2e8f0" stop-opacity="0.9" />
                </linearGradient>
                <filter
                  id="dayCloudShadow"
                  x="-15%"
                  y="-15%"
                  width="130%"
                  height="130%"
                >
                  <feDropShadow
                    dx="0"
                    dy="3"
                    stdDeviation="4"
                    flood-color="#0f172a"
                    flood-opacity="0.2"
                  />
                </filter>
              </defs>
              <circle
                cx="72"
                cy="24"
                r="14"
                fill="#fde047"
                filter="drop-shadow(0 0 8px rgba(250, 204, 21, 0.6))"
              />
              <path
                d="M 36 52 A 10 10 0 0 1 36 32 A 13 13 0 0 1 56 24 A 14 14 0 0 1 78 34 A 10 10 0 0 1 78 52 Z"
                fill="url(#cloudGradDay)"
                filter="url(#dayCloudShadow)"
              />
            </svg>
          {/if}
        </div>
      </div>

      <!-- Expand Indicator Toggle -->
      <div class="expand-indicator-row">
        <div class="expand-indicator">
          <span class="expand-hint"
            >{isForecastOpen ? 'Hide forecast' : 'Forecast'}</span
          >
          <ChevronDown
            size={12}
            class="chevron-icon {isForecastOpen ? 'open' : ''}"
          />
        </div>
      </div>

      <!-- Animated Forecast & Detailed Metrics Drawer -->
      {#if isForecastOpen}
        <div class="forecast-drawer" transition:slide={{ duration: 220 }}>
          <div class="drawer-metrics-row">
            <div class="drawer-chip">
              <Wind size={12} class="chip-icon" />
              <span>{weather.windSpeed} km/h</span>
            </div>
            <div class="drawer-chip">
              <Compass size={12} class="chip-icon" />
              <span>UV {weather.uvIndex}</span>
            </div>
            <div class="drawer-chip">
              <Droplets size={12} class="chip-icon" />
              <span>{weather.humidity}%</span>
            </div>
            <div class="drawer-chip">
              <span>H: {displayHigh.str} • L: {displayLow.str}</span>
            </div>
          </div>

          <div class="forecast-days-grid">
            {#each weather.forecast as f (f.day)}
              {@const FIcon = f.icon}
              <div class="f-day-item">
                <span class="f-day-title">{f.day}</span>
                <FIcon size={18} class="f-day-icon" />
                <span class="f-day-temp"
                  >{convertTemp(f.tempMax, $weatherUnitStore).str}</span
                >
              </div>
            {/each}
          </div>
        </div>
      {/if}
    {/if}
  </div>
{/if}

<style>
  /* MINIMAL */
  .weather-minimal {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--c-text-primary);
  }

  .minimal-skeleton {
    width: 45px;
    height: 16px;
    background: var(--c-item-bg);
    border-radius: 8px;
    animation: pulse 1.5s infinite alternate;
  }

  @keyframes pulse {
    0% {
      opacity: 0.4;
    }
    100% {
      opacity: 0.8;
    }
  }

  .temp-minimal {
    font-size: 1rem;
    font-weight: 500;
  }

  :global(.minimal-icon) {
    color: var(--weather-color, var(--c-amber, #f59e0b));
  }

  :global(.offline-icon) {
    color: var(--c-text-muted);
  }

  /* EXPANDED ILLUSTRATED BANNER MATCHING SCREENSHOT */
  .weather-banner-card {
    position: relative;
    width: 100%;
    min-height: 108px;
    border-radius: 22px;
    box-sizing: border-box;
    padding: 16px 18px 10px 18px;
    overflow: hidden;
    cursor: pointer;
    user-select: none;
    box-shadow: 0 4px 18px rgba(0, 0, 0, 0.18);
    transition:
      transform 0.2s ease,
      box-shadow 0.2s ease;
    color: #ffffff;
  }

  .weather-banner-card:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 22px rgba(0, 0, 0, 0.24);
  }

  .weather-banner-card:active {
    transform: scale(0.995);
  }

  /* Contextual Dynamic Gradient Backgrounds */
  .weather-banner-card.sunny {
    background: linear-gradient(135deg, #2596be 0%, #38bdf8 50%, #7dd3fc 100%);
  }

  .weather-banner-card.night {
    background: linear-gradient(135deg, #4338ca 0%, #6366f1 50%, #a855f7 100%);
  }

  .weather-banner-card.rain {
    background: linear-gradient(135deg, #0284c7 0%, #06b6d4 50%, #10b981 100%);
  }

  .weather-banner-card.thunder {
    background: linear-gradient(135deg, #312e81 0%, #4c1d95 50%, #7c3aed 100%);
  }

  .weather-banner-card.snow {
    background: linear-gradient(135deg, #0369a1 0%, #38bdf8 60%, #bae6fd 100%);
  }

  .weather-banner-card.cloudy,
  .weather-banner-card.fog {
    background: linear-gradient(135deg, #334155 0%, #475569 50%, #64748b 100%);
  }

  .banner-main-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    position: relative;
    z-index: 2;
  }

  .banner-left {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    gap: 4px;
    z-index: 3;
  }

  .condition-name {
    display: block;
    width: 100%;
    text-align: left;
    font-size: 0.85rem;
    font-weight: 500;
    color: rgba(255, 255, 255, 0.95);
    letter-spacing: -0.01em;
    margin: 0;
    padding: 0;
  }

  .metrics-split {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 1px;
  }

  .banner-temp {
    font-size: 2.65rem;
    font-weight: 350;
    letter-spacing: -0.04em;
    line-height: 1;
    font-variant-numeric: tabular-nums;
    color: #ffffff;
    display: inline-flex;
    align-items: flex-start;
  }

  .temp-degree {
    font-size: 1.4rem;
    font-weight: 300;
    opacity: 0.9;
    margin-left: 2px;
    margin-top: 1px;
  }

  .metrics-divider {
    width: 1px;
    height: 32px;
    background: rgba(255, 255, 255, 0.28);
    border-radius: 1px;
    flex-shrink: 0;
  }

  .meta-stack {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    gap: 3px;
  }

  .meta-date {
    font-size: 0.74rem;
    font-weight: 400;
    color: rgba(255, 255, 255, 0.92);
    letter-spacing: 0.01em;
    white-space: nowrap;
  }

  .meta-loc {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.74rem;
    font-weight: 500;
    color: #ffffff;
    white-space: nowrap;
  }

  :global(.loc-pin) {
    opacity: 0.95;
    flex-shrink: 0;
  }

  /* Right Illustrated Vector Art */
  .banner-right {
    position: relative;
    width: 110px;
    height: 75px;
    display: flex;
    align-items: center;
    justify-content: flex-end;
  }

  .weather-art-svg {
    width: 110px;
    height: 75px;
    overflow: visible;
    filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.12));
  }

  /* Bottom Drawer Chevron & Hint */
  .expand-indicator-row {
    display: flex;
    justify-content: center;
    width: 100%;
    margin-top: 6px;
  }

  .expand-indicator {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    margin: 8px auto 0 auto;
    padding: 2px 10px;
    border-radius: 12px;
    background: rgba(0, 0, 0, 0.14);
    opacity: 0.82;
    font-size: 0.68rem;
    font-weight: 500;
    color: rgba(255, 255, 255, 0.92);
    transition:
      opacity 0.2s ease,
      background 0.2s ease;
  }

  .weather-banner-card:hover .expand-indicator {
    opacity: 1;
    background: rgba(0, 0, 0, 0.22);
  }

  .weather-banner-card:hover .expand-indicator {
    opacity: 1;
  }

  :global(.chevron-icon) {
    transition: transform 0.25s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  :global(.chevron-icon.open) {
    transform: rotate(180deg);
  }

  /* Forecast Drawer */
  .forecast-drawer {
    width: 100%;
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid rgba(255, 255, 255, 0.22);
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .drawer-metrics-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 6px;
    flex-wrap: wrap;
  }

  .drawer-chip {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.7rem;
    font-weight: 500;
    color: rgba(255, 255, 255, 0.92);
    background: rgba(255, 255, 255, 0.16);
    padding: 3px 8px;
    border-radius: 10px;
  }

  .forecast-days-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 6px;
  }

  .f-day-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
    background: rgba(0, 0, 0, 0.12);
    padding: 6px 4px;
    border-radius: 12px;
  }

  .f-day-title {
    font-size: 0.68rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.9);
    text-transform: uppercase;
  }

  :global(.f-day-icon) {
    color: #ffffff;
    filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.2));
  }

  .f-day-temp {
    font-size: 0.74rem;
    font-weight: 700;
    color: #ffffff;
  }

  /* Loading & Error */
  .banner-loading,
  .banner-error {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    min-height: 80px;
    width: 100%;
  }

  .banner-error {
    flex-direction: column;
    gap: 6px;
    text-align: center;
  }

  .error-text {
    display: flex;
    flex-direction: column;
  }

  .error-title {
    font-weight: 700;
    font-size: 0.9rem;
  }

  .error-desc {
    font-size: 0.75rem;
    opacity: 0.8;
  }

  .retry-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.25);
    border: none;
    color: #ffffff;
    font-size: 0.72rem;
    font-weight: 600;
    cursor: pointer;
  }

  :global(.spin-icon) {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
