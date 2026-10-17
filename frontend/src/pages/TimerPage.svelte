<svelte:options />

<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 42, r: 16 },
    expanded: { w: 420, h: 370, r: 28 },
    borderColor: 'rgba(56, 189, 248, 0.35)',
    glowColor: 'rgba(56, 189, 248, 0.25)'
  }
</script>

<script lang="ts">
  import {
    Hourglass,
    Clock,
    Flame,
    Bell,
    Play,
    Pause,
    RotateCcw,
    SkipForward,
    SlidersHorizontal,
    Flag,
    Plus,
    Trash2,
    X,
    Check,
    Pencil
  } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import CustomTimePicker from '../components/CustomTimePicker.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from '../components/HorizontalTabSelection.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    timerAppStore,
    timeFormatStore,
    setActiveTimerTab,
    toggleCountdown,
    resetCountdown,
    setCountdownDuration,
    toggleStopwatch,
    resetStopwatch,
    recordLap,
    toggleFocus,
    resetFocus,
    skipFocusSession,
    setFocusPhase,
    updateFocusSettings,
    addAlarm,
    updateAlarm,
    toggleAlarm,
    deleteAlarm,
    dismissRingingAlarm,
    snoozeRingingAlarm,
    type TimerAppTab,
    type FocusPhase,
    type AlarmItem,
    type AlarmRepeat
  } from '../store'

  let { expanded = false }: { expanded?: boolean } = $props()

  let isSettingsOpen = $state(false)
  let isCustomDrawerOpen = $state(false)
  let isNewAlarmOpen = $state(false)
  let editingAlarmId = $state<string | null>(null)

  // Custom timer duration steppers state
  let customHours = $state(0)
  let customMinutes = $state(10)
  let customSeconds = $state(0)

  // New/edit alarm form state
  let newAlarmTime = $state('09:00')
  let newAlarmLabel = $state('')
  let newAlarmRepeat = $state<AlarmRepeat>('once')
  let newAlarmDays = $state<number[]>([1, 2, 3, 4, 5])

  function computeIs24H(): boolean {
    return $timeFormatStore === '24h'
  }
  let is24H = $derived.by(computeIs24H)

  const RADIUS = 70
  const CIRCUMFERENCE = 2 * Math.PI * RADIUS

  const DAY_NAMES = [
    'Sunday',
    'Monday',
    'Tuesday',
    'Wednesday',
    'Thursday',
    'Friday',
    'Saturday'
  ]
  const DAY_SHORT = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

  const TIMER_PRESETS = [
    { label: '1m', secs: 60 },
    { label: '3m', secs: 180 },
    { label: '5m', secs: 300 },
    { label: '10m', secs: 600 },
    { label: '15m', secs: 900 },
    { label: '25m', secs: 1500 }
  ]

  const QUICK_SECOND_PRESETS = [
    { label: '3s', secs: 3 },
    { label: '5s', secs: 5 },
    { label: '10s', secs: 10 },
    { label: '30s', secs: 30 },
    { label: '1m', secs: 60 },
    { label: '5m', secs: 300 }
  ]

  /**
   * Formats seconds into digital MM:SS or HH:MM:SS format.
   *
   * @param totalSecs Total duration in seconds.
   * @returns Formatted digital time string.
   */
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

  /**
   * Formats stopwatch milliseconds into minutes, seconds, and fractional hundredths.
   *
   * @param totalMs Milliseconds elapsed.
   * @returns Main digits and fractional hundredths.
   */
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

  /**
   * Formats a 24-hour HH:MM time string into 12-hour or 24-hour display string based on user setting.
   *
   * @param timeStr Stored HH:MM 24-hour time string.
   * @param use24H Whether 24-hour format is enabled.
   * @returns Formatted display time string.
   */
  function formatAlarmDisplayTime(timeStr: string, use24H: boolean): string {
    if (!timeStr) return ''
    const [hStr, mStr] = timeStr.split(':')
    const h24 = parseInt(hStr, 10) || 0
    const m = (mStr || '00').padStart(2, '0')

    if (use24H) {
      return `${h24.toString().padStart(2, '0')}:${m}`
    }

    const h12 = h24 % 12 || 12
    const ampm = h24 >= 12 ? 'PM' : 'AM'
    return `${h12}:${m} ${ampm}`
  }

  /**
   * Formats trigger epoch timestamp into 12-hour or 24-hour display format.
   *
   * @param timestamp Epoch millisecond timestamp.
   * @param use24H Whether 24-hour format is enabled.
   * @returns Formatted time string.
   */
  function formatTriggeredTime(timestamp: number, use24H: boolean): string {
    const d = new Date(timestamp)
    const h24 = d.getHours()
    const m = d.getMinutes().toString().padStart(2, '0')
    if (use24H) {
      return `${h24.toString().padStart(2, '0')}:${m}`
    }
    const h12 = h24 % 12 || 12
    const ampm = h24 >= 12 ? 'PM' : 'AM'
    return `${h12}:${m} ${ampm}`
  }

  /**
   * Formats alarm recurrence description badge.
   *
   * @param alarm Target alarm item.
   * @returns Human-readable recurrence label.
   */
  function formatAlarmRepeatBadge(alarm: AlarmItem): string {
    if (alarm.repeat === 'once') return 'Once'
    if (alarm.repeat === 'daily') return 'Every day'
    if (alarm.repeat === 'weekdays') return 'Mon - Fri'
    if (!alarm.days || alarm.days.length === 0) return 'Custom'
    if (alarm.days.length === 7) return 'Every day'
    return alarm.days.map((d) => DAY_SHORT[d]).join(', ')
  }

  // Active minimal mode derivation
  function computeActiveMinimalMode() {
    const s = $timerAppStore
    if (s.timer.isRunning) return 'timer'
    if (s.stopwatch.isRunning) return 'stopwatch'
    if (s.focus.isRunning) return 'focus'
    return s.activeTab === 'alarms' ? 'timer' : s.activeTab
  }
  let activeMinimalMode = $derived.by(computeActiveMinimalMode)

  function computeMinimalModeLabel() {
    if (activeMinimalMode === 'timer') return 'Timer'
    if (activeMinimalMode === 'stopwatch') return 'Stopwatch'
    if ($timerAppStore.focus.phase === 'focus') return 'Focus'
    if ($timerAppStore.focus.phase === 'shortBreak') return 'Break'
    return 'Rest'
  }
  let minimalModeLabel = $derived.by(computeMinimalModeLabel)

  function computeMinimalDigits() {
    const s = $timerAppStore
    if (activeMinimalMode === 'timer') {
      return formatDigitalTime(s.timer.remainingSeconds)
    }
    if (activeMinimalMode === 'stopwatch') {
      const sw = formatStopwatchDigits(s.stopwatch.elapsedMs)
      return `${sw.main}${sw.frac}`
    }
    return formatDigitalTime(s.focus.timeRemaining)
  }
  let minimalTimeDigits = $derived.by(computeMinimalDigits)

  function computeIsMinimalActiveRunning() {
    const s = $timerAppStore
    if (activeMinimalMode === 'timer') return s.timer.isRunning
    if (activeMinimalMode === 'stopwatch') return s.stopwatch.isRunning
    return s.focus.isRunning
  }
  let isMinimalActiveRunning = $derived.by(computeIsMinimalActiveRunning)

  // Radial stroke progress calculations
  function computeTimerProgressOffset() {
    const total = $timerAppStore.timer.totalDurationSeconds || 1
    const remaining = $timerAppStore.timer.remainingSeconds
    const progress = Math.max(0, Math.min(1, remaining / total))
    return (1 - progress) * CIRCUMFERENCE
  }
  let timerProgressOffset = $derived.by(computeTimerProgressOffset)

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

  let stopwatchDisplay = $derived.by(() =>
    formatStopwatchDigits($timerAppStore.stopwatch.elapsedMs)
  )

  let timerCountdownDisplay = $derived.by(() =>
    formatDigitalTime($timerAppStore.timer.remainingSeconds)
  )

  /**
   * Toggles the active timer in minimal mode.
   */
  function handleMinimalToggle(): void {
    if (activeMinimalMode === 'timer') {
      toggleCountdown()
    } else if (activeMinimalMode === 'stopwatch') {
      toggleStopwatch()
    } else {
      toggleFocus()
    }
  }

  /**
   * Resets the active timer in minimal mode.
   */
  function handleMinimalReset(): void {
    if (activeMinimalMode === 'timer') {
      resetCountdown()
    } else if (activeMinimalMode === 'stopwatch') {
      resetStopwatch()
    } else {
      resetFocus()
    }
  }

  /**
   * Switches the active tab in maximized mode.
   *
   * @param tab Target utility tab.
   */
  function handleTabSelect(tab: TimerAppTab): void {
    setActiveTimerTab(tab)
  }

  const TIMER_TABS: TabItem<TimerAppTab>[] = [
    {
      id: 'timer',
      label: 'Timer',
      icon: Hourglass,
      tooltip: 'Countdown timer'
    },
    {
      id: 'stopwatch',
      label: 'Stopwatch',
      icon: Clock,
      tooltip: 'Precision stopwatch'
    },
    {
      id: 'focus',
      label: 'Focus',
      icon: Flame,
      tooltip: 'Focus sprints & breaks'
    },
    {
      id: 'alarms',
      label: 'Alarms',
      icon: Bell,
      tooltip: 'Scheduled reminders'
    }
  ]

  /**
   * Switches focus phase directly.
   *
   * @param phase Target focus phase.
   */
  function handleFocusPhaseChange(phase: FocusPhase): void {
    setFocusPhase(phase)
  }

  /**
   * Adds one minute to the countdown timer.
   */
  function handleAddMinute(): void {
    const current = $timerAppStore.timer.remainingSeconds
    setCountdownDuration(current + 60)
  }

  /**
   * Opens the custom timer duration drawer.
   */
  function openCustomDrawer(): void {
    const total = $timerAppStore.timer.totalDurationSeconds
    customHours = Math.floor(total / 3600)
    customMinutes = Math.floor((total % 3600) / 60)
    customSeconds = total % 60
    isCustomDrawerOpen = true
  }

  /**
   * Applies custom duration from drawer steppers.
   */
  function applyCustomDuration(): void {
    const totalSecs = customHours * 3600 + customMinutes * 60 + customSeconds
    if (totalSecs > 0) {
      setCountdownDuration(totalSecs)
    }
    isCustomDrawerOpen = false
  }

  /**
   * Sets custom duration quickly from a preset second count.
   *
   * @param secs Duration in seconds.
   */
  function setQuickSecondPreset(secs: number): void {
    customHours = Math.floor(secs / 3600)
    customMinutes = Math.floor((secs % 3600) / 60)
    customSeconds = secs % 60
  }

  /**
   * Toggles day of week selection in custom recurrence mode.
   *
   * @param dayIndex Day index (0 to 6).
   */
  function toggleDaySelection(dayIndex: number): void {
    if (newAlarmDays.includes(dayIndex)) {
      if (newAlarmDays.length > 1) {
        newAlarmDays = newAlarmDays.filter((d) => d !== dayIndex)
      }
    } else {
      newAlarmDays = [...newAlarmDays, dayIndex].sort((a, b) => a - b)
    }
  }

  /**
   * Opens the alarm creator for a new alarm.
   */
  function openCreateAlarm(): void {
    editingAlarmId = null
    newAlarmTime = '09:00'
    newAlarmLabel = ''
    newAlarmRepeat = 'once'
    newAlarmDays = [1, 2, 3, 4, 5]
    isNewAlarmOpen = true
  }

  /**
   * Pre-populates the alarm form to edit an existing alarm.
   *
   * @param alarm Target alarm item.
   */
  function startEditAlarm(alarm: AlarmItem): void {
    editingAlarmId = alarm.id
    newAlarmTime = alarm.time
    newAlarmLabel = alarm.label || ''
    newAlarmRepeat = alarm.repeat || 'once'
    newAlarmDays =
      Array.isArray(alarm.days) && alarm.days.length > 0
        ? [...alarm.days]
        : [1, 2, 3, 4, 5]
    isNewAlarmOpen = true
  }

  /**
   * Cancels alarm creation or editing.
   */
  function cancelAlarmForm(): void {
    editingAlarmId = null
    isNewAlarmOpen = false
  }

  /**
   * Saves newly created or edited alarm item.
   */
  function handleSaveNewAlarm(): void {
    if (newAlarmTime) {
      if (editingAlarmId) {
        updateAlarm(editingAlarmId, {
          time: newAlarmTime,
          label: newAlarmLabel,
          repeat: newAlarmRepeat,
          days:
            newAlarmRepeat === 'daily'
              ? [0, 1, 2, 3, 4, 5, 6]
              : newAlarmRepeat === 'weekdays'
                ? [1, 2, 3, 4, 5]
                : newAlarmDays
        })
        editingAlarmId = null
      } else {
        addAlarm(newAlarmTime, newAlarmLabel, newAlarmRepeat, newAlarmDays)
      }
      newAlarmLabel = ''
      newAlarmRepeat = 'once'
      isNewAlarmOpen = false
    }
  }

  /**
   * Closes active drawers and forms when the island collapses.
   */
  function handleMinimizeEffect(): void {
    if (!expanded) {
      isSettingsOpen = false
      isCustomDrawerOpen = false
      isNewAlarmOpen = false
      editingAlarmId = null
    }
  }
  $effect(handleMinimizeEffect)
