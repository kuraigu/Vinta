<!--
  @fileoverview Weather forecast and environmental metrics display component.

  Fetches real-time temperature, humidity, UV index, wind speed, air pressure, and
  7-day outlooks from Open-Meteo with offline caching and condition artwork.
-->
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
    CloudRain,
    Gauge,
    Sun,
    CalendarDays,
    MoreHorizontal
  } from 'lucide-svelte'
  import {
    weatherLocationStore,
    weatherUnitStore,
    simulateOfflineStore,
    simulateWeatherStore,
    tStore,
    type IslandConfig
  } from '../store'
  import {
    fetchWeatherData,
    convertTemp,
    parseWmoCode,
    getWmoConditionKey,
    type WeatherData
  } from '../services/weatherService'
  import WeatherPage, {
    config as weatherPageConfig
  } from '../pages/WeatherPage.svelte'
  import BackButton from './BackButton.svelte'
  import RefreshButton from './RefreshButton.svelte'
  import ScrollView from './ScrollView.svelte'
  import WeatherAtmosphere from './WeatherAtmosphere.svelte'
  import WeatherArt, {
    type WeatherConditionType
  } from '../illustrations/weather/WeatherArt.svelte'
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
   * Queries Open-Meteo for meteorological data matching target coordinates.
   *
   * @param name - Location display name.
   * @param lat - Latitude coordinate.
   * @param lon - Longitude coordinate.
   * @param offline - Whether to read directly from local cache.
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

  function computeUnitLetter(): string {
    if ($weatherUnitStore === 'fahrenheit') return 'F'
    if ($weatherUnitStore === 'kelvin') return 'K'
    return 'C'
  }
  let unitLetter = $derived.by(computeUnitLetter)

  /**
   * Navigates directly to the dedicated WeatherPage.
   *
   * @param e - Optional click event.
   */
  function openWeatherPage(e?: MouseEvent) {
    if (e) e.stopPropagation()
    if (navigate) {
      navigate(WeatherPage, weatherPageConfig, true)
    }
  }

  /**
   * Derives parsed condition code metadata for current weather state.
   */
  function computeWeatherInfo() {
    if (!weather) return parseWmoCode(2, isNight)
    return parseWmoCode(weather.weatherCode, isNight)
  }
  let weatherInfo = $derived.by(computeWeatherInfo)

  /**
   * Evaluates whether current conditions reflect daytime or nighttime.
   */
  function computeIsNight() {
    if (typeof weather?.isNight === 'boolean') {
      return weather.isNight
    }
    const hour = new Date().getHours()
    return hour < 6 || hour >= 18
  }
  let isNight = $derived.by(computeIsNight)

  /**
   * Resolves the primary artwork illustration variant matching active conditions.
   *
   * @returns Condition identifier string for illustration component.
   */
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

  /**
   * Resolves the secondary illustration icon variant for daily forecast tiles.
   *
   * @param code - WMO weather interpretation code.
   * @returns Condition identifier string for forecast artwork.
   */
  function getForecastIllustrationType(code: number): WeatherConditionType {
    if (code >= 95) return 'thunder'
    if ((code >= 51 && code <= 67) || (code >= 80 && code <= 82)) return 'rain'
    if (code >= 71 && code <= 77) return 'snow'
    if (code === 0) return 'sunny'
    if (code >= 1 && code <= 3) return 'cloudy'
    if (code === 45 || code === 48) return 'fog'
    return 'cloudy'
  }

  /**
   * Derives localized condition title string corresponding to current WMO code.
   *
   * @returns Translated weather condition phrase.
   */
  function computeBannerConditionTitle() {
    if (!weather) {
      return isNight
        ? $tStore('weather.conditions.clear_night')
        : $tStore('weather.conditions.sunny')
    }
    const key = getWmoConditionKey(
      weather.weatherCode ?? 0,
      isNight,
      weather.condition
    )
    return $tStore(key)
  }
  let bannerConditionTitle = $derived.by(computeBannerConditionTitle)

  /**
   * Extracts a concise city name from longer location strings for compact headers.
   *
   * @returns Short city name string.
   */
  function computeShortCityName() {
    const raw = $weatherLocationStore.name || 'Auto'
    return raw.split(',')[0].trim()
  }
  let shortCityName = $derived.by(computeShortCityName)

  /**
   * Formats sunrise time into local clock format string.
   */
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

  /**
   * Formats sunset time into local clock format string.
   */
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

  /**
   * Evaluates numerical UV index value to determine qualitative safety rating.
   *
   * @param uvStr - Numeric UV index as string.
   * @returns Localized UV rating (Low, Moderate, High, Very High, Extreme).
   */
  function getUvRating(uvStr: string): string {
    const n = parseFloat(uvStr) || 0
    if (n >= 8) return 'Very High'
    if (n >= 6) return 'High'
    if (n >= 3) return 'Moderate'
    return 'Low'
  }
</script>

