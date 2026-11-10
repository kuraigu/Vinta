<!--
  @component ClockTab
  @description Clock display preferences. Switches between 12-hour and 24-hour modes,
  configures date formatting (US, EU, ISO), time zone offsets, and blinking separators.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import {
    Clock,
    Globe,
    Calendar,
    CalendarX,
    Trash2,
    Check
  } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import {
    timeFormatStore,
    timezoneStore,
    dateFormatStore,
    autoDiscardPastEventsStore,
    getInitialTimezone
  } from '../../../store'
  import { discardPastCalendarEvents } from '../../../stores/calendarStore'

  const timeFormatOptions = [
    { label: '12-Hour (1:30 PM)', value: '12h' },
    { label: '24-Hour (13:30)', value: '24h' }
  ]

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

  let clearedCount = $state<number | null>(null)

  function handleClearPast() {
    const count = discardPastCalendarEvents()
    clearedCount = count
    setTimeout(() => {
      clearedCount = null
    }, 2500)
  }

  function handleTimezoneSelectionEffect() {
    if (selectedTz === 'auto') {
      $timezoneStore = getInitialTimezone()
    } else {
      $timezoneStore = selectedTz
    }
  }
  $effect(handleTimezoneSelectionEffect)
</script>

<div class="setting-group">
  <SettingItem
    label="Time Format"
    description="Clock presentation style"
    icon={Clock}
    flexCol={true}
    isDefault={$timeFormatStore === timeFormatStore.defaultValue}
    onReset={() => timeFormatStore.reset()}
  >
    <CustomSelect
      options={timeFormatOptions}
      bind:value={$timeFormatStore}
      defaultValue={timeFormatStore.defaultValue}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem
    label="Timezone"
    description="Active timezone for clock and agenda calculations"
    icon={Globe}
    flexCol={true}
    isDefault={selectedTz === 'auto'}
    onReset={() => (selectedTz = 'auto')}
  >
    <CustomSelect
      options={timezoneOptions}
      bind:value={selectedTz}
      defaultValue="auto"
      searchable={true}
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem
    label="Date Display Format"
    description="Format used across date widgets"
    icon={Calendar}
    flexCol={true}
    isDefault={$dateFormatStore === 'short'}
    onReset={() => dateFormatStore.set('short')}
  >
    <CustomSelect
      options={dateFormatOptions}
      bind:value={$dateFormatStore}
      defaultValue="short"
      fullWidth={true}
      direction="auto"
    />
  </SettingItem>

  <SettingItem
    label="Auto-Discard Past Events"
    description="Automatically purges agenda events after their scheduled date has passed"
    icon={CalendarX}
  >
    <label class="toggle-switch">
      <input type="checkbox" bind:checked={$autoDiscardPastEventsStore} />
      <span class="slider"></span>
    </label>
  </SettingItem>

  <SettingItem
    label="Clear Past Events"
    description="Manually removes all expired agenda items preceding today"
    icon={Trash2}
    flexCol={true}
  >
    <SettingActionButton
      variant={clearedCount !== null ? 'success' : 'default'}
      fullWidth={true}
      onclick={handleClearPast}
    >
      {#if clearedCount !== null}
        <Check size={13} />
        <span
          >Removed {clearedCount}
          {clearedCount === 1 ? 'event' : 'events'}</span
        >
      {:else}
        <Trash2 size={13} />
        <span>Clear Past Events</span>
      {/if}
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
    width: 32px;
    height: 18px;
    flex-shrink: 0;
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
</style>
