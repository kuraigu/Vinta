<script lang="ts">
  import { onMount } from 'svelte'
  import {
    weatherLocationStore,
    weatherUnitStore,
    timezoneStore,
    simulateOfflineStore,
    type WeatherLocation,
    type TemperatureUnit
  } from '../store'
  import { Events } from '@wailsio/runtime'
  import {
    searchCities,
    type GeocodingResult
  } from '../services/weatherService'
  import {
    Search,
    MapPin,
    WifiOff,
    Thermometer,
    Globe,
    Trash2
  } from 'lucide-svelte'

  function broadcastWeatherEffect() {
    const location = $weatherLocationStore
    const unit = $weatherUnitStore
    const timezone = $timezoneStore
    const simulateOffline = $simulateOfflineStore
    try {
      Events.Emit('devtest-weather', {
        location,
        unit,
        timezone,
        simulateOffline
      })
    } catch {}
  }
  $effect(broadcastWeatherEffect)

  let currentLocation = $state<WeatherLocation>({
    name: 'Manila',
    lat: 14.5995,
    lon: 120.9842
  })

  let currentUnit = $state<TemperatureUnit>('celsius')
  let isOfflineSim = $state<boolean>(false)
  let searchQuery = $state('')
  let searchResults = $state<GeocodingResult[]>([])

  type PresetCity = WeatherLocation & { timezone: string }

  const PRESET_CITIES: PresetCity[] = [
    { name: 'Manila', lat: 14.5995, lon: 120.9842, timezone: 'Asia/Manila' },
    { name: 'Tokyo', lat: 35.6762, lon: 139.6503, timezone: 'Asia/Tokyo' },
    {
      name: 'New York',
      lat: 40.7128,
      lon: -74.006,
      timezone: 'America/New_York'
    },
    { name: 'London', lat: 51.5074, lon: -0.1278, timezone: 'Europe/London' },
    {
      name: 'Sydney',
      lat: -33.8688,
      lon: 151.2093,
      timezone: 'Australia/Sydney'
    },
    { name: 'Paris', lat: 48.8566, lon: 2.3522, timezone: 'Europe/Paris' }
  ]

  onMount(() => {
    const unsubLoc = weatherLocationStore.subscribe((loc) => {
      currentLocation = loc
    })
    const unsubUnit = weatherUnitStore.subscribe((unit) => {
      currentUnit = unit
    })
    const unsubOff = simulateOfflineStore.subscribe((off) => {
      isOfflineSim = off
    })
    return () => {
      unsubLoc()
      unsubUnit()
      unsubOff()
    }
  })

  function selectCity(city: WeatherLocation, tz?: string) {
    weatherLocationStore.set(city)
    if (tz) {
      timezoneStore.set(tz)
    }
    searchQuery = ''
    searchResults = []
  }

  function toggleUnit(unit: TemperatureUnit) {
    weatherUnitStore.set(unit)
  }

  function toggleOfflineMode() {
    simulateOfflineStore.update((v) => !v)
  }

  function clearWeatherCache() {
    Object.keys(localStorage).forEach((key) => {
      if (key.startsWith('vinta_weather_cache_')) {
        localStorage.removeItem(key)
      }
    })
    // Re-trigger location update to reload component state instantly
    weatherLocationStore.set({ ...currentLocation })
  }

  async function handleSearch() {
    if (!searchQuery || searchQuery.trim().length < 2) {
      searchResults = []
      return
    }
    searchResults = await searchCities(searchQuery)
  }
</script>

