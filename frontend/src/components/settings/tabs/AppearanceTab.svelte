<!--
  @component AppearanceTab
  @description Theme customization panel. Lets users select color presets (Dark, Light,
  Tokyo Night, Sunset, OLED), toggle acrylic backdrop blur, and adjust capsule opacity.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import {
    Moon,
    SunMedium,
    Eye,
    Maximize2,
    User,
    FolderOpen,
    Upload,
    RefreshCw,
    Info,
    ChevronDown,
    Scale,
    Gauge
  } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import SegmentedControl from '../../SegmentedControl.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import {
    themeStore,
    userNameStore,
    glowIntensityStore,
    opacityMinimalStore,
    opacityExpandedStore,
    uiScaleStore,
    animScaleStore,
    monitorUiScalesStore,
    getMonitorScaleKey,
    getMaxScaleForMonitor,
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
          label:
            t.manifest.name +
            (t.isBuiltIn ? '' : $tStore('appearance.custom_suffix')),
          value: t.manifest.id
        }))
      : [
          { label: 'Dark Gold', value: 'dark-gold' },
          { label: 'Nord Frost', value: 'nord-frost' },
          { label: 'Colorblind Accessible', value: 'colorblind-safe' }
        ]
  )

  const glowIntensityOptions = $derived([
    { label: $tStore('appearance.glow_disabled'), value: 'disabled' },
    { label: $tStore('appearance.glow_subtle'), value: 'subtle' },
    { label: $tStore('appearance.glow_vibrant'), value: 'vibrant' }
  ])

  const currentMonitor = $derived($displayLayoutStore.currentMonitor)
  const monitorHeight = $derived(currentMonitor?.height || 1080)
  const monitorWidth = $derived(currentMonitor?.width || 1920)
  const monitorDpi = $derived(
    currentMonitor?.dpiScale && currentMonitor.dpiScale > 0
      ? currentMonitor.dpiScale
      : 1.0
  )
  const monitorKey = $derived(getMonitorScaleKey(currentMonitor))

  const maxAllowedScale = $derived(
    getMaxScaleForMonitor(monitorWidth, monitorHeight, monitorDpi)
  )

  const uiScaleOptions = $derived([
    { label: '100% (1.0x)', value: '1', isDefault: true },
    {
      label:
        maxAllowedScale <= 1.1
          ? $tStore('appearance.scale_max', { pct: 110, scale: '1.1' })
          : '110% (1.1x)',
      value: '1.1'
    },
    ...(maxAllowedScale >= 1.2
      ? [
          {
            label:
              maxAllowedScale <= 1.2
                ? $tStore('appearance.scale_max', { pct: 120, scale: '1.2' })
                : '120% (1.2x)',
            value: '1.2'
          }
        ]
      : []),
    ...(maxAllowedScale >= 1.25
      ? [{ label: $tStore('appearance.scale_large'), value: '1.25' }]
      : []),
    ...(maxAllowedScale >= 1.5
      ? [
          {
            label:
              maxAllowedScale <= 1.5
                ? $tStore('appearance.scale_max', { pct: 150, scale: '1.5' })
                : '150% (1.5x)',
            value: '1.5'
          }
        ]
      : []),
    ...(maxAllowedScale >= 1.75
      ? [{ label: $tStore('appearance.scale_xl'), value: '1.75' }]
      : []),
    ...(maxAllowedScale >= 2.0
      ? [{ label: '200% (2.0x - 4K)', value: '2' }]
      : [])
  ])

  const activeMonitorScale = $derived.by(() => {
    const perMonitor = $monitorUiScalesStore?.[monitorKey]
    if (typeof perMonitor === 'number' && perMonitor >= 1.0) {
      return perMonitor
    }
    return currentMonitor?.isPrimary ? $uiScaleStore || 1.0 : 1.0
  })

  let selectedScale = $state('1')
  let lastSyncedMonitorKey = ''

  function syncScaleFromMonitor() {
    if (monitorKey !== lastSyncedMonitorKey) {
      lastSyncedMonitorKey = monitorKey
      selectedScale = String(activeMonitorScale)
    }
  }
  $effect(syncScaleFromMonitor)

  function handleScaleEffect() {
    const val = parseFloat(selectedScale)
    if (!isNaN(val) && val >= 1.0) {
      const clamped = Math.max(1.0, Math.min(maxAllowedScale, val))
      if (activeMonitorScale !== clamped) {
        monitorUiScalesStore.update((prev) => ({
          ...prev,
          [monitorKey]: clamped
        }))
        if (currentMonitor?.isPrimary && $uiScaleStore !== clamped) {
          $uiScaleStore = clamped
        }
      }
    }
  }
  $effect(handleScaleEffect)
