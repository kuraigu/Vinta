<script lang="ts">
  import { islandTooltip } from '../services/islandTooltip'
  import { wirelessDataStore } from '../store'

  let wireless = $derived($wirelessDataStore)

  let wifiSignalPercent = $derived(
    wireless.airplaneMode || !wireless.wifiEnabled || !wireless.wifiConnected
      ? 0
      : Math.max(0, Math.min(100, wireless.wifiSignal))
  )

  // Apex is at y=11.2, top arc reaches y=0.5 (Span is 10.7px)
  let clipY = $derived(11.2 - (10.7 * wifiSignalPercent) / 100)
  let clipH = $derived((10.7 * wifiSignalPercent) / 100)

  let wifiTitle = $derived(
    wireless.airplaneMode
      ? 'Airplane Mode: On'
      : wireless.wifiEnabled
        ? wireless.wifiConnected
          ? `Wi-Fi: ${wireless.wifiSsid || 'Connected'} (${wireless.wifiSignal}%)`
          : 'Wi-Fi: Disconnected'
        : 'Wi-Fi: Off'
  )

  let bluetoothTitle = $derived(
    wireless.bluetoothEnabled
      ? wireless.bluetoothConnected
        ? `Bluetooth: ${wireless.bluetoothDeviceName || 'Connected'}${wireless.bluetoothBattery !== undefined ? ` (${wireless.bluetoothBattery}%)` : ''}`
        : 'Bluetooth: Ready'
      : 'Bluetooth: Off'
  )
</script>

<div class="wireless-minimal-wrap">
  {#if wireless.airplaneMode}
    <!-- Solid Filled Airplane Glyph -->
    <span class="wireless-icon-btn" use:islandTooltip={wifiTitle}>
      <svg
        width="13"
        height="13"
        viewBox="0 0 16 16"
        fill="currentColor"
        class="solid-airplane-svg"
      >
        <path
          d="M7.2 1.4C7.5.8 8.5.8 8.8 1.4L9.5 5.5L15 8.2C15.6 8.5 15.6 9.4 15 9.7L9.5 8.5L9.5 12.3L11.2 13.8C11.5 14.1 11.3 14.7 10.8 14.7L8 13.8L5.2 14.7C4.7 14.7 4.5 14.1 4.8 13.8L6.5 12.3L6.5 8.5L1 9.7C.4 9.4.4 8.5 1 8.2L6.5 5.5L7.2 1.4Z"
        />
      </svg>
    </span>
  {:else}
    <!-- Bluetooth on the Left -->
    {#if wireless.hasBluetooth}
      <span class="wireless-icon-btn" use:islandTooltip={bluetoothTitle}>
        <svg
          width="13"
          height="13"
          viewBox="0 0 16 16"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          class="fluent-bluetooth-svg"
          class:off={!wireless.bluetoothEnabled || !wireless.bluetoothConnected}
        >
          <path
            d="M8.5 1v6l3.2-3.2c.4-.4 1-.4 1.4 0 .4.4.4 1 0 1.4L9.6 8.5l3.5 3.5c.4.4.4 1 0 1.4-.4.4-1 .4-1.4 0L8.5 10.2v4.8c0 .6-.4 1-1 1s-1-.4-1-1v-5L3.7 12.8c-.4.4-1 .4-1.4 0-.4-.4-.4-1 0-1.4l4.2-4.2-4.2-4.2c-.4-.4-.4-1 0-1.4.4-.4 1-.4 1.4 0L6.5 4.4V1c0-.6.4-1 1-1s1 .4 1 1zm-1 6.8l2-2-2-2v4zm0 2.4v4l2-2-2-2z"
            fill="#ffffff"
          />
        </svg>
      </span>
    {/if}

    <!-- Wi-Fi on the Right of Bluetooth -->
    {#if wireless.hasWifi}
      <span class="wireless-icon-btn" use:islandTooltip={wifiTitle}>
        <svg
          width="14.5"
          height="11"
          viewBox="0 0 16 12"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          class="solid-wifi-svg"
        >
          <defs>
            <clipPath id="vinta-wifi-fill-clip">
              <rect x="0" y={clipY} width="16" height={clipH} />
            </clipPath>
          </defs>

          <!-- Solid Background Fan Shape -->
          <path
            d="M8 11.2C7.5 11.2 7.1 10.9 6.8 10.5L0.6 3.8C0.2 3.3 0.3 2.6 0.8 2.2C4.9 -0.7 11.1 -0.7 15.2 2.2C15.7 2.6 15.8 3.3 15.4 3.8L9.2 10.5C8.9 10.9 8.5 11.2 8 11.2Z"
            fill="rgba(255, 255, 255, 0.22)"
          />

          <!-- Solid White Active Fill (Proportional to Signal Level) -->
          <path
            d="M8 11.2C7.5 11.2 7.1 10.9 6.8 10.5L0.6 3.8C0.2 3.3 0.3 2.6 0.8 2.2C4.9 -0.7 11.1 -0.7 15.2 2.2C15.7 2.6 15.8 3.3 15.4 3.8L9.2 10.5C8.9 10.9 8.5 11.2 8 11.2Z"
            fill="#ffffff"
            clip-path="url(#vinta-wifi-fill-clip)"
          />
        </svg>
      </span>
    {/if}
  {/if}
</div>

<style>
  .wireless-minimal-wrap {
    display: inline-flex;
    align-items: center;
    gap: 2.5px;
    background: transparent;
    user-select: none;
    flex-shrink: 0;
  }

  .wireless-icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: default;
    background: transparent;
    border: none;
    padding: 0;
    line-height: 1;
  }

  .solid-wifi-svg {
    display: block;
    flex-shrink: 0;
    transition: all 0.2s ease;
  }

  .solid-airplane-svg {
    color: #ffffff;
    display: block;
    flex-shrink: 0;
  }

  .fluent-bluetooth-svg {
    display: block;
    flex-shrink: 0;
    transition: opacity 0.2s ease;
  }

  .fluent-bluetooth-svg.off {
    opacity: 0.25;
  }
</style>
