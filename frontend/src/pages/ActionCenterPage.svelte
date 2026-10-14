<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 440, h: 380, r: 30 },
    borderColor: 'var(--c-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
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
  import BackButton from '../components/BackButton.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    wirelessDataStore,
    batteryDataStore,
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

  function openSystemDialog(target: string, e?: Event) {
    if (e) e.stopPropagation()
    if (typeof OpenWindowsSettings === 'function') {
      OpenWindowsSettings(target).catch(() => {})
    }
  }

  let batteryStateLabel = $derived.by(() => {
    if (!battery.hasBattery) return 'Desktop AC Power'
    if (battery.isCharging) return `Charging (${battery.percentage}%)`
    if (battery.state === 'full') return 'Fully Charged'
    if (battery.batterySaver) return `Battery Saver (${battery.percentage}%)`
    return `On Battery (${battery.percentage}%)`
  })

  function formatLifeTime(secs: number): string {
    if (secs <= 0 || secs >= 86400) return ''
    const hrs = Math.floor(secs / 3600)
    const mins = Math.floor((secs % 3600) / 60)
    return hrs > 0 ? `${hrs}h ${mins}m left` : `${mins}m left`
  }

  let remainingStr = $derived(formatLifeTime(battery.lifeTimeSeconds))
</script>

<div class="action-center-page">
  <!-- Header -->
  <div class="action-header">
    <BackButton />
    <span class="header-title">Action Center</span>
    <div class="header-spacer"></div>
  </div>

  <!-- Bento Action Cards Grid -->
  <div class="action-grid">
    <!-- Wi-Fi Card -->
    <div class="bento-action-card">
      <div class="card-left">
        <div
          class="card-icon-circle"
          class:wifi-active={wireless.wifiEnabled &&
            wireless.wifiConnected &&
            !wireless.airplaneMode}
        >
          {#if wireless.wifiEnabled && !wireless.airplaneMode}
            <Wifi size={18} />
          {:else}
            <WifiOff size={18} />
          {/if}
        </div>
        <div class="card-info">
          <div class="card-title-row">
            <span class="card-primary-text">Wi-Fi</span>
            <button
              class="quick-open-btn"
              use:islandTooltip={'Open Windows Network Flyout'}
              onclick={(e) => openSystemDialog('network', e)}
            >
              <ExternalLink size={11} />
            </button>
          </div>
          <span class="card-secondary-text">
            {#if wireless.airplaneMode}
              Turned off by Airplane mode
            {:else if wireless.wifiEnabled}
              {wireless.wifiConnected
                ? `${wireless.wifiSsid || 'Connected'} (${wireless.wifiSignal}%)`
                : 'Disconnected'}
            {:else}
              Off
            {/if}
          </span>
        </div>
      </div>
      <label
        class="toggle-switch"
        class:disabled-switch={wireless.airplaneMode}
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

    <!-- Bluetooth Card -->
    <div class="bento-action-card">
      <div class="card-left">
        <div
          class="card-icon-circle"
          class:bt-active={wireless.bluetoothEnabled &&
            wireless.bluetoothConnected &&
            !wireless.airplaneMode}
        >
          {#if wireless.bluetoothEnabled && !wireless.airplaneMode}
            <Bluetooth size={18} />
          {:else}
            <BluetoothOff size={18} />
          {/if}
        </div>
        <div class="card-info">
          <div class="card-title-row">
            <span class="card-primary-text">Bluetooth</span>
            <button
              class="quick-open-btn"
              use:islandTooltip={'Open Windows Bluetooth Settings'}
              onclick={(e) => openSystemDialog('bluetooth', e)}
            >
              <ExternalLink size={11} />
            </button>
          </div>
          <span class="card-secondary-text">
            {#if !wireless.hasBluetooth}
              No Bluetooth adapter installed
            {:else if wireless.airplaneMode}
              Turned off by Airplane mode
            {:else if wireless.bluetoothEnabled}
              {#if wireless.bluetoothConnected}
                <span>{wireless.bluetoothDeviceName || 'Connected'}</span>
                {#if wireless.bluetoothBattery !== undefined}
                  <span class="device-battery-badge"
                    >{wireless.bluetoothBattery}%</span
                  >
                {/if}
              {:else}
                Ready to connect
              {/if}
            {:else}
              Off
            {/if}
          </span>
        </div>
      </div>
      <label
        class="toggle-switch"
        class:disabled-switch={!wireless.hasBluetooth || wireless.airplaneMode}
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

    <!-- Battery & Power Card -->
    <div class="bento-action-card">
      <div class="card-left">
        <div
          class="card-icon-circle"
          class:battery-charging={battery.isCharging}
          class:battery-low={!battery.isCharging && battery.percentage <= 15}
        >
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
          <div class="card-title-row">
            <span class="card-primary-text">Battery & Power</span>
            <button
              class="quick-open-btn"
              use:islandTooltip={'Open Windows Power Settings'}
              onclick={(e) => openSystemDialog('battery', e)}
            >
              <ExternalLink size={11} />
            </button>
          </div>
          <span class="card-secondary-text">
            {batteryStateLabel}{remainingStr ? ` • ${remainingStr}` : ''}
          </span>
        </div>
      </div>
      {#if battery.hasBattery}
        <label class="toggle-switch" use:islandTooltip={'Toggle Battery Saver Mode'}>
          <input
            type="checkbox"
            checked={battery.batterySaver}
            onchange={toggleBatterySaver}
          />
          <span class="slider"></span>
        </label>
      {/if}
    </div>

    <!-- Airplane Mode Card -->
    <div class="bento-action-card">
      <div class="card-left">
        <div
          class="card-icon-circle"
          class:airplane-active={wireless.airplaneMode}
        >
          <Plane size={18} />
        </div>
        <div class="card-info">
          <div class="card-title-row">
            <span class="card-primary-text">Airplane Mode</span>
            <button
              class="quick-open-btn"
              use:islandTooltip={'Open Windows Airplane Mode Settings'}
              onclick={(e) => openSystemDialog('airplane', e)}
            >
              <ExternalLink size={11} />
            </button>
          </div>
          <span class="card-secondary-text">
            {wireless.airplaneMode
              ? 'All wireless radios disabled'
              : 'Radios operating normally'}
          </span>
        </div>
      </div>
      <label class="toggle-switch">
        <input
          type="checkbox"
          checked={wireless.airplaneMode}
          onchange={toggleAirplaneMode}
        />
        <span class="slider"></span>
      </label>
    </div>
  </div>
</div>

<style src="./styles/ActionCenterPage.css"></style>
