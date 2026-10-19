<script module lang="ts">
  export const config = {
    minimal: { w: 330, h: 35, r: 16 },
    expanded: { w: 430, h: 500, r: 32 },
    borderColor: 'var(--vinta-island-border-idle)',
    glowColor: 'var(--vinta-island-glow-idle)'
  }
</script>

<script lang="ts">
  import { onMount, onDestroy, getContext } from 'svelte'
  import { MapPin, Sparkles, X, RefreshCw } from 'lucide-svelte'
  import Calendar from '../components/Calendar.svelte'
  import Weather from '../components/Weather.svelte'
  import DeviceStatusWidget from '../components/DeviceStatusWidget.svelte'
  import ActionCenterGlance from '../components/ActionCenterGlance.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import CalendarPage, { config as calendarConfig } from './CalendarPage.svelte'
  import SettingsPage, { config as settingsConfig } from './SettingsPage.svelte'
  import {
    homeWidgetsVisibilityStore,
    userNameStore,
    timeFormatStore,
    timezoneStore,
    dateFormatStore,
    formatUserDate,
    updateInfoStore,
    updateLifecycleStatusStore,
    activeSettingsTabStore,
    triggerApplyUpdate,
    type IslandConfig
  } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let isUpdateSnoozed = $state(false)

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

  /**
   * Temporarily hides the update notification banner for the current session.
   */
  function snoozeUpdate(e: MouseEvent) {
    e.stopPropagation()
    isUpdateSnoozed = true
  }

  /**
   * Triggers immediate application of a downloaded and verified update.
   */
  function handleApplyUpdate(e: MouseEvent) {
    e.stopPropagation()
    if (triggerApplyUpdate) {
      triggerApplyUpdate()
    } else {
      openUpdatesTab()
    }
  }

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()

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
      const formatter = new Intl.DateTimeFormat('en-US', {
        timeZone: currentTimezone,
        hour: 'numeric',
        minute: 'numeric',
        second: 'numeric',
        hour12: false,
        weekday: 'long',
        month: 'long',
        day: 'numeric',
        year: 'numeric'
      } as Intl.DateTimeFormatOptions)
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

  /**
   * Resolves timezone label and GMT offset string for the active timezone.
   */
  function computeTimezoneDetails() {
    try {
      const now = new Date()
      const tzFormatter = new Intl.DateTimeFormat('en-US', {
        timeZone: currentTimezone,
        timeZoneName: 'shortOffset'
      })
      const tzParts = tzFormatter.formatToParts(now)
      const offsetPart =
        tzParts.find((p) => p.type === 'timeZoneName')?.value || ''
      const parts = currentTimezone.split('/')
      const friendlyName = parts[parts.length - 1].replace(/_/g, ' ')
      return {
        timezoneName: friendlyName,
        gmtOffset: offsetPart
      }
    } catch {
      return {
        timezoneName: currentTimezone,
        gmtOffset: ''
      }
    }
  }
  let timezoneDetails = $derived.by(computeTimezoneDetails)
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
    <div class="expanded-view">
      <!-- Standard Page Header (In compliance with Settings / SidebarLayout) -->
      <div class="page-header">
        <div class="page-title">Home</div>
      </div>

      <ScrollView class="home-feed-scroll">
        <div class="social-feed-container">
          <!-- Hero Greeting -->
          <header class="home-hero-greeting">
            <div class="greeting-dots-pattern" aria-hidden="true"></div>
            <span class="hero-salutation">{salutationText},</span>
            <h1 class="hero-username">{displayName}</h1>
          </header>

          <!-- Unified Date, Time and Location Row -->
          {#if $homeWidgetsVisibilityStore.clock}
            <section class="time-subtle-bar">
              <span class="meta-date-text">{timeParts.friendlyDateText}</span>
              <span class="meta-bullet">•</span>
              <span class="time-subtle-digits">
                {formattedHours}<span class="colon-blink">:</span>{timeParts.m}
              </span>
              {#if amPm}
                <span class="time-subtle-ampm">{amPm}</span>
              {/if}
              <span class="meta-bullet">•</span>
              <div
                class="meta-location-group"
                title={timezoneDetails.timezoneName}
              >
                <MapPin size={11} class="meta-pin-icon" />
                <span class="meta-city-name"
                  >{timezoneDetails.timezoneName}</span
                >
                {#if timezoneDetails.gmtOffset}
                  <span class="meta-offset">({timezoneDetails.gmtOffset})</span>
                {/if}
              </div>
            </section>
          {/if}

          <!-- Update Glance Banner (Issue #161) -->
          {#if $updateInfoStore?.updateAvailable && !isUpdateSnoozed}
            {@const isReady = $updateLifecycleStatusStore === 'ready'}
            <section
              class="update-glance-card {isReady ? 'ready' : 'available'}"
              role="button"
              tabindex="0"
              onclick={openUpdatesTab}
              onkeydown={(e) => e.key === 'Enter' && openUpdatesTab()}
            >
              <div class="update-glance-icon">
                {#if isReady}
                  <RefreshCw size={14} class="update-spin-icon" />
                {:else}
                  <Sparkles size={14} />
                {/if}
              </div>
              <div class="update-glance-text">
                <span class="update-glance-title">
                  {isReady
                    ? 'Update Ready to Install'
                    : `Vinta ${$updateInfoStore.latestVersion || 'Update'} Available`}
                </span>
                <span class="update-glance-sub">
                  {isReady
                    ? 'Click to restart and apply improvements'
                    : 'Click to review what is new in this release'}
                </span>
              </div>
              <div class="update-glance-actions">
                {#if isReady}
                  <button
                    class="update-action-btn restart"
                    onclick={handleApplyUpdate}
                    title="Restart Vinta"
                  >
                    Restart
                  </button>
                {:else}
                  <button
                    class="update-action-btn view"
                    onclick={(e) => {
                      e.stopPropagation()
                      openUpdatesTab()
                    }}
                    title="View update details"
                  >
                    View
                  </button>
                {/if}
                <button
                  class="update-dismiss-btn"
                  onclick={snoozeUpdate}
                  title="Dismiss for this session"
                  aria-label="Dismiss banner"
                >
                  <X size={12} />
                </button>
              </div>
            </section>
          {/if}

          <!-- Illustrated Weather Banner -->
          {#if $homeWidgetsVisibilityStore.weather}
            <section class="weather-card-feed">
              <Weather expanded={true} />
            </section>
          {/if}

          <!-- Quick Details from Action Center (Modular Bento Card) -->
          {#if $homeWidgetsVisibilityStore.actionCenter}
            <section class="glance-card-feed">
              <ActionCenterGlance />
            </section>
          {/if}

          <!-- Calendar Overview Card (Shared component with Go to Calendar button) -->
          {#if $homeWidgetsVisibilityStore.calendar}
            <section class="calendar-card-feed">
              <Calendar
                expanded={true}
                showGoToCalendar={true}
                onOpenCalendar={openCalendar}
              />
            </section>
          {/if}
        </div>
      </ScrollView>
    </div>
  {/if}
</div>

<style src="./styles/HomePage.css"></style>
