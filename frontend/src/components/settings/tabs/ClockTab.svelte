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
    getInitialTimezone,
    tStore
  } from '../../../store'
  import { discardPastCalendarEvents } from '../../../stores/calendarStore'

  const timeFormatOptions = $derived([
    { label: $tStore('settings.clock_12h'), value: '12h' },
    { label: $tStore('settings.clock_24h'), value: '24h' }
  ])

  const dateFormatOptions = $derived([
    { label: $tStore('settings.clock_date_short'), value: 'short' },
    { label: $tStore('settings.clock_date_full'), value: 'full' },
    { label: '10/15/2026 (MM/DD/YYYY)', value: 'us_numeric' },
    { label: '15/10/2026 (DD/MM/YYYY)', value: 'eu_numeric' },
    { label: $tStore('settings.clock_date_us_long'), value: 'us_long' },
    { label: $tStore('settings.clock_date_eu_long'), value: 'eu_long' },
    { label: '2026-10-15 (ISO YYYY-MM-DD)', value: 'iso' }
  ])

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

  const timezoneOptions = $derived([
    { label: $tStore('settings.clock_timezone_auto'), value: 'auto' },
    ...rawTimezones.map((tz) => ({
      label: tz.replace(/_/g, ' '),
      value: tz
    }))
  ])

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
      label={$tStore('settings.clock_time_format_label')}
      description={$tStore('settings.clock_time_format_desc')}
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
      label={$tStore('settings.clock_timezone_label')}
      description={$tStore('settings.clock_timezone_desc')}
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
      label={$tStore('settings.clock_date_format_label')}
      description={$tStore('settings.clock_date_format_desc')}
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
      label={$tStore('settings.clock_auto_discard_label')}
      description={$tStore('settings.clock_auto_discard_desc')}
      icon={CalendarX}
      isDefault={$autoDiscardPastEventsStore ===
        autoDiscardPastEventsStore.defaultValue}
      onReset={() => autoDiscardPastEventsStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$autoDiscardPastEventsStore}
        ariaLabel={$tStore('settings.clock_auto_discard_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.clock_clear_past_label')}
      description={$tStore('settings.clock_clear_past_desc')}
      icon={Trash2}
      flexCol={true}
    >
      <SettingActionButton
        variant="default"
        status={clearedCount !== null ? 'success' : 'idle'}
        successLabel={clearedCount !== null
          ? $tStore(
              clearedCount === 1
                ? 'settings.clock_removed_one'
                : 'settings.clock_removed_many',
              { count: clearedCount }
            )
          : undefined}
        fullWidth={true}
        onclick={handleClearPast}
      >
        <Trash2 size={14} />
        <span>{$tStore('settings.clock_clear_past_label')}</span>
      </SettingActionButton>
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/ClockTab.css"></style>
