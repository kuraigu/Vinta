import { writable, get } from 'svelte/store'
import {
  playFocusChime,
  playNotifChime,
  startAlarmRingingChime,
  stopAlarmRingingChime
} from '../services/audioService'
import { sendVintaNotification } from '../services/vintaNotification'
import { timeFormatStore } from './settings'

export type TimerAppTab = 'timer' | 'stopwatch' | 'focus' | 'alarms'

export type FocusPhase = 'focus' | 'shortBreak' | 'longBreak'

export type AlarmRepeat = 'once' | 'daily' | 'weekdays' | 'custom'

export interface AlarmItem {
  id: string
  time: string // HH:MM 24h
  label: string
  enabled: boolean
  repeat: AlarmRepeat
  days: number[] // 0 = Sun, 1 = Mon, ..., 6 = Sat
  lastTriggeredDate?: string
}

export interface FocusTimerSettings {
  focusDuration: number
  shortBreakDuration: number
  longBreakDuration: number
  longBreakInterval: number
  soundEnabled: boolean
  autoStartBreaks: boolean
  autoStartFocus: boolean
}

export interface LapRecord {
  id: number
  lapTimeMs: number
  overallTimeMs: number
}

export interface RingingAlarm {
  id: string
  alarmId: string
  time: string // Stored HH:MM 24h
  label: string
  repeat: AlarmRepeat
  days: number[]
  triggeredAt: number // epoch ms
}

export interface SnoozedAlarm {
  id: string
  alarmId: string
  time: string
  label: string
  repeat: AlarmRepeat
  days: number[]
  snoozeUntil: number // epoch ms
}

export interface TimerAppState {
  activeTab: TimerAppTab

  // Timer countdown
  timer: {
    remainingSeconds: number
    totalDurationSeconds: number
    isRunning: boolean
  }

  // Stopwatch
  stopwatch: {
    elapsedMs: number
    isRunning: boolean
    laps: LapRecord[]
  }

  // Focus / Pomodoro
  focus: {
    phase: FocusPhase
    isRunning: boolean
    timeRemaining: number
    totalDuration: number
    currentCycle: number
    completedSessions: number
    settings: FocusTimerSettings
  }

  // Alarms / Reminders
  alarms: AlarmItem[]
  ringingAlarm: RingingAlarm | null
  snoozedAlarms: SnoozedAlarm[]
}

const DEFAULT_FOCUS_SETTINGS: FocusTimerSettings = {
  focusDuration: 25,
  shortBreakDuration: 5,
  longBreakDuration: 15,
  longBreakInterval: 4,
  soundEnabled: true,
  autoStartBreaks: false,
  autoStartFocus: false
}

const FOCUS_SETTINGS_KEY = 'vinta_focus_settings'
const FOCUS_STATS_KEY = 'vinta_focus_stats'
const ALARMS_STORAGE_KEY = 'vinta_alarms'

/**
 * Loads stored settings from local storage.
 *
 * @returns Parsed user settings or defaults.
 */
function loadFocusSettings(): FocusTimerSettings {
  if (typeof localStorage === 'undefined') return { ...DEFAULT_FOCUS_SETTINGS }
  try {
    const raw = localStorage.getItem(FOCUS_SETTINGS_KEY)
    if (!raw) return { ...DEFAULT_FOCUS_SETTINGS }
    const parsed = JSON.parse(raw)
    return { ...DEFAULT_FOCUS_SETTINGS, ...parsed }
  } catch {
    return { ...DEFAULT_FOCUS_SETTINGS }
  }
}

/**
 * Loads total completed focus sessions from storage.
 *
 * @returns Completed sessions count.
 */
function loadCompletedSessions(): number {
  if (typeof localStorage === 'undefined') return 0
  try {
    const raw = localStorage.getItem(FOCUS_STATS_KEY)
    if (!raw) return 0
    const count = parseInt(raw, 10)
    return Number.isNaN(count) ? 0 : count
  } catch {
    return 0
  }
}

