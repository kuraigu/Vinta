<script lang="ts">
  import { Layout, EyeOff, Timer, MousePointer, RotateCcw } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import {
    homeAutoHideStore,
    mediaAutoHideStore,
    autoHideDelayStore,
    autoHideDisplayStore,
    wakeMethodStore,
    hiddenPillOffsetXStore,
    allowDragPositionStore
  } from '../../../store'

  const widgetContentOptions = [
    { label: 'Compact Dashboard', value: 'dash' },
    { label: 'Compact Clock', value: 'clock' }
  ]

  let isResettingOffset = $state(false)
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

  <SettingItem label="Auto-Hide Delay" icon={Timer} flexCol={true}>
    <SettingSlider
      min={5}
      max={30}
      step={1}
      unit="s"
      bind:value={$autoHideDelayStore}
    />
  </SettingItem>

  <SettingItem label="Collapsed Pill Content" icon={Layout} flexCol={true}>
    <CustomSelect
      options={widgetContentOptions}
      bind:value={$autoHideDisplayStore}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem label="Wake Trigger Method" icon={MousePointer}>
    <div class="segment-control">
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

  <SettingItem label="Center Island Position" icon={RotateCcw}>
    <button class="action-btn" onclick={handleResetOffset}>
      {isResettingOffset ? 'Centered' : 'Reset Offset'}
    </button>
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
    border: 1px solid var(--c-border);
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
    background-color: var(--c-primary, #6366f1);
  }

  input:checked + .slider:before {
    transform: translateX(14px);
  }

  .segment-control {
    display: flex;
    background: rgba(255, 255, 255, 0.06);
    padding: 2px;
    border-radius: 8px;
    border: 1px solid var(--c-border);
  }

  .segment-control button {
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    padding: 4px 10px;
    font-size: 11px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .segment-control button.active {
    background: rgba(255, 255, 255, 0.12);
    color: var(--c-text-primary);
    font-weight: 600;
  }

  .action-btn {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    padding: 4px 12px;
    font-size: 11.5px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .action-btn:hover {
    background: rgba(255, 255, 255, 0.14);
    border-color: var(--c-primary, #6366f1);
  }
</style>
