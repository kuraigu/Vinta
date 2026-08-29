<script lang="ts">
  import { flip } from 'svelte/animate'
  import { fly, fade, slide } from 'svelte/transition'
  import { cubicOut } from 'svelte/easing'
  import {
    Bell,
    MessageSquareDot,
    Trash2,
    X,
    ChevronDown,
    ExternalLink,
    Calendar,
    MessageCircle,
    Gamepad2,
    Globe,
    Play,
    Settings,
    Terminal
  } from 'lucide-svelte'
  import { notifData, notificationPermissionDenied, dndStore } from '../store'
  import { swipe } from '../utils/useSwipe'
  import { onMount } from 'svelte'
  import { SvelteSet, SvelteMap } from 'svelte/reactivity'
  import { appIcons, gameIcons } from '../appIcons'
  import ScrollView from '../components/ScrollView.svelte'
  import { FocusNotificationApp, FocusWindow } from '../../wailsjs/go/backend/App'

  export type NotificationItem = {
    id: number
    appName?: string
    aumid?: string
    icon?: string
    title: string
    message: string
    heroImage?: string
    timestamp: number
  }

  let { expanded = false, notifications = [] } = $props<{
    expanded?: boolean
    notifications?: NotificationItem[]
  }>()

  let expandedIds = new SvelteSet<number>()

  function toggleExpand(id: number) {
    if (expandedIds.has(id)) {
      expandedIds.delete(id)
    } else {
      expandedIds.add(id)
    }
  }

  let expandedGroups = new SvelteSet<string>()

  function toggleGroupExpand(brandName: string) {
    if (expandedGroups.has(brandName)) {
      expandedGroups.delete(brandName)
    } else {
      expandedGroups.add(brandName)
    }
  }

  let itemDragOffsets = new SvelteMap<string, number>()

  function dismissGroup(brandName: string) {
    notifData.update((n) =>
      n.filter((x) => getDisplayBrand(x).name !== brandName)
    )
  }

  type NotificationGroup = {
    brandName: string
    brandIcon: string | null
    brandAppIcon: string | null
    brandCategory?: string
    items: typeof notifications
    latestTimestamp: number
  }

  // Shared deduplication filter: strip control chars and hide unidentified "Notification"
  // entries when an identified version of the same content (title+message) exists.
  // Sequential messages sent >5s apart (like repeated chat tests) are preserved.
  let filteredNotifications = $derived.by(() => {
    const normTitle = (t: string) => {
      // eslint-disable-next-line no-control-regex
      return (t || '').replace(/[\x00-\x1f]/g, '').trim()
    }
    // eslint-disable-next-line svelte/prefer-svelte-reactivity
    const identifiedKeys = new Map<string, number>()
    for (const notif of notifications) {
      if (notif.appName) {
        const key = normTitle(notif.title) + '\0' + normTitle(notif.message)
        identifiedKeys.set(key, notif.timestamp || Date.now())
      }
    }
    return notifications.filter((notif) => {
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
  })

  const imageSizingCache = new SvelteMap<string, 'avatar' | 'hero'>()

  function checkImageSizing(url: string) {
    if (!url || !isImageUrl(url) || imageSizingCache.has(url)) return
    const img = new Image()
    img.onload = () => {
      const w = img.naturalWidth
      const h = img.naturalHeight
      if (w > 0 && h > 0) {
        const ratio = w / h
        const isAvatar =
          ratio >= 0.75 && ratio <= 1.33 && (w <= 512 || h <= 512)
        imageSizingCache.set(url, isAvatar ? 'avatar' : 'hero')
        notifData.update((items) => {
          let changed = false
          for (const it of items) {
            if (it.heroImage === url && isAvatar) {
              it.icon = it.heroImage
              it.heroImage = ''
              changed = true
            }
          }
          return changed ? [...items] : items
        })
      }
    }
    img.src = url
  }

  let processedNotifications = $derived.by(() => {
    const items = filteredNotifications.map((n) => ({ ...n }))
    const senderAvatars: Record<string, string> = {}

    for (const notif of items) {
      if (notif.heroImage && isImageUrl(notif.heroImage)) {
        const sizing = imageSizingCache.get(notif.heroImage)
        if (sizing === 'avatar') {
          notif.icon = notif.heroImage
          notif.heroImage = ''
        } else if (!sizing) {
          checkImageSizing(notif.heroImage)
        }
      }

      const senderKey = (notif.title || '').trim().toLowerCase()
      if (senderKey && notif.icon && isImageUrl(notif.icon)) {
        if (!senderAvatars[senderKey]) {
          senderAvatars[senderKey] = notif.icon
        }
      }
    }

    for (const notif of items) {
      const senderKey = (notif.title || '').trim().toLowerCase()
      if (
        senderKey &&
        senderAvatars[senderKey] &&
        (!notif.icon || !isImageUrl(notif.icon))
      ) {
        notif.icon = senderAvatars[senderKey]
      }
    }

    return items
  })

  let groupedNotifications = $derived.by(() => {
    const groups: Record<string, NotificationGroup> = {}
    for (const notif of processedNotifications) {
      const brand = getDisplayBrand(notif)
      const groupKey = brand.name
      if (!groups[groupKey]) {
        groups[groupKey] = {
          brandName: groupKey,
          brandIcon: brand.appIcon || brand.icon,
          brandAppIcon: null,
          brandCategory: brand.category,
          items: [],
          latestTimestamp: 0
        }
      }
      groups[groupKey].items.push(notif)
      const ts = notif.timestamp || now
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
  // notifications — this prevents the pill from flashing "Notification" for the ~50s
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
        title: 'No notifications',
        message: ''
      } as (typeof notifications)[0])
    )
  })

  let now = $state(Date.now())

  onMount(() => {
    const interval = setInterval(() => {
      now = Date.now()
    }, 1000)
    return () => clearInterval(interval)
  })

  function getRelativeTime(timestamp?: number): string {
    if (!timestamp) return 'NOW'
    const diffSeconds = Math.max(0, Math.floor((now - timestamp) / 1000))
    if (diffSeconds < 5) return 'NOW'
    if (diffSeconds < 60) return `${diffSeconds}s ago`

    const diffMinutes = Math.floor(diffSeconds / 60)
    if (diffMinutes < 60) return `${diffMinutes}m ago`

    const diffHours = Math.floor(diffMinutes / 60)
    if (diffHours < 24) return `${diffHours}h ago`

    const diffDays = Math.floor(diffHours / 24)
    return `${diffDays}d ago`
  }

  let isConfirmingClear = $state(false)
  let clearConfirmTimeout: number

  function handleClearAll() {
    if (notifications.length <= 2 || isConfirmingClear) {
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

  function dismiss(id: number) {
    notifData.update((n) => n.filter((x) => x.id !== id))
  }

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

  function cleanAppName(appStr: string): string {
    if (!appStr) return 'Notification'
    return appStr
  }

  function getDisplayBrand(notif: {
    appName?: string
    title?: string
    message?: string
    icon?: string
    heroImage?: string
  }): {
    name: string
    icon: string | null
    appIcon: string | null
    category?: string
  } {
    const rawName = notif.appName || 'Notification'
    let defaultIcon = notif.icon || null

    const allIcons = [...gameIcons, ...appIcons]

    // 1. Direct appName matching
    const matchString = (notif.appName || '').toLowerCase()
    if (matchString && matchString !== 'notification') {
      for (const brand of allIcons) {
        if (brand.matchers.some((m) => matchString.includes(m))) {
          return {
            name: brand.name,
            icon: defaultIcon,
            appIcon: brand.icon,
            category: brand.category
          }
        }
      }
    }

    // 2. Fallback regex matching against title, message, and combined text
    const titleText = notif.title || ''
    const messageText = notif.message || ''
    const combined = `${titleText} ${messageText}`.trim()

    if (combined) {
      for (const brand of allIcons) {
        if (brand.fallbackRegexes) {
          for (const regex of brand.fallbackRegexes) {
            if (
              regex.test(titleText) ||
              regex.test(messageText) ||
              regex.test(combined)
            ) {
              return {
                name: brand.name,
                icon: defaultIcon,
                appIcon: brand.icon,
                category: brand.category
              }
            }
          }
        }
      }
    }

    let displayName = cleanAppName(rawName)
    if (displayName.toLowerCase() === (notif.title || '').toLowerCase()) {
      displayName = 'Notification'
    }

    return { name: displayName, icon: defaultIcon, appIcon: null }
  }
</script>

{#snippet fallbackIcon(
  brand: { category?: string; brandCategory?: string },
  size: number
)}
  {@const cat = brand.category || brand.brandCategory}
  {#if cat === 'communication'}
    <MessageCircle {size} color="var(--c-text-secondary)" />
  {:else if cat === 'productivity'}
    <Calendar {size} color="var(--c-text-secondary)" />
  {:else if cat === 'gaming'}
    <Gamepad2 {size} color="var(--c-text-secondary)" />
  {:else if cat === 'media'}
    <Play {size} color="var(--c-text-secondary)" />
  {:else if cat === 'browser'}
    <Globe {size} color="var(--c-text-secondary)" />
  {:else if cat === 'system'}
    <Settings {size} color="var(--c-text-secondary)" />
  {:else if cat === 'dev'}
    <Terminal {size} color="var(--c-text-secondary)" />
  {:else}
    <MessageSquareDot {size} color="var(--c-text-secondary)" />
  {/if}
{/snippet}

<div class="notification-container">
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
              {#if isImageUrl(brand.icon)}
                <img
                  src={brand.icon}
                  alt="app icon"
                  class="notif-img-icon"
                  onerror={(e) => {
                    const img = e.currentTarget as HTMLImageElement
                    img.style.display = 'none'
                    if (img.nextElementSibling) {
                      ;(img.nextElementSibling as HTMLElement).style.display =
                        'block'
                    }
                  }}
                />
                <div style="display: none;">
                  {@render fallbackIcon(brand, 18)}
                </div>
              {:else if brand.icon}
                <span class="icon-text">{brand.icon}</span>
              {:else}
                {@render fallbackIcon(brand, 18)}
              {/if}
            </div>
            {#if brand.appIcon && isImageUrl(brand.appIcon) && brand.appIcon !== brand.icon}
              <img
                src={brand.appIcon}
                alt="app badge"
                class="app-badge"
                onerror={(e) => {
                  ;(e.currentTarget as HTMLImageElement).style.display = 'none'
                }}
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
        </div>
      {/key}
    </div>
  {:else}
    <div class="expanded-view">
      <div class="notif-header">
        <h3>Notifications</h3>
        <div class="notif-header-actions">
          <button
            class="dnd-btn"
            class:dnd-active={$dndStore}
            onclick={() => dndStore.update((v) => !v)}
            title={$dndStore ? 'Do Not Disturb: ON' : 'Do Not Disturb: OFF'}
            aria-label="Toggle Do Not Disturb"
          >
            {#if $dndStore}
              <div class="dnd-zzz-icon">
                <Bell size={14} />
                <span class="zzz-badge">z<sup>Z</sup></span>
              </div>
            {:else}
              <Bell size={14} />
            {/if}
            <span>DND</span>
          </button>
          {#if notifications.length > 0}
            <button
              class="clear-btn"
              class:is-confirming={isConfirmingClear}
              onclick={handleClearAll}
              aria-label="Clear all notifications"
            >
              <Trash2 size={14} />
              <span>{isConfirmingClear ? 'Confirm?' : 'Clear'}</span>
            </button>
          {/if}
        </div>
      </div>
      {#if $notificationPermissionDenied}
        <div class="permission-banner">
          <Settings size={16} />
          <span
            >Vinta needs permission to read notifications. Please enable
            "Notification access" for apps in Windows Settings.</span
          >
        </div>
      {/if}
      {#snippet renderNotifItem(
        notif: NotificationItem,
        brand: NotificationGroup,
        nested: boolean
      )}
        {@const displayBrand = getDisplayBrand(notif)}
        {@const iconUrl = displayBrand.icon}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
          class="notif-item"
          class:nested
          class:is-swiping={itemDragOffsets.has(`item-${notif.id}`)}
          style="transform: translateX({itemDragOffsets.get(
            `item-${notif.id}`
          ) || 0}px); transition: {itemDragOffsets.has(`item-${notif.id}`)
            ? 'none'
            : 'transform 0.3s cubic-bezier(0.16, 1, 0.3, 1)'};"
          use:swipe={{
            enableSwipe: true,
            ignoreSelector:
              '.notif-actions, .expand-indicator, input, textarea, .no-swipe',
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
                >{getRelativeTime(notif.timestamp)}</span
              >
              <div class="notif-actions">
                <div
                  class="expand-indicator"
                  class:is-expanded={expandedIds.has(notif.id)}
                >
                  <ChevronDown size={14} />
                </div>
                <button
                  class="dismiss-item-btn"
                  onclick={(e) => {
                    e.stopPropagation()
                    dismiss(notif.id)
                  }}
                  aria-label="Dismiss notification"
                >
                  <X size={14} />
                </button>
              </div>
            </div>
          {/snippet}

          {#if !nested}
            <div class="notif-item-header">
              <div class="notif-item-app-group">
                {#if displayBrand.appIcon && isImageUrl(displayBrand.appIcon) && displayBrand.appIcon !== iconUrl}
                  <img
                    src={displayBrand.appIcon}
                    alt="app badge"
                    class="notif-header-app-icon"
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
            <div class="notif-item-avatar-col">
              <div class="icon-wrapper">
                <div class="notif-avatar-icon">
                  {#if isImageUrl(iconUrl)}
                    <img
                      src={iconUrl!}
                      alt="app icon"
                      class="notif-img-icon"
                      onerror={(e) => {
                        const img = e.currentTarget as HTMLImageElement
                        img.style.display = 'none'
                        if (img.nextElementSibling) {
                          ;(
                            img.nextElementSibling as HTMLElement
                          ).style.display = 'block'
                        }
                      }}
                    />
                    <div style="display: none;">
                      {#if displayBrand.appIcon && isImageUrl(displayBrand.appIcon)}
                        <img
                          src={displayBrand.appIcon}
                          alt="app icon"
                          class="notif-img-icon"
                          onerror={(e) => {
                            const img = e.currentTarget as HTMLImageElement
                            img.style.display = 'none'
                            if (img.nextElementSibling) {
                              ;(
                                img.nextElementSibling as HTMLElement
                              ).style.display = 'block'
                            }
                          }}
                        />
                        <div style="display: none;">
                          {@render fallbackIcon(displayBrand, 24)}
                        </div>
                      {:else}
                        {@render fallbackIcon(displayBrand, 24)}
                      {/if}
                    </div>
                  {:else if displayBrand.appIcon && isImageUrl(displayBrand.appIcon)}
                    <img
                      src={displayBrand.appIcon}
                      alt="app icon"
                      class="notif-img-icon"
                      onerror={(e) => {
                        const img = e.currentTarget as HTMLImageElement
                        img.style.display = 'none'
                        if (img.nextElementSibling) {
                          ;(
                            img.nextElementSibling as HTMLElement
                          ).style.display = 'block'
                        }
                      }}
                    />
                    <div style="display: none;">
                      {@render fallbackIcon(displayBrand, 24)}
                    </div>
                  {:else if iconUrl}
                    <span class="icon-text">{iconUrl}</span>
                  {:else}
                    {@render fallbackIcon(displayBrand, 24)}
                  {/if}
                </div>
              </div>
            </div>

            <div class="notif-item-content-col">
              <div class="notif-item-body">
                <div class="notif-item-title-row">
                  <span
                    class="notif-item-title"
                    class:is-expanded={expandedIds.has(notif.id)}
                  >
                    {notif.title}
                  </span>
                </div>
                {#if notif.message}
                  <p
                    class="notif-item-message"
                    class:is-expanded={expandedIds.has(notif.id)}
                    class:is-visible={nested || expandedIds.has(notif.id)}
                  >
                    {notif.message}
                  </p>
                {/if}
                {#if notif.heroImage && isImageUrl(notif.heroImage) && notif.heroImage !== displayBrand.icon && notif.heroImage !== notif.icon && (!nested || expandedIds.has(notif.id))}
                  <div class="hero-image-wrapper">
                    <img src={notif.heroImage} alt="hero" class="hero-image" />
                  </div>
                {/if}
                {#if expandedIds.has(notif.id) && (notif.appName || notif.aumid)}
                  <!-- svelte-ignore a11y_click_events_have_key_events -->
                  <!-- svelte-ignore a11y_no_static_element_interactions -->
                  <div class="open-app-hint" onclick={(e) => {
                    e.stopPropagation()
                    FocusNotificationApp(notif.appName ?? '', notif.aumid ?? '')
                    FocusWindow()
                  }}>
                    <ExternalLink size={11} />
                    <span>Open app</span>
                  </div>
                {/if}
              </div>
            </div>
          </div>
        </div>
      {/snippet}

      <ScrollView class="notif-list">
        {#each groupedNotifications as group (group.brandName)}
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
                class:is-expanded={expandedGroups.has(group.brandName)}
                class:is-swiping={itemDragOffsets.has(
                  `group-${group.brandName}`
                )}
                style="transform: translateX({itemDragOffsets.get(
                  `group-${group.brandName}`
                ) || 0}px); transition: {itemDragOffsets.has(
                  `group-${group.brandName}`
                )
                  ? 'none'
                  : 'transform 0.3s cubic-bezier(0.16, 1, 0.3, 1)'};"
                use:swipe={{
                  enableSwipe: true,
                  ignoreSelector:
                    '.notif-actions, .expand-indicator, input, textarea, .no-swipe',
                  onDragMove: (offset) => {
                    itemDragOffsets.set(`group-${group.brandName}`, offset.x)
                  },
                  onDragEnd: () => {
                    itemDragOffsets.delete(`group-${group.brandName}`)
                  },
                  onSwipeRight: () => {
                    dismissGroup(group.brandName)
                  },
                  onSwipeLeft: () => {
                    dismissGroup(group.brandName)
                  }
                }}
                onclick={() => {
                  if (!expandedGroups.has(group.brandName)) {
                    toggleGroupExpand(group.brandName)
                  }
                }}
              >
                <!-- svelte-ignore a11y_click_events_have_key_events -->
                <!-- svelte-ignore a11y_no_static_element_interactions -->
                <div
                  class="notif-group-header"
                  onclick={(e) => {
                    if (expandedGroups.has(group.brandName)) {
                      e.stopPropagation()
                      toggleGroupExpand(group.brandName)
                    }
                  }}
                >
                  <div class="notif-group-header-top">
                    <div class="notif-item-app-group">
                      <div class="icon-wrapper">
                        <div class="notif-item-icon">
                          {#if isImageUrl(group.brandIcon)}
                            <img
                              src={group.brandIcon}
                              alt="app icon"
                              class="notif-img-icon"
                              onerror={(e) => {
                                const img = e.currentTarget as HTMLImageElement
                                img.style.display = 'none'
                                if (img.nextElementSibling) {
                                  ;(
                                    img.nextElementSibling as HTMLElement
                                  ).style.display = 'block'
                                }
                              }}
                            />
                            <div style="display: none;">
                              {#if group.brandAppIcon && isImageUrl(group.brandAppIcon)}
                                <img
                                  src={group.brandAppIcon}
                                  alt="app icon"
                                  class="notif-img-icon"
                                  onerror={(e) => {
                                    const img =
                                      e.currentTarget as HTMLImageElement
                                    img.style.display = 'none'
                                    if (img.nextElementSibling) {
                                      ;(
                                        img.nextElementSibling as HTMLElement
                                      ).style.display = 'block'
                                    }
                                  }}
                                />
                                <div style="display: none;">
                                  {@render fallbackIcon(group, 16)}
                                </div>
                              {:else}
                                {@render fallbackIcon(group, 16)}
                              {/if}
                            </div>
                          {:else if group.brandIcon}
                            <span class="icon-text">{group.brandIcon}</span>
                          {:else}
                            {@render fallbackIcon(group, 16)}
                          {/if}
                        </div>
                        {#if group.brandAppIcon && isImageUrl(group.brandAppIcon) && group.brandAppIcon !== group.brandIcon}
                          <img
                            src={group.brandAppIcon}
                            alt="app badge"
                            class="app-badge"
                            onerror={(e) => {
                              ;(
                                e.currentTarget as HTMLImageElement
                              ).style.display = 'none'
                            }}
                          />
                        {/if}
                      </div>
                      <span class="notif-item-app">{group.brandName}</span>
                      <span class="notif-count-badge">{group.items.length}</span
                      >
                    </div>
                    <div class="time-and-dismiss">
                      {#if !expandedGroups.has(group.brandName)}
                        <span class="notif-item-time"
                          >{getRelativeTime(group.latestTimestamp)}</span
                        >
                      {/if}
                      <div class="notif-actions">
                        <div
                          class="expand-indicator"
                          class:is-expanded={expandedGroups.has(
                            group.brandName
                          )}
                        >
                          <ChevronDown size={14} />
                        </div>
                        <button
                          class="dismiss-item-btn"
                          onclick={(e) => {
                            e.stopPropagation()
                            dismissGroup(group.brandName)
                          }}
                          aria-label="Dismiss group"
                        >
                          <X size={14} />
                        </button>
                      </div>
                    </div>
                  </div>
                  {#if !expandedGroups.has(group.brandName)}
                    <div class="notif-group-preview">
                      <div class="notif-item-layout">
                        {#if group.items[0].icon && isImageUrl(group.items[0].icon) && group.items[0].icon !== group.brandIcon}
                          <div class="notif-item-avatar-col">
                            <div class="icon-wrapper">
                              <div class="notif-avatar-icon">
                                <img
                                  src={group.items[0].icon}
                                  alt="avatar"
                                  class="notif-img-icon"
                                />
                              </div>
                            </div>
                          </div>
                        {/if}
                        <div class="notif-item-content-col">
                          <div class="notif-item-body">
                            <div class="notif-item-title-row">
                              <span class="notif-item-title"
                                >{group.items[0].title}</span
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

                {#if expandedGroups.has(group.brandName)}
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
        {#if notifications.length === 0}
          <div class="empty-state">No new notifications</div>
        {/if}
      </ScrollView>
    </div>
  {/if}
</div>

<style src="./styles/Notification.css"></style>
