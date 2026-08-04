<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 540, h: 440, r: 32 },
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
  import CustomSelect from '../components/CustomSelect.svelte'
  import {
    Palette,
    Sliders,
    Bell,
    Clock,
    CloudSun,
    Cpu,
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
    Database,
    Check
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
      // Fallback to Manila
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
        payload: r
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

  import {
    SetAutostart,
    SetHardwareAcceleration
  } from '../../wailsjs/go/backend/App.js'

  const themeOptions = [{ label: 'OLED Dark', value: 'dark' }]

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

  let cacheCleared = $state(false)
  let clearTimer: ReturnType<typeof setTimeout> | null = null
  let cacheSizeStr = $state('0 B')

  function formatBytes(bytes: number): string {
    if (bytes <= 0) return '0 B'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  function calculateCacheSize() {
    try {
      let totalBytes = 0
      const notifs = $notifData || []
      if (notifs.length > 0) {
        totalBytes += new Blob([JSON.stringify(notifs)]).size
      }

      const keepKeys = new Set([
        'vinta_theme',
        'vinta_shape',
        'vinta_glow_intensity',
        'vinta_home_autohide',
        'vinta_media_autohide',
        'vinta_notif_autohide',
        'vinta_autohide_delay',
        'vinta_autohide_display',
        'vinta_wake_method',
        'vinta_dnd',
        'vinta_notif_autowake',
        'vinta_time_format',
        'vinta_timezone',
        'vinta_date_format',
        'vinta_weather_location',
        'vinta_weather_unit',
        'vinta_opacity_minimal',
        'vinta_opacity_expanded',
        'vinta_opacity_hidden',
        'vinta_hidden_pill_offset_x',
        'vinta_allow_drag_pos',
        'vinta_hw_accel',
        'vinta_notif_sound',
        'vinta_notif_volume',
        'vinta_autostart'
      ])

      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i)
        if (key && !keepKeys.has(key)) {
          const val = localStorage.getItem(key) || ''
          totalBytes += (key.length + val.length) * 2
        }
      }

      cacheSizeStr = formatBytes(totalBytes)
    } catch {
      cacheSizeStr = '0 B'
    }
  }

  $effect(() => {
    calculateCacheSize()
  })

  function handleClearCache() {
    try {
      const keepKeys = new Set([
        'vinta_theme',
        'vinta_shape',
        'vinta_glow_intensity',
        'vinta_home_autohide',
        'vinta_media_autohide',
        'vinta_notif_autohide',
        'vinta_autohide_delay',
        'vinta_autohide_display',
        'vinta_wake_method',
        'vinta_dnd',
        'vinta_notif_autowake',
        'vinta_time_format',
        'vinta_timezone',
        'vinta_date_format',
        'vinta_weather_location',
        'vinta_weather_unit',
        'vinta_opacity_minimal',
        'vinta_opacity_expanded',
        'vinta_opacity_hidden',
        'vinta_hidden_pill_offset_x',
        'vinta_allow_drag_pos',
        'vinta_hw_accel',
        'vinta_notif_sound',
        'vinta_notif_volume',
        'vinta_autostart'
      ])
      for (let i = localStorage.length - 1; i >= 0; i--) {
        const key = localStorage.key(i)
        if (key && !keepKeys.has(key)) {
          localStorage.removeItem(key)
        }
      }
      cacheCleared = true
      calculateCacheSize()
      if (clearTimer) clearTimeout(clearTimer)
      clearTimer = setTimeout(() => {
        cacheCleared = false
      }, 2000)
    } catch {
      // Ignore
    }
  }

  // Sidebar Category Definition
  const categories = [
    { id: 'appearance', name: 'Appearance', icon: Palette },
    { id: 'behavior', name: 'Behavior', icon: Sliders },
    { id: 'notifications', name: 'Notifications', icon: Bell },
    { id: 'clock', name: 'Clock & Date', icon: Clock },
    { id: 'weather', name: 'Weather', icon: CloudSun },
    { id: 'system', name: 'System', icon: Cpu }
  ]

  let activeTab = $state('appearance')
