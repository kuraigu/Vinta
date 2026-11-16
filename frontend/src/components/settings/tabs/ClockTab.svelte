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
  import BentoCard from '../../BentoCard.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
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
  <BentoCard variant="group">
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
  </BentoCard>

  <BentoCard variant="group">
    <SettingItem
      label="Auto-Discard Past Events"
      description="Automatically purges agenda events after their scheduled date has passed"
      icon={CalendarX}
      isDefault={$autoDiscardPastEventsStore ===
        autoDiscardPastEventsStore.defaultValue}
      onReset={() => autoDiscardPastEventsStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$autoDiscardPastEventsStore}
        ariaLabel="Auto-Discard Past Events"
      />
    </SettingItem>

    <SettingItem
      label="Clear Past Events"
      description="Manually removes all expired agenda items preceding today"
      icon={Trash2}
      flexCol={true}
    >
      <SettingActionButton
        variant="default"
        status={clearedCount !== null ? 'success' : 'idle'}
        successLabel={clearedCount !== null
          ? `Removed ${clearedCount} ${clearedCount === 1 ? 'event' : 'events'}`
          : undefined}
        fullWidth={true}
        onclick={handleClearPast}
      >
        <Trash2 size={13} />
        <span>Clear Past Events</span>
      </SettingActionButton>
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/ClockTab.css"></style>
