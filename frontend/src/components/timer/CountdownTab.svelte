<!--
  @component CountdownTab
  @description Interactive countdown timer. Features quick preset chips (1m, 5m, 15m, 25m),
  circular progress ring, play/pause controls, and an audible chime on completion.
-->
<script lang="ts">
  import { Play, Pause, RotateCcw, Plus } from 'lucide-svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import {
    timerAppStore,
    tStore,
    toggleCountdown,
    resetCountdown,
    setCountdownDuration
  } from '../../store'

  interface Props {
    onOpenCustom: () => void
  }

  let { onOpenCustom }: Props = $props()

  const RADIUS = 70
  const CIRCUMFERENCE = 2 * Math.PI * RADIUS

  const TIMER_PRESETS = [
    { label: '1m', secs: 60 },
    { label: '3m', secs: 180 },
    { label: '5m', secs: 300 },
    { label: '10m', secs: 600 },
    { label: '15m', secs: 900 },
    { label: '25m', secs: 1500 }
  ]

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

  function computeTimerProgressOffset() {
    const total = $timerAppStore.timer.totalDurationSeconds || 1
    const remaining = $timerAppStore.timer.remainingSeconds
    const progress = Math.max(0, Math.min(1, remaining / total))
    return (1 - progress) * CIRCUMFERENCE
  }
  let timerProgressOffset = $derived.by(computeTimerProgressOffset)

  let timerCountdownDisplay = $derived.by(() =>
    formatDigitalTime($timerAppStore.timer.remainingSeconds)
  )

  function handleAddMinute(): void {
    const current = $timerAppStore.timer.remainingSeconds
    setCountdownDuration(current + 60)
  }
</script>

<div class="tab-panel">
  <div class="focus-cycle-bar timer-top-placeholder" aria-hidden="true"></div>

  <div class="dial-stage timer-dial-stage">
    <svg class="dial-svg" viewBox="0 0 160 160">
      <circle class="dial-track" cx="80" cy="80" r={RADIUS} />
      <circle
        class="dial-progress timer-stroke"
        cx="80"
        cy="80"
        r={RADIUS}
        stroke-dasharray={CIRCUMFERENCE}
        stroke-dashoffset={timerProgressOffset}
      />
    </svg>

    <button
      class="dial-center-btn"
      onclick={onOpenCustom}
      use:islandTooltip={$tStore('timer.click_to_edit')}
    >
      <span class="dial-large-digits">{timerCountdownDisplay}</span>
      <span class="dial-subtitle">
        {$timerAppStore.timer.isRunning
          ? $tStore('timer.counting_down')
          : $tStore('timer.tap_to_edit')}
      </span>
    </button>
  </div>

  <div class="bottom-controls-cluster">
    <div class="presets-scroll-row">
      {#each TIMER_PRESETS as preset (preset.secs)}
        <button
          class="preset-chip"
          class:active={$timerAppStore.timer.totalDurationSeconds ===
            preset.secs}
          onclick={() => setCountdownDuration(preset.secs)}
          use:islandTooltip={$tStore('timer.set_preset', {
            label: preset.label
          })}
        >
          {preset.label}
        </button>
      {/each}

      <button
        class="preset-chip custom-trigger"
        onclick={onOpenCustom}
        use:islandTooltip={$tStore('timer.custom')}
      >
        {$tStore('timer.custom')}...
      </button>
    </div>

    <div class="controls-row">
      <button
        class="ctrl-btn"
        onclick={resetCountdown}
        use:islandTooltip={$tStore('timer.reset')}
      >
        <RotateCcw size={15} />
      </button>

      <button
        class="primary-pill-btn"
        onclick={toggleCountdown}
        use:islandTooltip={$timerAppStore.timer.isRunning
          ? $tStore('timer.pause')
          : $tStore('timer.start')}
      >
        {#if $timerAppStore.timer.isRunning}
          <Pause size={16} />
          <span>{$tStore('timer.pause')}</span>
        {:else}
          <Play size={16} />
          <span>{$tStore('timer.start')}</span>
        {/if}
      </button>

      <button
        class="ctrl-btn"
        onclick={handleAddMinute}
        use:islandTooltip={$tStore('timer.add_minute')}
      >
        <Plus size={15} />
      </button>
    </div>
  </div>
</div>

<style src="../../pages/styles/TimerPage.css"></style>
