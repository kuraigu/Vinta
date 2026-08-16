<script lang="ts">
  import { Power, Cpu, Database, RotateCcw, Check } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import { autostartStore, hardwareAccelStore, notifData } from '../../../store'
  import {
    SetAutostart,
    SetHardwareAcceleration
  } from '../../../../wailsjs/go/backend/App.js'

  $effect(() => {
    SetAutostart($autostartStore).catch(() => {})
  })

  $effect(() => {
    SetHardwareAcceleration($hardwareAccelStore).catch(() => {})
  })

  let cacheCleared = $state(false)
  let clearTimer: ReturnType<typeof setTimeout> | null = null
  let cacheSizeStr = $state('0 B')

  function formatBytes(bytes: number): string {
    if (bytes <= 0) return '0 B'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  function calculateCacheSize() {
    try {
      let totalBytes = 0
      const notifs = $notifData || []
      if (notifs.length > 0) {
        totalBytes += new Blob([JSON.stringify(notifs)]).size
      }

      const keepKeys = new Set([
        'vinta_theme',
        'vinta_shape',
        'vinta_glow_intensity',
        'vinta_home_autohide',
        'vinta_media_autohide',
        'vinta_notif_autohide',
        'vinta_autohide_delay',
        'vinta_autohide_display',
        'vinta_wake_method',
        'vinta_dnd',
        'vinta_notif_autowake',
        'vinta_time_format',
        'vinta_timezone',
        'vinta_date_format',
        'vinta_weather_location',
        'vinta_weather_unit',
        'vinta_opacity_minimal',
        'vinta_opacity_expanded',
        'vinta_opacity_hidden',
        'vinta_hidden_pill_offset_x',
        'vinta_allow_drag_pos',
        'vinta_hw_accel',
        'vinta_notif_sound',
        'vinta_notif_volume',
        'vinta_autostart',
        'vinta_home_widgets_visibility',
        'vinta_page_visibility'
      ])

      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i)
        if (key && !keepKeys.has(key)) {
          const val = localStorage.getItem(key) || ''
          totalBytes += (key.length + val.length) * 2
        }
      }

      cacheSizeStr = formatBytes(totalBytes)
    } catch {
      cacheSizeStr = '0 B'
    }
  }

  $effect(() => {
    calculateCacheSize()
  })

  function handleClearCache() {
    try {
      const keepKeys = new Set([
        'vinta_theme',
        'vinta_shape',
        'vinta_glow_intensity',
        'vinta_home_autohide',
        'vinta_media_autohide',
        'vinta_notif_autohide',
        'vinta_autohide_delay',
        'vinta_autohide_display',
        'vinta_wake_method',
        'vinta_dnd',
        'vinta_notif_autowake',
        'vinta_time_format',
        'vinta_timezone',
        'vinta_date_format',
        'vinta_weather_location',
        'vinta_weather_unit',
        'vinta_opacity_minimal',
        'vinta_opacity_expanded',
        'vinta_opacity_hidden',
        'vinta_hidden_pill_offset_x',
        'vinta_allow_drag_pos',
        'vinta_hw_accel',
        'vinta_notif_sound',
        'vinta_notif_volume',
        'vinta_autostart',
        'vinta_home_widgets_visibility',
        'vinta_page_visibility'
      ])
      for (let i = localStorage.length - 1; i >= 0; i--) {
        const key = localStorage.key(i)
        if (key && !keepKeys.has(key)) {
          localStorage.removeItem(key)
        }
      }
      cacheCleared = true
      calculateCacheSize()
      if (clearTimer) clearTimeout(clearTimer)
      clearTimer = setTimeout(() => {
        cacheCleared = false
      }, 2000)
    } catch {
      // Ignore
    }
  }

  function handleResetSettings() {
    try {
      localStorage.clear()
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
  >
    <SettingActionButton
      variant={cacheCleared ? 'success' : 'default'}
      onclick={handleClearCache}
    >
      {#if cacheCleared}
        <Check size={12} /> Cleared
      {:else}
        Clear Cache
      {/if}
    </SettingActionButton>
  </SettingItem>

  <SettingItem
    label="Reset All Settings to Default"
    description="Restores all settings and layout states"
    icon={RotateCcw}
  >
    <SettingActionButton variant="danger" onclick={handleResetSettings}>
      Reset All
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
</style>
