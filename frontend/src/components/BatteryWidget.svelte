<script lang="ts">
  import { Zap, Plug, Leaf } from 'lucide-svelte'
  import { batteryDataStore } from '../store'

  let battery = $derived($batteryDataStore)

  function getBatteryFillColor(
    percentage: number,
    isCharging: boolean
  ): string {
    if (!isCharging && percentage <= 15) return '#ff3b30'
    return '#ffffff'
  }

  let batteryFillColor = $derived(
    getBatteryFillColor(battery.percentage, battery.isCharging)
  )

  let fillPercentage = $derived(
    battery.hasBattery ? Math.max(0, Math.min(100, battery.percentage)) : 0
  )

  function formatRemainingTime(seconds: number): string {
    if (seconds <= 0 || seconds >= 86400) return ''
    const hrs = Math.floor(seconds / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    if (hrs > 0) {
      return `${hrs}h ${mins}m remaining`
    }
    return `${mins}m remaining`
  }

  let remainingTimeStr = $derived(formatRemainingTime(battery.lifeTimeSeconds))

  function getStateLabel(b: typeof battery): string {
    if (!b.hasBattery && b.state === 'desktop') return 'Desktop AC Power'
    if (b.state === 'full') return 'Fully Charged'
    if (b.isCharging) return 'Charging'
    if (b.percentage <= 15) return 'Low Battery'
    if (b.batterySaver) return 'Battery Saver Active'
    return 'On Battery'
  }

  let stateLabel = $derived(getStateLabel(battery))
</script>

<!-- Minimal View on Base Island: Battery Capsule Pill (Fill State + Status Icon) -->
<div
  class="battery-capsule-minimal"
  class:charging={battery.isCharging}
  class:saver={!battery.isCharging && battery.batterySaver}
  class:low={!battery.isCharging && battery.percentage <= 15}
  title="{stateLabel} - {battery.percentage}%{remainingTimeStr
    ? ` (${remainingTimeStr})`
    : ''}"
>
  <div
    class="battery-pill-fill"
    style="width: {fillPercentage}%; background: {batteryFillColor}; border-radius: {fillPercentage >=
    95
      ? '9999px'
      : '9999px 0 0 9999px'};"
  ></div>

  <div class="battery-pill-content">
    {#if !battery.hasBattery}
      <Plug size={9} strokeWidth={2.4} class="battery-plug-icon" />
    {:else}
      {#if battery.isCharging}
        <Zap size={7.5} strokeWidth={2.8} class="battery-status-icon" />
      {:else if battery.batterySaver}
        <Leaf size={7.5} strokeWidth={2.8} class="battery-status-icon" />
      {/if}
      <span class="battery-percent-num">{battery.percentage}</span>
    {/if}
  </div>
</div>

<style>
  .battery-capsule-minimal {
    position: relative;
    height: 16px;
    width: 34px;
    min-width: 34px;
    max-width: 34px;
    border-radius: 9999px;
    background: rgba(255, 255, 255, 0.45);
    border: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    overflow: hidden;
    cursor: default;
    user-select: none;
    flex-shrink: 0;
  }

  .battery-capsule-minimal.low {
    background: rgba(255, 59, 48, 0.35);
  }

  .battery-pill-fill {
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    opacity: 1;
    transition:
      width 0.4s cubic-bezier(0.2, 0.8, 0.2, 1),
      background-color 0.3s ease;
  }

  .battery-pill-content {
    position: absolute;
    inset: 0;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1.5px;
    line-height: 1;
    pointer-events: none;
    transform: translateY(0.75px);
  }

  .battery-percent-num {
    font-size: 8.5px;
    font-weight: 800;
    letter-spacing: -0.2px;
    color: #0f172a;
    font-variant-numeric: tabular-nums;
    line-height: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0;
    padding: 0;
  }

  :global(.battery-status-icon) {
    color: #0f172a;
    flex-shrink: 0;
    display: block;
    margin-right: -0.5px;
  }

  :global(.battery-plug-icon) {
    color: #0f172a;
    flex-shrink: 0;
    display: block;
  }

  /* Low battery: crisp white text/icon over red fill & coral track */
  .battery-capsule-minimal.low .battery-percent-num {
    color: #ffffff;
  }

  .battery-capsule-minimal.low :global(.battery-status-icon) {
    color: #ffffff;
  }

  .battery-capsule-minimal.low :global(.battery-plug-icon) {
    color: #ffffff;
  }
</style>
