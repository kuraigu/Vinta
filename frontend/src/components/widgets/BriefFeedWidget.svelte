<!--
  @component BriefFeedWidget
  @description Unified Home brief widget grouping Announcements, Notifications, and Calendar Agenda
  inside a single Bento group card with `data-vinta-widget="brief"` scoping.
-->
<script lang="ts">
  import WidgetFrame from './WidgetFrame.svelte'
  import BentoCard from '../BentoCard.svelte'
  import AnnouncementGlance from '../AnnouncementGlance.svelte'
  import NotificationGlance from '../NotificationGlance.svelte'
  import CalendarGlance from '../CalendarGlance.svelte'
  import type { WidgetSizePreset } from '../../stores/settings'

  interface Props {
    showAnnouncements?: boolean
    showNotifications?: boolean
    showCalendar?: boolean
    size?: WidgetSizePreset
    showMenuButton?: boolean
    onOpenAnnouncements?: () => void
    onOpenCalendar?: () => void
    oncontextmenu?: (e: MouseEvent) => void
  }

  let {
    showAnnouncements = true,
    showNotifications = false,
    showCalendar = true,
    size = '2x1',
    showMenuButton = false,
    onOpenAnnouncements = undefined,
    onOpenCalendar = undefined,
    oncontextmenu = undefined
  }: Props = $props()
</script>

{#if showAnnouncements || showNotifications || showCalendar}
  <WidgetFrame id="brief" {size} {showMenuButton} {oncontextmenu}>
    <BentoCard variant="group" class="home-feed-group">
      {#if showAnnouncements}
        <section class="announcements-card-feed">
          <AnnouncementGlance onClick={onOpenAnnouncements} />
        </section>
      {/if}

      {#if showNotifications}
        <section class="notification-card-feed">
          <NotificationGlance />
        </section>
      {/if}

      {#if showCalendar}
        <section class="calendar-card-feed">
          <CalendarGlance onOpenCalendar={onOpenCalendar} />
        </section>
      {/if}
    </BentoCard>
  </WidgetFrame>
{/if}
