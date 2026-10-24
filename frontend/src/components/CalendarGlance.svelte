<script lang="ts">
  import { getContext } from 'svelte'
  import { CalendarDays, Plus, MapPin } from 'lucide-svelte'
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
  import DiamondPattern from './patterns/DiamondPattern.svelte'
  import BentoCard from './BentoCard.svelte'
  import {
    timeFormatStore,
    dateFormatStore,
    formatUserDate,
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

  function formatEventTime(timeStr: string): string {
    if (!timeStr) return ''
    if ($timeFormatStore === '24h') return timeStr
    const [hStr, mStr] = timeStr.split(':')
    const h = parseInt(hStr, 10)
    if (isNaN(h)) return timeStr
    const suffix = h >= 12 ? 'PM' : 'AM'
    const h12 = h % 12 || 12
    return `${h12}:${mStr} ${suffix}`
  }

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
      <div class="glance-empty-card">
        <div class="empty-main">
          <span class="empty-chip">All Day</span>
          <span class="empty-title">Free all day</span>
        </div>
        <DiamondPattern align="center" size="compact" />
      </div>
    {:else}
      <ScrollView class="glance-scroll-view">
        <div class="glance-events-list">
          {#each todayEvents as event (event.id)}
            <div class="glance-event-item card-color-{event.color}">
              <div class="event-main">
                <span class="event-time-chip">
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
                <span class="event-title">{event.title}</span>
                {#if event.location}
                  <span class="event-location">
                    <MapPin size={9} />
                    {event.location}
                  </span>
                {/if}
              </div>
              <DiamondPattern align="center" size="compact" />
            </div>
          {/each}
        </div>
      </ScrollView>
    {/if}
  </div>
</BentoCard>

<style src="./styles/CalendarGlance.css"></style>
