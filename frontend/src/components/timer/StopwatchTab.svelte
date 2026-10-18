<script lang="ts">
  import { Play, Pause, RotateCcw, Flag } from 'lucide-svelte'
  import ScrollView from '../ScrollView.svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import {
    timerAppStore,
    toggleStopwatch,
    resetStopwatch,
    recordLap
  } from '../../store'

  function formatStopwatchDigits(totalMs: number): {
    main: string
    frac: string
  } {
    const ms = Math.max(0, totalMs)
    const totalSecs = Math.floor(ms / 1000)
    const mins = Math.floor(totalSecs / 60)
    const secs = totalSecs % 60
    const hundredths = Math.floor((ms % 1000) / 10)

    return {
      main: `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`,
      frac: `.${hundredths.toString().padStart(2, '0')}`
    }
  }

  let stopwatchDisplay = $derived.by(() =>
    formatStopwatchDigits($timerAppStore.stopwatch.elapsedMs)
  )
</script>

<div class="tab-panel">
  <div
    class="stopwatch-stage"
    class:has-laps={$timerAppStore.stopwatch.laps.length > 0}
  >
    <div class="stopwatch-digits-row">
      <span class="stopwatch-main-digits">{stopwatchDisplay.main}</span>
      <span class="stopwatch-fraction-digits">{stopwatchDisplay.frac}</span>
    </div>
  </div>

  {#if $timerAppStore.stopwatch.laps.length > 0}
    <div class="laps-container">
      <ScrollView class="laps-scroll-view">
        {#each $timerAppStore.stopwatch.laps as lap (lap.id)}
          {@const lapSplit = formatStopwatchDigits(lap.lapTimeMs)}
          {@const lapTotal = formatStopwatchDigits(lap.overallTimeMs)}
          <div class="lap-item">
            <span class="lap-number">Lap {lap.id}</span>
            <span class="lap-split">+{lapSplit.main}{lapSplit.frac}</span>
            <span class="lap-total">{lapTotal.main}{lapTotal.frac}</span>
          </div>
        {/each}
      </ScrollView>
    </div>
  {/if}

  <div class="controls-row">
    <button
      class="ctrl-btn"
      onclick={resetStopwatch}
      use:islandTooltip={'Reset stopwatch'}
    >
      <RotateCcw size={15} />
    </button>

    <button
      class="primary-pill-btn phase-stopwatch"
      onclick={toggleStopwatch}
      use:islandTooltip={$timerAppStore.stopwatch.isRunning ? 'Pause' : 'Start'}
    >
      {#if $timerAppStore.stopwatch.isRunning}
        <Pause size={16} />
        <span>Pause</span>
      {:else}
        <Play size={16} />
        <span>Start</span>
      {/if}
    </button>

    <button
      class="ctrl-btn"
      onclick={recordLap}
      disabled={!$timerAppStore.stopwatch.isRunning}
      use:islandTooltip={'Record lap split'}
    >
      <Flag size={15} />
    </button>
  </div>
</div>
