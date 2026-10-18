<script lang="ts">
  import { Check, X } from 'lucide-svelte'
  import CustomTimePicker from '../CustomTimePicker.svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import type { AlarmRepeat } from '../../store'

  interface Props {
    isOpen: boolean
    isEditing: boolean
    initialTime?: string
    initialLabel?: string
    initialRepeat?: AlarmRepeat
    initialDays?: number[]
    onSave: (payload: {
      time: string
      label: string
      repeat: AlarmRepeat
      days: number[]
    }) => void
    onCancel: () => void
  }

  let {
    isOpen,
    isEditing,
    initialTime = '09:00',
    initialLabel = '',
    initialRepeat = 'once',
    initialDays = [1, 2, 3, 4, 5],
    onSave,
    onCancel
  }: Props = $props()

  let time = $state('09:00')
  let label = $state('')
  let repeat = $state<AlarmRepeat>('once')
  let days = $state<number[]>([1, 2, 3, 4, 5])

  const DAY_NAMES = [
    'Sunday',
    'Monday',
    'Tuesday',
    'Wednesday',
    'Thursday',
    'Friday',
    'Saturday'
  ]
  const DAY_SHORT = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

  function syncAlarmForm() {
    if (isOpen) {
      time = initialTime
      label = initialLabel
      repeat = initialRepeat
      days = [...initialDays]
    }
  }
  $effect(syncAlarmForm)

  function toggleDay(dayIdx: number) {
    if (days.includes(dayIdx)) {
      if (days.length > 1) {
        days = days.filter((d) => d !== dayIdx)
      }
    } else {
      days = [...days, dayIdx].sort((a, b) => a - b)
    }
  }

  function handleSave() {
    if (!time) return
    onSave({
      time,
      label,
      repeat,
      days:
        repeat === 'daily'
          ? [0, 1, 2, 3, 4, 5, 6]
          : repeat === 'weekdays'
            ? [1, 2, 3, 4, 5]
            : days
    })
  }
</script>

{#if isOpen}
  <div class="alarm-create-card">
    <span class="create-card-title">
      {isEditing ? 'Edit Alarm Reminder' : 'New Alarm Reminder'}
    </span>
    <div class="create-inputs-row">
      <CustomTimePicker bind:value={time} ariaLabel="Alarm time" />

      <input
        type="text"
        class="alarm-label-input"
        placeholder="Reminder label..."
        bind:value={label}
        aria-label="Alarm note"
      />
    </div>

    <div class="alarm-repeat-selector">
      <span class="repeat-row-label">Repeat:</span>
      <div class="repeat-pills-row">
        <button
          class="repeat-pill"
          class:active={repeat === 'once'}
          onclick={() => {
            repeat = 'once'
          }}>Once</button
        >
        <button
          class="repeat-pill"
          class:active={repeat === 'daily'}
          onclick={() => {
            repeat = 'daily'
          }}>Daily</button
        >
        <button
          class="repeat-pill"
          class:active={repeat === 'weekdays'}
          onclick={() => {
            repeat = 'weekdays'
          }}>Weekdays</button
        >
        <button
          class="repeat-pill"
          class:active={repeat === 'custom'}
          onclick={() => {
            repeat = 'custom'
          }}>Custom</button
        >
      </div>
    </div>

    {#if repeat === 'custom'}
      <div class="custom-days-row">
        {#each DAY_SHORT as dayLabel, idx (idx)}
          <button
            class="day-toggle-btn"
            class:active={days.includes(idx)}
            onclick={() => toggleDay(idx)}
            use:islandTooltip={DAY_NAMES[idx]}
          >
            {dayLabel}
          </button>
        {/each}
      </div>
    {/if}

    <div class="create-card-actions">
      <button
        class="card-act-btn save"
        onclick={handleSave}
        use:islandTooltip={'Save reminder'}
      >
        <Check size={14} />
        <span>Save</span>
      </button>
      <button
        class="card-act-btn cancel"
        onclick={onCancel}
        use:islandTooltip={'Cancel'}
      >
        <X size={14} />
        <span>Cancel</span>
      </button>
    </div>
  </div>
{/if}
