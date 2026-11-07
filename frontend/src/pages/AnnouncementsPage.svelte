<!--
  @fileoverview Developer announcements and community news feed.

  Displays official announcements, roadmap milestones, and release alerts fetched from
  GitHub, supporting read/unread state tracking.
-->
<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 36, r: 18 },
    expanded: { w: 440, h: 480, r: 28, minH: 320, maxH: 640 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { Megaphone, CheckCheck, ArrowUpRight, User } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import MarkdownViewer from '../components/MarkdownViewer.svelte'
  import {
    getDeveloperAnnouncements,
    markAnnouncementsRead,
    type AnnouncementItem
  } from '../services/announcementsService'
  import HomePage, { config as homeConfig } from './HomePage.svelte'
  import { tStore } from '../store'

  let announcements = $state<AnnouncementItem[]>([])
  let hasUnread = $state(false)
  let isLoading = $state(true)

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  /**
   * Navigates island viewport back to the previous page or Home feed.
   */
  function handleBack() {
    if (typeof navigate === 'function') {
      navigate(HomePage, homeConfig, true)
    }
  }

  /**
   * Loads announcement records from local cache and checks GitHub for updates.
   */
  async function loadAnnouncements() {
    isLoading = true
    const payload = await getDeveloperAnnouncements()
    if (payload && Array.isArray(payload.items)) {
      announcements = payload.items
      hasUnread = payload.hasUnread
    }
    isLoading = false
  }

  /**
   * Marks all pending announcements as read and clears notification glance badges.
   */
  async function handleMarkRead() {
    hasUnread = false
    announcements = announcements.map((it) => ({ ...it, isUnread: false }))
    await markAnnouncementsRead()
  }

  onMount(() => {
    void loadAnnouncements()
  })
</script>

<div data-vinta="announcements-page" class="announcements-page-layout">
  <header class="announcements-page-header">
    <div class="header-left">
      <BackButton onback={handleBack} title={$tStore('common.back')} />
      <div class="header-titles">
        <h2 class="header-main-title">{$tStore('announcements.page_title')}</h2>
        <span class="header-subtitle">Updates and news from the Vinta team</span
        >
      </div>
    </div>

    {#if hasUnread}
      <button
        type="button"
        class="mark-all-read-btn"
        onclick={handleMarkRead}
        title="Mark all as read"
      >
        <CheckCheck size={13} />
        <span>Mark Read</span>
      </button>
    {/if}
  </header>

  <div class="announcements-body-wrap">
    <ScrollView class="announcements-scroll-view">
      {#if isLoading}
        <div class="announcements-empty-state">
          <span>{$tStore('common.loading')}</span>
        </div>
      {:else if announcements.length === 0}
        <div class="announcements-empty-state">
          <Megaphone size={28} class="empty-icon" />
          <span class="empty-title">{$tStore('announcements.empty_title')}</span
          >
          <span class="empty-sub">{$tStore('announcements.empty_desc')}</span>
        </div>
      {:else}
        <div class="announcements-list">
          {#each announcements as item (item.id)}
            <BentoCard variant="bento" class="announcement-card-item">
              <div class="author-meta-row">
                <div class="author-info-box">
                  <div class="author-avatar-box">
                    {#if item.authorAvatar}
                      <img
                        src={item.authorAvatar}
                        alt=""
                        class="author-avatar-img"
                        onerror={(e) => {
                          const target = e.currentTarget as HTMLImageElement
                          target.style.display = 'none'
                          const fallback =
                            target.nextElementSibling as HTMLElement
                          if (fallback) fallback.style.display = 'flex'
                        }}
                      />
                      <div
                        class="author-avatar-fallback"
                        style="display: none;"
                      >
                        <User size={11} />
                      </div>
                    {:else}
                      <div class="author-avatar-fallback">
                        <User size={11} />
                      </div>
                    {/if}
                  </div>
                  <span class="author-name"
                    >{item.authorName || 'Vinta Team'}</span
                  >
                </div>

                <div class="post-meta-right">
                  {#if item.publishedAt}
                    <span class="card-item-date">{item.publishedAt}</span>
                  {/if}
                  {#if item.isUnread}
                    <span class="card-unread-pill">New</span>
                  {/if}
                </div>
              </div>

              <h3 class="card-item-title">{item.title}</h3>

              {#if item.body}
                <div class="card-markdown-wrap">
                  <MarkdownViewer content={item.body} />
                </div>
              {/if}

              {#if item.url}
                <a
                  href={item.url}
                  target="_blank"
                  rel="noreferrer"
                  class="card-item-link"
                >
                  <span>View on GitHub</span>
                  <ArrowUpRight size={11} />
                </a>
              {/if}
            </BentoCard>
          {/each}
        </div>
      {/if}
    </ScrollView>
  </div>
</div>

<style src="./styles/AnnouncementsPage.css"></style>
