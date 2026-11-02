<script lang="ts">
  import {
    Bell,
    Volume2,
    BellRing,
    Trash2,
    AlarmClock,
    Play,
    TimerReset,
    Orbit
  } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import {
    dndStore,
    notifSoundStore,
    notifVolumeStore,
    notifAutoWakeStore,
    notifSatelliteEnabledStore,
    alarmSoundPresetStore,
    alarmVolumeStore,
    alarmSnoozeMinutesStore,
    notifData
  } from '../../../store'
  import {
    previewAlarmSound,
    type AlarmSoundPreset
  } from '../../../services/audioService'

  const alarmSoundOptions: { value: AlarmSoundPreset; label: string }[] = [
    { value: 'chime', label: 'Chime (Soft Bell)' },
    { value: 'radar', label: 'Radar (Sonar Ping)' },
    { value: 'digital', label: 'Digital (Synth Beep)' },
    { value: 'marimba', label: 'Marimba (Warm Scale)' }
  ]

  const snoozeOptions = [
    { value: '5', label: '5 minutes' },
    { value: '10', label: '10 minutes' },
    { value: '15', label: '15 minutes' }
  ]

  let snoozeString = $state(String($alarmSnoozeMinutesStore))

  $effect(function syncSnoozeMinutes() {
    const parsed = parseInt(snoozeString, 10)
    if (!isNaN(parsed) && parsed !== $alarmSnoozeMinutesStore) {
      $alarmSnoozeMinutesStore = parsed
    }
  })

  function clearAllNotifications() {
    notifData.set([])
  }

  function handleAuditionAlarmSound() {
    previewAlarmSound($alarmSoundPresetStore, $alarmVolumeStore)
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
    label="Notification Sound Volume"
    description="Audio feedback level for incoming notifications (Default: 50%)"
    icon={Volume2}
    flexCol={true}
  >
    <SettingSlider
      min={0}
      max={100}
      step={5}
      unit="%"
      value={Math.round($notifVolumeStore * 100)}
      onchange={(val) => ($notifVolumeStore = val / 100)}
    />
  </SettingItem>

  <SettingItem label="Auto-Wake Island on Notification" icon={Bell}>
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$notifAutoWakeStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem
    label="Show Notification Satellite"
    description="Display alerts in a companion satellite capsule beside the island during media playback and other active views"
    icon={Orbit}
  >
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$notifSatelliteEnabledStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <!-- ======================================================================
       Alarms & Reminders Audio Settings
       ====================================================================== -->
  <div class="setting-section-header">
    <AlarmClock size={13} />
    <span>Alarms & Reminders</span>
  </div>

  <SettingItem
    label="Alarm Ringing Sound"
    description="Procedural synthesis chime played when scheduled alarms fire"
    icon={AlarmClock}
    flexCol={true}
  >
    <div class="sound-select-row">
      <div class="sound-select-box">
        <CustomSelect
          options={alarmSoundOptions}
          bind:value={$alarmSoundPresetStore}
          placeholder="Select sound..."
          fullWidth={true}
        />
      </div>
      <button
        type="button"
        class="sound-preview-btn"
        onclick={handleAuditionAlarmSound}
        aria-label="Preview alarm sound"
      >
        <Play size={11} />
        <span>Test</span>
      </button>
    </div>
  </SettingItem>

  <SettingItem
    label="Alarm Volume"
    description="Ringing volume level for active alarms and timer completions"
    icon={Volume2}
    flexCol={true}
  >
    <SettingSlider
      min={10}
      max={100}
      step={5}
      unit="%"
      value={Math.round($alarmVolumeStore * 100)}
      onchange={(val) => ($alarmVolumeStore = val / 100)}
    />
  </SettingItem>

  <SettingItem
    label="Default Snooze Duration"
    description="Interval duration before a snoozed alarm rings again"
    icon={TimerReset}
  >
    <div class="snooze-select-box">
      <CustomSelect
        options={snoozeOptions}
        bind:value={snoozeString}
        placeholder="Duration..."
      />
    </div>
  </SettingItem>

  <SettingItem
    label="Clear Notification History"
    description={`${$notifData?.length || 0} active notifications`}
    icon={Trash2}
    flexCol={true}
  >
    <SettingActionButton
      variant="danger"
      fullWidth={true}
      onclick={clearAllNotifications}
    >
      Clear Notification History
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
    background-color: var(
      --vinta-accent-primary,
      var(--vinta-primary, #38bdf8)
    );
  }

  input:checked + .slider:before {
    transform: translateX(14px);
  }

  .setting-section-header {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 10px 4px 2px 4px;
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--vinta-text-muted, #94a3b8);
  }

  .sound-select-row {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
  }

  .sound-select-box {
    flex: 1;
    min-width: 0;
  }

  .sound-preview-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    height: 32px;
    padding: 0 12px;
    border-radius: 9px;
    background: var(--vinta-item-bg, rgba(255, 255, 255, 0.08));
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.14));
    color: var(--vinta-text-primary, #ffffff);
    font-size: 11.5px;
    font-weight: 600;
    cursor: pointer;
    flex-shrink: 0;
    transition: all 0.15s ease;
  }

  .sound-preview-btn:hover {
    background: var(--vinta-hover, rgba(255, 255, 255, 0.16));
    border-color: rgba(255, 255, 255, 0.25);
    transform: translateY(-0.5px);
  }

  .snooze-select-box {
    width: 130px;
    flex-shrink: 0;
  }
</style>
