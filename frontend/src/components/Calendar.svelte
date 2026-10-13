<script lang="ts">
  import { onMount, onDestroy, getContext } from 'svelte'
  import {
    ChevronLeft,
    ChevronRight,
    CalendarDays,
    ArrowRight,
    Clock
  } from 'lucide-svelte'
  import { calendarStore } from '../stores/calendarStore'
  import CalendarPage, {
    config as calendarConfig
  } from '../pages/CalendarPage.svelte'
  import type { IslandConfig } from '../store'

  let {
    expanded = false,
    onOpenCalendar
  }: {
    expanded?: boolean
    onOpenCalendar?: () => void
  } = $props()

  const navigate = getContext<
    ((page: unknown, pageConfig: IslandConfig) => void) | undefined
  >('navigate')

  let realToday = $state(new Date())
  let viewDate = $state(new Date())
  let selectedDay = $state(new Date().getDate())
  let viewMode = $state<'month' | 'year' | 'decade'>('month')

  let timer: ReturnType<typeof setInterval>

  onMount(() => {
    timer = setInterval(() => {
      realToday = new Date()
    }, 60000)
  })

  onDestroy(() => {
    clearInterval(timer)
  })

  /**
   * Computes the current view year.
   */
  function computeViewYear(): number {
    return viewDate.getFullYear()
  }
  let viewYear = $derived.by(computeViewYear)

  /**
   * Computes the current view month index.
   */
  function computeViewMonth(): number {
    return viewDate.getMonth()
  }
  let viewMonth = $derived.by(computeViewMonth)

  /**
   * Evaluates if the current calendar view matches real today.
   */
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

  /**
   * Computes the interactive header navigation title.
   */
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

  /**
   * Computes the day offset for the first day of the current month.
   */
  function computeFirstDayOfWeek(): number {
    return new Date(viewYear, viewMonth, 1).getDay()
  }
  let firstDayOfWeek = $derived.by(computeFirstDayOfWeek)

  /**
   * Computes total days in the active view month.
   */
  function computeTotalDaysInMonth(): number {
    return new Date(viewYear, viewMonth + 1, 0).getDate()
  }
  let totalDaysInMonth = $derived.by(computeTotalDaysInMonth)

  /**
   * Computes empty blank cells preceding the first day of month.
   */
  function computeBlanks(): null[] {
    return Array(firstDayOfWeek).fill(null)
  }
  let blanks = $derived.by(computeBlanks)

  /**
   * Computes the array of days for the month.
   */
  function computeDaysArray(): number[] {
    return Array.from({ length: totalDaysInMonth }, (_, i) => i + 1)
  }
  let daysArray = $derived.by(computeDaysArray)

  /**
   * Computes the 12-year window for decade picker.
   */
  function computeDecadeYears(): number[] {
    const startYear = Math.floor(viewYear / 10) * 10 - 1
    return Array.from({ length: 12 }, (_, i) => startYear + i)
  }
  let decadeYears = $derived.by(computeDecadeYears)

  /**
   * Computes the date key string for the selected day cell.
   */
  function computeSelectedDateKey(): string {
    return `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(selectedDay).padStart(2, '0')}`
  }
  let selectedDateKey = $derived.by(computeSelectedDateKey)

  /**
   * Filters events matching the selected date.
   */
  function computeSelectedDateEvents() {
    return $calendarStore.events.filter((e) => e.date === selectedDateKey)
  }
  let selectedDateEvents = $derived.by(computeSelectedDateEvents)

  /**
   * Formats a human-readable title for the selected date agenda.
   */
  function computeSelectedDateLabel(): string {
    const today = new Date()
    if (
      viewYear === today.getFullYear() &&
      viewMonth === today.getMonth() &&
      selectedDay === today.getDate()
    ) {
      return 'Today'
    }
    const target = new Date(viewYear, viewMonth, selectedDay)
    return target.toLocaleDateString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric'
    })
  }
  let selectedDateLabel = $derived.by(computeSelectedDateLabel)

  /**
   * Computes minimal date string for minimal capsule.
   */
  function computeMinimalDate(): string {
    return realToday.toLocaleDateString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric'
    })
  }
  let minimalDate = $derived.by(computeMinimalDate)

  const weekDays = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

  /**
   * Advances backward one period based on current viewMode.
   */
  function prevPeriod() {
    if (viewMode === 'month') {
      viewDate = new Date(viewYear, viewMonth - 1, 1)
    } else if (viewMode === 'year') {
      viewDate = new Date(viewYear - 1, viewMonth, 1)
    } else {
      viewDate = new Date(viewYear - 10, viewMonth, 1)
    }
  }

  /**
   * Advances forward one period based on current viewMode.
   */
  function nextPeriod() {
    if (viewMode === 'month') {
      viewDate = new Date(viewYear, viewMonth + 1, 1)
    } else if (viewMode === 'year') {
      viewDate = new Date(viewYear + 1, viewMonth, 1)
    } else {
      viewDate = new Date(viewYear + 10, viewMonth, 1)
    }
  }

  /**
   * Cycles view modes on header button click.
   */
  function handleHeaderClick() {
    if (viewMode === 'month') {
      viewMode = 'year'
    } else if (viewMode === 'year') {
      viewMode = 'decade'
    } else {
      viewMode = 'month'
    }
  }

  /**
   * Selects month from year picker.
   */
  function selectMonth(mIndex: number) {
    viewDate = new Date(viewYear, mIndex, 1)
    viewMode = 'month'
  }

  /**
   * Selects year from decade picker.
   */
  function selectYear(yr: number) {
    viewDate = new Date(yr, viewMonth, 1)
    viewMode = 'year'
  }

  /**
   * Resets calendar view to today.
   */
  function resetToToday() {
    viewDate = new Date()
    selectedDay = realToday.getDate()
    viewMode = 'month'
  }

  /**
   * Selects day in month grid.
   */
  function selectDate(d: number) {
    selectedDay = d
  }

  /**
   * Navigates to the main Calendar page.
   */
  function handleGoToCalendar() {
    if (onOpenCalendar) {
      onOpenCalendar()
    } else if (navigate) {
      navigate(CalendarPage, calendarConfig)
    }
  }
