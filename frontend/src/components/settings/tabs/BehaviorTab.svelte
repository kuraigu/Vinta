<script lang="ts">
  import {
    Layout,
    EyeOff,
    Timer,
    MousePointer,
    RotateCcw,
    Monitor
  } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import type { MonitorInfo } from '../../../../bindings/Vinta/backend/models.js'
  import {
    homeAutoHideStore,
    mediaAutoHideStore,
    autoHideDelayStore,
    autoHideDisplayStore,
    wakeMethodStore,
    hiddenPillOffsetXStore,
    allowDragPositionStore,
    displayLayoutStore
  } from '../../../store'

  const widgetContentOptions = [
    { label: 'Compact Dashboard', value: 'dash' },
    { label: 'Compact Clock', value: 'clock' }
  ]

  let isResettingOffset = $state(false)

  /**
   * Switches the active display and centers the island on the selected monitor.
   */
  function handleSwitchMonitor(mon: MonitorInfo) {
    const baseCenter =
      $displayLayoutStore.virtualLeft + $displayLayoutStore.primaryCenter
    const monCenter = mon.left + mon.width / 2
    const targetOffset = Math.round(monCenter - baseCenter)
    $hiddenPillOffsetXStore = targetOffset
    $displayLayoutStore = {
      ...$displayLayoutStore,
      currentMonitor: mon
    }
  }

  /**
   * Resets the persistent horizontal position offset of the island back to monitor center.
   */
  function handleResetOffset() {
    $hiddenPillOffsetXStore = 0
    isResettingOffset = true
    setTimeout(() => {
      isResettingOffset = false
    }, 1200)
  }
</script>

<div class="setting-group">
  <SettingItem label="Auto-Hide Home Page" icon={EyeOff}>
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$homeAutoHideStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem label="Auto-Hide Media Widget" icon={EyeOff}>
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$mediaAutoHideStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem
    label="Auto-Hide Delay"
    description="Inactivity duration before collapsing (Default: 10s)"
    icon={Timer}
    flexCol={true}
  >
    <SettingSlider
      min={5}
      max={30}
      step={1}
      unit="s"
      bind:value={$autoHideDelayStore}
    />
  </SettingItem>

  <SettingItem
    label="Collapsed Pill Content"
    description="Widget display when collapsed (Default: Compact Dashboard)"
    icon={Layout}
    flexCol={true}
  >
    <CustomSelect
      options={widgetContentOptions}
      bind:value={$autoHideDisplayStore}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem
    label="Wake Trigger Method"
    description="Interaction used to expand the island (Default: Click)"
    icon={MousePointer}
    flexCol={true}
  >
    <div class="segment-control full-width">
      <button
        class:active={$wakeMethodStore === 'hover'}
        onclick={() => ($wakeMethodStore = 'hover')}>Hover</button
      >
      <button
        class:active={$wakeMethodStore === 'click'}
        onclick={() => ($wakeMethodStore = 'click')}>Click</button
      >
    </div>
  </SettingItem>

  <SettingItem label="Allow Dragging Position" icon={MousePointer}>
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$allowDragPositionStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  {#if $displayLayoutStore.monitors && $displayLayoutStore.monitors.length > 1}
    <SettingItem
      label="Active Display"
      description="Select which connected monitor Vinta is positioned on"
      icon={Monitor}
      flexCol={true}
    >
      <div class="segment-control full-width">
        {#each $displayLayoutStore.monitors as mon (mon.id)}
          <button
            class:active={$displayLayoutStore.currentMonitor?.id === mon.id}
            onclick={() => handleSwitchMonitor(mon)}
          >
            {mon.name || `Display ${mon.id}`}
          </button>
        {/each}
      </div>
    </SettingItem>
  {/if}

  <SettingItem
    label="Center Island Position"
    description="Resets the horizontal offset of the Dynamic Island to monitor center"
    icon={RotateCcw}
    flexCol={true}
  >
    <SettingActionButton
      variant="default"
      fullWidth={true}
      onclick={handleResetOffset}
    >
      <RotateCcw size={13} />
      <span>{isResettingOffset ? 'Centered' : 'Reset Offset to Center'}</span>
    </SettingActionButton>
  </SettingItem>
</div>

<style>
  .setting-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .toggle-switch {
    position: relative;
    display: inline-block;
    width: 34px;
    height: 18px;
  }

  .toggle-switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .slider {
    position: absolute;
    cursor: pointer;
    inset: 0;
    background-color: rgba(255, 255, 255, 0.2);
    transition: 0.3s;
    border-radius: 20px;
    border: 1px solid var(--vinta-border);
  }

  .slider:before {
    position: absolute;
    content: '';
    height: 14px;
    width: 14px;
    left: 2px;
    bottom: 1px;
    background-color: white;
    transition: 0.3s;
    border-radius: 50%;
  }

  input:checked + .slider {
    background-color: var(--vinta-accent-primary, var(--vinta-primary, #38bdf8));
  }

  input:checked + .slider:before {
    transform: translateX(14px);
  }

  .segment-control {
    display: flex;
    background: rgba(255, 255, 255, 0.06);
    padding: 2px;
    border-radius: 8px;
    border: 1px solid var(--vinta-border);
  }

  .segment-control.full-width {
    width: 100%;
    box-sizing: border-box;
    height: 32px;
    align-items: center;
  }

  .segment-control button {
    background: transparent;
    border: none;
    color: var(--vinta-text-secondary);
    padding: 4px 10px;
    font-size: 11px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .segment-control.full-width button {
    flex: 1;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .segment-control button.active {
    background: rgba(255, 255, 255, 0.12);
    color: var(--vinta-text-primary);
    font-weight: 600;
  }
</style>
