<script lang="ts">
  import { MapPin, Thermometer } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import { weatherLocationStore, weatherUnitStore } from '../../../store'
  import { searchCities } from '../../../services/weatherService'

  const baseWeatherPresets = [
    { name: 'Auto-detect (System IP)', lat: 14.5995, lon: 120.9842 },
    { name: 'Manila, Philippines', lat: 14.5995, lon: 120.9842 },
    { name: 'Quezon City, Philippines', lat: 14.676, lon: 121.0437 },
    { name: 'Cebu City, Philippines', lat: 10.3157, lon: 123.8854 },
    { name: 'Davao City, Philippines', lat: 7.1907, lon: 125.4553 },
    { name: 'Baguio, Philippines', lat: 16.4023, lon: 120.596 },
    { name: 'Singapore', lat: 1.3521, lon: 103.8198 },
    { name: 'Bangkok, Thailand', lat: 13.7563, lon: 100.5018 },
    { name: 'Kuala Lumpur, Malaysia', lat: 3.139, lon: 101.6869 },
    { name: 'Jakarta, Indonesia', lat: -6.2088, lon: 106.8456 },
    { name: 'Tokyo, Japan', lat: 35.6762, lon: 139.6503 },
    { name: 'Seoul, South Korea', lat: 37.5665, lon: 126.978 },
    { name: 'Beijing, China', lat: 39.9042, lon: 116.4074 },
    { name: 'Shanghai, China', lat: 31.2304, lon: 121.4737 },
    { name: 'Hong Kong', lat: 22.3193, lon: 114.1694 },
    { name: 'New York, United States', lat: 40.7128, lon: -74.006 },
    { name: 'Los Angeles, United States', lat: 34.0522, lon: -118.2437 },
    { name: 'London, United Kingdom', lat: 51.5074, lon: -0.1278 },
    { name: 'Paris, France', lat: 48.8566, lon: 2.3522 },
    { name: 'Sydney, Australia', lat: -33.8688, lon: 151.2093 }
  ]

  let weatherOptions = $state(
    baseWeatherPresets.map((p) => ({
      label: p.name,
      value: p.name,
      payload: p
    }))
  )

  let selectedWeatherCity = $derived($weatherLocationStore.name)

  async function detectIpLocation() {
    try {
      const res = await fetch('https://ipapi.co/json/')
      if (res.ok) {
        const data = await res.json()
        if (data.city && data.latitude && data.longitude) {
          $weatherLocationStore = {
            name: `${data.city}, ${data.country_name || 'Auto'}`,
            lat: data.latitude,
            lon: data.longitude
          }
          return
        }
      }
    } catch {
      // Fallback
    }
    $weatherLocationStore = {
      name: 'Manila, Philippines',
      lat: 14.5995,
      lon: 120.9842
    }
  }

  async function handleWeatherSearch(query: string) {
    if (!query || query.trim().length < 2) {
      weatherOptions = baseWeatherPresets.map((p) => ({
        label: p.name,
        value: p.name,
        payload: p
      }))
      return
    }

    const results = await searchCities(query)
    if (results && results.length > 0) {
      const searchedOpts = results.map((r) => ({
        label: r.name,
        value: r.name,
        payload: { name: r.name, lat: r.latitude, lon: r.longitude }
      }))

      const existingValues = new Set(searchedOpts.map((o) => o.value))
      const extraPresets = baseWeatherPresets
        .filter((p) => !existingValues.has(p.name))
        .map((p) => ({
          label: p.name,
          value: p.name,
          payload: p
        }))

      weatherOptions = [...searchedOpts, ...extraPresets]
    }
  }

  function handleSelectWeatherPayload(payload: unknown) {
    const p = payload as { name: string; lat: number; lon: number }
    if (p) {
      if (p.name.startsWith('Auto-detect')) {
        detectIpLocation()
      } else {
        $weatherLocationStore = { name: p.name, lat: p.lat, lon: p.lon }
      }
    }
  }
</script>

<div class="setting-group">
  <SettingItem label="Weather Location" icon={MapPin} flexCol={true}>
    <CustomSelect
      options={weatherOptions}
      value={selectedWeatherCity}
      searchable={true}
      fullWidth={true}
      direction="auto"
      onSearch={handleWeatherSearch}
      onSelectPayload={handleSelectWeatherPayload}
    />
  </SettingItem>

  <SettingItem label="Temperature Unit" icon={Thermometer} flexCol={true}>
    <div class="segment-control full-width">
      <button
        class:active={$weatherUnitStore === 'celsius'}
        onclick={() => ($weatherUnitStore = 'celsius')}>°C (Celsius)</button
      >
      <button
        class:active={$weatherUnitStore === 'fahrenheit'}
        onclick={() => ($weatherUnitStore = 'fahrenheit')}
        >°F (Fahrenheit)</button
      >
    </div>
  </SettingItem>
</div>

<style>
  .setting-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .segment-control {
    display: flex;
    background: rgba(255, 255, 255, 0.06);
    padding: 2px;
    border-radius: 8px;
    border: 1px solid var(--c-border);
  }

  .segment-control.full-width {
    width: 100%;
    box-sizing: border-box;
    height: 32px;
    align-items: center;
  }

  .segment-control button {
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    padding: 4px 10px;
    font-size: 11px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .segment-control.full-width button {
    flex: 1;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .segment-control button.active {
    background: rgba(255, 255, 255, 0.12);
    color: var(--c-text-primary);
    font-weight: 600;
  }
</style>
