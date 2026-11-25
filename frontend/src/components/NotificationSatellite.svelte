<!--
  @component NotificationSatellite
  @description Compact circular floating notification satellite. Docks beside the
  island when a notification arrives, displays the app icon or avatar with a small
  app badge and circular countdown ring, pauses on hover, and opens Notifications on click.
-->
<script lang="ts">
  import { getContext } from 'svelte'
  import { Bell, X } from 'lucide-svelte'
  import SatelliteHost from './SatelliteHost.svelte'
  import {
    satelliteNotification,
    dismissSatelliteNotification,
    pauseSatelliteTimer,
    resumeSatelliteTimer
  } from '../stores/notificationSatellite'
  import {
    autoHideDelayStore,
    notifSatelliteEnabledStore,
    pageVisibilityStore
  } from '../stores/settings'
  import {
    nativeIconsStore,
    parseAumid,
    requestNativeAppIcon
  } from '../services/iconService'
  import Notification, { config as notifConfig } from './Notification.svelte'
  import type { IslandConfig } from '../store'
  import type { FourPartGaps, FourPartRadii } from '../stores/satellites'
  import { swipe } from '../utils/useSwipe'
  import appIcon from '@appicon'

  interface Props {
    active?: boolean
    primaryCenter?: number
    clampedOffsetX?: number
    monitorLeft?: number
    activeMonitorWidth?: number
    islandWidth?: number
    expanded?: boolean
    minimalHeight?: number
    uiScale?: number
  }

  let {
    active = true,
    primaryCenter = 960,
    clampedOffsetX = 0,
    monitorLeft = 0,
    activeMonitorWidth = 1920,
    islandWidth = 368,
    expanded = false,
    minimalHeight = 36,
    uiScale = 1
  }: Props = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let currentNotif = $derived($satelliteNotification)
  let nativeIcons = $derived($nativeIconsStore)

  /**
   * Computes the auto-hide timeout duration in seconds for the circular ring.
   */
  function computeDurationSeconds(): number {
    if (typeof $autoHideDelayStore === 'number' && $autoHideDelayStore > 0) {
      return $autoHideDelayStore
    }
    return 8
  }

  let durationSeconds = $derived.by(computeDurationSeconds)

  let swipeOffsetX = $state(0)
  let isSwiping = $state(false)

  const satelliteSize = 40
  const satelliteRadii: FourPartRadii = [20, 20, 20, 20]
  const satelliteGaps: FourPartGaps = [0, 10, 0, 10]
  let verticalOffsetY = $derived(
    expanded ? 0 : Math.round((minimalHeight - satelliteSize) / 2)
  )

  /**
   * Computes whether the satellite should dock to the left or right flank of the island.
   */
  function computeSatelliteAnchor(): 'left' | 'right' {
    const islandCenterOnScreen = primaryCenter + clampedOffsetX - monitorLeft
    const scale = uiScale > 0 ? uiScale : 1
    const halfIslandW = (islandWidth / 2) * scale
    const scaledGap = 12 * scale
    const scaledSatW = satelliteSize * scale
    const leftEdge = islandCenterOnScreen - halfIslandW
    const leftClearance = leftEdge - scaledGap - scaledSatW
    const rightEdge = islandCenterOnScreen + halfIslandW
    const rightClearance =
      activeMonitorWidth - (rightEdge + scaledGap + scaledSatW)

    if (leftClearance < 6 && rightClearance >= 6) {
      return 'right'
    }
    return 'left'
  }

  let anchor = $derived.by(computeSatelliteAnchor)

  /**
   * Evaluates if a given string is a valid image URI.
   */
  function isImageUrl(url?: string): boolean {
    if (!url) return false
    return (
      url.startsWith('http://') ||
      url.startsWith('https://') ||
      url.startsWith('data:image/') ||
      url.startsWith('file://') ||
      url.startsWith('blob:') ||
      url.startsWith('/@fs/') ||
      url.startsWith('wails://') ||
      url.startsWith('/') ||
      url.endsWith('.png') ||
      url.endsWith('.jpg') ||
      url.endsWith('.jpeg') ||
      url.endsWith('.webp') ||
      url.endsWith('.ico')
    )
  }

  /**
   * Resolves the native or bundled app logo for the notification.
   */
  function resolveAppBadge(): string | null {
    if (!currentNotif) return null

    const rawInput = currentNotif.appName || currentNotif.aumid || ''
    const isVinta =
      rawInput.toLowerCase().includes('vinta') ||
      (currentNotif.title || '').toLowerCase().includes('vinta')

    if (isVinta) {
      return currentNotif.appLogo || appIcon
    }

    if (currentNotif.appLogo && isImageUrl(currentNotif.appLogo)) {
      return currentNotif.appLogo
    }

    const parsed = parseAumid(rawInput)
    const searchApp = (parsed.processName || currentNotif.appName || '').trim()
    const searchAumid = (currentNotif.aumid || '').trim()
    const iconKey = `${searchApp.toLowerCase()}|${searchAumid.toLowerCase()}`

    if (nativeIcons[iconKey]) return nativeIcons[iconKey]
    if (currentNotif.aumid && nativeIcons[currentNotif.aumid]) {
      return nativeIcons[currentNotif.aumid]
    }
    if (currentNotif.appName && nativeIcons[currentNotif.appName]) {
      return nativeIcons[currentNotif.appName]
    }

    if (searchApp || searchAumid) {
      const extracted = requestNativeAppIcon(searchApp, searchAumid)
      if (extracted) return extracted
    }

    return null
  }

  /**
   * Resolves the primary visual avatar or app icon for the circle.
   */
  function resolvePrimaryIcon(): string | null {
    if (!currentNotif) return null
    if (currentNotif.icon && isImageUrl(currentNotif.icon)) {
      return currentNotif.icon
    }
    return resolveAppBadge()
  }

  let resolvedAppBadge = $derived.by(resolveAppBadge)
  let resolvedIcon = $derived.by(resolvePrimaryIcon)
  let showMiniAppBadge = $derived(
    Boolean(
      resolvedAppBadge && resolvedIcon && resolvedAppBadge !== resolvedIcon
    )
  )

  let displayTitle = $derived.by(() => {
    if (!currentNotif) return ''
    if (currentNotif.title && currentNotif.title.trim()) {
      return currentNotif.title.trim()
    }
    if (currentNotif.appName && currentNotif.appName.trim()) {
      return currentNotif.appName.trim()
    }
    if (currentNotif.aumid) {
      return parseAumid(currentNotif.aumid).displayName
    }
    return 'Notification'
  })

  /**
   * Navigates to the Notification Center and dismisses the satellite circle.
   */
  function handleCircleClick() {
    if ($pageVisibilityStore.notifications !== false && navigate) {
      navigate(Notification, notifConfig, true)
    }
    dismissSatelliteNotification()
  }

  /**
   * Explicitly closes the notification satellite without opening Notifications.
   */
  function handleDismissClick(e: MouseEvent) {
    e.stopPropagation()
    dismissSatelliteNotification()
  }

  /**
   * Dismisses the notification satellite on right-click.
   */
  function handleContextDismiss(e: MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    dismissSatelliteNotification()
  }

  $effect(function handleSatelliteDisabled() {
    if (!$notifSatelliteEnabledStore && currentNotif !== null) {
      dismissSatelliteNotification()
    }
  })
