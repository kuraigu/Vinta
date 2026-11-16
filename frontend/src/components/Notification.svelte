<!--
  @fileoverview Notification history feed and management component.

  Captures Windows toast notifications via the backend service, groups notifications by
  application or sender, caches sender avatars, and provides inline dismiss controls.
-->
<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 360, h: 65, r: 18 },
    expanded: { w: 460, h: 410, r: 30 },
    borderColor: 'var(--vinta-island-border-notification)',
    glowColor: 'var(--vinta-island-glow-notification)'
  }
</script>

<script lang="ts">
  import { flip } from 'svelte/animate'
  import { fly, fade, slide } from 'svelte/transition'
  import { cubicOut } from 'svelte/easing'
  import { Events } from '$lib/events'
  import {
    Bell,
    MessageSquareDot,
    Trash2,
    X,
    ChevronDown,
    ChevronUp,
    ArrowUpRight,
    Calendar,
    MessageCircle,
    Gamepad2,
    Globe,
    Play,
    Settings,
    Terminal
  } from 'lucide-svelte'
  import {
    notifData,
    notificationPermissionDenied,
    dndStore,
    isIslandHiddenStore,
    tStore,
    t
  } from '../store'
  import { swipe } from '../utils/useSwipe'
  import { getContext, onMount } from 'svelte'
  import { SvelteSet, SvelteMap } from 'svelte/reactivity'
  import {
    nativeIconsStore,
    requestNativeAppIcon,
    parseAumid
  } from '../services/iconService'
  import ScrollView from '../components/ScrollView.svelte'
  import EmptyState from '../components/EmptyState.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import appIcon from '@appicon'
  import {
    FocusNotificationApp,
    FocusWindow
  } from '../../bindings/Vinta/backend/app.js'

  export type NotificationItem = {
    id: number
    nativeId?: number
    appName?: string
    aumid?: string
    icon?: string
    appLogo?: string
    title: string
    message: string
    heroImage?: string
    timestamp: number
    canOpen?: boolean
    isUpdate?: boolean
  }

  interface Props {
    expanded?: boolean
    inOverlay?: boolean
    notifications?: NotificationItem[]
  }

  let {
    expanded = false,
    inOverlay = false,
    notifications = undefined
  }: Props = $props()

  let activeNotifications = $derived(notifications ?? $notifData)

  let nativeIcons = $derived($nativeIconsStore)

  let expandedIds = new SvelteSet<number>()
  const foldMinimizedNotification = getContext<(() => void) | undefined>(
    'foldMinimizedNotification'
  )

  /**
   * Toggles single notification body expansion to view full message text.
   *
   * @param id - Unique notification record identifier.
   */
  function toggleExpand(id: number) {
    if (expandedIds.has(id)) {
      expandedIds.delete(id)
    } else {
      expandedIds.add(id)
    }
  }

  let expandedGroups = new SvelteSet<string>()

  /**
   * Toggles expansion state of a notification group accordion.
   *
   * @param groupKey - Group identifier string.
   */
  function toggleGroupExpand(groupKey: string) {
    if (expandedGroups.has(groupKey)) {
      expandedGroups.delete(groupKey)
    } else {
      expandedGroups.add(groupKey)
    }
  }

  let failedImageUrls = new SvelteSet<string>()

  function markImageFailed(url: string | undefined | null) {
    if (url) {
      failedImageUrls.add(url)
    }
  }

  function canRenderImage(url: string | undefined | null): url is string {
    return Boolean(url && isImageUrl(url) && !failedImageUrls.has(url))
  }

  let itemDragOffsets = new SvelteMap<string, number>()

  /**
   * Dismisses all notifications belonging to a specific application or sender group.
   *
   * @param groupKey - Group identifier string.
   */
  function dismissGroup(groupKey: string) {
    notifData.update((n) =>
      n.filter((x) => {
        const brand = getDisplayBrand(x)
        return brand.name !== groupKey
      })
    )
  }

  type NotificationGroup = {
    groupKey: string
    brandName: string
    headerTitle: string
    brandIcon: string | null
    brandAppIcon: string | null
    brandCategory?: string
    items: NotificationItem[]
    latestTimestamp: number
  }

  const SENDER_AVATAR_STORAGE_KEY = 'vinta_sender_avatars'

  /**
   * Retrieves the sender avatar mapping stored in persistent browser storage.
   * This preserves discovered contact profile pictures across app reloads.
   */
  function loadSenderAvatars(): Record<string, string> {
    try {
      const raw = localStorage.getItem(SENDER_AVATAR_STORAGE_KEY)
      if (raw) {
        const parsed = JSON.parse(raw)
        for (const k of Object.keys(parsed)) {
          if (k.toLowerCase().includes('vinta')) {
            delete parsed[k]
          }
        }
        return parsed
      }
    } catch {
      // Storage unavailable or unparseable
    }
    return {}
  }

  // Pure in-memory cache loaded once from localStorage, updated out-of-band without causing reactive loops
  const persistentSenderAvatars = loadSenderAvatars()

  /**
   * Persists a resolved avatar URL for a given sender identifier into local storage.
   */
  function saveSenderAvatar(key: string, url: string) {
    if (!key || !url) return
    if (persistentSenderAvatars[key] === url) return
    persistentSenderAvatars[key] = url
    try {
      localStorage.setItem(
        SENDER_AVATAR_STORAGE_KEY,
        JSON.stringify(persistentSenderAvatars)
      )
    } catch {
      // Storage write error
    }
  }

  /**
   * Generates a composite lookup key combining the source application with the sender title.
   */
  function getSenderKey(notif: {
    appName?: string
    aumid?: string
    title?: string
  }): string {
    const app = (notif.appName || notif.aumid || '').trim().toLowerCase()
    if (app.includes('vinta')) return ''
    const sender = (notif.title || '').trim().toLowerCase()
    if (!sender) return ''
    return `${app}::${sender}`
  }

  /**
   * Evaluates whether a notification's current icon is simply the generic application logo.
   */
  function isGenericAppLogo(notif: NotificationItem): boolean {
    if (!notif.icon) return true
    if (notif.appLogo && notif.icon === notif.appLogo) return true
    if (notif.icon === appIcon) return true
    const brand = getDisplayBrand(notif)
    if (brand.appIcon && notif.icon === brand.appIcon) return true
    return false
  }

  /**
   * Deduplicates notifications: strips control chars and hides unidentified entries
   * when an identified version of the same content exists.
   */
  function computeFilteredNotifications() {
    const normTitle = (t: string) => {
      // eslint-disable-next-line no-control-regex
      return (t || '').replace(/[\x00-\x1f]/g, '').trim()
    }
    // eslint-disable-next-line svelte/prefer-svelte-reactivity
    const identifiedKeys = new Map<string, number>()
    for (const notif of activeNotifications) {
      if (notif.appName) {
        const key = normTitle(notif.title) + '\0' + normTitle(notif.message)
        identifiedKeys.set(key, notif.timestamp || Date.now())
      }
    }
    return activeNotifications.filter((notif) => {
      if (!notif.appName) {
        const key = normTitle(notif.title) + '\0' + normTitle(notif.message)
        const identifiedTs = identifiedKeys.get(key)
        if (
          identifiedTs &&
          Math.abs((notif.timestamp || Date.now()) - identifiedTs) < 5000
        ) {
          return false
        }
      }
      return true
    })
  }

  let filteredNotifications = $derived.by(computeFilteredNotifications)

  const imageSizingCache = new SvelteMap<string, 'avatar' | 'hero'>()

  /**
   * Asynchronously inspects the natural dimensions of an incoming hero image.
   * When an incoming image is square, promote it to contact avatar and persist it.
   */
  function checkImageSizing(url: string) {
    if (!url || !isImageUrl(url) || imageSizingCache.has(url)) return
    const img = new Image()
    img.onload = () => {
      const w = img.naturalWidth
      const h = img.naturalHeight
      if (w > 0 && h > 0) {
        const ratio = w / h
        const isAvatar = ratio >= 0.75 && ratio <= 1.33
        imageSizingCache.set(url, isAvatar ? 'avatar' : 'hero')
        if (isAvatar) {
          notifData.update((items) => {
            let changed = false
            for (const it of items) {
              if (it.heroImage === url) {
                it.icon = url
                it.heroImage = ''
                changed = true
                const sKey = getSenderKey(it)
                if (sKey) {
                  saveSenderAvatar(sKey, url)
                }
              }
            }
            return changed ? [...items] : items
          })
        }
      }
    }
    img.src = url
  }

  /**
   * Derives display properties for notifications and restores persistent sender avatars.
   * Promotes avatar-sized hero images when the notification only has a generic app logo.
   */
  function computeProcessedNotifications() {
    const items = filteredNotifications.map((n) => ({ ...n }))

    for (const notif of items) {
      const sKey = getSenderKey(notif)

      // Promote avatar hero images when notification lacks a custom avatar
      if (notif.heroImage && isImageUrl(notif.heroImage)) {
        const isAvatarUrl =
          notif.heroImage.includes('/avatars/') ||
          notif.heroImage.includes('/avatar/') ||
          notif.heroImage.includes('avatar') ||
          notif.heroImage.includes('profile') ||
          notif.heroImage.includes('user')
        const sizing = imageSizingCache.get(notif.heroImage)
        if (isAvatarUrl || sizing === 'avatar') {
          notif.icon = notif.heroImage
          notif.heroImage = ''
          if (sKey) {
            saveSenderAvatar(sKey, notif.icon)
          }
        } else if (!sizing) {
          checkImageSizing(notif.heroImage)
        }
      }

      // Apply cached sender avatar when notification lacks a custom avatar
      if (sKey && isGenericAppLogo(notif) && persistentSenderAvatars[sKey]) {
        notif.icon = persistentSenderAvatars[sKey]
      }
    }

    return items
  }

  let processedNotifications = $derived.by(computeProcessedNotifications)

  /**
   * Persists newly resolved custom contact avatars to browser local storage.
   * Runs inside a reactive effect rather than during derived computations to keep derivations pure.
   */
  function handlePersistAvatarsEffect() {
    for (const notif of processedNotifications) {
      const sKey = getSenderKey(notif)
      if (
        sKey &&
        notif.icon &&
        isImageUrl(notif.icon) &&
        !isGenericAppLogo(notif)
      ) {
        saveSenderAvatar(sKey, notif.icon)
      }
    }
  }
  $effect(handlePersistAvatarsEffect)

  let groupedNotifications = $derived.by(() => {
    const groups: Record<string, NotificationGroup> = {}
    for (const notif of processedNotifications) {
      const brand = getDisplayBrand(notif)
      const headerTitle = (notif.title || '').trim() || 'Notification'
      const groupKey = brand.name
      if (!groups[groupKey]) {
        groups[groupKey] = {
          groupKey,
          brandName: brand.name,
          headerTitle,
          brandIcon: brand.appIcon,
          brandAppIcon: brand.appIcon,
          brandCategory: brand.category,
          items: [],
          latestTimestamp: 0
        }
      }
      groups[groupKey].items.push(notif)
      const ts = notif.timestamp || 0
      if (ts > groups[groupKey].latestTimestamp) {
        groups[groupKey].latestTimestamp = ts
      }
    }

    return Object.values(groups).sort(
      (a, b) => b.latestTimestamp - a.latestTimestamp
    )
  })

  // For the pill (minimized view), prefer the most recent IDENTIFIED notification.
  // Unidentified "Notification" entries are shown only when there are zero identified
  // notifications, which prevents the pill from flashing "Notification" for the ~50s
  // gap before the WPN copy arrives from dual-channel apps like Rise of Kingdoms.
  let latest = $derived.by(() => {
    const identified = processedNotifications.find(
      (n) =>
        (n.appName && n.appName !== 'Notification') ||
        getDisplayBrand(n).name !== 'Notification'
    )
    const fallback = processedNotifications[0]
    return (
      identified ||
      fallback ||
      ({
        id: 0,
        title: t('notification.empty_title'),
        message: ''
      } as NotificationItem)
    )
  })

  let now = $state(Date.now())

  function syncRelativeTimeTicker() {
    if ($isIslandHiddenStore || $notifData.length === 0) return

    const currentNow = Date.now()
    now = currentNow
    const hasFreshNotif = $notifData.some(
      (n) => typeof n.timestamp === 'number' && currentNow - n.timestamp < 60000
    )
    const intervalMs = hasFreshNotif ? 1000 : 15000
    const interval = setInterval(() => {
      now = Date.now()
    }, intervalMs)
    return () => clearInterval(interval)
  }
  $effect(syncRelativeTimeTicker)

  /**
   * Formats notification received timestamp into a human-readable relative time (e.g. "5m ago").
   *
   * @param timestamp - Millisecond epoch timestamp.
   * @param translate - Active translation lookup function.
   * @returns Relative time label.
   */
  function getRelativeTime(timestamp?: number, translate = t): string {
    if (!timestamp) return translate('notification.time_now')
    const diffSeconds = Math.max(0, Math.floor((now - timestamp) / 1000))
    if (diffSeconds < 5) return translate('notification.time_now')
    if (diffSeconds < 60) {
      return translate('notification.time_seconds_ago', { count: diffSeconds })
    }

    const diffMinutes = Math.floor(diffSeconds / 60)
    if (diffMinutes < 60) {
      return translate('notification.time_minutes_ago', { count: diffMinutes })
    }

    const diffHours = Math.floor(diffMinutes / 60)
    if (diffHours < 24) {
      return translate('notification.time_hours_ago', { count: diffHours })
    }

    const diffDays = Math.floor(diffHours / 24)
    return translate('notification.time_days_ago', { count: diffDays })
  }

  let isConfirmingClear = $state(false)
  let clearConfirmTimeout: number

  /**
   * Flushes all recorded notifications from persistent storage and resets view.
   */
  function handleClearAll() {
    if (activeNotifications.length <= 2 || isConfirmingClear) {
      clearTimeout(clearConfirmTimeout)
      isConfirmingClear = false
      notifData.set([])
    } else {
      isConfirmingClear = true
      clearTimeout(clearConfirmTimeout)
      clearConfirmTimeout = window.setTimeout(() => {
        isConfirmingClear = false
      }, 2500)
    }
  }

  /**
   * Dismisses a single notification and updates unread badge indicators.
   *
   * @param id - Unique notification record identifier.
   */
  function dismiss(id: number) {
    notifData.update((n) => n.filter((x) => x.id !== id))
  }

  /**
   * Evaluates whether an icon string points to a valid image URI or path.
   *
   * @param url - Image URI or filesystem path string.
   * @returns True if string represents an image target.
   */
  function isImageUrl(url: string | undefined): boolean {
    if (!url) return false
    return (
      url.startsWith('http') ||
      url.startsWith('data:') ||
      url.startsWith('/') ||
      url.startsWith('file:') ||
      url.startsWith('ms-appx:') ||
      url.startsWith('ms-appdata:') ||
      url.includes(':\\')
    )
  }

  /**
   * Determines whether clicking the notification card can bring the originating application forward.
   *
   * @param notif - Notification record.
   * @returns True if source application can be activated.
   */
  function canOpenApp(notif: NotificationItem): boolean {
    if (notif.isUpdate) return true
    if (notif.canOpen === false) return false
    if (notif.canOpen === true) return true
    const name = (notif.appName || '').trim().toLowerCase()
    const aumid = (notif.aumid || '').trim().toLowerCase()
    if (!name && !aumid) return false
    if (name.includes('vinta') || aumid.includes('vinta')) return false
    return true
  }

  /**
   * Resolves application brand name and display icon for a notification card.
   *
   * @param notif - Notification item record.
   * @returns Brand presentation metadata object.
   */
  function getDisplayBrand(notif: {
    appName?: string
    aumid?: string
    title?: string
    message?: string
    icon?: string
    appLogo?: string
    heroImage?: string
  }): {
    name: string
    icon: string | null
    appIcon: string | null
    category?: string
  } {
    const rawInput = notif.appName || notif.aumid || 'Notification'
    const isVinta =
      rawInput.toLowerCase().includes('vinta') ||
      (notif.title || '').toLowerCase().includes('vinta')

    if (isVinta) {
      const vintaIcon = notif.icon || appIcon
      return {
        name: 'Vinta',
        icon: vintaIcon,
        appIcon: notif.appLogo || appIcon,
        category: 'system'
      }
    }

    const parsed = parseAumid(rawInput)
    let defaultIcon = notif.icon || null
    let displayName = parsed.displayName || 'Notification'

    // Request native extracted icon from Windows (.exe / AUMID)
    const searchApp = (parsed.processName || notif.appName || '').trim()
    const searchAumid = (notif.aumid || '').trim()
    const iconKey = `${searchApp.toLowerCase()}|${searchAumid.toLowerCase()}`
    const nativeUri =
      nativeIcons[iconKey] || requestNativeAppIcon(searchApp, searchAumid)

    const effectiveAppIcon = notif.appLogo || nativeUri || null
    const effectiveIcon = defaultIcon || nativeUri || null

    return {
      name: displayName,
      icon: effectiveIcon,
      appIcon: effectiveAppIcon,
      category: 'system'
    }
  }
