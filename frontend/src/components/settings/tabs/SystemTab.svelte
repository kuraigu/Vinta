<!--
  @component SystemTab
  @description System performance settings. Configures hardware telemetry polling rates,
  run-at-startup preferences, hardware acceleration, and factory reset options.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import {
    Power,
    Cpu,
    Database,
    RotateCcw,
    Check,
    Download,
    Upload
  } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import {
    autostartStore,
    hardwareAccelStore,
    tStore,
    notifData,
    exportSettingsToFile,
    importSettingsFromFile
  } from '../../../store'
  import {
    GetAutostart,
    SetAutostart,
    SetHardwareAcceleration,
    GetCacheSize,
    ClearCache,
    ResetSettings
  } from '../../../../bindings/Vinta/backend/app.js'

  onMount(async () => {
    if (typeof GetAutostart === 'function') {
      try {
        const osEnabled = await GetAutostart()
        if (typeof osEnabled === 'boolean' && osEnabled !== $autostartStore) {
          $autostartStore = osEnabled
        }
      } catch {
        // Ignore if backend binding is unavailable
      }
    }
  })

  /**
   * Updates Windows autostart registration and reverts the store if the registry write fails.
   *
   * @param nextVal - Target autostart state.
   */
  async function handleAutostartChange(nextVal: boolean) {
    $autostartStore = nextVal
    if (typeof SetAutostart === 'function') {
      try {
        await SetAutostart(nextVal)
      } catch {
        $autostartStore = !nextVal
      }
    }
  }

  function handleHardwareAccelEffect() {
    SetHardwareAcceleration($hardwareAccelStore).catch(() => {})
  }
  $effect(handleHardwareAccelEffect)

  let cacheCleared = $state(false)
  let clearTimer: ReturnType<typeof setTimeout> | null = null
  let cacheSizeStr = $state('0 B')

  const SETTINGS_KEYS = new Set([
    'vinta_theme',
    'vinta_user_name',
    'vinta_glow_intensity',
    'vinta_opacity_minimal',
    'vinta_opacity_expanded',
    'vinta_opacity_hidden',
    'vinta_autohide',
    'vinta_home_autohide',
    'vinta_media_autohide',
    'vinta_autohide_delay',
    'vinta_autohide_display',
    'vinta_hidden_pill_position',
    'vinta_wakemethod',
    'vinta_allow_drag_position',
    'vinta_autohide_borderless',
    'vinta_global_hotkey',
    'vinta_hotkey_quick_notes',
    'vinta_hotkey_mode_switch',
    'vinta_pinned_media_source',
    'vinta_dnd',
    'vinta_notif_autowake',
    'vinta_notif_satellite_enabled',
    'vinta_media_autowake',
    'vinta_notif_sound',
    'vinta_notif_volume',
    'vinta_time_format',
    'vinta_timezone',
    'vinta_date_format',
    'vinta_weather_location',
    'vinta_weather_unit',
    'vinta_home_widgets_visibility',
    'vinta_page_visibility',
    'vinta_dock_pinned_apps',
    'vinta_autostart',
    'vinta_hardware_accel',
    'autoCheckUpdates',
    'vinta_update_channel',
    'vinta_custom_update_repo',
    'vinta_ui_scale',
    'vinta_auto_discard_past_events',
    'vinta_locale'
  ])

  /**
   * Determines if a localStorage key belongs to transient application cache.
   *
   * @param key - Storage key name.
   * @returns True if the key represents a purgeable cache.
   */
  function isCacheKey(key: string): boolean {
    return (
      key.startsWith('vinta_weather_cache_') || key === 'vinta_sender_avatars'
    )
  }

  function formatBytes(bytes: number): string {
    if (bytes <= 0) return '0 B'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  async function calculateCacheSize() {
    try {
      let totalBytes = 0
      const notifs = $notifData || []
      if (notifs.length > 0) {
        totalBytes += new Blob([JSON.stringify(notifs)]).size
      }

      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i)
        if (key && isCacheKey(key)) {
          const val = localStorage.getItem(key) || ''
          totalBytes += (key.length + val.length) * 2
        }
      }

      try {
        const diskBytes = await GetCacheSize()
        if (typeof diskBytes === 'number' && diskBytes > 0) {
          totalBytes += diskBytes
        }
      } catch {
        // Fall back to local calculation if native call is unavailable
      }

      cacheSizeStr = formatBytes(totalBytes)
    } catch {
      cacheSizeStr = '0 B'
    }
  }

  function handleCalculateCacheSizeEffect() {
    void calculateCacheSize()
  }
  $effect(handleCalculateCacheSizeEffect)

  async function handleClearCache() {
    try {
      await ClearCache().catch(() => {})
      notifData.set([])

      for (let i = localStorage.length - 1; i >= 0; i--) {
        const key = localStorage.key(i)
        if (key && isCacheKey(key)) {
          localStorage.removeItem(key)
        }
      }
      cacheCleared = true
      await calculateCacheSize()
      if (clearTimer) clearTimeout(clearTimer)
      clearTimer = setTimeout(() => {
        cacheCleared = false
      }, 2000)
    } catch {
      // Ignore
    }
  }

  let isExporting = $state(false)
  let exportSuccess = $state(false)
  let exportTimer: ReturnType<typeof setTimeout> | null = null

  let isImporting = $state(false)
  let importStatusMsg = $state('')
  let importSuccess = $state(false)
  let importTimer: ReturnType<typeof setTimeout> | null = null

  async function handleExportSettings() {
    isExporting = true
    try {
      const res = await exportSettingsToFile()
      if (res.success && res.path) {
        exportSuccess = true
        if (exportTimer) clearTimeout(exportTimer)
        exportTimer = setTimeout(() => {
          exportSuccess = false
        }, 2500)
      }
    } finally {
      isExporting = false
    }
  }

  async function handleImportSettings() {
    isImporting = true
    try {
      const res = await importSettingsFromFile()
      if (res.success) {
        importSuccess = true
        importStatusMsg = `Imported ${res.count ?? 0} settings successfully`
        if (importTimer) clearTimeout(importTimer)
        importTimer = setTimeout(() => {
          importSuccess = false
          importStatusMsg = ''
        }, 3000)
      } else if (res.error) {
        importStatusMsg = `Import failed: ${res.error}`
        if (importTimer) clearTimeout(importTimer)
        importTimer = setTimeout(() => {
          importStatusMsg = ''
        }, 4000)
      }
    } finally {
      isImporting = false
    }
  }

  async function handleResetSettings() {
    try {
      if (typeof ResetSettings === 'function') {
        await ResetSettings().catch(() => {})
      }
      for (const key of SETTINGS_KEYS) {
        localStorage.removeItem(key)
      }
      window.location.reload()
    } catch {
      // Ignore
    }
  }
