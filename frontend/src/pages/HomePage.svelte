<script module lang="ts">
  export const config = {
    minimal: { w: 330, h: 35, r: 16 },
    expanded: { w: 430, h: 640, r: 28, minH: 260, maxH: 640 },
    borderColor: 'var(--vinta-island-border-idle)',
    glowColor: 'var(--vinta-island-glow-idle)'
  }
</script>

<script lang="ts">
  /**
   * @fileoverview Main feed view for the expanded dynamic island.
   *
   * Displays the current time with localized greetings, mounts glance cards for
   * weather, calendar, and system alerts, and manages the downward scroll cue when
   * widget content overflows the viewport.
   */

  import { onMount, onDestroy, getContext } from 'svelte'
  import ScrollIndicator from '../components/ScrollIndicator.svelte'
  import Calendar from '../components/Calendar.svelte'
  import CalendarGlance from '../components/CalendarGlance.svelte'
  import Weather from '../components/Weather.svelte'
  import WeatherGlance from '../components/WeatherGlance.svelte'
  import DeviceStatusWidget from '../components/DeviceStatusWidget.svelte'
  import ActionCenterGlance from '../components/ActionCenterGlance.svelte'
  import SystemStats from '../components/SystemStats.svelte'
  import NotificationGlance from '../components/NotificationGlance.svelte'
  import UpdateGlanceCard from '../components/UpdateGlanceCard.svelte'
  import SpotlightCarousel from '../components/SpotlightCarousel.svelte'
  import AnnouncementGlance from '../components/AnnouncementGlance.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import CalendarPage, { config as calendarConfig } from './CalendarPage.svelte'
  import SettingsPage, { config as settingsConfig } from './SettingsPage.svelte'
  import AnnouncementsPage, {
    config as announcementsConfig
  } from './AnnouncementsPage.svelte'
  import {
    homeWidgetsVisibilityStore,
    homeMeasuredHeightStore,
    userNameStore,
    timeFormatStore,
    timezoneStore,
    dateFormatStore,
    formatUserDate,
    getCachedDateTimeFormat,
    updateInfoStore,
    appInfoStore,
    notifData,
    activeSettingsTabStore,
    tStore,
    simulatedClockStore,
    type IslandConfig
  } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let isUpdateSnoozed = $state(false)

  /**
   * Resolves the user-facing application version string for the footer pill.
   *
   * @returns Clean version string (e.g. "1.0.0").
   */
  function computeDisplayVersion() {
    const ver = $appInfoStore?.version
    if (!ver) return '0.1.0'
    const trimmed = ver.replace(/^v/, '').replace(/-dev$/i, '')
    return trimmed || '0.1.0'
  }
  let displayVersion = $derived.by(computeDisplayVersion)

  /**
   * Checks whether unread non-glance notifications exist in the backlog.
   *
   * @returns True if dismissible notifications are queued.
   */
  function computeHasStandardNotifications() {
    return ($notifData || []).some(
      (n) =>
        !n.isUpdate &&
        !n.title?.toLowerCase().includes('update available') &&
        !(
          (n.appName?.toLowerCase() === 'vinta' ||
            n.aumid?.toLowerCase().includes('vinta')) &&
          n.title?.toLowerCase().includes('update')
        )
    )
  }
  let hasStandardNotifications = $derived.by(computeHasStandardNotifications)

  /**
   * Navigates the island viewport to the full Calendar agenda view.
   */
  function openCalendar() {
    if (navigate) {
      navigate(CalendarPage, calendarConfig, true)
    }
  }

  /**
   * Navigates the island viewport to the Updates tab within Settings.
   */
  function openUpdatesTab() {
    if (navigate) {
      $activeSettingsTabStore = 'updates'
      navigate(SettingsPage, settingsConfig, true)
    }
  }

  /**
   * Navigates the island viewport to the Developer Announcements feed.
   */
  function openAnnouncementsPage() {
    if (navigate) {
      navigate(AnnouncementsPage, announcementsConfig, true)
    }
  }

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()

  let scrollContainer: HTMLElement | undefined = $state()
  let feedContainerEl: HTMLElement | undefined = $state()
  let isScrolledDown = $state(false)
  let canScrollMore = $state(false)

  /**
   * Synchronizes the measured height of the scrollable feed container
   * and updates the scroll cue pill in a single ResizeObserver pass.
   */
  function syncHomeFeedMeasuredHeight() {
    const el = feedContainerEl
    const scrollEl = scrollContainer
    if ((!el && !scrollEl) || typeof ResizeObserver === 'undefined') {
      canScrollMore = false
      return
    }
    updateScrollCueVisibility()
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        if (entry.target === el) {
          const height = Math.ceil(entry.contentRect.height)
          if (height > 0) {
            const totalContentHeight = $homeWidgetsVisibilityStore.weather
              ? height
              : height + 48
            homeMeasuredHeightStore.set(totalContentHeight)
          }
        }
      }
      updateScrollCueVisibility()
    })
    if (el) ro.observe(el)
    if (scrollEl) ro.observe(scrollEl)
    return () => {
      ro.disconnect()
    }
  }
  $effect(syncHomeFeedMeasuredHeight)

  /**
   * Checks container scroll offset against client height to toggle the
   * visibility of the bottom scroll cue indicator.
   */
  function updateScrollCueVisibility() {
    if (!scrollContainer) {
      canScrollMore = false
      return
    }
    const maxScroll =
      scrollContainer.scrollHeight - scrollContainer.clientHeight
    isScrolledDown = scrollContainer.scrollTop > 20
    canScrollMore = maxScroll > 8 && scrollContainer.scrollTop < maxScroll - 8
  }

  /**
   * Handles feed scroll events, updating scroll indicators and preventing bubbling.
   *
   * @param e - DOM scroll event from the feed container.
   */
  function handleFeedScroll(e: Event) {
    const target = e.currentTarget as HTMLElement
    if (target) {
      updateScrollCueVisibility()
    }
  }

  /**
   * Smoothly scrolls the primary home feed container downward by one viewport view.
   */
  function scrollFeedDown() {
    if (scrollContainer) {
      scrollContainer.scrollBy({
        top: 180,
        behavior: 'smooth'
      })
    }
  }

  let internalTime = $state(new Date())
  let time = $derived($simulatedClockStore ?? internalTime)
  let timer: ReturnType<typeof setInterval>

  onMount(() => {
    timer = setInterval(() => {
      internalTime = new Date()
    }, 10000)
  })

  onDestroy(() => {
    clearInterval(timer)
  })

  /**
   * Computes the active timezone identifier, falling back to system local.
   *
   * @returns IANA timezone string (e.g. "Asia/Tokyo").
   */
  function computeCurrentTimezone() {
    return $timezoneStore || 'Asia/Manila'
  }
  let currentTimezone = $derived.by(computeCurrentTimezone)

  /**
   * Derives decomposed 24-hour, 12-hour, minute, and second values adjusted for timezone.
   *
   * @returns Object containing h24, h12, min, and sec numbers.
   */
  function computeTimeParts() {
    try {
      if ($simulatedClockStore) {
        const h24 = $simulatedClockStore.getHours()
        const m = $simulatedClockStore.getMinutes().toString().padStart(2, '0')
        const s = $simulatedClockStore.getSeconds().toString().padStart(2, '0')
        const dateBadgeText = formatUserDate(
          $simulatedClockStore,
          $dateFormatStore
        )
        const friendlyDateText = formatUserDate(
          $simulatedClockStore,
          $dateFormatStore,
          {
            includeWeekday: true
          }
        )
        return {
          h24,
          m,
          s,
          friendlyDateText,
          fullDateStr: dateBadgeText
        }
      }
      const formatter = getCachedDateTimeFormat('en-US', {
        timeZone: currentTimezone,
        hour: 'numeric',
        minute: 'numeric',
        second: 'numeric',
        hour12: false,
        weekday: 'long',
        month: 'long',
        day: 'numeric',
        year: 'numeric'
      })
      const parts = formatter.formatToParts(time)
      const getVal = (type: string) =>
        parts.find((p) => p.type === type)?.value || '0'
      const h24 = parseInt(getVal('hour'), 10)
      const m = getVal('minute').padStart(2, '0')
      const s = parseInt(getVal('second'), 10).toString().padStart(2, '0')

      const dateBadgeText = formatUserDate(time, $dateFormatStore)
      const friendlyDateText = formatUserDate(time, $dateFormatStore, {
        includeWeekday: true
      })

      return {
        h24,
        m,
        s,
        friendlyDateText,
        fullDateStr: dateBadgeText
      }
    } catch {
      const dateBadgeText = formatUserDate(time, $dateFormatStore)
      const friendlyDateText = formatUserDate(time, $dateFormatStore, {
        includeWeekday: true
      })

      return {
        h24: time.getHours(),
        m: time.getMinutes().toString().padStart(2, '0'),
        s: time.getSeconds().toString().padStart(2, '0'),
        friendlyDateText,
        fullDateStr: dateBadgeText
      }
    }
  }
  let timeParts = $derived.by(computeTimeParts)

  /**
   * Computes a localized greeting phrase (Good morning, afternoon, evening, night)
   * corresponding to the current local hour.
   *
   * @returns Localized greeting from the active language catalog.
   */
  function computeSalutationText() {
    const hour = timeParts.h24
    if (hour >= 5 && hour < 12) return $tStore('home.good_morning')
    if (hour >= 12 && hour < 17) return $tStore('home.good_afternoon')
    if (hour >= 17 && hour < 22) return $tStore('home.good_evening')
    return $tStore('home.good_night')
  }
  let salutationText = $derived.by(computeSalutationText)

  /**
   * Resolves the display name to append to the salutation header.
   *
   * @returns User-configured name or empty string if unset.
   */
  function computeDisplayName() {
    return $userNameStore?.trim() || 'User'
  }
  let displayName = $derived.by(computeDisplayName)

  /**
   * Formats numeric hour value according to 12-hour or 24-hour presentation setting.
   *
   * @returns Formatted hour string.
   */
  function computeFormattedHours() {
    const h24 = timeParts.h24
    if ($timeFormatStore === '24h') {
      return h24.toString().padStart(2, '0')
    }
    const h = h24 % 12 || 12
    return h.toString().padStart(2, '0')
  }
  let formattedHours = $derived.by(computeFormattedHours)

  /**
   * Derives uppercase AM/PM designation for 12-hour clock modes.
   *
   * @returns "AM", "PM", or empty string when in 24-hour mode.
   */
  function computeAmPm() {
    if ($timeFormatStore === '24h') return ''
    return timeParts.h24 >= 12 ? 'PM' : 'AM'
  }
  let amPm = $derived.by(computeAmPm)
