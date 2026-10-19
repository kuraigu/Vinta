<script lang="ts">
  import {
    Power,
    Cpu,
    Database,
    RotateCcw,
    Check,
    Download,
    Upload
  } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import {
    autostartStore,
    hardwareAccelStore,
    notifData,
    exportSettingsToFile,
    importSettingsFromFile
  } from '../../../store'
  import {
    SetAutostart,
    SetHardwareAcceleration,
    GetCacheSize,
    ClearCache,
    ResetSettings
  } from '../../../../bindings/Vinta/backend/app.js'

  function handleAutostartEffect() {
    SetAutostart($autostartStore).catch(() => {})
  }
  $effect(handleAutostartEffect)

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
    'vinta_custom_update_repo'
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
  <SettingItem label="Run on Windows Startup" icon={Power}>
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$autostartStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem label="Hardware Acceleration" icon={Cpu}>
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$hardwareAccelStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem
    label="Clear Application Cache"
    description={`Calculated cache size: ${cacheSizeStr}`}
    icon={Database}
    flexCol={true}
  >
    <SettingActionButton
      variant={cacheCleared ? 'success' : 'default'}
      fullWidth={true}
      onclick={handleClearCache}
    >
      {#if cacheCleared}
        <Check size={13} /> Cleared
      {:else}
        Clear Cache
      {/if}
    </SettingActionButton>
  </SettingItem>

  <SettingItem
    label="Export Settings"
    description="Backup current configuration and preferences to a JSON file"
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
        <Check size={13} /> Exported Successfully
      {:else}
        <Download size={13} /> Export to JSON
      {/if}
    </SettingActionButton>
  </SettingItem>

  <SettingItem
    label="Import Settings"
    description={importStatusMsg ||
      'Restore previously exported settings with schema validation'}
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
        <Check size={13} /> Settings Imported
      {:else}
        <Upload size={13} /> Import from JSON
      {/if}
    </SettingActionButton>
  </SettingItem>

  <SettingItem
    label="Reset All Settings to Default"
    description="Restores all settings and layout states"
    icon={RotateCcw}
    flexCol={true}
  >
    <SettingActionButton
      variant="danger"
      fullWidth={true}
      onclick={handleResetSettings}
    >
      Reset All Settings
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
</style>
