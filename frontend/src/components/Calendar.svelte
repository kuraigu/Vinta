<script lang="ts">
  import { onMount, onDestroy, getContext } from 'svelte'
  import {
    ChevronLeft,
    ChevronRight,
    ChevronDown,
    CalendarDays,
    ArrowRight,
    Clock,
    Plus,
    Pencil,
    Trash2,
    MapPin,
    ListFilter,
    X,
    Calendar as CalendarIcon
  } from 'lucide-svelte'
  import {
    calendarStore,
    setSelectedCalendarDate,
    addCalendarEvent,
    updateCalendarEvent,
    deleteCalendarEvent,
    formatDateKey,
    type CalendarEvent
  } from '../stores/calendarStore'
  import CalendarPage, {
    config as calendarConfig
  } from '../pages/CalendarPage.svelte'
  import BackButton from './BackButton.svelte'
  import ScrollView from './ScrollView.svelte'
  import CustomDatePicker from './CustomDatePicker.svelte'
  import CustomTimePicker from './CustomTimePicker.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from './HorizontalTabSelection.svelte'
  import {
    timeFormatStore,
    dateFormatStore,
    formatUserDate,
    type IslandConfig
  } from '../store'
  import { islandTooltip } from '../services/islandTooltip'

  let {
    expanded = false,
    variant = 'widget',
    showGoToCalendar = false,
    onOpenCalendar
  }: {
    expanded?: boolean
    variant?: 'widget' | 'page'
    showGoToCalendar?: boolean
    onOpenCalendar?: () => void
  } = $props()

  const navigate = getContext<
    | ((page: unknown, pageConfig: IslandConfig, expand?: boolean) => void)
    | undefined
  >('navigate')

  let realToday = $state(new Date())
  let viewDate = $state(new Date())
  let viewMode = $state<'month' | 'year' | 'decade'>('month')
  let activeTab = $state<'calendar' | 'agenda'>('calendar')
  let agendaFilter = $state<'all' | 'today' | 'upcoming'>('all')

  // Event modal state
  let isModalOpen = $state(false)
  let modalMode = $state<'create' | 'edit'>('create')
  let editingEventId = $state<string | null>(null)
  let formTitle = $state('')
  let formDate = $state(formatDateKey(new Date()))
  let formStartTime = $state('09:00')
  let formEndTime = $state('10:00')
  let formIsAllDay = $state(false)
  let formColor = $state<CalendarEvent['color']>('indigo')
  let formLocation = $state('')
  let formDescription = $state('')

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

  const AGENDA_FILTER_TABS: TabItem<'all' | 'today' | 'upcoming'>[] = [
    { id: 'all', label: 'All', tooltip: 'Show all scheduled events' },
    { id: 'today', label: 'Today', tooltip: "Show today's events" },
    { id: 'upcoming', label: 'Upcoming', tooltip: 'Show upcoming events' }
  ]

  let timer: ReturnType<typeof setInterval>

  onMount(() => {
    timer = setInterval(() => {
      realToday = new Date()
    }, 60000)
  })

  onDestroy(() => {
    clearInterval(timer)
  })

  function computeViewYear(): number {
    return viewDate.getFullYear()
  }
  let viewYear = $derived.by(computeViewYear)

  function computeViewMonth(): number {
    return viewDate.getMonth()
  }
  let viewMonth = $derived.by(computeViewMonth)

  function computeIsCurrentMonth(): boolean {
    return (
      viewYear === realToday.getFullYear() &&
      viewMonth === realToday.getMonth() &&
      viewMode === 'month'
    )
  }
  let isCurrentMonth = $derived.by(computeIsCurrentMonth)

  const monthNames = [
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

  function computeHeaderTitle(): string {
    if (viewMode === 'month') {
      return viewDate.toLocaleDateString(undefined, {
        month: 'long',
        year: 'numeric'
      })
    }
    if (viewMode === 'year') {
      return viewYear.toString()
    }
    const startYear = Math.floor(viewYear / 10) * 10
    return `${startYear} - ${startYear + 11}`
  }
  let headerTitle = $derived.by(computeHeaderTitle)

  function computeFirstDayOfWeek(): number {
    return new Date(viewYear, viewMonth, 1).getDay()
  }
  let firstDayOfWeek = $derived.by(computeFirstDayOfWeek)

  function computeTotalDaysInMonth(): number {
    return new Date(viewYear, viewMonth + 1, 0).getDate()
  }
  let totalDaysInMonth = $derived.by(computeTotalDaysInMonth)

  function computeBlanks(): null[] {
    return Array(firstDayOfWeek).fill(null)
  }
  let blanks = $derived.by(computeBlanks)

  function computeDaysArray(): number[] {
    return Array.from({ length: totalDaysInMonth }, (_, i) => i + 1)
  }
  let daysArray = $derived.by(computeDaysArray)

  function computeDecadeYears(): number[] {
    const startYear = Math.floor(viewYear / 10) * 10 - 1
    return Array.from({ length: 12 }, (_, i) => startYear + i)
  }
  let decadeYears = $derived.by(computeDecadeYears)

  function computeSelectedDateEvents(): CalendarEvent[] {
    return $calendarStore.events.filter(
      (e) => e.date === $calendarStore.selectedDate
    )
  }
  let selectedDateEvents = $derived.by(computeSelectedDateEvents)

  function computeSelectedDateLabel(): string {
    return formatUserDate($calendarStore.selectedDate, $dateFormatStore)
  }
  let selectedDateLabel = $derived.by(computeSelectedDateLabel)

  function computeAllAgendaEvents(): CalendarEvent[] {
    return [...$calendarStore.events].sort((a, b) => {
      const cmpDate = a.date.localeCompare(b.date)
      if (cmpDate !== 0) return cmpDate
      return (a.startTime || '').localeCompare(b.startTime || '')
    })
  }
  let allAgendaEvents = $derived.by(computeAllAgendaEvents)

  function computeFilteredAgendaEvents(): CalendarEvent[] {
    const todayKey = formatDateKey(new Date())
    if (agendaFilter === 'today') {
      return allAgendaEvents.filter((e) => e.date === todayKey)
    }
    if (agendaFilter === 'upcoming') {
      return allAgendaEvents.filter((e) => e.date >= todayKey)
    }
    return allAgendaEvents
  }
  let filteredAgendaEvents = $derived.by(computeFilteredAgendaEvents)

  function computeNextUpcomingEvent(): CalendarEvent | null {
    const todayKey = formatDateKey(new Date())
    const evts = $calendarStore.events
      .filter((e) => e.date >= todayKey)
      .sort((a, b) => {
        const c = a.date.localeCompare(b.date)
        if (c !== 0) return c
        return (a.startTime || '').localeCompare(b.startTime || '')
      })
    return evts[0] || null
  }
  let nextUpcomingEvent = $derived.by(computeNextUpcomingEvent)

  function formatEventTime(timeStr?: string): string {
    if (!timeStr) return ''
    if ($timeFormatStore === '24h') return timeStr
    const parts = timeStr.split(':')
    if (parts.length < 2) return timeStr
    let h = parseInt(parts[0], 10)
    const m = parts[1]
    const ampm = h >= 12 ? 'PM' : 'AM'
    h = h % 12
    if (h === 0) h = 12
    return `${h}:${m} ${ampm}`
  }

  function handlePrev() {
    if (viewMode === 'month') {
      viewDate = new Date(viewYear, viewMonth - 1, 1)
    } else if (viewMode === 'year') {
      viewDate = new Date(viewYear - 1, viewMonth, 1)
    } else {
      viewDate = new Date(viewYear - 10, viewMonth, 1)
    }
  }

  function handleNext() {
    if (viewMode === 'month') {
      viewDate = new Date(viewYear, viewMonth + 1, 1)
    } else if (viewMode === 'year') {
      viewDate = new Date(viewYear + 1, viewMonth, 1)
    } else {
      viewDate = new Date(viewYear + 10, viewMonth, 1)
    }
  }

  function cycleViewMode() {
    if (viewMode === 'month') viewMode = 'year'
    else if (viewMode === 'year') viewMode = 'decade'
    else viewMode = 'month'
  }

  function selectMonth(idx: number) {
    viewDate = new Date(viewYear, idx, 1)
    viewMode = 'month'
  }

  function selectYear(yr: number) {
    viewDate = new Date(yr, viewMonth, 1)
    viewMode = 'year'
  }

  function jumpToToday() {
    const now = new Date()
    viewDate = now
    viewMode = 'month'
    setSelectedCalendarDate(formatDateKey(now))
  }

  function selectDate(day: number) {
    const dateStr = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    setSelectedCalendarDate(dateStr)
  }

  function handleOpenCalendarPage() {
    if (onOpenCalendar) {
      onOpenCalendar()
    } else if (navigate) {
      navigate(CalendarPage, calendarConfig, true)
    }
  }

  function handleOpenAddModal(targetDate?: string) {
    editingEventId = null
    modalMode = 'create'
    formTitle = ''
    formDate = targetDate || $calendarStore.selectedDate
    formStartTime = '09:00'
    formEndTime = '10:00'
    formIsAllDay = false
    formColor = 'indigo'
    formLocation = ''
    formDescription = ''
    isModalOpen = true
  }

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

  function handleDeleteEditingEvent() {
    if (editingEventId) {
      deleteCalendarEvent(editingEventId)
      isModalOpen = false
      editingEventId = null
    }
  }
</script>

<div
  class="calendar-container"
  class:expanded
  class:minimal={!expanded}
  class:page-mode={variant === 'page'}
>
  {#if !expanded}
    <!-- MINIMAL VIEWS -->
    {#if variant === 'page'}
      <div class="minimal-calendar-pill">
        <div class="pill-date-badge">
          <CalendarIcon size={13} />
          <span>{formatUserDate(new Date(), $dateFormatStore)}</span>
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
      <span class="minimal-date">
        {formatUserDate(realToday, $dateFormatStore)}
      </span>
    {/if}
  {:else}
    <!-- EXPANDED VIEWS -->
    {#if variant === 'page'}
      <!-- Page View Header: BackButton + View Tabs (Calendar vs Agenda) -->
      <div class="calendar-page-header">
        <div class="page-header-left">
          <BackButton />
        </div>
        <div class="page-header-tabs">
          <HorizontalTabSelection
            tabs={CALENDAR_TABS}
            bind:value={activeTab}
            ariaLabel="Calendar views"
          />
        </div>
        <div class="page-header-spacer"></div>
      </div>
    {/if}

    {#if variant === 'page'}
      <ScrollView class="calendar-scroll-view">
        <div class="calendar-card variant-page">
          {#if activeTab === 'agenda'}
            <!-- FULL AGENDA STREAM TAB -->
            <div class="agenda-tab-view">
              <div class="agenda-toolbar-row">
                <HorizontalTabSelection
                  tabs={AGENDA_FILTER_TABS}
                  bind:value={agendaFilter}
                  colorVariant="primary"
                  ariaLabel="Filter agenda events"
                />

                <button
                  type="button"
                  class="agenda-primary-add-btn"
                  onclick={() => handleOpenAddModal()}
                >
                  <Plus size={13} />
                  <span>Add Event</span>
                </button>
              </div>

              {#if filteredAgendaEvents.length === 0}
                <div class="agenda-empty-state">
                  <CalendarDays size={26} />
                  <span>No events match this filter</span>
                </div>
              {:else}
                <div class="agenda-stream-list">
                  {#each filteredAgendaEvents as event (event.id)}
                    <div class="agenda-card card-color-{event.color}">
                      <div class="agenda-card-left">
                        <div class="agenda-card-title-row">
                          <span class="agenda-card-title">{event.title}</span>
                          <span class="agenda-card-date-badge"
                            >{formatUserDate(
                              event.date,
                              $dateFormatStore
                            )}</span
                          >
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

                      <div class="tray-card-actions">
                        <button
                          type="button"
                          class="tray-card-action-btn"
                          onclick={() => handleOpenEditModal(event)}
                          aria-label="Edit event"
                          use:islandTooltip={'Edit event'}
                        >
                          <Pencil size={12} />
                        </button>
                        <button
                          type="button"
                          class="tray-card-action-btn danger"
                          onclick={() => deleteCalendarEvent(event.id)}
                          aria-label="Delete event"
                          use:islandTooltip={'Delete event'}
                        >
                          <Trash2 size={12} />
                        </button>
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {:else}
            {@render monthGridView()}
          {/if}
        </div>
      </ScrollView>
    {:else}
      <div class="calendar-card widget-mode">
        {@render monthGridView()}
      </div>
    {/if}

    {#snippet monthGridView()}
      <!-- MONTH GRID VIEW (Shared by Widget and Page Calendar Tab) -->
      <div class="grid-section">
        <!-- Navigation Header -->
        <div class="month-nav">
          <div class="month-nav-left">
            <div class="stepper-title-group">
              <button
                type="button"
                class="header-title-btn"
                onclick={cycleViewMode}
              >
                <span>{headerTitle}</span>
                <ChevronDown
                  size={11}
                  style="display:inline; margin-left:2px; vertical-align:middle;"
                />
              </button>

              <button
                type="button"
                class="icon-nav-btn"
                onclick={handlePrev}
                aria-label="Previous interval"
                use:islandTooltip={'Previous'}
              >
                <ChevronLeft size={13} />
              </button>
              <button
                type="button"
                class="icon-nav-btn"
                onclick={handleNext}
                aria-label="Next interval"
                use:islandTooltip={'Next'}
              >
                <ChevronRight size={13} />
              </button>
            </div>
          </div>

          <div class="month-nav-right">
            <button
              type="button"
              class="today-btn"
              onclick={jumpToToday}
              use:islandTooltip={'Jump to current day'}
            >
              Today
            </button>

            {#if showGoToCalendar}
              <button
                type="button"
                class="go-to-calendar-btn"
                onclick={handleOpenCalendarPage}
                use:islandTooltip={'Open full Calendar page'}
              >
                <CalendarDays size={12} />
                <span>Go to Calendar</span>
                <ArrowRight size={11} />
              </button>
            {/if}
          </div>
        </div>

        {#if viewMode === 'month'}
          <!-- Weekdays Header -->
          <div class="weekdays-row">
            {#each weekDays as wd (wd)}
              <div class="weekday-cell">{wd}</div>
            {/each}
          </div>

          <!-- Days Grid -->
          <div class="days-grid">
            {#each blanks as blank, i (i)}
              <div class="day-cell empty">{blank}</div>
            {/each}
            {#each daysArray as day (day)}
              {@const isToday = isCurrentMonth && day === realToday.getDate()}
              {@const dayKey = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`}
              {@const isSelected = dayKey === $calendarStore.selectedDate}
              {@const dayEvts = $calendarStore.events.filter(
                (e) => e.date === dayKey
              )}
              <button
                type="button"
                class="day-cell"
                class:today={isToday}
                class:selected={isSelected}
                onclick={() => selectDate(day)}
              >
                <span class="day-num">{day}</span>
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

          <!-- Selected Day Agenda Tray -->
          <div class="calendar-agenda-tray">
            <div class="agenda-tray-header">
              <div class="agenda-tray-title">
                <Clock size={11} />
                <span>{selectedDateLabel}</span>
                <span class="agenda-tray-subtitle">
                  • {selectedDateEvents.length}
                  {selectedDateEvents.length === 1 ? 'event' : 'events'}
                </span>
              </div>
              <div class="agenda-tray-actions">
                {#if variant === 'page'}
                  <button
                    type="button"
                    class="tray-agenda-jump-btn"
                    onclick={() => (activeTab = 'agenda')}
                  >
                    <span>View in Agenda</span>
                  </button>
                {/if}
                <button
                  type="button"
                  class="tray-add-btn"
                  onclick={() =>
                    handleOpenAddModal($calendarStore.selectedDate)}
                >
                  <Plus size={12} />
                  <span>Add Event</span>
                </button>
              </div>
            </div>

            {#if selectedDateEvents.length === 0}
              <div class="agenda-tray-empty">
                <span>No events scheduled for this date</span>
              </div>
            {:else}
              <div class="agenda-tray-list">
                {#each selectedDateEvents as event (event.id)}
                  <div class="tray-event-card card-color-{event.color}">
                    <div class="tray-card-main">
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
                      <span class="tray-card-title">{event.title}</span>
                      {#if event.location}
                        <span class="tray-card-loc">
                          <MapPin size={9} />
                          {event.location}
                        </span>
                      {/if}
                    </div>
                    <div class="tray-card-actions">
                      <button
                        type="button"
                        class="tray-card-action-btn"
                        onclick={() => handleOpenEditModal(event)}
                        aria-label="Edit event"
                        use:islandTooltip={'Edit event'}
                      >
                        <Pencil size={11} />
                      </button>
                      <button
                        type="button"
                        class="tray-card-action-btn danger"
                        onclick={() => deleteCalendarEvent(event.id)}
                        aria-label="Delete event"
                        use:islandTooltip={'Delete event'}
                      >
                        <Trash2 size={11} />
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {:else if viewMode === 'year'}
          <!-- 12 Months Grid -->
          <div class="months-grid">
            {#each monthNames as mName, idx (mName)}
              {@const isCurrentM =
                viewYear === realToday.getFullYear() &&
                idx === realToday.getMonth()}
              {@const isSelectedM = idx === viewMonth}
              <button
                type="button"
                class="grid-picker-cell"
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
          <div class="years-grid">
            {#each decadeYears as yr (yr)}
              {@const isCurrentY = yr === realToday.getFullYear()}
              {@const isSelectedY = yr === viewYear}
              <button
                type="button"
                class="grid-picker-cell"
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
    {/snippet}

    <!-- EVENT CREATION / EDITING MODAL (Self-contained in Calendar component) -->
    {#if isModalOpen}
      <div
        class="modal-backdrop"
        onclick={() => (isModalOpen = false)}
        onkeydown={(e) => {
          if (e.key === 'Escape') isModalOpen = false
        }}
        role="presentation"
      >
        <div
          class="modal-dialog"
          onclick={(e) => e.stopPropagation()}
          onkeydown={(e) => e.stopPropagation()}
          role="dialog"
          tabindex="-1"
          aria-modal="true"
        >
          <div class="modal-header">
            <span class="modal-title">
              {modalMode === 'edit' ? 'Edit Event' : 'Add New Event'}
            </span>
            <button
              type="button"
              class="modal-close-btn"
              onclick={() => (isModalOpen = false)}
              aria-label="Close modal"
            >
              <X size={15} />
            </button>
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

<style src="./styles/Calendar.css"></style>
