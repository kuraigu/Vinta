<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 440, h: 380, r: 30 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  /**
   * @fileoverview Quick settings and system hardware controls view.
   *
   * Provides quick toggles for Windows Wi-Fi flyouts, Bluetooth settings, battery saver,
   * and airplane mode, while monitoring live battery percentages and charge status.
   */

  import {
    Wifi,
    WifiOff,
    Bluetooth,
    BluetoothOff,
    Battery,
    Zap,
    Leaf,
    Plug,
    Plane,
    ExternalLink
  } from 'lucide-svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import BackButton from '../components/BackButton.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    wirelessDataStore,
    batteryDataStore,
    tStore,
    type WirelessData,
    type BatteryData
  } from '../store'
  import {
    OpenWindowsSettings,
    SetWifiEnabled,
    SetBluetoothEnabled,
    SetAirplaneMode,
    SetBatterySaver,
    GetWirelessStatus,
    GetBatteryStatus
  } from '../../bindings/Vinta/backend/app.js'

  let wireless = $derived($wirelessDataStore)
  let battery = $derived($batteryDataStore)

  /**
   * Queries the backend for latest wireless adapter states (Wi-Fi, Bluetooth, Airplane Mode).
   */
  async function refreshWireless() {
    if (typeof GetWirelessStatus === 'function') {
      try {
        const status = await GetWirelessStatus()
        if (status) {
          wirelessDataStore.set(status as unknown as WirelessData)
        }
      } catch {
        // Ignore status query error
      }
    }
  }

  /**
   * Queries the backend for real-time battery charge percentage and AC power connectivity.
   */
  async function refreshBattery() {
    if (typeof GetBatteryStatus === 'function') {
      try {
        const status = await GetBatteryStatus()
        if (status) {
          batteryDataStore.set(status as unknown as BatteryData)
        }
      } catch {
        // Ignore status query error
      }
    }
  }

  /**
   * Dispatches a request to open the native Windows Quick Settings network flyout.
   *
   * @param e - Click or key event triggering the action.
   */
  async function toggleWifi(e: Event) {
    e.stopPropagation()
    const next = !wireless.wifiEnabled
    wirelessDataStore.update((w) => ({
      ...w,
      wifiEnabled: next,
      airplaneMode: next ? false : w.airplaneMode,
      wifiConnected: next ? w.wifiConnected : false
    }))
    if (typeof SetWifiEnabled === 'function') {
      await SetWifiEnabled(next).catch(() => {})
      setTimeout(refreshWireless, 600)
    }
  }

  /**
   * Dispatches a request to open the native Windows Bluetooth settings panel.
   *
   * @param e - Click or key event triggering the action.
   */
  async function toggleBluetooth(e: Event) {
    e.stopPropagation()
    if (!wireless.hasBluetooth) return
    const next = !wireless.bluetoothEnabled
    wirelessDataStore.update((w) => ({
      ...w,
      bluetoothEnabled: next,
      airplaneMode: next ? false : w.airplaneMode,
      bluetoothConnected: next ? w.bluetoothConnected : false
    }))
    if (typeof SetBluetoothEnabled === 'function') {
      await SetBluetoothEnabled(next).catch(() => {})
      setTimeout(refreshWireless, 600)
    }
  }

  /**
   * Toggles Windows Airplane Mode state via the backend radio manager.
   *
   * @param e - Click or key event triggering the action.
   */
  async function toggleAirplaneMode(e: Event) {
    e.stopPropagation()
    const next = !wireless.airplaneMode
    wirelessDataStore.update((w) => ({
      ...w,
      airplaneMode: next,
      wifiEnabled: next ? false : true,
      bluetoothEnabled: next ? false : w.bluetoothEnabled,
      wifiConnected: next ? false : w.wifiConnected
    }))
    if (typeof SetAirplaneMode === 'function') {
      await SetAirplaneMode(next).catch(() => {})
      setTimeout(refreshWireless, 600)
    }
  }

  /**
   * Toggles Windows Battery Saver power scheme mode.
   *
   * @param e - Click or key event triggering the action.
   */
  async function toggleBatterySaver(e: Event) {
    e.stopPropagation()
    const next = !battery.batterySaver
    batteryDataStore.update((b) => ({
      ...b,
      batterySaver: next
    }))
    if (typeof SetBatterySaver === 'function') {
      await SetBatterySaver(next).catch(() => {})
      setTimeout(refreshBattery, 600)
    }
  }

  /**
   * Launches an external Windows settings URI (e.g. ms-settings:network, ms-settings:powersleep).
   *
   * @param target - Target URI scheme string.
   * @param e - Optional DOM event to stop propagation.
   */
  function openSystemDialog(target: string, e?: Event) {
    if (e) e.stopPropagation()
    if (typeof OpenWindowsSettings === 'function') {
      OpenWindowsSettings(target).catch(() => {})
    }
  }

  let batteryStateLabel = $derived.by(() => {
    if (!battery.hasBattery) return $tStore('action_center.desktop_ac')
    if (battery.isCharging)
      return $tStore('action_center.charging', { percent: battery.percentage })
    if (battery.state === 'full') return $tStore('action_center.fully_charged')
    if (battery.batterySaver)
      return $tStore('action_center.battery_saver_pct', {
        percent: battery.percentage
      })
    return $tStore('action_center.on_battery', { percent: battery.percentage })
  })

  /**
   * Converts remaining battery discharge seconds into hours and minutes display.
   *
   * @param secs - Estimated remaining battery lifetime in seconds.
   * @returns Localized time remaining string.
   */
  function formatLifeTime(secs: number): string {
    if (secs <= 0 || secs >= 86400) return ''
    const hrs = Math.floor(secs / 3600)
    const mins = Math.floor((secs % 3600) / 60)
    return hrs > 0
      ? $tStore('action_center.hours_mins_left', { hours: hrs, mins })
      : $tStore('action_center.mins_left', { mins })
  }

  let remainingStr = $derived(formatLifeTime(battery.lifeTimeSeconds))
