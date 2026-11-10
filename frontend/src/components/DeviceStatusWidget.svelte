<!--
  @component DeviceStatusWidget
  @description Combined status chip showing battery charge, audio volume, and network connectivity.
-->
<script lang="ts">
  import WirelessWidget from './WirelessWidget.svelte'
  import BatteryWidget from './BatteryWidget.svelte'
  import Clock from './Clock.svelte'
  import { homeWidgetsVisibilityStore, batteryDataStore } from '../store'

  let hasBattery = $derived($batteryDataStore.hasBattery)
  let showWireless = $derived($homeWidgetsVisibilityStore.wireless)
  let showBattery = $derived($homeWidgetsVisibilityStore.battery && hasBattery)
  let showClock = $derived($homeWidgetsVisibilityStore.clock)
</script>

{#if showWireless || showBattery || showClock}
  <div class="device-status-widget">
    <!-- Connectivity & Power Group -->
    {#if showWireless || showBattery}
      <div class="connectivity-power-group">
        {#if showWireless}
          <WirelessWidget />
        {/if}
        {#if showBattery}
          <BatteryWidget />
        {/if}
      </div>
    {/if}

    <!-- Clock Time Group -->
    {#if showClock}
      <div class="clock-group">
        <Clock expanded={false} />
      </div>
    {/if}
  </div>
{/if}

<style src="./styles/DeviceStatusWidget.css"></style>
