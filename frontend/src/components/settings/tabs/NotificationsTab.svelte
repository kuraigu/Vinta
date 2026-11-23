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
  import SettingSectionHeader from '../SettingSectionHeader.svelte'
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
    notifData,
    tStore
  } from '../../../store'
  import {
    previewAlarmSound,
    type AlarmSoundPreset
  } from '../../../services/audioService'

  const alarmSoundOptions: { value: AlarmSoundPreset; label: string }[] =
    $derived([
      { value: 'chime', label: $tStore('settings.notif_sound_chime') },
      { value: 'radar', label: $tStore('settings.notif_sound_radar') },
      { value: 'digital', label: $tStore('settings.notif_sound_digital') },
      { value: 'marimba', label: $tStore('settings.notif_sound_marimba') }
    ])

  const snoozeOptions = $derived([
    { value: '5', label: $tStore('settings.notif_snooze_mins', { count: 5 }) },
    {
      value: '10',
      label: $tStore('settings.notif_snooze_mins', { count: 10 })
    },
    {
      value: '15',
      label: $tStore('settings.notif_snooze_mins', { count: 15 })
    }
  ])

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
      label={$tStore('settings.notif_dnd_label')}
      icon={Bell}
      isDefault={$dndStore === dndStore.defaultValue}
      onReset={() => dndStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$dndStore}
        ariaLabel={$tStore('settings.notif_dnd_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.notif_sound_label')}
      icon={BellRing}
      isDefault={$notifSoundStore === notifSoundStore.defaultValue}
      onReset={() => notifSoundStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$notifSoundStore}
        ariaLabel={$tStore('settings.notif_sound_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.notif_volume_label')}
      description={$tStore('settings.notif_volume_desc')}
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
      label={$tStore('settings.notif_autowake_label')}
      icon={Bell}
      isDefault={$notifAutoWakeStore === notifAutoWakeStore.defaultValue}
      onReset={() => notifAutoWakeStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$notifAutoWakeStore}
        ariaLabel={$tStore('settings.notif_autowake_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.notif_satellite_label')}
      description={$tStore('settings.notif_satellite_desc')}
      icon={Orbit}
      isDefault={$notifSatelliteEnabledStore ===
        notifSatelliteEnabledStore.defaultValue}
      onReset={() => notifSatelliteEnabledStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$notifSatelliteEnabledStore}
        ariaLabel={$tStore('settings.notif_satellite_label')}
      />
    </SettingItem>
  </BentoCard>

  <SettingSectionHeader title={$tStore('settings.notif_alarms_section')} />

  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.notif_alarm_sound_label')}
      description={$tStore('settings.notif_alarm_sound_desc')}
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
            placeholder={$tStore('settings.notif_alarm_sound_label')}
            fullWidth={true}
          />
        </div>
        <button
          type="button"
          class="sound-preview-btn"
          onclick={handleAuditionAlarmSound}
          aria-label={$tStore('settings.notif_alarm_preview_aria')}
        >
          <Play size={11} />
          <span>{$tStore('settings.notif_alarm_test')}</span>
        </button>
      </div>
    </SettingItem>

    <SettingItem
      label={$tStore('settings.notif_alarm_volume_label')}
      description={$tStore('settings.notif_alarm_volume_desc')}
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
      label={$tStore('settings.notif_snooze_label')}
      description={$tStore('settings.notif_snooze_desc')}
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
          placeholder={$tStore('settings.notif_snooze_label')}
        />
      </div>
    </SettingItem>
  </BentoCard>

  <SettingItem
    label={$tStore('settings.notif_clear_label')}
    description={$tStore('settings.notif_clear_desc', {
      count: $notifData?.length || 0
    })}
    icon={Trash2}
    flexCol={true}
  >
    <SettingActionButton
      variant="danger"
      fullWidth={true}
      onclick={clearAllNotifications}
    >
      {$tStore('settings.notif_clear_label')}
    </SettingActionButton>
  </SettingItem>
</div>

<style src="../../styles/settings/NotificationsTab.css"></style>
