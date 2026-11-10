<!--
  @component AlarmsTab
  @description Scheduled alarms manager. Lists active and inactive alarms with toggle switches,
  repeat day indicators, and buttons to create or edit alarms.
-->
<script lang="ts">
  import { Bell, Pencil, Trash2 } from 'lucide-svelte'
  import ScrollView from '../ScrollView.svelte'
  import EmptyState from '../EmptyState.svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import {
    timerAppStore,
    toggleAlarm,
    deleteAlarm,
    type AlarmItem
  } from '../../store'

  interface Props {
    is24H: boolean
    onStartEdit: (alarm: AlarmItem) => void
    onOpenCreate: () => void
  }

  let { is24H, onStartEdit, onOpenCreate }: Props = $props()

  const DAY_SHORT = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

  function formatAlarmDisplayTime(timeStr: string, use24H: boolean): string {
    if (!timeStr) return ''
    const [hStr, mStr] = timeStr.split(':')
    const h24 = parseInt(hStr, 10) || 0
    const m = (mStr || '00').padStart(2, '0')

    if (use24H) {
      return `${h24.toString().padStart(2, '0')}:${m}`
    }

    const h12 = h24 % 12 || 12
    const ampm = h24 >= 12 ? 'PM' : 'AM'
    return `${h12}:${m} ${ampm}`
  }
</script>

<div class="tab-panel alarms-panel">
  <div class="alarms-list-container">
    {#if $timerAppStore.alarms.length === 0}
      <EmptyState
        icon={Bell}
        title="No active alarms"
        description="Set a scheduled reminder to keep track of tasks or alerts."
        actionLabel="Add Alarm"
        onAction={onOpenCreate}
      />
    {:else}
      <ScrollView class="alarms-scroll-view">
        {#each $timerAppStore.alarms as alarm (alarm.id)}
          <div class="alarm-item-card" class:disabled={!alarm.enabled}>
            <button
              class="alarm-item-info"
              onclick={() => onStartEdit(alarm)}
              use:islandTooltip={'Click to edit alarm'}
            >
              <div class="alarm-time-row">
                <span class="alarm-item-time"
                  >{formatAlarmDisplayTime(alarm.time, is24H)}</span
                >
                {#if alarm.repeat === 'once'}
                  <span class="alarm-repeat-badge">Once</span>
                {:else if alarm.repeat === 'daily'}
                  <span class="alarm-repeat-badge">Every day</span>
                {:else if alarm.repeat === 'weekdays'}
                  <div class="alarm-days-circles">
                    {#each [1, 2, 3, 4, 5] as d (d)}
                      <span class="alarm-day-circle">{DAY_SHORT[d]}</span>
                    {/each}
                  </div>
                {:else if alarm.repeat === 'custom' && alarm.days && alarm.days.length > 0}
                  <div class="alarm-days-circles">
                    {#each [...alarm.days].sort((a, b) => a - b) as d (d)}
                      <span class="alarm-day-circle">{DAY_SHORT[d]}</span>
                    {/each}
                  </div>
                {:else}
                  <span class="alarm-repeat-badge">Custom</span>
                {/if}
              </div>
              <span class="alarm-item-label">{alarm.label || 'Reminder'}</span>
            </button>

            <div class="alarm-item-actions">
              <button
                class="alarm-action-btn"
                onclick={() => onStartEdit(alarm)}
                use:islandTooltip={'Edit alarm'}
              >
                <Pencil size={13} />
              </button>

              <label class="drawer-switch mini-switch">
                <input
                  type="checkbox"
                  checked={alarm.enabled}
                  onchange={() => toggleAlarm(alarm.id)}
                />
                <span class="switch-slider"></span>
              </label>

              <button
                class="alarm-delete-btn"
                onclick={() => deleteAlarm(alarm.id)}
                use:islandTooltip={'Delete'}
              >
                <Trash2 size={13} />
              </button>
            </div>
          </div>
        {/each}
      </ScrollView>
    {/if}
  </div>
</div>

<style src="../../pages/styles/TimerPage.css"></style>