</script>

{#snippet fallbackIcon(
  brand: { category?: string; brandCategory?: string },
  size: number
)}
  {@const cat = brand.category || brand.brandCategory}
  {#if cat === 'communication'}
    <MessageCircle {size} color="var(--vinta-text-secondary)" />
  {:else if cat === 'productivity'}
    <Calendar {size} color="var(--vinta-text-secondary)" />
  {:else if cat === 'gaming'}
    <Gamepad2 {size} color="var(--vinta-text-secondary)" />
  {:else if cat === 'media'}
    <Play {size} color="var(--vinta-text-secondary)" />
  {:else if cat === 'browser'}
    <Globe {size} color="var(--vinta-text-secondary)" />
  {:else if cat === 'system'}
    <Settings {size} color="var(--vinta-text-secondary)" />
  {:else if cat === 'dev'}
    <Terminal {size} color="var(--vinta-text-secondary)" />
  {:else}
    <MessageSquareDot {size} color="var(--vinta-text-secondary)" />
  {/if}
{/snippet}

<div
  data-vinta="notification"
  class="notification-container"
  class:minimal={!expanded}
>
  {#if !expanded}
    <div class="banner-view-container">
      {#key latest.id}
        {@const brand = getDisplayBrand(latest)}
        <div
          class="banner-view"
          in:fly={{ y: -10, duration: 300 }}
          out:fade={{ duration: 150 }}
        >
          <div class="icon-wrapper">
            <div class="banner-icon">
              {#if canRenderImage(brand.icon)}
                <img
                  src={brand.icon}
                  alt={brand.name}
                  class="notif-img-icon"
                  onerror={() => markImageFailed(brand.icon)}
                />
              {:else if brand.icon && !isImageUrl(brand.icon)}
                <span class="icon-text">{brand.icon}</span>
              {:else}
                {@render fallbackIcon(brand, 18)}
              {/if}
            </div>
            {#if canRenderImage(brand.appIcon) && brand.appIcon !== brand.icon}
              <img
                src={brand.appIcon}
                alt={brand.name}
                class="app-badge"
                onerror={() => markImageFailed(brand.appIcon)}
              />
            {/if}
          </div>
          <div class="banner-content">
            <span class="banner-app">{brand.name}</span>
            <span class="banner-title">{latest.title}</span>
            {#if latest.message}
              <span class="banner-message">{latest.message}</span>
            {/if}
          </div>
          {#if canRenderImage(latest.heroImage) && latest.heroImage !== brand.icon && latest.heroImage !== brand.appIcon}
            <div class="banner-hero-thumb-wrapper">
              <img
                src={latest.heroImage}
                alt={latest.title}
                class="banner-hero-thumb"
                onerror={() => markImageFailed(latest.heroImage)}
              />
            </div>
          {/if}
          {#if foldMinimizedNotification}
            <button
              type="button"
              class="banner-fold-btn"
              onclick={(e) => {
                e.stopPropagation()
                foldMinimizedNotification()
              }}
              use:islandTooltip={{
                text: $tStore('notification.fold'),
                position: 'bottom'
              }}
              aria-label={$tStore('notification.fold')}
            >
              <ChevronUp size={15} />
            </button>
          {/if}
        </div>
      {/key}
    </div>
  {:else}
    <div class="expanded-view" class:in-overlay={inOverlay}>
      {#if !inOverlay}
        <div class="notif-header">
          <h3>{$tStore('notification.page_title')}</h3>
          <div class="notif-header-actions">
            <button
              type="button"
              class="dnd-btn"
              class:dnd-active={$dndStore}
              onclick={() => dndStore.update((v) => !v)}
              use:islandTooltip={$dndStore
                ? $tStore('notification.dnd_on')
                : $tStore('notification.dnd_off')}
              aria-label={$tStore('notification.dnd_aria')}
            >
              {#if $dndStore}
                <div class="dnd-zzz-icon">
                  <Bell size={14} />
                  <span class="zzz-badge">z<sup>Z</sup></span>
                </div>
              {:else}
                <Bell size={14} />
              {/if}
            </button>
            {#if activeNotifications.length > 0}
              <button
                type="button"
                class="clear-btn"
                class:is-confirming={isConfirmingClear}
                onclick={handleClearAll}
                use:islandTooltip={isConfirmingClear
                  ? $tStore('notification.clear_confirm_tooltip')
                  : $tStore('notification.clear_all')}
                aria-label={$tStore('notification.clear_aria')}
              >
                <Trash2 size={14} />
                {#if isConfirmingClear}
                  <span>{$tStore('notification.confirm_clear')}</span>
                {/if}
              </button>
            {/if}
          </div>
        </div>
      {/if}
      {#if $notificationPermissionDenied}
        <div class="permission-banner">
          <Settings size={16} />
          <span>{$tStore('notification.permission_required')}</span>
        </div>
      {/if}
      {#snippet renderNotifItem(
        notif: NotificationItem,
        brand: NotificationGroup,
        nested: boolean
      )}
        {@const displayBrand = getDisplayBrand(notif)}
        {@const iconUrl = displayBrand.icon}
        {@const isAvatarIcon = Boolean(iconUrl && !isGenericAppLogo(notif))}
        {@const isExpanded = expandedIds.has(notif.id)}
        {@const canOpen = canOpenApp(notif)}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
          class="notif-item"
          class:nested
          class:is-expanded={isExpanded}
          class:can-open={isExpanded && canOpen}
          class:is-swiping={itemDragOffsets.has(`item-${notif.id}`)}
          style="transform: translateX({itemDragOffsets.get(
            `item-${notif.id}`
          ) || 0}px); opacity: {itemDragOffsets.has(`item-${notif.id}`)
            ? Math.max(
                0.2,
                1 - Math.abs(itemDragOffsets.get(`item-${notif.id}`) || 0) / 220
              )
            : 1}; transition: {itemDragOffsets.has(`item-${notif.id}`)
            ? 'none'
            : 'transform 0.25s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.2s ease'};"
          use:swipe={{
            axis: 'x',
            enableSwipe: true,
            threshold: 45,
            ignoreSelector:
              '.notif-actions, .expand-indicator, .expand-indicator-btn, input, textarea, .no-swipe',
            onDragMove: (offset) => {
              itemDragOffsets.set(`item-${notif.id}`, offset.x)
            },
            onDragEnd: () => {
              itemDragOffsets.delete(`item-${notif.id}`)
            },
            onSwipeRight: () => {
              dismiss(notif.id)
            },
            onSwipeLeft: () => {
              dismiss(notif.id)
            }
          }}
          onclick={(e) => {
            if (nested && !expandedGroups.has(brand.brandName)) {
              e.stopPropagation()
              toggleGroupExpand(brand.brandName)
            } else {
              e.stopPropagation()
              toggleExpand(notif.id)
            }
          }}
        >
          {#snippet timeAndDismiss()}
            <div class="time-and-dismiss">
              <span class="notif-item-time"
                >{getRelativeTime(notif.timestamp, $tStore)}</span
              >
              <div class="notif-actions">
                <button
                  type="button"
                  class="expand-indicator-btn"
                  class:is-expanded={isExpanded}
                  onclick={(e) => {
                    e.stopPropagation()
                    toggleExpand(notif.id)
                  }}
                  use:islandTooltip={isExpanded
                    ? $tStore('notification.collapse_item')
                    : $tStore('notification.expand_item')}
                  aria-label={isExpanded
                    ? $tStore('notification.collapse_item')
                    : $tStore('notification.expand_item')}
                >
                  <ChevronDown size={14} />
                </button>
                <button
                  type="button"
                  class="dismiss-item-btn"
                  onclick={(e) => {
                    e.stopPropagation()
                    dismiss(notif.id)
                  }}
                  aria-label={$tStore('notification.dismiss_tooltip')}
                >
                  <X size={14} />
                </button>
              </div>
            </div>
          {/snippet}

          {#if !nested}
            <div class="notif-item-header">
              <div class="notif-item-app-group">
                {#if canRenderImage(displayBrand.appIcon) && displayBrand.appIcon !== iconUrl}
                  <img
                    src={displayBrand.appIcon}
                    alt={brand.brandName}
                    class="notif-header-app-icon"
                    onerror={() => markImageFailed(displayBrand.appIcon)}
                  />
                {/if}
                <span class="notif-item-app">{brand.brandName}</span>
              </div>
              {@render timeAndDismiss()}
            </div>
          {/if}

          {#if nested}
            <div class="nested-time-and-dismiss">
              {@render timeAndDismiss()}
            </div>
          {/if}

          <div class="notif-item-layout">
            {#if !nested || isAvatarIcon}
              <div class="notif-item-avatar-col">
                <div class="icon-wrapper">
                  <div class="notif-avatar-icon">
                    {#if canRenderImage(iconUrl)}
                      <img
                        src={iconUrl}
                        alt={displayBrand.name}
                        class="notif-img-icon"
                        onerror={() => markImageFailed(iconUrl)}
                      />
                    {:else if canRenderImage(displayBrand.appIcon)}
                      <img
                        src={displayBrand.appIcon}
                        alt={displayBrand.name}
                        class="notif-img-icon"
                        onerror={() => markImageFailed(displayBrand.appIcon)}
                      />
                    {:else if iconUrl && !isImageUrl(iconUrl)}
                      <span class="icon-text">{iconUrl}</span>
                    {:else}
                      {@render fallbackIcon(displayBrand, 24)}
                    {/if}
                  </div>
                </div>
              </div>
            {/if}

            <div class="notif-item-content-col">
              <div class="notif-item-body">
                <div class="notif-item-title-row">
                  <span class="notif-item-title" class:is-expanded={isExpanded}>
                    {notif.title}
                  </span>
                </div>
                {#if notif.message}
                  <p
                    class="notif-item-message"
                    class:is-expanded={isExpanded}
                    class:is-visible={nested || isExpanded}
                  >
                    {notif.message}
                  </p>
                {/if}
                {#if canRenderImage(notif.heroImage) && notif.heroImage !== notif.icon && notif.heroImage !== displayBrand.icon && notif.heroImage !== displayBrand.appIcon && (!nested || isExpanded)}
                  <div class="hero-image-wrapper">
                    <img
                      src={notif.heroImage}
                      alt={notif.title}
                      class="hero-image"
                      onerror={() => markImageFailed(notif.heroImage)}
                    />
                  </div>
                {/if}
                {#if isExpanded && canOpen}
                  <div class="notif-item-actions-row">
                    <button
                      type="button"
                      class="open-app-btn"
                      onclick={(e) => {
                        e.stopPropagation()
                        if (notif.isUpdate) {
                          Events.Emit('vinta:navigate-settings', {
                            tab: 'updates'
                          })
                          return
                        }
                        FocusNotificationApp(
                          notif.appName ?? '',
                          notif.aumid ?? ''
                        )
                        FocusWindow()
                      }}
                      use:islandTooltip={notif.isUpdate
                        ? $tStore('notification.view_update_details')
                        : $tStore('notification.open_in_app', {
                            app: notif.appName || brand.brandName
                          })}
                    >
                      <ArrowUpRight size={12} />
                      <span
                        >{notif.isUpdate
                          ? $tStore('notification.view_update')
                          : $tStore('notification.open')}</span
                      >
                    </button>
                  </div>
                {/if}
              </div>
            </div>
          </div>
        </div>
      {/snippet}

      <ScrollView class="notif-list">
        {#each groupedNotifications as group (group.groupKey)}
          <div
            animate:flip={{ duration: 300 }}
            in:fly={{ x: -20, duration: 300 }}
            out:fade={{ duration: 150 }}
            style="width: 100%; display: flex; flex-direction: column;"
          >
            {#if group.items.length === 1}
              {@render renderNotifItem(group.items[0], group, false)}
            {:else}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div
                class="notif-group unified-card"
                class:is-expanded={expandedGroups.has(group.groupKey)}
                class:is-swiping={itemDragOffsets.has(
                  `group-${group.groupKey}`
                )}
                style="transform: translateX({itemDragOffsets.get(
                  `group-${group.groupKey}`
                ) || 0}px); opacity: {itemDragOffsets.has(
                  `group-${group.groupKey}`
                )
                  ? Math.max(
                      0.2,
                      1 -
                        Math.abs(
                          itemDragOffsets.get(`group-${group.groupKey}`) || 0
                        ) /
                          220
                    )
                  : 1}; transition: {itemDragOffsets.has(
                  `group-${group.groupKey}`
                )
                  ? 'none'
                  : 'transform 0.25s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.2s ease'};"
                use:swipe={{
                  axis: 'x',
                  enableSwipe: true,
                  threshold: 45,
                  ignoreSelector:
                    '.notif-actions, .expand-indicator, input, textarea, .no-swipe',
                  onDragMove: (offset) => {
                    itemDragOffsets.set(`group-${group.groupKey}`, offset.x)
                  },
                  onDragEnd: () => {
                    itemDragOffsets.delete(`group-${group.groupKey}`)
                  },
                  onSwipeRight: () => {
                    dismissGroup(group.groupKey)
                  },
                  onSwipeLeft: () => {
                    dismissGroup(group.groupKey)
                  }
                }}
                onclick={() => {
                  if (!expandedGroups.has(group.groupKey)) {
                    toggleGroupExpand(group.groupKey)
                  }
                }}
              >
                <!-- svelte-ignore a11y_click_events_have_key_events -->
                <!-- svelte-ignore a11y_no_static_element_interactions -->
                <div
                  class="notif-group-header"
                  onclick={(e) => {
                    if (expandedGroups.has(group.groupKey)) {
                      e.stopPropagation()
                      toggleGroupExpand(group.groupKey)
                    }
                  }}
                >
                  <div class="notif-group-header-top">
                    <div class="notif-item-app-group">
                      <div class="notif-item-icon">
                        {#if canRenderImage(group.brandAppIcon)}
                          <img
                            src={group.brandAppIcon}
                            alt={group.brandName}
                            class="notif-img-icon"
                            onerror={() => markImageFailed(group.brandAppIcon)}
                          />
                        {:else if group.brandIcon && !isImageUrl(group.brandIcon)}
                          <span class="icon-text">{group.brandIcon}</span>
                        {:else}
                          {@render fallbackIcon(group, 16)}
                        {/if}
                      </div>
                      <span class="notif-item-app">{group.brandName}</span>
                      <span class="notif-count-badge">{group.items.length}</span
                      >
                    </div>
                    <div class="time-and-dismiss">
                      {#if !expandedGroups.has(group.groupKey)}
                        <span class="notif-item-time"
                          >{getRelativeTime(
                            group.latestTimestamp,
                            $tStore
                          )}</span
                        >
                      {/if}
                      <div class="notif-actions">
                        <button
                          type="button"
                          class="expand-indicator-btn"
                          class:is-expanded={expandedGroups.has(group.groupKey)}
                          onclick={(e) => {
                            e.stopPropagation()
                            toggleGroupExpand(group.groupKey)
                          }}
                          use:islandTooltip={expandedGroups.has(group.groupKey)
                            ? $tStore('notification.fold')
                            : $tStore('notification.expand')}
                          aria-label={expandedGroups.has(group.groupKey)
                            ? $tStore('notification.fold')
                            : $tStore('notification.expand')}
                        >
                          {#if expandedGroups.has(group.groupKey)}
                            <ChevronUp size={14} />
                          {:else}
                            <ChevronDown size={14} />
                          {/if}
                        </button>
                        <button
                          type="button"
                          class="dismiss-item-btn"
                          onclick={(e) => {
                            e.stopPropagation()
                            dismissGroup(group.groupKey)
                          }}
                          aria-label={$tStore('notification.dismiss_group')}
                        >
                          <X size={14} />
                        </button>
                      </div>
                    </div>
                  </div>
                  {#if !expandedGroups.has(group.groupKey)}
                    <div class="notif-group-preview">
                      <div class="notif-item-layout">
                        {#if canRenderImage(group.items[0].icon) && !isGenericAppLogo(group.items[0])}
                          <div class="notif-item-avatar-col">
                            <div class="icon-wrapper">
                              <div class="notif-avatar-icon">
                                <img
                                  src={group.items[0].icon}
                                  alt={group.brandName}
                                  class="notif-img-icon"
                                  onerror={() =>
                                    markImageFailed(group.items[0].icon)}
                                />
                              </div>
                            </div>
                          </div>
                        {/if}
                        <div class="notif-item-content-col">
                          <div class="notif-item-body">
                            <div class="notif-item-title-row">
                              <span class="notif-item-title"
                                >{group.headerTitle}</span
                              >
                            </div>
                            {#if group.items[0].message}
                              <p class="notif-item-message is-visible">
                                {group.items[0].message}
                              </p>
                            {/if}
                          </div>
                        </div>
                      </div>
                    </div>
                  {/if}
                </div>

                {#if expandedGroups.has(group.groupKey)}
                  <div
                    class="notif-group-items"
                    transition:slide={{ duration: 250, easing: cubicOut }}
                  >
                    {#each group.items as notif (notif.id)}
                      {@render renderNotifItem(notif, group, true)}
                    {/each}
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
        {#if activeNotifications.length === 0}
          <EmptyState
            icon={Bell}
            title={$tStore('notification.empty_title')}
            description={$tStore('notification.empty_desc')}
          />
        {/if}
      </ScrollView>
    </div>
  {/if}
</div>

<style src="./styles/Notification.css"></style>
