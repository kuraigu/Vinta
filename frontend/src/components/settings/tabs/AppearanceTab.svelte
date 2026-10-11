<script lang="ts">
  import { Moon, Sparkles, Eye, Maximize2, User } from 'lucide-svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import {
    themeStore,
    userNameStore,
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
  <SettingItem
    label="Display Name"
    description="Personalizes the greeting on the Home feed (defaults to User)"
    icon={User}
    flexCol={true}
  >
    <div class="input-wrapper">
      <input
        type="text"
        class="text-input"
        placeholder="User"
        bind:value={$userNameStore}
        maxlength="20"
        onblur={() => {
          if (!$userNameStore || !$userNameStore.trim()) {
            $userNameStore = 'User'
          } else {
            $userNameStore = $userNameStore.trim()
          }
        }}
      />
    </div>
  </SettingItem>

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

  .input-wrapper {
    width: 100%;
  }

  .text-input {
    width: 100%;
    padding: 8px 12px;
    border-radius: 8px;
    background: var(--c-input-bg, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    font-size: 12px;
    font-weight: 500;
    outline: none;
    box-sizing: border-box;
    transition:
      border-color 0.2s ease,
      background 0.2s ease;
  }

  .text-input:focus {
    border-color: var(--c-primary, #6366f1);
    background: var(--c-item-bg);
  }

  .text-input::placeholder {
    color: var(--c-text-muted);
  }
</style>
