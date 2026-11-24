<!--
  @component AgendaEventCard
  @description Calendar event row showing event title, start and end times, color pill,
  and relative countdown ('in 15 mins').
-->
<script lang="ts">
  import { Pencil, Trash2, MapPin, BellRing, Video, Flag } from 'lucide-svelte'
  import DiamondPattern from './patterns/DiamondPattern.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { timeFormatStore, dateFormatStore, formatUserDate } from '../store'
  import {
    extractMeetingUrl,
    type CalendarEvent
  } from '../stores/calendarStore'
  import { OpenURL } from '../../bindings/Vinta/backend/app.js'
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

  function getSourceLabel(evt: CalendarEvent): string {
    if (evt.source === 'holiday' || evt.isHoliday) return 'Holiday'
    if (evt.source === 'google') return 'Google'
    if (evt.source === 'synced') return 'iCal'
    return ''
  }

  function openMeeting(url: string) {
    if (typeof OpenURL === 'function') {
      void OpenURL(url)
    } else {
      window.open(url, '_blank', 'noopener,noreferrer')
    }
  }

  let meetingUrl = $derived(event ? extractMeetingUrl(event) : null)
  let isHoliday = $derived(
    event ? Boolean(event.isHoliday || event.source === 'holiday') : false
  )
  let isReadOnly = $derived(event ? event.source !== 'local' : false)
  let sourceLabel = $derived(event ? getSourceLabel(event) : '')
</script>

{#if isEmpty || !event}
  <div class="tray-event-card tray-empty-card">
    <div class="tray-card-main">
      <span class="tray-time-chip">{emptyChip}</span>
      <span class="tray-card-title">{emptyTitle}</span>
    </div>
    <DiamondPattern align="right" size="compact" />
  </div>
{:else if variant === 'detailed'}
  <div class="agenda-card card-color-{event.color}">
    <div class="agenda-card-left">
      <div class="agenda-card-title-row">
        {#if isHoliday}
          <span class="agenda-holiday-icon" aria-hidden="true">
            <Flag size={14} />
          </span>
        {/if}
        <span class="agenda-card-title">{event.title}</span>
        <div class="agenda-card-badges">
          {#if sourceLabel}
            <span class="tray-source-pill">{sourceLabel}</span>
          {/if}
          {#if event.notified && !isHoliday}
            <span
              class="agenda-card-notified-badge"
              use:islandTooltip={'Notified'}
              aria-label="Notification delivered"
            >
              <BellRing size={12} />
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

    {#if meetingUrl || (!isReadOnly && (onEdit || onDelete))}
      <div class="tray-card-actions">
        {#if meetingUrl}
          <button
            type="button"
            class="tray-card-action-btn"
            onclick={() => openMeeting(meetingUrl)}
            aria-label="Join meeting"
            use:islandTooltip={'Join meeting'}
          >
            <Video size={14} />
          </button>
        {/if}
        {#if !isReadOnly && onEdit}
          <button
            type="button"
            class="tray-card-action-btn"
            onclick={() => onEdit(event)}
            aria-label="Edit event"
            use:islandTooltip={'Edit event'}
          >
            <Pencil size={14} />
          </button>
        {/if}
        {#if !isReadOnly && onDelete}
          <button
            type="button"
            class="tray-card-action-btn danger"
            onclick={() => onDelete(event.id)}
            aria-label="Delete event"
            use:islandTooltip={'Delete event'}
          >
            <Trash2 size={14} />
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
        {#if isHoliday}
          <span class="agenda-holiday-icon" aria-hidden="true">
            <Flag size={12} />
          </span>
        {/if}
        {#if event.notified && !isHoliday}
          <span
            class="tray-card-notified-chip"
            use:islandTooltip={'Notified'}
            aria-label="Notification delivered"
          >
            <BellRing size={11} />
          </span>
        {/if}
      </div>
      <span class="tray-card-title">{event.title}</span>
      {#if event.location}
        <span class="tray-card-loc">
          <MapPin size={11} />
          {event.location}
        </span>
      {/if}
    </div>

    <DiamondPattern align="right" size="compact" />

    {#if sourceLabel}
      <span class="tray-source-pill">{sourceLabel}</span>
    {/if}

    {#if meetingUrl || (!isReadOnly && (onEdit || onDelete))}
      <div class="tray-card-actions">
        {#if meetingUrl}
          <button
            type="button"
            class="tray-card-action-btn"
            onclick={() => openMeeting(meetingUrl)}
            aria-label="Join meeting"
            use:islandTooltip={'Join meeting'}
          >
            <Video size={13} />
          </button>
        {/if}
        {#if !isReadOnly && onEdit}
          <button
            type="button"
            class="tray-card-action-btn"
            onclick={() => onEdit(event)}
            aria-label="Edit event"
            use:islandTooltip={'Edit event'}
          >
            <Pencil size={13} />
          </button>
        {/if}
        {#if !isReadOnly && onDelete}
          <button
            type="button"
            class="tray-card-action-btn danger"
            onclick={() => onDelete(event.id)}
            aria-label="Delete event"
            use:islandTooltip={'Delete event'}
          >
            <Trash2 size={13} />
          </button>
        {/if}
      </div>
    {/if}
  </div>
{/if}
