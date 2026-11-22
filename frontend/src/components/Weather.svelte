<!--
  @fileoverview Weather widget, minimal capsule pill, and full-page forecast view.
-->
<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { get } from 'svelte/store'
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
    pageVisibilityStore,
    tStore,
    type IslandConfig
  } from '../store'
  import {
    fetchWeatherData,
    getCachedWeatherData,
    convertTemp,
    parseWmoCode,
    getWmoConditionKey,
    type WeatherData
  } from '../services/weatherService'
  import WeatherPage, {
    config as weatherPageConfig
  } from '../pages/WeatherPage.svelte'
  import BentoCard from './BentoCard.svelte'
  import PageHeader from './PageHeader.svelte'
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

  // Seed initial state synchronously from cache or mock store to prevent loading flashes
  const initialLoc = get(weatherLocationStore)
  const initialMock = get(simulateWeatherStore)
  const initialCache =
    initialMock ?? getCachedWeatherData(initialLoc.lat, initialLoc.lon, true)

  let weather = $state<WeatherData | null>(initialCache)
  let loading = $state<boolean>(!initialCache)
  let error = $state<boolean>(false)
  let isForecastOpen = $state(false)

  let lastFetchKey = ''
  let loadedLocationKey = initialCache
    ? `${initialLoc.lat.toFixed(4)},${initialLoc.lon.toFixed(4)}`
    : ''
  let fetchSeq = 0

  /**
   * Synchronizes weather state when the simulateOffline component prop changes after mount.
   */
  function syncOfflinePropEffect() {
    if (!lastFetchKey) return
    if (get(simulateWeatherStore)) return
    const isOffline = Boolean($simulateOfflineStore || simulateOffline)
    const loc = $weatherLocationStore
    const key = `${loc.lat.toFixed(4)}_${loc.lon.toFixed(4)}_${isOffline}`
    if (key === lastFetchKey) return
    lastFetchKey = key
    void loadWeather(loc.name, loc.lat, loc.lon, isOffline)
  }
  $effect(syncOfflinePropEffect)

  /**
   * Fetches weather for the given coordinates and guards against stale responses.
   */
  async function loadWeather(
    name: string,
    lat: number,
    lon: number,
    offline: boolean,
    forceRefresh = false
  ) {
    const seq = ++fetchSeq
    const locKey = `${lat.toFixed(4)},${lon.toFixed(4)}`

    if (!forceRefresh) {
      const cached = getCachedWeatherData(lat, lon, offline)
      if (cached) {
        weather = { ...cached, location: name, isOffline: offline }
        loadedLocationKey = locKey
        loading = false
        error = false
        return
      }
    }

    if (loadedLocationKey !== locKey) {
      const stale = getCachedWeatherData(lat, lon, true)
      if (stale) {
        weather = { ...stale, location: name, isOffline: offline }
        loadedLocationKey = locKey
      } else {
        weather = null
        loadedLocationKey = ''
      }
    }

    // Only show loading spinner when no cached data is available (stale-while-revalidate).
    loading = !weather
    error = false

    try {
      const data = await fetchWeatherData(name, lat, lon, offline, forceRefresh)
      if (seq !== fetchSeq) return

      if (data) {
        weather = data
        loadedLocationKey = locKey
        error = false
      } else if (!weather) {
        error = true
      }
    } catch {
      if (seq === fetchSeq && !weather) error = true
    } finally {
      if (seq === fetchSeq) loading = false
    }
  }

  function handleRetry(e?: MouseEvent) {
    if (e) e.stopPropagation()
    const isOffline = Boolean($simulateOfflineStore || simulateOffline)
    const loc = $weatherLocationStore
    loadWeather(loc.name, loc.lat, loc.lon, isOffline, true)
  }

  onMount(() => {
    let hadMock = Boolean(get(simulateWeatherStore))

    function checkAndFetch() {
      if (get(simulateWeatherStore)) return
      const isOffline = Boolean($simulateOfflineStore || simulateOffline)
      const loc = $weatherLocationStore
      const key = `${loc.lat.toFixed(4)}_${loc.lon.toFixed(4)}_${isOffline}`
      if (key === lastFetchKey) return
      lastFetchKey = key
      loadWeather(loc.name, loc.lat, loc.lon, isOffline)
    }

    const unsubLoc = weatherLocationStore.subscribe(checkAndFetch)
    const unsubOff = simulateOfflineStore.subscribe(checkAndFetch)
    const unsubMock = simulateWeatherStore.subscribe((mock) => {
      if (mock) {
        hadMock = true
        weather = mock
        loading = false
        error = false
      } else if (hadMock) {
        hadMock = false
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

  let displayTemp = $derived(
    weather
      ? convertTemp(weather.temp, $weatherUnitStore)
      : { val: 0, symbol: '°', str: '--' }
  )

  let displayFeelsLike = $derived(
    weather
      ? convertTemp(weather.feelsLike, $weatherUnitStore)
      : { val: 0, symbol: '°', str: '--' }
  )

  let unitLetter = $derived(
    $weatherUnitStore === 'fahrenheit'
      ? 'F'
      : $weatherUnitStore === 'kelvin'
        ? 'K'
        : 'C'
  )

  let isNight = $derived.by(() => {
    if (typeof weather?.isNight === 'boolean') return weather.isNight
    const hour = new Date().getHours()
    return hour < 6 || hour >= 18
  })

  let weatherInfo = $derived(parseWmoCode(weather?.weatherCode ?? 2, isNight))

  function mapCodeToArt(code: number, night = false): WeatherConditionType {
    if (code >= 95) return 'thunder'
    if ((code >= 51 && code <= 67) || (code >= 80 && code <= 82)) return 'rain'
    if (code >= 71 && code <= 77) return 'snow'
    if (code === 0) return night ? 'night' : 'sunny'
    if (code >= 1 && code <= 3) return night ? 'cloudy-night' : 'cloudy'
    if (code === 45 || code === 48) return 'fog'
    return night ? 'cloudy-night' : 'cloudy'
  }

  let illustrationType = $derived<WeatherConditionType>(
    weather
      ? mapCodeToArt(weather.weatherCode, isNight)
      : isNight
        ? 'cloudy-night'
        : 'cloudy'
  )

  let bannerConditionTitle = $derived(
    weather
      ? $tStore(
          getWmoConditionKey(
            weather.weatherCode ?? 0,
            isNight,
            weather.condition
          )
        )
      : $tStore(
          isNight
            ? 'weather.conditions.clear_night'
            : 'weather.conditions.sunny'
        )
  )

  let shortCityName = $derived(
    ($weatherLocationStore.name || 'Auto').split(',')[0].trim()
  )

  function formatClockTime(iso?: string): string {
    if (!iso) return '--:--'
    try {
      return new Date(iso).toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit'
      })
    } catch {
      return '--:--'
    }
  }

  let formattedSunrise = $derived(formatClockTime(weather?.sunrise))
  let formattedSunset = $derived(formatClockTime(weather?.sunset))

  function getUvRating(
    uvStr: string,
    translate: (key: string) => string
  ): string {
    const n = parseFloat(uvStr) || 0
    if (n >= 8) return translate('weather.uv_very_high')
    if (n >= 6) return translate('weather.uv_high')
    if (n >= 3) return translate('weather.uv_moderate')
    return translate('weather.uv_low')
  }

  function openWeatherPage(e?: MouseEvent) {
    if (e) e.stopPropagation()
    if ($pageVisibilityStore.weather === false) return
    navigate?.(WeatherPage, weatherPageConfig, true)
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
    <WeatherAtmosphere
      condition={illustrationType}
      seamless={false}
      paused={!expanded}
    />

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
        <div class="banner-left-content">
          <div class="banner-temp-row">
            <span class="banner-temp">{displayTemp.val}</span>
            <span class="banner-unit">°{unitLetter}</span>
          </div>

          <div class="banner-sub-meta">
            <span class="meta-sub-item">{bannerConditionTitle}</span>
            <span class="sub-divider">|</span>
            <span class="meta-sub-item">
              {$tStore('weather.feels_like_temp', {
                temp: `${displayFeelsLike.val}°${unitLetter}`
              })}
            </span>
          </div>

          <div class="banner-location">
            <MapPin size={11} class="loc-pin" />
            <span>{$weatherLocationStore.name || shortCityName}</span>
          </div>

          {#if !isPage}
            <button
              type="button"
              class="see-forecast-btn"
              onclick={(e) => {
                e.stopPropagation()
                isForecastOpen = !isForecastOpen
              }}
            >
              <span>
                {isForecastOpen
                  ? $tStore('weather.hide_forecast')
                  : $tStore('weather.see_forecast')}
              </span>
              <ChevronDown
                size={13}
                strokeWidth={2.4}
                class="chevron-icon {isForecastOpen ? 'open' : ''}"
              />
            </button>
          {/if}
        </div>

        <div class="banner-right-art">
          {#if !isPage && $pageVisibilityStore.weather !== false}
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
            <WeatherArt
              type={illustrationType}
              variant="hero"
              paused={!expanded}
            />
          </div>
        </div>
      </div>

      {#if !isPage && isForecastOpen}
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
              {@const forecastArtType = mapCodeToArt(f.weatherCode)}
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
  <div
    class="weather-minimal"
    style="--weather-color: {weatherInfo.theme.color}"
  >
    {#if loading && !weather}
      <div class="minimal-skeleton"></div>
    {:else if error && !weather}
      <WifiOff size={14} class="offline-icon" />
      <span class="temp-minimal">--</span>
    {:else}
      <WeatherArt type={illustrationType} variant="compact" />
      <span class="temp-minimal">{displayTemp.val}°</span>
    {/if}
  </div>
{:else if variant === 'page'}
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
          {@render bannerCard('page')}

          {#if weather.forecast && weather.forecast.length > 0}
            <section class="forecast-section">
              <BentoCard variant="bento" class="forecast-bento-card">
                <div class="forecast-bento-header">
                  <CalendarDays size={13} />
                  <span>{$tStore('weather.forecast_7day')}</span>
                </div>
                <div class="forecast-days-strip">
                  {#each weather.forecast as f, idx (f.day + idx)}
                    {@const forecastArtType = mapCodeToArt(f.weatherCode)}
                    {@const maxConv = convertTemp(f.tempMax, $weatherUnitStore)}
                    {@const minConv = convertTemp(f.tempMin, $weatherUnitStore)}
                    <div class="forecast-col" class:is-today={idx === 0}>
                      <span class="fc-day-title">{f.day}</span>
                      <div class="fc-art-badge">
                        <WeatherArt
                          type={forecastArtType}
                          variant="mini"
                          monochrome={false}
                        />
                      </div>
                      <div class="fc-temp-split">
                        <span class="fc-max">{maxConv.val}°</span>
                        <span class="fc-min">{minConv.val}°</span>
                      </div>
                    </div>
                  {/each}
                </div>
              </BentoCard>
            </section>
          {/if}

          <div class="bento-metrics-grid">
            <BentoCard variant="bento" class="forecast-bento-card">
              <div class="tile-head">
                <Wind size={13} class="tile-head-icon" />
                <span>{$tStore('weather.wind')}</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num">{weather.windSpeed}</span>
                <span class="tile-val-unit">{$tStore('weather.unit_kmh')}</span>
              </div>
              <div class="tile-sub">
                {weather.windDirectionLabel} • {weather.windDirection}°
              </div>
            </BentoCard>

            <BentoCard variant="bento" class="forecast-bento-card">
              <div class="tile-head">
                <Compass size={13} class="tile-head-icon" />
                <span>{$tStore('weather.uv_index')}</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num"
                  >{parseFloat(weather.uvIndex) || weather.uvIndex}</span
                >
                <span class="tile-val-unit">UV</span>
              </div>
              <div class="tile-sub">
                {getUvRating(weather.uvIndex, $tStore)}
              </div>
            </BentoCard>

            <BentoCard variant="bento" class="forecast-bento-card">
              <div class="tile-head">
                <Droplets size={13} class="tile-head-icon" />
                <span>{$tStore('weather.humidity')}</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num">{weather.humidity}</span>
                <span class="tile-val-unit">%</span>
              </div>
              <div class="tile-sub">
                {weather.humidity >= 70
                  ? $tStore('weather.humidity_high')
                  : weather.humidity <= 35
                    ? $tStore('weather.humidity_dry')
                    : $tStore('weather.humidity_comfortable')}
              </div>
            </BentoCard>

            <BentoCard variant="bento" class="forecast-bento-card">
              <div class="tile-head">
                <Gauge size={13} class="tile-head-icon" />
                <span>{$tStore('weather.pressure')}</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num">{weather.pressure}</span>
                <span class="tile-val-unit">{$tStore('weather.unit_hpa')}</span>
              </div>
              <div class="tile-sub">{$tStore('weather.atmospheric')}</div>
            </BentoCard>

            <BentoCard variant="bento" class="forecast-bento-card">
              <div class="tile-head">
                <Sun size={13} class="tile-head-icon" />
                <span>{$tStore('weather.sun_timing')}</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num time-val">{formattedSunrise}</span>
                <span class="tile-val-unit"
                  >{$tStore('weather.sunrise_short')}</span
                >
              </div>
              <div class="tile-sub">
                {$tStore('weather.sunset_time', { time: formattedSunset })}
              </div>
            </BentoCard>

            <BentoCard variant="bento" class="forecast-bento-card">
              <div class="tile-head">
                <CloudRain size={13} class="tile-head-icon" />
                <span>{$tStore('weather.precipitation')}</span>
              </div>
              <div class="tile-body">
                <span class="tile-val-num"
                  >{weather.precipitationProbability}</span
                >
                <span class="tile-val-unit">%</span>
              </div>
              <div class="tile-sub">{$tStore('weather.peak_today')}</div>
            </BentoCard>
          </div>
        </div>
      </ScrollView>

      <PageHeader position="bottom" title={$tStore('apps.weather_name')}>
        {#snippet actions()}
          <RefreshButton
            {loading}
            onclick={() => handleRetry()}
            title={$tStore('common.retry')}
          />
        {/snippet}
      </PageHeader>
    {/if}
  </div>
{:else}
  {@render bannerCard('widget')}
{/if}

<style src="./styles/Weather.css"></style>
