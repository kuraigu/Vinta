<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { slide } from 'svelte/transition'
  import {
    Wind,
    Droplets,
    Compass,
    RefreshCw,
    WifiOff,
    MapPin,
    ChevronDown,
    ArrowRight,
    CloudRain,
    Gauge,
    Sun,
    Sunset,
    CalendarDays
  } from 'lucide-svelte'
  import {
    weatherLocationStore,
    weatherUnitStore,
    simulateOfflineStore,
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
  import BackButton from './BackButton.svelte'
  import ScrollView from './ScrollView.svelte'
  import { islandTooltip } from '../services/islandTooltip'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let {
    expanded = false,
    variant = 'widget',
    simulateOffline = false
  }: {
    expanded?: boolean
    variant?: 'widget' | 'page'
    simulateOffline?: boolean
  } = $props()

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

  function computeDisplayFeelsLike() {
    if (!weather) return { val: 0, symbol: '°', str: '--' }
    return convertTemp(weather.feelsLike, $weatherUnitStore)
  }
  let displayFeelsLike = $derived.by(computeDisplayFeelsLike)

  /**
   * Navigates directly to the dedicated WeatherPage.
   */
  function openWeatherPage(e?: MouseEvent) {
    if (e) e.stopPropagation()
    if (navigate) {
      navigate(WeatherPage, weatherPageConfig, true)
    }
  }

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

  function computeShortCityName() {
    const raw = $weatherLocationStore.name || 'Auto'
    return raw.split(',')[0].trim()
  }
  let shortCityName = $derived.by(computeShortCityName)

  function computeFormattedSunrise() {
    if (!weather?.sunrise) return '--:--'
    try {
      const d = new Date(weather.sunrise)
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    } catch {
      return '--:--'
    }
  }
  let formattedSunrise = $derived.by(computeFormattedSunrise)

  function computeFormattedSunset() {
    if (!weather?.sunset) return '--:--'
    try {
      const d = new Date(weather.sunset)
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    } catch {
      return '--:--'
    }
  }
  let formattedSunset = $derived.by(computeFormattedSunset)

  function computeHeroIcon() {
    return weatherInfo.icon
  }
  let HeroIcon = $derived.by(computeHeroIcon)

  function getUvRating(uvStr: string): string {
    const n = parseFloat(uvStr) || 0
    if (n >= 8) return 'Very High'
    if (n >= 6) return 'High'
    if (n >= 3) return 'Moderate'
    return 'Low'
  }
</script>

{#snippet bannerCard(isPage: boolean)}
  <div
    class="weather-banner-card {illustrationType}"
    class:variant-page={isPage}
    onclick={() => (!isPage ? (isForecastOpen = !isForecastOpen) : null)}
    role={isPage ? 'region' : 'button'}
    tabindex={isPage ? -1 : 0}
    onkeydown={(e) => {
      if (!isPage && (e.key === 'Enter' || e.key === ' ')) {
        isForecastOpen = !isForecastOpen
      }
    }}
  >
    {#if loading && !weather}
      <div class="banner-loading">
        <RefreshCw size={18} class="spin-icon" />
        <span>Loading live weather...</span>
      </div>
    {:else if error && !weather}
      <div class="banner-error">
        <div class="error-text">
          <span class="error-title">Weather Unavailable</span>
          <span class="error-desc">Check connection or location settings</span>
        </div>
        <button class="retry-btn" onclick={handleRetry}>
          <RefreshCw size={12} />
          <span>Retry</span>
        </button>
      </div>
    {:else if weather}
      <div class="banner-main-row">
        <!-- Left Condition & Temperature Block -->
        <div class="banner-left">
          <span class="condition-name">{bannerConditionTitle}</span>
          <div class="metrics-split">
            <span class="banner-temp">
              {displayTemp.val}<span class="temp-degree">°</span>
            </span>
            <div class="metrics-divider"></div>
            <div class="meta-stack">
              <span class="meta-loc">
                <MapPin size={11} class="loc-pin" />
                {shortCityName}
              </span>
              <span class="meta-hl">
                H: {displayHigh.str} • L: {displayLow.str}
              </span>
              <span class="meta-feels">
                Feels like {displayFeelsLike.val}°
              </span>
            </div>
          </div>
        </div>

        <!-- Right Illustrated Graphic -->
        <div class="banner-right">
          {#if illustrationType === 'sunny'}
            <!-- Sun with Radiating Glow -->
            <svg class="weather-art-svg" viewBox="0 0 100 80">
              <defs>
                <radialGradient id="sunGlow" cx="50%" cy="50%" r="50%">
                  <stop offset="0%" stop-color="#fef08a" stop-opacity="0.9" />
                  <stop offset="60%" stop-color="#f59e0b" stop-opacity="0.6" />
                  <stop offset="100%" stop-color="#f59e0b" stop-opacity="0" />
                </radialGradient>
              </defs>
              <circle cx="50" cy="40" r="34" fill="url(#sunGlow)" />
              <circle
                cx="50"
                cy="40"
                r="18"
                fill="#fde047"
                filter="drop-shadow(0 0 10px rgba(250, 204, 21, 0.8))"
              />
            </svg>
          {:else if illustrationType === 'night'}
            <!-- Crescent Moon with Stars -->
            <svg class="weather-art-svg" viewBox="0 0 100 80">
              <path
                d="M 52 14 A 24 24 0 1 0 76 60 A 20 20 0 0 1 52 14 Z"
                fill="#fef08a"
                filter="drop-shadow(0 0 12px rgba(254, 240, 138, 0.7))"
              />
              <circle cx="28" cy="22" r="1.5" fill="#ffffff" opacity="0.8" />
              <circle cx="34" cy="54" r="2" fill="#ffffff" opacity="0.9" />
              <circle cx="78" cy="26" r="1.5" fill="#ffffff" opacity="0.75" />
            </svg>
          {:else if illustrationType === 'rain'}
            <!-- Cloud with Rain Showers -->
            <svg class="weather-art-svg" viewBox="0 0 100 80">
              <defs>
                <linearGradient id="cloudGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#ffffff" stop-opacity="0.95" />
                  <stop offset="100%" stop-color="#cbd5e1" stop-opacity="0.9" />
                </linearGradient>
              </defs>
              <path
                d="M 28 44 A 12 12 0 0 1 28 20 A 16 16 0 0 1 52 12 A 18 18 0 0 1 76 24 A 12 12 0 0 1 76 44 Z"
                fill="url(#cloudGrad)"
                filter="drop-shadow(0 3px 6px rgba(0,0,0,0.15))"
              />
              <line
                x1="32"
                y1="50"
                x2="28"
                y2="60"
                stroke="#67e8f9"
                stroke-width="2.5"
                stroke-linecap="round"
              />
              <line
                x1="46"
                y1="50"
                x2="42"
                y2="62"
                stroke="#67e8f9"
                stroke-width="2.5"
                stroke-linecap="round"
              />
              <line
                x1="60"
                y1="50"
                x2="56"
                y2="60"
                stroke="#67e8f9"
                stroke-width="2.5"
                stroke-linecap="round"
              />
            </svg>
          {:else if illustrationType === 'thunder'}
            <!-- Dark Cloud with Lightning Bolt -->
            <svg class="weather-art-svg" viewBox="0 0 100 80">
              <path
                d="M 28 44 A 12 12 0 0 1 28 20 A 16 16 0 0 1 52 12 A 18 18 0 0 1 76 24 A 12 12 0 0 1 76 44 Z"
                fill="#475569"
                filter="drop-shadow(0 3px 8px rgba(0,0,0,0.3))"
              />
              <polygon
                points="48,42 42,54 48,54 44,66 56,50 50,50"
                fill="#fde047"
                filter="drop-shadow(0 0 6px #fde047)"
              />
            </svg>
          {:else}
            <!-- Sun Partially Covered by Day Cloud -->
            <svg class="weather-art-svg" viewBox="0 0 100 80">
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

      {#if !isPage}
        <!-- Expand Indicator Toggle & Weather Page Jump -->
        <div class="expand-indicator-row">
          <button
            type="button"
            class="weather-jump-btn"
            onclick={openWeatherPage}
            use:islandTooltip={'Open full weather forecast'}
          >
            <span>Weather details</span>
            <ArrowRight size={11} />
          </button>
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
                <span>{weather.humidity}% Humidity</span>
              </div>
            </div>

            <div class="forecast-days-grid">
              {#each weather.forecast.slice(0, 4) as f (f.day)}
                {@const DayIcon = f.icon}
                {@const maxConv = convertTemp(f.tempMax, $weatherUnitStore)}
                {@const minConv = convertTemp(f.tempMin, $weatherUnitStore)}
                <div class="f-day-item">
                  <span class="f-day-title">{f.day}</span>
                  <DayIcon size={16} class="f-day-icon" />
                  <span class="f-day-temp">{maxConv.val}° / {minConv.val}°</span
                  >
                </div>
              {/each}
            </div>
          </div>
        {/if}
      {/if}
    {/if}
  </div>
{/snippet}

{#if !expanded}
  <!-- MINIMAL CAPSULE PILL -->
  <div
    class="weather-minimal"
    style="--weather-color: {weatherInfo.theme.color}"
  >
    {#if loading}
      <div class="minimal-skeleton"></div>
    {:else if error}
      <WifiOff size={14} class="offline-icon" />
      <span class="temp-minimal">--</span>
    {:else}
      {@const Icon = weatherInfo.icon}
      <Icon size={14} class="minimal-icon" />
      <span class="temp-minimal">{displayTemp.val}°</span>
    {/if}
  </div>
{:else if variant === 'page'}
  <!-- FULL PAGE STANDALONE WEATHER VIEW -->
  <div
    class="weather-page-view"
    style="--weather-color: {weatherInfo.theme
      .color}; --weather-glow: {weatherInfo.theme.glow};"
  >
    <div class="weather-header">
      <div class="header-left">
        <BackButton />
        <span class="header-title-text">Weather</span>
      </div>
      <button
        type="button"
        class="refresh-btn"
        class:spinning={loading}
        onclick={() => handleRetry()}
        use:islandTooltip={'Refresh weather'}
        aria-label="Refresh weather data"
      >
        <RefreshCw size={14} />
      </button>
    </div>

    {#if loading && !weather}
      <div class="bento-tile-loading">
        <span>Loading live weather...</span>
      </div>
    {:else if error && !weather}
      <div class="banner-error">
        <WifiOff size={24} class="offline-icon" />
        <span class="error-title">Unable to fetch weather</span>
        <button type="button" class="retry-btn" onclick={() => handleRetry()}>
          <RefreshCw size={12} />
          <span>Retry</span>
        </button>
      </div>
    {:else if weather}
      <ScrollView class="weather-scroll-area">
        <div class="weather-content-stream">
          <!-- Hero Illustrated Card (Identical to Home feed card) -->
          {@render bannerCard(true)}

          <!-- 7-Day Forecast Section: Consolidated into 1 Bento Card with Dividers -->
          {#if weather.forecast && weather.forecast.length > 0}
            <section class="forecast-section">
              <div class="forecast-bento-card">
                <div class="forecast-bento-header">
                  <CalendarDays size={13} />
                  <span>7-Day Forecast</span>
                </div>
                <div class="forecast-days-strip">
                  {#each weather.forecast as f, idx (f.day + idx)}
                    {@const DayIcon = f.icon}
                    {@const maxConv = convertTemp(f.tempMax, $weatherUnitStore)}
                    {@const minConv = convertTemp(f.tempMin, $weatherUnitStore)}
                    <div class="forecast-col">
                      <span class="fc-day-title">{f.day}</span>
                      <DayIcon size={18} class="f-day-icon" />
                      <div class="fc-temp-split">
                        <span class="fc-max">{maxConv.val}°</span>
                        <span class="fc-min">{minConv.val}°</span>
                      </div>
                    </div>
                    {#if idx < weather.forecast.length - 1}
                      <div class="forecast-col-divider"></div>
                    {/if}
                  {/each}
                </div>
              </div>
            </section>
          {/if}

          <!-- Standardized Bento Metrics Grid -->
          <div class="bento-metrics-grid">
            <!-- Wind -->
            <div class="bento-tile">
              <div class="tile-head">
                <Wind size={13} class="tile-head-icon" />
                <span>Wind</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num">{weather.windSpeed}</span>
                <span class="tile-val-unit">km/h</span>
              </div>
              <div class="tile-sub">
                {weather.windDirectionLabel} • {weather.windDirection}°
              </div>
            </div>

            <!-- UV Index -->
            <div class="bento-tile">
              <div class="tile-head">
                <Compass size={13} class="tile-head-icon" />
                <span>UV Index</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num"
                  >{parseFloat(weather.uvIndex) || weather.uvIndex}</span
                >
                <span class="tile-val-unit">UV</span>
              </div>
              <div class="tile-sub">
                {getUvRating(weather.uvIndex)}
              </div>
            </div>

            <!-- Humidity -->
            <div class="bento-tile">
              <div class="tile-head">
                <Droplets size={13} class="tile-head-icon" />
                <span>Humidity</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num">{weather.humidity}</span>
                <span class="tile-val-unit">%</span>
              </div>
              <div class="tile-sub">
                {weather.humidity >= 70
                  ? 'High Humidity'
                  : weather.humidity <= 35
                    ? 'Dry Air'
                    : 'Comfortable'}
              </div>
            </div>

            <!-- Pressure -->
            <div class="bento-tile">
              <div class="tile-head">
                <Gauge size={13} class="tile-head-icon" />
                <span>Pressure</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num">{weather.pressure}</span>
                <span class="tile-val-unit">hPa</span>
              </div>
              <div class="tile-sub">Atmospheric</div>
            </div>

            <!-- Sun Timing -->
            <div class="bento-tile">
              <div class="tile-head">
                <Sun size={13} class="tile-head-icon" />
                <span>Sun Timing</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num time-val">{formattedSunrise}</span>
                <span class="tile-val-unit">rise</span>
              </div>
              <div class="tile-sub">
                Sunset: {formattedSunset}
              </div>
            </div>

            <!-- Precipitation -->
            <div class="bento-tile">
              <div class="tile-head">
                <CloudRain size={13} class="tile-head-icon" />
                <span>Precipitation</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num"
                  >{weather.precipitationProbability}</span
                >
                <span class="tile-val-unit">%</span>
              </div>
              <div class="tile-sub">Peak probability today</div>
            </div>
          </div>
        </div>
      </ScrollView>
    {/if}
  </div>
{:else}
  <!-- WIDGET VARIANT (Home Feed Card) -->
  {@render bannerCard(false)}
{/if}

<style src="./styles/Weather.css"></style>
