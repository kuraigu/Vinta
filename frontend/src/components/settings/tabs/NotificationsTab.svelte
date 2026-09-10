<script lang="ts">
  import {
    Bell,
    Volume2,
    BellRing,
    Trash2,
    ClipboardCopy,
    ClipboardPaste
  } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import {
    dndStore,
    notifSoundStore,
    notifVolumeStore,
    notifAutoWakeStore,
    clipboardCopyNotifStore,
    clipboardPasteNotifStore,
    notifData
  } from '../../../store'

  function clearAllNotifications() {
    notifData.set([])
  }
</script>

<div class="setting-group">
  <SettingItem label="Do Not Disturb Mode" icon={Bell}>
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$dndStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem label="Play Sound on Notification" icon={BellRing}>
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$notifSoundStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem
    label={`Notification Sound Volume (${Math.round($notifVolumeStore * 100)}%)`}
    icon={Volume2}
    flexCol={true}
  >
    <input
      type="range"
      min="0"
      max="1"
      step="0.05"
      bind:value={$notifVolumeStore}
      class="range-slider no-swipe"
    />
  </SettingItem>

  <SettingItem label="Auto-Wake Island on Notification" icon={Bell}>
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$notifAutoWakeStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem
    label="Notify on Clipboard Copy"
    description="Display notification when text, files, or images are copied"
    icon={ClipboardCopy}
  >
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$clipboardCopyNotifStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem
    label="Notify on Clipboard Paste"
    description="Display notification when content is pasted via keyboard"
    icon={ClipboardPaste}
  >
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$clipboardPasteNotifStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem
    label="Clear Notification History"
    description={`${$notifData?.length || 0} active notifications`}
    icon={Trash2}
  >
    <SettingActionButton variant="danger" onclick={clearAllNotifications}>
      Clear History
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

  .range-slider {
    width: 100%;
    accent-color: var(--c-primary, #6366f1);
  }
</style>
