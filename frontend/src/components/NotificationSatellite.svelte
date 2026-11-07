<!--
  @component NotificationSatellite
  @description Floating toast notification pill. Slides in beside the island when a
  system notification arrives, shows app icon and preview text, pauses dismiss countdown
  on mouse hover, and opens the notification on click.
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
    notifSatelliteEnabledStore
  } from '../stores/settings'
  import {
    nativeIconsStore,
    parseAumid,
    requestNativeAppIcon
  } from '../services/iconService'
  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'
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
  }

  let {
    active = true,
    primaryCenter = 960,
    clampedOffsetX = 0,
    monitorLeft = 0,
    activeMonitorWidth = 1920,
    islandWidth = 368
  }: Props = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let currentNotif = $derived($satelliteNotification)
  let nativeIcons = $derived($nativeIconsStore)

  /**
   * Computes the auto-hide timeout duration in seconds for the satellite bar.
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

  const satelliteWidth = 236
  const satelliteHeight = 44
  const satelliteRadii: FourPartRadii = [0, 0, 16, 16]
  const satelliteGaps: FourPartGaps = [0, 12, 0, 12]

  /**
   * Computes whether the satellite should dock to the left or right flank of the island.
   * If the island is pushed to the far left of the monitor such that the left flank
   * would clip off-screen, dynamically switch anchoring to the right flank.
   */
  function computeSatelliteAnchor(): 'left' | 'right' {
    const islandCenterOnScreen = primaryCenter + clampedOffsetX - monitorLeft
    const halfIslandW = islandWidth / 2
    const leftEdge = islandCenterOnScreen - halfIslandW
    const leftClearance = leftEdge - 12 - satelliteWidth
    const rightEdge = islandCenterOnScreen + halfIslandW
    const rightClearance =
      activeMonitorWidth - (rightEdge + 12 + satelliteWidth)

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
   * Resolves the primary visual icon or contact avatar for the alert.
   */
  function resolveIcon(): string | null {
    if (!currentNotif) return null

    const rawInput = currentNotif.appName || currentNotif.aumid || ''
    const isVinta =
      rawInput.toLowerCase().includes('vinta') ||
      (currentNotif.title || '').toLowerCase().includes('vinta')

    if (isVinta) {
      return currentNotif.icon || currentNotif.appLogo || appIcon
    }

    if (currentNotif.icon && isImageUrl(currentNotif.icon)) {
      return currentNotif.icon
    }

    if (currentNotif.appLogo && isImageUrl(currentNotif.appLogo)) {
      return currentNotif.appLogo
    }

    const parsed = parseAumid(rawInput)
    const searchApp = (parsed.processName || currentNotif.appName || '').trim()
    const searchAumid = (currentNotif.aumid || '').trim()
    const iconKey = `${searchApp.toLowerCase()}|${searchAumid.toLowerCase()}`

    if (nativeIcons[iconKey]) {
      return nativeIcons[iconKey]
    }
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

  let resolvedIcon = $derived(resolveIcon())

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

  let displayMessage = $derived.by(() => {
    if (!currentNotif) return ''
    if (currentNotif.message && currentNotif.message.trim()) {
      return currentNotif.message.trim()
    }
    if (currentNotif.appName && currentNotif.appName !== displayTitle) {
      return currentNotif.appName
    }
    return ''
  })

  /**
   * Navigates to the Notification Center page and dismisses the satellite capsule.
   */
  function handleCapsuleClick() {
    if (navigate) {
      navigate(NotificationPage, notifConfig, true)
    }
    dismissSatelliteNotification()
  }

  /**
   * Explicitly closes the notification alert.
   */
  function handleDismissClick(e: MouseEvent) {
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
  width={satelliteWidth}
  height={satelliteHeight}
  radius={satelliteRadii}
  gap={satelliteGaps}
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
    onclick={handleCapsuleClick}
    use:swipe={{
      axis: 'x',
      enableSwipe: true,
      threshold: 30,
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
      opacity: {isSwiping
      ? Math.max(0.2, 1 - Math.abs(swipeOffsetX) / 140)
      : 1};
      transition: {isSwiping
      ? 'none'
      : 'transform 0.22s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.2s ease'};
    "
  >
    {#if currentNotif}
      {#key currentNotif}
        <div
          class="satellite-timeout-bar"
          style="animation-duration: {durationSeconds}s;"
        ></div>
      {/key}
    {/if}

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
          <Bell size={12} />
        </div>
      {:else}
        <div class="satellite-icon-fallback">
          <Bell size={12} />
        </div>
      {/if}
    </div>

    <div class="satellite-content">
      <span class="satellite-header" title={displayTitle}>{displayTitle}</span>
      {#if displayMessage}
        <span class="satellite-snippet" title={displayMessage}
          >{displayMessage}</span
        >
      {/if}
    </div>

    <button
      type="button"
      class="satellite-dismiss-btn"
      onclick={handleDismissClick}
      aria-label="Dismiss notification"
    >
      <X size={12} />
    </button>
  </div>
</SatelliteHost>

<style src="./styles/NotificationSatellite.css"></style>