</script>

<div data-vinta="home-page" class="page-container">
  {#if !expanded}
    <div class="minimal-view">
      <div class="minimal-slot">
        <DeviceStatusWidget />
      </div>
      {#if $homeWidgetsVisibilityStore.weather}
        <div class="minimal-slot">
          <Weather expanded={false} />
        </div>
      {/if}
      {#if $homeWidgetsVisibilityStore.calendar}
        <div
          class="minimal-slot clickable-slot"
          role="button"
          tabindex="0"
          onclick={(e) => {
            e.stopPropagation()
            openCalendar()
          }}
          onkeydown={(e) => e.key === 'Enter' && openCalendar()}
        >
          <Calendar expanded={false} />
        </div>
      {/if}
    </div>
  {:else}
    <div
      class="expanded-view"
      class:has-hero-weather={$homeWidgetsVisibilityStore.weather}
    >
      <ScrollView
        class="home-feed-scroll"
        bind:element={scrollContainer}
        onscroll={handleFeedScroll}
      >
        <div class="social-feed-container" bind:this={feedContainerEl}>
          {#if $homeWidgetsVisibilityStore.weather}
            <!-- Full-Bleed Weather Hero Header (Background for Salutation + User + Time) -->
            <WeatherGlance
              salutation={salutationText}
              userName={displayName}
              timeText={$homeWidgetsVisibilityStore.clock
                ? `${formattedHours}:${timeParts.m}`
                : ''}
              amPmText={$homeWidgetsVisibilityStore.clock ? amPm : ''}
            />
          {:else}
            <!-- Social Feed Header Bar (Fallback when weather hero is disabled) -->
            <header class="home-feed-status-header">
              <div class="home-hero-greeting">
                <span class="hero-salutation">{salutationText},</span>
                <h1 class="hero-username">{displayName}</h1>
              </div>

              {#if $homeWidgetsVisibilityStore.clock}
                <div class="time-subtle-bar">
                  <span class="meta-date-text"
                    >{timeParts.friendlyDateText}</span
                  >
                  <span class="meta-bullet">•</span>
                  <span class="time-subtle-digits">
                    {formattedHours}:{timeParts.m}
                  </span>
                  {#if amPm}
                    <span class="time-subtle-ampm">{amPm}</span>
                  {/if}
                </div>
              {/if}
            </header>
          {/if}

          <!-- Update Glance Banner (Issue #161) -->
          {#if $updateInfoStore?.updateAvailable && !isUpdateSnoozed}
            <UpdateGlanceCard onClick={openUpdatesTab} />
          {/if}

          <!-- Spotlight Promotional Carousel (Issue #201) -->
          {#if $homeWidgetsVisibilityStore.spotlight !== false}
            <section class="spotlight-card-feed">
              <SpotlightCarousel />
            </section>
          {/if}

          <!-- Grouped Brief Card: Announcements, Notifications & Calendar Agenda -->
          {#if $homeWidgetsVisibilityStore.announcements !== false || hasStandardNotifications || $homeWidgetsVisibilityStore.calendar}
            <BentoCard variant="group" class="home-feed-group">
              {#if $homeWidgetsVisibilityStore.announcements !== false}
                <section class="announcements-card-feed">
                  <AnnouncementGlance onClick={openAnnouncementsPage} />
                </section>
              {/if}

              {#if hasStandardNotifications}
                <section class="notification-card-feed">
                  <NotificationGlance />
                </section>
              {/if}

              {#if $homeWidgetsVisibilityStore.calendar}
                <section class="calendar-card-feed">
                  <CalendarGlance onOpenCalendar={openCalendar} />
                </section>
              {/if}
            </BentoCard>
          {/if}

          <!-- Grouped System Controls Card: Quick Controls & Hardware Stats -->
          {#if $homeWidgetsVisibilityStore.actionCenter || $homeWidgetsVisibilityStore.systemStats}
            <BentoCard variant="group" class="home-feed-group">
              {#if $homeWidgetsVisibilityStore.actionCenter}
                <section class="glance-card-feed">
                  <ActionCenterGlance />
                </section>
              {/if}

              {#if $homeWidgetsVisibilityStore.systemStats}
                <section class="glance-card-feed">
                  <SystemStats expanded={true} />
                </section>
              {/if}
            </BentoCard>
          {/if}

          <!-- Studio Branding Footer -->
          <footer class="home-brand-footer">
            <span class="home-brand-text">{$tStore('home.brand_footer')}</span>
            <span class="home-version-pill">v{displayVersion}</span>
          </footer>
        </div>
      </ScrollView>

      <!-- Scroll indicator cue -->
      <ScrollIndicator
        visible={canScrollMore && !isScrolledDown}
        onclick={scrollFeedDown}
        label={$tStore('home.scroll_cue_more')}
        ariaLabel={$tStore('home.scroll_cue_aria')}
      />
    </div>
  {/if}
</div>

<style src="./styles/HomePage.css"></style>
