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
    Wifi,
    MoreHorizontal
  } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
  import SettingItem from '../SettingItem.svelte'
  import {
    pageVisibilityStore,
    homeWidgetsVisibilityStore,
    homeWidgetMenuEnabledStore,
    dockPinnedAppsStore,
    batteryDataStore,
    tStore,
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

<div class="apps-modules-tab setting-group">
  <BentoCard variant="group">
    <SettingItem
      label={$tStore('apps.actioncenter_name')}
      icon={SlidersHorizontal}
    >
      <ToggleSwitch
        checked={$pageVisibilityStore.actioncenter}
        onchange={() => togglePage('actioncenter')}
        ariaLabel={$tStore('apps.actioncenter_name')}
      />
    </SettingItem>

    <SettingItem label={$tStore('apps.calendar_name')} icon={Calendar}>
      <ToggleSwitch
        checked={$pageVisibilityStore.calendar !== false}
        onchange={() => togglePage('calendar')}
        ariaLabel={$tStore('apps.calendar_name')}
      />
    </SettingItem>

    <SettingItem label={$tStore('apps.calculator_name')} icon={Calculator}>
      <ToggleSwitch
        checked={$pageVisibilityStore.calculator}
        onchange={() => togglePage('calculator')}
        ariaLabel={$tStore('apps.calculator_name')}
      />
    </SettingItem>

    <SettingItem label={$tStore('apps.fileshelf_name')} icon={FolderDown}>
      <ToggleSwitch
        checked={$pageVisibilityStore.fileshelf !== false}
        onchange={() => togglePage('fileshelf')}
        ariaLabel={$tStore('apps.fileshelf_name')}
      />
    </SettingItem>

    <SettingItem label={$tStore('apps.weather_name')} icon={CloudSun}>
      <ToggleSwitch
        checked={$pageVisibilityStore.weather !== false}
        onchange={() => togglePage('weather')}
        ariaLabel={$tStore('apps.weather_name')}
      />
    </SettingItem>

    <SettingItem label={$tStore('settings.modules_timer_focus')} icon={Timer}>
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
        ariaLabel={$tStore('settings.modules_timer_focus')}
      />
    </SettingItem>

    <SettingItem label={$tStore('settings.modules_now_playing')} icon={Music}>
      <ToggleSwitch
        checked={$pageVisibilityStore.media}
        onchange={() => togglePage('media')}
        ariaLabel={$tStore('settings.modules_now_playing')}
      />
    </SettingItem>

    <SettingItem label={$tStore('apps.notes_name')} icon={FileText}>
      <ToggleSwitch
        checked={$pageVisibilityStore.notes}
        onchange={() => togglePage('notes')}
        ariaLabel={$tStore('apps.notes_name')}
      />
    </SettingItem>

    <SettingItem label={$tStore('apps.notifs_name')} icon={Bell}>
      <ToggleSwitch
        checked={$pageVisibilityStore.notifications}
        onchange={() => togglePage('notifications')}
        ariaLabel={$tStore('apps.notifs_name')}
      />
    </SettingItem>

    <SettingItem label={$tStore('apps.performance_name')} icon={Activity}>
      <ToggleSwitch
        checked={$pageVisibilityStore.performance !== false}
        onchange={() => togglePage('performance')}
        ariaLabel={$tStore('apps.performance_name')}
      />
    </SettingItem>
  </BentoCard>

  <div class="setting-section-header">
    <span>{$tStore('settings.modules_home_widgets')}</span>
  </div>

  <BentoCard variant="group">
    <!-- Clock Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'clock'}>
      <SettingItem label={$tStore('settings.modules_clock_time')} icon={Clock}>
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
            ariaLabel={$tStore('settings.modules_clock_time')}
          />
        </div>
      </SettingItem>
      {#if isClockLocked && shakeKey === 'clock'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>{$tStore('settings.modules_one_widget_required')}</span>
        </div>
      {/if}
    </div>

    <!-- Calendar Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'calendar'}>
      <SettingItem
        label={$tStore('settings.modules_calendar_date')}
        icon={Calendar}
      >
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
            ariaLabel={$tStore('settings.modules_calendar_date')}
          />
        </div>
      </SettingItem>
      {#if isCalLocked && shakeKey === 'calendar'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>{$tStore('settings.modules_one_widget_required')}</span>
        </div>
      {/if}
    </div>

    <!-- Weather Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'weather'}>
      <SettingItem
        label={$tStore('settings.modules_weather_forecast')}
        icon={CloudSun}
      >
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
            ariaLabel={$tStore('settings.modules_weather_forecast')}
          />
        </div>
      </SettingItem>
      {#if isWeatherLocked && shakeKey === 'weather'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>{$tStore('settings.modules_one_widget_required')}</span>
        </div>
      {/if}
    </div>

    <!-- Battery Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'battery'}>
      <SettingItem
        label={$tStore('settings.modules_battery_power')}
        icon={Battery}
        disabled={!hasBattery}
      >
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
            ariaLabel={$tStore('settings.modules_battery_power')}
          />
        </div>
      </SettingItem>
      {#if isBatteryLocked && shakeKey === 'battery'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>{$tStore('settings.modules_one_widget_required')}</span>
        </div>
      {/if}
    </div>

    <!-- Wireless & Connectivity Widget -->
    <div class="item-wrapper" class:shake={shakeKey === 'wireless'}>
      <SettingItem label={$tStore('settings.modules_wireless')} icon={Wifi}>
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
            ariaLabel={$tStore('settings.modules_wireless')}
          />
        </div>
      </SettingItem>
      {#if isWirelessLocked && shakeKey === 'wireless'}
        <div class="locked-warning">
          <AlertCircle size={12} />
          <span>{$tStore('settings.modules_one_widget_required')}</span>
        </div>
      {/if}
    </div>

    <SettingItem
      label={$tStore('settings.modules_action_glance')}
      icon={SlidersHorizontal}
    >
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.actionCenter}
        onchange={() => toggleHomeWidget('actionCenter')}
        ariaLabel={$tStore('settings.modules_action_glance')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.modules_stats_glance')}
      icon={Activity}
    >
      <ToggleSwitch
        checked={Boolean($homeWidgetsVisibilityStore.systemStats)}
        onchange={() => toggleHomeWidget('systemStats')}
        ariaLabel={$tStore('settings.modules_stats_glance')}
      />
    </SettingItem>

    <SettingItem label={$tStore('settings.modules_spotlight')} icon={Sparkles}>
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.spotlight !== false}
        onchange={() => toggleHomeWidget('spotlight')}
        ariaLabel={$tStore('settings.modules_spotlight')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.modules_announcements')}
      icon={Megaphone}
    >
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.announcements !== false}
        onchange={() => toggleHomeWidget('announcements')}
        ariaLabel={$tStore('settings.modules_announcements')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.widgets_media_label')}
      description={$tStore('settings.widgets_media_desc')}
      icon={Music}
      isDefault={$homeWidgetsVisibilityStore.media !== false}
      onReset={() =>
        homeWidgetsVisibilityStore.update((v) => ({ ...v, media: true }))}
    >
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.media !== false}
        onchange={() => toggleHomeWidget('media')}
        ariaLabel={$tStore('settings.widgets_media_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.widgets_menu_label')}
      description={$tStore('settings.widgets_menu_desc')}
      icon={MoreHorizontal}
      isDefault={$homeWidgetMenuEnabledStore ===
        homeWidgetMenuEnabledStore.defaultValue}
      onReset={() => homeWidgetMenuEnabledStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$homeWidgetMenuEnabledStore}
        ariaLabel={$tStore('settings.widgets_menu_label')}
      />
    </SettingItem>
  </BentoCard>

  <div class="setting-section-header">
    <span>{$tStore('settings.modules_dock_pinned')}</span>
  </div>

  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.modules_pin_calendar')}
      icon={Calendar}
    >
      <ToggleSwitch
        checked={$dockPinnedAppsStore.calendar}
        onchange={() => toggleDockPinned('calendar')}
        ariaLabel={$tStore('settings.modules_pin_calendar')}
      />
    </SettingItem>

    <SettingItem label={$tStore('settings.modules_pin_notes')} icon={FileText}>
      <ToggleSwitch
        checked={$dockPinnedAppsStore.notes}
        onchange={() => toggleDockPinned('notes')}
        ariaLabel={$tStore('settings.modules_pin_notes')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.modules_pin_calculator')}
      icon={Calculator}
    >
      <ToggleSwitch
        checked={$dockPinnedAppsStore.calculator}
        onchange={() => toggleDockPinned('calculator')}
        ariaLabel={$tStore('settings.modules_pin_calculator')}
      />
    </SettingItem>

    <SettingItem label={$tStore('settings.modules_pin_timer')} icon={Timer}>
      <ToggleSwitch
        checked={$dockPinnedAppsStore.timer}
        onchange={() => toggleDockPinned('timer')}
        ariaLabel={$tStore('settings.modules_pin_timer')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.modules_pin_fileshelf')}
      icon={FolderDown}
    >
      <ToggleSwitch
        checked={$dockPinnedAppsStore.fileshelf}
        onchange={() => toggleDockPinned('fileshelf')}
        ariaLabel={$tStore('settings.modules_pin_fileshelf')}
      />
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/AppsModulesTab.css"></style>
