<!--
  @component FocusTab
  @description Pomodoro focus session manager. Tracks work and break intervals, displays
  cycle counts, and integrates with the procedural soundscape generator for background white noise.
-->
<script lang="ts">
  import {
    Play,
    Pause,
    RotateCcw,
    SkipForward,
    Waves,
    CloudRain,
    Radio
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
  import {
    soundscapeStore,
    playPreset,
    type SoundscapePreset
  } from '../../services/soundscapeService'

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

  let soundscape = $derived($soundscapeStore)

  function computeActiveSoundscapeLabel() {
    if (!soundscape.isPlaying || soundscape.preset === 'none') return 'Ambience'
    if (soundscape.preset === 'rain') return 'Rain'
    if (soundscape.preset === 'brown') return 'Brown'
    if (soundscape.preset === 'static') return 'Static'
    return 'Ambience'
  }
  let activeSoundscapeLabel = $derived.by(computeActiveSoundscapeLabel)

  function handleSoundscapeChipClick() {
    if (!soundscape.isPlaying) {
      playPreset(soundscape.preset === 'none' ? 'rain' : soundscape.preset)
    } else {
      const order: SoundscapePreset[] = ['rain', 'brown', 'static', 'none']
      const curIdx = order.indexOf(soundscape.preset)
      const next = order[(curIdx + 1) % order.length]
      playPreset(next)
    }
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

    <div class="cycle-right-group">
      <button
        type="button"
        class="focus-soundscape-chip"
        class:is-active={soundscape.isPlaying}
        onclick={handleSoundscapeChipClick}
        use:islandTooltip={{
          text: soundscape.isPlaying
            ? `Ambience: ${activeSoundscapeLabel} (Click to switch/stop)`
            : 'Focus soundscape: Click to play offline ambience'
        }}
        aria-label="Focus soundscape preset"
      >
        {#if soundscape.preset === 'rain'}
          <CloudRain size={12} />
        {:else if soundscape.preset === 'static'}
          <Radio size={12} />
        {:else}
          <Waves size={12} />
        {/if}
        <span>{activeSoundscapeLabel}</span>
      </button>
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
        <RotateCcw size={16} />
      </button>

      <button
        class="primary-pill-btn"
        class:phase-focus={$timerAppStore.focus.phase === 'focus'}
        class:phase-break={$timerAppStore.focus.phase !== 'focus'}
        onclick={toggleFocus}
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
        use:islandTooltip={'Next cycle'}
      >
        <SkipForward size={16} />
      </button>
    </div>
  </div>
</div>

<style src="../../pages/styles/TimerPage.css"></style>