</script>

<div
  class="timer-page"
  class:minimal-mode={!expanded}
  class:expanded-mode={expanded}
>
  {#if !expanded}
    {#if $timerAppStore.ringingAlarm}
      <!-- Minimized Ringing Alarm Capsule -->
      <div class="timer-minimal ringing-capsule">
        <div class="minimal-ringing-pulse" aria-hidden="true"></div>

        <div class="minimal-ringing-info">
          <div class="ringing-bell-badge">
            <Bell size={13} class="bell-shake" />
          </div>
          <div class="ringing-text-col">
            <span class="ringing-title">
              {$timerAppStore.ringingAlarm.label || 'Alarm'}
            </span>
            <span class="ringing-subtitle">
              {formatAlarmDisplayTime($timerAppStore.ringingAlarm.time, is24H)}
            </span>
          </div>
        </div>

        <div class="ringing-capsule-actions">
          <button
            class="ringing-action-pill snooze"
            onclick={(e: MouseEvent) => {
              e.stopPropagation()
              snoozeRingingAlarm()
            }}
          >
            <Clock size={11} />
            <span>Snooze</span>
          </button>
          <button
            class="ringing-action-pill stop"
            onclick={(e: MouseEvent) => {
              e.stopPropagation()
              dismissRingingAlarm()
            }}
          >
            <X size={11} />
            <span>Stop</span>
          </button>
        </div>
      </div>
    {:else}
      <!-- Minimal Mode Capsule -->
      <div class="timer-minimal">
        <!-- Ambient dot matrix background pattern across capsule with soft organic fade -->
        <div
          class="minimal-dots-bg"
          class:mode-timer={activeMinimalMode === 'timer'}
          class:mode-stopwatch={activeMinimalMode === 'stopwatch'}
          class:mode-focus={activeMinimalMode === 'focus' &&
            $timerAppStore.focus.phase === 'focus'}
          class:mode-break={activeMinimalMode === 'focus' &&
            $timerAppStore.focus.phase !== 'focus'}
          aria-hidden="true"
        ></div>

        <div class="minimal-info-trigger">
          <span class="minimal-mode-label">{minimalModeLabel}</span>

          <div class="minimal-center-group">
            <span class="minimal-digits">{minimalTimeDigits}</span>

            {#if activeMinimalMode === 'focus'}
              <div class="minimal-cycle-dots">
                {#each Array($timerAppStore.focus.settings.longBreakInterval) as _, idx (idx)}
                  <span
                    class="cycle-dot"
                    class:active={idx < $timerAppStore.focus.currentCycle}
                  ></span>
                {/each}
              </div>
            {/if}
          </div>
        </div>

        <!-- Minimal Mode Controls (Play/Pause and Reset without tooltips) -->
        <div class="minimal-actions">
          <button
            class="mini-btn primary"
            class:phase-focus={activeMinimalMode === 'focus' &&
              $timerAppStore.focus.phase === 'focus'}
            class:phase-break={activeMinimalMode === 'focus' &&
              $timerAppStore.focus.phase !== 'focus'}
            class:phase-stopwatch={activeMinimalMode === 'stopwatch'}
            onclick={(e: MouseEvent) => {
              e.stopPropagation()
              handleMinimalToggle()
            }}
          >
            {#if isMinimalActiveRunning}
              <Pause size={13} />
            {:else}
              <Play size={13} />
            {/if}
          </button>

          <button
            class="mini-btn"
            onclick={(e: MouseEvent) => {
              e.stopPropagation()
              handleMinimalReset()
            }}
          >
            <RotateCcw size={12} />
          </button>
        </div>
      </div>
    {/if}
  {:else}
    <!-- Maximized Full Canvas View -->
    <div class="timer-expanded">
      <!-- Ambient dot matrix background pattern matching Home greeting aesthetic -->
      <div
        class="expanded-dots-bg"
        class:mode-timer={$timerAppStore.activeTab === 'timer'}
        class:mode-stopwatch={$timerAppStore.activeTab === 'stopwatch'}
        class:mode-focus={$timerAppStore.activeTab === 'focus' &&
          $timerAppStore.focus.phase === 'focus'}
        class:mode-break={$timerAppStore.activeTab === 'focus' &&
          $timerAppStore.focus.phase !== 'focus'}
        aria-hidden="true"
      ></div>

      <!-- Uniform App Header with 4-Tab Navigation Bar -->
      <div class="app-header">
        <BackButton />

        <HorizontalTabSelection
          tabs={TIMER_TABS}
          value={$timerAppStore.activeTab}
          onchange={handleTabSelect}
          ariaLabel="Timer utilities"
        />

        {#if $timerAppStore.activeTab === 'focus'}
          <button
            class="header-tool-btn"
            class:active={isSettingsOpen}
            onclick={() => {
              isSettingsOpen = !isSettingsOpen
            }}
            use:islandTooltip={'Focus settings'}
          >
            <SlidersHorizontal size={14} />
          </button>
        {:else if $timerAppStore.activeTab === 'alarms'}
          <button
            class="header-tool-btn"
            class:active={isNewAlarmOpen}
            onclick={() => {
              if (isNewAlarmOpen) {
                cancelAlarmForm()
              } else {
                openCreateAlarm()
              }
            }}
            use:islandTooltip={'Add new alarm'}
          >
            <Plus size={15} />
          </button>
        {:else}
          <div style="width: 30px;"></div>
        {/if}
      </div>

      <!-- Active Ringing Alarm Banner Card -->
      {#if $timerAppStore.ringingAlarm}
        <div class="expanded-ringing-banner">
          <div class="ringing-banner-glow" aria-hidden="true"></div>
          <div class="ringing-banner-left">
            <div class="banner-bell-badge">
              <Bell size={16} class="bell-shake" />
            </div>
            <div class="banner-info-group">
              <div class="banner-title-row">
                <span class="banner-alarm-title">
                  {$timerAppStore.ringingAlarm.label || 'Alarm Reminder'}
                </span>
                <span class="banner-alarm-time">
                  {formatAlarmDisplayTime(
                    $timerAppStore.ringingAlarm.time,
                    is24H
                  )}
                </span>
              </div>
              <span class="banner-alarm-meta">
                Scheduled for {formatAlarmDisplayTime(
                  $timerAppStore.ringingAlarm.time,
                  is24H
                )} (Triggered at {formatTriggeredTime(
                  $timerAppStore.ringingAlarm.triggeredAt,
                  is24H
                )}) • {formatAlarmRepeatBadge(
                  $timerAppStore.ringingAlarm as unknown as AlarmItem
                )}
              </span>
            </div>
          </div>

          <div class="banner-actions">
            <button
              class="banner-action-btn snooze"
              onclick={() => snoozeRingingAlarm()}
              use:islandTooltip={'Snooze alarm'}
            >
              <Clock size={12} />
              <span>Snooze</span>
            </button>
            <button
              class="banner-action-btn stop"
              onclick={dismissRingingAlarm}
              use:islandTooltip={'Stop alarm'}
            >
              <X size={12} />
              <span>Stop</span>
            </button>
          </div>
        </div>
      {/if}

      <!-- Tab 1: Countdown Timer -->
      {#if $timerAppStore.activeTab === 'timer'}
        <div class="tab-panel">
          <!-- Top alignment spacer matching Focus cycle bar for uniform layout -->
          <div
            class="focus-cycle-bar timer-top-placeholder"
            aria-hidden="true"
          ></div>

          <!-- Expanded Radial Progress Dial for Timer -->
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

            <!-- Clickable Center to open custom timer setter directly -->
            <button
              class="dial-center-btn"
              onclick={openCustomDrawer}
              use:islandTooltip={'Click to edit duration'}
            >
              <span class="dial-large-digits">{timerCountdownDisplay}</span>
              <span class="dial-subtitle">
                {$timerAppStore.timer.isRunning
                  ? 'Counting down'
                  : 'Tap to edit'}
              </span>
            </button>
          </div>

          <!-- Bottom Controls Cluster matching Focus layout -->
          <div class="bottom-controls-cluster">
            <!-- Presets and Custom Trigger Row -->
            <div class="presets-scroll-row">
              {#each TIMER_PRESETS as preset (preset.secs)}
                <button
                  class="preset-chip"
                  class:active={$timerAppStore.timer.totalDurationSeconds ===
                    preset.secs}
                  onclick={() => setCountdownDuration(preset.secs)}
                  use:islandTooltip={`Set ${preset.label}`}
                >
                  {preset.label}
                </button>
              {/each}

              <button
                class="preset-chip custom-trigger"
                onclick={openCustomDrawer}
                use:islandTooltip={'Set custom duration'}
              >
                Custom...
              </button>
            </div>

            <!-- Controls Row -->
            <div class="controls-row">
              <button
                class="ctrl-btn"
                onclick={resetCountdown}
                use:islandTooltip={'Reset countdown'}
              >
                <RotateCcw size={15} />
              </button>

              <button
                class="primary-pill-btn"
                onclick={toggleCountdown}
                use:islandTooltip={$timerAppStore.timer.isRunning
                  ? 'Pause'
                  : 'Start'}
              >
                {#if $timerAppStore.timer.isRunning}
                  <Pause size={16} />
                  <span>Pause</span>
                {:else}
                  <Play size={16} />
                  <span>Start</span>
                {/if}
              </button>

              <button
                class="ctrl-btn"
                onclick={handleAddMinute}
                use:islandTooltip={'Add 1 minute'}
              >
                <Plus size={15} />
              </button>
            </div>
          </div>
        </div>
      {/if}

      <!-- Tab 2: Stopwatch -->
      {#if $timerAppStore.activeTab === 'stopwatch'}
        <div class="tab-panel">
          <div
            class="stopwatch-stage"
            class:has-laps={$timerAppStore.stopwatch.laps.length > 0}
          >
            <div class="stopwatch-digits-row">
              <span class="stopwatch-main-digits">{stopwatchDisplay.main}</span>
              <span class="stopwatch-fraction-digits"
                >{stopwatchDisplay.frac}</span
              >
            </div>
          </div>

          <!-- Laps list wrapped in Vinta ScrollView with custom controls -->
          {#if $timerAppStore.stopwatch.laps.length > 0}
            <div class="laps-container">
              <ScrollView class="laps-scroll-view">
                {#each $timerAppStore.stopwatch.laps as lap (lap.id)}
                  {@const lapSplit = formatStopwatchDigits(lap.lapTimeMs)}
                  {@const lapTotal = formatStopwatchDigits(lap.overallTimeMs)}
                  <div class="lap-item">
                    <span class="lap-number">Lap {lap.id}</span>
                    <span class="lap-split"
                      >+{lapSplit.main}{lapSplit.frac}</span
                    >
                    <span class="lap-total">{lapTotal.main}{lapTotal.frac}</span
                    >
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
              use:islandTooltip={$timerAppStore.stopwatch.isRunning
                ? 'Pause'
                : 'Start'}
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
      {/if}

      <!-- Tab 3: Focus & Break Sessions (Unboxed Clean Dial) -->
      {#if $timerAppStore.activeTab === 'focus'}
        <div class="tab-panel">
          <!-- Cycle Progress Bar above Dial with Settings Trigger -->
          <div class="focus-cycle-bar">
            <div class="cycle-left-group">
              <span class="cycle-text">
                Cycle {$timerAppStore.focus.currentCycle} / {$timerAppStore
                  .focus.settings.longBreakInterval}
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

            <button
              class="focus-settings-trigger-btn"
              onclick={() => {
                isSettingsOpen = true
              }}
              use:islandTooltip={'Focus settings'}
              aria-label="Focus settings"
            >
              <SlidersHorizontal size={13} />
            </button>
          </div>

          <!-- Unboxed Radial Dial directly on canvas -->
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

          <!-- Bottom Controls Cluster matching Timer layout -->
          <div class="bottom-controls-cluster">
            <!-- Phase Switch Pills -->
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
                class:phase-shortBreak={$timerAppStore.focus.phase ===
                  'shortBreak'}
                onclick={() => handleFocusPhaseChange('shortBreak')}
                use:islandTooltip={'Short break'}>Short Break</button
              >

              <button
                class="phase-pill-toggle"
                class:active={$timerAppStore.focus.phase === 'longBreak'}
                class:phase-longBreak={$timerAppStore.focus.phase ===
                  'longBreak'}
                onclick={() => handleFocusPhaseChange('longBreak')}
                use:islandTooltip={'Long break'}>Long Break</button
              >
            </div>

            <!-- Focus Controls Bar -->
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
                use:islandTooltip={$timerAppStore.focus.isRunning
                  ? 'Pause'
                  : 'Start'}
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
      {/if}

      <!-- Tab 4: Alarms & Reminders -->
      {#if $timerAppStore.activeTab === 'alarms'}
        <div class="tab-panel alarms-panel">
          {#if isNewAlarmOpen}
            <div class="alarm-create-card">
              <span class="create-card-title">
                {editingAlarmId ? 'Edit Alarm Reminder' : 'New Alarm Reminder'}
              </span>
              <div class="create-inputs-row">
                <!-- Custom Time Steppers (No native OS flyouts) -->
                <CustomTimePicker
                  bind:value={newAlarmTime}
                  ariaLabel="Alarm time"
                />

                <input
                  type="text"
                  class="alarm-label-input"
                  placeholder="Reminder label..."
                  bind:value={newAlarmLabel}
                  aria-label="Alarm note"
                />
              </div>

              <!-- Recurrence Selector -->
              <div class="alarm-repeat-selector">
                <span class="repeat-row-label">Repeat:</span>
                <div class="repeat-pills-row">
                  <button
                    class="repeat-pill"
                    class:active={newAlarmRepeat === 'once'}
                    onclick={() => {
                      newAlarmRepeat = 'once'
                    }}>Once</button
                  >
                  <button
                    class="repeat-pill"
                    class:active={newAlarmRepeat === 'daily'}
                    onclick={() => {
                      newAlarmRepeat = 'daily'
                    }}>Daily</button
                  >
                  <button
                    class="repeat-pill"
                    class:active={newAlarmRepeat === 'weekdays'}
                    onclick={() => {
                      newAlarmRepeat = 'weekdays'
                    }}>Weekdays</button
                  >
                  <button
                    class="repeat-pill"
                    class:active={newAlarmRepeat === 'custom'}
                    onclick={() => {
                      newAlarmRepeat = 'custom'
                    }}>Custom</button
                  >
                </div>
              </div>

              {#if newAlarmRepeat === 'custom'}
                <div class="custom-days-row">
                  {#each DAY_SHORT as dayLabel, idx (idx)}
                    <button
                      class="day-toggle-btn"
                      class:active={newAlarmDays.includes(idx)}
                      onclick={() => toggleDaySelection(idx)}
                      use:islandTooltip={DAY_NAMES[idx]}
                    >
                      {dayLabel}
                    </button>
                  {/each}
                </div>
              {/if}

              <div class="create-card-actions">
                <button
                  class="card-act-btn save"
                  onclick={handleSaveNewAlarm}
                  use:islandTooltip={'Save reminder'}
                >
                  <Check size={14} />
                  <span>Save</span>
                </button>
                <button
                  class="card-act-btn cancel"
                  onclick={cancelAlarmForm}
                  use:islandTooltip={'Cancel'}
                >
                  <X size={14} />
                  <span>Cancel</span>
                </button>
              </div>
            </div>
          {/if}

          <!-- Alarms list wrapped in Vinta ScrollView -->
          <div class="alarms-list-container">
            {#if $timerAppStore.alarms.length === 0 && !isNewAlarmOpen}
              <div class="alarms-empty-state">
                <Bell size={24} class="empty-icon" />
                <span class="empty-text">No active alarms</span>
                <button
                  class="add-first-alarm-btn"
                  onclick={() => {
                    isNewAlarmOpen = true
                  }}
                  use:islandTooltip={'Create reminder'}
                >
                  <Plus size={13} />
                  <span>Add Alarm</span>
                </button>
              </div>
            {:else}
              <ScrollView class="alarms-scroll-view">
                {#each $timerAppStore.alarms as alarm (alarm.id)}
                  <div class="alarm-item-card" class:disabled={!alarm.enabled}>
                    <button
                      class="alarm-item-info"
                      onclick={() => startEditAlarm(alarm)}
                      use:islandTooltip={'Click to edit alarm'}
                    >
                      <div class="alarm-time-row">
                        <span class="alarm-item-time"
                          >{formatAlarmDisplayTime(alarm.time, is24H)}</span
                        >
                        {#if alarm.repeat === 'once'}
                          <span class="alarm-repeat-badge">Once</span>
                        {:else if alarm.repeat === 'daily'}
                          <span class="alarm-repeat-badge">Every day</span>
                        {:else if alarm.repeat === 'weekdays'}
                          <div class="alarm-days-circles">
                            {#each [1, 2, 3, 4, 5] as d (d)}
                              <span class="alarm-day-circle"
                                >{DAY_SHORT[d]}</span
                              >
                            {/each}
                          </div>
                        {:else if alarm.repeat === 'custom' && alarm.days && alarm.days.length > 0}
                          <div class="alarm-days-circles">
                            {#each [...alarm.days].sort((a, b) => a - b) as d (d)}
                              <span class="alarm-day-circle"
                                >{DAY_SHORT[d]}</span
                              >
                            {/each}
                          </div>
                        {:else}
                          <span class="alarm-repeat-badge">Custom</span>
                        {/if}
                      </div>
                      <span class="alarm-item-label"
                        >{alarm.label || 'Reminder'}</span
                      >
                    </button>

                    <div class="alarm-item-actions">
                      <button
                        class="alarm-action-btn"
                        onclick={() => startEditAlarm(alarm)}
                        use:islandTooltip={'Edit alarm'}
                      >
                        <Pencil size={13} />
                      </button>

                      <label class="drawer-switch mini-switch">
                        <input
                          type="checkbox"
                          checked={alarm.enabled}
                          onchange={() => toggleAlarm(alarm.id)}
                        />
                        <span class="switch-slider"></span>
                      </label>

                      <button
                        class="alarm-delete-btn"
                        onclick={() => deleteAlarm(alarm.id)}
                        use:islandTooltip={'Delete alarm'}
                      >
                        <Trash2 size={13} />
                      </button>
                    </div>
                  </div>
                {/each}
              </ScrollView>
            {/if}
          </div>
        </div>
      {/if}

      <!-- Custom Timer Duration Drawer Overlay (Supports single-second precision) -->
      {#if isCustomDrawerOpen}
        <div class="drawer-overlay">
          <div class="drawer-header">
            <span class="drawer-title">Custom Duration</span>
            <button
              class="header-tool-btn"
              onclick={() => {
                isCustomDrawerOpen = false
              }}
              use:islandTooltip={'Close'}
            >
              <X size={15} />
            </button>
          </div>

          <div class="drawer-content">
            <!-- Steppers with exact single-unit buttons and editable inputs -->
            <div class="drawer-row">
              <span class="drawer-row-label">Hours</span>
              <div class="drawer-stepper">
                <button
                  class="stepper-btn"
                  onclick={() => {
                    customHours = Math.max(0, customHours - 1)
                  }}>-</button
                >
                <input
                  type="number"
                  min="0"
                  max="23"
                  class="stepper-input"
                  bind:value={customHours}
                  aria-label="Hours"
                />
                <button
                  class="stepper-btn"
                  onclick={() => {
                    customHours = Math.min(23, customHours + 1)
                  }}>+</button
                >
              </div>
            </div>

            <div class="drawer-row">
              <span class="drawer-row-label">Minutes</span>
              <div class="drawer-stepper">
                <button
                  class="stepper-btn"
                  onclick={() => {
                    customMinutes = Math.max(0, customMinutes - 1)
                  }}>-</button
                >
                <input
                  type="number"
                  min="0"
                  max="59"
                  class="stepper-input"
                  bind:value={customMinutes}
                  aria-label="Minutes"
                />
                <button
                  class="stepper-btn"
                  onclick={() => {
                    customMinutes = Math.min(59, customMinutes + 1)
                  }}>+</button
                >
              </div>
            </div>

            <div class="drawer-row">
              <span class="drawer-row-label">Seconds</span>
              <div class="drawer-stepper">
                <button
                  class="stepper-btn"
                  onclick={() => {
                    customSeconds = Math.max(0, customSeconds - 1)
                  }}>-</button
                >
                <input
                  type="number"
                  min="0"
                  max="59"
                  class="stepper-input"
                  bind:value={customSeconds}
                  aria-label="Seconds"
                />
                <button
                  class="stepper-btn"
                  onclick={() => {
                    customSeconds = Math.min(59, customSeconds + 1)
                  }}>+</button
                >
              </div>
            </div>

            <!-- Quick duration chips for instant precision -->
            <div class="quick-seconds-row">
              {#each QUICK_SECOND_PRESETS as qp (qp.secs)}
                <button
                  class="quick-sec-chip"
                  onclick={() => setQuickSecondPreset(qp.secs)}
                >
                  {qp.label}
                </button>
              {/each}
            </div>

            <button class="drawer-confirm-btn" onclick={applyCustomDuration}>
              <Check size={15} />
              <span>Set Duration</span>
            </button>
          </div>
        </div>
      {/if}

      <!-- Settings Drawer Overlay for Focus Sessions -->
      {#if isSettingsOpen}
        <div class="drawer-overlay">
          <div class="drawer-header">
            <span class="drawer-title">Focus Settings</span>
            <button
              class="header-tool-btn"
              onclick={() => {
                isSettingsOpen = false
              }}
              use:islandTooltip={'Close settings'}
            >
              <X size={15} />
            </button>
          </div>

          <div class="drawer-content">
            <!-- Pomodoro Interval Presets -->
            <div class="focus-presets-row">
              <button
                class="focus-preset-chip"
                class:active={$timerAppStore.focus.settings.focusDuration ===
                  25 && $timerAppStore.focus.settings.shortBreakDuration === 5}
                onclick={() =>
                  updateFocusSettings({
                    focusDuration: 25,
                    shortBreakDuration: 5
                  })}
                use:islandTooltip={'Standard Pomodoro (25m focus / 5m break)'}
              >
                25 / 5m
              </button>
              <button
                class="focus-preset-chip"
                class:active={$timerAppStore.focus.settings.focusDuration ===
                  45 && $timerAppStore.focus.settings.shortBreakDuration === 15}
                onclick={() =>
                  updateFocusSettings({
                    focusDuration: 45,
                    shortBreakDuration: 15
                  })}
                use:islandTooltip={'Deep Work (45m focus / 15m break)'}
              >
                45 / 15m
              </button>
              <button
                class="focus-preset-chip"
                class:active={$timerAppStore.focus.settings.focusDuration ===
                  50 && $timerAppStore.focus.settings.shortBreakDuration === 10}
                onclick={() =>
                  updateFocusSettings({
                    focusDuration: 50,
                    shortBreakDuration: 10
                  })}
                use:islandTooltip={'Extended Focus (50m focus / 10m break)'}
              >
                50 / 10m
              </button>
            </div>

            <div class="drawer-row">
              <span class="drawer-row-label">Focus Length</span>
              <div class="drawer-stepper">
                <button
                  class="stepper-btn"
                  onclick={() =>
                    updateFocusSettings({
                      focusDuration: Math.max(
                        1,
                        $timerAppStore.focus.settings.focusDuration - 5
                      )
                    })}>-</button
                >
                <span class="stepper-value"
                  >{$timerAppStore.focus.settings.focusDuration}m</span
                >
                <button
                  class="stepper-btn"
                  onclick={() =>
                    updateFocusSettings({
                      focusDuration: Math.min(
                        90,
                        $timerAppStore.focus.settings.focusDuration + 5
                      )
                    })}>+</button
                >
              </div>
            </div>

            <div class="drawer-row">
              <span class="drawer-row-label">Short Break</span>
              <div class="drawer-stepper">
                <button
                  class="stepper-btn"
                  onclick={() =>
                    updateFocusSettings({
                      shortBreakDuration: Math.max(
                        1,
                        $timerAppStore.focus.settings.shortBreakDuration - 1
                      )
                    })}>-</button
                >
                <span class="stepper-value"
                  >{$timerAppStore.focus.settings.shortBreakDuration}m</span
                >
                <button
                  class="stepper-btn"
                  onclick={() =>
                    updateFocusSettings({
                      shortBreakDuration: Math.min(
                        30,
                        $timerAppStore.focus.settings.shortBreakDuration + 1
                      )
                    })}>+</button
                >
              </div>
            </div>

            <div class="drawer-row">
              <span class="drawer-row-label">Long Break</span>
              <div class="drawer-stepper">
                <button
                  class="stepper-btn"
                  onclick={() =>
                    updateFocusSettings({
                      longBreakDuration: Math.max(
                        1,
                        $timerAppStore.focus.settings.longBreakDuration - 5
                      )
                    })}>-</button
                >
                <span class="stepper-value"
                  >{$timerAppStore.focus.settings.longBreakDuration}m</span
                >
                <button
                  class="stepper-btn"
                  onclick={() =>
                    updateFocusSettings({
                      longBreakDuration: Math.min(
                        60,
                        $timerAppStore.focus.settings.longBreakDuration + 5
                      )
                    })}>+</button
                >
              </div>
            </div>

            <div class="drawer-row">
              <span class="drawer-row-label">Long Break Cycle</span>
              <div class="drawer-stepper">
                <button
                  class="stepper-btn"
                  onclick={() =>
                    updateFocusSettings({
                      longBreakInterval: Math.max(
                        2,
                        $timerAppStore.focus.settings.longBreakInterval - 1
                      )
                    })}>-</button
                >
                <span class="stepper-value"
                  >{$timerAppStore.focus.settings.longBreakInterval}</span
                >
                <button
                  class="stepper-btn"
                  onclick={() =>
                    updateFocusSettings({
                      longBreakInterval: Math.min(
                        10,
                        $timerAppStore.focus.settings.longBreakInterval + 1
                      )
                    })}>+</button
                >
              </div>
            </div>

            <div class="drawer-row">
              <span class="drawer-row-label">Sound Chimes</span>
              <label class="drawer-switch">
                <input
                  type="checkbox"
                  checked={$timerAppStore.focus.settings.soundEnabled}
                  onchange={(e) =>
                    updateFocusSettings({
                      soundEnabled: (e.target as HTMLInputElement).checked
                    })}
                />
                <span class="switch-slider"></span>
              </label>
            </div>

            <div class="drawer-row">
              <span class="drawer-row-label">Auto-start Breaks</span>
              <label class="drawer-switch">
                <input
                  type="checkbox"
                  checked={$timerAppStore.focus.settings.autoStartBreaks}
                  onchange={(e) =>
                    updateFocusSettings({
                      autoStartBreaks: (e.target as HTMLInputElement).checked
                    })}
                />
                <span class="switch-slider"></span>
              </label>
            </div>

            <div class="drawer-row">
              <span class="drawer-row-label">Auto-start Focus</span>
              <label class="drawer-switch">
                <input
                  type="checkbox"
                  checked={$timerAppStore.focus.settings.autoStartFocus}
                  onchange={(e) =>
                    updateFocusSettings({
                      autoStartFocus: (e.target as HTMLInputElement).checked
                    })}
                />
                <span class="switch-slider"></span>
              </label>
            </div>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style src="./styles/TimerPage.css"></style>
