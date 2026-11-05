<script lang="ts">
  import { getContext } from 'svelte'
  import {
    SlidersHorizontal,
    Wifi,
    WifiOff,
    Bluetooth,
    BluetoothOff,
    Plane,
    Leaf,
    Battery,
    ChevronRight
  } from 'lucide-svelte'
  import {
    wirelessDataStore,
    batteryDataStore,
    type WirelessData,
    type BatteryData,
    type IslandConfig
  } from '../store'
  import ActionCenterPage, {
    config as actionCenterConfig
  } from '../pages/ActionCenterPage.svelte'
  import {
    SetWifiEnabled,
    SetBluetoothEnabled,
    SetAirplaneMode,
    SetBatterySaver,
    GetWirelessStatus,
    GetBatteryStatus
  } from '../../bindings/Vinta/backend/app.js'
  import { islandTooltip } from '../services/islandTooltip'
  import BentoCard from './BentoCard.svelte'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  function handleOpenActionCenter() {
    if (typeof navigate === 'function') {
      navigate(ActionCenterPage, actionCenterConfig, true)
    }
  }

  let wireless = $derived($wirelessDataStore)
  let battery = $derived($batteryDataStore)

  async function refreshWireless() {
    if (typeof GetWirelessStatus === 'function') {
      try {
        const status = await GetWirelessStatus()
        if (status) {
          wirelessDataStore.set(status as unknown as WirelessData)
        }
      } catch {}
    }
  }

  async function refreshBattery() {
    if (typeof GetBatteryStatus === 'function') {
      try {
        const status = await GetBatteryStatus()
        if (status) {
          batteryDataStore.set(status as unknown as BatteryData)
        }
      } catch {}
    }
  }

  async function handleToggleWifi(e: MouseEvent) {
    e.stopPropagation()
    const next = !wireless.wifiEnabled
    wirelessDataStore.update((w) => ({
      ...w,
      wifiEnabled: next,
      airplaneMode: next ? false : w.airplaneMode,
      wifiConnected: next ? w.wifiConnected : false
    }))
    if (typeof SetWifiEnabled === 'function') {
      try {
        await SetWifiEnabled(next)
        setTimeout(refreshWireless, 600)
      } catch {}
    }
  }

  async function handleToggleBluetooth(e: MouseEvent) {
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
      try {
        await SetBluetoothEnabled(next)
        setTimeout(refreshWireless, 600)
      } catch {}
    }
  }

  async function handleToggleAirplane(e: MouseEvent) {
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
      try {
        await SetAirplaneMode(next)
        setTimeout(refreshWireless, 600)
      } catch {}
    }
  }

  async function handleToggleSaver(e: MouseEvent) {
    e.stopPropagation()
    if (!battery.hasBattery) return
    const next = !battery.batterySaver
    batteryDataStore.update((b) => ({
      ...b,
      batterySaver: next
    }))
    if (typeof SetBatterySaver === 'function') {
      try {
        await SetBatterySaver(next)
        setTimeout(refreshBattery, 600)
      } catch {}
    }
  }

  let isWifiActive = $derived(wireless.wifiEnabled && !wireless.airplaneMode)
  let isBtActive = $derived(
    wireless.hasBluetooth && wireless.bluetoothEnabled && !wireless.airplaneMode
  )
  let isAirplaneActive = $derived(wireless.airplaneMode)
  let isSaverActive = $derived(battery.hasBattery && battery.batterySaver)

  function computeWifiTooltip(): string {
    if (wireless.airplaneMode) return 'Wi-Fi: Disabled by Airplane mode'
    if (!wireless.wifiEnabled) return 'Wi-Fi: Off (click to enable)'
    if (wireless.wifiConnected && wireless.wifiSsid) {
      return `Wi-Fi: Connected to ${wireless.wifiSsid}${wireless.wifiSignal > 0 ? ` (${wireless.wifiSignal}%)` : ''}`
    }
    return 'Wi-Fi: On (click to disable)'
  }
  let wifiTooltip = $derived.by(computeWifiTooltip)

  function computeBtTooltip(): string {
    if (!wireless.hasBluetooth) return 'Bluetooth: Not supported'
    if (wireless.airplaneMode) return 'Bluetooth: Disabled by Airplane mode'
    if (!wireless.bluetoothEnabled) return 'Bluetooth: Off (click to enable)'
    if (wireless.bluetoothConnected && wireless.bluetoothDeviceName) {
      return `Bluetooth: Connected to ${wireless.bluetoothDeviceName}`
    }
    return 'Bluetooth: On (click to disable)'
  }
  let btTooltip = $derived.by(computeBtTooltip)

  function computeAirplaneTooltip(): string {
    return isAirplaneActive
      ? 'Airplane Mode: On (click to disable)'
      : 'Airplane Mode: Off (click to enable)'
  }
  let airplaneTooltip = $derived.by(computeAirplaneTooltip)

  function computeSaverTooltip(): string {
    if (!battery.hasBattery) return 'Power: Desktop AC Connected'
    return isSaverActive
      ? `Battery Saver: On (${battery.percentage}%)`
      : `Battery Saver: Off (${battery.percentage}%)`
  }
  let saverTooltip = $derived.by(computeSaverTooltip)
