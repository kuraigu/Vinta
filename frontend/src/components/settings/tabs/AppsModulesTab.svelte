<!--
  @component AppsModulesTab
  @description Application and widget visibility toggles. Configures which apps appear
  in the dock, which glance cards show on the collapsed island, and what page opens on wake.
-->
<script lang="ts">
  import {
    FileText,
    Music,
    Bell,
    AlertCircle,
    SlidersHorizontal,
    Calculator,
    Timer,
    Calendar,
    CloudSun,
    FolderDown,
    Activity,
    Sparkles,
    Megaphone,
    Clock,
    Battery,
    Wifi
  } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
  import SettingItem from '../SettingItem.svelte'
  import {
    pageVisibilityStore,
    homeWidgetsVisibilityStore,
    dockPinnedAppsStore,
    batteryDataStore,
    type HomeWidgetsVisibility,
    type PageVisibility,
    type DockPinnedApps
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

  function toggleDockPinned(key: keyof DockPinnedApps) {
    $dockPinnedAppsStore = {
      ...$dockPinnedAppsStore,
      [key]: !$dockPinnedAppsStore[key]
    }
  }
</script>

<div class="apps-modules-tab">
  <!-- Section 1: Apps -->
  <div class="section-header">
    <div class="section-title">
      <span>Apps & Features</span>
    </div>
  </div>

  <BentoCard variant="group">
    <SettingItem label="Action Center" icon={SlidersHorizontal}>
      <ToggleSwitch
        checked={$pageVisibilityStore.actioncenter}
        onchange={() => togglePage('actioncenter')}
        ariaLabel="Action Center"
      />
    </SettingItem>

    <SettingItem label="Calendar" icon={Calendar}>
      <ToggleSwitch
        checked={$pageVisibilityStore.calendar !== false}
        onchange={() => togglePage('calendar')}
        ariaLabel="Calendar"
      />
    </SettingItem>

    <SettingItem label="Calculator" icon={Calculator}>
      <ToggleSwitch
        checked={$pageVisibilityStore.calculator}
        onchange={() => togglePage('calculator')}
        ariaLabel="Calculator"
      />
    </SettingItem>

    <SettingItem label="File Shelf" icon={FolderDown}>
      <ToggleSwitch
        checked={$pageVisibilityStore.fileshelf !== false}
        onchange={() => togglePage('fileshelf')}
        ariaLabel="File Shelf"
      />
    </SettingItem>

    <SettingItem label="Weather" icon={CloudSun}>
      <ToggleSwitch
        checked={$pageVisibilityStore.weather !== false}
        onchange={() => togglePage('weather')}
        ariaLabel="Weather"
      />
    </SettingItem>

    <SettingItem label="Timer & Focus" icon={Timer}>
      <ToggleSwitch
        checked={$pageVisibilityStore.timer !== false}
        onchange={() => {
          const next = !($pageVisibilityStore.timer !== false)
          $pageVisibilityStore = {
            ...$pageVisibilityStore,
            timer: next,
            focus: next
          }
        }}
        ariaLabel="Timer & Focus"
      />
    </SettingItem>

    <SettingItem label="Now Playing / Media" icon={Music}>
      <ToggleSwitch
        checked={$pageVisibilityStore.media}
        onchange={() => togglePage('media')}
        ariaLabel="Now Playing / Media"
      />
    </SettingItem>

    <SettingItem label="Quick Notes" icon={FileText}>
      <ToggleSwitch
        checked={$pageVisibilityStore.notes}
        onchange={() => togglePage('notes')}
        ariaLabel="Quick Notes"
      />
    </SettingItem>

    <SettingItem label="Notifications" icon={Bell}>
      <ToggleSwitch
        checked={$pageVisibilityStore.notifications}
        onchange={() => togglePage('notifications')}
        ariaLabel="Notifications"
      />
    </SettingItem>

    <SettingItem label="System Stats" icon={Activity}>
      <ToggleSwitch
        checked={$pageVisibilityStore.performance !== false}
        onchange={() => togglePage('performance')}
        ariaLabel="System Stats"
      />
    </SettingItem>
  </BentoCard>

  <div class="divider"></div>

  <!-- Section 2: Home Widgets -->
  <div class="section-header">
    <div class="section-title">
      <span>Home Widgets</span>
    </div>
  </div>

  <BentoCard variant="group">
    <!-- Clock Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'clock'}>
      <SettingItem label="Clock & Time" icon={Clock}>
        <div
          class="toggle-container"
          class:locked={isClockLocked}
          role="button"
          tabindex="0"
          onclick={() => isClockLocked && triggerShake('clock')}
          onkeydown={(e) =>
            e.key === 'Enter' && isClockLocked && triggerShake('clock')}
        >
          <ToggleSwitch
            checked={$homeWidgetsVisibilityStore.clock}
            disabled={isClockLocked}
            onchange={() => toggleHomeWidget('clock')}
            ariaLabel="Clock & Time"
          />
        </div>
      </SettingItem>
      {#if isClockLocked && shakeKey === 'clock'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span>
        </div>
      {/if}
    </div>

    <!-- Calendar Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'calendar'}>
      <SettingItem label="Calendar & Date" icon={Calendar}>
        <div
          class="toggle-container"
          class:locked={isCalLocked}
          role="button"
          tabindex="0"
          onclick={() => isCalLocked && triggerShake('calendar')}
          onkeydown={(e) =>
            e.key === 'Enter' && isCalLocked && triggerShake('calendar')}
        >
          <ToggleSwitch
            checked={$homeWidgetsVisibilityStore.calendar}
            disabled={isCalLocked}
            onchange={() => toggleHomeWidget('calendar')}
            ariaLabel="Calendar & Date"
          />
        </div>
      </SettingItem>
      {#if isCalLocked && shakeKey === 'calendar'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span>
        </div>
      {/if}
    </div>

    <!-- Weather Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'weather'}>
      <SettingItem label="Weather & Forecast" icon={CloudSun}>
        <div
          class="toggle-container"
          class:locked={isWeatherLocked}
          role="button"
          tabindex="0"
          onclick={() => isWeatherLocked && triggerShake('weather')}
          onkeydown={(e) =>
            e.key === 'Enter' && isWeatherLocked && triggerShake('weather')}
        >
          <ToggleSwitch
            checked={$homeWidgetsVisibilityStore.weather}
            disabled={isWeatherLocked}
            onchange={() => toggleHomeWidget('weather')}
            ariaLabel="Weather & Forecast"
          />
        </div>
      </SettingItem>
      {#if isWeatherLocked && shakeKey === 'weather'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span>
        </div>
      {/if}
    </div>

    <!-- Battery Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'battery'}>
      <SettingItem label="Battery & Power" icon={Battery} disabled={!hasBattery}>
        <div
          class="toggle-container"
          class:locked={isBatteryLocked || !hasBattery}
          role="button"
          tabindex="0"
          onclick={() => isBatteryLocked && triggerShake('battery')}
          onkeydown={(e) =>
            e.key === 'Enter' && isBatteryLocked && triggerShake('battery')}
        >
          <ToggleSwitch
            checked={hasBattery && $homeWidgetsVisibilityStore.battery}
            disabled={isBatteryLocked || !hasBattery}
            onchange={() => toggleHomeWidget('battery')}
            ariaLabel="Battery & Power"
          />
        </div>
      </SettingItem>
      {#if isBatteryLocked && shakeKey === 'battery'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span>
        </div>
      {/if}
    </div>

    <!-- Wireless & Connectivity Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'wireless'}>
      <SettingItem label="Wireless & Connectivity" icon={Wifi}>
        <div
          class="toggle-container"
          class:locked={isWirelessLocked}
          role="button"
          tabindex="0"
          onclick={() => isWirelessLocked && triggerShake('wireless')}
          onkeydown={(e) =>
            e.key === 'Enter' && isWirelessLocked && triggerShake('wireless')}
        >
          <ToggleSwitch
            checked={$homeWidgetsVisibilityStore.wireless}
            disabled={isWirelessLocked}
            onchange={() => toggleHomeWidget('wireless')}
            ariaLabel="Wireless & Connectivity"
          />
        </div>
      </SettingItem>
      {#if isWirelessLocked && shakeKey === 'wireless'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>At least one widget must remain visible on the home island</span>
        </div>
      {/if}
    </div>

    <SettingItem label="Action Center Glance" icon={SlidersHorizontal}>
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.actionCenter}
        onchange={() => toggleHomeWidget('actionCenter')}
        ariaLabel="Action Center Glance"
      />
    </SettingItem>

    <SettingItem label="System Stats Glance" icon={Activity}>
      <ToggleSwitch
        checked={Boolean($homeWidgetsVisibilityStore.systemStats)}
        onchange={() => toggleHomeWidget('systemStats')}
        ariaLabel="System Stats Glance"
      />
    </SettingItem>

    <SettingItem label="Spotlight Carousel" icon={Sparkles}>
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.spotlight !== false}
        onchange={() => toggleHomeWidget('spotlight')}
        ariaLabel="Spotlight Carousel"
      />
    </SettingItem>

    <SettingItem label="Announcements Feed" icon={Megaphone}>
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.announcements !== false}
        onchange={() => toggleHomeWidget('announcements')}
        ariaLabel="Announcements Feed"
      />
    </SettingItem>
  </BentoCard>

  <div class="divider"></div>

  <!-- Section 3: Dock Pinned Favorites -->
  <div class="section-header">
    <div class="section-title">
      <span>Dock Pinned Favorites</span>
    </div>
  </div>

  <BentoCard variant="group">
    <SettingItem label="Pin Calendar" icon={Calendar}>
      <ToggleSwitch
        checked={$dockPinnedAppsStore.calendar}
        onchange={() => toggleDockPinned('calendar')}
        ariaLabel="Pin Calendar"
      />
    </SettingItem>

    <SettingItem label="Pin Quick Notes" icon={FileText}>
      <ToggleSwitch
        checked={$dockPinnedAppsStore.notes}
        onchange={() => toggleDockPinned('notes')}
        ariaLabel="Pin Quick Notes"
      />
    </SettingItem>

    <SettingItem label="Pin Calculator" icon={Calculator}>
      <ToggleSwitch
        checked={$dockPinnedAppsStore.calculator}
        onchange={() => toggleDockPinned('calculator')}
        ariaLabel="Pin Calculator"
      />
    </SettingItem>

    <SettingItem label="Pin Timer" icon={Timer}>
      <ToggleSwitch
        checked={$dockPinnedAppsStore.timer}
        onchange={() => toggleDockPinned('timer')}
        ariaLabel="Pin Timer"
      />
    </SettingItem>

    <SettingItem label="Pin Weather" icon={CloudSun}>
      <ToggleSwitch
        checked={$dockPinnedAppsStore.weather}
        onchange={() => toggleDockPinned('weather')}
        ariaLabel="Pin Weather"
      />
    </SettingItem>

    <SettingItem label="Pin File Shelf" icon={FolderDown}>
      <ToggleSwitch
        checked={$dockPinnedAppsStore.fileshelf}
        onchange={() => toggleDockPinned('fileshelf')}
        ariaLabel="Pin File Shelf"
      />
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/AppsModulesTab.css"></style>
