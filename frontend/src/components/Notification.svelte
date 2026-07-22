<script lang="ts">
  import { flip } from 'svelte/animate'
  import { fly, fade, slide } from 'svelte/transition'
  import { cubicOut } from 'svelte/easing'
  import {
    Bell,
    BellOff,
    MessageSquareDot,
    Trash2,
    X,
    ChevronDown,
    Calendar,
    MessageCircle,
    Gamepad2,
    Globe,
    Play,
    Settings,
    Terminal
  } from 'lucide-svelte'
  import { notifData, notificationPermissionDenied, dndStore } from '../store'
  import { FocusWindow } from '../../wailsjs/go/backend/App.js'
  import { onMount } from 'svelte'
  import { SvelteSet } from 'svelte/reactivity'
  import { appIcons, gameIcons, getCachedBase64Url } from '../appIcons'

  export type NotificationItem = {
    id: number
    appName?: string
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

  let groupedNotifications = $derived.by(() => {
    const groups: Record<string, NotificationGroup> = {}
    for (const notif of filteredNotifications) {
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
    
    // Backfill avatar icons per sender (by title/sender handle) so rapid notifications inherit their own valid avatar
    for (const group of Object.values(groups)) {
      const senderAvatars: Record<string, string> = {}
      for (const item of group.items) {
        const senderKey = (item.title || '').trim().toLowerCase()
        if (senderKey && item.icon && isImageUrl(item.icon)) {
          if (!senderAvatars[senderKey]) {
            senderAvatars[senderKey] = item.icon
          }
        }
      }

      for (const item of group.items) {
        const senderKey = (item.title || '').trim().toLowerCase()
        if (senderKey && senderAvatars[senderKey]) {
          if (!item.icon || !isImageUrl(item.icon)) {
            item.icon = senderAvatars[senderKey]
          }
        }
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
    const identified = filteredNotifications.find((n) => n.appName)
    const fallback = filteredNotifications[0]
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
    icon?: string
  }): {
    name: string
    icon: string | null
    appIcon: string | null
    category?: string
  } {
    const rawName = notif.appName || 'Notification'
    const defaultIcon = notif.icon || null

    const matchString = (notif.appName || '').toLowerCase()
    if (matchString) {
      const allIcons = [...gameIcons, ...appIcons]
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
              <img src={brand.appIcon} alt="app badge" class="app-badge" />
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
              <BellOff size={14} />
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
      <div class="notif-list">
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
                        src={getCachedBase64Url(iconUrl!)}
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
                          <img src={displayBrand.appIcon} alt="app icon" class="notif-img-icon" />
                        {:else}
                          {@render fallbackIcon(displayBrand, 24)}
                        {/if}
                      </div>
                    {:else if displayBrand.appIcon && isImageUrl(displayBrand.appIcon)}
                      <img src={displayBrand.appIcon} alt="app icon" class="notif-img-icon" />
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
                  {#if notif.heroImage && isImageUrl(notif.heroImage) && expandedIds.has(notif.id)}
                    <div class="hero-image-wrapper">
                      <img
                        src={notif.heroImage}
                        alt="hero"
                        class="hero-image"
                      />
                    </div>
                  {/if}
                </div>
              </div>
            </div>
          </div>
        {/snippet}

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
                              {@render fallbackIcon(group, 16)}
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
                      <span class="notif-item-title"
                        >{group.items[0].title}</span
                      >
                      {#if group.items[0].message}
                        <p class="notif-item-message is-visible">
                          {group.items[0].message}
                        </p>
                      {/if}
                    </div>
                  {/if}
                </div>

                {#if expandedGroups.has(group.brandName)}
                  <div class="notif-group-items">
                    {#each group.items as notif (notif.id)}
                      <div
                        transition:slide|local={{
                          duration: 250,
                          easing: cubicOut
                        }}
                      >
                        <div
                          in:fade={{ duration: 150, delay: 50 }}
                          out:fade={{ duration: 100 }}
                        >
                          {@render renderNotifItem(notif, group, true)}
                        </div>
                      </div>
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
      </div>
    </div>
  {/if}
</div>

<style>
  .notification-container {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    color: var(--c-text-primary);
  }

  .banner-view-container {
    width: 100%;
    height: 100%;
    display: grid;
    place-items: center;
  }

  .banner-view {
    grid-area: 1 / 1;
    display: flex;
    align-items: center;
    width: 100%;
    height: 100%;
    padding: 0 24px;
    box-sizing: border-box;
    gap: 16px;
  }

  .icon-wrapper {
    position: relative;
    display: inline-flex;
    flex-shrink: 0;
  }

  .app-badge {
    position: absolute;
    bottom: -6px;
    right: -6px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 2px solid var(--c-bg);
    background: var(--c-bg);
    object-fit: contain;
    z-index: 2;
  }

  .banner-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    flex-shrink: 0;
    font-size: 22px;
    background: var(--c-item-bg);
    border-radius: 8px;
    overflow: hidden;
  }

  .notif-img-icon {
    width: 100%;
    height: 100%;
    object-fit: contain;
    border-radius: 4px;
  }

  .banner-content {
    display: flex;
    flex-direction: column;
    justify-content: center;
    text-align: left;
    flex-grow: 1;
    min-width: 0;
  }

  .banner-app {
    font-size: 0.7rem;
    color: var(--c-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 700;
    line-height: 1;
    margin-bottom: 0px;
  }

  .banner-title {
    display: -webkit-box;
    -webkit-line-clamp: 1;
    line-clamp: 1;
    -webkit-box-orient: vertical;
    width: 100%;
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--c-text-primary);
    line-height: 1.2;
    margin-bottom: 0px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .banner-message {
    display: -webkit-box;
    -webkit-line-clamp: 1;
    line-clamp: 1;
    -webkit-box-orient: vertical;
    width: 100%;
    font-size: 0.85rem;
    color: var(--c-text-secondary);
    line-height: 1.3;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .hero-image-wrapper {
    margin-top: 8px;
    width: 100%;
    border-radius: 6px;
    overflow: hidden;
    display: flex;
    justify-content: center;
    background: var(--c-subtle-bg, rgba(0, 0, 0, 0.1));
  }

  .hero-image {
    max-width: 100%;
    width: auto;
    height: auto;
    display: block;
    object-fit: contain;
    max-height: 120px;
  }

  .expanded-view {
    padding: 20px 20px 60px 20px;
    height: 100%;
    width: 100%;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
  }

  .notif-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 2px;
    width: 100%;
    margin-bottom: 24px;
    flex-shrink: 0;
  }

  h3 {
    color: var(--c-text-primary);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.02em;
  }

  .dismiss-item-btn {
    background: var(--c-contrast-bg);
    border: 1px solid var(--c-contrast-bg);
    color: var(--c-text-secondary);
    width: 26px;
    height: 26px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .dismiss-item-btn:hover {
    background: var(--c-danger-hover-bg);
    color: var(--c-danger);
    border-color: var(--c-danger);
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(255, 59, 48, 0.2);
  }

  .clear-btn,
  .dnd-btn {
    border: none;
    outline: none;
    background: var(--c-contrast-bg);
    box-shadow: var(--c-contrast-bg-border);
    color: var(--c-text-secondary);
    font-size: 11px;
    font-weight: 600;
    padding: 5px 10px;
    border-radius: 14px;
    display: inline-flex;
    align-items: center;
    gap: 5px;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    line-height: 1;
  }

  .dnd-btn:hover {
    background: var(--c-contrast-bg-hover);
    color: var(--c-text-primary);
    transform: translateY(-1px);
  }

  .dnd-btn:active,
  .clear-btn:active {
    transform: scale(0.95);
  }

  .dnd-btn.dnd-active {
    color: var(--c-dnd-active-text);
    background: var(--c-dnd-active-bg);
    box-shadow: var(--c-dnd-active-border);
  }

  .dnd-btn.dnd-active:hover {
    background: var(--c-dnd-active-bg);
    transform: translateY(-1px);
  }

  .clear-btn:hover {
    background: var(--c-clear-btn-hover-bg);
    color: var(--c-clear-btn-hover-text);
    box-shadow: var(--c-clear-btn-hover-border);
    transform: translateY(-1px);
  }

  .notif-list {
    width: 100%;
    flex: 1;
    min-height: 0; /* CRITICAL: prevents flex child from overflowing parent */
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding-right: 5px; /* for scrollbar */
    outline: none;
  }

  .notif-list::-webkit-scrollbar {
    width: 4px;
  }
  .notif-list::-webkit-scrollbar-track {
    background: transparent;
  }
  .notif-list::-webkit-scrollbar-thumb {
    background: var(--c-scrollbar-thumb);
    border-radius: 4px;
  }

  .notif-item {
    padding: 14px 12px;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    position: relative;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    border-radius: 14px;
  }

  .notif-item:hover {
    background-color: var(--c-contrast-bg);
    border-color: var(--c-hover);
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }

  .notif-item:active {
    transform: scale(0.98);
  }

  .notif-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    margin-bottom: 6px;
  }

  .notif-item.nested .notif-item-header {
    justify-content: flex-end;
  }

  .time-and-dismiss {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    position: relative;
    min-height: 26px;
    padding: 2px 0;
    white-space: nowrap;
    flex-shrink: 0;
    margin-left: auto;
  }

  .nested-time-and-dismiss {
    position: absolute;
    top: 10px;
    right: 12px;
    z-index: 10;
  }

  .notif-item-time {
    font-size: 0.68rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--c-text-muted);
    opacity: 0.6;
    transition: opacity 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    flex-shrink: 0;
  }

  .notif-actions {
    display: flex;
    align-items: center;
    gap: 6px;
    opacity: 0;
    pointer-events: none;
    position: absolute;
    right: 0;
    top: 50%;
    transform: translateY(-50%);
    transition: opacity 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-item:hover .notif-actions,
  .notif-group-header:hover .notif-actions {
    opacity: 1;
    pointer-events: auto;
  }

  .notif-item:hover .notif-item-time,
  .notif-group-header:hover .notif-item-time {
    opacity: 0;
  }

  /* Connector layout for grouped notifications */
  .notif-group {
    background: transparent;
    border: none;
    position: relative;
    width: 100%;
    margin-bottom: 8px;
    box-sizing: border-box;
    transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-group:not(.is-expanded) {
    cursor: pointer;
  }

  .notif-group-header {
    padding: 14px 12px;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    border-radius: 14px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
    box-sizing: border-box;
    cursor: pointer;
    transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-group-header:hover {
    background-color: var(--c-contrast-bg);
    border-color: var(--c-hover);
  }

  .notif-group-header-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
  }

  .notif-group-preview {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    width: 100%;
    gap: 4px;
    padding-top: 4px;
  }

  .notif-group-items {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-top: 10px;
    padding-left: 12px;
    border-left: 1px dashed var(--c-hover);
  }

  .expand-indicator {
    color: var(--c-text-secondary);
    background: var(--c-contrast-bg);
    border: 1px solid var(--c-contrast-bg);
    border-radius: 50%;
    width: 26px;
    height: 26px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .expand-indicator:hover {
    background: var(--c-hover);
    color: var(--c-text-primary);
  }

  /* Keep visible when expanded so user can collapse */
  .expand-indicator.is-expanded {
    transform: rotate(180deg);
  }

  .notif-item-app-group {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .notif-item-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    flex-shrink: 0;
    font-size: 14px;
    background: var(--c-hover);
    border-radius: 6px;
    overflow: hidden;
    box-shadow:
      inset 0 1px 0 var(--c-hover),
      0 2px 4px rgba(0, 0, 0, 0.2);
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-avatar-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    flex-shrink: 0;
    font-size: 18px;
    background: var(--c-icon-badge-bg);
    border: 1px solid var(--c-icon-badge-border);
    border-radius: 50%;
    overflow: hidden;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-img-icon {
    width: 100%;
    height: 100%;
    object-fit: contain;
    border-radius: 50%;
  }

  .notif-item-layout {
    display: flex;
    gap: 12px;
    width: 100%;
  }

  .notif-item-avatar-col {
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
  }

  .notif-item-content-col {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    min-width: 0;
  }

  .notif-item:hover .notif-avatar-icon {
    transform: scale(1.02);
  }

  .notif-header-app-icon {
    width: 16px;
    height: 16px;
    border-radius: 4px;
    object-fit: contain;
  }

  .notif-item-app {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--c-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 140px;
  }

  .notif-item-time {
    font-size: 0.65rem;
    font-weight: 600;
    color: var(--c-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    opacity: 0.8;
  }

  .notif-item-body {
    display: flex;
    flex-direction: column;
    justify-content: center;
    text-align: left;
    gap: 4px;
    min-width: 0;
  }

  .notif-item-title-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    width: 100%;
  }

  .notif-item-title {
    font-weight: 600;
    font-size: 0.95rem;
    color: var(--c-text-primary);
    line-height: 1.4;
    padding-right: 8px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    width: 100%;
    flex: 1;
    min-width: 0;
  }

  .notif-item-title.is-expanded {
    white-space: normal;
  }

  .notif-item-message {
    margin: 0;
    font-size: 0.85rem;
    color: var(--c-text-muted);
    line-height: 1.4;
    width: 100%;
    display: -webkit-box;
    -webkit-line-clamp: 1;
    line-clamp: 1;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* When item is expanded, reveal the full multi-line message */
  .notif-item-message.is-expanded {
    display: block;
    -webkit-line-clamp: none;
    line-clamp: none;
    white-space: pre-wrap;
    word-break: break-word;
  }

  /* Keyboard Accessibility Focus Rings */
  .dnd-btn:focus-visible,
  .clear-btn:focus-visible,
  .dismiss-item-btn:focus-visible,
  .notif-item:focus-visible {
    outline: 2px solid var(--c-primary);
    outline-offset: 2px;
  }

  .hero-image-wrapper {
    margin-top: 8px;
    width: 100%;
    border-radius: 6px;
    overflow: hidden;
  }

  .hero-image {
    width: 100%;
    height: auto;
    max-height: 250px;
    display: block;
    object-fit: scale-down;
    background-color: var(--c-contrast-bg);
  }

  .empty-state {
    text-align: center;
    color: var(--c-text-muted);
    margin-top: 2rem;
    font-size: 0.9rem;
  }

  /* Grouping Styles */
  .notif-group.unified-card {
    display: flex;
    flex-direction: column;
    margin-bottom: 8px;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    border-radius: 14px;
    position: relative;
    overflow: hidden;
    transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-group.unified-card:hover {
    background: var(--c-contrast-bg);
    border-color: var(--c-hover);
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.2);
  }

  .notif-group-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px;
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .notif-count-badge {
    background: var(--c-primary);
    color: #ffffff;
    font-size: 0.68rem;
    font-weight: 700;
    padding: 2px 8px;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.15);
    margin-left: 6px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
  }

  .notification-container {
    height: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .permission-banner {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 16px 12px;
    padding: 12px;
    background: var(--c-danger-hover-bg);
    border: 1px solid var(--c-danger);
    border-radius: 8px;
    color: var(--c-danger);
    font-size: 13px;
    line-height: 1.4;
  }

  .notif-group-items {
    display: flex;
    flex-direction: column;
    padding: 0;
    margin: 0;
    border: none;
    gap: 0;
  }

  .notif-item.nested {
    background: transparent;
    border: none;
    border-top: 1px solid var(--c-contrast-bg);
    border-radius: 0;
    padding: 10px 12px;
    box-shadow: none;
    transition: background-color 0.2s ease;
  }

  .notif-item.nested .notif-item-content-col {
    padding-right: 75px;
  }

  /* Make the first item NOT have a top border to blend with header */
  .notif-item.nested:first-child {
    border-top: none;
  }

  .notif-item.nested:hover {
    background-color: var(--c-contrast-bg);
    transform: none;
    border-color: var(--c-contrast-bg);
  }
</style>