/**
 * Loads stored alarms from local storage.
 *
 * @returns Parsed user alarms or default list.
 */
function loadAlarms(): AlarmItem[] {
  if (typeof localStorage === 'undefined') return []
  try {
    const raw = localStorage.getItem(ALARMS_STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed.map((item) => ({
      ...item,
      repeat: item.repeat || 'once',
      days: Array.isArray(item.days) ? item.days : []
    }))
  } catch {
    return []
  }
}

/**
 * Persists updated alarms to local storage.
 *
 * @param alarms List of alarm items.
 */
function persistAlarms(alarms: AlarmItem[]): void {
  if (typeof localStorage === 'undefined') return
  try {
    localStorage.setItem(ALARMS_STORAGE_KEY, JSON.stringify(alarms))
  } catch {
    // Ignore storage failure
  }
}

const initialFocusSettings = loadFocusSettings()
const initialFocusDuration = initialFocusSettings.focusDuration * 60

const initialState: TimerAppState = {
  activeTab: 'timer',
  timer: {
    remainingSeconds: 300,
    totalDurationSeconds: 300,
    isRunning: false
  },
  stopwatch: {
    elapsedMs: 0,
    isRunning: false,
    laps: []
  },
  focus: {
    phase: 'focus',
    isRunning: false,
    timeRemaining: initialFocusDuration,
    totalDuration: initialFocusDuration,
    currentCycle: 1,
    completedSessions: loadCompletedSessions(),
    settings: initialFocusSettings
  },
  alarms: loadAlarms(),
  ringingAlarm: null,
  snoozedAlarms: []
}

export const timerAppStore = writable<TimerAppState>(initialState)

// Interval & timestamp tracking
let timerInterval: ReturnType<typeof setInterval> | null = null
let timerTargetEnd: number | null = null

let stopwatchInterval: ReturnType<typeof setInterval> | null = null
let stopwatchStartTime: number | null = null
let stopwatchBaseElapsed = 0

let focusInterval: ReturnType<typeof setInterval> | null = null
let focusTargetEnd: number | null = null

/**
 * Persists updated focus settings.
 *
 * @param settings Target configuration to save.
 */
function persistFocusSettings(settings: FocusTimerSettings): void {
  if (typeof localStorage === 'undefined') return
  try {
    localStorage.setItem(FOCUS_SETTINGS_KEY, JSON.stringify(settings))
  } catch {
    // Ignore storage failure
  }
}

/**
 * Dispatches a desktop notification toast.
 *
 * @param title Notification title.
 * @param message Notification body.
 */
function sendToastNotification(title: string, message: string): void {
  sendVintaNotification({
    title,
    message,
    subApp: 'Timer'
  })
}

// ============================================================================
// Active Tab Switcher
// ============================================================================

/**
 * Switches the active sub-utility tab.
 *
 * @param tab Target utility tab.
 */
export function setActiveTimerTab(tab: TimerAppTab): void {
  timerAppStore.update((s) => ({ ...s, activeTab: tab }))
}

// ============================================================================
// Countdown Timer Logic
// ============================================================================

/**
 * Timer tick handler.
 */
function handleCountdownTick(): void {
  if (!timerTargetEnd) return
  const now = Date.now()
  const remaining = Math.max(0, Math.ceil((timerTargetEnd - now) / 1000))

  timerAppStore.update((s) => ({
    ...s,
    timer: {
      ...s.timer,
      remainingSeconds: remaining
    }
  }))

  if (remaining <= 0) {
    pauseCountdown()
    playNotifChime(0.85)
    sendToastNotification('Timer Expired', 'Your countdown timer has completed.')
  }
}

/**
 * Starts the countdown timer.
 */
export function startCountdown(): void {
  const state = get(timerAppStore)
  if (state.timer.isRunning || state.timer.remainingSeconds <= 0) return

  timerTargetEnd = Date.now() + state.timer.remainingSeconds * 1000
  if (timerInterval) clearInterval(timerInterval)
  timerInterval = setInterval(handleCountdownTick, 250)

  timerAppStore.update((s) => ({
    ...s,
    timer: { ...s.timer, isRunning: true }
  }))
}

/**
 * Pauses the countdown timer.
 */
export function pauseCountdown(): void {
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }
  timerTargetEnd = null

  timerAppStore.update((s) => ({
    ...s,
    timer: { ...s.timer, isRunning: false }
  }))
}

