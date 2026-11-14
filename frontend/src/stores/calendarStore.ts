/**
 * @fileoverview Manages calendar events and remote iCal feed subscriptions.
 *
 * Stores local events, fetches remote calendar subscriptions, triggers reminder
 * notifications, and automatically removes past events when configured.
 */

import { writable, get } from 'svelte/store'
import {
  GetCalendarData,
  SaveCalendarData
} from '../../bindings/Vinta/backend/app.js'
import { sendVintaNotification } from '../services/vintaNotification'
import {
  autoDiscardPastEventsStore,
  pageVisibilityStore,
  homeWidgetsVisibilityStore
} from './settings'

export interface CalendarEvent {
  id: string
  title: string
  description?: string
  date: string // Format: 'YYYY-MM-DD'
  startTime: string // Format: 'HH:mm' or empty for all-day
  endTime?: string // Format: 'HH:mm'
  isAllDay?: boolean
  color: 'indigo' | 'emerald' | 'amber' | 'purple' | 'rose' | 'sky'
  locationUrl?: string
  location?: string
  source: 'local' | 'synced'
  feedId?: string
  notified?: boolean
}

export interface CalendarFeed {
  id: string
  name: string
  url: string
  color: 'indigo' | 'emerald' | 'amber' | 'purple' | 'rose' | 'sky'
  lastSyncedAt?: number
  enabled: boolean
}

export interface CalendarAppState {
  events: CalendarEvent[]
  feeds: CalendarFeed[]
  selectedDate: string // 'YYYY-MM-DD'
  isSyncing: boolean
  lastSyncError?: string
}

export interface CalendarPendingAction {
  action: 'add'
  date?: string
}

export const calendarPendingActionStore = writable<CalendarPendingAction | null>(null)

const STORAGE_KEY_EVENTS = 'vinta_calendar_events'
const STORAGE_KEY_FEEDS = 'vinta_calendar_feeds'

let isInitialCalendarDiskLoadComplete = false
let saveCalendarTimeout: ReturnType<typeof setTimeout> | null = null

/**
 * Persists the current calendar events and feeds to disk.
 *
 * @param state - Active calendar state.
 */
function syncCalendarToDisk(state: CalendarAppState) {
  if (saveCalendarTimeout) clearTimeout(saveCalendarTimeout)
  saveCalendarTimeout = setTimeout(() => {
    try {
      if (typeof SaveCalendarData === 'function') {
        const payload = JSON.stringify({
          events: state.events,
          feeds: state.feeds
        })
        SaveCalendarData(payload).catch(() => {})
      }
    } catch {
      // Ignore disk write failure
    }
  }, 300)
}

/**
 * Formats a Date object to YYYY-MM-DD local format.
 *
 * @param d - Target Date.
 * @returns Local ISO formatted date string.
 */
export function formatDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function loadInitialEvents(): CalendarEvent[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY_EVENTS)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) {
        return parsed
      }
    }
  } catch {
    // Fall back to empty array on storage error
  }
  return []
}

function loadInitialFeeds(): CalendarFeed[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY_FEEDS)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) {
        return parsed
      }
    }
  } catch {
    // Fall back to empty list
  }
  return []
}

const initialState: CalendarAppState = {
  events: loadInitialEvents(),
  feeds: loadInitialFeeds(),
  selectedDate: formatDateKey(new Date()),
  isSyncing: false
}

export const calendarStore = writable<CalendarAppState>(initialState)

calendarStore.subscribe((state) => {
  try {
    localStorage.setItem(STORAGE_KEY_EVENTS, JSON.stringify(state.events))
    localStorage.setItem(STORAGE_KEY_FEEDS, JSON.stringify(state.feeds))
  } catch {
    // Ignore storage failure
  }

  if (isInitialCalendarDiskLoadComplete) {
    syncCalendarToDisk(state)
  }
})

if (typeof window !== 'undefined') {
  setTimeout(async () => {
    try {
      if (typeof GetCalendarData === 'function') {
        const json = await GetCalendarData()
        if (json && json !== '{}') {
          const parsed = JSON.parse(json)
          const diskEvents: CalendarEvent[] = Array.isArray(parsed.events)
            ? parsed.events
            : []
          const diskFeeds: CalendarFeed[] = Array.isArray(parsed.feeds)
            ? parsed.feeds
            : []

          calendarStore.update((state) => ({
            ...state,
            events: diskEvents,
            feeds: diskFeeds
          }))
        } else {
          const currentState = get(calendarStore)
          if (currentState.events.length > 0 || currentState.feeds.length > 0) {
            syncCalendarToDisk(currentState)
          }
        }
      }
    } catch {
      // Fallback to local storage state on disk read error
    } finally {
      isInitialCalendarDiskLoadComplete = true
      if (get(autoDiscardPastEventsStore)) {
        discardPastCalendarEvents(true)
      }
    }
  }, 0)
}

