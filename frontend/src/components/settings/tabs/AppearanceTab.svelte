<!--
  @component AppearanceTab
  @description Theme customization panel. Lets users select color presets (Dark, Light,
  Tokyo Night, Sunset, OLED), toggle acrylic backdrop blur, and adjust capsule opacity.
-->
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
    RefreshCw,
    Info,
    ChevronDown,
    Scale
  } from 'lucide-svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import {
    themeStore,
    userNameStore,
    glowIntensityStore,
    opacityMinimalStore,
    opacityExpandedStore,
    uiScaleStore,
    displayLayoutStore,
    tStore
  } from '../../../store'
  import {
    listThemes,
    importTheme,
    openThemesFolder,
    BUILTIN_THEMES
  } from '../../../themes/engine'
  import type { ThemeInfo } from '../../../themes/types'

  let availableThemes = $state<ThemeInfo[]>(BUILTIN_THEMES)
  let isLoadingThemes = $state(false)
  let isDetailsOpen = $state(false)

  function toggleDetails() {
    isDetailsOpen = !isDetailsOpen
  }

  const currentTheme = $derived(
    availableThemes.find((t) => t.manifest.id === $themeStore)
  )

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
          { label: 'Dark Gold', value: 'dark-gold' },
          { label: 'Nord Frost', value: 'nord-frost' },
          { label: 'Colorblind Accessible', value: 'colorblind-safe' }
        ]
  )

  const glowIntensityOptions = [
    { label: 'Disabled', value: 'disabled' },
    { label: 'Subtle', value: 'subtle' },
    { label: 'Vibrant Glow', value: 'vibrant' }
  ]

  const currentMonitor = $derived($displayLayoutStore.currentMonitor)
  const monitorHeight = $derived(currentMonitor?.height || 1080)
  const monitorWidth = $derived(currentMonitor?.width || 1920)

  const maxAllowedScale = $derived(
    monitorHeight <= 900 ? 1.1 : monitorHeight <= 1080 ? 1.2 : 1.25
  )

  const uiScaleOptions = $derived([
    { label: '100% (1.0x)', value: '1', isDefault: true },
    {
      label:
        monitorHeight <= 900 ? '110% (1.1x - Max for display)' : '110% (1.1x)',
      value: '1.1'
    },
    ...(maxAllowedScale >= 1.2
      ? [
          {
            label:
              monitorHeight <= 1080
                ? '120% (1.2x - Max for 1080p)'
                : '120% (1.2x)',
            value: '1.2'
          }
        ]
      : []),
    ...(maxAllowedScale >= 1.25
      ? [{ label: '125% (1.25x - Large)', value: '1.25' }]
      : [])
  ])

  let selectedScale = $state(String($uiScaleStore || 1.0))

  function handleScaleEffect() {
    const val = parseFloat(selectedScale)
    if (!isNaN(val) && val >= 1.0) {
      const clamped = Math.max(1.0, Math.min(maxAllowedScale, val))
      if ($uiScaleStore !== clamped) {
        $uiScaleStore = clamped
      }
    }
  }
  $effect(handleScaleEffect)
</script>

