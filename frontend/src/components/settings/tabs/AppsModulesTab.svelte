<script lang="ts">
  import {
    FileText,
    Music,
    Bell,
    Layers,
    AlertCircle,
    SlidersHorizontal,
    Calculator,
    Timer,
    LayoutGrid
  } from 'lucide-svelte'
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
      <span>Apps & Features</span>
    </div>
  </div>

  <div class="setting-group">
    <SettingItem label="Action Center" icon={SlidersHorizontal}>
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={$pageVisibilityStore.actioncenter}
          onchange={() => togglePage('actioncenter')}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>

    <SettingItem label="Calculator" icon={Calculator}>
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={$pageVisibilityStore.calculator}
          onchange={() => togglePage('calculator')}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>

    <SettingItem label="Timer" icon={Timer}>
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={$pageVisibilityStore.timer !== false}
          onchange={() => {
            const next = !($pageVisibilityStore.timer !== false)
            $pageVisibilityStore = {
              ...$pageVisibilityStore,
              timer: next,
              focus: next
            }
          }}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>

    <SettingItem label="Now Playing / Media" icon={Music}>
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={$pageVisibilityStore.media}
          onchange={() => togglePage('media')}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>

    <SettingItem label="Quick Notes" icon={FileText}>
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={$pageVisibilityStore.notes}
          onchange={() => togglePage('notes')}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>

    <SettingItem label="Notifications" icon={Bell}>
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
      <LayoutGrid size={13} />
      <span>Home Widgets</span>
    </div>
  </div>

  <div class="setting-group">
    <!-- Clock Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'clock'}>
      <SettingItem label="Clock & Time">
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
      <SettingItem label="Calendar & Date">
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
      <SettingItem label="Weather & Forecast">
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
      <SettingItem label="Wireless & Connectivity">
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

<style src="../../styles/settings/AppsModulesTab.css"></style>
