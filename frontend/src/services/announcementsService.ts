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
 * Fetches release notes markdown for the specified version, prioritizing offline cache.
 *
 * @param version - Version tag to fetch release notes for
 * @returns Release notes markdown content
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
 * Fetches developer announcements from the official release feed.
 *
 * @returns Announcements payload containing items and unread indicator
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
 * Marks announcements as read by recording the active content hash.
 */
export async function markAnnouncementsRead(): Promise<void> {
  try {
    await Call.ByName('Vinta/backend.App.MarkAnnouncementsRead')
  } catch (err) {
    console.error('Failed to mark announcements as read:', err)
  }
}
