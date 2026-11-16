<!--
  @fileoverview Multi-function time management suite.

  Combines an interval countdown timer, millisecond-precision stopwatch with lap logging,
  Pomodoro focus sprint cycles, and scheduled audio reminder alarms.
-->
<svelte:options />

<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 42, r: 16 },
    expanded: { w: 420, h: 405, r: 28 },
    borderColor: 'var(--vinta-island-border-idle)',
    glowColor: 'var(--vinta-island-glow-idle)'
  }
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { Events } from '$lib/events'
  import {
    Hourglass,
    Clock,
    Flame,
    Bell,
    Play,
    Pause,
    RotateCcw,
    SlidersHorizontal,
    Plus,
    X
  } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import PageHeader from '../components/PageHeader.svelte'
  import DotPattern from '../components/patterns/DotPattern.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from '../components/HorizontalTabSelection.svelte'
  import CountdownTab from '../components/timer/CountdownTab.svelte'
  import StopwatchTab from '../components/timer/StopwatchTab.svelte'
  import FocusTab from '../components/timer/FocusTab.svelte'
  import AlarmsTab from '../components/timer/AlarmsTab.svelte'
  import CustomTimerDrawer from '../components/timer/CustomTimerDrawer.svelte'
  import FocusSettingsDrawer from '../components/timer/FocusSettingsDrawer.svelte'
  import AlarmEditorDrawer from '../components/timer/AlarmEditorDrawer.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    timerAppStore,
    timeFormatStore,
    tStore,
    setActiveTimerTab,
    toggleCountdown,
    resetCountdown,
    setCountdownDuration,
    toggleStopwatch,
    resetStopwatch,
    toggleFocus,
    resetFocus,
    addAlarm,
    updateAlarm,
    dismissRingingAlarm,
    snoozeRingingAlarm,
    type TimerAppTab,
    type AlarmItem,
    type AlarmRepeat
  } from '../store'

  let { expanded = false }: { expanded?: boolean } = $props()

  let isSettingsOpen = $state(false)
  let isCustomDrawerOpen = $state(false)
  let isNewAlarmOpen = $state(false)
  let editingAlarmId = $state<string | null>(null)
  let defaultCreateTime = $state('16:00')

  Events.On('devtest-open-new-alarm', (ev: any) => {
    const payload = ev?.data?.[0] ?? ev?.data ?? ev
    openCreateAlarm(payload?.time || '16:00')
  })

  /**
   * Checks whether user settings or locale use 24-hour time formatting.
   *
   * @returns True if 24-hour clock is active.
   */
  function computeIs24H(): boolean {
    return $timeFormatStore === '24h'
  }
  let is24H = $derived.by(computeIs24H)

  const DAY_SHORT = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

  /**
   * Formats total seconds into standard digital timer format (MM:SS or HH:MM:SS).
   *
   * @param totalSecs - Total elapsed or remaining seconds.
   * @returns Digital clock formatted string.
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
   * Formats millisecond duration into stopwatch precision digits (MM:SS.ms).
   *
   * @param totalMs - Total stopwatch duration in milliseconds.
   * @returns Stopwatch display string object with main and fractional components.
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
   * Formats alarm trigger time string into user-configured 12h or 24h clock format.
   *
   * @param timeStr - Raw time string (HH:MM).
   * @param use24H - Whether 24-hour mode is enabled.
   * @returns Formatted alarm time string.
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
   * Formats the trigger timestamp of a ringing alarm.
   *
   * @param timestamp - Epoch timestamp in milliseconds.
   * @param use24H - Whether 24-hour format is active.
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
   * Generates localized summary text for alarm recurrence schedules.
   *
   * @param alarm - Alarm record.
   * @returns Formatted recurrence description.
   */
  function formatAlarmRepeatBadge(alarm: AlarmItem): string {
    if (alarm.repeat === 'once') return $tStore('timer.once')
    if (alarm.repeat === 'daily') return $tStore('timer.every_day')
    if (alarm.repeat === 'weekdays') return $tStore('timer.weekdays')
    if (!alarm.days || alarm.days.length === 0) return $tStore('timer.custom')
    if (alarm.days.length === 7) return $tStore('timer.every_day')
    return alarm.days.map((d) => DAY_SHORT[d]).join(', ')
  }

  function computeActiveMinimalMode() {
    const s = $timerAppStore
    if (s.timer.isRunning) return 'timer'
    if (s.stopwatch.isRunning) return 'stopwatch'
    if (s.focus.isRunning) return 'focus'
    return s.activeTab === 'alarms' ? 'timer' : s.activeTab
  }
  let activeMinimalMode = $derived.by(computeActiveMinimalMode)

  function computeMinimalModeLabel() {
    if (activeMinimalMode === 'timer') return $tStore('timer.countdown_tab')
    if (activeMinimalMode === 'stopwatch') return $tStore('timer.stopwatch_tab')
    if ($timerAppStore.focus.phase === 'focus')
      return $tStore('timer.focus_tab')
    if ($timerAppStore.focus.phase === 'shortBreak')
      return $tStore('timer.break')
    return $tStore('timer.rest')
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

  /**
   * Toggles active countdown, stopwatch, or focus timer depending on active mode in minimal pill.
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
   * Resets active countdown, stopwatch, or focus timer depending on active mode in minimal pill.
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
   * Switches the active timer tab (countdown, stopwatch, focus, alarms).
   *
   * @param tab - Target tab identifier.
   */
  function handleTabSelect(tab: TimerAppTab): void {
    setActiveTimerTab(tab)
  }

  let timerTabs: TabItem<TimerAppTab>[] = $derived([
    {
      id: 'timer',
      label: $tStore('timer.countdown_tab'),
      icon: Hourglass
    },
    {
      id: 'stopwatch',
      label: $tStore('timer.stopwatch_tab'),
      icon: Clock
    },
    {
      id: 'focus',
      label: $tStore('timer.focus_tab'),
      icon: Flame
    },
    {
      id: 'alarms',
      label: $tStore('timer.alarms_tab'),
      icon: Bell
    }
  ])

  let editingAlarm = $derived(
    editingAlarmId
      ? $timerAppStore.alarms.find((a) => a.id === editingAlarmId)
      : null
  )

  /**
   * Opens the new alarm creation modal drawer.
   */
  function openCreateAlarm(initialTime?: string): void {
    editingAlarmId = null
    if (initialTime) {
      defaultCreateTime = initialTime
    }
    isNewAlarmOpen = true
  }

  /**
   * Opens the alarm editor modal populated with existing alarm properties.
   *
   * @param alarm - Alarm item record to edit.
   */
  function startEditAlarm(alarm: AlarmItem): void {
    editingAlarmId = alarm.id
    isNewAlarmOpen = true
  }

  /**
   * Closes active alarm modal drawer without saving changes.
   */
  function cancelAlarmForm(): void {
    editingAlarmId = null
    isNewAlarmOpen = false
  }

  /**
   * Commits created or modified alarm properties to the alarms store.
   *
   * @param payload - Alarm time, label, repeat schedule, and days array.
   */
  function handleSaveAlarm(payload: {
    time: string
    label: string
    repeat: AlarmRepeat
    days: number[]
  }): void {
    if (editingAlarmId) {
      updateAlarm(editingAlarmId, payload)
      editingAlarmId = null
    } else {
      addAlarm(payload.time, payload.label, payload.repeat, payload.days)
    }
    isNewAlarmOpen = false
  }

  /**
   * Applies a custom countdown duration from the custom timer drawer.
   *
   * @param totalSecs - Chosen duration in seconds.
   */
  function handleApplyCustomDuration(totalSecs: number): void {
    setCountdownDuration(totalSecs)
    isCustomDrawerOpen = false
  }

  function handleMinimizeEffect(): void {
    if (!expanded) {
      isSettingsOpen = false
      isCustomDrawerOpen = false
      isNewAlarmOpen = false
      editingAlarmId = null
    }
  }
  $effect(handleMinimizeEffect)

  /**
   * Handles Spacebar shortcut to toggle Play/Pause on the active timer/stopwatch/focus tab.
   *
   * @param e - Window keyboard event.
   */
  function handleTimerWindowKeyDown(e: KeyboardEvent): void {
    if (!expanded || isCustomDrawerOpen || isSettingsOpen || isNewAlarmOpen) {
      return
    }
    const target = e.target as HTMLElement | null
    if (
      target &&
      (target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.isContentEditable)
    ) {
      return
    }
    if (e.code === 'Space') {
      const tab = $timerAppStore.activeTab
      if (tab === 'timer') {
        e.preventDefault()
        toggleCountdown()
      } else if (tab === 'stopwatch') {
        e.preventDefault()
        toggleStopwatch()
      } else if (tab === 'focus') {
        e.preventDefault()
        toggleFocus()
      }
    }
  }
</script>

<svelte:window onkeydown={handleTimerWindowKeyDown} />

<div
  data-vinta="timer-page"
  class="timer-page"
  class:minimal-mode={!expanded}
  class:expanded-mode={expanded}
>
  {#if !expanded}
    {#if $timerAppStore.ringingAlarm}
      <div class="timer-minimal ringing-capsule">
        <div class="minimal-ringing-pulse" aria-hidden="true"></div>

        <div class="minimal-ringing-info">
          <div class="ringing-bell-badge">
            <Bell size={13} class="bell-shake" />
          </div>
          <div class="ringing-text-col">
            <span class="ringing-alarm-name">
              {$timerAppStore.ringingAlarm.label || 'Alarm Reminder'}
            </span>
            <span class="ringing-alarm-digits">
              {formatAlarmDisplayTime($timerAppStore.ringingAlarm.time, is24H)}
            </span>
          </div>
        </div>

        <div class="minimal-actions">
          <button
            class="mini-btn ringing-snooze-btn"
            onclick={(e: MouseEvent) => {
              e.stopPropagation()
              snoozeRingingAlarm()
            }}
            use:islandTooltip={'Snooze alarm'}
            aria-label="Snooze alarm"
          >
            <Clock size={12} />
          </button>
          <button
            class="mini-btn primary ringing-dismiss-btn"
            onclick={(e: MouseEvent) => {
              e.stopPropagation()
              dismissRingingAlarm()
            }}
            use:islandTooltip={'Dismiss alarm'}
            aria-label="Dismiss alarm"
          >
            <X size={13} />
          </button>
        </div>
      </div>
    {:else}
      <div class="timer-minimal">
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
    <div class="timer-expanded">
      <DotPattern
        variant="corner-tr"
        class="expanded-dots-bg {$timerAppStore.activeTab === 'timer'
          ? 'mode-timer'
          : $timerAppStore.activeTab === 'stopwatch'
            ? 'mode-stopwatch'
            : $timerAppStore.activeTab === 'focus' &&
                $timerAppStore.focus.phase === 'focus'
              ? 'mode-focus'
              : $timerAppStore.activeTab === 'focus'
                ? 'mode-break'
                : ''}"
      />

      <PageHeader>
        {#snippet center()}
          <HorizontalTabSelection
            tabs={timerTabs}
            value={$timerAppStore.activeTab}
            onchange={handleTabSelect}
            ariaLabel={$tStore('timer.title')}
          />
        {/snippet}

        {#snippet actions()}
          {#if $timerAppStore.activeTab === 'focus'}
            <button
              class="header-tool-btn"
              class:active={isSettingsOpen}
              onclick={() => {
                isSettingsOpen = !isSettingsOpen
              }}
              use:islandTooltip={'Focus settings'}
              aria-label="Focus settings"
            >
              <SlidersHorizontal size={14} />
            </button>
          {:else if $timerAppStore.activeTab === 'alarms'}
            <button
              class="header-tool-btn"
              data-action="create-alarm"
              class:active={isNewAlarmOpen}
              onclick={() => {
                if (isNewAlarmOpen) {
                  cancelAlarmForm()
                } else {
                  openCreateAlarm()
                }
              }}
              use:islandTooltip={$tStore('timer.new_alarm')}
              aria-label={$tStore('timer.new_alarm')}
            >
              <Plus size={15} />
            </button>
          {:else}
            <div class="header-spacer" aria-hidden="true"></div>
          {/if}
        {/snippet}
      </PageHeader>

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
            >
              <Clock size={12} />
              <span>{$tStore('timer.snooze')}</span>
            </button>
            <button
              class="banner-action-btn stop"
              onclick={dismissRingingAlarm}
            >
              <X size={12} />
              <span>{$tStore('timer.dismiss')}</span>
            </button>
          </div>
        </div>
      {/if}

      {#key $timerAppStore.activeTab}
        <div class="timer-tab-transition">
          {#if $timerAppStore.activeTab === 'timer'}
            <CountdownTab onOpenCustom={() => (isCustomDrawerOpen = true)} />
          {:else if $timerAppStore.activeTab === 'stopwatch'}
            <StopwatchTab />
          {:else if $timerAppStore.activeTab === 'focus'}
            <FocusTab />
          {:else if $timerAppStore.activeTab === 'alarms'}
            <AlarmsTab
              {is24H}
              onStartEdit={startEditAlarm}
              onOpenCreate={openCreateAlarm}
            />
          {/if}
        </div>
      {/key}

      <CustomTimerDrawer
        isOpen={isCustomDrawerOpen}
        initialTotalSeconds={$timerAppStore.timer.totalDurationSeconds}
        onClose={() => (isCustomDrawerOpen = false)}
        onApply={handleApplyCustomDuration}
      />

      <FocusSettingsDrawer
        isOpen={isSettingsOpen}
        onClose={() => (isSettingsOpen = false)}
      />

      <AlarmEditorDrawer
        isOpen={isNewAlarmOpen}
        isEditing={Boolean(editingAlarmId)}
        initialTime={editingAlarm?.time || defaultCreateTime || '16:00'}
        initialLabel={editingAlarm?.label || ''}
        initialRepeat={editingAlarm?.repeat || 'once'}
        initialDays={editingAlarm?.days || [1, 2, 3, 4, 5]}
        onSave={handleSaveAlarm}
        onCancel={cancelAlarmForm}
      />
    </div>
  {/if}
</div>

<style src="./styles/TimerPage.css"></style>
