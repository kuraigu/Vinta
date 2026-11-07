<!--
  @component AnnouncementGlance
  @description Minimized glance strip that alerts the user when a new system announcement
  or release update is available.
-->
<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { Megaphone, ChevronRight } from 'lucide-svelte'
  import BentoCard from './BentoCard.svelte'
  import {
    getDeveloperAnnouncements,
    type AnnouncementItem
  } from '../services/announcementsService'
  import AnnouncementsPage, {
    config as announcementsConfig
  } from '../pages/AnnouncementsPage.svelte'
  import { tStore, type IslandConfig } from '../store'

  let {
    onClick
  }: {
    onClick?: () => void
  } = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let latestItem = $state<AnnouncementItem | null>(null)
  let hasUnread = $state(false)

  /**
   * Scans markdown text to find the first valid non-badge preview image.
   */
  function extractFirstImage(text: string | undefined): string | null {
    if (!text) return null
    const mdRegex = /!\[.*?\]\((https?:\/\/[^\s)"]+)(?:\s+["'][^"']*["'])?\)/g
    let match: RegExpExecArray | null
    while ((match = mdRegex.exec(text)) !== null) {
      const url = match[1]
      if (!url.includes('shields.io') && !url.includes('badge')) {
        return url
      }
    }
    const htmlRegex = /<img[^>]+src=["'](https?:\/\/[^"']+)["']/gi
    while ((match = htmlRegex.exec(text)) !== null) {
      const url = match[1]
      if (!url.includes('shields.io') && !url.includes('badge')) {
        return url
      }
    }
    return null
  }

  /**
   * Cleans and converts inline markdown to lightweight styled HTML for the glance card.
   */
  function renderGlanceSnippet(raw: string | undefined, maxLen = 110): string {
    if (!raw) return ''

    let text = raw
      .replace(/!\[.*?\]\((https?:\/\/[^\s)"]+)(?:\s+["'][^"']*["'])?\)/g, '')
      .replace(/<img[^>]*>/gi, '')
      .replace(/<details[\s\S]*?<\/details>/gi, '')
      .replace(/^#+\s+/gm, '')
      .replace(/^---+$/gm, '')
      .replace(/\r?\n+/g, ' ')
      .trim()

    text = text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')

    const isTruncated = text.length > maxLen
    if (isTruncated) {
      text = text.slice(0, maxLen).trim()
    }

    text = text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    text = text.replace(/(?<!\*)\*([^*\s][^*]*[^*\s]|\S)\*(?!\*)/g, '<em>$1</em>')
    text = text.replace(/`([^`]+)`/g, '<code class="glance-inline-code">$1</code>')
    text = text.replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, '<span class="glance-inline-link">$1</span>')

    return isTruncated ? text + '...' : text
  }

  function getPreviewImage(): string | null {
    return extractFirstImage(latestItem?.body)
  }

  function getSnippetHtml(): string {
    if (!latestItem?.body) return $tStore('announcements.default_desc')
    return renderGlanceSnippet(latestItem.body, previewImage ? 85 : 120)
  }

  let previewImage = $derived(getPreviewImage())
  let snippetHtml = $derived(getSnippetHtml())

  onMount(async () => {
    try {
      const payload = await getDeveloperAnnouncements()
      if (payload && Array.isArray(payload.items) && payload.items.length > 0) {
        latestItem = payload.items[0]
        hasUnread = payload.hasUnread
      }
    } catch {
      // Ignore background fetch error
    }
  })

  function handleOpen() {
    if (onClick) {
      onClick()
    } else if (typeof navigate === 'function') {
      navigate(AnnouncementsPage, announcementsConfig, true)
    }
  }
</script>

<div
  class="announcement-glance-wrap"
  role="button"
  tabindex="0"
  onclick={handleOpen}
  onkeydown={(e) => e.key === 'Enter' && handleOpen()}
>
  <BentoCard variant="bento" class="announcement-glance-card">
    <div class="glance-top-row">
      <div class="glance-brand-group">
        <div class="glance-icon-badge">
          <Megaphone size={12} />
        </div>
        <span class="glance-brand-name">{$tStore('announcements.title')}</span>
        {#if hasUnread}
          <span class="glance-unread-pill"
            >{$tStore('announcements.badge_new')}</span
          >
        {/if}
      </div>

      <ChevronRight size={14} class="glance-nav-chevron" />
    </div>

    <div class="glance-content-preview" class:has-image={!!previewImage}>
      <div class="glance-text-col">
        <h4 class="glance-title">
          {latestItem?.title || $tStore('announcements.default_title')}
        </h4>
        <p class="glance-message">
          <!-- eslint-disable-next-line svelte/no-at-html-tags -->
          {@html snippetHtml}
        </p>
      </div>

      {#if previewImage}
        <div class="glance-thumb-wrapper">
          <img
            src={previewImage}
            alt=""
            class="glance-thumb-image"
            onerror={(e) => {
              const target = e.currentTarget as HTMLImageElement
              if (target.parentElement)
                target.parentElement.style.display = 'none'
            }}
          />
        </div>
      {/if}
    </div>
  </BentoCard>
</div>

<style src="./styles/AnnouncementGlance.css"></style>
