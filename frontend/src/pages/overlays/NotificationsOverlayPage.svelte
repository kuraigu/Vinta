<!--
  @fileoverview Overlay page unifying Notifications, Developer Announcements, and Release Notes
  on top of the active workspace page without destroying workspace state.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import {
    Bell,
    Megaphone,
    ScrollText,
    CheckCheck,
    ExternalLink
  } from 'lucide-svelte'
  import OverlayPage from '../../components/OverlayPage.svelte'
  import Notification from '../../components/Notification.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from '../../components/HorizontalTabSelection.svelte'
  import ScrollView from '../../components/ScrollView.svelte'
  import MarkdownViewer from '../../components/MarkdownViewer.svelte'
  import {
    getCachedAnnouncements,
    getDeveloperAnnouncements,
    getReleaseNotes,
    markAnnouncementsRead,
    type AnnouncementItem
  } from '../../services/announcementsService'
  import {
    notifData,
    appInfoStore,
    notificationsOverlayTabStore,
    tStore,
    type NotificationsOverlayTab
  } from '../../store'

  interface Props {
    onClose: () => void
  }

  let { onClose }: Props = $props()

  const cached = getCachedAnnouncements()
  let announcements = $state<AnnouncementItem[]>(cached?.items || [])
  let hasUnread = $state<boolean>(cached?.hasUnread || false)
  let isLoadingAnnouncements = $state<boolean>(!cached)
  let releaseNotes = $state<string>('')
  let isLoadingNotes = $state<boolean>(true)
  let loadedExtraTabs = false

  const headerTabs = $derived<TabItem<NotificationsOverlayTab>[]>([
    {
      id: 'notifications',
      label: $tStore('dock.notifications'),
      icon: Bell
    },
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

  function formatVersion(ver: string | undefined): string {
    if (!ver) return 'v0.1.0'
    const clean = ver.trim().replace(/^v+/i, '')
    return `v${clean}`
  }

  async function loadAnnouncements() {
    isLoadingAnnouncements = !getCachedAnnouncements()
    const payload = await getDeveloperAnnouncements()
    if (payload && Array.isArray(payload.items)) {
      announcements = payload.items
      hasUnread = payload.hasUnread
    }
    isLoadingAnnouncements = false
  }

  async function loadReleaseNotes() {
    isLoadingNotes = true
    const version = $appInfoStore?.version || '0.1.0'
    releaseNotes = await getReleaseNotes(version)
    isLoadingNotes = false
  }

  async function handleMarkRead() {
    hasUnread = false
    announcements = announcements.map((it) => ({ ...it, isUnread: false }))
    await markAnnouncementsRead()
  }

  function ensureFeedLoaded() {
    if (loadedExtraTabs) return
    if (
      $notificationsOverlayTabStore === 'announcements' ||
      $notificationsOverlayTabStore === 'releases'
    ) {
      loadedExtraTabs = true
      void loadAnnouncements()
      void loadReleaseNotes()
    }
  }

  $effect(ensureFeedLoaded)

  onMount(() => {
    ensureFeedLoaded()
  })
</script>

<OverlayPage
  dataVinta="notifications-overlay-page"
  title={$tStore('dock.notifications')}
  {onClose}
>
  {#snippet headerLeft()}
    <HorizontalTabSelection
      tabs={headerTabs}
      bind:value={$notificationsOverlayTabStore}
      colorVariant="subtle"
      ariaLabel={$tStore('dock.notifications')}
    />
  {/snippet}

  {#snippet headerActions()}
    {#if $notificationsOverlayTabStore === 'announcements' && hasUnread}
      <button type="button" class="overlay-read-btn" onclick={handleMarkRead}>
        <CheckCheck size={13} />
        <span>{$tStore('announcements.mark_read')}</span>
      </button>
    {/if}
  {/snippet}

  <div class="notif-overlay-body">
    {#key $notificationsOverlayTabStore}
      <div class="notif-overlay-tab-pane">
        {#if $notificationsOverlayTabStore === 'notifications'}
          <Notification expanded={true} notifications={$notifData} />
        {:else if $notificationsOverlayTabStore === 'announcements'}
      <ScrollView class="notif-overlay-scroll">
        {#if isLoadingAnnouncements}
          <div class="notif-overlay-empty">
            <span>{$tStore('common.loading')}</span>
          </div>
        {:else if announcements.length === 0}
          <div class="notif-overlay-empty">
            <Megaphone size={26} />
            <span class="empty-title">{$tStore('announcements.empty_title')}</span>
            <span class="empty-sub">{$tStore('announcements.empty_desc')}</span>
          </div>
        {:else}
          <div class="overlay-feed-list">
            {#each announcements as item (item.id)}
              <article
                class="overlay-feed-card"
                class:is-unread={item.isUnread}
              >
                <div class="feed-card-top">
                  <div class="feed-card-title-row">
                    {#if item.isUnread}
                      <span class="feed-unread-dot"></span>
                    {/if}
                    <h3 class="feed-card-title">{item.title}</h3>
                  </div>
                  {#if item.url}
                    <a
                      href={item.url}
                      target="_blank"
                      rel="noreferrer"
                      class="feed-ext-link"
                      aria-label={$tStore('announcements.view_github')}
                    >
                      <ExternalLink size={13} />
                    </a>
                  {/if}
                </div>
                <div class="feed-card-markdown">
                  <MarkdownViewer content={item.body} />
                </div>
              </article>
            {/each}
          </div>
        {/if}
      </ScrollView>
    {:else}
      <ScrollView class="notif-overlay-scroll">
        {#if isLoadingNotes}
          <div class="notif-overlay-empty">
            <span>{$tStore('common.loading')}</span>
          </div>
        {:else}
          <div class="overlay-feed-card">
            <div class="feed-card-top">
              <div class="overlay-version-pill">
                <ScrollText size={11} />
                <span>{formatVersion($appInfoStore?.version)}</span>
              </div>
            </div>
            <div class="feed-card-markdown">
              <MarkdownViewer content={releaseNotes} />
            </div>
          </div>
        {/if}
      </ScrollView>
        {/if}
      </div>
    {/key}
  </div>
</OverlayPage>

<style src="../styles/NotificationsOverlayPage.css"></style>
