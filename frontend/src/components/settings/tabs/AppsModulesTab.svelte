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
    MoreHorizontal,
    RotateCcw
  } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
  import SettingItem from '../SettingItem.svelte'
  import {
    pageVisibilityStore,
    defaultPageVisibility,
    homeWidgetsVisibilityStore,
    defaultHomeWidgetsVisibility,
    homeWidgetSizesStore,
    homeWidgetMenuEnabledStore,
    dockPinnedAppsStore,
    defaultDockPinnedApps,
    batteryDataStore,
    tStore,
    type HomeWidgetsVisibility,
    type PageVisibility,
    type DockPinnedApps
  } from '../../../store'

  let hasBattery = $derived($batteryDataStore.hasBattery)

  function isRecordEqual<T extends object>(a: T, b: T): boolean {
    const keys = new Set([...Object.keys(a), ...Object.keys(b)]) as Set<keyof T>
    for (const k of keys) {
      if (a[k] !== b[k]) return false
    }
    return true
  }

  let isPagesDefault = $derived(
    isRecordEqual($pageVisibilityStore, defaultPageVisibility)
  )

  let isHomeWidgetsDefault = $derived(
    isRecordEqual($homeWidgetsVisibilityStore, defaultHomeWidgetsVisibility) &&
      $homeWidgetMenuEnabledStore === homeWidgetMenuEnabledStore.defaultValue
  )

  let isDockPinnedDefault = $derived(
    isRecordEqual($dockPinnedAppsStore, defaultDockPinnedApps)
  )

  let isAllDefault = $derived(
    isPagesDefault && isHomeWidgetsDefault && isDockPinnedDefault
  )

  function resetPages() {
    pageVisibilityStore.reset()
  }

  function resetHomeWidgets() {
    homeWidgetsVisibilityStore.reset()
    homeWidgetSizesStore.reset()
    homeWidgetMenuEnabledStore.reset()
  }

  function resetDockPinned() {
    dockPinnedAppsStore.reset()
  }

  function resetAllAppsAndWidgets() {
    resetPages()
    resetHomeWidgets()
    resetDockPinned()
  }

  function toggleHomeWidget(key: keyof HomeWidgetsVisibility) {
    const raw = $homeWidgetsVisibilityStore[key]
    const isCurrentlyActive =
      key === 'spotlight' ||
      key === 'announcements' ||
      key === 'media' ||
      key === 'notifications'
        ? raw !== false
        : Boolean(raw)
    $homeWidgetsVisibilityStore = {
      ...$homeWidgetsVisibilityStore,
      [key]: !isCurrentlyActive
    }
  }

  function togglePage(key: keyof PageVisibility) {
    const raw = $pageVisibilityStore[key]
    const isCurrentlyActive =
      key === 'calendar' || key === 'performance' ? raw !== false : Boolean(raw)
    $pageVisibilityStore = {
      ...$pageVisibilityStore,
      [key]: !isCurrentlyActive
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
  <div class="setting-section-header section-header-row">
    <span>{$tStore('settings.modules_pages_features')}</span>
    {#if !isPagesDefault}
      <button type="button" class="section-reset-link" onclick={resetPages}>
        <RotateCcw size={11} strokeWidth={2.2} />
        <span>{$tStore('common.reset')}</span>
      </button>
    {/if}
  </div>

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

  <div class="setting-section-header section-header-row">
    <span>{$tStore('settings.modules_home_widgets')}</span>
    {#if !isHomeWidgetsDefault}
      <button
        type="button"
        class="section-reset-link"
        onclick={resetHomeWidgets}
      >
        <RotateCcw size={11} strokeWidth={2.2} />
        <span>{$tStore('common.reset')}</span>
      </button>
    {/if}
  </div>

  <BentoCard variant="group">
    <SettingItem label={$tStore('settings.modules_clock_time')} icon={Clock}>
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.clock}
        onchange={() => toggleHomeWidget('clock')}
        ariaLabel={$tStore('settings.modules_clock_time')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.modules_calendar_date')}
      icon={Calendar}
    >
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.calendar}
        onchange={() => toggleHomeWidget('calendar')}
        ariaLabel={$tStore('settings.modules_calendar_date')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.modules_weather_forecast')}
      icon={CloudSun}
    >
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.weather}
        onchange={() => toggleHomeWidget('weather')}
        ariaLabel={$tStore('settings.modules_weather_forecast')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.modules_battery_power')}
      icon={Battery}
      disabled={!hasBattery}
    >
      <ToggleSwitch
        checked={hasBattery && $homeWidgetsVisibilityStore.battery}
        disabled={!hasBattery}
        onchange={() => toggleHomeWidget('battery')}
        ariaLabel={$tStore('settings.modules_battery_power')}
      />
    </SettingItem>

    <SettingItem label={$tStore('settings.modules_wireless')} icon={Wifi}>
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.wireless}
        onchange={() => toggleHomeWidget('wireless')}
        ariaLabel={$tStore('settings.modules_wireless')}
      />
    </SettingItem>

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
      label={$tStore('settings.modules_notifications_glance')}
      icon={Bell}
    >
      <ToggleSwitch
        checked={$homeWidgetsVisibilityStore.notifications !== false}
        onchange={() => toggleHomeWidget('notifications')}
        ariaLabel={$tStore('settings.modules_notifications_glance')}
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

  <div class="setting-section-header section-header-row">
    <span>{$tStore('settings.modules_dock_pinned')}</span>
    {#if !isDockPinnedDefault}
      <button
        type="button"
        class="section-reset-link"
        onclick={resetDockPinned}
      >
        <RotateCcw size={11} strokeWidth={2.2} />
        <span>{$tStore('common.reset')}</span>
      </button>
    {/if}
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

  <BentoCard variant="group">
    <SettingItem label={$tStore('settings.modules_reset_all')} icon={RotateCcw}>
      <button
        type="button"
        class="reset-all-btn"
        disabled={isAllDefault}
        onclick={resetAllAppsAndWidgets}
      >
        <RotateCcw size={12} strokeWidth={2.2} />
        <span>{$tStore('common.reset')}</span>
      </button>
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/AppsModulesTab.css"></style>
