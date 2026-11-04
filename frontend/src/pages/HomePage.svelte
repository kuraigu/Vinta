<script module lang="ts">
  export const config = {
    minimal: { w: 330, h: 35, r: 16 },
    expanded: { w: 430, h: 640, r: 28, minH: 260, maxH: 640 },
    borderColor: 'var(--vinta-island-border-idle)',
    glowColor: 'var(--vinta-island-glow-idle)'
  }
</script>

<script lang="ts">
  import { onMount, onDestroy, getContext } from 'svelte'
  import { ChevronDown } from 'lucide-svelte'
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
  import ScrollView from '../components/ScrollView.svelte'
  import CalendarPage, { config as calendarConfig } from './CalendarPage.svelte'
  import SettingsPage, { config as settingsConfig } from './SettingsPage.svelte'
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
    type IslandConfig
  } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let isUpdateSnoozed = $state(false)

  /**
   * Formats the application version for display in the footer.
   */
  function computeDisplayVersion() {
    const ver = $appInfoStore?.version
    if (!ver) return '0.1.0'
    const trimmed = ver.replace(/^v/, '').replace(/-dev$/i, '')
    return trimmed || '0.1.0'
  }
  let displayVersion = $derived.by(computeDisplayVersion)

  /**
   * Evaluates if there are active non-update notifications to display in the glance slot.
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
   * Navigates directly to the dedicated Calendar application page.
   */
  function openCalendar() {
    if (navigate) {
      navigate(CalendarPage, calendarConfig, true)
    }
  }

  /**
   * Navigates directly to the Updates tab in Settings.
   */
  function openUpdatesTab() {
    if (navigate) {
      $activeSettingsTabStore = 'updates'
      navigate(SettingsPage, settingsConfig, true)
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
   * Dynamically synchronizes the measured height of the home social feed with the island frame.
   */
  function syncHomeFeedMeasuredHeight() {
    const el = feedContainerEl
    if (!el || typeof ResizeObserver === 'undefined') return
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const height = Math.ceil(entry.contentRect.height)
        if (height > 0) {
          const totalContentHeight = $homeWidgetsVisibilityStore.weather
            ? height
            : height + 48
          homeMeasuredHeightStore.set(totalContentHeight)
        }
      }
    })
    ro.observe(el)
    return () => {
      ro.disconnect()
    }
  }
  $effect(syncHomeFeedMeasuredHeight)

  /**
   * Evaluates whether the home feed has vertical overflow below the current scroll position.
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
   * Updates scroll indicator state based on current scroll position.
   */
  function handleFeedScroll(e: Event) {
    const target = e.currentTarget as HTMLElement
    if (target) {
      updateScrollCueVisibility()
    }
  }

  /**
   * Observes the home feed scroll container and content height to keep the scroll cue pill in sync.
   */
  function syncScrollCueObserverEffect() {
    const el = scrollContainer
    if (!el || typeof ResizeObserver === 'undefined') {
      canScrollMore = false
      return
    }

    updateScrollCueVisibility()
    const ro = new ResizeObserver(() => {
      updateScrollCueVisibility()
    })
    ro.observe(el)
    const inner = el.firstElementChild
    if (inner) {
      ro.observe(inner)
    }

    return () => {
      ro.disconnect()
    }
  }
  $effect(syncScrollCueObserverEffect)

  /**
   * Smoothly scrolls the social feed downward when the scroll cue is clicked.
   */
  function scrollFeedDown() {
    if (scrollContainer) {
      scrollContainer.scrollBy({
        top: 180,
        behavior: 'smooth'
      })
    }
  }

  let time = $state(new Date())
  let timer: ReturnType<typeof setInterval>

  onMount(() => {
    timer = setInterval(() => {
      time = new Date()
    }, 1000)
  })

  onDestroy(() => {
    clearInterval(timer)
  })

  /**
   * Resolves the user configured timezone identifier or defaults to Asia/Manila.
   */
  function computeCurrentTimezone() {
    return $timezoneStore || 'Asia/Manila'
  }
  let currentTimezone = $derived.by(computeCurrentTimezone)

  /**
   * Extracts localized date and time components formatted according to active timezone and date settings.
   */
  function computeTimeParts() {
    try {
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
   * Calculates dynamic time-of-day salutation.
   */
  function computeSalutationText() {
    const hour = timeParts.h24
    if (hour >= 5 && hour < 12) return 'Good morning'
    if (hour >= 12 && hour < 17) return 'Good afternoon'
    if (hour >= 17 && hour < 22) return 'Good evening'
    return 'Good night'
  }
  let salutationText = $derived.by(computeSalutationText)

  /**
   * Resolves the user display name defaulting to User.
   */
  function computeDisplayName() {
    return $userNameStore?.trim() || 'User'
  }
  let displayName = $derived.by(computeDisplayName)

  /**
   * Formats display hours honoring 12-hour or 24-hour preference.
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
   * Formats AM/PM indicator for 12-hour clock mode.
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
            <!-- Hero Greeting (Fallback when weather widget is disabled) -->
            <header class="home-hero-greeting">
              <span class="hero-salutation">{salutationText},</span>
              <h1 class="hero-username">{displayName}</h1>
            </header>

            {#if $homeWidgetsVisibilityStore.clock}
              <section class="time-subtle-bar">
                <span class="meta-date-text">{timeParts.friendlyDateText}</span>
                <span class="meta-bullet">•</span>
                <span class="time-subtle-digits">
                  {formattedHours}<span class="colon-blink">:</span
                  >{timeParts.m}
                </span>
                {#if amPm}
                  <span class="time-subtle-ampm">{amPm}</span>
                {/if}
              </section>
            {/if}
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

          <!-- Notification Glance Card (High priority alerts right after updates) -->
          {#if hasStandardNotifications}
            <section class="notification-card-feed">
              <NotificationGlance />
            </section>
          {/if}

          <!-- Calendar Agenda Glance Card -->
          {#if $homeWidgetsVisibilityStore.calendar}
            <section class="calendar-card-feed">
              <CalendarGlance onOpenCalendar={openCalendar} />
            </section>
          {/if}

          <!-- Quick Controls from Action Center (Modular Bento Card) -->
          {#if $homeWidgetsVisibilityStore.actionCenter}
            <section class="glance-card-feed">
              <ActionCenterGlance />
            </section>
          {/if}

          <!-- System Hardware Stats Glance (Modular Bento Card) -->
          {#if $homeWidgetsVisibilityStore.systemStats}
            <section class="glance-card-feed">
              <SystemStats expanded={true} />
            </section>
          {/if}

          <!-- Studio Branding Footer (Issue #185 / Category 7) -->
          <footer class="home-brand-footer">
            <span class="home-brand-text"
              >Vinta by FreeMatrix Studios & Contributors</span
            >
            <span class="home-version-pill">v{displayVersion}</span>
          </footer>
        </div>
      </ScrollView>

      <!-- Scroll cue indicator: hides when feed has no overflow or is scrolled down -->
      <button
        type="button"
        class="home-scroll-cue-pill"
        class:hidden={!canScrollMore || isScrolledDown}
        onclick={scrollFeedDown}
        aria-label="Scroll down for more content"
      >
        <span class="scroll-cue-text">More</span>
        <ChevronDown size={13} class="scroll-cue-icon" />
      </button>
    </div>
  {/if}
</div>

<style src="./styles/HomePage.css"></style>