</script>

<div data-vinta="action-center-page" class="action-center-page">
  <!-- Header -->
  <div class="action-header">
    <BackButton />
    <span class="header-title">{$tStore('action_center.title')}</span>
    <div class="header-spacer"></div>
  </div>

  <!-- Bento Action Cards Grid -->
  <div class="action-grid">
    <!-- Wi-Fi Card -->
    <BentoCard
      variant="bento"
      class="bento-action-card"
      onclick={(e) => {
        if (!wireless.airplaneMode) toggleWifi(e)
      }}
    >
      <div class="card-left">
        <div
          class="card-icon-circle"
          class:is-off={!wireless.wifiEnabled || wireless.airplaneMode}
        >
          {#if wireless.wifiEnabled && !wireless.airplaneMode}
            <Wifi size={18} />
          {:else}
            <WifiOff size={18} />
          {/if}
        </div>
        <div class="card-info">
          <span class="card-primary-text">{$tStore('action_center.wifi')}</span>
          <span class="card-secondary-text">
            {#if wireless.airplaneMode}
              {$tStore('action_center.turned_off_airplane')}
            {:else if wireless.wifiEnabled}
              {wireless.wifiConnected
                ? `${wireless.wifiSsid || $tStore('action_center.connected')} (${wireless.wifiSignal}%)`
                : $tStore('action_center.disconnected')}
            {:else}
              {$tStore('common.off')}
            {/if}
          </span>
        </div>
      </div>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <div class="card-controls-right">
        <button
          type="button"
          class="quick-open-btn"
          use:islandTooltip={$tStore('action_center.open_network')}
          onclick={(e) => openSystemDialog('network', e)}
          aria-label={$tStore('action_center.open_network')}
        >
          <ExternalLink size={13} />
        </button>
        <label
          class="toggle-switch"
          class:disabled-switch={wireless.airplaneMode}
          onclick={(e) => e.stopPropagation()}
        >
          <input
            type="checkbox"
            checked={wireless.wifiEnabled && !wireless.airplaneMode}
            disabled={wireless.airplaneMode}
            onchange={toggleWifi}
          />
          <span class="slider"></span>
        </label>
      </div>
    </BentoCard>

    <!-- Bluetooth Card -->
    <BentoCard
      variant="bento"
      class="bento-action-card"
      onclick={(e) => {
        if (wireless.hasBluetooth && !wireless.airplaneMode) toggleBluetooth(e)
      }}
    >
      <div class="card-left">
        <div
          class="card-icon-circle"
          class:is-off={!wireless.bluetoothEnabled || wireless.airplaneMode}
        >
          {#if wireless.bluetoothEnabled && !wireless.airplaneMode}
            <Bluetooth size={18} />
          {:else}
            <BluetoothOff size={18} />
          {/if}
        </div>
        <div class="card-info">
          <span class="card-primary-text"
            >{$tStore('action_center.bluetooth')}</span
          >
          <span class="card-secondary-text">
            {#if !wireless.hasBluetooth}
              {$tStore('action_center.no_bluetooth')}
            {:else if wireless.airplaneMode}
              {$tStore('action_center.turned_off_airplane')}
            {:else if wireless.bluetoothEnabled}
              {#if wireless.bluetoothConnected}
                <span
                  >{wireless.bluetoothDeviceName ||
                    $tStore('action_center.connected')}</span
                >
                {#if wireless.bluetoothBattery !== undefined}
                  <span class="device-battery-badge"
                    >{wireless.bluetoothBattery}%</span
                  >
                {/if}
              {:else}
                {$tStore('action_center.ready_to_connect')}
              {/if}
            {:else}
              {$tStore('common.off')}
            {/if}
          </span>
        </div>
      </div>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <div class="card-controls-right">
        <button
          type="button"
          class="quick-open-btn"
          use:islandTooltip={$tStore('action_center.open_bluetooth')}
          onclick={(e) => openSystemDialog('bluetooth', e)}
          aria-label={$tStore('action_center.open_bluetooth')}
        >
          <ExternalLink size={13} />
        </button>
        <label
          class="toggle-switch"
          class:disabled-switch={!wireless.hasBluetooth ||
            wireless.airplaneMode}
          onclick={(e) => e.stopPropagation()}
        >
          <input
            type="checkbox"
            checked={wireless.hasBluetooth &&
              wireless.bluetoothEnabled &&
              !wireless.airplaneMode}
            disabled={!wireless.hasBluetooth || wireless.airplaneMode}
            onchange={toggleBluetooth}
          />
          <span class="slider"></span>
        </label>
      </div>
    </BentoCard>

    <!-- Battery & Power Card -->
    <BentoCard
      variant="bento"
      class="bento-action-card"
      onclick={(e) => {
        if (battery.hasBattery) toggleBatterySaver(e)
      }}
    >
      <div class="card-left">
        <div class="card-icon-circle">
          {#if !battery.hasBattery}
            <Plug size={18} />
          {:else if battery.isCharging}
            <Zap size={18} />
          {:else if battery.batterySaver}
            <Leaf size={18} />
          {:else}
            <Battery size={18} />
          {/if}
        </div>
        <div class="card-info">
          <span class="card-primary-text"
            >{$tStore('action_center.battery_power')}</span
          >
          <span class="card-secondary-text">
            {batteryStateLabel}{remainingStr ? ` • ${remainingStr}` : ''}
          </span>
        </div>
      </div>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <div class="card-controls-right">
        <button
          type="button"
          class="quick-open-btn"
          use:islandTooltip={$tStore('action_center.open_power')}
          onclick={(e) => openSystemDialog('battery', e)}
          aria-label={$tStore('action_center.open_power')}
        >
          <ExternalLink size={13} />
        </button>
        {#if battery.hasBattery}
          <label class="toggle-switch" onclick={(e) => e.stopPropagation()}>
            <input
              type="checkbox"
              checked={battery.batterySaver}
              onchange={toggleBatterySaver}
            />
            <span class="slider"></span>
          </label>
        {/if}
      </div>
    </BentoCard>

    <!-- Airplane Mode Card -->
    <BentoCard
      variant="bento"
      class="bento-action-card"
      onclick={toggleAirplaneMode}
    >
      <div class="card-left">
        <div class="card-icon-circle" class:is-off={!wireless.airplaneMode}>
          <Plane size={18} />
        </div>
        <div class="card-info">
          <span class="card-primary-text"
            >{$tStore('action_center.airplane_mode')}</span
          >
          <span class="card-secondary-text">
            {wireless.airplaneMode
              ? $tStore('action_center.radios_disabled')
              : $tStore('action_center.radios_normal')}
          </span>
        </div>
      </div>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <div class="card-controls-right">
        <button
          type="button"
          class="quick-open-btn"
          use:islandTooltip={$tStore('action_center.open_airplane')}
          onclick={(e) => openSystemDialog('airplane', e)}
          aria-label={$tStore('action_center.open_airplane')}
        >
          <ExternalLink size={13} />
        </button>
        <label class="toggle-switch" onclick={(e) => e.stopPropagation()}>
          <input
            type="checkbox"
            checked={wireless.airplaneMode}
            onchange={toggleAirplaneMode}
          />
          <span class="slider"></span>
        </label>
      </div>
    </BentoCard>
  </div>
</div>

<style src="./styles/ActionCenterPage.css"></style>
