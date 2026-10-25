<script lang="ts">
  import { getContext } from 'svelte'
  import { CalendarDays, Plus } from 'lucide-svelte'
  import {
    calendarStore,
    calendarPendingActionStore,
    formatDateKey,
    type CalendarEvent
  } from '../stores/calendarStore'
  import CalendarPage, {
    config as calendarConfig
  } from '../pages/CalendarPage.svelte'
  import ActionCapsule from './ActionCapsule.svelte'
  import ScrollView from './ScrollView.svelte'
  import BentoCard from './BentoCard.svelte'
  import AgendaEventCard from './AgendaEventCard.svelte'
  import { dateFormatStore, formatUserDate, type IslandConfig } from '../store'

  interface Props {
    onOpenCalendar?: () => void
  }

  let { onOpenCalendar }: Props = $props()

  const navigate = getContext<
    | ((page: unknown, pageConfig: IslandConfig, expand?: boolean) => void)
    | undefined
  >('navigate')

  let realToday = $state(new Date())

  function computeTodayEvents(): CalendarEvent[] {
    const todayStr = formatDateKey(realToday)
    return $calendarStore.events
      .filter((e) => e.date === todayStr)
      .sort((a, b) => {
        if (a.isAllDay && !b.isAllDay) return -1
        if (!a.isAllDay && b.isAllDay) return 1
        return (a.startTime || '').localeCompare(b.startTime || '')
      })
  }

  let todayEvents = $derived.by(computeTodayEvents)

  function computeTodayDateLabel(): string {
    return formatUserDate(realToday, $dateFormatStore)
  }

  let todayDateLabel = $derived.by(computeTodayDateLabel)

  function handleOpenCalendarPage() {
    if (onOpenCalendar) {
      onOpenCalendar()
    } else if (navigate) {
      navigate(CalendarPage, calendarConfig, true)
    }
  }

  function handleAddEvent() {
    calendarPendingActionStore.set({
      action: 'add',
      date: formatDateKey(realToday)
    })
    handleOpenCalendarPage()
  }
</script>

<BentoCard variant="bento" class="calendar-glance-card">
  <div class="glance-header">
    <div class="glance-title-group">
      <CalendarDays size={13} />
      <span>{todayDateLabel}</span>
      {#if todayEvents.length > 0}
        <span class="glance-event-count">
          • {todayEvents.length}
          {todayEvents.length === 1 ? 'event' : 'events'}
        </span>
      {/if}
    </div>
    <div class="glance-actions">
      <button
        type="button"
        class="glance-add-btn"
        onclick={handleAddEvent}
        aria-label="Add new calendar event"
      >
        <Plus size={12} />
        <span>Add Event</span>
      </button>
      <ActionCapsule
        label="Calendar"
        onclick={handleOpenCalendarPage}
        tooltip="Open full Calendar page"
      />
    </div>
  </div>

  <div class="glance-agenda-container">
    {#if todayEvents.length === 0}
      <AgendaEventCard isEmpty={true} />
    {:else}
      <ScrollView class="glance-scroll-view">
        <div class="glance-events-list">
          {#each todayEvents as event (event.id)}
            <AgendaEventCard {event} variant="compact" />
          {/each}
        </div>
      </ScrollView>
    {/if}
  </div>
</BentoCard>

<style src="./styles/CalendarGlance.css"></style>
