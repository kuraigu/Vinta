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

    <div class="glance-content-preview">
      <h4 class="glance-title">
        {latestItem?.title || $tStore('announcements.default_title')}
      </h4>
      <p class="glance-message">
        {latestItem?.body
          ? latestItem.body.replace(/\r?\n+/g, ' ').slice(0, 110) + '...'
          : $tStore('announcements.default_desc')}
      </p>
    </div>
  </BentoCard>
</div>

<style src="./styles/AnnouncementGlance.css"></style>
