<!--
  @component AlarmEditorDrawer
  @description Slide-over drawer for creating or editing an alarm. Provides time picker,
  label input, repeat weekday check-boxes, and chime sound preview selector.
-->
<script lang="ts">
  import { Check } from 'lucide-svelte'
  import CustomTimePicker from '../CustomTimePicker.svelte'
  import DrawerPanel from '../DrawerPanel.svelte'
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
  <DrawerPanel
    title={isEditing ? 'Edit Alarm Reminder' : 'New Alarm Reminder'}
    variant="overlay"
    closeTooltip="Close"
    onClose={onCancel}
  >
    <div class="drawer-content">
      <div class="drawer-row">
        <span class="drawer-row-label">Time</span>
        <div class="alarm-drawer-time-wrapper">
          <CustomTimePicker
            bind:value={time}
            ariaLabel="Alarm time"
            align="right"
          />
        </div>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Label</span>
        <input
          type="text"
          class="alarm-label-input"
          placeholder="Reminder label..."
          bind:value={label}
          aria-label="Alarm note"
        />
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Repeat</span>
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
        <div class="drawer-row">
          <span class="drawer-row-label">Days</span>
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
        </div>
      {/if}

      <button class="drawer-confirm-btn" onclick={handleSave}>
        <Check size={16} />
        <span>{isEditing ? 'Save Changes' : 'Create Alarm'}</span>
      </button>
    </div>
  </DrawerPanel>
{/if}

<style src="../../pages/styles/TimerPage.css"></style>
