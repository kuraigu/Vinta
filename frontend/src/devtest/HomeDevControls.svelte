<script lang="ts">
  import { CloudSun, BatteryCharging, Radio } from 'lucide-svelte'
  import WeatherDevControls from './WeatherDevControls.svelte'
  import BatteryDevControls from './BatteryDevControls.svelte'
  import WirelessDevControls from './WirelessDevControls.svelte'

  let activeTab = $state<'battery' | 'wireless' | 'weather'>('battery')
</script>

<div class="home-dev-container">
  <div class="subtab-bar">
    <button
      class="subtab-btn"
      class:active={activeTab === 'battery'}
      onclick={() => (activeTab = 'battery')}
    >
      <BatteryCharging size={12} />
      <span>Battery</span>
    </button>
    <button
      class="subtab-btn"
      class:active={activeTab === 'wireless'}
      onclick={() => (activeTab = 'wireless')}
    >
      <Radio size={12} />
      <span>Wireless & Status</span>
    </button>
    <button
      class="subtab-btn"
      class:active={activeTab === 'weather'}
      onclick={() => (activeTab = 'weather')}
    >
      <CloudSun size={12} />
      <span>Weather</span>
    </button>
  </div>

  <div class="subtab-content">
    {#if activeTab === 'battery'}
      <BatteryDevControls />
    {:else if activeTab === 'wireless'}
      <WirelessDevControls />
    {:else}
      <WeatherDevControls />
    {/if}
  </div>
</div>

<style>
  .home-dev-container {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }

  .subtab-bar {
    display: flex;
    gap: 4px;
    background: rgba(255, 255, 255, 0.04);
    padding: 3px;
    border-radius: 8px;
    border: 1px solid var(--c-border);
  }

  .subtab-btn {
    flex: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    padding: 5px 6px;
    font-size: 10.5px;
    font-weight: 600;
    color: var(--c-text-secondary);
    background: transparent;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .subtab-btn:hover {
    color: var(--c-text-primary);
  }

  .subtab-btn.active {
    color: var(--c-text-primary);
    background: rgba(255, 255, 255, 0.1);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }

  .subtab-content {
    width: 100%;
  }
</style>