</script>

<BentoCard variant="bento" class="action-center-glance-card">
  <div
    class="glance-header"
    role="button"
    tabindex="0"
    onclick={handleOpenActionCenter}
    onkeydown={(e) => e.key === 'Enter' && handleOpenActionCenter()}
  >
    <div class="glance-title-group">
      <div class="glance-icon-badge">
        <SlidersHorizontal size={12} />
      </div>
      <span class="glance-title-text">Quick Controls</span>
    </div>
    <ChevronRight size={14} class="glance-nav-chevron" />
  </div>

  <div class="quick-toggle-row">
    <!-- Wi-Fi Toggle -->
    <button
      type="button"
      class="quick-toggle-btn"
      class:active={isWifiActive}
      class:disabled={wireless.airplaneMode}
      onclick={handleToggleWifi}
      use:islandTooltip={wifiTooltip}
      aria-label="Toggle Wi-Fi"
    >
      <div class="toggle-icon-wrap">
        {#if isWifiActive}
          <Wifi size={18} />
        {:else}
          <WifiOff size={18} />
        {/if}
      </div>
      <span class="toggle-label">Wi-Fi</span>
    </button>

    <!-- Bluetooth Toggle -->
    <button
      type="button"
      class="quick-toggle-btn"
      class:active={isBtActive}
      class:disabled={!wireless.hasBluetooth || wireless.airplaneMode}
      onclick={handleToggleBluetooth}
      use:islandTooltip={btTooltip}
      aria-label="Toggle Bluetooth"
    >
      <div class="toggle-icon-wrap">
        {#if isBtActive}
          <Bluetooth size={18} />
        {:else}
          <BluetoothOff size={18} />
        {/if}
      </div>
      <span class="toggle-label">Bluetooth</span>
    </button>

    <!-- Airplane Mode Toggle -->
    <button
      type="button"
      class="quick-toggle-btn"
      class:active={isAirplaneActive}
      onclick={handleToggleAirplane}
      use:islandTooltip={airplaneTooltip}
      aria-label="Toggle Airplane Mode"
    >
      <div class="toggle-icon-wrap">
        <Plane size={18} />
      </div>
      <span class="toggle-label">Airplane</span>
    </button>

    <!-- Battery Saver Toggle -->
    <button
      type="button"
      class="quick-toggle-btn"
      class:active={isSaverActive}
      class:disabled={!battery.hasBattery}
      onclick={handleToggleSaver}
      use:islandTooltip={saverTooltip}
      aria-label="Toggle Battery Saver"
    >
      <div class="toggle-icon-wrap">
        {#if isSaverActive}
          <Leaf size={18} />
        {:else}
          <Battery size={18} />
        {/if}
      </div>
      <span class="toggle-label">Saver</span>
    </button>
  </div>
</BentoCard>

<style src="./styles/ActionCenterGlance.css"></style>
