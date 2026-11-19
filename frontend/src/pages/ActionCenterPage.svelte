<script lang="ts" module>
  import { ISLAND_WIDTH_TIERS, type IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: ISLAND_WIDTH_TIERS.standard, h: 380, r: 28 },
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
    Plane
  } from 'lucide-svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import PageHeader from '../components/PageHeader.svelte'
  import IconBadge from '../components/IconBadge.svelte'
  import ToggleSwitch from '../components/ToggleSwitch.svelte'
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
  <PageHeader title={$tStore('action_center.title')} />

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
        <IconBadge
          icon={wireless.wifiEnabled && !wireless.airplaneMode ? Wifi : WifiOff}
          color={wireless.wifiEnabled && !wireless.airplaneMode
            ? 'blue'
            : 'slate'}
          size="lg"
        />
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
      <div class="card-controls-right">
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div onclick={(e) => e.stopPropagation()}>
          <ToggleSwitch
            checked={wireless.wifiEnabled && !wireless.airplaneMode}
            disabled={wireless.airplaneMode}
            onchange={(_, e) => toggleWifi(e)}
            ariaLabel={$tStore('action_center.wifi')}
          />
        </div>
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
        <IconBadge
          icon={wireless.bluetoothEnabled && !wireless.airplaneMode
            ? Bluetooth
            : BluetoothOff}
          color={wireless.bluetoothEnabled && !wireless.airplaneMode
            ? 'indigo'
            : 'slate'}
          size="lg"
        />
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
      <div class="card-controls-right">
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div onclick={(e) => e.stopPropagation()}>
          <ToggleSwitch
            checked={wireless.hasBluetooth &&
              wireless.bluetoothEnabled &&
              !wireless.airplaneMode}
            disabled={!wireless.hasBluetooth || wireless.airplaneMode}
            onchange={(_, e) => toggleBluetooth(e)}
            ariaLabel={$tStore('action_center.bluetooth')}
          />
        </div>
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
        <IconBadge
          icon={!battery.hasBattery
            ? Plug
            : battery.isCharging
              ? Zap
              : battery.batterySaver
                ? Leaf
                : Battery}
          color={battery.batterySaver || battery.isCharging
            ? 'emerald'
            : 'teal'}
          size="lg"
        />
        <div class="card-info">
          <span class="card-primary-text"
            >{$tStore('action_center.battery_power')}</span
          >
          <span class="card-secondary-text">
            {batteryStateLabel}{remainingStr ? ` • ${remainingStr}` : ''}
          </span>
        </div>
      </div>
      <div class="card-controls-right">
        {#if battery.hasBattery}
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div onclick={(e) => e.stopPropagation()}>
            <ToggleSwitch
              checked={battery.batterySaver}
              onchange={(_, e) => toggleBatterySaver(e)}
              ariaLabel={$tStore('action_center.battery_power')}
            />
          </div>
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
        <IconBadge
          icon={Plane}
          color={wireless.airplaneMode ? 'amber' : 'slate'}
          size="lg"
        />
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
      <div class="card-controls-right">
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div onclick={(e) => e.stopPropagation()}>
          <ToggleSwitch
            checked={wireless.airplaneMode}
            onchange={(_, e) => toggleAirplaneMode(e)}
            ariaLabel={$tStore('action_center.airplane_mode')}
          />
        </div>
      </div>
    </BentoCard>
  </div>
</div>

<style src="./styles/ActionCenterPage.css"></style>
