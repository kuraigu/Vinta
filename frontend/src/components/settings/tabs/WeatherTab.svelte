<!--
  @component WeatherTab
  @description Weather configuration tab. Manages location presets, coordinates custom city
  searches, toggles Celsius vs. Fahrenheit, and provides offline testing mock data.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import { SvelteSet } from 'svelte/reactivity'
  import { MapPin, Thermometer } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import SegmentedControl from '../../SegmentedControl.svelte'
  import SettingItem from '../SettingItem.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import {
    weatherLocationStore,
    weatherUnitStore,
    tStore
  } from '../../../store'
  import {
    searchCities,
    formatLocationName
  } from '../../../services/weatherService'
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

  function buildPresetOptions(): {
    label: string
    value: string
    payload: WeatherLocationPreset
  }[] {
    const list: {
      label: string
      value: string
      payload: WeatherLocationPreset
    }[] = []
    const seen = new SvelteSet<string>()

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
    if (
      currentName &&
      !seen.has(currentName) &&
      !currentName.startsWith('Auto-detect')
    ) {
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
  let searchSeq = 0

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
    const seq = ++searchSeq
    if (!query || query.trim().length < 2) {
      weatherOptions = buildPresetOptions()
      return
    }

    const results = await searchCities(query)
    if (seq !== searchSeq) return
    if (results && results.length > 0) {
      const searchedOpts = results.map((r) => {
        const displayName = formatLocationName(r)
        return {
          label: displayName,
          value: displayName,
          payload: { name: displayName, lat: r.latitude, lon: r.longitude }
        }
      })

      const existingValues = new SvelteSet(searchedOpts.map((o) => o.value))
      const extraPresets = GLOBAL_WEATHER_PRESETS.filter(
        (p) => !existingValues.has(p.name)
      ).map((p) => ({
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
  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.weather_location_label')}
      description={$tStore('settings.weather_location_desc')}
      icon={MapPin}
      flexCol={true}
      isDefault={selectedWeatherCity === weatherLocationStore.defaultValue.name}
      onReset={() => weatherLocationStore.reset()}
    >
      <CustomSelect
        options={weatherOptions}
        value={selectedWeatherCity}
        defaultValue={GLOBAL_WEATHER_PRESETS[0]?.name}
        searchable={true}
        fullWidth={true}
        direction="auto"
        onSearch={handleWeatherSearch}
        onSelectPayload={handleSelectWeatherPayload}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.weather_unit_label')}
      description={$tStore('settings.weather_unit_desc')}
      icon={Thermometer}
      flexCol={true}
      isDefault={$weatherUnitStore === weatherUnitStore.defaultValue}
      onReset={() => weatherUnitStore.reset()}
    >
      <SegmentedControl
        fullWidth={true}
        options={[
          { value: 'celsius', label: $tStore('settings.weather_unit_c') },
          { value: 'fahrenheit', label: $tStore('settings.weather_unit_f') }
        ]}
        bind:value={$weatherUnitStore}
      />
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/WeatherTab.css"></style>
