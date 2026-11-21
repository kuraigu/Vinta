<script module lang="ts">
  import { ISLAND_WIDTH_TIERS } from '../store'

  export const config = {
    minimal: { w: 330, h: 35, r: 16 },
    expanded: {
      w: ISLAND_WIDTH_TIERS.standard,
      h: 640,
      r: 28,
      minH: 260,
      maxH: 640
    },
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
  import { SlidersHorizontal } from 'lucide-svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import Calendar from '../components/Calendar.svelte'
  import Weather from '../components/Weather.svelte'
  import DeviceStatusWidget from '../components/DeviceStatusWidget.svelte'
  import UpdateGlanceCard from '../components/UpdateGlanceCard.svelte'
  import SpotlightWidget from '../components/widgets/SpotlightWidget.svelte'
  import WeatherWidget from '../components/widgets/WeatherWidget.svelte'
  import MediaGlanceWidget from '../components/widgets/MediaGlanceWidget.svelte'
  import BriefFeedWidget from '../components/widgets/BriefFeedWidget.svelte'
  import SystemStatsWidget from '../components/widgets/SystemStatsWidget.svelte'
  import ActionCenterWidget from '../components/widgets/ActionCenterWidget.svelte'
  import CustomContextMenu, {
    type ContextMenuItem
  } from '../components/CustomContextMenu.svelte'
  import IslandPage from '../components/IslandPage.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import CalendarPage, { config as calendarConfig } from './CalendarPage.svelte'
  import SettingsPage, { config as settingsConfig } from './SettingsPage.svelte'
  import {
    homeWidgetsVisibilityStore,
    homeWidgetSizesStore,
    homeWidgetMenuEnabledStore,
    type WidgetSizePreset,
    homeMeasuredHeightStore,
    userNameStore,
    timeFormatStore,
    timezoneStore,
    dateFormatStore,
    formatUserDate,
    getCachedDateTimeFormat,
    updateInfoStore,
    appInfoStore,
    batteryDataStore,
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
   * Navigates the island viewport to the Apps & Widgets tab within Settings.
   */
  function openModulesTab() {
    if (navigate) {
      $activeSettingsTabStore = 'modules'
      navigate(SettingsPage, settingsConfig, true)
    }
  }

  let hasMinimalStatusSlot = $derived(
    $homeWidgetsVisibilityStore.clock ||
      $homeWidgetsVisibilityStore.wireless ||
      ($homeWidgetsVisibilityStore.battery && $batteryDataStore.hasBattery)
  )

  let allMinimalWidgetsDisabled = $derived(
    !hasMinimalStatusSlot &&
      !$homeWidgetsVisibilityStore.weather &&
      !$homeWidgetsVisibilityStore.calendar
  )

  let allHomeWidgetsDisabled = $derived(
    !$homeWidgetsVisibilityStore.clock &&
      !$homeWidgetsVisibilityStore.calendar &&
      !$homeWidgetsVisibilityStore.weather &&
      !$homeWidgetsVisibilityStore.battery &&
      !$homeWidgetsVisibilityStore.wireless &&
      !$homeWidgetsVisibilityStore.actionCenter &&
      !$homeWidgetsVisibilityStore.systemStats &&
      $homeWidgetsVisibilityStore.spotlight === false &&
      $homeWidgetsVisibilityStore.announcements === false &&
      $homeWidgetsVisibilityStore.notifications === false &&
      $homeWidgetsVisibilityStore.media === false
  )

  const openOverlay =
    getContext<
      (
        overlay: 'notifications' | 'apps' | null,
        opts?: { tab?: 'notifications' | 'announcements' | 'releases' }
      ) => void
    >('openOverlay')

  /**
   * Opens the Developer Announcements tab inside NotificationsOverlayPage.
   */
  function openAnnouncementsPage() {
    if (typeof openOverlay === 'function') {
      openOverlay('notifications', { tab: 'announcements' })
    }
  }

  const WIDGET_SUPPORTED_SIZES: Record<string, WidgetSizePreset[]> = {
    brief: ['2x1', '2x2'],
    media: ['1x1', '2x1', '2x2'],
    actionCenter: ['1x1', '2x1'],
    systemStats: ['1x1', '2x1', '2x2']
  }

  let ctxOpen = $state(false)
  let ctxCoords = $state({ x: 0, y: 0 })
  let ctxWidgetId = $state<string | null>(null)

  function getWidgetSize(id: string): WidgetSizePreset {
    const allowed = WIDGET_SUPPORTED_SIZES[id] || ['2x1']
    const saved = $homeWidgetSizesStore?.[id]
    if (saved && allowed.includes(saved)) {
      return saved
    }
    return allowed[0]
  }

  function setWidgetSize(id: string, size: WidgetSizePreset) {
    homeWidgetSizesStore.update((prev) => ({
      ...prev,
      [id]: size
    }))
  }

  function handleWidgetContextMenu(e: MouseEvent, widgetId: string) {
    if (!$homeWidgetMenuEnabledStore) return
    e.preventDefault()
    e.stopPropagation()
    if (ctxOpen && ctxWidgetId === widgetId && e.type !== 'contextmenu') {
      ctxOpen = false
      return
    }
    ctxCoords = { x: e.clientX, y: e.clientY }
    ctxWidgetId = widgetId
    ctxOpen = true
  }

  function getWidgetContextMenuItems(): ContextMenuItem[] {
    const id = ctxWidgetId
    if (!id) return []
    const currentSize = getWidgetSize(id)
    const allowed = WIDGET_SUPPORTED_SIZES[id] || ['2x1']
    const sizeLabels: Record<WidgetSizePreset, string> = {
      '1x1': $tStore('home.widget_size_1x1'),
      '2x1': $tStore('home.widget_size_2x1'),
      '2x2': $tStore('home.widget_size_2x2')
    }
    const items: ContextMenuItem[] = allowed.map((sz) => ({
      label: `${currentSize === sz ? '• ' : ''}${sizeLabels[sz]}`,
      action: () => setWidgetSize(id, sz)
    }))
    if (id !== 'brief') {
      items.push(
        { label: '', action: () => {}, divider: true },
        {
          label: $tStore('home.widget_hide'),
          action: () => {
            homeWidgetsVisibilityStore.update((vis) => ({
              ...vis,
              [id]: false
            }))
          }
        }
      )
    }
    return items
  }

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()

  let feedContainerEl: HTMLElement | undefined = $state()

  /**
   * Synchronizes the measured height of the scrollable feed container
   * in a single ResizeObserver pass.
   */
  function syncHomeFeedMeasuredHeight() {
    const el = feedContainerEl
    if (!el || typeof ResizeObserver === 'undefined') {
      return
    }
    let measureTimer: ReturnType<typeof setTimeout> | undefined
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        if (entry.target === el) {
          const height = Math.ceil(entry.contentRect.height)
          if (height > 0) {
            if (measureTimer !== undefined) clearTimeout(measureTimer)
            measureTimer = setTimeout(() => {
              measureTimer = undefined
              const totalContentHeight = $homeWidgetsVisibilityStore.weather
                ? height
                : height + 48
              homeMeasuredHeightStore.set(totalContentHeight)
            }, 60)
          }
        }
      }
    })
    ro.observe(el)
    return () => {
      if (measureTimer !== undefined) clearTimeout(measureTimer)
      ro.disconnect()
    }
  }
  $effect(syncHomeFeedMeasuredHeight)

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
  <IslandPage {expanded}>
    {#snippet minimal()}
      <div class="minimal-view">
        {#if allMinimalWidgetsDisabled}
          <div class="minimal-slot minimal-empty-brand">
            <span class="minimal-empty-dot" aria-hidden="true"></span>
            <span>Vinta</span>
          </div>
        {:else}
          {#if hasMinimalStatusSlot}
            <div class="minimal-slot">
              <DeviceStatusWidget />
            </div>
          {/if}
          {#if $homeWidgetsVisibilityStore.weather}
            <div class="minimal-slot">
              <Weather expanded={false} />
            </div>
          {/if}
          {#if $homeWidgetsVisibilityStore.calendar}
            <div class="minimal-slot">
              <Calendar expanded={false} />
            </div>
          {/if}
        {/if}
      </div>
    {/snippet}

    {#snippet expandedContent()}
      <div
        class="expanded-view"
        class:has-hero-weather={$homeWidgetsVisibilityStore.weather}
      >
        <ScrollView
          class="home-feed-scroll"
          showScrollIndicators={true}
          hideIndicatorOnScroll={true}
          scrollStep={180}
          scrollIndicatorLabel={$tStore('home.scroll_cue_more')}
          scrollIndicatorAriaLabel={$tStore('home.scroll_cue_aria')}
        >
          <div class="social-feed-container" bind:this={feedContainerEl}>
            {#if $homeWidgetsVisibilityStore.weather}
              <!-- Full-Bleed Weather Hero Header (Background for Salutation + User + Time) -->
              <WeatherWidget
                salutation={salutationText}
                userName={displayName}
                timeText={$homeWidgetsVisibilityStore.clock
                  ? `${formattedHours}:${timeParts.m}`
                  : ''}
                amPmText={$homeWidgetsVisibilityStore.clock ? amPm : ''}
                paused={!expanded}
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
                <SpotlightWidget />
              </section>
            {/if}

            <!-- Modular Bento Widget Grid -->
            {#if allHomeWidgetsDisabled}
              <BentoCard variant="bento" class="home-empty-widgets-card">
                <div class="home-empty-widgets-body">
                  <div class="home-empty-widgets-icon" aria-hidden="true">
                    <SlidersHorizontal size={18} strokeWidth={2} />
                  </div>
                  <div class="home-empty-widgets-copy">
                    <h2 class="home-empty-widgets-title">
                      {$tStore('home.widgets_empty_title')}
                    </h2>
                    <p class="home-empty-widgets-desc">
                      {$tStore('home.widgets_empty_desc')}
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  class="home-empty-widgets-cta"
                  onclick={openModulesTab}
                >
                  {$tStore('home.widgets_empty_cta')}
                </button>
              </BentoCard>
            {:else}
              <div class="home-widgets-bento">
                {#if $homeWidgetsVisibilityStore.media !== false}
                  <MediaGlanceWidget
                    size={getWidgetSize('media')}
                    active={expanded}
                    showMenuButton={$homeWidgetMenuEnabledStore}
                    oncontextmenu={(e) => handleWidgetContextMenu(e, 'media')}
                  />
                {/if}

                {#if $homeWidgetsVisibilityStore.announcements !== false || $homeWidgetsVisibilityStore.calendar || ($homeWidgetsVisibilityStore.notifications !== false && hasStandardNotifications)}
                  <BriefFeedWidget
                    showAnnouncements={$homeWidgetsVisibilityStore.announcements !==
                      false}
                    showNotifications={$homeWidgetsVisibilityStore.notifications !==
                      false && hasStandardNotifications}
                    showCalendar={$homeWidgetsVisibilityStore.calendar}
                    size={getWidgetSize('brief')}
                    showMenuButton={$homeWidgetMenuEnabledStore}
                    onOpenAnnouncements={openAnnouncementsPage}
                    onOpenCalendar={openCalendar}
                    oncontextmenu={(e) => handleWidgetContextMenu(e, 'brief')}
                  />
                {/if}

                {#if $homeWidgetsVisibilityStore.actionCenter}
                  <ActionCenterWidget
                    size={getWidgetSize('actionCenter')}
                    showMenuButton={$homeWidgetMenuEnabledStore}
                    oncontextmenu={(e) =>
                      handleWidgetContextMenu(e, 'actionCenter')}
                  />
                {/if}

                {#if $homeWidgetsVisibilityStore.systemStats}
                  <SystemStatsWidget
                    size={getWidgetSize('systemStats')}
                    active={expanded}
                    showMenuButton={$homeWidgetMenuEnabledStore}
                    oncontextmenu={(e) =>
                      handleWidgetContextMenu(e, 'systemStats')}
                  />
                {/if}
              </div>
            {/if}

            <!-- Studio Branding Footer -->
            <footer class="home-brand-footer">
              <span class="home-brand-text">{$tStore('home.brand_footer')}</span
              >
              <span class="home-version-pill">v{displayVersion}</span>
            </footer>
          </div>
        </ScrollView>
      </div>
    {/snippet}
  </IslandPage>

  {#if ctxOpen}
    <CustomContextMenu
      x={ctxCoords.x}
      y={ctxCoords.y}
      items={getWidgetContextMenuItems()}
      onclose={() => {
        ctxOpen = false
      }}
    />
  {/if}
</div>

<style src="./styles/HomePage.css"></style>