</script>

{#snippet appearanceSection()}
  <div class="setting-group">
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

    <div class="setting-item flex-col">
      <div class="setting-label">
        <Eye size={14} class="item-icon" />
        <span>Compact Opacity ({Math.round($opacityMinimalStore * 100)}%)</span>
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
        <span>Expanded Opacity ({Math.round($opacityExpandedStore * 100)}%)</span>
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
        <span>Hidden Opacity ({Math.round($opacityHiddenStore * 100)}%)</span>
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
  </div>
{/snippet}

{#snippet behaviorSection()}
  <div class="setting-group">
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
        <span>Hidden Island Display</span>
      </div>
      <div class="segment-control full-width">
        <button
          class:active={$autoHideDisplayStore === 'dash'}
          onclick={() => ($autoHideDisplayStore = 'dash')}>None (Dash)</button
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
        <MousePointer size={14} class="item-icon" />
        <span>Allow Drag Repositioning</span>
      </div>
      <label class="toggle">
        <input type="checkbox" bind:checked={$allowDragPositionStore} />
        <span class="slider"></span>
      </label>
    </div>
  </div>
{/snippet}

{#snippet notificationsSection()}
  <div class="setting-group">
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
        <Volume2 size={14} class="item-icon" />
        <span>Sound Effects</span>
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
        <Bell size={14} class="item-icon" />
        <span>Active Notifications</span>
      </div>
      <span class="badge-count">{$notifData.length}</span>
    </div>

    <button class="action-btn danger" onclick={clearAllNotifications}>
      <Trash2 size={14} />
      <span>Clear All Notifications</span>
    </button>
  </div>
{/snippet}

{#snippet clockSection()}
  <div class="setting-group">
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
  </div>
{/snippet}

{#snippet weatherSection()}
  <div class="setting-group">
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
  </div>
{/snippet}

{#snippet systemSection()}
  <div class="setting-group">
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

    <button
      class="action-btn"
      class:success={cacheCleared}
      onclick={handleClearCache}
    >
      {#if cacheCleared}
        <Check size={14} />
        <span>Cache Cleared!</span>
        <span class="cache-size-badge">0 B</span>
      {:else}
        <Database size={14} />
        <span>Clear App Cache</span>
        <span class="cache-size-badge">{cacheSizeStr}</span>
      {/if}
    </button>
  </div>
{/snippet}

<div class="settings-page">
  <div class="header">
    <div class="title">Settings</div>
  </div>

  <div class="settings-layout">
    <!-- Left Category Navigation Sidebar -->
    <div class="settings-sidebar">
      {#each categories as cat (cat.id)}
        <button
          class="sidebar-tab"
          class:active={activeTab === cat.id}
          onclick={(e) => {
            e.stopPropagation()
            activeTab = cat.id
          }}
        >
          <cat.icon size={15} class="tab-icon" />
          <span class="tab-label">{cat.name}</span>
        </button>
      {/each}
    </div>

    <!-- Right Dynamic Settings Content View -->
    <div class="settings-content">
      <ScrollView class="content-scroll">
        {#if activeTab === 'appearance'}
          {@render appearanceSection()}
        {:else if activeTab === 'behavior'}
          {@render behaviorSection()}
        {:else if activeTab === 'notifications'}
          {@render notificationsSection()}
        {:else if activeTab === 'clock'}
          {@render clockSection()}
        {:else if activeTab === 'weather'}
          {@render weatherSection()}
        {:else if activeTab === 'system'}
          {@render systemSection()}
        {/if}
      </ScrollView>
    </div>
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
    justify-content: center;
    margin-bottom: 14px;
    flex-shrink: 0;
    width: 100%;
  }

  .title {
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--c-text-primary);
    text-align: center;
  }

  .settings-layout {
    flex: 1;
    min-height: 0;
    display: flex;
    gap: 12px;
    width: 100%;
  }

  /* Left Sidebar Navigation */
  .settings-sidebar {
    width: 140px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    border-right: 1px solid var(--c-border);
    padding-right: 10px;
    box-sizing: border-box;
  }

  .sidebar-tab {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    border-radius: 10px;
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    text-align: left;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    width: 100%;
    box-sizing: border-box;
  }

  .sidebar-tab:hover {
    background: var(--c-hover, rgba(255, 255, 255, 0.08));
    color: var(--c-text-primary);
  }

  .sidebar-tab.active {
    background: var(--c-item-bg, rgba(255, 255, 255, 0.12));
    color: var(--c-primary, #6366f1);
    font-weight: 600;
    box-shadow: inset 0 1px 0 var(--c-border);
  }

  :global(.tab-icon) {
    flex-shrink: 0;
  }

  .tab-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Right Settings Scroll Content */
  .settings-content {
    flex: 1;
    min-width: 0;
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .setting-group {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding-right: 4px;
    padding-bottom: 8px;
  }

  .setting-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    border-radius: 12px;
    background: var(--c-item-bg, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.08));
    box-sizing: border-box;
  }

  .setting-item.flex-col {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .setting-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    font-weight: 500;
    color: var(--c-text-primary);
  }

  :global(.item-icon) {
    color: var(--c-text-muted);
    flex-shrink: 0;
  }

  /* Segment Control */
  .segment-control {
    display: flex;
    background: var(--c-contrast-bg, rgba(0, 0, 0, 0.2));
    padding: 2px;
    border-radius: 8px;
    border: 1px solid var(--c-border);
    gap: 2px;
  }

  .segment-control.full-width {
    width: 100%;
  }

  .segment-control.full-width button {
    flex: 1;
  }

  .segment-control button {
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    font-size: 11px;
    font-weight: 500;
    padding: 4px 10px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .segment-control button.active {
    background: var(--c-item-bg);
    color: var(--c-text-primary);
    font-weight: 600;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
  }

  /* Range Slider */
  .range-slider {
    width: 100%;
    accent-color: var(--c-primary);
    cursor: pointer;
  }

  /* Toggle Switch */
  .toggle {
    position: relative;
    display: inline-block;
    width: 36px;
    height: 20px;
    flex-shrink: 0;
  }

  .toggle input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .slider {
    position: absolute;
    cursor: pointer;
    inset: 0;
    background-color: var(--c-contrast-bg, rgba(255, 255, 255, 0.2));
    transition: 0.3s;
    border-radius: 20px;
    border: 1px solid var(--c-border);
  }

  .slider:before {
    position: absolute;
    content: '';
    height: 14px;
    width: 14px;
    left: 2px;
    bottom: 2px;
    background-color: white;
    transition: 0.3s;
    border-radius: 50%;
  }

  input:checked + .slider {
    background-color: var(--c-primary);
  }

  input:checked + .slider:before {
    transform: translateX(16px);
  }

  /* Action Buttons */
  .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    width: 100%;
    padding: 9px 14px;
    border-radius: 10px;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .action-btn:hover {
    background: var(--c-hover);
  }

  .action-btn.danger {
    color: var(--c-danger);
    border-color: rgba(255, 59, 48, 0.3);
  }

  .action-btn.danger:hover {
    background: var(--c-danger-hover-bg);
  }

  .action-btn.success {
    color: var(--c-success);
    border-color: rgba(52, 199, 89, 0.4);
  }

  .badge-count {
    font-size: 11px;
    font-weight: 700;
    background: var(--c-primary);
    color: #ffffff;
    padding: 2px 7px;
    border-radius: 10px;
  }

  .cache-size-badge {
    margin-left: auto;
    font-size: 10px;
    opacity: 0.7;
    font-weight: 500;
  }
</style>
