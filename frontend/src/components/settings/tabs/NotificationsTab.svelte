<!--
  @component NotificationsTab
  @description Toast notification preferences. Toggles the floating satellite notification
  pill, configures auto-dismiss durations, and allows testing notification chimes.
-->
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
  import BentoCard from '../../BentoCard.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
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
  <BentoCard variant="group">
    <SettingItem
      label="Do Not Disturb Mode"
      icon={Bell}
      isDefault={$dndStore === dndStore.defaultValue}
      onReset={() => dndStore.reset()}
    >
      <ToggleSwitch bind:checked={$dndStore} ariaLabel="Do Not Disturb Mode" />
    </SettingItem>

    <SettingItem
      label="Play Sound on Notification"
      icon={BellRing}
      isDefault={$notifSoundStore === notifSoundStore.defaultValue}
      onReset={() => notifSoundStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$notifSoundStore}
        ariaLabel="Play Sound on Notification"
      />
    </SettingItem>

    <SettingItem
      label="Notification Sound Volume"
      description="Audio feedback level for incoming notifications"
      icon={Volume2}
      flexCol={true}
      isDefault={Math.round($notifVolumeStore * 100) ===
        Math.round(notifVolumeStore.defaultValue * 100)}
      onReset={() => notifVolumeStore.reset()}
    >
      <SettingSlider
        min={0}
        max={100}
        step={5}
        unit="%"
        defaultValue={Math.round(notifVolumeStore.defaultValue * 100)}
        value={Math.round($notifVolumeStore * 100)}
        onchange={(val) => ($notifVolumeStore = val / 100)}
      />
    </SettingItem>

    <SettingItem
      label="Auto-Wake Island on Notification"
      icon={Bell}
      isDefault={$notifAutoWakeStore === notifAutoWakeStore.defaultValue}
      onReset={() => notifAutoWakeStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$notifAutoWakeStore}
        ariaLabel="Auto-Wake Island on Notification"
      />
    </SettingItem>

    <SettingItem
      label="Show Notification Satellite"
      description="Display alerts in a companion satellite capsule beside the island during media playback and other active views"
      icon={Orbit}
      isDefault={$notifSatelliteEnabledStore ===
        notifSatelliteEnabledStore.defaultValue}
      onReset={() => notifSatelliteEnabledStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$notifSatelliteEnabledStore}
        ariaLabel="Show Notification Satellite"
      />
    </SettingItem>
  </BentoCard>

  <div class="setting-section-header">
    <span>Alarms & Reminders</span>
  </div>

  <BentoCard variant="group">
    <SettingItem
      label="Alarm Ringing Sound"
      description="Procedural synthesis chime played when scheduled alarms fire"
      icon={AlarmClock}
      flexCol={true}
      isDefault={$alarmSoundPresetStore === alarmSoundPresetStore.defaultValue}
      onReset={() => alarmSoundPresetStore.reset()}
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
      isDefault={Math.round($alarmVolumeStore * 100) ===
        Math.round(alarmVolumeStore.defaultValue * 100)}
      onReset={() => alarmVolumeStore.reset()}
    >
      <SettingSlider
        min={10}
        max={100}
        step={5}
        unit="%"
        defaultValue={Math.round(alarmVolumeStore.defaultValue * 100)}
        value={Math.round($alarmVolumeStore * 100)}
        onchange={(val) => ($alarmVolumeStore = val / 100)}
      />
    </SettingItem>

    <SettingItem
      label="Default Snooze Duration"
      description="Interval duration before a snoozed alarm rings again"
      icon={TimerReset}
      isDefault={$alarmSnoozeMinutesStore ===
        alarmSnoozeMinutesStore.defaultValue}
      onReset={() => {
        alarmSnoozeMinutesStore.reset()
        snoozeString = String(alarmSnoozeMinutesStore.defaultValue)
      }}
    >
      <div class="snooze-select-box">
        <CustomSelect
          options={snoozeOptions}
          bind:value={snoozeString}
          placeholder="Duration..."
        />
      </div>
    </SettingItem>
  </BentoCard>

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

<style src="../../styles/settings/NotificationsTab.css"></style>
