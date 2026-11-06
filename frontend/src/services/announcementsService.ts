/**
 * @fileoverview Fetches developer announcements and GitHub release notes.
 *
 * Pulls update changelogs and announcements through the backend, caches the results,
 * and tracks whether the user has unread announcements to show a badge on the home feed.
 */

import { Call } from '@wailsio/runtime'

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

/**
 * Fetches release notes markdown for a specific version tag.
 */
export async function getReleaseNotes(version: string): Promise<string> {
  try {
    const res = await Call.ByName('Vinta/backend.App.GetReleaseNotes', version)
    return String(res || '')
  } catch (err) {
    console.error('Failed to retrieve release notes:', err)
    return 'Release notes are currently unavailable.'
  }
}

/**
 * Fetches announcements from the release feed.
 */
export async function getDeveloperAnnouncements(): Promise<AnnouncementsPayload> {
  try {
    const res = await Call.ByName('Vinta/backend.App.GetDeveloperAnnouncements')
    return (res as AnnouncementsPayload) || { items: [], hasUnread: false }
  } catch (err) {
    console.error('Failed to retrieve developer announcements:', err)
    return { items: [], hasUnread: false }
  }
}

/**
 * Marks announcements as read by saving the current content hash.
 */
export async function markAnnouncementsRead(): Promise<void> {
  try {
    await Call.ByName('Vinta/backend.App.MarkAnnouncementsRead')
  } catch (err) {
    console.error('Failed to mark announcements as read:', err)
  }
}
