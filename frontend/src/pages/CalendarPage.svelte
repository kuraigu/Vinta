<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 36, r: 18 },
    expanded: { w: 440, h: 390, r: 28 },
    borderColor: 'var(--c-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { SvelteSet } from 'svelte/reactivity'
  import {
    Calendar as CalendarIcon,
    CalendarDays,
    ListFilter,
    ChevronLeft,
    ChevronRight,
    ChevronDown,
    Plus,
    Pencil,
    Trash2,
    ExternalLink,
    X,
    Clock,
    MapPin
  } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import CustomTimePicker from '../components/CustomTimePicker.svelte'
  import CustomDatePicker from '../components/CustomDatePicker.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from '../components/HorizontalTabSelection.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    calendarStore,
    setSelectedCalendarDate,
    addCalendarEvent,
    updateCalendarEvent,
    deleteCalendarEvent,
    formatDateKey,
    type CalendarEvent
  } from '../stores/calendarStore'

  let { expanded = false } = $props<{ expanded?: boolean }>()

  // Sub-view mode tabs: 'calendar' (month grid) or 'agenda' (chronological stream)
  let activeTab = $state<'calendar' | 'agenda'>('calendar')

  const CALENDAR_TABS: TabItem<'calendar' | 'agenda'>[] = [
    {
      id: 'calendar',
      label: 'Calendar',
      icon: CalendarDays,
      tooltip: 'Month overview'
    },
    {
      id: 'agenda',
      label: 'Agenda',
      icon: ListFilter,
      tooltip: 'Scheduled events agenda'
    }
  ]

  // Calendar view navigation state
  let viewDate = $state(new Date())
  let calViewMode = $state<'month' | 'year' | 'decade'>('month')
  let agendaYearFilter = $state<number | 'all'>('all')
  let isModalOpen = $state(false)
  let modalMode = $state<'create' | 'edit'>('create')
  let editingEventId = $state<string | null>(null)

  const COLOR_OPTIONS: CalendarEvent['color'][] = [
    'indigo',
    'emerald',
    'amber',
    'purple',
    'rose',
    'sky'
  ]

  /**
   * Returns a random category color for newly initiated events.
   */
  function getRandomEventColor(): CalendarEvent['color'] {
    return COLOR_OPTIONS[Math.floor(Math.random() * COLOR_OPTIONS.length)]
  }

  // New Event form fields
  let formTitle = $state('')
  let formDate = $state(formatDateKey(new Date()))
  let formStartTime = $state('09:00')
  let formEndTime = $state('10:00')
  let formIsAllDay = $state(false)
  let formColor = $state<CalendarEvent['color']>(getRandomEventColor())
  let formLocation = $state('')
  let formDescription = $state('')

  const monthNames = [
    'January',
    'February',
    'March',
    'April',
    'May',
    'June',
    'July',
    'August',
    'September',
    'October',
    'November',
    'December'
  ]
  const monthShortNames = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec'
  ]
  const weekDays = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

  /**
   * Computes the current year of the calendar view.
   */
  function computeViewYear(): number {
    return viewDate.getFullYear()
  }
  let viewYear = $derived.by(computeViewYear)

  /**
   * Computes the current month index of the calendar view.
   */
  function computeViewMonth(): number {
    return viewDate.getMonth()
  }
  let viewMonth = $derived.by(computeViewMonth)

  /**
   * Generates blank padding cells for leading days in the monthly view.
   */
  function computeBlanks(): null[] {
    const firstDay = new Date(viewYear, viewMonth, 1).getDay()
    return Array(firstDay).fill(null)
  }
  let blanks = $derived.by(computeBlanks)

  /**
   * Generates the array of day numbers for the active view month.
   */
  function computeDaysInMonth(): number[] {
    const totalDays = new Date(viewYear, viewMonth + 1, 0).getDate()
    return Array.from({ length: totalDays }, (_, i) => i + 1)
  }
  let daysArray = $derived.by(computeDaysInMonth)

  /**
   * Computes the interactive navigation title for month, year, or decade.
   */
  function computeHeaderTitle(): string {
    if (calViewMode === 'month') {
      return `${monthNames[viewMonth]} ${viewYear}`
    }
    if (calViewMode === 'year') {
      return viewYear.toString()
    }
    const startYear = Math.floor(viewYear / 10) * 10
    return `${startYear} - ${startYear + 11}`
  }
  let headerTitle = $derived.by(computeHeaderTitle)

  /**
   * Computes the 12-year window for decade mode navigation.
   */
  function computeDecadeYears(): number[] {
    const startYear = Math.floor(viewYear / 10) * 10 - 1
    return Array.from({ length: 12 }, (_, i) => startYear + i)
  }
  let decadeYears = $derived.by(computeDecadeYears)

  /**
   * Retrieves events matching the selected date key.
   */
  function computeSelectedDayEvents(): CalendarEvent[] {
    return $calendarStore.events.filter(
      (e) => e.date === $calendarStore.selectedDate
    )
  }
  let selectedDayEvents = $derived.by(computeSelectedDayEvents)

  /**
   * Retrieves all scheduled events sorted chronologically for the agenda tab.
   */
  function computeAllAgendaEvents(): CalendarEvent[] {
    return [...$calendarStore.events].sort((a, b) => {
      const cmpDate = a.date.localeCompare(b.date)
      if (cmpDate !== 0) return cmpDate
      return (a.startTime || '').localeCompare(b.startTime || '')
    })
  }
  let allAgendaEvents = $derived.by(computeAllAgendaEvents)

  /**
   * Derives distinct years represented across scheduled events for agenda filtering.
   */
  function computeAgendaYears(): number[] {
    const yearsSet = new SvelteSet<number>()
    const currentY = new Date().getFullYear()
    yearsSet.add(currentY)
    yearsSet.add(currentY + 1)
    for (const evt of $calendarStore.events) {
      const y = parseInt(evt.date.slice(0, 4), 10)
      if (!isNaN(y)) yearsSet.add(y)
    }
    return Array.from(yearsSet).sort((a, b) => a - b)
  }
  let agendaYears = $derived.by(computeAgendaYears)

  /**
   * Computes agenda events filtered by the active year selection.
   */
  function computeFilteredAgendaEvents(): CalendarEvent[] {
    if (agendaYearFilter === 'all') {
      return allAgendaEvents
    }
    return allAgendaEvents.filter((e) =>
      e.date.startsWith(`${agendaYearFilter}-`)
    )
  }
  let filteredAgendaEvents = $derived.by(computeFilteredAgendaEvents)

  /**
   * Finds the upcoming event scheduled for today or subsequent dates.
   */
  function computeNextUpcomingEvent(): CalendarEvent | null {
    const todayKey = formatDateKey(new Date())
    const future = $calendarStore.events
      .filter((e) => e.date >= todayKey)
      .sort((a, b) => {
        const cmpDate = a.date.localeCompare(b.date)
        if (cmpDate !== 0) return cmpDate
        return (a.startTime || '').localeCompare(b.startTime || '')
      })
    return future[0] || null
  }
  let nextUpcomingEvent = $derived.by(computeNextUpcomingEvent)

  /**
   * Formats the active date label into a human-readable string.
   */
  function computeSelectedDateLabel(): string {
    const todayKey = formatDateKey(new Date())
    if ($calendarStore.selectedDate === todayKey) {
      return 'Today'
    }
    const [y, m, d] = $calendarStore.selectedDate.split('-').map(Number)
    const target = new Date(y, m - 1, d)
    return target.toLocaleDateString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric'
    })
  }
  let selectedDateLabel = $derived.by(computeSelectedDateLabel)

  /**
   * Evaluates if the current calendar view matches the real current month.
   */
  function computeIsCurrentMonth(): boolean {
    const now = new Date()
    return (
      viewYear === now.getFullYear() &&
      viewMonth === now.getMonth() &&
      calViewMode === 'month'
    )
  }
  let isCurrentMonth = $derived.by(computeIsCurrentMonth)

  /**
   * Formats an arbitrary date key into a compact agenda group header.
   */
  function formatAgendaDateHeader(dateStr: string): string {
    const todayKey = formatDateKey(new Date())
    if (dateStr === todayKey) return 'Today'
    const tomorrow = new Date(Date.now() + 86400000)
    if (dateStr === formatDateKey(tomorrow)) return 'Tomorrow'

    const [y, m, d] = dateStr.split('-').map(Number)
    const dt = new Date(y, m - 1, d)
    return dt.toLocaleDateString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric'
    })
  }

  /**
   * Retrieves events occurring on a specified day number in the current view month.
   *
   * @param dayNum Day index (1-31).
   */
  function getEventsForDay(dayNum: number): CalendarEvent[] {
    const dateStr = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(dayNum).padStart(2, '0')}`
    return $calendarStore.events.filter((e) => e.date === dateStr)
  }

  /**
   * Advances backward one period based on current calendar view mode.
   */
  function handlePrevPeriod() {
    if (calViewMode === 'month') {
      viewDate = new Date(viewYear, viewMonth - 1, 1)
    } else if (calViewMode === 'year') {
      viewDate = new Date(viewYear - 1, viewMonth, 1)
    } else {
      viewDate = new Date(viewYear - 10, viewMonth, 1)
    }
  }

  /**
   * Advances forward one period based on current calendar view mode.
   */
  function handleNextPeriod() {
    if (calViewMode === 'month') {
      viewDate = new Date(viewYear, viewMonth + 1, 1)
    } else if (calViewMode === 'year') {
      viewDate = new Date(viewYear + 1, viewMonth, 1)
    } else {
      viewDate = new Date(viewYear + 10, viewMonth, 1)
    }
  }

  /**
   * Cycles view mode between month, year, and decade.
   */
  function handleHeaderClick() {
    if (calViewMode === 'month') {
      calViewMode = 'year'
    } else if (calViewMode === 'year') {
      calViewMode = 'decade'
    } else {
      calViewMode = 'month'
    }
  }

  /**
   * Selects a month from the year picker.
   */
  function selectMonth(mIndex: number) {
    viewDate = new Date(viewYear, mIndex, 1)
    calViewMode = 'month'
  }

  /**
   * Selects a year from the decade picker.
   */
  function selectYear(yr: number) {
    viewDate = new Date(yr, viewMonth, 1)
    calViewMode = 'year'
  }

  /**
   * Jumps the calendar view and selection to today.
   */
  function handleJumpToday() {
    const now = new Date()
    viewDate = now
    calViewMode = 'month'
    setSelectedCalendarDate(formatDateKey(now))
  }

  /**
   * Selects an individual day cell in the monthly grid.
   *
   * @param dayNum Selected day of the month.
   */
  function handleSelectDay(dayNum: number) {
    const dateStr = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(dayNum).padStart(2, '0')}`
    setSelectedCalendarDate(dateStr)
  }

  /**
   * Opens the event creation modal prefilled for the active date.
   *
   * @param targetDate Optional target date string.
   */
  function handleOpenAddModal(targetDate?: string) {
    editingEventId = null
    modalMode = 'create'
    formTitle = ''
    formDate = targetDate || $calendarStore.selectedDate
    formStartTime = '09:00'
    formEndTime = '10:00'
    formIsAllDay = false
    formColor = getRandomEventColor()
    formLocation = ''
    formDescription = ''
    isModalOpen = true
  }

  /**
   * Opens the event modal in edit mode prefilled with existing event data.
   *
   * @param event Calendar event to view or edit.
   */
  function handleOpenEditModal(event: CalendarEvent) {
    editingEventId = event.id
    modalMode = 'edit'
    formTitle = event.title
    formDate = event.date
    formStartTime = event.startTime || '09:00'
    formEndTime = event.endTime || '10:00'
    formIsAllDay = Boolean(event.isAllDay)
    formColor = event.color
    formLocation = event.location || ''
    formDescription = event.description || ''
    isModalOpen = true
  }

  /**
   * Persists a new or edited local calendar event.
   */
  function handleSaveEvent() {
    const cleanTitle = formTitle.trim().slice(0, 60)
    if (!cleanTitle) return

    const cleanDesc = formDescription.trim().slice(0, 200) || undefined
    const cleanLoc = formLocation.trim().slice(0, 80) || undefined

    if (modalMode === 'edit' && editingEventId) {
      updateCalendarEvent(editingEventId, {
        title: cleanTitle,
        description: cleanDesc,
        date: formDate,
        startTime: formIsAllDay ? '' : formStartTime,
        endTime: formIsAllDay ? '' : formEndTime,
        isAllDay: formIsAllDay,
        color: formColor,
        location: cleanLoc
      })
    } else {
      addCalendarEvent({
        title: cleanTitle,
        description: cleanDesc,
        date: formDate,
        startTime: formIsAllDay ? '' : formStartTime,
        endTime: formIsAllDay ? '' : formEndTime,
        isAllDay: formIsAllDay,
        color: formColor,
        location: cleanLoc,
        source: 'local'
      })
    }
    isModalOpen = false
    editingEventId = null
  }

  /**
   * Deletes the currently edited event from the schedule.
   */
  function handleDeleteEditingEvent() {
    if (editingEventId) {
      deleteCalendarEvent(editingEventId)
      isModalOpen = false
      editingEventId = null
    }
  }

  /**
   * Switches from the calendar view directly to the agenda view for the selected date.
   *
   * @param dateStr Target date key.
   */
  function handleJumpToAgenda(dateStr: string) {
    activeTab = 'agenda'
    const y = parseInt(dateStr.slice(0, 4), 10)
    if (!isNaN(y)) agendaYearFilter = y
  }
