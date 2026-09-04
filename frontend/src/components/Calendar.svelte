<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { ChevronLeft, ChevronRight } from '@lucide/svelte'

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
    padding: 0.8rem 1rem;
    box-sizing: border-box;
    color: var(--c-text-primary, #ffffff);
  }

  /* GRID SECTION */
  .grid-section {
    flex: 1;
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    min-width: 0;
  }

  .month-nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 6px;
  }

  .header-title-btn {
    font-size: 0.95rem;
    font-weight: 700;
    letter-spacing: -0.01em;
    background: transparent;
    border: none;
    color: var(--c-text-primary, #ffffff);
    cursor: pointer;
    padding: 2px 6px;
    border-radius: 6px;
    transition: background 0.2s ease;
  }

  .header-title-btn:hover {
    background: var(--c-hover);
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
    font-size: 0.68rem;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 10px;
    background: var(--c-hover);
    color: var(--c-primary);
    border: 1px solid var(--c-border);
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .today-btn:hover {
    background: var(--c-item-bg);
  }

  .icon-nav-btn {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    padding: 0;
    transition: background 0.2s ease;
  }

  .icon-nav-btn:hover {
    background: var(--c-hover);
  }

  /* WEEKDAYS */
  .weekdays-row {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    text-align: center;
    margin-bottom: 4px;
  }

  .weekday-cell {
    font-size: 0.68rem;
    font-weight: 600;
    color: var(--c-text-muted);
    text-transform: uppercase;
  }

  /* DAYS GRID */
  .days-grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    grid-auto-rows: 1fr;
    gap: 3px;
    flex: 1;
  }

  .day-cell {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    background: transparent;
    border: none;
    color: var(--c-text-primary);
    font-size: 0.8rem;
    font-weight: 500;
    cursor: pointer;
    padding: 0;
    transition:
      background 0.2s ease,
      transform 0.15s ease;
  }

  .day-cell.empty {
    cursor: default;
    pointer-events: none;
  }

  .day-cell:hover:not(.empty) {
    background: var(--c-hover);
  }

  .day-cell.today {
    color: var(--c-success);
    font-weight: 700;
  }

  .day-cell.selected {
    background: var(--c-primary) !important;
    color: #ffffff !important;
    font-weight: 700;
    box-shadow: var(--c-calendar-today-glow);
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
    border-radius: 10px;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    font-size: 0.82rem;
    font-weight: 600;
    cursor: pointer;
    transition:
      background 0.2s ease,
      border-color 0.2s ease;
  }

  .grid-picker-cell:hover {
    background: var(--c-hover);
    border-color: var(--c-border);
  }

  .grid-picker-cell.today {
    color: var(--c-success);
    border-color: var(--c-success);
  }

  .grid-picker-cell.selected {
    background: var(--c-primary) !important;
    color: #ffffff !important;
    border-color: transparent;
  }
</style>
