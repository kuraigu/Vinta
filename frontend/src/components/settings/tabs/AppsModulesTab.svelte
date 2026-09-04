<script lang="ts">
  import {
    FileText,
    Music,
    Bell,
    Layers,
    AlertCircle,
    SlidersHorizontal
  } from '@lucide/svelte'
  import SettingItem from '../SettingItem.svelte'
  import {
    pageVisibilityStore,
    homeWidgetsVisibilityStore,
    batteryDataStore,
    type HomeWidgetsVisibility,
    type PageVisibility
  } from '../../../store'

  let shakeKey = $state<string | null>(null)
  let shakeTimeout: ReturnType<typeof setTimeout> | null = null

  function triggerShake(key: string) {
    shakeKey = key
    if (shakeTimeout) clearTimeout(shakeTimeout)
    shakeTimeout = setTimeout(() => {
      shakeKey = null
    }, 600)
  }

  let hasBattery = $derived($batteryDataStore.hasBattery)
  let isBatteryActive = $derived(
    $homeWidgetsVisibilityStore.battery && hasBattery
  )

  let activeHomeCount = $derived(
    ($homeWidgetsVisibilityStore.clock ? 1 : 0) +
      ($homeWidgetsVisibilityStore.calendar ? 1 : 0) +
      ($homeWidgetsVisibilityStore.weather ? 1 : 0) +
      ($homeWidgetsVisibilityStore.wireless ? 1 : 0) +
      (isBatteryActive ? 1 : 0)
  )

  let isClockLocked = $derived(
    $homeWidgetsVisibilityStore.clock && activeHomeCount <= 1
  )
  let isCalLocked = $derived(
    $homeWidgetsVisibilityStore.calendar && activeHomeCount <= 1
  )
  let isWeatherLocked = $derived(
    $homeWidgetsVisibilityStore.weather && activeHomeCount <= 1
  )
  let isWirelessLocked = $derived(
    $homeWidgetsVisibilityStore.wireless && activeHomeCount <= 1
  )
  let isBatteryLocked = $derived(
    hasBattery && $homeWidgetsVisibilityStore.battery && activeHomeCount <= 1
  )

  function toggleHomeWidget(key: keyof HomeWidgetsVisibility) {
    const isCurrentlyActive = $homeWidgetsVisibilityStore[key]
    if (isCurrentlyActive && activeHomeCount <= 1) {
      triggerShake(key)
      return
    }
    $homeWidgetsVisibilityStore = {
      ...$homeWidgetsVisibilityStore,
      [key]: !isCurrentlyActive
    }
  }

  function togglePage(key: keyof PageVisibility) {
    $pageVisibilityStore = {
      ...$pageVisibilityStore,
      [key]: !$pageVisibilityStore[key]
    }
  }
</script>

