<script lang="ts">
  import { onMount } from 'svelte'
  import {
    Wind,
    Droplets,
    Compass,
    RefreshCw,
    WifiOff,
    MapPin
  } from 'lucide-svelte'
  import {
    weatherLocationStore,
    weatherUnitStore,
    simulateOfflineStore,
    type WeatherLocation,
    type TemperatureUnit
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

  let currentLocation = $state<WeatherLocation>({
    name: 'Manila',
    lat: 14.5995,
    lon: 120.9842
  })

  let currentUnit = $state<TemperatureUnit>('celsius')
  let offlineSim = $state<boolean>(false)
  let weather = $state<WeatherData | null>(null)
  let loading = $state<boolean>(true)
  let error = $state<boolean>(false)

  // Subscribe to persistent stores safely in onMount
  onMount(() => {
    const unsubLocation = weatherLocationStore.subscribe((loc) => {
      currentLocation = loc
      loadWeather()
    })
    const unsubUnit = weatherUnitStore.subscribe((unit) => {
      currentUnit = unit
    })
    const unsubOffline = simulateOfflineStore.subscribe((off) => {
      offlineSim = off
      loadWeather()
    })

    return () => {
      unsubLocation()
      unsubUnit()
      unsubOffline()
    }
  })

  $effect(() => {
    // Re-fetch when simulateOffline prop toggles in DevTest
    if (simulateOffline !== undefined) {
      loadWeather()
    }
  })

  async function loadWeather() {
    loading = true
    error = false

    try {
      const data = await fetchWeatherData(
        currentLocation.name,
        currentLocation.lat,
        currentLocation.lon,
        offlineSim || simulateOffline
      )

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

  // Reactive temperature conversions
  let displayTemp = $derived.by(() => {
    if (!weather) return { val: 0, symbol: '', str: '--' }
    return convertTemp(weather.temp, currentUnit)
  })

  let displayHigh = $derived.by(() => {
    if (!weather) return { val: 0, symbol: '', str: '--' }
    return convertTemp(weather.high, currentUnit)
  })

  let displayLow = $derived.by(() => {
    if (!weather) return { val: 0, symbol: '', str: '--' }
    return convertTemp(weather.low, currentUnit)
  })

  let weatherInfo = $derived.by(() => {
    if (!weather) return parseWmoCode(2)
    return parseWmoCode(weather.weatherCode)
  })
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
  <div
    class="weather-card"
    style="--weather-color: {weatherInfo.theme
      .color}; --weather-glow: {weatherInfo.theme
      .glow}; --weather-badge-bg: {weatherInfo.theme.badgeBg}"
  >
    <div class="glow-backdrop"></div>

    {#if loading && !weather}
      <div class="card-loading">
        <RefreshCw size={24} class="spin-icon" />
        <span>Loading Weather...</span>
      </div>
    {:else if error && !weather}
      <div class="card-error">
        <WifiOff size={32} class="offline-hero-icon" />
        <span class="error-title">Connection Offline</span>
        <span class="error-desc">Unable to fetch live weather data</span>
        <button class="retry-btn" onclick={loadWeather}>
          <RefreshCw size={12} /> Retry
        </button>
      </div>
    {:else if weather}
      {@const MainWeatherIcon = weatherInfo.icon}

      <!-- Top Main Overview Row -->
      <div class="top-row">
        <div class="main-left">
          <div class="location" title={currentLocation.name}>
            <MapPin size={12} class="location-pin-icon" />
            <span>{currentLocation.name}</span>
          </div>
          <div class="temp-group">
            <span class="temp-val"
              >{displayTemp.val}<span class="temp-unit-symbol"
                >{displayTemp.symbol}</span
              ></span
            >
            <div class="temp-details">
              <span class="hilo-badge"
                >H: {displayHigh.str} &bull; L: {displayLow.str}</span
              >
              <span
                class="condition-badge"
                class:offline-badge={weather.isOffline}
              >
                {weather.isOffline ? 'Offline (Cached)' : weather.condition}
              </span>
            </div>
          </div>
        </div>

        <div class="main-right">
          <div class="icon-wrapper">
            <MainWeatherIcon size={52} class="weather-icon-animated" />
          </div>
        </div>
      </div>

      <!-- Metrics Stats Bar -->
      <div class="metrics-bar">
        <div class="metric-item">
          <Droplets size={14} class="metric-icon" />
          <span class="metric-val">{weather.humidity}%</span>
          <span class="metric-lbl">Humidity</span>
        </div>
        <div class="metric-sep"></div>
        <div class="metric-item">
          <Wind size={14} class="metric-icon" />
          <span class="metric-val">{weather.windSpeed} km/h</span>
          <span class="metric-lbl">Wind</span>
        </div>
        <div class="metric-sep"></div>
        <div class="metric-item">
          <Compass size={14} class="metric-icon" />
          <span class="metric-val">{weather.uvIndex}</span>
          <span class="metric-lbl">UV</span>
        </div>
      </div>

      <!-- 4-Day Forecast Bar -->
      <div class="forecast-bar">
        {#each weather.forecast as f (f.day)}
          {@const ForecastIcon = f.icon}
          <div class="forecast-item">
            <span class="f-day">{f.day}</span>
            <ForecastIcon size={18} class="f-icon" />
            <span class="f-temp">{convertTemp(f.tempMax, currentUnit).str}</span
            >
          </div>
        {/each}
      </div>
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

  /* EXPANDED DESIGN */
  .weather-card {
    position: relative;
    width: 100%;
    height: 100%;
    padding: 0.9rem 1.1rem;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    overflow: hidden;
    color: var(--c-text-primary);
  }

  .glow-backdrop {
    position: absolute;
    top: -20px;
    right: -20px;
    width: 180px;
    height: 180px;
    background: radial-gradient(
      circle,
      var(--weather-glow, rgba(245, 158, 11, 0.15)) 0%,
      transparent 75%
    );
    filter: blur(24px);
    pointer-events: none;
    transition: background 0.5s ease;
  }

  .card-loading,
  .card-error {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: var(--c-text-muted);
    font-size: 0.85rem;
    z-index: 2;
  }

  .error-title {
    font-weight: 700;
    color: var(--c-text-primary);
    font-size: 0.95rem;
  }

  .error-desc {
    font-size: 0.75rem;
    color: var(--c-text-muted);
  }

  .retry-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-top: 6px;
    padding: 6px 14px;
    border-radius: 12px;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .retry-btn:hover {
    background: var(--c-hover);
  }

  :global(.spin-icon) {
    animation: spin 1.2s linear infinite;
    color: var(--c-primary);
  }

  @keyframes spin {
    100% {
      transform: rotate(360deg);
    }
  }

  :global(.offline-hero-icon) {
    color: var(--c-text-muted);
  }

  .top-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    z-index: 2;
  }

  .main-left {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }

  .location {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--c-text-muted);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    max-width: 170px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :global(.location-pin-icon) {
    color: var(--weather-color, var(--c-primary));
    flex-shrink: 0;
    transition: color 0.3s ease;
  }

  .temp-group {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 2px;
  }

  .temp-val {
    font-size: 3.2rem;
    font-weight: 800;
    line-height: 1;
    letter-spacing: -1.8px;
    color: var(--c-text-primary);
  }

  .temp-unit-symbol {
    font-size: 1.6rem;
    font-weight: 600;
    margin-left: 2px;
    opacity: 0.85;
  }

  .temp-details {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 3px;
  }

  .hilo-badge {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--c-text-muted);
    letter-spacing: 0.02em;
  }

  .condition-badge {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--weather-color, var(--c-warning));
    background: var(--weather-badge-bg, var(--c-item-bg));
    padding: 3px 10px;
    border-radius: 12px;
    border: 1px solid var(--weather-color, var(--c-border));
    transition: all 0.3s ease;
  }

  .offline-badge {
    color: var(--c-text-muted);
    border-style: dashed;
  }

  .main-right {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .icon-wrapper {
    color: var(--weather-color, var(--c-warning));
    filter: drop-shadow(
      0 4px 14px var(--weather-glow, rgba(245, 158, 11, 0.4))
    );
    transition: all 0.3s ease;
  }

  :global(.weather-icon-animated) {
    animation: float 4s ease-in-out infinite alternate;
  }

  @keyframes float {
    0% {
      transform: translateY(0px) rotate(0deg);
    }
    100% {
      transform: translateY(-4px) rotate(3deg);
    }
  }

  /* Metrics Bar */
  .metrics-bar {
    display: flex;
    align-items: center;
    justify-content: space-around;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    border-radius: 12px;
    padding: 8px 10px;
    z-index: 2;
  }

  .metric-item {
    display: flex;
    align-items: center;
    gap: 5px;
  }

  :global(.metric-icon) {
    color: var(--c-primary);
  }

  .metric-val {
    font-size: 0.78rem;
    font-weight: 700;
    color: var(--c-text-primary);
  }

  .metric-lbl {
    font-size: 0.68rem;
    color: var(--c-text-muted);
  }

  .metric-sep {
    width: 1px;
    height: 14px;
    background: var(--c-border);
  }

  /* Forecast Bar */
  .forecast-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    z-index: 2;
  }

  .forecast-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 5px;
    padding: 8px 4px;
    background: var(--c-item-bg);
    border-radius: 12px;
    border: 1px solid var(--c-border);
  }

  .f-day {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--c-text-muted);
  }

  :global(.f-icon) {
    color: var(--c-warning);
  }

  .f-temp {
    font-size: 0.78rem;
    font-weight: 700;
    color: var(--c-text-primary);
  }
</style>
