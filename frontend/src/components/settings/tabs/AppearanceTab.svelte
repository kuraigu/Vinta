<script lang="ts">
  import { Moon, Maximize2, Sparkles, Eye } from '@lucide/svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import {
    themeStore,
    shapeStore,
    glowIntensityStore,
    opacityMinimalStore,
    opacityExpandedStore
  } from '../../../store'

  const themeOptions = [{ label: 'OLED Dark', value: 'dark' }]

  const glowIntensityOptions = [
    { label: 'Disabled', value: 'disabled' },
    { label: 'Subtle', value: 'subtle' },
    { label: 'Vibrant Glow', value: 'vibrant' }
  ]
</script>

<div class="setting-group">
  <SettingItem label="Theme" icon={Moon} flexCol={true}>
    <CustomSelect
      options={themeOptions}
      bind:value={$themeStore}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem label="Shape" icon={Maximize2}>
    <div class="segment-control">
      <button
        class:active={$shapeStore === 'pill'}
        onclick={() => ($shapeStore = 'pill')}>Pill</button
      >
      <button
        class:active={$shapeStore === 'box'}
        onclick={() => ($shapeStore = 'box')}>Box</button
      >
    </div>
  </SettingItem>

  <SettingItem label="Ambient Glow Intensity" icon={Sparkles} flexCol={true}>
    <CustomSelect
      options={glowIntensityOptions}
      bind:value={$glowIntensityStore}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem label="Compact Opacity" icon={Eye} flexCol={true}>
    <SettingSlider
      min={0.2}
      max={1.0}
      step={0.05}
      bind:value={$opacityMinimalStore}
    />
  </SettingItem>

  <SettingItem label="Expanded Opacity" icon={Maximize2} flexCol={true}>
    <SettingSlider
      min={0.2}
      max={1.0}
      step={0.05}
      bind:value={$opacityExpandedStore}
    />
  </SettingItem>
</div>

<style>
  .setting-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
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
</style>