/**
 * Toggles the countdown timer between running and paused.
 */
export function toggleCountdown(): void {
  const state = get(timerAppStore)
  if (state.timer.isRunning) {
    pauseCountdown()
  } else {
    startCountdown()
  }
}

/**
 * Resets the countdown timer to its configured total duration.
 */
export function resetCountdown(): void {
  pauseCountdown()
  timerAppStore.update((s) => ({
    ...s,
    timer: {
      ...s.timer,
      remainingSeconds: s.timer.totalDurationSeconds
    }
  }))
}

/**
 * Sets the countdown timer to a specified duration in seconds.
 *
 * @param seconds Duration in seconds.
 */
export function setCountdownDuration(seconds: number): void {
  pauseCountdown()
  const clamped = Math.max(1, seconds)
  timerAppStore.update((s) => ({
    ...s,
    timer: {
      remainingSeconds: clamped,
      totalDurationSeconds: clamped,
      isRunning: false
    }
  }))
}

// ============================================================================
// Stopwatch Logic
// ============================================================================

/**
 * Stopwatch tick handler.
 */
function handleStopwatchTick(): void {
  if (!stopwatchStartTime) return
  const now = Date.now()
  const elapsed = stopwatchBaseElapsed + (now - stopwatchStartTime)

  timerAppStore.update((s) => ({
    ...s,
    stopwatch: {
      ...s.stopwatch,
      elapsedMs: elapsed
    }
  }))
}

/**
 * Starts the stopwatch.
 */
export function startStopwatch(): void {
  const state = get(timerAppStore)
  if (state.stopwatch.isRunning) return

  stopwatchStartTime = Date.now()
  if (stopwatchInterval) clearInterval(stopwatchInterval)
  stopwatchInterval = setInterval(handleStopwatchTick, 35)

  timerAppStore.update((s) => ({
    ...s,
    stopwatch: { ...s.stopwatch, isRunning: true }
  }))
}

/**
 * Pauses the stopwatch.
 */
export function pauseStopwatch(): void {
  if (stopwatchInterval) {
    clearInterval(stopwatchInterval)
    stopwatchInterval = null
  }
  if (stopwatchStartTime) {
    stopwatchBaseElapsed += Date.now() - stopwatchStartTime
    stopwatchStartTime = null
  }

  timerAppStore.update((s) => ({
    ...s,
    stopwatch: { ...s.stopwatch, isRunning: false }
  }))
}

/**
 * Toggles the stopwatch between running and paused.
 */
export function toggleStopwatch(): void {
  const state = get(timerAppStore)
  if (state.stopwatch.isRunning) {
    pauseStopwatch()
  } else {
    startStopwatch()
  }
}

/**
 * Resets the stopwatch to zero.
 */
export function resetStopwatch(): void {
  pauseStopwatch()
  stopwatchBaseElapsed = 0
  stopwatchStartTime = null

  timerAppStore.update((s) => ({
    ...s,
    stopwatch: {
      elapsedMs: 0,
      isRunning: false,
      laps: []
    }
  }))
}

/**
 * Records a lap split time in the stopwatch.
 */
export function recordLap(): void {
  const state = get(timerAppStore)
  const currentTotal = state.stopwatch.elapsedMs
  const previousLaps = state.stopwatch.laps
  const previousTotal =
    previousLaps.length > 0 ? previousLaps[0].overallTimeMs : 0
  const lapTime = Math.max(0, currentTotal - previousTotal)

  const newLap: LapRecord = {
    id: previousLaps.length + 1,
    lapTimeMs: lapTime,
    overallTimeMs: currentTotal
  }

  timerAppStore.update((s) => ({
    ...s,
    stopwatch: {
      ...s.stopwatch,
      laps: [newLap, ...s.stopwatch.laps]
    }
  }))
}

