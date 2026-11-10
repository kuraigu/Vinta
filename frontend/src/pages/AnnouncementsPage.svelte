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
  import {
    Megaphone,
    ScrollText,
    CheckCheck,
    ArrowUpRight,
    User
  } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import MarkdownViewer from '../components/MarkdownViewer.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from '../components/HorizontalTabSelection.svelte'
  import {
    getDeveloperAnnouncements,
    getReleaseNotes,
    markAnnouncementsRead,
    type AnnouncementItem
  } from '../services/announcementsService'
  import HomePage, { config as homeConfig } from './HomePage.svelte'
  import { appInfoStore, tStore } from '../store'

  let {
    initialTab = 'announcements'
  }: {
    initialTab?: 'announcements' | 'releases'
  } = $props()

  let activeTab = $state<'announcements' | 'releases'>('announcements')
  $effect(function syncInitialAnnouncementsTab() {
    activeTab = initialTab
  })

  let announcements = $state<AnnouncementItem[]>([])
  let hasUnread = $state(false)
  let isLoadingAnnouncements = $state(true)
  let releaseNotes = $state('')
  let isLoadingNotes = $state(true)

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const headerTabs = $derived<TabItem<'announcements' | 'releases'>[]>([
    {
      id: 'announcements',
      label: $tStore('apps.announcements_name'),
      icon: Megaphone
    },
    {
      id: 'releases',
      label: $tStore('release_notes.title'),
      icon: ScrollText
    }
  ])

  /**
   * Navigates island viewport back to the previous page or Home feed.
   */
  function handleBack() {
    if (typeof navigate === 'function') {
      navigate(HomePage, homeConfig, true)
    }
  }

  /**
   * Normalizes version string to ensure a single consistent leading 'v' prefix.
   */
  function formatVersion(ver: string | undefined): string {
    if (!ver) return 'v0.1.0'
    const clean = ver.trim().replace(/^v+/i, '')
    return `v${clean}`
  }

  /**
   * Loads announcement records from local cache and checks GitHub for updates.
   */
  async function loadAnnouncements() {
    isLoadingAnnouncements = true
    const payload = await getDeveloperAnnouncements()
    if (payload && Array.isArray(payload.items)) {
      announcements = payload.items
      hasUnread = payload.hasUnread
    }
    isLoadingAnnouncements = false
  }

  /**
   * Loads release notes markdown from GitHub or fallback cache.
   */
  async function loadReleaseNotes() {
    isLoadingNotes = true
    const version = $appInfoStore?.version || '0.1.0'
    releaseNotes = await getReleaseNotes(version)
    isLoadingNotes = false
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
    void loadReleaseNotes()
  })
</script>

<div data-vinta="announcements-page" class="announcements-page-layout">
  <header class="announcements-page-header">
    <div class="header-left">
      <BackButton onback={handleBack} title={$tStore('common.back')} />
      <HorizontalTabSelection
        tabs={headerTabs}
        bind:value={activeTab}
        colorVariant="subtle"
        ariaLabel="Announcements and Release Notes"
      />
    </div>

    {#if activeTab === 'announcements' && hasUnread}
      <button
        type="button"
        class="mark-all-read-btn"
        onclick={handleMarkRead}
      >
        <CheckCheck size={13} />
        <span>{$tStore('announcements.mark_read')}</span>
      </button>
    {:else if activeTab === 'releases'}
      <div class="version-tag-pill">
        <ScrollText size={11} />
        <span>{formatVersion($appInfoStore?.version)}</span>
      </div>
    {/if}
  </header>

  <div class="announcements-body-wrap">
    <ScrollView class="announcements-scroll-view">
      {#key activeTab}
        <div class="announcements-tab-transition">
          {#if activeTab === 'announcements'}
        {#if isLoadingAnnouncements}
          <div class="announcements-empty-state">
            <span>{$tStore('common.loading')}</span>
          </div>
        {:else if announcements.length === 0}
          <div class="announcements-empty-state">
            <Megaphone size={28} class="empty-icon" />
            <span class="empty-title">{$tStore('announcements.empty_title')}</span>
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
                      <span class="card-unread-pill">{$tStore('announcements.new_badge')}</span>
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
                    <span>{$tStore('announcements.view_github')}</span>
                    <ArrowUpRight size={11} />
                  </a>
                {/if}
              </BentoCard>
            {/each}
          </div>
        {/if}
      {:else if isLoadingNotes}
        <div class="announcements-empty-state">
          <span>{$tStore('release_notes.loading')}</span>
        </div>
      {:else}
        <div class="announcements-list">
          <BentoCard variant="bento" class="announcement-card-item release-notes-card">
            <div class="card-markdown-wrap">
              <MarkdownViewer content={releaseNotes} />
            </div>

            {#if $appInfoStore?.repoUrl}
              <a
                href="{$appInfoStore.repoUrl}/releases"
                target="_blank"
                rel="noreferrer"
                class="card-item-link"
              >
                <span>{$tStore('release_notes.view_all_github')}</span>
                <ArrowUpRight size={12} />
              </a>
            {/if}
          </BentoCard>
        </div>
      {/if}
    </div>
  {/key}
</ScrollView>
  </div>
</div>

<style src="./styles/AnnouncementsPage.css"></style>