autoDiscardPastEventsStore.subscribe((enabled) => {
  if (enabled && isInitialCalendarDiskLoadComplete) {
    discardPastCalendarEvents(false)
  }
})

/**
 * Calculates an implied 1-hour end time for an event with start time but no end time.
 */
function getImpliedEndTime(startTime: string): string {
  const [hStr, mStr] = startTime.split(':')
  const h = parseInt(hStr, 10)
  const m = parseInt(mStr, 10)
  if (isNaN(h) || isNaN(m)) return startTime
  const totalM = h * 60 + m + 60
  if (totalM >= 24 * 60) return '23:59'
  const endH = String(Math.floor(totalM / 60)).padStart(2, '0')
  const endM = String(totalM % 60).padStart(2, '0')
  return `${endH}:${endM}`
}

/**
 * Evaluates whether a calendar event has passed based on current local date and time.
 *
 * An event is past if:
 * 1. Its scheduled date precedes today.
 * 2. Its scheduled date is today, it is not an all-day event, and its end time has passed.
 *
 * @param evt - Target calendar event to evaluate.
 * @param now - Reference timestamp (default: current system time).
 * @returns True if event is in the past.
 */
export function isCalendarEventPast(evt: CalendarEvent, now = new Date()): boolean {
  const todayKey = formatDateKey(now)
  if (evt.date < todayKey) {
    return true
  }
  if (evt.date > todayKey) {
    return false
  }
  if (evt.isAllDay) {
    return false
  }
  const currentHours = String(now.getHours()).padStart(2, '0')
  const currentMinutes = String(now.getMinutes()).padStart(2, '0')
  const currentTimeKey = `${currentHours}:${currentMinutes}`

  const effectiveEnd =
    evt.endTime?.trim() || (evt.startTime ? getImpliedEndTime(evt.startTime) : '')
  if (effectiveEnd) {
    return currentTimeKey >= effectiveEnd
  }
  return false
}

/**
 * Updates the currently active selected date on the calendar.
 *
 * @param dateStr - Target date in YYYY-MM-DD format.
 */
export function setSelectedCalendarDate(dateStr: string) {
  calendarStore.update((state) => ({
    ...state,
    selectedDate: dateStr
  }))
}

/**
 * Inserts a new event into the calendar schedule.
 *
 * @param eventData - Event parameters without identifier.
 */