</script>

<SatelliteHost
  id="notification"
  {anchor}
  width={satelliteSize}
  height={satelliteSize}
  radius={satelliteRadii}
  gap={satelliteGaps}
  offsetY={verticalOffsetY}
  active={active && currentNotif !== null && $notifSatelliteEnabledStore}
  onmouseenter={pauseSatelliteTimer}
  onmouseleave={resumeSatelliteTimer}
>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="notification-satellite"
    role="button"
    tabindex="0"
    aria-label={displayTitle || 'Notification'}
    onclick={handleCircleClick}
    oncontextmenu={handleContextDismiss}
    use:swipe={{
      axis: 'x',
      enableSwipe: true,
      threshold: 24,
      ignoreSelector: '.satellite-dismiss-btn',
      onDragStart: () => {
        isSwiping = true
      },
      onDragMove: (offset) => {
        swipeOffsetX = offset.x
      },
      onDragEnd: () => {
        isSwiping = false
        swipeOffsetX = 0
      },
      onSwipeLeft: () => {
        dismissSatelliteNotification()
      },
      onSwipeRight: () => {
        dismissSatelliteNotification()
      }
    }}
    style="
      transform: translateX({swipeOffsetX}px);
      opacity: {isSwiping ? Math.max(0.2, 1 - Math.abs(swipeOffsetX) / 90) : 1};
      transition: {isSwiping
      ? 'none'
      : 'transform 0.22s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.2s ease'};
    "
  >
    {#if currentNotif}
      {#key currentNotif}
        <div class="satellite-arrival-pulse" aria-hidden="true"></div>
        <div
          class="satellite-timer-ring"
          style="animation-duration: {durationSeconds}s;"
          aria-hidden="true"
        ></div>
      {/key}
    {/if}

    {#key currentNotif}
      <div class="satellite-core">
        <div class="satellite-avatar-wrap">
          {#if resolvedIcon}
            <img
              src={resolvedIcon}
              alt=""
              class="satellite-icon-img"
              onerror={(e) => {
                const target = e.currentTarget as HTMLImageElement
                target.style.display = 'none'
                const fallback = target.nextElementSibling as HTMLElement
                if (fallback) fallback.style.display = 'flex'
              }}
            />
            <div class="satellite-icon-fallback" style="display: none;">
              <Bell size={14} />
            </div>
          {:else}
            <div class="satellite-icon-fallback">
              <Bell size={14} />
            </div>
          {/if}
        </div>
      </div>
    {/key}

    {#if showMiniAppBadge && resolvedAppBadge}
      <img
        src={resolvedAppBadge}
        alt=""
        class="satellite-mini-app-badge"
        onerror={(e) => {
          ;(e.currentTarget as HTMLImageElement).style.display = 'none'
        }}
      />
    {/if}

    <button
      type="button"
      class="satellite-dismiss-btn"
      onclick={handleDismissClick}
      aria-label="Dismiss notification"
    >
      <X size={12} strokeWidth={2.5} />
    </button>
  </div>
</SatelliteHost>

<style src="./styles/NotificationSatellite.css"></style>