<div class="weather-dev-controls">
  <div class="control-header">
    <MapPin size={14} class="header-icon" />
    <span>Home &amp; Weather Controls</span>
  </div>

  <!-- Consolidated Location & Timezone Presets -->
  <div class="preset-row">
    <div class="section-title">
      <Globe size={12} />
      <span class="label">Location &amp; Timezone Presets:</span>
    </div>
    <div class="preset-pills">
      {#each PRESET_CITIES as city (city.name)}
        <button
          class="city-pill"
          class:active={currentLocation.name.startsWith(city.name)}
          onclick={() => selectCity(city, city.timezone)}
        >
          {city.name}
        </button>
      {/each}
    </div>
  </div>

  <!-- Search Custom City -->
  <div class="search-row">
    <div class="search-input-wrapper">
      <Search size={13} class="search-icon" />
      <input
        type="text"
        placeholder="Search city name &amp; timezone..."
        bind:value={searchQuery}
        oninput={handleSearch}
        class="search-input"
      />
    </div>

    {#if searchResults.length > 0}
      <div class="search-dropdown">
        {#each searchResults as res (res.latitude + res.longitude)}
          <button
            class="dropdown-item"
            onclick={() =>
              selectCity(
                {
                  name: `${res.name}${res.country ? `, ${res.country}` : ''}`,
                  lat: res.latitude,
                  lon: res.longitude
                },
                res.timezone
              )}
          >
            {res.name}
            {res.admin1 ? `(${res.admin1})` : ''}
            {res.country ? `- ${res.country}` : ''}
            {res.timezone ? `[${res.timezone}]` : ''}
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Unit & Offline Controls -->
  <div class="toggles-row">
    <div class="unit-toggle">
      <Thermometer size={13} class="toggle-icon" />
      <span class="label">Unit:</span>
      <button
        class="unit-btn"
        class:active={currentUnit === 'celsius'}
        onclick={() => toggleUnit('celsius')}
      >
        °C
      </button>
      <button
        class="unit-btn"
        class:active={currentUnit === 'fahrenheit'}
        onclick={() => toggleUnit('fahrenheit')}
      >
        °F
      </button>
      <button
        class="unit-btn"
        class:active={currentUnit === 'kelvin'}
        onclick={() => toggleUnit('kelvin')}
        title="Kelvin (Nerd Mode)"
      >
        K
      </button>
    </div>

    <div class="dev-btn-group">
      <button
        class="offline-toggle-btn"
        class:active={isOfflineSim}
        onclick={toggleOfflineMode}
        title="Simulate Disconnected Network"
      >
        <WifiOff size={12} />
        <span>{isOfflineSim ? 'Offline Mode (On)' : 'Simulate Offline'}</span>
      </button>

      <button
        class="clear-cache-btn"
        onclick={clearWeatherCache}
        title="Clear Local Weather Cache to test Zero-Cache Error View"
      >
        <Trash2 size={12} />
        <span>Clear Cache</span>
      </button>
    </div>
  </div>
</div>

<style>
  .weather-dev-controls {
    display: flex;
    flex-direction: column;
    gap: 10px;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    border-radius: 14px;
    padding: 12px;
    color: var(--c-text-primary);
    font-size: 0.78rem;
  }

  .control-header {
    display: flex;
    align-items: center;
    gap: 6px;
    font-weight: 700;
    color: var(--c-amber, #f59e0b);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-size: 0.72rem;
  }

  .label {
    font-size: 0.72rem;
    color: var(--c-text-muted);
    font-weight: 600;
  }

  .section-title {
    display: flex;
    align-items: center;
    gap: 4px;
    color: var(--c-primary);
  }

  .preset-row {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .preset-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .city-pill {
    padding: 4px 10px;
    border-radius: 12px;
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    font-size: 0.72rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .city-pill.active {
    background: var(--c-primary);
    color: #ffffff;
    border-color: var(--c-primary);
  }

  .search-row {
    position: relative;
  }

  .search-input-wrapper {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    border-radius: 10px;
    padding: 6px 10px;
  }

  .search-input {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: var(--c-text-primary);
    font-size: 0.75rem;
  }

  .search-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    margin-top: 4px;
    background: var(--c-dock-popover-bg, var(--c-bg));
    border: 1px solid var(--c-border);
    border-radius: 10px;
    overflow: hidden;
    z-index: 10;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .dropdown-item {
    width: 100%;
    text-align: left;
    padding: 6px 10px;
    background: transparent;
    border: none;
    color: var(--c-text-primary);
    font-size: 0.72rem;
    cursor: pointer;
    text-overflow: ellipsis;
    overflow: hidden;
    white-space: nowrap;
  }

  .dropdown-item:hover {
    background: var(--c-hover);
  }

  .toggles-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 2px;
  }

  .unit-toggle {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .unit-btn {
    padding: 3px 8px;
    border-radius: 8px;
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    font-size: 0.72rem;
    font-weight: 600;
    cursor: pointer;
  }

  .unit-btn.active {
    background: var(--c-primary);
    color: #ffffff;
    border-color: var(--c-primary);
  }

  .offline-toggle-btn {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 4px 10px;
    border-radius: 10px;
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-muted);
    font-size: 0.72rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .offline-toggle-btn.active {
    background: var(--c-warning);
    color: #000000;
    border-color: var(--c-warning);
  }

  .dev-btn-group {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .clear-cache-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    border-radius: 10px;
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-muted);
    font-size: 0.72rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .clear-cache-btn:hover {
    background: var(--c-hover);
    color: var(--c-danger, #ef4444);
  }
</style>
