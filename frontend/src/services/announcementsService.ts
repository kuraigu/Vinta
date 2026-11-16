/**
 * @fileoverview Fetches developer announcements and GitHub release notes.
 *
 * Pulls update changelogs and announcements through the backend, caches the results
 * in memory and localStorage for instant synchronous hydration on startup, and tracks
 * whether the user has unread announcements to show a badge on the home feed.
 */

import {
  GetDeveloperAnnouncements,
  GetReleaseNotes,
  MarkAnnouncementsRead
} from '../../bindings/Vinta/backend/app'

/**
 * Announcement item representation.
 */
export interface AnnouncementItem {
  id: string
  title: string
  body: string
  url: string
  publishedAt: string
  authorName?: string
  authorAvatar?: string
  isUnread: boolean
}

/**
 * Payload returning announcement feed items and unread status.
 */
export interface AnnouncementsPayload {
  items: AnnouncementItem[]
  hasUnread: boolean
}

const ANNOUNCEMENTS_CACHE_KEY = 'vinta_announcements_cache_v1'
const ANNOUNCEMENTS_TTL_MS = 10 * 60 * 1000
let inMemoryAnnouncements: AnnouncementsPayload | null = null
let lastAnnouncementsFetchMs = 0
let inFlightAnnouncementsPromise: Promise<AnnouncementsPayload> | null = null

/**
 * Synchronously returns the cached announcements payload from memory or localStorage
 * so Home feed glance cards render at their exact height on frame 0 without layout shift.
 */
export function getCachedAnnouncements(): AnnouncementsPayload | null {
  if (inMemoryAnnouncements) return inMemoryAnnouncements
  if (typeof localStorage === 'undefined') return null
  try {
    const raw = localStorage.getItem(ANNOUNCEMENTS_CACHE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as AnnouncementsPayload
    if (parsed && Array.isArray(parsed.items)) {
      inMemoryAnnouncements = parsed
      return parsed
    }
  } catch {
    // Ignore storage read errors
  }
  return null
}

function saveCachedAnnouncements(payload: AnnouncementsPayload): void {
  inMemoryAnnouncements = payload
  if (typeof localStorage === 'undefined') return
  try {
    localStorage.setItem(ANNOUNCEMENTS_CACHE_KEY, JSON.stringify(payload))
  } catch {
    // Ignore storage write errors
  }
}

/**
 * Fetches release notes markdown for a specific version tag.
 */
export async function getReleaseNotes(version: string): Promise<string> {
  try {
    const res = await GetReleaseNotes(version)
    return String(res || '')
  } catch (err) {
    console.error('Failed to retrieve release notes:', err)
    return 'Release notes are currently unavailable.'
  }
}

/**
 * Fetches announcements from the release feed and updates the persistent cache,
 * deduplicating concurrent calls and respecting a 10-minute TTL unless forced.
 */
export async function getDeveloperAnnouncements(
  forceRefresh: boolean = false
): Promise<AnnouncementsPayload> {
  if (
    !forceRefresh &&
    inMemoryAnnouncements &&
    Date.now() - lastAnnouncementsFetchMs < ANNOUNCEMENTS_TTL_MS
  ) {
    return inMemoryAnnouncements
  }
  if (inFlightAnnouncementsPromise) {
    return inFlightAnnouncementsPromise
  }

  inFlightAnnouncementsPromise = (async () => {
    try {
      const res = await GetDeveloperAnnouncements()
      const payload = (res as AnnouncementsPayload) || {
        items: [],
        hasUnread: false
      }
      lastAnnouncementsFetchMs = Date.now()
      if (Array.isArray(payload.items) && payload.items.length > 0) {
        saveCachedAnnouncements(payload)
      }
      return payload
    } catch (err) {
      console.error('Failed to retrieve developer announcements:', err)
      return getCachedAnnouncements() || { items: [], hasUnread: false }
    } finally {
      inFlightAnnouncementsPromise = null
    }
  })()

  return inFlightAnnouncementsPromise
}

/**
 * Marks announcements as read by saving the current content hash.
 */
export async function markAnnouncementsRead(): Promise<void> {
  try {
    await MarkAnnouncementsRead()
    if (inMemoryAnnouncements) {
      saveCachedAnnouncements({
        ...inMemoryAnnouncements,
        hasUnread: false,
        items: inMemoryAnnouncements.items.map((item) => ({
          ...item,
          isUnread: false
        }))
      })
    }
  } catch (err) {
    console.error('Failed to mark announcements as read:', err)
  }
}
