import { writable, get } from 'svelte/store'
import {
  GetCalendarData,
  SaveCalendarData
} from '../../bindings/Vinta/backend/app.js'
import { sendVintaNotification } from '../services/vintaNotification'

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
    }
  }, 0)
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
    events: state.events.map((evt) =>
      evt.id === id ? { ...evt, ...updates } : evt
    )
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

    calendarStore.update((s) => ({
      ...s,
      feeds: updatedFeeds,
      // Retain local events while updating synced ones
      events: [...s.events.filter((e) => e.source === 'local'), ...allSyncedEvents],
      isSyncing: false
    }))
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
const notifiedEventKeys = new Set<string>()
let reminderWatcherInterval: ReturnType<typeof setInterval> | null = null

/**
 * Checks upcoming and active calendar events against the current system time
 * and dispatches one-shot Vinta notifications when scheduled start times arrive.
 */
export function checkCalendarEventReminders(): void {
  const state = get(calendarStore)
  if (!state.events || state.events.length === 0) return

  const now = new Date()
  const todayKey = formatDateKey(now)
  const currentHours = String(now.getHours()).padStart(2, '0')
  const currentMinutes = String(now.getMinutes()).padStart(2, '0')
  const currentTimeKey = `${currentHours}:${currentMinutes}`

  for (const event of state.events) {
    if (event.date !== todayKey) continue

    const isTimedMatch =
      !event.isAllDay &&
      event.startTime &&
      event.startTime === currentTimeKey

    const isAllDayMorningMatch =
      Boolean(event.isAllDay) &&
      currentHours === '09' &&
      currentMinutes === '00'

    if (!isTimedMatch && !isAllDayMorningMatch) continue

    const notifyKey = `${event.id}_${todayKey}_${event.startTime || 'allday'}`
    if (notifiedEventKeys.has(notifyKey)) continue

    notifiedEventKeys.add(notifyKey)

    let message = ''
    if (event.location && event.description) {
      message = `${event.location} • ${event.description}`
    } else if (event.location) {
      message = `Location: ${event.location}`
    } else if (event.description) {
      message = event.description
    } else if (event.isAllDay) {
      message = 'All-day event scheduled for today'
    } else {
      message = event.endTime
        ? `Scheduled until ${event.endTime}`
        : `Scheduled event starting at ${event.startTime}`
    }

    sendVintaNotification({
      title: event.title,
      message,
      subApp: 'Agenda'
    })
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
