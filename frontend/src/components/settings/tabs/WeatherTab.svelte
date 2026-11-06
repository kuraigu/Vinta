<script lang="ts">
  import { onMount } from 'svelte'
  import { MapPin, Thermometer } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import { weatherLocationStore, weatherUnitStore } from '../../../store'
  import { searchCities, formatLocationName } from '../../../services/weatherService'
  import {
    GLOBAL_WEATHER_PRESETS,
    detectAutoWeatherLocation,
    type WeatherLocationPreset
  } from '../../../services/weatherPresets'

  const RECENTS_STORAGE_KEY = 'vinta_weather_recents'

  let recentLocations = $state<WeatherLocationPreset[]>([])

  onMount(() => {
    try {
      const raw = localStorage.getItem(RECENTS_STORAGE_KEY)
      if (raw) {
        recentLocations = JSON.parse(raw)
      }
    } catch {
      // Ignore storage errors
    }
  })

  function saveRecentLocation(loc: WeatherLocationPreset) {
    if (loc.name.startsWith('Auto-detect')) return
    const filtered = recentLocations.filter((item) => item.name !== loc.name)
    const updated = [loc, ...filtered].slice(0, 5)
    recentLocations = updated
    try {
      localStorage.setItem(RECENTS_STORAGE_KEY, JSON.stringify(updated))
    } catch {
      // Ignore storage errors
    }
  }

  function buildPresetOptions(): { label: string; value: string; payload: WeatherLocationPreset }[] {
    const list: { label: string; value: string; payload: WeatherLocationPreset }[] = []
    const seen = new Set<string>()

    // 1. Auto-detect first
    const autoPreset = GLOBAL_WEATHER_PRESETS[0]
    list.push({
      label: autoPreset.name,
      value: autoPreset.name,
      payload: autoPreset
    })
    seen.add(autoPreset.name)

    // 2. User recents (if any)
    for (const r of recentLocations) {
      if (!seen.has(r.name)) {
        list.push({ label: r.name, value: r.name, payload: r })
        seen.add(r.name)
      }
    }

    // 3. Current active location if not already in presets
    const currentName = $weatherLocationStore.name
    if (currentName && !seen.has(currentName) && !currentName.startsWith('Auto-detect')) {
      list.push({
        label: currentName,
        value: currentName,
        payload: {
          name: currentName,
          lat: $weatherLocationStore.lat,
          lon: $weatherLocationStore.lon
        }
      })
      seen.add(currentName)
    }

    // 4. Global curated presets
    for (let i = 1; i < GLOBAL_WEATHER_PRESETS.length; i++) {
      const p = GLOBAL_WEATHER_PRESETS[i]
      if (!seen.has(p.name)) {
        list.push({ label: p.name, value: p.name, payload: p })
        seen.add(p.name)
      }
    }

    return list
  }

  let weatherOptions = $state(buildPresetOptions())

  let selectedWeatherCity = $derived($weatherLocationStore.name)

  async function handleAutoDetect() {
    const loc = await detectAutoWeatherLocation()
    $weatherLocationStore = {
      name: loc.name,
      lat: loc.lat,
      lon: loc.lon
    }
  }

  async function handleWeatherSearch(query: string) {
    if (!query || query.trim().length < 2) {
      weatherOptions = buildPresetOptions()
      return
    }

    const results = await searchCities(query)
    if (results && results.length > 0) {
      const searchedOpts = results.map((r) => {
        const displayName = formatLocationName(r)
        return {
          label: displayName,
          value: displayName,
          payload: { name: displayName, lat: r.latitude, lon: r.longitude }
        }
      })

      const existingValues = new Set(searchedOpts.map((o) => o.value))
      const extraPresets = GLOBAL_WEATHER_PRESETS
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
    const p = payload as WeatherLocationPreset
    if (p) {
      if (p.name.startsWith('Auto-detect')) {
        handleAutoDetect()
      } else {
        $weatherLocationStore = { name: p.name, lat: p.lat, lon: p.lon }
        saveRecentLocation(p)
      }
    }
  }
</script>

<div class="setting-group">
  <SettingItem
    label="Weather Location"
    description="Target location for forecasts (Default: Auto-detect)"
    icon={MapPin}
    flexCol={true}
  >
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

  <SettingItem
    label="Temperature Unit"
    description="Measurement unit for temperatures (Default: Celsius)"
    icon={Thermometer}
    flexCol={true}
  >
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
    border: 1px solid var(--vinta-border);
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
    color: var(--vinta-text-secondary);
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
    color: var(--vinta-text-primary);
    font-weight: 600;
  }
</style>