<div class="setting-group">
  <SettingItem
    label="Display Name"
    description="Personalizes the greeting on the Home feed"
    icon={User}
    flexCol={true}
    isDefault={$userNameStore === userNameStore.defaultValue}
    onReset={() => userNameStore.reset()}
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
    label={$tStore('appearance.theme_label')}
    description={$tStore('appearance.theme_desc')}
    icon={Moon}
    flexCol={true}
    isDefault={$themeStore === themeStore.defaultValue}
    onReset={() => themeStore.reset()}
  >
    <div class="theme-setting-content">
      <CustomSelect
        options={themeOptions}
        bind:value={$themeStore}
        defaultValue={themeStore.defaultValue}
        fullWidth={true}
        direction="auto"
      />
      <div class="theme-actions">
        <button
          type="button"
          class="theme-btn"
          onclick={handleImport}
        >
          <Upload size={13} strokeWidth={2.2} />
          <span>{$tStore('appearance.import_theme')}</span>
        </button>
        <button
          type="button"
          class="theme-btn"
          onclick={handleOpenFolder}
        >
          <FolderOpen size={13} strokeWidth={2.2} />
          <span>{$tStore('appearance.open_folder')}</span>
        </button>
        <button
          type="button"
          class="theme-btn"
          onclick={refreshThemes}
          disabled={isLoadingThemes}
        >
          <RefreshCw
            size={13}
            strokeWidth={2.2}
            class={isLoadingThemes ? 'spinning' : ''}
          />
          <span>{$tStore('appearance.refresh_themes')}</span>
        </button>
      </div>

      {#if currentTheme}
        <div class="theme-details-wrapper">
          <button
            type="button"
            class="theme-details-toggle"
            onclick={toggleDetails}
            aria-expanded={isDetailsOpen}
          >
            <div class="toggle-left">
              <Info size={12} strokeWidth={2} />
              <span>{$tStore('appearance.theme_details')}</span>
            </div>
            <ChevronDown
              size={12}
              strokeWidth={2}
              class="toggle-chevron {isDetailsOpen ? 'open' : ''}"
            />
          </button>

          {#if isDetailsOpen}
            <div class="theme-details-card">
              <div class="details-header">
                <span class="details-author">
                  {$tStore('appearance.created_by', {
                    author: currentTheme.manifest.author
                  })}
                </span>
                <div class="details-badges">
                  <span class="badge">v{currentTheme.manifest.version}</span>
                  <span class="badge mode"
                    >{currentTheme.manifest.mode || 'dark'}</span
                  >
                  {#if !currentTheme.isBuiltIn}
                    <span class="badge custom">Custom</span>
                  {/if}
                </div>
              </div>

              {#if currentTheme.manifest.description}
                <p class="details-desc">{currentTheme.manifest.description}</p>
              {/if}

              {#if currentTheme.manifest.accentColor}
                <div class="details-accent">
                  <span class="accent-label">Accent</span>
                  <span
                    class="accent-preview-pill"
                    style="background-color: {currentTheme.manifest
                      .accentColor}"
                  ></span>
                  <span class="accent-value"
                    >{currentTheme.manifest.accentColor}</span
                  >
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </SettingItem>

  <SettingItem
    label="Ambient Glow Intensity"
    description="Edge glow accents around the dynamic island"
    icon={Sparkles}
    flexCol={true}
    isDefault={$glowIntensityStore === glowIntensityStore.defaultValue}
    onReset={() => glowIntensityStore.reset()}
  >
    <CustomSelect
      options={glowIntensityOptions}
      bind:value={$glowIntensityStore}
      defaultValue={glowIntensityStore.defaultValue}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem
    label="UI Scale"
    description="Proportionally scales the island interface (Display: {monitorWidth}×{monitorHeight})"
    icon={Scale}
    flexCol={true}
    isDefault={$uiScaleStore === uiScaleStore.defaultValue}
    onReset={() => {
      selectedScale = '1'
      uiScaleStore.reset()
    }}
  >
    <CustomSelect
      options={uiScaleOptions}
      bind:value={selectedScale}
      defaultValue="1"
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem
    label="Compact Opacity"
    description="Background opacity when collapsed"
    icon={Eye}
    flexCol={true}
    isDefault={$opacityMinimalStore === opacityMinimalStore.defaultValue}
    onReset={() => opacityMinimalStore.reset()}
  >
    <SettingSlider
      min={0.2}
      max={1.0}
      step={0.05}
      defaultValue={opacityMinimalStore.defaultValue}
      bind:value={$opacityMinimalStore}
    />
  </SettingItem>

  <SettingItem
    label="Expanded Opacity"
    description="Background opacity when expanded"
    icon={Maximize2}
    flexCol={true}
    isDefault={$opacityExpandedStore === opacityExpandedStore.defaultValue}
    onReset={() => opacityExpandedStore.reset()}
  >
    <SettingSlider
      min={0.2}
      max={1.0}
      step={0.05}
      defaultValue={opacityExpandedStore.defaultValue}
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

  .theme-details-wrapper {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
    margin-top: 2px;
  }

  .theme-details-toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 4px 6px;
    border: none;
    background: transparent;
    color: var(--vinta-text-muted, rgba(255, 255, 255, 0.45));
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    border-radius: var(--vinta-radius-button, 6px);
    transition:
      color 0.15s ease,
      background 0.15s ease;
  }

  .theme-details-toggle:hover {
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.7));
    background: var(--vinta-hover, rgba(255, 255, 255, 0.04));
  }

  .toggle-left {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  :global(.toggle-chevron) {
    transition: transform 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  :global(.toggle-chevron.open) {
    transform: rotate(180deg);
  }

  .theme-details-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px 12px;
    border-radius: var(--vinta-radius-button, 8px);
    background: var(--vinta-item-bg, rgba(255, 255, 255, 0.04));
    border: 1px solid var(--vinta-border-light, rgba(255, 255, 255, 0.06));
    animation: fadeSlideDown 0.18s ease-out;
  }

  @keyframes fadeSlideDown {
    from {
      opacity: 0;
      transform: translateY(-4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .details-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    font-size: 11px;
  }

  .details-author {
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.7));
  }

  .author-name {
    color: var(--vinta-text-primary, #ffffff);
    font-weight: 600;
  }

  .details-badges {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .badge {
    font-size: 9.5px;
    font-weight: 600;
    padding: 1px 5px;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.08);
    color: var(--vinta-text-muted, rgba(255, 255, 255, 0.5));
    letter-spacing: 0.02em;
  }

  .badge.mode {
    text-transform: capitalize;
  }

  .badge.custom {
    background: var(--vinta-accent-muted, rgba(56, 189, 248, 0.15));
    color: var(--vinta-accent-primary, #38bdf8);
  }

  .details-desc {
    margin: 0;
    font-size: 11px;
    line-height: 1.45;
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.65));
  }

  .details-accent {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 10.5px;
    color: var(--vinta-text-muted, rgba(255, 255, 255, 0.45));
    margin-top: 2px;
  }

  .accent-preview-pill {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.2);
    flex-shrink: 0;
  }

  .accent-value {
    font-family: monospace;
    font-size: 10px;
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.6));
  }
</style>
