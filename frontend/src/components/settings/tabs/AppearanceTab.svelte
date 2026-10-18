<script lang="ts">
  import { onMount } from 'svelte'
  import {
    Moon,
    Sparkles,
    Eye,
    Maximize2,
    User,
    FolderOpen,
    Upload,
    RefreshCw
  } from 'lucide-svelte'
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
  import {
    listThemes,
    importTheme,
    openThemesFolder
  } from '../../../themes/engine'
  import type { ThemeInfo } from '../../../themes/types'

  let availableThemes = $state<ThemeInfo[]>([])
  let isLoadingThemes = $state(false)

  async function refreshThemes() {
    isLoadingThemes = true
    try {
      availableThemes = await listThemes()
    } finally {
      isLoadingThemes = false
    }
  }

  async function handleImport() {
    try {
      const imported = await importTheme()
      if (imported) {
        await refreshThemes()
        $themeStore = imported.manifest.id
      }
    } catch (err) {
      console.error('Failed to import theme:', err)
    }
  }

  async function handleOpenFolder() {
    try {
      await openThemesFolder()
    } catch (err) {
      console.error('Failed to open themes folder:', err)
    }
  }

  onMount(() => {
    refreshThemes()
  })

  const themeOptions = $derived(
    availableThemes.length > 0
      ? availableThemes.map((t) => ({
          label: t.manifest.name + (t.isBuiltIn ? '' : ' (Custom)'),
          value: t.manifest.id
        }))
      : [
          { label: 'OLED Dark', value: 'oled-dark' },
          { label: 'Nord Frost', value: 'nord-frost' }
        ]
  )

  const glowIntensityOptions = [
    { label: 'Disabled', value: 'disabled' },
    { label: 'Subtle', value: 'subtle' },
    { label: 'Vibrant Glow', value: 'vibrant' }
  ]
</script>

<div class="setting-group">
  <SettingItem
    label="Display Name"
    description="Personalizes the greeting on the Home feed (Default: User)"
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

  <SettingItem
    label="Theme"
    description="Color scheme for UI surfaces (Default: OLED Dark)"
    icon={Moon}
    flexCol={true}
  >
    <div class="theme-setting-content">
      <CustomSelect
        options={themeOptions}
        bind:value={$themeStore}
        fullWidth={true}
        direction="auto"
      />
      <div class="theme-actions">
        <button
          type="button"
          class="theme-btn"
          onclick={handleImport}
          title="Import .vintatheme package"
        >
          <Upload size={13} strokeWidth={2.2} />
          <span>Import Theme</span>
        </button>
        <button
          type="button"
          class="theme-btn"
          onclick={handleOpenFolder}
          title="Open Themes Folder in Windows Explorer"
        >
          <FolderOpen size={13} strokeWidth={2.2} />
          <span>Open Folder</span>
        </button>
        <button
          type="button"
          class="theme-btn"
          onclick={refreshThemes}
          title="Refresh Themes"
          disabled={isLoadingThemes}
        >
          <RefreshCw
            size={13}
            strokeWidth={2.2}
            class={isLoadingThemes ? 'spinning' : ''}
          />
          <span>Refresh</span>
        </button>
      </div>
    </div>
  </SettingItem>

  <SettingItem
    label="Ambient Glow Intensity"
    description="Edge glow accents around the dynamic island (Default: Subtle)"
    icon={Sparkles}
    flexCol={true}
  >
    <CustomSelect
      options={glowIntensityOptions}
      bind:value={$glowIntensityStore}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem
    label="Compact Opacity"
    description="Background opacity when collapsed (Default: 100%)"
    icon={Eye}
    flexCol={true}
  >
    <SettingSlider
      min={0.2}
      max={1.0}
      step={0.05}
      bind:value={$opacityMinimalStore}
    />
  </SettingItem>

  <SettingItem
    label="Expanded Opacity"
    description="Background opacity when expanded (Default: 100%)"
    icon={Maximize2}
    flexCol={true}
  >
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
    border-radius: var(--vinta-radius-button, 8px);
    background: var(--vinta-item-bg, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.08));
    color: var(--vinta-text-primary, #ffffff);
    font-size: 12px;
    font-weight: 500;
    outline: none;
    box-sizing: border-box;
    transition:
      border-color 0.2s ease,
      background 0.2s ease;
  }

  .text-input:focus {
    border-color: var(--vinta-accent-primary, #00ffaa);
    background: var(--vinta-hover, rgba(255, 255, 255, 0.08));
  }

  .text-input::placeholder {
    color: var(--vinta-text-muted, rgba(255, 255, 255, 0.4));
  }

  .theme-setting-content {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }

  .theme-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
  }

  .theme-btn {
    flex: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    height: 32px;
    padding: 0 8px;
    border-radius: var(--vinta-radius-button, 8px);
    background: var(--vinta-item-bg, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.08));
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.7));
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    box-sizing: border-box;
    white-space: nowrap;
    transition: all 0.15s ease;
  }

  .theme-btn:hover {
    background: var(--vinta-hover, rgba(255, 255, 255, 0.1));
    color: var(--vinta-text-primary, #ffffff);
    border-color: var(--vinta-border-accent, rgba(0, 255, 170, 0.3));
  }

  :global(.spinning) {
    animation: vinta-spin 0.8s linear infinite;
  }

  @keyframes vinta-spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
