<!--
  @fileoverview Overlay page unifying Notifications, Developer Announcements, and Release Notes
  on top of the active workspace page without destroying workspace state.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import { SvelteSet } from 'svelte/reactivity'
  import {
    Bell,
    Megaphone,
    ScrollText,
    CheckCheck,
    ArrowUpRight,
    User,
    Trash2
  } from 'lucide-svelte'
  import OverlayPage from '../../components/OverlayPage.svelte'
  import BentoCard from '../../components/BentoCard.svelte'
  import EmptyState from '../../components/EmptyState.svelte'
  import Notification from '../../components/Notification.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from '../../components/HorizontalTabSelection.svelte'
  import ScrollView from '../../components/ScrollView.svelte'
  import MarkdownViewer from '../../components/MarkdownViewer.svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import {
    getCachedAnnouncements,
    getDeveloperAnnouncements,
    getReleaseNotes,
    markAnnouncementsRead,
    type AnnouncementItem
  } from '../../services/announcementsService'
  import {
    notifData,
    dndStore,
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
  let failedAvatars = new SvelteSet<string>()
  let isConfirmingClear = $state(false)
  let clearConfirmTimeout: number | undefined

  function handleClearAll() {
    if ($notifData.length <= 2 || isConfirmingClear) {
      if (clearConfirmTimeout) clearTimeout(clearConfirmTimeout)
      isConfirmingClear = false
      notifData.set([])
    } else {
      isConfirmingClear = true
      if (clearConfirmTimeout) clearTimeout(clearConfirmTimeout)
      clearConfirmTimeout = window.setTimeout(() => {
        isConfirmingClear = false
      }, 2500)
    }
  }

  function syncClearConfirmState() {
    if (
      $notificationsOverlayTabStore !== 'notifications' ||
      $notifData.length === 0
    ) {
      if (clearConfirmTimeout) clearTimeout(clearConfirmTimeout)
      isConfirmingClear = false
    }
  }
  $effect(syncClearConfirmState)

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
    {#if $notificationsOverlayTabStore === 'notifications'}
      <button
        type="button"
        class="overlay-action-btn overlay-dnd-btn"
        class:dnd-active={$dndStore}
        onclick={() => dndStore.update((v) => !v)}
        use:islandTooltip={{
          text: $dndStore
            ? $tStore('notification.dnd_on')
            : $tStore('notification.dnd_off'),
          position: 'bottom'
        }}
        aria-label={$tStore('notification.dnd_aria')}
      >
        {#if $dndStore}
          <div class="dnd-zzz-icon">
            <Bell size={15} />
            <span class="zzz-badge">z<sup>Z</sup></span>
          </div>
        {:else}
          <Bell size={15} />
        {/if}
      </button>
      {#if $notifData.length > 0}
        <button
          type="button"
          class="overlay-action-btn overlay-clear-btn"
          class:is-confirming={isConfirmingClear}
          onclick={handleClearAll}
          use:islandTooltip={{
            text: isConfirmingClear
              ? $tStore('notification.clear_confirm_tooltip')
              : $tStore('notification.clear_all'),
            position: 'bottom'
          }}
          aria-label={$tStore('notification.clear_aria')}
        >
          <Trash2 size={15} />
          {#if isConfirmingClear}
            <span>{$tStore('notification.confirm_clear')}</span>
          {/if}
        </button>
      {/if}
    {:else if $notificationsOverlayTabStore === 'announcements' && hasUnread}
      <button
        type="button"
        class="overlay-action-btn overlay-read-btn"
        onclick={handleMarkRead}
        use:islandTooltip={{
          text: $tStore('announcements.mark_all_read'),
          position: 'bottom'
        }}
        aria-label={$tStore('announcements.mark_all_read')}
      >
        <CheckCheck size={15} />
      </button>
    {/if}
  {/snippet}

  <div class="notif-overlay-body">
    {#key $notificationsOverlayTabStore}
      <div class="notif-overlay-tab-pane">
        {#if $notificationsOverlayTabStore === 'notifications'}
          <Notification
            expanded={true}
            inOverlay={true}
            notifications={$notifData}
          />
        {:else if $notificationsOverlayTabStore === 'announcements'}
          <ScrollView class="notif-overlay-scroll">
            {#if isLoadingAnnouncements}
              <div class="notif-overlay-empty">
                <span>{$tStore('common.loading')}</span>
              </div>
            {:else if announcements.length === 0}
              <EmptyState
                icon={Megaphone}
                title={$tStore('announcements.empty_title')}
                description={$tStore('announcements.empty_desc')}
              />
            {:else}
              <div class="overlay-feed-list">
                {#each announcements as item (item.id)}
                  <BentoCard
                    variant="bento"
                    class="overlay-feed-card {item.isUnread ? 'is-unread' : ''}"
                  >
                    <div class="author-meta-row">
                      <div class="author-info-box">
                        <div class="author-avatar-box">
                          {#if item.authorAvatar && !failedAvatars.has(item.authorAvatar)}
                            <img
                              src={item.authorAvatar}
                              alt=""
                              class="author-avatar-img"
                              onerror={() => {
                                if (item.authorAvatar)
                                  failedAvatars.add(item.authorAvatar)
                              }}
                            />
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
                          <span class="card-unread-pill"
                            >{$tStore('announcements.new_badge')}</span
                          >
                        {/if}
                      </div>
                    </div>

                    <h3 class="feed-card-title">{item.title}</h3>

                    {#if item.body}
                      <div class="feed-card-markdown">
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
          </ScrollView>
        {:else}
          <ScrollView class="notif-overlay-scroll">
            {#if isLoadingNotes}
              <div class="notif-overlay-empty">
                <span>{$tStore('common.loading')}</span>
              </div>
            {:else}
              <div class="overlay-feed-list">
                <BentoCard variant="bento" class="overlay-feed-card">
                  <div class="feed-card-top">
                    <div class="overlay-version-pill">
                      <ScrollText size={11} />
                      <span>{formatVersion($appInfoStore?.version)}</span>
                    </div>
                  </div>
                  <div class="feed-card-markdown">
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
          </ScrollView>
        {/if}
      </div>
    {/key}
  </div>
</OverlayPage>

<style src="../styles/NotificationsOverlayPage.css"></style>