</script>

<div class="setting-group">
  <BentoCard variant="group">
    <SettingItem label={$tStore('settings.startup_label')} icon={Power}>
      <ToggleSwitch
        checked={$autostartStore}
        onchange={(checked) => void handleAutostartChange(checked)}
        ariaLabel={$tStore('settings.startup_label')}
      />
    </SettingItem>

    <SettingItem label={$tStore('settings.hardware_accel_label')} icon={Cpu}>
      <ToggleSwitch
        bind:checked={$hardwareAccelStore}
        ariaLabel={$tStore('settings.hardware_accel_label')}
      />
    </SettingItem>
  </BentoCard>

  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.cache_label')}
      description={$tStore('settings.cache_desc', { size: cacheSizeStr })}
      icon={Database}
      flexCol={true}
    >
      <SettingActionButton
        variant={cacheCleared ? 'success' : 'default'}
        fullWidth={true}
        onclick={handleClearCache}
      >
        {#if cacheCleared}
          <Check size={13} />
          {$tStore('settings.cache_cleared')}
        {:else}
          {$tStore('settings.cache_clear_btn')}
        {/if}
      </SettingActionButton>
    </SettingItem>

    <SettingItem
      label={$tStore('settings.export_label')}
      description={$tStore('settings.export_desc')}
      icon={Download}
      flexCol={true}
    >
      <SettingActionButton
        variant={exportSuccess ? 'success' : 'default'}
        fullWidth={true}
        disabled={isExporting}
        onclick={handleExportSettings}
      >
        {#if exportSuccess}
          <Check size={13} />
          {$tStore('settings.export_success')}
        {:else}
          <Download size={13} />
          {$tStore('settings.export_btn')}
        {/if}
      </SettingActionButton>
    </SettingItem>

    <SettingItem
      label={$tStore('settings.import_label')}
      description={importStatusMsg || $tStore('settings.import_desc')}
      icon={Upload}
      flexCol={true}
    >
      <SettingActionButton
        variant={importSuccess ? 'success' : 'default'}
        fullWidth={true}
        disabled={isImporting}
        onclick={handleImportSettings}
      >
        {#if importSuccess}
          <Check size={13} />
          {$tStore('settings.import_success')}
        {:else}
          <Upload size={13} />
          {$tStore('settings.import_btn')}
        {/if}
      </SettingActionButton>
    </SettingItem>
  </BentoCard>

  <SettingItem
    label={$tStore('settings.reset_label')}
    description={$tStore('settings.reset_desc')}
    icon={RotateCcw}
    flexCol={true}
  >
    <SettingActionButton
      variant="danger"
      fullWidth={true}
      onclick={handleResetSettings}
    >
      {$tStore('settings.reset_btn')}
    </SettingActionButton>
  </SettingItem>
</div>

<style src="../../styles/settings/SystemTab.css"></style>
