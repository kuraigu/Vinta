<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 440, h: 370, r: 30 },
    borderColor: 'var(--c-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import {
    themeStore,
    shapeStore,
    homeAutoHideStore,
    mediaAutoHideStore,
    notifAutoHideStore,
    autoHideDelayStore,
    autoHideDisplayStore,
    wakeMethodStore,
    dndStore,
    notifAutoWakeStore,
    notifData,
    timeFormatStore,
    timezoneStore,
    dateFormatStore,
    weatherLocationStore,
    weatherUnitStore,
    opacityMinimalStore,
    opacityExpandedStore,
    opacityHiddenStore,
    hiddenPillOffsetXStore,
    allowDragPositionStore,
    hardwareAccelStore,
    notifSoundStore,
    notifVolumeStore,
    glowIntensityStore,
    autostartStore,
    getInitialTimezone
  } from '../store'
  import ScrollView from '../components/ScrollView.svelte'
  import ComponentGallery from '../components/ComponentGallery.svelte'
  import CustomSelect from '../components/CustomSelect.svelte'
  import BentoItem from '../components/BentoItem.svelte'
  import {
    Palette,
    Sliders,
    Bell,
    Clock,
    CloudSun,
    Trash2,
    Moon,
    Maximize2,
    EyeOff,
    Timer,
    Eye,
    MousePointer,
    BellOff,
    BellRing,
    Calendar,
    Globe,
    Thermometer,
    MapPin,
    Sparkles,
    Volume2,
    RotateCcw,
    Power,
    Cpu
  } from 'lucide-svelte'

  import { searchCities } from '../services/weatherService'

  // Full standard IANA timezones (400+ world timezones)
  const rawTimezones =
    typeof Intl !== 'undefined' && typeof Intl.supportedValuesOf === 'function'
      ? Intl.supportedValuesOf('timeZone')
      : [
          'UTC',
          'Asia/Manila',
          'Asia/Tokyo',
          'Asia/Singapore',
          'Asia/Seoul',
          'Asia/Hong_Kong',
          'Asia/Dubai',
          'Asia/Bangkok',
          'Europe/London',
          'Europe/Paris',
          'Europe/Berlin',
          'Europe/Rome',
          'America/New_York',
          'America/Chicago',
          'America/Denver',
          'America/Los_Angeles',
          'America/Toronto',
          'America/Sao_Paulo',
          'Australia/Sydney'
        ]

  const timezoneOptions = [
    { label: 'Auto-detect (System Local)', value: 'auto' },
    ...rawTimezones.map((tz) => ({
      label: tz.replace(/_/g, ' '),
      value: tz
    }))
  ]

  const baseWeatherPresets = [
    { name: 'Auto-detect (System IP)', lat: 14.5995, lon: 120.9842 },
    // Philippines & Southeast Asia
    { name: 'Manila, Philippines', lat: 14.5995, lon: 120.9842 },
    { name: 'Quezon City, Philippines', lat: 14.676, lon: 121.0437 },
    { name: 'Cebu City, Philippines', lat: 10.3157, lon: 123.8854 },
    { name: 'Davao City, Philippines', lat: 7.1907, lon: 125.4553 },
    { name: 'Baguio, Philippines', lat: 16.4023, lon: 120.596 },
    { name: 'Singapore', lat: 1.3521, lon: 103.8198 },
    { name: 'Bangkok, Thailand', lat: 13.7563, lon: 100.5018 },
    { name: 'Kuala Lumpur, Malaysia', lat: 3.139, lon: 101.6869 },
    { name: 'Jakarta, Indonesia', lat: -6.2088, lon: 106.8456 },
    { name: 'Bali, Indonesia', lat: -8.4095, lon: 115.1889 },
    { name: 'Hanoi, Vietnam', lat: 21.0285, lon: 105.8542 },
    { name: 'Ho Chi Minh City, Vietnam', lat: 10.8231, lon: 106.6297 },

    // East & South Asia
    { name: 'Tokyo, Japan', lat: 35.6762, lon: 139.6503 },
    { name: 'Osaka, Japan', lat: 34.6937, lon: 135.5023 },
    { name: 'Kyoto, Japan', lat: 35.0116, lon: 135.7681 },
    { name: 'Seoul, South Korea', lat: 37.5665, lon: 126.978 },
    { name: 'Busan, South Korea', lat: 35.1796, lon: 129.0756 },
    { name: 'Beijing, China', lat: 39.9042, lon: 116.4074 },
    { name: 'Shanghai, China', lat: 31.2304, lon: 121.4737 },
    { name: 'Hong Kong', lat: 22.3193, lon: 114.1694 },
    { name: 'Taipei, Taiwan', lat: 25.033, lon: 121.5654 },
    { name: 'Mumbai, India', lat: 19.076, lon: 72.8777 },
    { name: 'New Delhi, India', lat: 28.6139, lon: 77.209 },

    // Americas
    { name: 'New York, United States', lat: 40.7128, lon: -74.006 },
    { name: 'Los Angeles, United States', lat: 34.0522, lon: -118.2437 },
    { name: 'Chicago, United States', lat: 41.8781, lon: -87.6298 },
    { name: 'Houston, United States', lat: 29.7604, lon: -95.3698 },
    { name: 'San Francisco, United States', lat: 37.7749, lon: -122.4194 },
    { name: 'Seattle, United States', lat: 47.6062, lon: -122.3321 },
    { name: 'Miami, United States', lat: 25.7617, lon: -80.1918 },
    { name: 'Las Vegas, United States', lat: 36.1699, lon: -115.1398 },
    { name: 'Toronto, Canada', lat: 43.6532, lon: -79.3832 },
    { name: 'Vancouver, Canada', lat: 49.2827, lon: -123.1207 },
    { name: 'Montreal, Canada', lat: 45.5017, lon: -73.5673 },
    { name: 'Mexico City, Mexico', lat: 19.4326, lon: -99.1332 },
    { name: 'São Paulo, Brazil', lat: -23.5505, lon: -46.6333 },
    { name: 'Rio de Janeiro, Brazil', lat: -22.9068, lon: -43.1729 },
    { name: 'Buenos Aires, Argentina', lat: -34.6037, lon: -58.3816 },

    // Europe
    { name: 'London, United Kingdom', lat: 51.5074, lon: -0.1278 },
    { name: 'Paris, France', lat: 48.8566, lon: 2.3522 },
    { name: 'Berlin, Germany', lat: 52.52, lon: 13.405 },
    { name: 'Munich, Germany', lat: 48.1351, lon: 11.582 },
    { name: 'Frankfurt, Germany', lat: 50.1109, lon: 8.6821 },
    { name: 'Rome, Italy', lat: 41.9028, lon: 12.4964 },
    { name: 'Milan, Italy', lat: 45.4642, lon: 9.19 },
    { name: 'Madrid, Spain', lat: 40.4168, lon: -3.7038 },
    { name: 'Barcelona, Spain', lat: 41.3851, lon: 2.1734 },
    { name: 'Amsterdam, Netherlands', lat: 52.3676, lon: 4.9041 },
    { name: 'Vienna, Austria', lat: 48.2082, lon: 16.3738 },
    { name: 'Zurich, Switzerland', lat: 47.3769, lon: 8.5417 },
    { name: 'Stockholm, Sweden', lat: 59.3293, lon: 18.0686 },
    { name: 'Oslo, Norway', lat: 59.9139, lon: 10.7522 },
    { name: 'Copenhagen, Denmark', lat: 55.6761, lon: 12.5683 },
    { name: 'Dublin, Ireland', lat: 53.3498, lon: -6.2603 },
    { name: 'Lisbon, Portugal', lat: 38.7223, lon: -9.1393 },
    { name: 'Athens, Greece', lat: 37.9838, lon: 23.7275 },
    { name: 'Istanbul, Turkey', lat: 41.0082, lon: 28.9784 },
    { name: 'Prague, Czechia', lat: 50.0755, lon: 14.4378 },
    { name: 'Warsaw, Poland', lat: 52.2297, lon: 21.0122 },

    // Oceania, Middle East & Africa
    { name: 'Sydney, Australia', lat: -33.8688, lon: 151.2093 },
    { name: 'Melbourne, Australia', lat: -37.8136, lon: 144.9631 },
    { name: 'Auckland, New Zealand', lat: -36.8485, lon: 174.7633 },
    { name: 'Dubai, United Arab Emirates', lat: 25.2048, lon: 55.2708 },
    { name: 'Abu Dhabi, United Arab Emirates', lat: 24.4539, lon: 54.3773 },
    { name: 'Riyadh, Saudi Arabia', lat: 24.7136, lon: 46.6753 },
    { name: 'Doha, Qatar', lat: 25.2854, lon: 51.531 },
    { name: 'Cairo, Egypt', lat: 30.0444, lon: 31.2357 },
    { name: 'Cape Town, South Africa', lat: -33.9249, lon: 18.4241 }
  ]

  let dynamicSearchResults = $state<
    { value: string; label: string; payload: (typeof baseWeatherPresets)[0] }[]
  >([])

  let weatherOptions = $derived([
    ...baseWeatherPresets.map((p) => ({
      value: p.name,
      label: p.name,
      payload: p
    })),
    ...dynamicSearchResults.filter(
      (ds) => !baseWeatherPresets.some((bp) => bp.name === ds.value)
    )
  ])

  async function detectIpLocation() {
    try {
      const res = await fetch('https://ipapi.co/json/')
      if (res.ok) {
        const data = await res.json()
        if (data.latitude && data.longitude) {
          $weatherLocationStore = {
            name: `Auto-detect (${data.city || 'IP'})`,
            lat: data.latitude,
            lon: data.longitude
          }
          if (data.timezone) {
            $timezoneStore = data.timezone
          }
          return
        }
      }
    } catch {
      // Fallback
    }
    $weatherLocationStore = baseWeatherPresets[0]
  }

  async function handleWeatherSearch(query: string) {
    if (!query || query.trim().length < 2) {
      dynamicSearchResults = []
      return
    }
    const results = await searchCities(query)
    dynamicSearchResults = results.map((r) => {
      const cityLabel = `${r.name}${r.country ? `, ${r.country}` : ''}`
      return {
        value: cityLabel,
        label: `${cityLabel} ${r.timezone ? `[${r.timezone}]` : ''}`,
        payload: {
          name: cityLabel,
          lat: r.latitude,
          lon: r.longitude
        }
      }
    })
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

  import {
    SetAutostart,
    SetHardwareAcceleration
  } from '../../wailsjs/go/backend/App.js'

  const themeOptions = [
    { label: 'Dark Mode', value: 'dark' },
    { label: 'Light Mode', value: 'light' }
  ]

  const glowIntensityOptions = [
    { label: 'Disabled', value: 'disabled' },
    { label: 'Subtle', value: 'subtle' },
    { label: 'Vibrant Glow', value: 'vibrant' }
  ]

  $effect(() => {
    SetAutostart($autostartStore).catch(() => {})
  })

  let selectedTz = $state(
    $timezoneStore === getInitialTimezone() ? 'auto' : $timezoneStore
  )

  $effect(() => {
    if (selectedTz === 'auto') {
      $timezoneStore = getInitialTimezone()
    } else {
      $timezoneStore = selectedTz
    }
  })

  let selectedWeatherCity = $derived($weatherLocationStore.name)

  function clearAllNotifications() {
    notifData.set([])
  }
</script>

{#snippet appearanceSlide()}
  <div class="card-wrapper">
    <BentoItem>
      <ScrollView class="section-content">
        <div class="card-category-header">
          <div class="category-icon-badge purple">
            <Palette size={15} />
          </div>
          <div class="category-text">
            <span class="category-name">Appearance</span>
            <span class="category-desc">Visual theme and island shape</span>
          </div>
        </div>

        <div class="setting-item flex-col">
          <div class="setting-label">
            <Moon size={14} class="item-icon" />
            <span>Theme</span>
          </div>
          <CustomSelect
            options={themeOptions}
            bind:value={$themeStore}
            fullWidth={true}
            direction="auto"
          />
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <Maximize2 size={14} class="item-icon" />
            <span>Shape</span>
          </div>
          <div class="segment-control">
            <button
              class:active={$shapeStore === 'pill'}
              onclick={() => ($shapeStore = 'pill')}>Pill</button
            >
            <button
              class:active={$shapeStore === 'box'}
              onclick={() => ($shapeStore = 'box')}>Box</button
            >
          </div>
        </div>

        <div class="setting-item flex-col">
          <div class="setting-label">
            <Sparkles size={14} class="item-icon" />
            <span>Ambient Glow Intensity</span>
          </div>
          <CustomSelect
            options={glowIntensityOptions}
            bind:value={$glowIntensityStore}
            fullWidth={true}
            direction="auto"
          />
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <MousePointer size={14} class="item-icon" />
            <span>Allow Screen Repositioning</span>
          </div>
          <label class="toggle">
            <input type="checkbox" bind:checked={$allowDragPositionStore} />
            <span class="slider"></span>
          </label>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <Cpu size={14} class="item-icon" />
            <span>Hardware Acceleration</span>
          </div>
          <label class="toggle">
            <input
              type="checkbox"
              bind:checked={$hardwareAccelStore}
              onchange={(e) => {
                SetHardwareAcceleration(e.currentTarget.checked).catch(() => {})
              }}
            />
            <span class="slider"></span>
          </label>
        </div>

        {#if $hiddenPillOffsetXStore !== 0}
          <div class="setting-item">
            <div class="setting-label">
              <RotateCcw size={14} class="item-icon" />
              <span>Island Position</span>
            </div>
            <button
              class="action-btn"
              onclick={() => ($hiddenPillOffsetXStore = 0)}
            >
              Reset to Center
            </button>
          </div>
        {/if}

        <div class="setting-item flex-col">
          <div class="setting-label">
            <Eye size={14} class="item-icon" />
            <span
              >Compact Mode Opacity ({Math.round(
                $opacityMinimalStore * 100
              )}%)</span
            >
          </div>
          <input
            type="range"
            min="0.2"
            max="1.0"
            step="0.05"
            bind:value={$opacityMinimalStore}
            class="range-slider no-swipe"
          />
        </div>

        <div class="setting-item flex-col">
          <div class="setting-label">
            <Maximize2 size={14} class="item-icon" />
            <span
              >Expanded Mode Opacity ({Math.round(
                $opacityExpandedStore * 100
              )}%)</span
            >
          </div>
          <input
            type="range"
            min="0.2"
            max="1.0"
            step="0.05"
            bind:value={$opacityExpandedStore}
            class="range-slider no-swipe"
          />
        </div>

        <div class="setting-item flex-col">
          <div class="setting-label">
            <EyeOff size={14} class="item-icon" />
            <span
              >Hidden Widget Opacity ({Math.round(
                $opacityHiddenStore * 100
              )}%)</span
            >
          </div>
          <input
            type="range"
            min="0.2"
            max="1.0"
            step="0.05"
            bind:value={$opacityHiddenStore}
            class="range-slider no-swipe"
          />
        </div>
      </ScrollView>
    </BentoItem>
  </div>
{/snippet}

{#snippet behaviorSlide()}
  <div class="card-wrapper">
    <BentoItem>
      <ScrollView class="section-content">
        <div class="card-category-header">
          <div class="category-icon-badge primary">
            <Sliders size={15} />
          </div>
          <div class="category-text">
            <span class="category-name">Behavior</span>
            <span class="category-desc">Auto-hide and wake controls</span>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <EyeOff size={14} class="item-icon" />
            <span>Auto-hide Home</span>
          </div>
          <label class="toggle">
            <input type="checkbox" bind:checked={$homeAutoHideStore} />
            <span class="slider"></span>
          </label>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <EyeOff size={14} class="item-icon" />
            <span>Auto-hide Now Playing</span>
          </div>
          <label class="toggle">
            <input type="checkbox" bind:checked={$mediaAutoHideStore} />
            <span class="slider"></span>
          </label>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <EyeOff size={14} class="item-icon" />
            <span>Auto-hide Notifications</span>
          </div>
          <label class="toggle">
            <input type="checkbox" bind:checked={$notifAutoHideStore} />
            <span class="slider"></span>
          </label>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <Timer size={14} class="item-icon" />
            <span>Auto-hide Delay</span>
          </div>
          <div class="segment-control">
            <button
              class:active={$autoHideDelayStore === 5}
              onclick={() => ($autoHideDelayStore = 5)}>5s</button
            >
            <button
              class:active={$autoHideDelayStore === 10}
              onclick={() => ($autoHideDelayStore = 10)}>10s</button
            >
            <button
              class:active={$autoHideDelayStore === 20}
              onclick={() => ($autoHideDelayStore = 20)}>20s</button
            >
          </div>
        </div>

        <div class="setting-item flex-col">
          <div class="setting-label">
            <Eye size={14} class="item-icon" />
            <span>Hidden Island Display Widget</span>
          </div>
          <div class="segment-control full-width">
            <button
              class:active={$autoHideDisplayStore === 'dash'}
              onclick={() => ($autoHideDisplayStore = 'dash')}
              >None (Dash)</button
            >
            <button
              class:active={$autoHideDisplayStore === 'clock'}
              onclick={() => ($autoHideDisplayStore = 'clock')}>Time</button
            >
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <MousePointer size={14} class="item-icon" />
            <span>Wake Method</span>
          </div>
          <div class="segment-control">
            <button
              class:active={$wakeMethodStore === 'hover'}
              onclick={() => ($wakeMethodStore = 'hover')}>Hover</button
            >
            <button
              class:active={$wakeMethodStore === 'click'}
              onclick={() => ($wakeMethodStore = 'click')}>Click</button
            >
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <Volume2 size={14} class="item-icon" />
            <span>Notification Sound Effects</span>
          </div>
          <label class="toggle">
            <input type="checkbox" bind:checked={$notifSoundStore} />
            <span class="slider"></span>
          </label>
        </div>

        {#if $notifSoundStore}
          <div class="setting-item flex-col">
            <div class="setting-label">
              <Volume2 size={14} class="item-icon" />
              <span>Chime Volume ({Math.round($notifVolumeStore * 100)}%)</span>
            </div>
            <input
              type="range"
              min="0.1"
              max="1.0"
              step="0.05"
              bind:value={$notifVolumeStore}
              class="range-slider no-swipe"
            />
          </div>
        {/if}

        <div class="setting-item">
          <div class="setting-label">
            <Power size={14} class="item-icon" />
            <span>Launch on Windows Startup</span>
          </div>
          <label class="toggle">
            <input type="checkbox" bind:checked={$autostartStore} />
            <span class="slider"></span>
          </label>
        </div>
      </ScrollView>
    </BentoItem>
  </div>
{/snippet}

{#snippet notificationsSlide()}
  <div class="card-wrapper">
    <BentoItem>
      <ScrollView class="section-content">
        <div class="card-category-header">
          <div class="category-icon-badge amber">
            <Bell size={15} />
          </div>
          <div class="category-text">
            <span class="category-name">Notifications</span>
            <span class="category-desc">Do Not Disturb & wake options</span>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <BellOff size={14} class="item-icon" />
            <span>Do Not Disturb (DND)</span>
          </div>
          <label class="toggle">
            <input type="checkbox" bind:checked={$dndStore} />
            <span class="slider"></span>
          </label>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <BellRing size={14} class="item-icon" />
            <span>Auto-wake on Notification</span>
          </div>
          <label class="toggle">
            <input type="checkbox" bind:checked={$notifAutoWakeStore} />
            <span class="slider"></span>
          </label>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <Bell size={14} class="item-icon" />
            <span>Active Notifications</span>
          </div>
          <span class="badge-count">{$notifData.length}</span>
        </div>

        <button class="action-btn danger" onclick={clearAllNotifications}>
          <Trash2 size={14} />
          <span>Clear All Notifications</span>
        </button>
      </ScrollView>
    </BentoItem>
  </div>
{/snippet}

{#snippet clockSlide()}
  <div class="card-wrapper">
    <BentoItem>
      <ScrollView class="section-content">
        <div class="card-category-header">
          <div class="category-icon-badge primary">
            <Clock size={15} />
          </div>
          <div class="category-text">
            <span class="category-name">Clock & Timezone</span>
            <span class="category-desc">Time format and regional timezone</span>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <Clock size={14} class="item-icon" />
            <span>Time Format</span>
          </div>
          <div class="segment-control">
            <button
              class:active={$timeFormatStore === '12h'}
              onclick={() => ($timeFormatStore = '12h')}>12H</button
            >
            <button
              class:active={$timeFormatStore === '24h'}
              onclick={() => ($timeFormatStore = '24h')}>24H</button
            >
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <Calendar size={14} class="item-icon" />
            <span>Date Format</span>
          </div>
          <div class="segment-control">
            <button
              class:active={$dateFormatStore === 'default'}
              onclick={() => ($dateFormatStore = 'default')}>Def</button
            >
            <button
              class:active={$dateFormatStore === 'iso'}
              onclick={() => ($dateFormatStore = 'iso')}>ISO</button
            >
            <button
              class:active={$dateFormatStore === 'us'}
              onclick={() => ($dateFormatStore = 'us')}>US</button
            >
            <button
              class:active={$dateFormatStore === 'eu'}
              onclick={() => ($dateFormatStore = 'eu')}>EU</button
            >
          </div>
        </div>

        <div class="setting-item flex-col">
          <div class="setting-label">
            <Globe size={14} class="item-icon" />
            <span>Timezone</span>
          </div>
          <CustomSelect
            options={timezoneOptions}
            bind:value={selectedTz}
            fullWidth={true}
            direction="auto"
            searchable={true}
            searchPlaceholder="Search timezones..."
          />
        </div>
      </ScrollView>
    </BentoItem>
  </div>
{/snippet}

{#snippet weatherSlide()}
  <div class="card-wrapper">
    <BentoItem>
      <ScrollView class="section-content">
        <div class="card-category-header">
          <div class="category-icon-badge amber">
            <CloudSun size={15} />
          </div>
          <div class="category-text">
            <span class="category-name">Weather</span>
            <span class="category-desc">Temperature unit and location</span>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <Thermometer size={14} class="item-icon" />
            <span>Temperature Unit</span>
          </div>
          <div class="segment-control">
            <button
              class:active={$weatherUnitStore === 'celsius'}
              onclick={() => ($weatherUnitStore = 'celsius')}>°C</button
            >
            <button
              class:active={$weatherUnitStore === 'fahrenheit'}
              onclick={() => ($weatherUnitStore = 'fahrenheit')}>°F</button
            >
            <button
              class:active={$weatherUnitStore === 'kelvin'}
              onclick={() => ($weatherUnitStore = 'kelvin')}>K</button
            >
          </div>
        </div>

        <div class="setting-item flex-col">
          <div class="setting-label">
            <MapPin size={14} class="item-icon" />
            <span>Weather Location</span>
          </div>
          <CustomSelect
            options={weatherOptions}
            bind:value={selectedWeatherCity}
            fullWidth={true}
            direction="auto"
            searchable={true}
            searchPlaceholder="Search city name (e.g. Manila, New York, Tokyo)..."
            onSearch={handleWeatherSearch}
            onSelectPayload={handleSelectWeatherPayload}
          />
        </div>
      </ScrollView>
    </BentoItem>
  </div>
{/snippet}

<div class="settings-page">
  <div class="header">
    <div class="header-side left"></div>
    <div class="header-center">
      <div class="title">Settings</div>
    </div>
    <div class="header-side right"></div>
  </div>

  <div class="gallery-wrapper">
    <ComponentGallery
      slides={[
        { id: 'appearance', snippet: appearanceSlide },
        { id: 'behavior', snippet: behaviorSlide },
        { id: 'notifications', snippet: notificationsSlide },
        { id: 'clock', snippet: clockSlide },
        { id: 'weather', snippet: weatherSlide }
      ]}
      loop={true}
      showControls={true}
    />
  </div>
</div>

<style>
  .settings-page {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 14px 16px 58px 16px;
    box-sizing: border-box;
    color: var(--c-text-primary);
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
    flex-shrink: 0;
    width: 100%;
    box-sizing: border-box;
  }

  .header-side {
    flex: 1 1 0;
    display: flex;
    align-items: center;
    min-width: 0;
  }

  .header-side.left {
    pointer-events: auto;
  }

  .header-side.right {
    justify-content: flex-end;
  }

  .header-center {
    flex: 1 1 0;
    display: flex;
    justify-content: center;
    align-items: center;
    white-space: nowrap;
  }

  .title {
    font-weight: 600;
    font-size: 0.9rem;
    letter-spacing: -0.01em;
    text-align: center;
    color: var(--c-text-primary);
  }

  .gallery-wrapper {
    flex: 1;
    min-height: 0;
    width: 100%;
  }

  .card-wrapper {
    width: 100%;
    height: 100%;
    padding: 0;
    box-sizing: border-box;
  }

  .section-content {
    display: flex;
    flex-direction: column;
    gap: 8px;
    height: 100%;
    padding: 10px 24px 30px 24px; /* Internal card padding with space for dropdowns */
    box-sizing: border-box;
  }

  /* Category Header Banner */
  .card-category-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 8px 10px 4px;
    border-bottom: 1px solid var(--c-border);
    margin-bottom: 2px;
  }

  .category-icon-badge {
    width: 28px;
    height: 28px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: white;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
  }

  .category-icon-badge.purple {
    background: linear-gradient(135deg, var(--c-purple), var(--c-primary));
  }

  .category-icon-badge.primary {
    background: linear-gradient(
      135deg,
      var(--c-primary),
      var(--c-surface-elevated, #3b82f6)
    );
  }

  .category-icon-badge.amber {
    background: linear-gradient(135deg, var(--c-warning), var(--c-amber));
  }

  .category-text {
    display: flex;
    flex-direction: column;
    gap: 1px;
  }

  .category-name {
    font-size: 13px;
    font-weight: 700;
    color: var(--c-text-primary);
  }

  .category-desc {
    font-size: 10px;
    color: var(--c-text-muted);
  }

  /* Setting Items */
  .setting-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--c-item-bg);
    padding: 8px 10px;
    border-radius: 10px;
    font-size: 12px;
    font-weight: 500;
    border: 1px solid var(--c-border);
  }

  .setting-item.flex-col {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }

  .setting-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 500;
    color: var(--c-text-primary);
  }
  :global(.item-icon) {
    color: var(--c-primary);
    flex-shrink: 0;
  }

  .badge-count {
    font-size: 11px;
    font-weight: 700;
    background: var(--c-primary);
    color: white;
    padding: 2px 8px;
    border-radius: 10px;
  }

  .segment-control {
    display: flex;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 6px;
    padding: 2px;
    gap: 2px;
  }

  .segment-control.full-width {
    width: 100%;
  }

  .segment-control.full-width button {
    flex: 1;
    text-align: center;
    padding: 5px 4px;
  }

  :global([data-theme='light']) .segment-control {
    background: rgba(0, 0, 0, 0.05);
  }

  .segment-control button {
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    padding: 4px 10px;
    font-size: 11px;
    font-weight: 600;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .segment-control button:hover {
    color: var(--c-text-primary);
  }

  .segment-control button.active {
    background: var(--c-border);
    color: var(--c-text-primary);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .custom-select {
    width: 100%;
    background: var(--c-contrast-bg, rgba(0, 0, 0, 0.3));
    color: var(--c-text-primary);
    border: 1px solid var(--c-border);
    border-radius: 6px;
    padding: 6px 10px;
    font-size: 11px;
    font-weight: 500;
    outline: none;
    cursor: pointer;
  }

  :global([data-theme='light']) .custom-select {
    background: rgba(255, 255, 255, 0.85);
  }

  .custom-select option {
    background: var(--c-bg);
    color: var(--c-text-primary);
  }

  .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    width: 100%;
    padding: 8px;
    border-radius: 10px;
    border: 1px solid var(--c-border);
    background: var(--c-item-bg);
    color: var(--c-text-primary);
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .action-btn:hover {
    background: var(--c-hover);
  }

  .action-btn.danger {
    color: var(--c-warning);
    border-color: rgba(248, 113, 113, 0.3);
  }

  .action-btn.danger:hover {
    background: rgba(248, 113, 113, 0.15);
  }

  /* Toggle Switch CSS */
  .toggle {
    position: relative;
    display: inline-block;
    width: 38px;
    height: 20px;
  }
  .toggle input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  .slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--c-border);
    transition: 0.3s;
    border-radius: 20px;
  }
  .slider:before {
    position: absolute;
    content: '';
    height: 14px;
    width: 14px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: 0.3s;
    border-radius: 50%;
  }
  input:checked + .slider {
    background-color: var(--c-primary);
  }
  input:checked + .slider:before {
    transform: translateX(18px);
  }

  .range-slider {
    width: 100%;
    accent-color: var(--c-primary);
    cursor: pointer;
    margin-top: 4px;
  }
</style>