{#snippet bannerCard(cardMode: 'widget' | 'page')}
  {@const isPage = cardMode === 'page'}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="weather-banner-card {illustrationType}"
    class:variant-page={isPage}
    onclick={() => (!isPage ? (isForecastOpen = !isForecastOpen) : null)}
  >
    <WeatherAtmosphere condition={illustrationType} seamless={isPage} />

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
      <div class="banner-top-layout">
        <!-- Left Content Details Block -->
        <div class="banner-left-content">
          <!-- Temperature -->
          <div class="banner-temp-row">
            <span class="banner-temp">{displayTemp.val}</span>
            <span class="banner-unit">°{unitLetter}</span>
          </div>

          <!-- Secondary Condition & Heat Index (Feels like) -->
          <div class="banner-sub-meta">
            <span class="meta-sub-item">{bannerConditionTitle}</span>
            <span class="sub-divider">|</span>
            <span class="meta-sub-item">
              {$tStore('weather.feels_like_temp', {
                temp: `${displayFeelsLike.val}°${unitLetter}`
              })}
            </span>
          </div>

          <!-- Location Name -->
          <div class="banner-location">
            <MapPin size={11} class="loc-pin" />
            <span>{$weatherLocationStore.name || shortCityName}</span>
          </div>

          <!-- Forecast Button -->
          {#if !isPage}
            <button
              type="button"
              class="see-forecast-btn"
              onclick={(e) => {
                e.stopPropagation()
                isForecastOpen = !isForecastOpen
              }}
            >
              <span
                >{isForecastOpen
                  ? $tStore('weather.hide_forecast')
                  : $tStore('weather.see_forecast')}</span
              >
              <ChevronDown
                size={13}
                strokeWidth={2.4}
                class="chevron-icon {isForecastOpen ? 'open' : ''}"
              />
            </button>
          {/if}
        </div>

        <!-- Right Illustrated Graphic & More Action -->
        <div class="banner-right-art">
          {#if !isPage}
            <button
              type="button"
              class="weather-more-btn"
              onclick={(e) => {
                e.stopPropagation()
                openWeatherPage()
              }}
              use:islandTooltip={$tStore('weather.details')}
              aria-label={$tStore('weather.details')}
            >
              <MoreHorizontal size={14} />
            </button>
          {/if}

          <div class="weather-art-wrap">
            <WeatherArt type={illustrationType} variant="hero" />
          </div>
        </div>
      </div>

      {#if !isPage && isForecastOpen}
        <!-- Animated Forecast & Detailed Metrics Drawer -->
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
              <span>{weather.humidity}% {$tStore('weather.humidity')}</span>
            </div>
          </div>

          <div class="forecast-strip-box">
            {#each weather.forecast.slice(0, 4) as f (f.day)}
              {@const forecastArtType = getForecastIllustrationType(
                f.weatherCode
              )}
              {@const maxConv = convertTemp(f.tempMax, $weatherUnitStore)}
              {@const minConv = convertTemp(f.tempMin, $weatherUnitStore)}
              <div class="f-day-col">
                <span class="f-day-title">{f.day}</span>
                <div class="f-day-art-wrap">
                  <WeatherArt
                    type={forecastArtType}
                    variant="mini"
                    monochrome={true}
                  />
                </div>
                <span class="f-day-temp">{maxConv.val}° / {minConv.val}°</span>
              </div>
            {/each}
          </div>
        </div>
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
      <WeatherArt type={illustrationType} variant="compact" />
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
    {#if loading && !weather}
      <div class="bento-tile-loading">
        <span>{$tStore('weather.loading')}</span>
      </div>
    {:else if error && !weather}
      <div class="banner-error">
        <WifiOff size={24} class="offline-icon" />
        <span class="error-title">{$tStore('weather.unable_to_fetch')}</span>
        <button type="button" class="retry-btn" onclick={() => handleRetry()}>
          <RefreshCw size={12} />
          <span>{$tStore('common.retry')}</span>
        </button>
      </div>
    {:else if weather}
      <ScrollView class="weather-scroll-area">
        <div class="weather-content-stream">
          <!-- Hero Illustrated Card (Identical to Home feed card) -->
          {@render bannerCard('page')}

          <!-- 7-Day Forecast Section: Consolidated into 1 Bento Card with Dividers -->
          {#if weather.forecast && weather.forecast.length > 0}
            <section class="forecast-section">
              <div class="forecast-bento-card">
                <div class="forecast-bento-header">
                  <CalendarDays size={13} />
                  <span>{$tStore('weather.forecast_7day')}</span>
                </div>
                <div class="forecast-days-strip">
                  {#each weather.forecast as f, idx (f.day + idx)}
                    {@const forecastArtType = getForecastIllustrationType(
                      f.weatherCode
                    )}
                    {@const maxConv = convertTemp(f.tempMax, $weatherUnitStore)}
                    {@const minConv = convertTemp(f.tempMin, $weatherUnitStore)}
                    <div class="forecast-col">
                      <span class="fc-day-title">{f.day}</span>
                      <div class="fc-art-badge">
                        <WeatherArt
                          type={forecastArtType}
                          variant="mini"
                          monochrome={true}
                        />
                      </div>
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
            <div class="forecast-bento-card">
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
            <div class="forecast-bento-card">
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
            <div class="forecast-bento-card">
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
            <div class="forecast-bento-card">
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
            <div class="forecast-bento-card">
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
            <div class="forecast-bento-card">
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

      <!-- Header: title is centered via a 1fr / auto / 1fr grid (see CSS) -->
      <div class="weather-header">
        <div class="header-left">
          <BackButton />
        </div>
        <span class="header-title-text">{$tStore('apps.weather_name')}</span>
        <div class="header-right">
          <RefreshButton
            {loading}
            onclick={() => handleRetry()}
            title={$tStore('common.retry')}
          />
        </div>
      </div>
    {/if}
  </div>
{:else}
  <!-- WIDGET VARIANT (Home Feed Card) -->
  {@render bannerCard('widget')}
{/if}

<style src="./styles/Weather.css"></style>
