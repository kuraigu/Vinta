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
    RotateCw,
    History,
    Download,
    Upload
  } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton, {
    type SettingButtonStatus
  } from '../SettingActionButton.svelte'
  import {
    autostartStore,
    hardwareAccelStore,
    tStore,
    notifData,
    syncSettingsToDisk,
    exportSettingsToFile,
    importSettingsFromFile,
    restoreLatestSettingsBackup
  } from '../../../store'
  import type { SettingsBackupInfo } from '../../../../bindings/Vinta/backend/models.js'
  import {
    GetAutostart,
    SetAutostart,
    GetBootHardwareAcceleration,
    RestartApp,
    GetCacheSize,
    ClearCache,
    ResetSettings,
    GetSettingsBackups
  } from '../../../../bindings/Vinta/backend/app.js'

  let bootHwAccel = $state<boolean | null>(null)
  let detectedRenderer = $state('')
  let needsHwAccelRestart = $derived(
    bootHwAccel !== null && $hardwareAccelStore !== bootHwAccel
  )
  let hardwareAccelDescription = $derived.by(() => {
    if (needsHwAccelRestart) {
      return $tStore('settings.hardware_accel_restart_desc')
    }
    if (bootHwAccel === false) {
      return $tStore('settings.hardware_accel_software_desc')
    }
    if (bootHwAccel === true) {
      return $tStore('settings.hardware_accel_active_desc', {
        renderer: detectedRenderer || 'GPU'
      })
    }
    return undefined
  })
  let isRestarting = $state(false)

  let latestBackup = $state<SettingsBackupInfo | null>(null)
  let restoreStatus = $state<SettingButtonStatus>('idle')
  let restoreStatusMsg = $state('')
  let restoreTimer: ReturnType<typeof setTimeout> | null = null

  /**
   * Queries WebGL debug renderer metadata to verify whether GPU or software rasterization is active.
   */
  function detectActiveRenderer(): string {
    try {
      const canvas = document.createElement('canvas')
      const gl =
        canvas.getContext('webgl2') ||
        canvas.getContext('webgl') ||
        canvas.getContext('experimental-webgl')
      if (
        !gl ||
        !(
          gl instanceof WebGLRenderingContext ||
          gl instanceof WebGL2RenderingContext
        )
      ) {
        return ''
      }
      const debugInfo = gl.getExtension('WEBGL_debug_renderer_info')
      const raw = debugInfo
        ? String(gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL) || '')
        : String(gl.getParameter(gl.RENDERER) || '')
      gl.getExtension('WEBGL_lose_context')?.loseContext()

      if (!raw) return ''
      const angleMatch = raw.match(/^ANGLE\s*\([^,]+,\s*([^,(]+)/i)
      const cleaned = (angleMatch ? angleMatch[1] : raw)
        .replace(/^ANGLE\s*\(/i, '')
        .replace(/\s+Direct3D.*$/i, '')
        .replace(/\)$/, '')
        .trim()
      return cleaned.length > 40 ? `${cleaned.slice(0, 37)}...` : cleaned
    } catch {
      return ''
    }
  }

  /**
   * Loads available automatic rotating settings snapshots from disk.
   */
  async function refreshBackups() {
    if (typeof GetSettingsBackups !== 'function') return
    try {
      const list = await GetSettingsBackups()
      latestBackup = Array.isArray(list) && list.length > 0 ? list[0] : null
    } catch {
      latestBackup = null
    }
  }

  onMount(async () => {
    detectedRenderer = detectActiveRenderer()

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

    if (typeof GetBootHardwareAcceleration === 'function') {
      try {
        const activeAtBoot = await GetBootHardwareAcceleration()
        if (typeof activeAtBoot === 'boolean') {
          bootHwAccel = activeAtBoot
        }
      } catch {
        // Ignore if backend binding is unavailable
      }
    }

    await refreshBackups()
  })

  /**
   * Flushes pending settings to disk and relaunches Vinta to apply boot-time graphics flags.
   */
  async function handleRestartApp() {
    isRestarting = true
    await syncSettingsToDisk(true)
    try {
      await RestartApp()
    } catch {
      isRestarting = false
    }
  }

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

  let cacheStatus = $state<SettingButtonStatus>('idle')
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
    'vinta_weather_animations',
    'vinta_home_widgets_visibility',
    'vinta_home_widgets_order',
    'vinta_home_widget_sizes',
    'vinta_home_widget_menu_enabled',
    'vinta_page_visibility',
    'vinta_dock_pinned_apps',
    'vinta_autostart',
    'vinta_hardware_accel',
    'autoCheckUpdates',
    'vinta_update_channel',
    'vinta_custom_update_repo',
    'vinta_ui_scale',
    'vinta_anim_scale',
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
    cacheStatus = 'loading'
    try {
      await ClearCache().catch(() => {})
      notifData.set([])

      for (let i = localStorage.length - 1; i >= 0; i--) {
        const key = localStorage.key(i)
        if (key && isCacheKey(key)) {
          localStorage.removeItem(key)
        }
      }
      cacheStatus = 'success'
      await calculateCacheSize()
      if (clearTimer) clearTimeout(clearTimer)
      clearTimer = setTimeout(() => {
        cacheStatus = 'idle'
      }, 2000)
    } catch {
      cacheStatus = 'error'
    }
  }

  let exportStatus = $state<SettingButtonStatus>('idle')
  let exportTimer: ReturnType<typeof setTimeout> | null = null

  let importStatus = $state<SettingButtonStatus>('idle')
  let importStatusMsg = $state('')
  let importTimer: ReturnType<typeof setTimeout> | null = null

  async function handleExportSettings() {
    exportStatus = 'loading'
    try {
      const res = await exportSettingsToFile()
      if (res.success && res.path) {
        exportStatus = 'success'
        if (exportTimer) clearTimeout(exportTimer)
        exportTimer = setTimeout(() => {
          exportStatus = 'idle'
        }, 2500)
      } else {
        exportStatus = 'idle'
      }
    } catch {
      exportStatus = 'error'
    }
  }

  async function handleImportSettings() {
    importStatus = 'loading'
    try {
      const res = await importSettingsFromFile()
      if (res.success) {
        importStatus = 'success'
        importStatusMsg = `Imported ${res.count ?? 0} settings successfully`
        await refreshBackups()
        if (importTimer) clearTimeout(importTimer)
        importTimer = setTimeout(() => {
          importStatus = 'idle'
          importStatusMsg = ''
        }, 3000)
      } else if (res.error) {
        importStatus = 'error'
        importStatusMsg = `Import failed: ${res.error}`
        if (importTimer) clearTimeout(importTimer)
        importTimer = setTimeout(() => {
          importStatus = 'idle'
          importStatusMsg = ''
        }, 4000)
      } else {
        importStatus = 'idle'
      }
    } catch {
      importStatus = 'error'
    }
  }

  /**
   * Restores settings from the most recent automatic rotating backup snapshot.
   */
  async function handleRestoreBackup() {
    restoreStatus = 'loading'
    try {
      const res = await restoreLatestSettingsBackup()
      if (res.success) {
        restoreStatus = 'success'
        restoreStatusMsg = ''
        await refreshBackups()
        if (restoreTimer) clearTimeout(restoreTimer)
        restoreTimer = setTimeout(() => {
          restoreStatus = 'idle'
        }, 3000)
      } else {
        restoreStatus = 'error'
        restoreStatusMsg = res.error || ''
        if (restoreTimer) clearTimeout(restoreTimer)
        restoreTimer = setTimeout(() => {
          restoreStatus = 'idle'
          restoreStatusMsg = ''
        }, 4000)
      }
    } catch {
      restoreStatus = 'error'
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
    <SettingItem
      label={$tStore('settings.startup_label')}
      icon={Power}
      isDefault={$autostartStore === autostartStore.defaultValue}
      onReset={() => void handleAutostartChange(autostartStore.defaultValue)}
    >
      <ToggleSwitch
        checked={$autostartStore}
        onchange={(checked) => void handleAutostartChange(checked)}
        ariaLabel={$tStore('settings.startup_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.hardware_accel_label')}
      description={hardwareAccelDescription}
      icon={Cpu}
      isDefault={$hardwareAccelStore === hardwareAccelStore.defaultValue}
      onReset={() => hardwareAccelStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$hardwareAccelStore}
        ariaLabel={$tStore('settings.hardware_accel_label')}
      />
    </SettingItem>

    {#if needsHwAccelRestart}
      <div class="hw-restart-row">
        <SettingActionButton
          variant="default"
          disabled={isRestarting}
          fullWidth={true}
          onclick={handleRestartApp}
        >
          <RotateCw size={13} />
          {$tStore('settings.hardware_accel_restart_btn')}
        </SettingActionButton>
      </div>
    {/if}
  </BentoCard>

  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.cache_label')}
      description={$tStore('settings.cache_desc', { size: cacheSizeStr })}
      icon={Database}
      flexCol={true}
    >
      <SettingActionButton
        variant="default"
        status={cacheStatus}
        successLabel={$tStore('settings.cache_cleared')}
        fullWidth={true}
        onclick={handleClearCache}
      >
        {$tStore('settings.cache_clear_btn')}
      </SettingActionButton>
    </SettingItem>

    <SettingItem
      label={$tStore('settings.export_label')}
      description={$tStore('settings.export_desc')}
      icon={Download}
      flexCol={true}
    >
      <SettingActionButton
        variant="default"
        status={exportStatus}
        successLabel={$tStore('settings.export_success')}
        fullWidth={true}
        onclick={handleExportSettings}
      >
        <Download size={13} />
        {$tStore('settings.export_btn')}
      </SettingActionButton>
    </SettingItem>

    <SettingItem
      label={$tStore('settings.import_label')}
      description={importStatusMsg || $tStore('settings.import_desc')}
      icon={Upload}
      flexCol={true}
    >
      <SettingActionButton
        variant="default"
        status={importStatus}
        successLabel={$tStore('settings.import_success')}
        fullWidth={true}
        onclick={handleImportSettings}
      >
        <Upload size={13} />
        {$tStore('settings.import_btn')}
      </SettingActionButton>
    </SettingItem>

    <SettingItem
      label={$tStore('settings.backup_restore_label')}
      description={restoreStatusMsg ||
        (latestBackup
          ? $tStore('settings.backup_restore_desc', {
              date: latestBackup.createdAt
            })
          : $tStore('settings.backup_restore_empty_desc'))}
      icon={History}
      flexCol={true}
    >
      <SettingActionButton
        variant="default"
        status={restoreStatus}
        successLabel={$tStore('settings.backup_restore_success')}
        disabled={!latestBackup}
        fullWidth={true}
        onclick={handleRestoreBackup}
      >
        <History size={13} />
        {$tStore('settings.backup_restore_btn')}
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
