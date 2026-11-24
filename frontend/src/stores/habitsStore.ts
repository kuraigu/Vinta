/**
 * @fileoverview Daily habits and micro-routine reminders store persisted in %APPDATA%/FreeMatrix Studios/Vinta/habits.json.
 */

import { writable, derived, get } from 'svelte/store'
import {
  GetHabitsData,
  SaveHabitsData
} from '../../bindings/Vinta/backend/app.js'

export interface HabitItem {
  id: string
  title: string
  reminderMinutes: number
  streak: number
  lastCompletedDate: string
  lastRemindedAt: number
}

export interface HabitsState {
  habits: HabitItem[]
  initialized: boolean
}

const STORAGE_FALLBACK_KEY = 'vinta_habits_fallback'

export function getTodayDateKey(date = new Date()): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function getYesterdayDateKey(): string {
  const dt = new Date()
  dt.setDate(dt.getDate() - 1)
  return getTodayDateKey(dt)
}

function createStarterHabits(): HabitItem[] {
  return [
    {
      id: 'habit-hydrate',
      title: 'Drink Water',
      reminderMinutes: 60,
      streak: 0,
      lastCompletedDate: '',
      lastRemindedAt: Date.now()
    },
    {
      id: 'habit-posture',
      title: '20-20-20 Eye & Posture Break',
      reminderMinutes: 0,
      streak: 0,
      lastCompletedDate: '',
      lastRemindedAt: 0
    }
  ]
}

export const habitsStore = writable<HabitsState>({
  habits: [],
  initialized: false
})

export const habitsProgress = derived(habitsStore, ($state) => {
  const today = getTodayDateKey()
  const total = $state.habits.length
  const done = $state.habits.filter((h) => h.lastCompletedDate === today).length
  return { done, total }
})

let saveTimer: ReturnType<typeof setTimeout> | null = null

function persistHabits(state: HabitsState): void {
  if (!state.initialized) return
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    const payload = JSON.stringify({ habits: state.habits })
    try {
      localStorage.setItem(STORAGE_FALLBACK_KEY, payload)
    } catch {
      // ignore
    }
    if (typeof SaveHabitsData === 'function') {
      void SaveHabitsData(payload).catch(() => {})
    }
  }, 250)
}

/**
 * Loads habits once from habits.json (with localStorage fallback).
 */
export async function initHabitsStore(): Promise<void> {
  if (get(habitsStore).initialized) return

  let raw = ''
  try {
    if (typeof GetHabitsData === 'function') {
      raw = await GetHabitsData()
    }
  } catch {
    // fallback below
  }

  if (!raw || raw === '{}') {
    try {
      raw = localStorage.getItem(STORAGE_FALLBACK_KEY) || ''
    } catch {
      raw = ''
    }
  }

  let habits: HabitItem[] = []
  if (raw && raw !== '{}') {
    try {
      const parsed = JSON.parse(raw) as { habits?: HabitItem[] }
      if (Array.isArray(parsed.habits)) {
        habits = parsed.habits
      }
    } catch {
      // ignore
    }
  }

  if (habits.length === 0) {
    habits = createStarterHabits()
  }

  const nextState: HabitsState = { habits, initialized: true }
  habitsStore.set(nextState)
  persistHabits(nextState)
}

function updateAndPersist(updater: (prev: HabitsState) => HabitsState): void {
  habitsStore.update((prev) => {
    const next = updater(prev)
    persistHabits(next)
    return next
  })
}

export function addHabit(title: string, reminderMinutes = 0): void {
  const trimmed = title.trim()
  if (!trimmed) return
  updateAndPersist((prev) => ({
    ...prev,
    habits: [
      ...prev.habits,
      {
        id: `habit-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
        title: trimmed,
        reminderMinutes,
        streak: 0,
        lastCompletedDate: '',
        lastRemindedAt: reminderMinutes > 0 ? Date.now() : 0
      }
    ]
  }))
}

export function deleteHabit(id: string): void {
  updateAndPersist((prev) => ({
    ...prev,
    habits: prev.habits.filter((h) => h.id !== id)
  }))
}

export function toggleHabitToday(id: string): void {
  const today = getTodayDateKey()
  const yesterday = getYesterdayDateKey()
  updateAndPersist((prev) => ({
    ...prev,
    habits: prev.habits.map((h) => {
      if (h.id !== id) return h
      if (h.lastCompletedDate === today) {
        return {
          ...h,
          lastCompletedDate: '',
          streak: Math.max(0, h.streak - 1)
        }
      }
      const nextStreak =
        h.lastCompletedDate === yesterday ? h.streak + 1 : Math.max(1, h.streak)
      return {
        ...h,
        lastCompletedDate: today,
        streak: nextStreak
      }
    })
  }))
}

/**
 * Checks due habit reminders piggybacking on the existing 60-second clock ticker.
 * Returns habit titles whose reminder interval has elapsed so callers can notify without extra loops.
 */
export function collectDueHabitReminders(now = Date.now()): HabitItem[] {
  const state = get(habitsStore)
  if (!state.initialized) return []
  const today = getTodayDateKey(new Date(now))
  const due: HabitItem[] = []

  for (const h of state.habits) {
    if (h.reminderMinutes <= 0) continue
    if (h.lastCompletedDate === today) continue
    const intervalMs = h.reminderMinutes * 60 * 1000
    if (now - (h.lastRemindedAt || 0) >= intervalMs) {
      due.push(h)
    }
  }

  if (due.length > 0) {
    const dueIds = new Set(due.map((d) => d.id))
    updateAndPersist((prev) => ({
      ...prev,
      habits: prev.habits.map((h) =>
        dueIds.has(h.id) ? { ...h, lastRemindedAt: now } : h
      )
    }))
  }

  return due
}