export function addCalendarEvent(
  eventData: Omit<CalendarEvent, 'id' | 'source'> & { source?: 'local' | 'synced' }
) {
  const newEvent: CalendarEvent = {
    source: 'local',
    ...eventData,
    notified: false,
    id: `evt-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
  }
  calendarStore.update((state) => ({
    ...state,
    events: [...state.events, newEvent]
  }))
}

/**
 * Modifies an existing calendar event.
 *
 * @param id - Event identifier.
 * @param updates - Fields to mutate.
 */
export function updateCalendarEvent(
  id: string,
  updates: Partial<Omit<CalendarEvent, 'id'>>
) {
  calendarStore.update((state) => ({
    ...state,
    events: state.events.map((evt) => {
      if (evt.id !== id) return evt
      const timeOrDateChanged =
        (updates.date && updates.date !== evt.date) ||
        (updates.startTime && updates.startTime !== evt.startTime)
      return {
        ...evt,
        ...updates,
        ...(timeOrDateChanged ? { notified: false } : {})
      }
    })
  }))
}

/**
 * Removes an event from the schedule.
 *
 * @param id - Event identifier to delete.
 */
export function deleteCalendarEvent(id: string) {
  calendarStore.update((state) => ({
    ...state,
    events: state.events.filter((evt) => evt.id !== id)
  }))
}

/**
 * Prunes calendar events whose scheduled date or time has passed.
 *
 * @param onlyNotified - When true, only purges past events that have already been notified.
 * @returns Total count of removed past events.
 */
export function discardPastCalendarEvents(onlyNotified = false): number {
  const now = new Date()
  let count = 0
  calendarStore.update((state) => {
    const retained = state.events.filter((evt) => {
      if (isCalendarEventPast(evt, now)) {
        if (onlyNotified && !evt.notified) {
          return true
        }
        count++
        return false
      }
      return true
    })
    return {
      ...state,
      events: retained
    }
  })
  return count
}

/**
 * Registers an external read-only calendar feed (e.g. Google Calendar or Outlook iCal feed).
 *
 * @param name - Display label for the subscription.
 * @param rawUrl - Subscription URL (supports webcal:// and https://).
 * @param color - Color badge for imported events.
 */
export function addCalendarFeed(
  name: string,
  rawUrl: string,
  color: CalendarFeed['color'] = 'sky'
) {
  const cleanUrl = rawUrl.trim().replace(/^webcal:\/\//i, 'https://')
  const newFeed: CalendarFeed = {
    id: `feed-${Date.now()}`,
    name: name.trim() || 'Subscribed Calendar',
    url: cleanUrl,
    color,
    enabled: true
  }
  calendarStore.update((state) => ({
    ...state,
    feeds: [...state.feeds, newFeed]
  }))
}

/**
 * Removes a calendar feed subscription and prunes its associated events.
 *
 * @param feedId - Identifier of feed to remove.
 */
export function removeCalendarFeed(feedId: string) {
  calendarStore.update((state) => ({
    ...state,
    feeds: state.feeds.filter((f) => f.id !== feedId),
    events: state.events.filter((evt) => evt.feedId !== feedId)
  }))
}

/**
 * Minimal iCalendar (RFC 5545) VEVENT parser for read-only calendar sync.
 * Extracts events with summary, description, start time, and location.
 *
 * @param icsText - Raw ICS payload.
 * @param feed - Source subscription feed metadata.
 * @returns Parsed array of calendar events.
 */
export function parseICSContent(
  icsText: string,
  feed: CalendarFeed
): CalendarEvent[] {
  const events: CalendarEvent[] = []
  const veventBlocks = icsText.split(/BEGIN:VEVENT/i).slice(1)

  for (const block of veventBlocks) {
    const endIdx = block.search(/END:VEVENT/i)
    const content = endIdx !== -1 ? block.slice(0, endIdx) : block

    const getProp = (name: string): string => {
      const regex = new RegExp(`(?:^|\\r?\\n)${name}(?:;[^:]*)?:([^\\r\\n]+)`, 'i')
      const match = content.match(regex)
      return match ? match[1].trim() : ''
    }

    const summary = getProp('SUMMARY') || 'Untitled Event'
    const description = getProp('DESCRIPTION')
    const location = getProp('LOCATION')
    const dtstart = getProp('DTSTART')
    const dtend = getProp('DTEND')

    if (!dtstart) continue

    let dateKey = ''
    let startTime = ''
    let endTime = ''
    let isAllDay = false

    // Format: YYYYMMDD or YYYYMMDDTHHMMSSZ
    if (dtstart.length >= 8) {
      const year = dtstart.slice(0, 4)
      const month = dtstart.slice(4, 6)
      const day = dtstart.slice(6, 8)
      dateKey = `${year}-${month}-${day}`

      if (dtstart.includes('T')) {
        const timePart = dtstart.split('T')[1]
        if (timePart.length >= 4) {
          startTime = `${timePart.slice(0, 2)}:${timePart.slice(2, 4)}`
        }
      } else {
        isAllDay = true
      }
    }

    if (dtend && dtend.includes('T')) {
      const timePart = dtend.split('T')[1]
      if (timePart.length >= 4) {
        endTime = `${timePart.slice(0, 2)}:${timePart.slice(2, 4)}`
      }
    }

    if (dateKey) {
      events.push({
        id: `synced-${feed.id}-${events.length}-${dateKey}`,
        title: summary.replace(/\\,/g, ',').replace(/\\;/g, ';'),
        description: description ? description.replace(/\\n/g, '\n') : undefined,
        date: dateKey,
        startTime,
        endTime,
        isAllDay,
        color: feed.color,
        locationUrl: location || undefined,
        source: 'synced',
        feedId: feed.id
      })
    }
  }

  return events
}

/**
 * Synchronizes all registered external feeds over HTTP/HTTPS.
 */
export async function syncAllCalendarFeeds() {
  const state = get(calendarStore)
  if (state.feeds.length === 0) return

  calendarStore.update((s) => ({ ...s, isSyncing: true, lastSyncError: undefined }))

  try {
    let allSyncedEvents: CalendarEvent[] = []
    const updatedFeeds: CalendarFeed[] = []

    for (const feed of state.feeds) {
      if (!feed.enabled) {
        updatedFeeds.push(feed)
        continue
      }

      try {
        const response = await fetch(feed.url)
        if (!response.ok) {
          throw new Error(`HTTP ${response.status} ${response.statusText}`)
        }
        const text = await response.text()
        const parsed = parseICSContent(text, feed)
        allSyncedEvents = [...allSyncedEvents, ...parsed]
        updatedFeeds.push({ ...feed, lastSyncedAt: Date.now() })
      } catch (_err) {
        updatedFeeds.push(feed)
      }
    }

    calendarStore.update((s) => {
      const previouslyNotifiedIds = new Set(
        s.events.filter((e) => e.notified).map((e) => e.id)
      )
      const mergedSynced = allSyncedEvents.map((evt) =>
        previouslyNotifiedIds.has(evt.id) ? { ...evt, notified: true } : evt
      )
      return {
        ...s,
        feeds: updatedFeeds,
        // Retain local events while updating synced ones
        events: [...s.events.filter((e) => e.source === 'local'), ...mergedSynced],
        isSyncing: false
      }
    })
  } catch (err) {
    calendarStore.update((s) => ({
      ...s,
      isSyncing: false,
      lastSyncError: err instanceof Error ? err.message : 'Calendar sync failed'
    }))
  }
}

// ============================================================================
// Event Reminder Watcher
// ============================================================================
let reminderWatcherInterval: ReturnType<typeof setInterval> | null = null

/**
 * Checks upcoming, in-progress, and concluded calendar events against current system time
 * and dispatches one-shot Vinta notifications when scheduled times arrive or when launching
 * after scheduled events have already started.
 */
export function checkCalendarEventReminders(): void {
  if (
    get(pageVisibilityStore).calendar === false &&
    get(homeWidgetsVisibilityStore).calendar === false
  ) {
    return
  }
  const state = get(calendarStore)
  if (!state.events || state.events.length === 0) return

  const now = new Date()
  const todayKey = formatDateKey(now)
  const currentHours = String(now.getHours()).padStart(2, '0')
  const currentMinutes = String(now.getMinutes()).padStart(2, '0')
  const currentTimeKey = `${currentHours}:${currentMinutes}`

  const newlyNotifiedIds = new Set<string>()

  for (const event of state.events) {
    if (event.notified) continue

    // If event was scheduled before today and never notified, mark as notified to prevent stale spam
    if (event.date < todayKey) {
      newlyNotifiedIds.add(event.id)
      continue
    }

    // Only process events scheduled for today
    if (event.date !== todayKey) continue

    const isAllDay = Boolean(event.isAllDay)
    const isPast = isCalendarEventPast(event, now)

    let shouldNotify = false
    let statusPrefix = ''

    if (isAllDay) {
      if (currentTimeKey >= '09:00') {
        shouldNotify = true
        statusPrefix = 'All Day'
      }
    } else if (event.startTime) {
      if (currentTimeKey >= event.startTime) {
        shouldNotify = true
        if (isPast) {
          statusPrefix = 'Missed event'
        } else if (currentTimeKey === event.startTime) {
          statusPrefix = 'Starting now'
        } else {
          statusPrefix = 'In progress'
        }
      }
    }

    if (!shouldNotify) continue

    newlyNotifiedIds.add(event.id)

    const timeRangeStr = event.startTime
      ? `${event.startTime}${event.endTime ? ` - ${event.endTime}` : ''}`
      : 'Scheduled'

    const parts: string[] = []
    if (statusPrefix) {
      parts.push(`${statusPrefix} (${timeRangeStr})`)
    } else {
      parts.push(timeRangeStr)
    }

    if (event.location?.trim()) {
      parts.push(event.location.trim())
    }
    if (event.description?.trim()) {
      parts.push(event.description.trim())
    }

    const message = parts.join(' • ')

    sendVintaNotification({
      title: event.title?.trim() || 'Calendar Event',
      message,
      subApp: 'Agenda'
    })
  }

  if (newlyNotifiedIds.size > 0) {
    calendarStore.update((s) => ({
      ...s,
      events: s.events.map((e) =>
        newlyNotifiedIds.has(e.id) ? { ...e, notified: true } : e
      )
    }))
  }

  // If auto-discard setting is active, prune past events that have been notified
  if (get(autoDiscardPastEventsStore)) {
    discardPastCalendarEvents(true)
  }
}

/**
 * Starts the background periodic timer checking calendar events every 15 seconds.
 */
export function startCalendarReminderWatcher(): void {
  if (reminderWatcherInterval) return
  checkCalendarEventReminders()
  reminderWatcherInterval = setInterval(checkCalendarEventReminders, 15000)
}

if (typeof window !== 'undefined') {
  startCalendarReminderWatcher()
}
