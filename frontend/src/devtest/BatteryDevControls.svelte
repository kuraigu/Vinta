<script lang="ts">
  import {
    Zap,
    Battery,
    BatteryWarning,
    Leaf,
    Plug,
    RefreshCw
  } from '@lucide/svelte'
  import {
    batteryDataStore,
    isSimulatingBatteryStore,
    type BatteryData
  } from '../store'
  import { GetBatteryStatus } from '../../wailsjs/go/backend/App.js'

  let simulatedPercentage = $state($batteryDataStore.percentage || 75)
  let simulatedCharging = $state(true)
  let simulatedBatterySaver = $state(false)
  let simulatedHasBattery = $state(true)
  let simulatedState = $state<
    'charging' | 'discharging' | 'full' | 'desktop' | 'unknown'
  >('charging')

  function applySimulation() {
    isSimulatingBatteryStore.set(true)
    batteryDataStore.set({
      hasBattery: simulatedHasBattery,
      isCharging: simulatedCharging,
      percentage: simulatedPercentage,
      batterySaver: simulatedBatterySaver,
      lifeTimeSeconds: simulatedCharging
        ? 3600
        : Math.round((simulatedPercentage / 100) * 14400),
      state: simulatedHasBattery ? simulatedState : 'desktop'
    })
  }

  async function resetToHardware() {
    isSimulatingBatteryStore.set(false)
    try {
      if (typeof GetBatteryStatus === 'function') {
        const live = await GetBatteryStatus()
        if (live) {
          batteryDataStore.set(live as BatteryData)
          simulatedPercentage = live.percentage
          simulatedCharging = live.isCharging
          simulatedBatterySaver = live.batterySaver
          simulatedHasBattery = live.hasBattery
          simulatedState =
            (live.state as
              'charging' | 'discharging' | 'full' | 'desktop' | 'unknown') ||
            'desktop'
        }
      }
    } catch {
      // Ignore
    }
  }

  function setPreset(
    pct: number,
    charging: boolean,
    saver: boolean,
    hasBat: boolean,
    state: 'charging' | 'discharging' | 'full' | 'desktop'
  ) {
    simulatedPercentage = pct
    simulatedCharging = charging
    simulatedBatterySaver = saver
    simulatedHasBattery = hasBat
    simulatedState = state
    applySimulation()
  }
</script>

<div class="battery-dev-panel">
  <div class="dev-header-row">
    <span class="dev-panel-title">Battery & Power Simulation</span>
    <button
      class="reset-btn"
      onclick={resetToHardware}
      title="Reset to actual hardware power status"
    >
      <RefreshCw size={11} />
      <span>Live Hardware</span>
    </button>
  </div>

  <!-- Presets -->
  <div class="presets-row">
    <button
      class="preset-chip"
      onclick={() => setPreset(85, true, false, true, 'charging')}
    >
      <Zap size={11} style="color: var(--c-success, #34c759)" />
      <span>Charging (85%)</span>
    </button>

    <button
      class="preset-chip"
      onclick={() => setPreset(50, false, false, true, 'discharging')}
    >
      <Battery size={11} style="color: var(--c-primary, #007aff)" />
      <span>Battery (50%)</span>
    </button>

    <button
      class="preset-chip"
      onclick={() => setPreset(15, false, true, true, 'discharging')}
    >
      <BatteryWarning size={11} style="color: var(--c-warning, #ffcc00)" />
      <span>Low (15%)</span>
    </button>

    <button
      class="preset-chip"
      onclick={() => setPreset(5, false, true, true, 'discharging')}
    >
      <BatteryWarning size={11} style="color: var(--c-danger, #ff3b30)" />
      <span>Critical (5%)</span>
    </button>

    <button
      class="preset-chip"
      onclick={() => setPreset(100, true, false, true, 'full')}
    >
      <Zap size={11} style="color: var(--c-success, #34c759)" />
      <span>Full (100%)</span>
    </button>

    <button
      class="preset-chip"
      onclick={() => setPreset(100, true, false, false, 'desktop')}
    >
      <Plug size={11} style="color: var(--c-text-secondary)" />
      <span>Desktop (No Bat)</span>
    </button>
  </div>

  <!-- Interactive Controls -->
  <div class="controls-grid">
    <div class="control-item slider-row">
      <div class="control-label-row">
        <span class="control-label">Battery Level</span>
        <span class="control-value">{simulatedPercentage}%</span>
      </div>
      <input
        type="range"
        min="0"
        max="100"
        bind:value={simulatedPercentage}
        oninput={applySimulation}
        class="dev-slider"
      />
    </div>

    <div class="control-item toggle-row">
      <span class="control-label">Hardware Battery</span>
      <button
        class="dev-toggle-btn"
        class:active={simulatedHasBattery}
        onclick={() => {
          simulatedHasBattery = !simulatedHasBattery
          applySimulation()
        }}
      >
        {simulatedHasBattery ? 'Present' : 'None (Desktop)'}
      </button>
    </div>

    <div class="control-item toggle-row">
      <span class="control-label">Charging State</span>
      <button
        class="dev-toggle-btn"
        class:active={simulatedCharging}
        onclick={() => {
          simulatedCharging = !simulatedCharging
          simulatedState = simulatedCharging ? 'charging' : 'discharging'
          applySimulation()
        }}
      >
        {simulatedCharging ? '⚡ Charging' : '🔋 Discharging'}
      </button>
    </div>

    <div class="control-item toggle-row">
      <span class="control-label">Battery Saver</span>
      <button
        class="dev-toggle-btn"
        class:active={simulatedBatterySaver}
        onclick={() => {
          simulatedBatterySaver = !simulatedBatterySaver
          applySimulation()
        }}
      >
        <Leaf size={11} />
        <span>{simulatedBatterySaver ? 'Active' : 'Off'}</span>
      </button>
    </div>
  </div>
</div>

<style>
  .battery-dev-panel {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px 14px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--c-border);
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
    color: var(--c-text-primary);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .reset-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 10px;
    font-weight: 600;
    color: var(--c-text-secondary);
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--c-border);
    padding: 3px 8px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .reset-btn:hover {
    color: var(--c-text-primary);
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
    color: var(--c-text-primary);
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--c-border);
    padding: 4px 8px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .preset-chip:hover {
    background: rgba(255, 255, 255, 0.1);
    transform: translateY(-1px);
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
    border: 1px solid var(--c-border);
    border-radius: 8px;
    text-align: left;
  }

  .control-item.slider-row {
    grid-column: span 2;
  }

  .control-label-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .control-label {
    font-size: 10px;
    font-weight: 600;
    color: var(--c-text-secondary);
  }

  .control-value {
    font-size: 11px;
    font-weight: 700;
    color: var(--c-primary, #007aff);
  }

  .dev-slider {
    width: 100%;
    accent-color: var(--c-primary, #007aff);
    cursor: pointer;
    margin: 2px 0;
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
    color: var(--c-text-secondary);
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--c-border);
    padding: 3px 8px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .dev-toggle-btn.active {
    color: #ffffff;
    background: var(--c-primary, #007aff);
    border-color: var(--c-primary, #007aff);
  }
</style>