// ============================================================================
// Focus / Pomodoro Logic
// ============================================================================

/**
 * Calculates phase duration in seconds.
 *
 * @param phase Phase identifier.
 * @param settings Configured durations.
 * @returns Phase length in seconds.
 */
function getFocusDurationSeconds(
  phase: FocusPhase,
  settings: FocusTimerSettings
): number {
  if (phase === 'shortBreak') return Math.max(1, settings.shortBreakDuration) * 60
  if (phase === 'longBreak') return Math.max(1, settings.longBreakDuration) * 60
  return Math.max(1, settings.focusDuration) * 60
}

/**
 * Focus timer tick handler.
 */
function handleFocusTick(): void {
  if (!focusTargetEnd) return
  const now = Date.now()
  const remaining = Math.max(0, Math.ceil((focusTargetEnd - now) / 1000))

  timerAppStore.update((s) => ({
    ...s,
    focus: {
      ...s.focus,
      timeRemaining: remaining
    }
  }))

  if (remaining <= 0) {
    handleFocusCompletion()
  }
}

/**
 * Handles focus phase transition upon countdown completion.
 */
function handleFocusCompletion(): void {
  pauseFocus()
  const state = get(timerAppStore)
  const currentPhase = state.focus.phase
  const settings = state.focus.settings

  if (currentPhase === 'focus') {
    const updatedCompleted = state.focus.completedSessions + 1
    if (typeof localStorage !== 'undefined') {
      try {
        localStorage.setItem(FOCUS_STATS_KEY, updatedCompleted.toString())
      } catch {}
    }

    if (settings.soundEnabled) {
      playFocusChime('focusComplete')
    }

    const isLongBreakDue = state.focus.currentCycle >= settings.longBreakInterval
    const nextPhase: FocusPhase = isLongBreakDue ? 'longBreak' : 'shortBreak'
    const nextDuration = getFocusDurationSeconds(nextPhase, settings)

    sendToastNotification(
      'Focus Complete',
      isLongBreakDue
        ? 'Great sprint! Time for an extended rest.'
        : 'Sprint finished! Enjoy a short breather.'
    )

    timerAppStore.update((s) => ({
      ...s,
      focus: {
        ...s.focus,
        phase: nextPhase,
        timeRemaining: nextDuration,
        totalDuration: nextDuration,
        completedSessions: updatedCompleted
      }
    }))

    if (settings.autoStartBreaks) {
      startFocus()
    }
  } else {
    if (settings.soundEnabled) {
      playFocusChime('breakComplete')
    }

    const nextCycle =
      currentPhase === 'longBreak' ? 1 : state.focus.currentCycle + 1
    const nextDuration = getFocusDurationSeconds('focus', settings)

    sendToastNotification('Break Over', 'Ready for the next sprint?')

    timerAppStore.update((s) => ({
      ...s,
      focus: {
        ...s.focus,
        phase: 'focus',
        currentCycle: nextCycle,
        timeRemaining: nextDuration,
        totalDuration: nextDuration
      }
    }))

    if (settings.autoStartFocus) {
      startFocus()
    }
  }
}

/**
 * Starts the focus countdown.
 */
export function startFocus(): void {
  const state = get(timerAppStore)
  if (state.focus.isRunning) return

  focusTargetEnd = Date.now() + state.focus.timeRemaining * 1000
  if (focusInterval) clearInterval(focusInterval)
  focusInterval = setInterval(handleFocusTick, 250)

  timerAppStore.update((s) => ({
    ...s,
    focus: { ...s.focus, isRunning: true }
  }))
}

/**
 * Pauses the focus countdown.
 */
