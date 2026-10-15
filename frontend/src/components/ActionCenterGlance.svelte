<script lang="ts">
  import { getContext } from 'svelte'
  import {
    SlidersHorizontal,
    Wifi,
    WifiOff,
    Battery,
    Zap,
    Leaf,
    ExternalLink
  } from 'lucide-svelte'
  import ActionCenterPage, {
    config as actionCenterConfig
  } from '../pages/ActionCenterPage.svelte'
  import {
    wirelessDataStore,
    batteryDataStore,
    type IslandConfig
  } from '../store'
  import {
    SetWifiEnabled,
    SetBatterySaver
  } from '../../bindings/Vinta/backend/app.js'
  import { islandTooltip } from '../services/islandTooltip'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let wireless = $derived($wirelessDataStore)
  let battery = $derived($batteryDataStore)

  function openActionCenter(e?: Event) {
    if (e) e.stopPropagation()
    if (navigate) {
      navigate(ActionCenterPage, actionCenterConfig, true)
    }
  }

  async function handleToggleWifi(e: Event) {
    e.stopPropagation()
    const next = !wireless.wifiEnabled
    wirelessDataStore.update((w) => ({
      ...w,
      wifiEnabled: next,
      wifiConnected: next ? w.wifiConnected : false
    }))
    if (typeof SetWifiEnabled === 'function') {
      try {
        await SetWifiEnabled(next)
      } catch {}
    }
  }

  async function handleToggleSaver(e: Event) {
    e.stopPropagation()
    const next = !battery.batterySaver
    batteryDataStore.update((b) => ({
      ...b,
      batterySaver: next
    }))
    if (typeof SetBatterySaver === 'function') {
      try {
        await SetBatterySaver(next)
      } catch {}
    }
  }

  function computeWifiSub(): string {
    if (wireless.airplaneMode) return 'Airplane Mode'
    if (!wireless.wifiEnabled) return 'Off'
    if (wireless.wifiConnected && wireless.wifiSsid) {
      return `${wireless.wifiSsid}${wireless.wifiSignal > 0 ? ` (${wireless.wifiSignal}%)` : ''}`
    }
    if (wireless.wifiConnected) return 'Connected'
    return 'Disconnected'
  }
  let wifiSub = $derived.by(computeWifiSub)

  function computeBatterySub(): string {
    if (!battery.hasBattery) return 'Desktop Power'
    if (battery.isCharging) return `Charging (${battery.percentage}%)`
    if (battery.batterySaver) return `Battery Saver (${battery.percentage}%)`
    return `${battery.percentage}% remaining`
  }
  let batterySub = $derived.by(computeBatterySub)
</script>

<div class="action-center-glance-card">
  <div class="glance-header">
    <div class="glance-title-group">
      <SlidersHorizontal size={12} />
      <span>Quick Details</span>
    </div>
  </div>

  <div class="glance-cards-list">
    <!-- Wi-Fi Bento Card -->
    <div class="glance-bento-card">
      <div class="glance-card-left">
        <div
          class="glance-icon-circle"
          class:is-off={!wireless.wifiEnabled || wireless.airplaneMode}
        >
          {#if wireless.wifiEnabled && !wireless.airplaneMode}
            <Wifi size={16} />
          {:else}
            <WifiOff size={16} />
          {/if}
        </div>
        <div class="glance-card-info">
          <span class="glance-primary-text">Wi-Fi</span>
          <span class="glance-secondary-text">{wifiSub}</span>
        </div>
      </div>

      <div class="glance-card-right">
        <button
          type="button"
          class="glance-open-near-toggle"
          onclick={openActionCenter}
          use:islandTooltip={'Open Action Center'}
          aria-label="Open Action Center"
        >
          <ExternalLink size={12} />
        </button>

        <label
          class="toggle-switch"
          class:disabled-switch={wireless.airplaneMode}
          use:islandTooltip={wireless.wifiEnabled
            ? 'Turn Wi-Fi off'
            : 'Turn Wi-Fi on'}
        >
          <input
            type="checkbox"
            checked={wireless.wifiEnabled && !wireless.airplaneMode}
            disabled={wireless.airplaneMode}
            onchange={handleToggleWifi}
          />
          <span class="slider"></span>
        </label>
      </div>
    </div>

    <!-- Battery / Saver Bento Card -->
    {#if battery.hasBattery}
      <div class="glance-bento-card">
        <div class="glance-card-left">
          <div class="glance-icon-circle" class:is-off={!battery.batterySaver}>
            {#if battery.isCharging}
              <Zap size={16} />
            {:else if battery.batterySaver}
              <Leaf size={16} />
            {:else}
              <Battery size={16} />
            {/if}
          </div>
          <div class="glance-card-info">
            <span class="glance-primary-text">Battery Saver</span>
            <span class="glance-secondary-text">{batterySub}</span>
          </div>
        </div>

        <div class="glance-card-right">
          <button
            type="button"
            class="glance-open-near-toggle"
            onclick={openActionCenter}
            use:islandTooltip={'Open Action Center'}
            aria-label="Open Action Center"
          >
            <ExternalLink size={12} />
          </button>

          <label
            class="toggle-switch"
            use:islandTooltip={battery.batterySaver
              ? 'Turn Battery Saver off'
              : 'Turn Battery Saver on'}
          >
            <input
              type="checkbox"
              checked={battery.batterySaver}
              onchange={handleToggleSaver}
            />
            <span class="slider"></span>
          </label>
        </div>
      </div>
    {/if}
  </div>
</div>

<style src="./styles/ActionCenterGlance.css"></style>
