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

  import { onMount } from 'svelte'

  const dateFormatOptions = [
    { label: 'Thu, Oct 15 (Short)', value: 'short' },
    { label: 'Thursday, October 15 (Full)', value: 'full' },
    { label: '10/15/2026 (MM/DD/YYYY)', value: 'us_numeric' },
    { label: '15/10/2026 (DD/MM/YYYY)', value: 'eu_numeric' },
    { label: 'October 15, 2026 (Month D, YYYY)', value: 'us_long' },
    { label: '15 October 2026 (D Month YYYY)', value: 'eu_long' },
    { label: '2026-10-15 (ISO YYYY-MM-DD)', value: 'iso' }
  ]

  onMount(() => {
    if ($dateFormatStore === 'default') {
      dateFormatStore.set('short')
    } else if ($dateFormatStore === 'us') {
      dateFormatStore.set('us_numeric')
    } else if ($dateFormatStore === 'eu') {
      dateFormatStore.set('eu_numeric')
    }
  })

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
  <SettingItem
    label="Time Format"
    description="Clock presentation style (Default: 12-Hour)"
    icon={Clock}
    flexCol={true}
  >
    <CustomSelect
      options={timeFormatOptions}
      bind:value={$timeFormatStore}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem
    label="Timezone"
    description="Active timezone for clock and agenda calculations (Default: Auto-detect)"
    icon={Globe}
    flexCol={true}
  >
    <CustomSelect
      options={timezoneOptions}
      bind:value={selectedTz}
      searchable={true}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem
    label="Date Display Format"
    description="Format used across date widgets (Default: Short)"
    icon={Calendar}
    flexCol={true}
  >
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
