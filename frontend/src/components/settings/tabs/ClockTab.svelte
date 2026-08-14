<script lang="ts">
  import { Clock, Globe, Calendar } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import {
    timeFormatStore,
    timezoneStore,
    dateFormatStore,
    getInitialTimezone
  } from '../../../store'

  const timeFormatOptions = [
    { label: '12-Hour (1:30 PM)', value: '12h' },
    { label: '24-Hour (13:30)', value: '24h' }
  ]

  const dateFormatOptions = [
    { label: 'Short (Mon, Oct 24)', value: 'short' },
    { label: 'Full (Monday, October 24)', value: 'full' },
    { label: 'ISO (2026-10-24)', value: 'iso' }
  ]

  const rawTimezones =
    typeof Intl !== 'undefined' && typeof Intl.supportedValuesOf === 'function'
      ? Intl.supportedValuesOf('timeZone')
      : [
          'UTC',
          'Asia/Manila',
          'Asia/Tokyo',
          'Asia/Singapore',
          'Asia/Seoul',
          'Asia/Hong_Kong',
          'Asia/Dubai',
          'Asia/Bangkok',
          'Europe/London',
          'Europe/Paris',
          'Europe/Berlin',
          'Europe/Rome',
          'America/New_York',
          'America/Chicago',
          'America/Denver',
          'America/Los_Angeles',
          'America/Toronto',
          'America/Sao_Paulo',
          'Australia/Sydney'
        ]

  const timezoneOptions = [
    { label: 'Auto-detect (System Local)', value: 'auto' },
    ...rawTimezones.map((tz) => ({
      label: tz.replace(/_/g, ' '),
      value: tz
    }))
  ]

  let selectedTz = $state(
    $timezoneStore === getInitialTimezone() ? 'auto' : $timezoneStore
  )

  $effect(() => {
    if (selectedTz === 'auto') {
      $timezoneStore = getInitialTimezone()
    } else {
      $timezoneStore = selectedTz
    }
  })
</script>

<div class="setting-group">
  <SettingItem label="Time Format" icon={Clock} flexCol={true}>
    <CustomSelect
      options={timeFormatOptions}
      bind:value={$timeFormatStore}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem label="Timezone" icon={Globe} flexCol={true}>
    <CustomSelect
      options={timezoneOptions}
      bind:value={selectedTz}
      searchable={true}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem label="Date Display Format" icon={Calendar} flexCol={true}>
    <CustomSelect
      options={dateFormatOptions}
      bind:value={$dateFormatStore}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>
</div>

<style>
  .setting-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
</style>