</script>

<div class="calendar-container {expanded ? 'expanded' : 'minimal'}">
  {#if !expanded}
    <div class="minimal-date">{minimalDate}</div>
  {:else}
    <div class="calendar-card">
      <div class="grid-section">
        <!-- Interactive Header Bar: Single-line month title with symmetrical stepper -->
        <div class="month-nav">
          <div class="month-nav-left">
            <div class="stepper-title-group">
              <button
                class="icon-nav-btn"
                onclick={prevPeriod}
                title="Previous"
                aria-label="Previous"
              >
                <ChevronLeft size={13} />
              </button>

              <button
                class="header-title-btn"
                onclick={handleHeaderClick}
                title="Click to switch view mode"
              >
                {headerTitle}
              </button>

              <button
                class="icon-nav-btn"
                onclick={nextPeriod}
                title="Next"
                aria-label="Next"
              >
                <ChevronRight size={13} />
              </button>
            </div>
          </div>

          <div class="month-nav-right">
            {#if !isCurrentMonth}
              <button
                class="today-btn"
                onclick={resetToToday}
                title="Jump to Today">Today</button
              >
            {/if}

            <button
              class="go-to-calendar-btn"
              onclick={handleGoToCalendar}
              title="Open Calendar Page"
            >
              <CalendarDays size={12} />
              <span>Go to Calendar</span>
              <ArrowRight size={11} />
            </button>
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
              {@const isSelected = day === selectedDay}
              {@const dayKey = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`}
              {@const dayEvts = $calendarStore.events.filter(
                (e) => e.date === dayKey
              )}
              <button
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

          <!-- Selected Day Agenda Tray (Subtle unified bento surface) -->
          <div class="home-calendar-agenda">
            <div class="home-agenda-header">
              <div class="home-agenda-title">
                <Clock size={11} />
                <span>{selectedDateLabel}</span>
              </div>
              {#if selectedDateEvents.length > 0}
                <span class="home-agenda-count"
                  >{selectedDateEvents.length}</span
                >
              {/if}
            </div>

            {#if selectedDateEvents.length === 0}
              <div class="home-agenda-empty">
                <span>No events scheduled</span>
              </div>
            {:else}
              <div class="home-agenda-list">
                {#each selectedDateEvents.slice(0, 3) as event (event.id)}
                  <div class="home-agenda-item">
                    <span class="item-color-dot {event.color}"></span>
                    <span class="item-time">
                      {#if event.isAllDay}
                        All Day
                      {:else if event.startTime}
                        {event.startTime}
                      {:else}
                        Event
                      {/if}
                    </span>
                    <span class="item-title">{event.title}</span>
                  </div>
                {/each}
                {#if selectedDateEvents.length > 3}
                  <div class="home-agenda-overflow">
                    +{selectedDateEvents.length - 3} more
                  </div>
                {/if}
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
    </div>
  {/if}
</div>

<style>
  /* MINIMAL */
  .minimal .minimal-date {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--c-text-muted, rgba(255, 255, 255, 0.7));
  }

  /* EXPANDED */
  .calendar-container.expanded {
    width: 100%;
    height: 100%;
    box-sizing: border-box;
  }

  .calendar-card {
    display: flex;
    width: 100%;
    height: 100%;
    padding: 10px 14px;
    box-sizing: border-box;
    color: var(--c-text-primary, #ffffff);
  }

  /* GRID SECTION */
  .grid-section {
    flex: 1;
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: flex-start;
    gap: 6px;
    min-width: 0;
  }

  .month-nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 2px;
    gap: 8px;
    flex-wrap: nowrap;
  }

  .month-nav-left {
    display: flex;
    align-items: center;
    min-width: 0;
    flex-shrink: 0;
  }

  .stepper-title-group {
    display: flex;
    align-items: center;
    gap: 2px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.08));
    border-radius: 9px;
    padding: 2px 4px;
  }

  .header-title-btn {
    font-size: 0.8rem;
    font-weight: 700;
    letter-spacing: -0.01em;
    background: transparent;
    border: none;
    color: var(--c-text-primary, #ffffff);
    cursor: pointer;
    padding: 2px 6px;
    border-radius: 6px;
    white-space: nowrap;
    transition:
      background 0.2s ease,
      color 0.2s ease;
  }

  .header-title-btn:hover {
    background: rgba(255, 255, 255, 0.08);
  }

  .icon-nav-btn {
    width: 20px;
    height: 20px;
    border-radius: 5px;
    background: transparent;
    border: none;
    color: var(--c-text-secondary, #a1a1aa);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    padding: 0;
    transition: all 0.15s ease;
  }

  .icon-nav-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--c-text-primary, #ffffff);
  }

  .month-nav-right {
    display: flex;
    align-items: center;
    gap: 5px;
    flex-shrink: 0;
  }

  .today-btn {
    font-size: 0.65rem;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 9999px;
    background: rgba(56, 189, 248, 0.12);
    color: var(--c-primary, #38bdf8);
    border: 1px solid rgba(56, 189, 248, 0.28);
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.2s ease;
  }

  .today-btn:hover {
    background: rgba(56, 189, 248, 0.2);
    border-color: rgba(56, 189, 248, 0.45);
  }

  .go-to-calendar-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.68rem;
    font-weight: 600;
    padding: 3px 8px;
    border-radius: 8px;
    background: var(--c-item-active, rgba(255, 255, 255, 0.1));
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.14));
    color: var(--c-text-primary, #ffffff);
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.15s ease;
  }

  .go-to-calendar-btn:hover {
    background: var(--c-hover, rgba(255, 255, 255, 0.18));
    border-color: rgba(255, 255, 255, 0.25);
    transform: translateY(-0.5px);
  }

  /* WEEKDAYS */
  .weekdays-row {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    text-align: center;
    margin-bottom: 1px;
  }

  .weekday-cell {
    font-size: 0.65rem;
    font-weight: 600;
    color: var(--c-text-muted, #71717a);
    text-transform: uppercase;
    letter-spacing: 0.02em;
  }

  /* DAYS GRID */
  .days-grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 2px;
  }

  .day-cell {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 25px;
    border-radius: 7px;
    background: transparent;
    border: none;
    color: var(--c-text-primary);
    font-size: 0.74rem;
    font-weight: 500;
    cursor: pointer;
    padding: 2px 0;
    transition:
      background 0.15s ease,
      color 0.15s ease;
  }

  .day-cell.empty {
    cursor: default;
    pointer-events: none;
  }

  .day-cell:hover:not(.empty) {
    background: rgba(255, 255, 255, 0.08);
  }

  .day-cell.today {
    color: var(--c-primary, #38bdf8);
    font-weight: 700;
  }

  .day-cell.selected {
    background: var(--c-item-active, rgba(255, 255, 255, 0.16)) !important;
    color: #ffffff !important;
    font-weight: 700;
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.22));
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.25);
  }

  .day-dot-indicator-tray {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 2px;
    margin-top: 1px;
    height: 3px;
  }

  .mini-day-dot {
    width: 3px;
    height: 3px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .mini-day-dot.indigo {
    background: #818cf8;
  }
  .mini-day-dot.emerald {
    background: #34d399;
  }
  .mini-day-dot.amber {
    background: #fbbf24;
  }
  .mini-day-dot.purple {
    background: #c084fc;
  }
  .mini-day-dot.rose {
    background: #fb7185;
  }
  .mini-day-dot.sky {
    background: #38bdf8;
  }

  /* HOME CALENDAR AGENDA TRAY (Clean subtle bento surface without divider line) */
  .home-calendar-agenda {
    margin-top: 4px;
    padding: 6px 10px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.035);
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.06));
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .home-agenda-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 0.68rem;
    color: var(--c-text-secondary, #a1a1aa);
  }

  .home-agenda-title {
    display: flex;
    align-items: center;
    gap: 4px;
    font-weight: 600;
    color: var(--c-text-primary, #ffffff);
  }

  .home-agenda-count {
    padding: 1px 5px;
    border-radius: 8px;
    background: rgba(56, 189, 248, 0.12);
    color: var(--c-primary, #38bdf8);
    font-size: 0.62rem;
    font-weight: 700;
  }

  .home-agenda-empty {
    font-size: 0.68rem;
    color: var(--c-text-muted, #71717a);
    padding: 2px 0 1px;
  }

  .home-agenda-list {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .home-agenda-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 3px 6px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.04);
    font-size: 0.68rem;
  }

  .item-color-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .item-color-dot.indigo {
    background: #818cf8;
  }
  .item-color-dot.emerald {
    background: #34d399;
  }
  .item-color-dot.amber {
    background: #fbbf24;
  }
  .item-color-dot.purple {
    background: #c084fc;
  }
  .item-color-dot.rose {
    background: #fb7185;
  }
  .item-color-dot.sky {
    background: #38bdf8;
  }

  .item-time {
    font-weight: 600;
    color: var(--c-text-muted, #a1a1aa);
    flex-shrink: 0;
  }

  .item-title {
    color: var(--c-text-primary, #ffffff);
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .home-agenda-overflow {
    font-size: 0.62rem;
    color: var(--c-text-muted, #71717a);
    padding-left: 6px;
  }

  /* MONTHS & YEARS PICKER GRIDS */
  .months-grid,
  .years-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    grid-template-rows: repeat(3, 1fr);
    gap: 6px;
    flex: 1;
    padding-top: 4px;
    min-height: 120px;
  }

  .grid-picker-cell {
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.07);
    color: var(--c-text-primary);
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .grid-picker-cell:hover {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.14);
  }

  .grid-picker-cell.today {
    color: var(--c-primary, #38bdf8);
    border-color: rgba(56, 189, 248, 0.4);
  }

  .grid-picker-cell.selected {
    background: var(--c-item-active, rgba(255, 255, 255, 0.18)) !important;
    color: #ffffff !important;
    border-color: var(--c-border, rgba(255, 255, 255, 0.25));
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
  }
</style>