export function pauseFocus(): void {
  if (focusInterval) {
    clearInterval(focusInterval)
    focusInterval = null
  }
  focusTargetEnd = null

  timerAppStore.update((s) => ({
    ...s,
    focus: { ...s.focus, isRunning: false }
  }))
}

/**
 * Toggles the focus countdown.
 */
export function toggleFocus(): void {
  const state = get(timerAppStore)
  if (state.focus.isRunning) {
    pauseFocus()
  } else {
    startFocus()
  }
}

/**
 * Resets the focus countdown to the configured duration of the active phase.
 */
export function resetFocus(): void {
  pauseFocus()
  timerAppStore.update((s) => {
    const duration = getFocusDurationSeconds(s.focus.phase, s.focus.settings)
    return {
      ...s,
      focus: {
        ...s.focus,
        timeRemaining: duration,
        totalDuration: duration
      }
    }
  })
}

/**
 * Skips the active focus session and advances the cycle count.
 */
export function skipFocusSession(): void {
  pauseFocus()
  const state = get(timerAppStore)
  const maxCycles = state.focus.settings.longBreakInterval || 4
  const nextCycle =
    state.focus.currentCycle >= maxCycles ? 1 : state.focus.currentCycle + 1
  const duration = getFocusDurationSeconds('focus', state.focus.settings)

  timerAppStore.update((s) => ({
    ...s,
    focus: {
      ...s.focus,
      phase: 'focus',
      currentCycle: nextCycle,
      timeRemaining: duration,
      totalDuration: duration
    }
  }))
}

/**
 * Selects a focus phase directly.
 *
 * @param phase Target phase.
 */
export function setFocusPhase(phase: FocusPhase): void {
  pauseFocus()
  timerAppStore.update((s) => {
    const duration = getFocusDurationSeconds(phase, s.focus.settings)
    return {
      ...s,
      focus: {
        ...s.focus,
        phase,
        timeRemaining: duration,
        totalDuration: duration
      }
    }
  })
}

/**
 * Updates focus timer duration settings.
 *
 * @param updates Partial settings updates.
 */
export function updateFocusSettings(
  updates: Partial<FocusTimerSettings>
): void {
  timerAppStore.update((s) => {
    const newSettings = { ...s.focus.settings, ...updates }
    persistFocusSettings(newSettings)

    let remaining = s.focus.timeRemaining
    let total = s.focus.totalDuration

    if (!s.focus.isRunning) {
      total = getFocusDurationSeconds(s.focus.phase, newSettings)
      remaining = total
    }

    return {
      ...s,
      focus: {
        ...s.focus,
        settings: newSettings,
        timeRemaining: remaining,
        totalDuration: total
      }
    }
  })
}

// ============================================================================
// Alarms & Reminders Logic
// ============================================================================

/**
 * Adds a new scheduled alarm reminder.
 *
 * @param time Time in HH:MM format.
 * @param label Optional note label.
 * @param repeat Recurrence rule.
 * @param days Active days of the week (0 = Sun .. 6 = Sat).
 */