</script>

<div class="calendar-page-container">
  {#if !expanded}
    <!-- Minimal Mode Capsule -->
    <div class="minimal-calendar-pill">
      <div class="pill-date-badge">
        <CalendarIcon size={13} />
        <span
          >{new Date().toLocaleDateString(undefined, {
            month: 'short',
            day: 'numeric'
          })}</span
        >
      </div>

      <div class="pill-event-marquee">
        {#if nextUpcomingEvent}
          <span class="pill-event-dot {nextUpcomingEvent.color}"></span>
          {#if nextUpcomingEvent.startTime}
            <span class="pill-time-tag">{nextUpcomingEvent.startTime}</span>
          {/if}
          <span class="pill-title-tag">{nextUpcomingEvent.title}</span>
        {:else}
          <span class="pill-time-tag">No events today</span>
        {/if}
      </div>
    </div>
  {:else}
    <!-- Expanded Mode Header with Segmented Calendar & Agenda Tabs -->
    <div class="calendar-header">
      <div class="header-left">
        <BackButton title="Back to Apps" />
      </div>

      <!-- Segmented View Tabs (Calendar vs Agenda) -->
      <HorizontalTabSelection
        tabs={CALENDAR_TABS}
        bind:value={activeTab}
        ariaLabel="Calendar views"
      />

      <div class="header-right"></div>
    </div>

    <!-- Scrollable Body Content (Clear of bottom dock) -->
    <ScrollView class="calendar-scroll-body">
      {#if activeTab === 'calendar'}
        <!-- CALENDAR TAB: Mirrored Home Calendar card layout & compact grid -->
        <div class="calendar-tab-content">
          <div class="calendar-card-surface">
            <!-- Header bar inside the card matching Home widget -->
            <div class="card-month-nav">
              <div class="stepper-title-group">
                <button
                  class="icon-nav-btn"
                  onclick={handlePrevPeriod}
                  aria-label="Previous period"
                >
                  <ChevronLeft size={13} />
                </button>

                <button
                  class="month-title-btn"
                  onclick={handleHeaderClick}
                  aria-label="Change view scale"
                >
                  <span>{headerTitle}</span>
                  <ChevronDown size={11} class="title-chevron" />
                </button>

                <button
                  class="icon-nav-btn"
                  onclick={handleNextPeriod}
                  aria-label="Next period"
                >
                  <ChevronRight size={13} />
                </button>
              </div>

              {#if !isCurrentMonth || calViewMode !== 'month'}
                <button class="today-chip-btn" onclick={handleJumpToday}
                  >Today</button
                >
              {/if}
            </div>

            {#if calViewMode === 'month'}
              <!-- Weekdays Row -->
              <div class="weekdays-row">
                {#each weekDays as wd (wd)}
                  <div class="weekday-cell">{wd}</div>
                {/each}
              </div>

              <!-- Days Grid matching compact 25px cell sizing from Home -->
              <div class="days-grid">
                {#each blanks as _, i (i)}
                  <div class="day-cell empty" aria-hidden="true"></div>
                {/each}

                {#each daysArray as dayNum (dayNum)}
                  {@const today = new Date()}
                  {@const isToday =
                    viewYear === today.getFullYear() &&
                    viewMonth === today.getMonth() &&
                    dayNum === today.getDate()}
                  {@const dateStr = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(dayNum).padStart(2, '0')}`}
                  {@const isSelected = $calendarStore.selectedDate === dateStr}
                  {@const dayEvts = getEventsForDay(dayNum)}

                  <button
                    class="day-cell"
                    class:today={isToday}
                    class:selected={isSelected}
                    onclick={() => handleSelectDay(dayNum)}
                  >
                    <span class="day-num">{dayNum}</span>
                    {#if dayEvts.length > 0}
                      <div class="day-dot-indicator-tray">
                        {#each dayEvts.slice(0, 3) as evt (evt.id)}
                          <span class="mini-day-dot {evt.color}"></span>
                        {/each}
                      </div>
                    {/if}
                  </button>
                {/each}
              </div>

              <!-- Day-Level Agenda Tray directly inside the card -->
              <div class="calendar-agenda-tray">
                <div class="agenda-tray-header">
                  <div class="agenda-tray-title">
                    <Clock size={12} />
                    <span>{selectedDateLabel}</span>
                    <span class="agenda-tray-subtitle">
                      • {selectedDayEvents.length}
                      {selectedDayEvents.length === 1 ? 'event' : 'events'}
                    </span>
                  </div>
                  <div class="agenda-tray-actions">
                    <button
                      type="button"
                      class="tray-agenda-jump-btn"
                      onclick={() =>
                        handleJumpToAgenda($calendarStore.selectedDate)}
                    >
                      <span>View in Agenda</span>
                    </button>
                    <button
                      type="button"
                      class="tray-add-btn"
                      onclick={() =>
                        handleOpenAddModal($calendarStore.selectedDate)}
                    >
                      <Plus size={13} />
                      <span>Add Event</span>
                    </button>
                  </div>
                </div>

                {#if selectedDayEvents.length === 0}
                  <div class="agenda-tray-empty">
                    <span>No events scheduled for this date</span>
                  </div>
                {:else}
                  <div class="agenda-tray-list">
                    {#each selectedDayEvents as event (event.id)}
                      <div class="tray-event-card card-color-{event.color}">
                        <div class="tray-card-main">
                          <span class="tray-time-chip">
                            {#if event.isAllDay}
                              All Day
                            {:else if event.startTime}
                              {event.startTime}{event.endTime
                                ? ` - ${event.endTime}`
                                : ''}
                            {:else}
                              Event
                            {/if}
                          </span>
                          <span class="tray-card-title">{event.title}</span>
                          {#if event.location}
                            <span class="tray-card-loc">
                              <MapPin size={10} />
                              {event.location}
                            </span>
                          {/if}
                        </div>
                        <div class="tray-card-actions">
                          <button
                            type="button"
                            class="tray-card-action-btn edit"
                            onclick={() => handleOpenEditModal(event)}
                            aria-label="Edit event"
                            title="Edit event"
                          >
                            <Pencil size={11} />
                          </button>
                          <button
                            type="button"
                            class="tray-card-action-btn danger"
                            onclick={() => deleteCalendarEvent(event.id)}
                            aria-label="Delete event"
                            title="Delete event"
                          >
                            <Trash2 size={11} />
                          </button>
                        </div>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {:else if calViewMode === 'year'}
              <!-- 12 Months Grid -->
              <div class="cal-months-grid">
                {#each monthShortNames as mName, idx (mName)}
                  {@const today = new Date()}
                  {@const isCurrentM =
                    viewYear === today.getFullYear() &&
                    idx === today.getMonth()}
                  {@const isSelectedM = idx === viewMonth}
                  <button
                    class="cal-picker-cell"
                    class:today={isCurrentM}
                    class:selected={isSelectedM}
                    onclick={() => selectMonth(idx)}
                  >
                    {mName}
                  </button>
                {/each}
              </div>
            {:else}
              <!-- 12 Years Decade Grid -->
              <div class="cal-years-grid">
                {#each decadeYears as yr (yr)}
                  {@const today = new Date()}
                  {@const isCurrentY = yr === today.getFullYear()}
                  {@const isSelectedY = yr === viewYear}
                  <button
                    class="cal-picker-cell"
                    class:today={isCurrentY}
                    class:selected={isSelectedY}
                    onclick={() => selectYear(yr)}
                  >
                    {yr}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      {:else}
        <!-- AGENDA TAB: Dedicated Chronological Stream with Clean Card Layout -->
        <div class="agenda-tab-content">
          <!-- Unified Agenda Controls: Year Filters aligned with Add Event button -->
          <div class="agenda-toolbar-row">
            <div class="agenda-year-filter-row">
              <button
                class="agenda-year-pill"
                class:active={agendaYearFilter === 'all'}
                onclick={() => (agendaYearFilter = 'all')}
              >
                All
              </button>
              {#each agendaYears as yr (yr)}
                <button
                  class="agenda-year-pill"
                  class:active={agendaYearFilter === yr}
                  onclick={() => (agendaYearFilter = yr)}
                >
                  {yr}
                </button>
              {/each}
            </div>

            <button
              class="agenda-primary-add-btn"
              onclick={() => handleOpenAddModal(formatDateKey(new Date()))}
            >
              <Plus size={13} />
              <span>Add Event</span>
            </button>
          </div>

          {#if filteredAgendaEvents.length === 0}
            <div class="agenda-empty-state">
              <div class="empty-icon-circle">
                <CalendarIcon size={24} />
              </div>
              <span class="empty-primary-text">No Scheduled Events</span>
              <p class="empty-secondary-text">
                No events scheduled for {agendaYearFilter === 'all'
                  ? 'the selected period'
                  : agendaYearFilter}. Click "Add Event" to plan an activity.
              </p>
            </div>
          {:else}
            <div class="agenda-stream">
              {#each filteredAgendaEvents as event (event.id)}
                <div class="agenda-card card-color-{event.color}">
                  <div class="agenda-card-main">
                    <div class="agenda-card-top">
                      <div class="agenda-card-heading">
                        <span class="agenda-card-title">{event.title}</span>
                        <span class="agenda-date-chip"
                          >{formatAgendaDateHeader(event.date)}</span
                        >
                      </div>
                      <div class="agenda-card-actions">
                        {#if event.locationUrl}
                          <a
                            class="action-icon-link"
                            href={event.locationUrl}
                            target="_blank"
                            rel="noreferrer"
                            use:islandTooltip={'Open link'}
                          >
                            <ExternalLink size={13} />
                          </a>
                        {/if}
                        {#if event.source === 'local'}
                          <button
                            type="button"
                            class="action-icon-btn edit"
                            onclick={() => handleOpenEditModal(event)}
                            use:islandTooltip={'Edit event'}
                            aria-label="Edit event"
                          >
                            <Pencil size={13} />
                          </button>
                          <button
                            type="button"
                            class="action-icon-btn danger"
                            onclick={() => deleteCalendarEvent(event.id)}
                            use:islandTooltip={'Delete event'}
                            aria-label="Delete event"
                          >
                            <Trash2 size={13} />
                          </button>
                        {/if}
                      </div>
                    </div>

                    <div class="agenda-card-meta">
                      <span class="event-time-badge">
                        <Clock size={11} />
                        {#if event.isAllDay}
                          All Day
                        {:else if event.startTime}
                          {event.startTime}{event.endTime
                            ? ` - ${event.endTime}`
                            : ''}
                        {:else}
                          Scheduled
                        {/if}
                      </span>
                      {#if event.location}
                        <span class="agenda-loc-text">
                          <MapPin size={11} />
                          {event.location}
                        </span>
                      {/if}
                    </div>

                    {#if event.description}
                      <div class="agenda-card-desc">{event.description}</div>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </ScrollView>

    <!-- Modal: Add or Edit Event with Custom Date & Time Steppers -->
    {#if isModalOpen}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="modal-overlay" onclick={() => (isModalOpen = false)}>
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="modal-card" onclick={(e) => e.stopPropagation()}>
          <div class="modal-header">
            <span class="modal-title">
              {modalMode === 'edit' ? 'Edit Event' : 'New Event'}
            </span>
            <div class="modal-header-actions">
              {#if modalMode === 'edit'}
                <button
                  type="button"
                  class="modal-agenda-jump-btn"
                  onclick={() => {
                    handleJumpToAgenda(formDate)
                    isModalOpen = false
                  }}
                  title="View this event in Agenda"
                >
                  <ExternalLink size={12} />
                  <span>View in Agenda</span>
                </button>
              {/if}
              <button
                class="modal-close-btn"
                onclick={() => (isModalOpen = false)}
                aria-label="Close dialog"
              >
                <X size={14} />
              </button>
            </div>
          </div>

          <ScrollView class="modal-form-scroll">
            <div class="modal-form">
              <div class="form-group">
                <div class="form-label-row">
                  <label class="form-label" for="cal-event-title">Title</label>
                  <span class="char-count">{formTitle.length}/60</span>
                </div>
                <input
                  id="cal-event-title"
                  class="form-input"
                  type="text"
                  placeholder="Meeting or event title"
                  maxlength="60"
                  bind:value={formTitle}
                />
              </div>

              <div class="form-group">
                <span class="form-label">Date</span>
                <CustomDatePicker
                  bind:value={formDate}
                  ariaLabel="Event date"
                />
              </div>

              {#if !formIsAllDay}
                <div class="form-row time-pickers-row">
                  <div class="form-group time-field">
                    <span class="form-label">Start Time</span>
                    <CustomTimePicker
                      bind:value={formStartTime}
                      ariaLabel="Start time"
                      align="left"
                    />
                  </div>
                  <div class="form-group time-field">
                    <span class="form-label">End Time</span>
                    <CustomTimePicker
                      bind:value={formEndTime}
                      ariaLabel="End time"
                      align="right"
                    />
                  </div>
                </div>
              {/if}

              <div class="form-group">
                <span class="form-label">Category Color</span>
                <div class="color-pills-row">
                  {#each ['indigo', 'emerald', 'amber', 'purple', 'rose', 'sky'] as color (color)}
                    <button
                      class="color-pill-btn {color}"
                      class:active={formColor === color}
                      onclick={() =>
                        (formColor = color as CalendarEvent['color'])}
                      type="button"
                      aria-label="{color} category"
                    ></button>
                  {/each}
                </div>
              </div>

              <div class="form-group">
                <div class="form-label-row">
                  <label class="form-label" for="cal-event-loc"
                    >Location (Optional)</label
                  >
                  <span class="char-count">{formLocation.length}/80</span>
                </div>
                <input
                  id="cal-event-loc"
                  class="form-input"
                  type="text"
                  placeholder="Room, office, or meeting link"
                  maxlength="80"
                  bind:value={formLocation}
                />
              </div>

              <div class="form-group">
                <div class="form-label-row">
                  <label class="form-label" for="cal-event-desc"
                    >Description (Optional)</label
                  >
                  <span class="char-count">{formDescription.length}/200</span>
                </div>
                <textarea
                  id="cal-event-desc"
                  class="form-input form-textarea"
                  placeholder="Additional notes"
                  maxlength="200"
                  rows="2"
                  bind:value={formDescription}></textarea>
              </div>
            </div>
          </ScrollView>

          <div class="modal-footer">
            {#if modalMode === 'edit'}
              <button
                type="button"
                class="modal-btn-delete"
                onclick={handleDeleteEditingEvent}
                aria-label="Delete event"
              >
                <Trash2 size={13} />
                <span>Delete</span>
              </button>
            {:else}
              <div></div>
            {/if}

            <div class="modal-footer-right">
              <button
                type="button"
                class="modal-btn-cancel"
                onclick={() => (isModalOpen = false)}
              >
                Cancel
              </button>
              <button
                type="button"
                class="modal-btn-submit"
                onclick={handleSaveEvent}
                disabled={!formTitle.trim()}
              >
                {modalMode === 'edit' ? 'Save Changes' : 'Create Event'}
              </button>
            </div>
          </div>
        </div>
      </div>
    {/if}
  {/if}
</div>

<style src="./styles/CalendarPage.css"></style>
