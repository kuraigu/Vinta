<script lang="ts">
  import {
    Play,
    Pause,
    RotateCcw,
    SkipForward
  } from 'lucide-svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import {
    timerAppStore,
    toggleFocus,
    resetFocus,
    skipFocusSession,
    setFocusPhase,
    type FocusPhase
  } from '../../store'

  const RADIUS = 70
  const CIRCUMFERENCE = 2 * Math.PI * RADIUS

  function formatDigitalTime(totalSecs: number): string {
    const s = Math.max(0, totalSecs)
    const hrs = Math.floor(s / 3600)
    const mins = Math.floor((s % 3600) / 60)
    const secs = s % 60

    if (hrs > 0) {
      return `${hrs.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
    }
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  function computeFocusProgressOffset() {
    const total = $timerAppStore.focus.totalDuration || 1
    const remaining = $timerAppStore.focus.timeRemaining
    const progress = Math.max(0, Math.min(1, remaining / total))
    return (1 - progress) * CIRCUMFERENCE
  }
  let focusProgressOffset = $derived.by(computeFocusProgressOffset)

  let focusFormattedDigits = $derived.by(() =>
    formatDigitalTime($timerAppStore.focus.timeRemaining)
  )

  function handleFocusPhaseChange(phase: FocusPhase): void {
    setFocusPhase(phase)
  }
</script>

<div class="tab-panel">
  <div class="focus-cycle-bar">
    <div class="cycle-left-group">
      <span class="cycle-text">
        Cycle {$timerAppStore.focus.currentCycle} / {$timerAppStore.focus
          .settings.longBreakInterval}
      </span>
      <div class="cycle-dots-row">
        {#each Array($timerAppStore.focus.settings.longBreakInterval) as _, idx (idx)}
          <span
            class="cycle-dot"
            class:active={idx < $timerAppStore.focus.currentCycle}
          ></span>
        {/each}
      </div>
    </div>
  </div>

  <div class="dial-stage focus-dial-stage">
    <svg class="dial-svg" viewBox="0 0 160 160">
      <circle class="dial-track" cx="80" cy="80" r={RADIUS} />
      <circle
        class="dial-progress phase-{$timerAppStore.focus.phase}"
        cx="80"
        cy="80"
        r={RADIUS}
        stroke-dasharray={CIRCUMFERENCE}
        stroke-dashoffset={focusProgressOffset}
      />
    </svg>

    <div class="dial-center-content">
      <span class="dial-large-digits">{focusFormattedDigits}</span>
      <span class="dial-subtitle">
        {#if $timerAppStore.focus.phase === 'focus'}
          Stay in flow
        {:else if $timerAppStore.focus.phase === 'shortBreak'}
          Short break
        {:else}
          Extended rest
        {/if}
      </span>
    </div>
  </div>

  <div class="bottom-controls-cluster">
    <div class="phase-switch-bar">
      <button
        class="phase-pill-toggle"
        class:active={$timerAppStore.focus.phase === 'focus'}
        class:phase-focus={$timerAppStore.focus.phase === 'focus'}
        onclick={() => handleFocusPhaseChange('focus')}
        use:islandTooltip={'Focus session'}>Focus</button
      >

      <button
        class="phase-pill-toggle"
        class:active={$timerAppStore.focus.phase === 'shortBreak'}
        class:phase-shortBreak={$timerAppStore.focus.phase === 'shortBreak'}
        onclick={() => handleFocusPhaseChange('shortBreak')}
        use:islandTooltip={'Short break'}>Short Break</button
      >

      <button
        class="phase-pill-toggle"
        class:active={$timerAppStore.focus.phase === 'longBreak'}
        class:phase-longBreak={$timerAppStore.focus.phase === 'longBreak'}
        onclick={() => handleFocusPhaseChange('longBreak')}
        use:islandTooltip={'Long break'}>Long Break</button
      >
    </div>

    <div class="controls-row">
      <button
        class="ctrl-btn"
        onclick={resetFocus}
        use:islandTooltip={'Reset session'}
      >
        <RotateCcw size={15} />
      </button>

      <button
        class="primary-pill-btn"
        class:phase-focus={$timerAppStore.focus.phase === 'focus'}
        class:phase-break={$timerAppStore.focus.phase !== 'focus'}
        onclick={toggleFocus}
        use:islandTooltip={$timerAppStore.focus.isRunning ? 'Pause' : 'Start'}
      >
        {#if $timerAppStore.focus.isRunning}
          <Pause size={16} />
          <span>Pause</span>
        {:else}
          <Play size={16} />
          <span>Start</span>
        {/if}
      </button>

      <button
        class="ctrl-btn"
        onclick={skipFocusSession}
        use:islandTooltip={'Advance to next cycle'}
      >
        <SkipForward size={15} />
      </button>
    </div>
  </div>
</div>
