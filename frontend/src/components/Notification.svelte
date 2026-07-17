<script lang="ts">
  import { flip } from 'svelte/animate'
  import { fly, fade, slide } from 'svelte/transition'
  import { cubicOut } from 'svelte/easing'
  import {
    Bell,
    BellOff,
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
  import { onMount } from 'svelte'
  import { SvelteSet } from 'svelte/reactivity'
  import { appIcons, gameIcons } from '../appIcons'

  let { expanded = false, notifications = [] } = $props<{
    expanded?: boolean
    notifications?: Array<{
      id: number
      appName?: string
      icon?: string
      title: string
      message: string
      heroImage?: string
      timestamp?: number
    }>
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
    brandCategory?: string
    items: typeof notifications
    latestTimestamp: number
  }

  let groupedNotifications = $derived.by(() => {
    const groups: Record<string, NotificationGroup> = {}
    for (const notif of notifications) {
      const brand = getDisplayBrand(notif)
      const key = brand.name
      if (!groups[key]) {
        groups[key] = {
          brandName: key,
          brandIcon: brand.icon,
          brandCategory: brand.category,
          items: [],
          latestTimestamp: 0
        }
      }
      groups[key].items.push(notif)
      const ts = notif.timestamp || now
      if (ts > groups[key].latestTimestamp) {
        groups[key].latestTimestamp = ts
      }
    }
    return Object.values(groups).sort(
      (a, b) => b.latestTimestamp - a.latestTimestamp
    )
  })

  let latest = $derived(
    notifications[0] ||
      ({
        id: 0,
        title: 'No notifications',
        message: ''
      } as (typeof notifications)[0])
  )

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

  function clearAll() {
    notifData.set([])
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

  function getDisplayBrand(notif: { appName?: string; icon?: string }): {
    name: string
    icon: string | null
    category?: string
  } {
    const rawName = notif.appName || 'Notification'
    const defaultIcon = notif.icon || null

    // Match against the WinRT display name (already human-readable, e.g. "Discord", "NVIDIA App")
    const matchString = (notif.appName || '').toLowerCase()
    if (!matchString) return { name: rawName, icon: defaultIcon }

    const allIcons = [...gameIcons, ...appIcons]
    for (const brand of allIcons) {
      if (brand.matchers.some((m) => matchString.includes(m))) {
        return { name: brand.name, icon: brand.icon, category: brand.category }
      }
    }
    return { name: rawName, icon: defaultIcon }
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
    <Bell {size} color="var(--c-text-secondary)" />
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
          </button>
          {#if notifications.length > 0}
            <button class="clear-btn" onclick={clearAll}>
              <Trash2 size={14} />
              Clear
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
        {#snippet renderNotifItem(notif: any, brand: any, nested: boolean)}
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
                toggleExpand(notif.id)
              }
            }}
          >
            {#if !nested}
              {@const iconUrl = brand.icon || brand.brandIcon}
              <div class="notif-item-header">
                <div class="notif-item-app-group">
                  <div class="notif-item-icon">
                    {#if isImageUrl(iconUrl)}
                      <img
                        src={iconUrl}
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
                        {@render fallbackIcon(brand, 16)}
                      </div>
                    {:else if iconUrl}
                      <span class="icon-text">{iconUrl}</span>
                    {:else}
                      {@render fallbackIcon(brand, 16)}
                    {/if}
                  </div>
                  <span class="notif-item-app"
                    >{brand.name || brand.brandName}</span
                  >
                </div>
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
              </div>
            {/if}
            <div class="notif-item-body">
              <div class="notif-item-title-row">
                <span class="notif-item-title">{notif.title}</span>
                {#if nested}
                  <div class="time-and-dismiss nested-actions">
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
                {/if}
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
                <div class="hero-image-wrapper" style="display: none;">
                  <img
                    src={notif.heroImage}
                    alt="hero"
                    class="hero-image"
                    onload={(e) => {
                      if (
                        (e.currentTarget as HTMLImageElement).naturalWidth >=
                        150
                      ) {
                        ;(
                          (e.currentTarget as HTMLImageElement)
                            .parentElement as HTMLElement
                        ).style.display = 'block'
                      }
                    }}
                  />
                </div>
              {/if}
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
                    // Only toggle group if clicking header
                    if (expandedGroups.has(group.brandName)) {
                      e.stopPropagation()
                      toggleGroupExpand(group.brandName)
                    }
                  }}
                >
                  <div class="notif-item-app-group">
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
                    <span class="notif-item-app">{group.brandName}</span>
                    <span class="notif-count-badge">{group.items.length}</span>
                  </div>
                  <div class="time-and-dismiss">
                    <span class="notif-item-time"
                      >{getRelativeTime(group.latestTimestamp)}</span
                    >
                    <div class="notif-actions">
                      <div
                        class="expand-indicator"
                        class:is-expanded={expandedGroups.has(group.brandName)}
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

                <div class="notif-group-items">
                  {#each group.items as notif, i (notif.id)}
                    {#if i === 0 || expandedGroups.has(group.brandName)}
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
                    {/if}
                  {/each}
                </div>
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
  }

  .hero-image {
    width: 100%;
    height: auto;
    display: block;
    object-fit: cover;
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
    gap: 10px;
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
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.05);
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
    background: rgba(255, 59, 48, 0.15);
    color: #ff453a;
    border-color: rgba(255, 59, 48, 0.2);
    transform: scale(1.1);
  }

  .clear-btn {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.05);
    color: var(--c-text-secondary);
    font-size: 12px;
    font-weight: 600;
    padding: 6px 12px;
    border-radius: 16px;
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .clear-btn:hover {
    background: rgba(255, 59, 48, 0.15);
    color: #ff453a;
    border-color: rgba(255, 59, 48, 0.2);
    transform: scale(1.02);
  }

  .clear-btn:active {
    transform: scale(0.96);
  }

  .notif-header-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .dnd-btn {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.05);
    color: var(--c-text-secondary);
    width: 30px;
    height: 30px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .dnd-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--c-text-primary);
    transform: scale(1.08);
  }

  .dnd-btn.dnd-active {
    background: rgba(255, 184, 0, 0.15);
    border-color: rgba(255, 184, 0, 0.25);
    color: #ffb800;
  }

  .dnd-btn.dnd-active:hover {
    background: rgba(255, 184, 0, 0.25);
    color: #ffca40;
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
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.04);
    position: relative;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    border-radius: 14px;
  }

  .notif-item:hover {
    background-color: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.08);
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

  .time-and-dismiss {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    width: 80px;
    height: 26px;
  }

  .notif-item-time {
    font-size: 0.72rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--c-text-muted);
    opacity: 0.6;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-actions {
    position: absolute;
    right: 0;
    display: flex;
    align-items: center;
    gap: 6px;
    opacity: 0;
    pointer-events: none;
    transform: translateX(4px);
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-item:hover .notif-item-time,
  .notif-group-header:hover .notif-item-time {
    opacity: 0;
    transform: translateX(-4px);
  }

  .notif-item:hover .notif-actions,
  .notif-group-header:hover .notif-actions {
    opacity: 1;
    pointer-events: auto;
    transform: translateX(0);
  }

  .expand-indicator {
    color: var(--c-text-secondary);
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.05);
    border-radius: 50%;
    width: 26px;
    height: 26px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .expand-indicator:hover {
    background: rgba(255, 255, 255, 0.1);
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
    background: rgba(255, 255, 255, 0.06);
    border-radius: 6px;
    overflow: hidden;
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.1),
      0 2px 4px rgba(0, 0, 0, 0.2);
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-item:hover .notif-item-icon,
  .notif-group-header:hover .notif-item-icon {
    transform: scale(1.08);
  }

  .notif-item-app {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--c-text-primary);
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
  }

  .time-and-dismiss.nested-actions {
    opacity: 0;
    transition: opacity 0.2s ease;
  }

  .notif-item:hover .time-and-dismiss.nested-actions {
    opacity: 1;
  }

  .notif-item-message {
    margin: 0;
    font-size: 0.85rem;
    color: var(--c-text-muted);
    line-height: 1.4;
    width: 100%;
    display: none;
  }

  /* Show message when explicitly made visible (nested items always show) */
  .notif-item-message.is-visible {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .notif-item-message.is-expanded {
    display: block;
    white-space: pre-wrap;
    word-break: break-word;
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
    background-color: rgba(255, 255, 255, 0.05);
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
    margin-bottom: 4px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.04);
    border-radius: 14px;
    overflow: hidden;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .notif-group.unified-card:hover {
    background: rgba(255, 255, 255, 0.04);
    border-color: rgba(255, 255, 255, 0.06);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }

  .notif-group-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 12px 8px 12px;
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .notif-count-badge {
    background: linear-gradient(
      135deg,
      var(--c-primary, #007aff) 0%,
      #005bb5 100%
    );
    color: #ffffff;
    font-size: 0.7rem;
    font-weight: 700;
    padding: 2px 8px;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    margin-left: 8px;
    box-shadow: 0 2px 4px rgba(0, 122, 255, 0.4);
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
    background: rgba(255, 69, 58, 0.1);
    border: 1px solid rgba(255, 69, 58, 0.2);
    border-radius: 8px;
    color: #ff453a;
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
    border-top: 1px solid rgba(255, 255, 255, 0.05);
    border-radius: 0;
    padding: 10px 12px;
    box-shadow: none;
    transition: background-color 0.2s ease;
  }

  /* Make the first item NOT have a top border to blend with header */
  .notif-item.nested:first-child {
    border-top: none;
  }

  .notif-item.nested:hover {
    background-color: rgba(255, 255, 255, 0.02);
    transform: none;
    border-color: rgba(255, 255, 255, 0.05);
  }
</style>
