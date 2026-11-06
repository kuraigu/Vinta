<!--
  @component AgendaEventCard
  @description Calendar event row showing event title, start and end times, color pill,
  and relative countdown ('in 15 mins').
-->
<script lang="ts">
  import { Pencil, Trash2, MapPin, BellRing } from 'lucide-svelte'
  import DiamondPattern from './patterns/DiamondPattern.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { timeFormatStore, dateFormatStore, formatUserDate } from '../store'
  import type { CalendarEvent } from '../stores/calendarStore'
  import './styles/AgendaEventCard.css'

  interface Props {
    event?: CalendarEvent
    variant?: 'compact' | 'detailed'
    isEmpty?: boolean
    emptyTitle?: string
    emptyChip?: string
    onEdit?: (event: CalendarEvent) => void
    onDelete?: (eventId: string) => void
  }

  let {
    event,
    variant = 'compact',
    isEmpty = false,
    emptyTitle = 'Free all day',
    emptyChip = 'All Day',
    onEdit,
    onDelete
  }: Props = $props()

  function formatEventTime(timeStr?: string): string {
    if (!timeStr) return ''
    if ($timeFormatStore === '24h') return timeStr
    const [hStr, mStr] = timeStr.split(':')
    const h = parseInt(hStr, 10)
    if (isNaN(h)) return timeStr
    const suffix = h >= 12 ? 'PM' : 'AM'
    const h12 = h % 12 || 12
    return `${h12}:${mStr} ${suffix}`
  }
</script>

{#if isEmpty || !event}
  <div class="tray-event-card tray-empty-card">
    <div class="tray-card-main">
      <span class="tray-time-chip">{emptyChip}</span>
      <span class="tray-card-title">{emptyTitle}</span>
    </div>
    <DiamondPattern align="center" size="compact" />
  </div>
{:else if variant === 'detailed'}
  <div class="agenda-card card-color-{event.color}">
    <div class="agenda-card-left">
      <div class="agenda-card-title-row">
        <span class="agenda-card-title">{event.title}</span>
        <div class="agenda-card-badges">
          {#if event.notified}
            <span
              class="agenda-card-notified-badge"
              use:islandTooltip={'Notified'}
              aria-label="Notification delivered"
            >
              <BellRing size={9.5} />
            </span>
          {/if}
          <span class="agenda-card-date-badge">
            {formatUserDate(event.date, $dateFormatStore)}
          </span>
        </div>
      </div>
      <div class="agenda-card-meta-row">
        <span class="agenda-card-time">
          {#if event.isAllDay}
            All Day
          {:else if event.startTime}
            {formatEventTime(event.startTime)}{event.endTime
              ? ` - ${formatEventTime(event.endTime)}`
              : ''}
          {:else}
            Scheduled
          {/if}
        </span>
        {#if event.location}
          <span>• {event.location}</span>
        {/if}
      </div>
    </div>

    <DiamondPattern align="center" />

    {#if onEdit || onDelete}
      <div class="tray-card-actions">
        {#if onEdit}
          <button
            type="button"
            class="tray-card-action-btn"
            onclick={() => onEdit(event)}
            aria-label="Edit event"
            use:islandTooltip={'Edit event'}
          >
            <Pencil size={12} />
          </button>
        {/if}
        {#if onDelete}
          <button
            type="button"
            class="tray-card-action-btn danger"
            onclick={() => onDelete(event.id)}
            aria-label="Delete event"
            use:islandTooltip={'Delete event'}
          >
            <Trash2 size={12} />
          </button>
        {/if}
      </div>
    {/if}
  </div>
{:else}
  <div class="tray-event-card card-color-{event.color}">
    <div class="tray-card-main">
      <div class="tray-card-chips-row">
        <span class="tray-time-chip">
          {#if event.isAllDay}
            All Day
          {:else if event.startTime}
            {formatEventTime(event.startTime)}{event.endTime
              ? ` - ${formatEventTime(event.endTime)}`
              : ''}
          {:else}
            Event
          {/if}
        </span>
        {#if event.notified}
          <span
            class="tray-card-notified-chip"
            use:islandTooltip={'Notified'}
            aria-label="Notification delivered"
          >
            <BellRing size={8} />
          </span>
        {/if}
      </div>
      <span class="tray-card-title">{event.title}</span>
      {#if event.location}
        <span class="tray-card-loc">
          <MapPin size={9} />
          {event.location}
        </span>
      {/if}
    </div>

    <DiamondPattern align="center" size="compact" />

    {#if onEdit || onDelete}
      <div class="tray-card-actions">
        {#if onEdit}
          <button
            type="button"
            class="tray-card-action-btn"
            onclick={() => onEdit(event)}
            aria-label="Edit event"
            use:islandTooltip={'Edit event'}
          >
            <Pencil size={11} />
          </button>
        {/if}
        {#if onDelete}
          <button
            type="button"
            class="tray-card-action-btn danger"
            onclick={() => onDelete(event.id)}
            aria-label="Delete event"
            use:islandTooltip={'Delete event'}
          >
            <Trash2 size={11} />
          </button>
        {/if}
      </div>
    {/if}
  </div>
{/if}
