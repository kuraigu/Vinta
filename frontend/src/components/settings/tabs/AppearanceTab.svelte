<script lang="ts">
  import { Moon, Sparkles, Eye, Maximize2 } from 'lucide-svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import {
    themeStore,
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
</style>
