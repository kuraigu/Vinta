<script lang="ts">
  import {
    Wifi,
    WifiOff,
    Bluetooth,
    Plane,
    Radio,
    RefreshCw
  } from 'lucide-svelte'
  import {
    wirelessDataStore,
    defaultWirelessData,
    isSimulatingWirelessStore,
    type WirelessData
  } from '../store'
  import { Events } from '$lib/events'
  import { GetWirelessStatus } from '../../bindings/Vinta/backend/app.js'
  import { islandTooltip } from '../services/islandTooltip'

  let wireless = $derived($wirelessDataStore)
  let isSimulating = $derived($isSimulatingWirelessStore)

  function broadcastWirelessEffect() {
    const isSim = $isSimulatingWirelessStore
    const data = $wirelessDataStore
    try {
      Events.Emit('devtest-wireless', { isSimulating: isSim, data })
    } catch {}
  }
  $effect(broadcastWirelessEffect)

  function resetToDefaults() {
    isSimulatingWirelessStore.set(true)
    wirelessDataStore.set(JSON.parse(JSON.stringify(defaultWirelessData)))
  }

  async function resetToLive() {
    isSimulatingWirelessStore.set(false)
    try {
      if (typeof GetWirelessStatus === 'function') {
        const live = await GetWirelessStatus()
        if (live) {
          wirelessDataStore.set(live as unknown as WirelessData)
        }
      }
    } catch {
      // Fallback
    }
  }

  function setSignalPreset(level: number) {
    isSimulatingWirelessStore.set(true)
    wirelessDataStore.update((w) => ({
      ...w,
      wifiSignal: level,
      wifiConnected: level > 0,
      wifiEnabled: true,
      airplaneMode: false
    }))
  }

  function setPreset(preset: 'wifi_only' | 'wifi_bt' | 'airplane' | 'offline') {
    isSimulatingWirelessStore.set(true)
    if (preset === 'wifi_only') {
      wirelessDataStore.set({
        hasWifi: true,
        wifiEnabled: true,
        wifiConnected: true,
        wifiSsid: 'Home_5GHz',
        wifiSignal: 50,
        hasBluetooth: false,
        bluetoothEnabled: false,
        bluetoothConnected: false,
        bluetoothDeviceName: '',
        bluetoothBattery: undefined,
        airplaneMode: false
      })
    } else if (preset === 'wifi_bt') {
      wirelessDataStore.set({
        hasWifi: true,
        wifiEnabled: true,
        wifiConnected: true,
        wifiSsid: 'Office_Mesh_5G',
        wifiSignal: 75,
        hasBluetooth: true,
        bluetoothEnabled: true,
        bluetoothConnected: true,
        bluetoothDeviceName: 'Wireless Device',
        bluetoothBattery: 85,
        airplaneMode: false
      })
    } else if (preset === 'airplane') {
      wirelessDataStore.update((w) => ({
        ...w,
        airplaneMode: true,
        wifiEnabled: false,
        bluetoothEnabled: false
      }))
    } else if (preset === 'offline') {
      wirelessDataStore.update((w) => ({
        ...w,
        airplaneMode: false,
        wifiConnected: false,
        bluetoothConnected: false
      }))
    }
  }
</script>

