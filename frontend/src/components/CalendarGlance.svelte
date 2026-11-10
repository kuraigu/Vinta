<!--
  @component CalendarGlance
  @description Minimized calendar card displaying the upcoming agenda event and its countdown.
-->
<script lang="ts">
  import { getContext } from 'svelte'
  import { CalendarDays, Plus, ChevronRight } from 'lucide-svelte'
  import {
    calendarStore,
    calendarPendingActionStore,
    formatDateKey,
    type CalendarEvent
  } from '../stores/calendarStore'
  import CalendarPage, {
    config as calendarConfig
  } from '../pages/CalendarPage.svelte'
  import ScrollView from './ScrollView.svelte'
  import BentoCard from './BentoCard.svelte'
  import AgendaEventCard from './AgendaEventCard.svelte'
  import {
    dateFormatStore,
    formatUserDate,
    tStore,
    type IslandConfig
  } from '../store'

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
    <div
      class="glance-title-group"
      role="button"
      tabindex="0"
      onclick={handleOpenCalendarPage}
      onkeydown={(e) => e.key === 'Enter' && handleOpenCalendarPage()}
    >
      <div class="glance-icon-badge">
        <CalendarDays size={12} />
      </div>
      <span class="glance-title-text">{todayDateLabel}</span>
      {#if todayEvents.length > 0}
        <span class="glance-event-count">
          • {todayEvents.length === 1
            ? $tStore('calendar.event_single', { count: todayEvents.length })
            : $tStore('calendar.event_plural', { count: todayEvents.length })}
        </span>
      {/if}
    </div>
    <div class="glance-actions">
      <button
        type="button"
        class="glance-add-btn"
        onclick={handleAddEvent}
        aria-label={$tStore('calendar.add_event')}
      >
        <Plus size={12} />
        <span>{$tStore('calendar.add_event')}</span>
      </button>
      <button
        type="button"
        class="glance-nav-btn"
        onclick={handleOpenCalendarPage}
        aria-label={$tStore('calendar.open_calendar')}
      >
        <ChevronRight size={14} class="glance-nav-chevron" />
      </button>
    </div>
  </div>

  <div class="glance-agenda-container">
    {#if todayEvents.length === 0}
      <AgendaEventCard
        isEmpty={true}
        emptyTitle={$tStore('calendar.no_events_today')}
        emptyChip={$tStore('calendar.all_day')}
      />
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