export function addAlarm(
  time: string,
  label: string,
  repeat: AlarmRepeat = 'once',
  days: number[] = []
): void {
  const computedDays =
    repeat === 'daily'
      ? [0, 1, 2, 3, 4, 5, 6]
      : repeat === 'weekdays'
        ? [1, 2, 3, 4, 5]
        : days

  const newAlarm: AlarmItem = {
    id: `alarm_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
    time,
    label: label.trim() || 'Reminder',
    enabled: true,
    repeat,
    days: computedDays
  }

  timerAppStore.update((s) => {
    const updated = [...s.alarms, newAlarm].sort((a, b) =>
      a.time.localeCompare(b.time)
    )
    persistAlarms(updated)
    return { ...s, alarms: updated }
  })
}

/**
 * Toggles an alarm's active state.
 *
 * @param id Alarm identifier.
 */
export function toggleAlarm(id: string): void {
  timerAppStore.update((s) => {
    const updated = s.alarms.map((item) =>
      item.id === id ? { ...item, enabled: !item.enabled } : item
    )
    persistAlarms(updated)
    return { ...s, alarms: updated }
  })
}

/**
 * Removes an alarm.
 *
 * @param id Alarm identifier.
 */
export function deleteAlarm(id: string): void {
  timerAppStore.update((s) => {
    const updated = s.alarms.filter((item) => item.id !== id)
    persistAlarms(updated)
    return { ...s, alarms: updated }
  })
}

/**
 * Updates an alarm's details.
 *
 * @param id Alarm identifier.
 * @param updates Partial updates.
 */
export function updateAlarm(
  id: string,
  updates: Partial<Omit<AlarmItem, 'id'>>
): void {
  timerAppStore.update((s) => {
    const updated = s.alarms.map((item) =>
      item.id === id ? { ...item, ...updates } : item
    )
    persistAlarms(updated)
    return { ...s, alarms: updated }
  })
}

/**
 * Formats a 24-hour HH:MM time string according to user 12-hour or 24-hour preference.
 *
 * @param timeStr Stored HH:MM string.
 * @param use24H Whether 24-hour format is preferred.
 * @returns Formatted time string.
 */
function formatTimeBySetting(timeStr: string, use24H: boolean): string {
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
 * Returns human-readable repetition description for an alarm reminder.
 *
 * @param repeat Recurrence rule.
 * @param days Active day numbers.
 * @returns Human-readable recurrence string.
 */
function formatAlarmRepeatDescription(
  repeat: AlarmRepeat,
  days: number[] = []
): string {
  if (repeat === 'once') return 'One-time reminder'
  if (repeat === 'daily') return 'Every day'
  if (repeat === 'weekdays') return 'Mon - Fri'
  if (!days || days.length === 0) return 'Custom days'
  if (days.length === 7) return 'Every day'
  const names = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  return days.map((d) => names[d]).join(', ')
}

/**
 * Checks wall clock against registered and snoozed alarms.
 */
function checkAlarmsClock(): void {
  const now = new Date()
  const currentDay = now.getDay() // 0 to 6
  const hrs = now.getHours().toString().padStart(2, '0')
  const mins = now.getMinutes().toString().padStart(2, '0')
  const currentTime = `${hrs}:${mins}`
  const todayStr = `${now.getFullYear()}-${(now.getMonth() + 1).toString().padStart(2, '0')}-${now.getDate().toString().padStart(2, '0')}`
  const nowTimestamp = now.getTime()

  const state = get(timerAppStore)

  // 1. Check snoozed alarms ready to ring again
  const pendingSnoozes = state.snoozedAlarms.filter(
    (s) => nowTimestamp >= s.snoozeUntil
  )
  if (pendingSnoozes.length > 0) {
    const target = pendingSnoozes[0]
    const remainingSnoozes = state.snoozedAlarms.filter(
      (s) => s.id !== target.id
    )

    const ringing: RingingAlarm = {
      id: `ring_${Date.now()}`,
      alarmId: target.alarmId,
      time: target.time,
      label: target.label,
      repeat: target.repeat,
      days: target.days,
      triggeredAt: nowTimestamp
    }

    startAlarmRingingChime()

    const use24H = get(timeFormatStore) === '24h'
    const scheduledDisplay = formatTimeBySetting(target.time, use24H)
    const sentDisplay = formatTimeBySetting(currentTime, use24H)
    const repeatDesc = formatAlarmRepeatDescription(target.repeat, target.days)

    sendToastNotification(
      `Alarm: ${target.label} • ${scheduledDisplay}`,
      `Scheduled for ${scheduledDisplay} (Triggered at ${sentDisplay}) • ${repeatDesc}`
    )

    timerAppStore.update((s) => ({
      ...s,
      ringingAlarm: ringing,
      snoozedAlarms: remainingSnoozes
    }))
    return
  }

  // 2. Check scheduled registered alarms
  let changed = false
  let newlyRinging: RingingAlarm | null = null

  const updatedAlarms = state.alarms.map((alarm) => {
    if (
      !alarm.enabled ||
      alarm.time !== currentTime ||
      alarm.lastTriggeredDate === todayStr
    ) {
      return alarm
    }

    let shouldRing = false
    if (alarm.repeat === 'once' || alarm.repeat === 'daily') {
      shouldRing = true
    } else if (alarm.repeat === 'weekdays') {
      shouldRing = currentDay >= 1 && currentDay <= 5
    } else if (alarm.repeat === 'custom') {
      shouldRing = Array.isArray(alarm.days) && alarm.days.includes(currentDay)
    }

    if (!shouldRing) {
      return alarm
    }

    changed = true
    if (!newlyRinging && !state.ringingAlarm) {
      newlyRinging = {
        id: `ring_${Date.now()}`,
        alarmId: alarm.id,
        time: alarm.time,
        label: alarm.label || 'Reminder',
        repeat: alarm.repeat,
        days: alarm.days || [],
        triggeredAt: nowTimestamp
      }
    }

    const nextEnabled = alarm.repeat === 'once' ? false : alarm.enabled

    return {
      ...alarm,
      enabled: nextEnabled,
      lastTriggeredDate: todayStr
    }
  })

  if (newlyRinging) {
    const ringingItem = newlyRinging as RingingAlarm
    startAlarmRingingChime()

    const use24H = get(timeFormatStore) === '24h'
    const scheduledDisplay = formatTimeBySetting(ringingItem.time, use24H)
    const sentDisplay = formatTimeBySetting(currentTime, use24H)
    const repeatDesc = formatAlarmRepeatDescription(
      ringingItem.repeat,
      ringingItem.days
    )

    sendToastNotification(
      `Alarm: ${ringingItem.label} • ${scheduledDisplay}`,
      `Scheduled for ${scheduledDisplay} (Triggered at ${sentDisplay}) • ${repeatDesc}`
    )
  }

  if (changed || newlyRinging) {
    if (changed) {
      persistAlarms(updatedAlarms)
    }
    timerAppStore.update((s) => ({
      ...s,
      alarms: updatedAlarms,
      ringingAlarm: newlyRinging || s.ringingAlarm
    }))
  }
}

// Start alarms background check every 3 seconds
if (typeof setInterval !== 'undefined') {
  setInterval(checkAlarmsClock, 3000)
}

/**
 * Dismisses and stops the actively ringing alarm.
 */
export function dismissRingingAlarm(): void {
  stopAlarmRingingChime()
  timerAppStore.update((s) => ({ ...s, ringingAlarm: null }))
}

/**
 * Snoozes the actively ringing alarm for a designated duration.
 *
 * @param minutes Snooze duration in minutes (defaults to 5 minutes).
 */
export function snoozeRingingAlarm(minutes = 5): void {
  stopAlarmRingingChime()
  timerAppStore.update((s) => {
    const current = s.ringingAlarm
    if (!current) return s

    const snoozed: SnoozedAlarm = {
      id: `snooze_${Date.now()}`,
      alarmId: current.alarmId,
      time: current.time,
      label: current.label,
      repeat: current.repeat,
      days: current.days,
      snoozeUntil: Date.now() + Math.max(1, minutes) * 60 * 1000
    }

    sendToastNotification(
      'Alarm Snoozed',
      `"${current.label}" will ring again in ${minutes} minutes.`
    )

    return {
      ...s,
      ringingAlarm: null,
      snoozedAlarms: [...s.snoozedAlarms, snoozed]
    }
  })
}

// ============================================================================
// Global Query Helpers
// ============================================================================

/**
 * Returns whether any timer, stopwatch, focus sprint, or alarm is active.
 *
 * @param state Snapshot of timer app state.
 * @returns True if any countdown, stopwatch, or ringing alarm is active.
 */
export function isAnyTimerRunning(state: TimerAppState): boolean {
  return (
    state.timer.isRunning ||
    state.stopwatch.isRunning ||
    state.focus.isRunning ||
    Boolean(state.ringingAlarm)
  )
}

