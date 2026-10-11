<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { ChevronLeft, ChevronRight } from 'lucide-svelte'

  let { expanded = false } = $props<{ expanded?: boolean }>()

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

  let viewYear = $derived(viewDate.getFullYear())
  let viewMonth = $derived(viewDate.getMonth())

  let isCurrentMonth = $derived(
    viewYear === realToday.getFullYear() &&
      viewMonth === realToday.getMonth() &&
      viewMode === 'month'
  )

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

  let headerTitle = $derived.by(() => {
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
  })

  let firstDayOfWeek = $derived(new Date(viewYear, viewMonth, 1).getDay())
  let totalDaysInMonth = $derived(
    new Date(viewYear, viewMonth + 1, 0).getDate()
  )

  let blanks = $derived(Array(firstDayOfWeek).fill(null))
  let daysArray = $derived(
    Array.from({ length: totalDaysInMonth }, (_, i) => i + 1)
  )

  let decadeYears = $derived.by(() => {
    const startYear = Math.floor(viewYear / 10) * 10 - 1
    return Array.from({ length: 12 }, (_, i) => startYear + i)
  })

  const weekDays = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

  function prevPeriod() {
    if (viewMode === 'month') {
      viewDate = new Date(viewYear, viewMonth - 1, 1)
    } else if (viewMode === 'year') {
      viewDate = new Date(viewYear - 1, viewMonth, 1)
    } else {
      viewDate = new Date(viewYear - 10, viewMonth, 1)
    }
  }

  function nextPeriod() {
    if (viewMode === 'month') {
      viewDate = new Date(viewYear, viewMonth + 1, 1)
    } else if (viewMode === 'year') {
      viewDate = new Date(viewYear + 1, viewMonth, 1)
    } else {
      viewDate = new Date(viewYear + 10, viewMonth, 1)
    }
  }

  function handleHeaderClick() {
    if (viewMode === 'month') {
      viewMode = 'year'
    } else if (viewMode === 'year') {
      viewMode = 'decade'
    } else {
      viewMode = 'month'
    }
  }

  function selectMonth(mIndex: number) {
    viewDate = new Date(viewYear, mIndex, 1)
    viewMode = 'month'
  }

  function selectYear(yr: number) {
    viewDate = new Date(yr, viewMonth, 1)
    viewMode = 'year'
  }

  function resetToToday() {
    viewDate = new Date()
    selectedDay = realToday.getDate()
    viewMode = 'month'
  }

  function selectDate(d: number) {
    selectedDay = d
  }

  let minimalDate = $derived(
    realToday.toLocaleDateString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric'
    })
  )
</script>

<div class="calendar-container {expanded ? 'expanded' : 'minimal'}">
  {#if !expanded}
    <div class="minimal-date">{minimalDate}</div>
  {:else}
    <div class="calendar-card">
      <div class="grid-section">
        <!-- Windows Style Interactive Header -->
        <div class="month-nav">
          <button
            class="header-title-btn"
            onclick={handleHeaderClick}
            title="Click to switch view mode"
          >
            {headerTitle}
          </button>

          <div class="nav-actions">
            {#if !isCurrentMonth}
              <button
                class="today-btn"
                onclick={resetToToday}
                title="Jump to Today">Jump to Today</button
              >
            {/if}
            <div class="stepper-group">
              <button
                class="icon-nav-btn"
                onclick={prevPeriod}
                title="Previous"
                aria-label="Previous"
              >
                <ChevronLeft size={16} />
              </button>
              <button
                class="icon-nav-btn"
                onclick={nextPeriod}
                title="Next"
                aria-label="Next"
              >
                <ChevronRight size={16} />
              </button>
            </div>
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
              <button
                class="day-cell"
                class:today={isToday}
                class:selected={isSelected}
                onclick={() => selectDate(day)}
              >
                <span class="day-num">{day}</span>
              </button>
            {/each}
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
    gap: 4px;
    min-width: 0;
  }

  .month-nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 2px;
  }

  .header-title-btn {
    font-size: 0.88rem;
    font-weight: 700;
    letter-spacing: -0.01em;
    background: transparent;
    border: none;
    color: var(--c-text-primary, #ffffff);
    cursor: pointer;
    padding: 2px 8px;
    border-radius: 8px;
    transition:
      background 0.2s ease,
      color 0.2s ease;
  }

  .header-title-btn:hover {
    background: rgba(255, 255, 255, 0.08);
  }

  .nav-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .stepper-group {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .today-btn {
    font-size: 0.66rem;
    font-weight: 600;
    padding: 2px 9px;
    border-radius: 9999px;
    background: rgba(56, 189, 248, 0.12);
    color: var(--c-primary, #38bdf8);
    border: 1px solid rgba(56, 189, 248, 0.28);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .today-btn:hover {
    background: rgba(56, 189, 248, 0.2);
    border-color: rgba(56, 189, 248, 0.45);
  }

  .icon-nav-btn {
    width: 24px;
    height: 24px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: var(--c-text-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    padding: 0;
    transition: all 0.2s ease;
  }

  .icon-nav-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    border-color: rgba(255, 255, 255, 0.16);
  }

  /* WEEKDAYS */
  .weekdays-row {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    text-align: center;
    margin-bottom: 2px;
  }

  .weekday-cell {
    font-size: 0.65rem;
    font-weight: 600;
    color: var(--c-text-muted);
    opacity: 0.75;
    text-transform: uppercase;
    letter-spacing: 0.02em;
  }

  /* DAYS GRID */
  .days-grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    grid-auto-rows: 24px;
    gap: 2px;
    flex: 1;
  }

  .day-cell {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    margin: 0 auto;
    border-radius: 7px;
    background: transparent;
    border: none;
    color: var(--c-text-primary);
    font-size: 0.75rem;
    font-weight: 500;
    cursor: pointer;
    padding: 0;
    transition:
      background 0.2s ease,
      transform 0.15s ease,
      color 0.2s ease;
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

  .day-cell.today::after {
    content: '';
    position: absolute;
    bottom: 2px;
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: var(--c-primary, #38bdf8);
  }

  .day-cell.selected {
    background: linear-gradient(135deg, #38bdf8 0%, #0284c7 100%) !important;
    color: #ffffff !important;
    font-weight: 700;
    box-shadow: 0 2px 8px rgba(56, 189, 248, 0.4);
  }

  .day-cell.selected::after {
    display: none;
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
  }

  .grid-picker-cell {
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.07);
    color: var(--c-text-primary);
    font-size: 0.82rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
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
    background: linear-gradient(135deg, #38bdf8 0%, #0284c7 100%) !important;
    color: #ffffff !important;
    border-color: transparent;
    box-shadow: 0 2px 8px rgba(56, 189, 248, 0.35);
  }
</style>