</script>

<div class="setting-group">
  <BentoCard variant="group">
    <SettingItem
      label={$tStore('appearance.user_label')}
      description={$tStore('appearance.user_desc')}
      icon={User}
      flexCol={true}
      isDefault={$userNameStore === userNameStore.defaultValue}
      onReset={() => userNameStore.reset()}
    >
      <div class="input-wrapper">
        <input
          type="text"
          class="text-input"
          placeholder={$tStore('appearance.user_placeholder')}
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
          <button type="button" class="theme-btn" onclick={handleImport}>
            <Upload size={14} strokeWidth={2.2} />
            <span>{$tStore('appearance.import_theme')}</span>
          </button>
          <button type="button" class="theme-btn" onclick={handleOpenFolder}>
            <FolderOpen size={14} strokeWidth={2.2} />
            <span>{$tStore('appearance.open_folder')}</span>
          </button>
          <button
            type="button"
            class="theme-btn"
            onclick={refreshThemes}
            disabled={isLoadingThemes}
          >
            <RefreshCw
              size={14}
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
                      <span class="badge custom"
                        >{$tStore('appearance.custom_badge')}</span
                      >
                    {/if}
                  </div>
                </div>

                {#if currentTheme.manifest.description}
                  <p class="details-desc">
                    {currentTheme.manifest.description}
                  </p>
                {/if}

                {#if currentTheme.manifest.accentColor}
                  <div class="details-accent">
                    <span class="accent-label"
                      >{$tStore('appearance.accent_label')}</span
                    >
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
  </BentoCard>

  <BentoCard variant="group">
    <SettingItem
      label={$tStore('appearance.glow_label')}
      description={$tStore('appearance.glow_desc')}
      icon={SunMedium}
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
      label={$tStore('appearance.scale_label')}
      description={`${$tStore('appearance.scale_desc')} (${monitorWidth}×${monitorHeight} · ${Math.round(monitorDpi * 100)}%)`}
      icon={Scale}
      flexCol={true}
      isDefault={activeMonitorScale === 1.0}
      onReset={() => {
        selectedScale = '1'
        monitorUiScalesStore.update((prev) => {
          const next = { ...prev }
          delete next[monitorKey]
          return next
        })
        if (currentMonitor?.isPrimary) {
          uiScaleStore.reset()
        }
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
  </BentoCard>

  <BentoCard variant="group">
    <SettingItem
      label={$tStore('appearance.opacity_minimal_label')}
      description={$tStore('appearance.opacity_minimal_desc')}
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
      label={$tStore('appearance.opacity_expanded_label')}
      description={$tStore('appearance.opacity_expanded_desc')}
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

    <SettingItem
      label={$tStore('appearance.anim_scale_label')}
      description={$tStore('appearance.anim_scale_desc')}
      icon={Gauge}
      flexCol={true}
      isDefault={$animScaleStore === animScaleStore.defaultValue}
      onReset={() => animScaleStore.reset()}
    >
      <SegmentedControl
        fullWidth={true}
        options={[
          { value: '0', label: $tStore('appearance.anim_off') },
          { value: '0.7', label: $tStore('appearance.anim_snappy') },
          { value: '1', label: $tStore('appearance.anim_default') },
          { value: '1.3', label: $tStore('appearance.anim_smooth') },
          { value: '1.65', label: $tStore('appearance.anim_relaxed') }
        ]}
        value={String($animScaleStore)}
        onchange={(val) => {
          $animScaleStore = parseFloat(val)
        }}
      />
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/AppearanceTab.css"></style>