<div class="apps-modules-tab">
  <!-- Section 1: Apps -->
  <div class="section-header">
    <div class="section-title">
      <Layers size={13} />
      <span>Apps</span>
    </div>
    <span class="section-desc">
      Toggle showing of apps in the App Drawer and Dock
    </span>
  </div>

  <div class="setting-group">
    <SettingItem
      label="Action Center"
      description="Quick wireless toggles, Bluetooth devices, battery status, and airplane mode"
      icon={SlidersHorizontal}
    >
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={$pageVisibilityStore.actioncenter}
          onchange={() => togglePage('actioncenter')}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>

    <SettingItem
      label="Now Playing / Media"
      description="System media playback controls and audio peak visualizer"
      icon={Music}
    >
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={$pageVisibilityStore.media}
          onchange={() => togglePage('media')}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>

    <SettingItem
      label="Quick Notes"
      description="Scratchpad notes with Markdown support and clipboard exports"
      icon={FileText}
    >
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={$pageVisibilityStore.notes}
          onchange={() => togglePage('notes')}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>

    <SettingItem
      label="Notifications"
      description="Desktop alert feeds, sound chimes, and system notifications"
      icon={Bell}
    >
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={$pageVisibilityStore.notifications}
          onchange={() => togglePage('notifications')}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>
  </div>

  <div class="divider"></div>

  <!-- Section 2: Home Widgets -->
  <div class="section-header">
    <div class="section-title">
      <span>Home Widgets</span>
    </div>
  </div>

  <div class="setting-group">
    <!-- Clock Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'clock'}>
      <SettingItem
        label="Clock & Time"
        description="Live clock with customizable timezones and 12h/24h formats"
      >
        <div
          class="toggle-container"
          class:locked={isClockLocked}
          title={isClockLocked
            ? 'At least one widget must remain visible on the home island'
            : ''}
          role="button"
          tabindex="0"
          onclick={() => isClockLocked && triggerShake('clock')}
          onkeydown={(e) =>
            e.key === 'Enter' && isClockLocked && triggerShake('clock')}
        >
          <label class="toggle-switch" class:disabled-switch={isClockLocked}>
            <input
              type="checkbox"
              checked={$homeWidgetsVisibilityStore.clock}
              disabled={isClockLocked}
              onchange={() => toggleHomeWidget('clock')}
            />
            <span class="slider"></span>
          </label>
        </div>
      </SettingItem>
      {#if isClockLocked && shakeKey === 'clock'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span
          >
        </div>
      {/if}
    </div>

    <!-- Calendar Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'calendar'}>
      <SettingItem
        label="Calendar & Date"
        description="Date indicator and interactive month/year calendar picker"
      >
        <div
          class="toggle-container"
          class:locked={isCalLocked}
          title={isCalLocked
            ? 'At least one widget must remain visible on the home island'
            : ''}
          role="button"
          tabindex="0"
          onclick={() => isCalLocked && triggerShake('calendar')}
          onkeydown={(e) =>
            e.key === 'Enter' && isCalLocked && triggerShake('calendar')}
        >
          <label class="toggle-switch" class:disabled-switch={isCalLocked}>
            <input
              type="checkbox"
              checked={$homeWidgetsVisibilityStore.calendar}
              disabled={isCalLocked}
              onchange={() => toggleHomeWidget('calendar')}
            />
            <span class="slider"></span>
          </label>
        </div>
      </SettingItem>
      {#if isCalLocked && shakeKey === 'calendar'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span
          >
        </div>
      {/if}
    </div>

    <!-- Weather Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'weather'}>
      <SettingItem
        label="Weather & Forecast"
        description="Live temperature, condition icons, and 7-day forecast cards"
      >
        <div
          class="toggle-container"
          class:locked={isWeatherLocked}
          title={isWeatherLocked
            ? 'At least one widget must remain visible on the home island'
            : ''}
          role="button"
          tabindex="0"
          onclick={() => isWeatherLocked && triggerShake('weather')}
          onkeydown={(e) =>
            e.key === 'Enter' && isWeatherLocked && triggerShake('weather')}
        >
          <label class="toggle-switch" class:disabled-switch={isWeatherLocked}>
            <input
              type="checkbox"
              checked={$homeWidgetsVisibilityStore.weather}
              disabled={isWeatherLocked}
              onchange={() => toggleHomeWidget('weather')}
            />
            <span class="slider"></span>
          </label>
        </div>
      </SettingItem>
      {#if isWeatherLocked && shakeKey === 'weather'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span
          >
        </div>
      {/if}
    </div>

    <!-- Battery Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'battery'}>
      <SettingItem
        label="Battery & Power"
        description={hasBattery
          ? 'Battery percentage, charging animations, and power saver indicator'
          : 'Battery hardware is not available on this device'}
        badge={hasBattery ? undefined : 'Unavailable'}
        disabled={!hasBattery}
      >
        <div
          class="toggle-container"
          class:locked={isBatteryLocked || !hasBattery}
          title={!hasBattery
            ? 'Battery hardware is not available on this device'
            : isBatteryLocked
              ? 'At least one widget must remain visible on the home island'
              : ''}
          role="button"
          tabindex="0"
          onclick={() => isBatteryLocked && triggerShake('battery')}
          onkeydown={(e) =>
            e.key === 'Enter' && isBatteryLocked && triggerShake('battery')}
        >
          <label
            class="toggle-switch"
            class:disabled-switch={isBatteryLocked || !hasBattery}
          >
            <input
              type="checkbox"
              checked={hasBattery && $homeWidgetsVisibilityStore.battery}
              disabled={isBatteryLocked || !hasBattery}
              onchange={() => toggleHomeWidget('battery')}
            />
            <span class="slider"></span>
          </label>
        </div>
      </SettingItem>
      {#if isBatteryLocked && shakeKey === 'battery'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span
          >
        </div>
      {/if}
    </div>

    <!-- Wireless & Connectivity Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'wireless'}>
      <SettingItem
        label="Wireless & Connectivity"
        description="Wi-Fi signal, Bluetooth devices, and Airplane mode status capsule"
      >
        <div
          class="toggle-container"
          class:locked={isWirelessLocked}
          title={isWirelessLocked
            ? 'At least one widget must remain visible on the home island'
            : ''}
          role="button"
          tabindex="0"
          onclick={() => isWirelessLocked && triggerShake('wireless')}
          onkeydown={(e) =>
            e.key === 'Enter' && isWirelessLocked && triggerShake('wireless')}
        >
          <label class="toggle-switch" class:disabled-switch={isWirelessLocked}>
            <input
              type="checkbox"
              checked={$homeWidgetsVisibilityStore.wireless}
              disabled={isWirelessLocked}
              onchange={() => toggleHomeWidget('wireless')}
            />
            <span class="slider"></span>
          </label>
        </div>
      </SettingItem>
      {#if isWirelessLocked && shakeKey === 'wireless'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span
          >
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .apps-modules-tab {
    display: flex;
    flex-direction: column;
    gap: 12px;
    text-align: left;
  }

  .section-header {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    gap: 2px;
    margin-bottom: 2px;
  }

  .section-title {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 700;
    color: var(--c-text-primary);
    text-align: left;
  }

  .section-desc {
    font-size: 10px;
    color: var(--c-text-muted);
    text-align: left;
  }

  .divider {
    height: 1px;
    background: var(--c-border);
    margin: 4px 0;
  }

  .setting-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .item-wrapper {
    display: flex;
    flex-direction: column;
    gap: 4px;
    transition: transform 0.2s ease;
  }

  .item-wrapper.shake {
    animation: shakeAnim 0.4s cubic-bezier(0.36, 0.07, 0.19, 0.97) both;
  }

  @keyframes shakeAnim {
    10%,
    90% {
      transform: translateX(-2px);
    }
    20%,
    80% {
      transform: translateX(3px);
    }
    30%,
    50%,
    70% {
      transform: translateX(-4px);
    }
    40%,
    60% {
      transform: translateX(4px);
    }
  }

  .locked-warning {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 10px;
    color: var(--c-danger, #ef4444);
    padding: 4px 10px;
    border-radius: 8px;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.2);
  }

  .toggle-container.locked {
    cursor: not-allowed;
  }

  .toggle-switch {
    position: relative;
    display: inline-block;
    width: 34px;
    height: 18px;
  }

  .toggle-switch.disabled-switch {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .toggle-switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .slider {
    position: absolute;
    cursor: pointer;
    inset: 0;
    background-color: rgba(255, 255, 255, 0.2);
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
    bottom: 1px;
    background-color: white;
    transition: 0.3s;
    border-radius: 50%;
  }

  input:checked + .slider {
    background-color: var(--c-primary, #6366f1);
  }

  input:checked + .slider:before {
    transform: translateX(14px);
  }

  .toggle-switch.disabled-switch .slider {
    cursor: not-allowed;
  }
</style>
