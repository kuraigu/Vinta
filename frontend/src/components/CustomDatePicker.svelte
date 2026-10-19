<svelte:options />

<script lang="ts">
  import { onMount } from 'svelte'
  import {
    Calendar as CalendarIcon,
    ChevronLeft,
    ChevronRight,
    ChevronDown
  } from 'lucide-svelte'
  import { dateFormatStore, formatUserDate } from '../store'

  let {
    value = $bindable(''),
    ariaLabel = 'Select date',
    align = 'left'
  }: {
    value?: string
    ariaLabel?: string
    align?: 'left' | 'right'
  } = $props()

  let isOpen = $state(false)
  let containerRef = $state<HTMLDivElement | null>(null)

  // Current view mode: 'calendar' (days), 'months' (12-month grid), 'decade' (12-year decade grid)
  let viewMode = $state<'calendar' | 'months' | 'decade'>('calendar')

  // Internal view year and month for calendar navigation
  let viewYear = $state(new Date().getFullYear())
  let viewMonth = $state(new Date().getMonth())

  // Direct year jump input state
  let jumpYearInput = $state('')

  const MONTH_NAMES = [
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
  const MONTH_SHORT = [
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
  const WEEK_DAYS = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

  /**
   * Synchronizes internal view parameters with the bound value string.
   */
  function syncViewFromValue() {
    if (value && /^\d{4}-\d{2}-\d{2}$/.test(value)) {
      const [y, m] = value.split('-').map(Number)
      viewYear = y
      viewMonth = m - 1
    } else {
      const now = new Date()
      viewYear = now.getFullYear()
      viewMonth = now.getMonth()
    }
    jumpYearInput = String(viewYear)
  }

  $effect(() => {
    syncViewFromValue()
  })

  /**
   * Computes the formatted readable label for the trigger button.
   */
  function computeDisplayLabel(): string {
    if (!value || !/^\d{4}-\d{2}-\d{2}$/.test(value)) {
      return 'Pick a date'
    }
    const [y, m, d] = value.split('-').map(Number)
    const target = new Date(y, m - 1, d)
    return formatUserDate(target, $dateFormatStore, { includeWeekday: true })
  }
  let displayLabel = $derived.by(computeDisplayLabel)

  /**
   * Calculates the days and leading blank offsets for the active month view.
   */
  function computeCalendarDays() {
    const firstDayIndex = new Date(viewYear, viewMonth, 1).getDay()
    const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate()
    const blanks = Array.from({ length: firstDayIndex })
    const days = Array.from({ length: daysInMonth }, (_, i) => i + 1)
    return { blanks, days }
  }
  let calendarGrid = $derived.by(computeCalendarDays)

  /**
   * Computes the 12-year window for the active decade view.
   */
  function computeDecadeYears(): number[] {
    const startYear = Math.floor(viewYear / 10) * 10 - 1
    return Array.from({ length: 12 }, (_, i) => startYear + i)
  }
  let decadeYears = $derived.by(computeDecadeYears)

  /**
   * Computes the header title text for the active decade.
   */
  function computeDecadeTitle(): string {
    const startYear = Math.floor(viewYear / 10) * 10
    return `${startYear} - ${startYear + 9}`
  }
  let decadeTitle = $derived.by(computeDecadeTitle)

  /**
   * Navigates to the preceding month, year, or decade.
   */
  function handlePrev(e: MouseEvent): void {
    e.stopPropagation()
    if (viewMode === 'decade') {
      viewYear -= 10
      jumpYearInput = String(viewYear)
    } else if (viewMode === 'months') {
      viewYear -= 1
      jumpYearInput = String(viewYear)
    } else {
      if (viewMonth === 0) {
        viewMonth = 11
        viewYear -= 1
        jumpYearInput = String(viewYear)
      } else {
        viewMonth -= 1
      }
    }
  }

  /**
   * Navigates to the subsequent month, year, or decade.
   */
  function handleNext(e: MouseEvent): void {
    e.stopPropagation()
    if (viewMode === 'decade') {
      viewYear += 10
      jumpYearInput = String(viewYear)
    } else if (viewMode === 'months') {
      viewYear += 1
      jumpYearInput = String(viewYear)
    } else {
      if (viewMonth === 11) {
        viewMonth = 0
        viewYear += 1
        jumpYearInput = String(viewYear)
      } else {
        viewMonth += 1
      }
    }
  }

  /**
   * Selects a specific day in the active month view.
   */
  function handleSelectDay(day: number): void {
    const mStr = String(viewMonth + 1).padStart(2, '0')
    const dStr = String(day).padStart(2, '0')
    value = `${viewYear}-${mStr}-${dStr}`
    isOpen = false
    viewMode = 'calendar'
  }

  /**
   * Selects a year and navigates to the months grid.
   */
  function handleSelectYear(year: number): void {
    viewYear = year
    jumpYearInput = String(year)
    viewMode = 'months'
  }

  /**
   * Selects a month and navigates to the days grid.
   */
  function handleSelectMonth(monthIndex: number): void {
    viewMonth = monthIndex
    viewMode = 'calendar'
  }

  /**
   * Applies direct manual year input from the jump input box.
   */
  function handleJumpYearSubmit(e?: Event): void {
    if (e) e.preventDefault()
    const parsed = parseInt(jumpYearInput, 10)
    if (!isNaN(parsed) && parsed >= 1900 && parsed <= 2100) {
      viewYear = parsed
      viewMode = 'months'
    }
  }

  /**
   * Sets the selected date to today.
   */
  function handleJumpToday(e: MouseEvent): void {
    e.stopPropagation()
    const today = new Date()
    viewYear = today.getFullYear()
    viewMonth = today.getMonth()
    jumpYearInput = String(viewYear)
    const mStr = String(today.getMonth() + 1).padStart(2, '0')
    const dStr = String(today.getDate()).padStart(2, '0')
    value = `${today.getFullYear()}-${mStr}-${dStr}`
    isOpen = false
    viewMode = 'calendar'
  }

  /**
   * Toggles popover visibility.
   */
  function toggleOpen(): void {
    isOpen = !isOpen
    if (isOpen) {
      viewMode = 'calendar'
      syncViewFromValue()
    }
  }

  onMount(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        isOpen &&
        containerRef &&
        !containerRef.contains(event.target as Node)
      ) {
        isOpen = false
        viewMode = 'calendar'
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  })
</script>

<div class="custom-date-picker" bind:this={containerRef}>
  <button
    type="button"
    class="date-trigger-btn"
    class:open={isOpen}
    onclick={toggleOpen}
    aria-label={ariaLabel}
    aria-haspopup="dialog"
    aria-expanded={isOpen}
  >
    <div class="trigger-label-group">
      <CalendarIcon size={13} class="trigger-icon" />
      <span class="trigger-text">{displayLabel}</span>
    </div>
    <ChevronDown size={13} class={`chevron-icon ${isOpen ? 'rotate' : ''}`} />
  </button>

  {#if isOpen}
    <div
      class="date-popover"
      class:align-right={align === 'right'}
      role="dialog"
      aria-modal="true"
    >
      <!-- Industry Standard Navigation Header -->
      <div class="popover-header">
        <button
          type="button"
          class="popover-nav-btn"
          onclick={handlePrev}
          aria-label="Previous"
        >
          <ChevronLeft size={13} />
        </button>

        <!-- Dynamic Header Switchers based on View Mode -->
        <div class="header-titles-group">
          {#if viewMode === 'calendar'}
            <button
              type="button"
              class="header-toggle-btn"
              onclick={() => (viewMode = 'months')}
              aria-label="Change month"
            >
              <span>{MONTH_NAMES[viewMonth]}</span>
              <ChevronDown size={10} class="small-chevron" />
            </button>

            <button
              type="button"
              class="header-toggle-btn"
              onclick={() => (viewMode = 'decade')}
              aria-label="Change year"
            >
              <span>{viewYear}</span>
              <ChevronDown size={10} class="small-chevron" />
            </button>
          {:else if viewMode === 'months'}
            <button
              type="button"
              class="header-toggle-btn active"
              onclick={() => (viewMode = 'decade')}
              aria-label="Change decade"
            >
              <span>{viewYear}</span>
              <ChevronDown size={10} class="small-chevron" />
            </button>
          {:else}
            <span class="decade-header-label">{decadeTitle}</span>
          {/if}
        </div>

        <button
          type="button"
          class="popover-nav-btn"
          onclick={handleNext}
          aria-label="Next"
        >
          <ChevronRight size={13} />
        </button>
      </div>

      <!-- View Mode 1: Calendar Days View -->
      {#if viewMode === 'calendar'}
        <div class="popover-weekdays">
          {#each WEEK_DAYS as wd (wd)}
            <span class="popover-weekday">{wd}</span>
          {/each}
        </div>

        <div class="popover-days-grid">
          {#each calendarGrid.blanks as _, idx (idx)}
            <div class="popover-day-cell blank" aria-hidden="true"></div>
          {/each}

          {#each calendarGrid.days as day (day)}
            {@const today = new Date()}
            {@const isToday =
              viewYear === today.getFullYear() &&
              viewMonth === today.getMonth() &&
              day === today.getDate()}
            {@const dateStr = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`}
            {@const isSelected = value === dateStr}

            <button
              type="button"
              class="popover-day-cell"
              class:today={isToday}
              class:selected={isSelected}
              onclick={() => handleSelectDay(day)}
            >
              <span>{day}</span>
            </button>
          {/each}
        </div>

        <!-- View Mode 2: 12 Months Grid -->
      {:else if viewMode === 'months'}
        <div class="months-selector-grid">
          {#each MONTH_SHORT as mName, idx (mName)}
            <button
              type="button"
              class="month-item-btn"
              class:active={viewMonth === idx}
              onclick={() => handleSelectMonth(idx)}
            >
              {mName}
            </button>
          {/each}
        </div>

        <!-- View Mode 3: 12-Year Decade Grid with Fast Year Jump -->
      {:else if viewMode === 'decade'}
        <!-- Fast Year Jump input for far years -->
        <form class="decade-jump-form" onsubmit={handleJumpYearSubmit}>
          <input
            type="number"
            class="decade-jump-input"
            placeholder="Go to year..."
            min="1900"
            max="2100"
            bind:value={jumpYearInput}
          />
          <button type="submit" class="decade-jump-btn">Go</button>
        </form>

        <div class="decade-selector-grid">
          {#each decadeYears as yr (yr)}
            {@const isSelected = viewYear === yr}
            {@const startYear = Math.floor(viewYear / 10) * 10}
            {@const isOuter = yr < startYear || yr > startYear + 9}
            <button
              type="button"
              class="decade-item-btn"
              class:active={isSelected}
              class:outer={isOuter}
              onclick={() => handleSelectYear(yr)}
            >
              {yr}
            </button>
          {/each}
        </div>
      {/if}

      <!-- Popover Footer -->
      <div class="popover-footer">
        <button
          type="button"
          class="popover-today-btn"
          onclick={handleJumpToday}
        >
          Today
        </button>
        {#if viewMode !== 'calendar'}
          <button
            type="button"
            class="popover-back-btn"
            onclick={() => (viewMode = 'calendar')}
          >
            Back to Calendar
          </button>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .custom-date-picker {
    position: relative;
    width: 100%;
    user-select: none;
  }

  .date-trigger-btn {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    height: 32px;
    padding: 0 10px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.1));
    border-radius: 9px;
    color: var(--vinta-text-primary, #ffffff);
    font-size: 11.5px;
    font-weight: 500;
    cursor: pointer;
    outline: none;
    transition: all 0.16s ease;
    box-sizing: border-box;
  }

  .date-trigger-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.18);
  }

  .date-trigger-btn.open {
    border-color: var(--vinta-primary, #38bdf8);
    background: rgba(255, 255, 255, 0.09);
    box-shadow: 0 0 0 2px rgba(56, 189, 248, 0.2);
  }

  .trigger-label-group {
    display: flex;
    align-items: center;
    gap: 7px;
    min-width: 0;
  }

  :global(.trigger-icon) {
    color: var(--vinta-primary, #38bdf8);
    flex-shrink: 0;
  }

  .trigger-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :global(.chevron-icon) {
    color: var(--vinta-text-secondary, #a1a1aa);
    transition: transform 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
    flex-shrink: 0;
  }

  :global(.chevron-icon.rotate) {
    transform: rotate(180deg);
  }

  /* Popover Floating Card */
  .date-popover {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    width: 256px;
    z-index: 250;
    background: #181920;
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 14px;
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.14),
      0 12px 32px rgba(0, 0, 0, 0.65);
    padding: 10px 12px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 7px;
    animation: popover-fade 0.15s ease-out;
  }

  .date-popover.align-right {
    left: auto;
    right: 0;
  }

  @keyframes popover-fade {
    from {
      opacity: 0;
      transform: translateY(-4px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  /* Popover Header */
  .popover-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 2px;
  }

  .header-titles-group {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .header-toggle-btn {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 2.5px 7px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.08));
    color: var(--vinta-text-primary, #ffffff);
    font-size: 11px;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .header-toggle-btn:hover {
    background: rgba(255, 255, 255, 0.12);
  }

  .header-toggle-btn.active {
    background: rgba(56, 189, 248, 0.18);
    border-color: rgba(56, 189, 248, 0.35);
    color: var(--vinta-primary, #38bdf8);
  }

  .decade-header-label {
    font-size: 11.5px;
    font-weight: 700;
    color: var(--vinta-text-primary, #ffffff);
    letter-spacing: -0.01em;
  }

  :global(.small-chevron) {
    color: var(--vinta-text-secondary, #a1a1aa);
  }

  .popover-nav-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.08));
    color: var(--vinta-text-secondary, #b0b0b0);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .popover-nav-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #ffffff;
  }

  /* Days View */
  .popover-weekdays {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    text-align: center;
    gap: 2px;
  }

  .popover-weekday {
    font-size: 9px;
    font-weight: 700;
    color: var(--vinta-text-muted, #71717a);
    text-transform: uppercase;
  }

  .popover-days-grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 2px;
  }

  .popover-day-cell {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 25px;
    border-radius: 6px;
    background: transparent;
    border: none;
    color: var(--vinta-text-primary, #e4e4e7);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.12s ease;
    padding: 0;
  }

  .popover-day-cell:hover:not(.blank) {
    background: rgba(255, 255, 255, 0.1);
    color: #ffffff;
  }

  .popover-day-cell.today {
    color: var(--vinta-primary, #38bdf8);
    font-weight: 700;
  }

  .popover-day-cell.selected {
    background: var(--vinta-primary, #38bdf8) !important;
    color: #0b0f19 !important;
    font-weight: 700;
  }

  .popover-day-cell.blank {
    cursor: default;
    pointer-events: none;
  }

  /* Months Selector Grid */
  .months-selector-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 5px;
    padding: 4px 0;
  }

  .month-item-btn {
    padding: 8px 0;
    border-radius: 7px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.06));
    color: var(--vinta-text-primary, #e4e4e7);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .month-item-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #ffffff;
  }

  .month-item-btn.active {
    background: rgba(56, 189, 248, 0.22);
    border-color: rgba(56, 189, 248, 0.4);
    color: var(--vinta-primary, #38bdf8);
  }

  /* Fast Year Jump Form */
  .decade-jump-form {
    display: flex;
    align-items: center;
    gap: 5px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.08));
    border-radius: 7px;
    padding: 2px 6px;
  }

  .decade-jump-input {
    flex: 1;
    min-width: 0;
    background: transparent;
    border: none;
    outline: none;
    color: var(--vinta-text-primary, #ffffff);
    font-size: 11px;
    font-weight: 600;
    font-family: inherit;
    padding: 2px 0;
  }

  .decade-jump-input::placeholder {
    color: var(--vinta-text-muted, #71717a);
  }

  .decade-jump-btn {
    padding: 2px 7px;
    border-radius: 5px;
    background: rgba(56, 189, 248, 0.15);
    border: 1px solid rgba(56, 189, 248, 0.3);
    color: var(--vinta-primary, #38bdf8);
    font-size: 10px;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .decade-jump-btn:hover {
    background: rgba(56, 189, 248, 0.25);
  }

  /* Decade Selector Grid */
  .decade-selector-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 4px;
    padding: 4px 0;
  }

  .decade-item-btn {
    padding: 7px 0;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.06));
    color: var(--vinta-text-primary, #e4e4e7);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .decade-item-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #ffffff;
  }

  .decade-item-btn.outer {
    color: var(--vinta-text-muted, #71717a);
    opacity: 0.6;
  }

  .decade-item-btn.active {
    background: rgba(56, 189, 248, 0.22);
    border-color: rgba(56, 189, 248, 0.4);
    color: var(--vinta-primary, #38bdf8);
    font-weight: 700;
  }

  /* Popover Footer */
  .popover-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 4px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
  }

  .popover-today-btn {
    font-size: 10px;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 6px;
    background: rgba(56, 189, 248, 0.12);
    border: 1px solid rgba(56, 189, 248, 0.25);
    color: var(--vinta-primary, #38bdf8);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .popover-today-btn:hover {
    background: rgba(56, 189, 248, 0.22);
    border-color: rgba(56, 189, 248, 0.4);
  }

  .popover-back-btn {
    font-size: 10px;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 6px;
    background: transparent;
    border: none;
    color: var(--vinta-text-secondary, #a1a1aa);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .popover-back-btn:hover {
    color: var(--vinta-text-primary, #ffffff);
  }
</style>
