<script lang="ts">
  import { onMount, onDestroy } from 'svelte'

  let { expanded = false } = $props<{ expanded?: boolean }>()

  let now = $state(new Date())
  let timer: ReturnType<typeof setInterval>

  onMount(() => {
    // Update every minute to catch day changes safely
    timer = setInterval(() => {
      now = new Date()
    }, 60000)
  })

  onDestroy(() => {
    clearInterval(timer)
  })

  // Date formatted for minimal view (e.g. "Mon, Jul 7")
  let minimalDate = $derived(
    now.toLocaleDateString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric'
    })
  )

  // Calendar Header (e.g. "July 2026")
  let currentMonthStr = $derived(
    now.toLocaleDateString(undefined, {
      month: 'long',
      year: 'numeric'
    })
  )

  // Calendar grid math
  let year = $derived(now.getFullYear())
  let month = $derived(now.getMonth()) // 0-11
  let todayDate = $derived(now.getDate())

  // Find day of week for 1st of month (0 = Sun, 1 = Mon ... 6 = Sat)
  let firstDayOfMonth = $derived(new Date(year, month, 1).getDay())
  let daysInMonth = $derived(new Date(year, month + 1, 0).getDate())

  // Create arrays for Svelte {#each} blocks
  let blanks = $derived(Array(firstDayOfMonth).fill(null))
  let days = $derived(Array.from({ length: daysInMonth }, (_, i) => i + 1))

  const weekDays = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']
</script>

<div class="calendar-container {expanded ? 'expanded' : 'minimal'}">
  {#if !expanded}
    <div class="minimal-date">{minimalDate}</div>
  {:else}
    <div class="calendar-grid">
      <div class="month-header">{currentMonthStr}</div>
      <div class="weekdays">
        {#each weekDays as wd (wd)}
          <div class="weekday">{wd}</div>
        {/each}
      </div>
      <div class="days">
        {#each blanks as _, i (i)}
          <div class="day empty" data-index={_}></div>
        {/each}
        {#each days as day (day)}
          <div class="day {day === todayDate ? 'today' : ''}">{day}</div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .calendar-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: var(--c-text-primary);
  }

  /* MINIMAL */
  .minimal .minimal-date {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--c-text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* EXPANDED */
  .expanded {
    width: 100%;
    height: 100%;
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .calendar-grid {
    width: 100%;
    display: flex;
    flex-direction: column;
  }

  .month-header {
    font-size: 1.1rem;
    font-weight: 600;
    margin-bottom: 0.5rem;
    text-align: left;
    padding-left: 0.2rem;
  }

  .weekdays,
  .days {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    text-align: center;
    gap: 2px;
  }

  .weekday {
    font-size: 0.75rem;
    color: #a0a0a0;
    font-weight: 600;
    margin-bottom: 4px;
  }

  .day {
    font-size: 0.85rem;
    font-weight: 500;
    padding: 6px 0;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .day.today {
    background-color: #34c759; /* Vinta Green */
    color: black;
    font-weight: 700;
    box-shadow: 0 4px 10px rgba(52, 199, 89, 0.4);
    transform: scale(1.05);
  }
</style>