<div class="wireless-dev-panel">
  <div class="dev-header-row">
    <span class="dev-panel-title">Wireless & Connectivity</span>
    <div class="header-actions">
      {#if isSimulating}
        <span class="simulating-badge">Simulating</span>
      {/if}
      <button
        class="live-btn"
        onclick={resetToLive}
        use:islandTooltip={'Switch to live hardware state'}
      >
        <Radio size={11} />
        <span>Live Hardware</span>
      </button>
      <button
        class="reset-btn"
        onclick={resetToDefaults}
        use:islandTooltip={'Reset to default wireless status'}
      >
        <RefreshCw size={11} />
        <span>Defaults</span>
      </button>
    </div>
  </div>

  <!-- Main Presets -->
  <div class="presets-row">
    <button class="preset-chip" onclick={() => setPreset('wifi_only')}>
      <Wifi size={11} style="color: var(--vinta-success, #34c759)" />
      <span>Wi-Fi Only (No BT)</span>
    </button>

    <button class="preset-chip" onclick={() => setPreset('wifi_bt')}>
      <Bluetooth size={11} style="color: var(--vinta-primary, #6366f1)" />
      <span>Wi-Fi + Bluetooth</span>
    </button>

    <button class="preset-chip" onclick={() => setPreset('airplane')}>
      <Plane size={11} style="color: var(--vinta-warning, #ffcc00)" />
      <span>Airplane Mode</span>
    </button>

    <button class="preset-chip" onclick={() => setPreset('offline')}>
      <WifiOff size={11} style="color: var(--vinta-danger, #ff3b30)" />
      <span>Disconnected</span>
    </button>
  </div>

  <!-- Signal Strength Quick Presets -->
  <div class="signal-presets-row">
    <span class="sub-label">Signal Level:</span>
    <button class="signal-chip" onclick={() => setSignalPreset(100)}
      >Full (100%)</button
    >
    <button class="signal-chip" onclick={() => setSignalPreset(75)}
      >3 Bars (75%)</button
    >
    <button class="signal-chip" onclick={() => setSignalPreset(50)}
      >2 Bars (50%)</button
    >
    <button class="signal-chip" onclick={() => setSignalPreset(25)}
      >1 Bar (25%)</button
    >
    <button class="signal-chip" onclick={() => setSignalPreset(0)}>0%</button>
  </div>

  <!-- Interactive Controls Grid -->
  <div class="controls-grid">
    <!-- Wi-Fi Controls -->
    <div class="control-item toggle-row">
      <span class="control-label">Wi-Fi Hardware</span>
      <button
        class="dev-toggle-btn"
        class:active={wireless.hasWifi}
        onclick={() => {
          isSimulatingWirelessStore.set(true)
          wirelessDataStore.update((w) => ({ ...w, hasWifi: !w.hasWifi }))
        }}
      >
        {wireless.hasWifi ? 'Present' : 'None'}
      </button>
    </div>

    <div class="control-item toggle-row">
      <span class="control-label">Wi-Fi Connected</span>
      <button
        class="dev-toggle-btn"
        class:active={wireless.wifiConnected && !wireless.airplaneMode}
        onclick={() => {
          isSimulatingWirelessStore.set(true)
          wirelessDataStore.update((w) => ({
            ...w,
            wifiConnected: !w.wifiConnected,
            wifiEnabled: true,
            airplaneMode: false
          }))
        }}
      >
        {wireless.wifiConnected ? 'Connected' : 'Disconnected'}
      </button>
    </div>

    <!-- Wi-Fi Signal Strength Slider -->
    <div class="control-item slider-row full-width">
      <div class="control-label-row">
        <span class="control-label">Wi-Fi Signal Strength</span>
        <span class="control-value">{wireless.wifiSignal}%</span>
      </div>
      <input
        type="range"
        min="0"
        max="100"
        value={wireless.wifiSignal}
        oninput={(e) => {
          isSimulatingWirelessStore.set(true)
          const val = parseInt((e.target as HTMLInputElement).value, 10)
          wirelessDataStore.update((w) => ({ ...w, wifiSignal: val }))
        }}
        class="dev-slider"
      />
    </div>

    <!-- Bluetooth Controls -->
    <div class="control-item toggle-row">
      <span class="control-label">Bluetooth Hardware</span>
      <button
        class="dev-toggle-btn"
        class:active={wireless.hasBluetooth}
        onclick={() => {
          isSimulatingWirelessStore.set(true)
          wirelessDataStore.update((w) => ({
            ...w,
            hasBluetooth: !w.hasBluetooth
          }))
        }}
      >
        {wireless.hasBluetooth ? 'Present' : 'None (No Adapter)'}
      </button>
    </div>

    <div class="control-item toggle-row">
      <span class="control-label">Bluetooth Device</span>
      <button
        class="dev-toggle-btn"
        class:active={wireless.bluetoothConnected && !wireless.airplaneMode}
        onclick={() => {
          isSimulatingWirelessStore.set(true)
          wirelessDataStore.update((w) => ({
            ...w,
            hasBluetooth: true,
            bluetoothEnabled: true,
            bluetoothConnected: !w.bluetoothConnected,
            bluetoothDeviceName: !w.bluetoothConnected ? 'Wireless Device' : '',
            airplaneMode: false
          }))
        }}
      >
        {wireless.bluetoothConnected ? 'Connected' : 'Not Connected'}
      </button>
    </div>

    <!-- Airplane Mode Toggle -->
    <div class="control-item toggle-row full-width">
      <span class="control-label">Airplane Mode</span>
      <button
        class="dev-toggle-btn"
        class:active={wireless.airplaneMode}
        onclick={() => {
          isSimulatingWirelessStore.set(true)
          wirelessDataStore.update((w) => ({
            ...w,
            airplaneMode: !w.airplaneMode
          }))
        }}
      >
        <Plane size={11} />
        <span>{wireless.airplaneMode ? 'Active' : 'Off'}</span>
      </button>
    </div>
  </div>
</div>

<style>
  .wireless-dev-panel {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px 14px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--vinta-border);
    border-radius: 12px;
    margin-top: 6px;
    user-select: none;
  }

  .dev-header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
  }

  .dev-panel-title {
    font-size: 11px;
    font-weight: 700;
    color: var(--vinta-text-primary);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .simulating-badge {
    font-size: 9px;
    font-weight: 700;
    color: var(--vinta-warning, #ffcc00);
    background: rgba(255, 204, 0, 0.12);
    border: 1px solid rgba(255, 204, 0, 0.3);
    padding: 2px 6px;
    border-radius: 4px;
    text-transform: uppercase;
  }

  .live-btn,
  .reset-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 10px;
    font-weight: 600;
    color: var(--vinta-text-secondary);
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--vinta-border);
    padding: 3px 8px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .live-btn:hover,
  .reset-btn:hover {
    color: var(--vinta-text-primary);
    background: rgba(255, 255, 255, 0.12);
  }

  .presets-row {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .preset-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 10px;
    font-weight: 600;
    color: var(--vinta-text-primary);
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--vinta-border);
    padding: 4px 8px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .preset-chip:hover {
    background: rgba(255, 255, 255, 0.1);
    transform: translateY(-1px);
  }

  .signal-presets-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px;
    padding: 6px 8px;
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid var(--vinta-border);
    border-radius: 8px;
  }

  .sub-label {
    font-size: 10px;
    font-weight: 600;
    color: var(--vinta-text-secondary);
    margin-right: 2px;
  }

  .signal-chip {
    font-size: 9.5px;
    font-weight: 600;
    color: var(--vinta-text-primary);
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--vinta-border);
    padding: 2px 6px;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.12s ease;
  }

  .signal-chip:hover {
    background: rgba(255, 255, 255, 0.14);
    border-color: var(--vinta-primary, #6366f1);
  }

  .controls-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }

  .control-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px 10px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--vinta-border);
    border-radius: 8px;
    text-align: left;
  }

  .control-item.full-width {
    grid-column: span 2;
  }

  .control-label-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
  }

  .control-label {
    font-size: 10px;
    font-weight: 600;
    color: var(--vinta-text-secondary);
  }

  .control-value {
    font-size: 10px;
    font-weight: 700;
    color: var(--vinta-text-primary);
  }

  .dev-slider {
    width: 100%;
    accent-color: var(--vinta-primary, #6366f1);
    cursor: pointer;
  }

  .toggle-row {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }

  .dev-toggle-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 10px;
    font-weight: 600;
    color: var(--vinta-text-secondary);
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--vinta-border);
    padding: 3px 8px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .dev-toggle-btn.active {
    color: #ffffff;
    background: var(--vinta-primary, #6366f1);
    border-color: var(--vinta-primary, #6366f1);
  }
</style>
