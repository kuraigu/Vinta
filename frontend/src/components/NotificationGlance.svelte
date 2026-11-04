<script lang="ts">
  import { getContext } from 'svelte'
  import { Bell, X } from 'lucide-svelte'
  import { notifData, type IslandConfig } from '../store'
  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'
  import {
    nativeIconsStore,
    parseAumid,
    requestNativeAppIcon
  } from '../services/iconService'
  import { islandTooltip } from '../services/islandTooltip'
  import ActionCapsule from './ActionCapsule.svelte'
  import BentoCard from './BentoCard.svelte'
  import appIcon from '@appicon'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let notifications = $derived(
    $notifData.filter(
      (n) =>
        !n.isUpdate &&
        !n.title?.toLowerCase().includes('update available') &&
        !(
          (n.appName?.toLowerCase() === 'vinta' ||
            n.aumid?.toLowerCase().includes('vinta')) &&
          n.title?.toLowerCase().includes('update')
        )
    )
  )
  let latestNotif = $derived(notifications.length > 0 ? notifications[0] : null)
  let unreadCount = $derived(notifications.length)
  let nativeIcons = $derived($nativeIconsStore)

  function openNotifications(e?: Event) {
    if (e) e.stopPropagation()
    if (navigate) {
      navigate(NotificationPage, notifConfig, true)
    }
  }

  function dismissLatest(e: Event) {
    e.stopPropagation()
    if (!latestNotif) return
    const idToDismiss = latestNotif.id
    notifData.update((items) => items.filter((n) => n.id !== idToDismiss))
  }

  function resolveIconUrl(notif: typeof latestNotif): string | null {
    if (!notif) return null

    const rawInput = notif.appName || notif.aumid || ''
    const isVinta =
      rawInput.toLowerCase().includes('vinta') ||
      (notif.title || '').toLowerCase().includes('vinta')

    if (isVinta) {
      return notif.icon || notif.appLogo || appIcon
    }

    if (notif.appLogo && notif.appLogo.length > 0) {
      return notif.appLogo
    }

    const parsed = parseAumid(rawInput)
    const searchApp = (parsed.processName || notif.appName || '').trim()
    const searchAumid = (notif.aumid || '').trim()
    const iconKey = `${searchApp.toLowerCase()}|${searchAumid.toLowerCase()}`

    if (nativeIcons[iconKey]) {
      return nativeIcons[iconKey]
    }
    if (notif.aumid && nativeIcons[notif.aumid]) {
      return nativeIcons[notif.aumid]
    }
    if (notif.appName && nativeIcons[notif.appName]) {
      return nativeIcons[notif.appName]
    }

    // Trigger asynchronous extraction if not in cache yet
    if (searchApp || searchAumid) {
      const extracted = requestNativeAppIcon(searchApp, searchAumid)
      if (extracted) return extracted
    }

    // 3. Fall back to notif.icon if it doesn't match a wide hero image
    if (notif.icon && notif.icon.length > 0 && notif.icon !== notif.heroImage) {
      return notif.icon
    }

    // 4. Default to standard app icon if available
    return notif.icon || null
  }

  let resolvedIcon = $derived(resolveIconUrl(latestNotif))
  let appDisplayName = $derived(
    latestNotif?.appName ||
      (latestNotif?.aumid
        ? parseAumid(latestNotif.aumid).displayName
        : 'Notification')
  )
</script>

{#if latestNotif}
  <BentoCard
    variant="bento"
    class="notification-glance-card"
    onclick={openNotifications}
  >
    <div class="glance-top-row">
      <div class="glance-brand-group">
        <div class="glance-avatar-container">
          {#key latestNotif?.id}
            {#if resolvedIcon}
              <img
                src={resolvedIcon}
                alt=""
                class="glance-app-icon"
                onerror={(e) => {
                  const target = e.currentTarget as HTMLImageElement
                  target.style.display = 'none'
                  const fallback = target.nextElementSibling as HTMLElement
                  if (fallback) fallback.style.display = 'flex'
                }}
              />
              <div class="glance-fallback-icon" style="display: none;">
                <Bell size={13} />
              </div>
            {:else}
              <div class="glance-fallback-icon">
                <Bell size={13} />
              </div>
            {/if}
          {/key}
        </div>
        <span class="glance-app-name">{appDisplayName}</span>
        {#if unreadCount > 1}
          <span class="glance-count-badge"
            >{unreadCount > 9 ? '9+' : unreadCount}</span
          >
        {/if}
      </div>

      <div class="glance-actions-group">
        <ActionCapsule
          label="View"
          onclick={openNotifications}
          tooltip="Open Notification Center"
        />
        <button
          type="button"
          class="glance-action-btn dismiss"
          onclick={dismissLatest}
          use:islandTooltip={'Dismiss notification'}
          aria-label="Dismiss this notification"
        >
          <X size={12} />
        </button>
      </div>
    </div>

    <div class="glance-content-preview">
      {#if latestNotif.title}
        <h4 class="glance-title">{latestNotif.title}</h4>
      {/if}
      {#if latestNotif.message}
        <p class="glance-message">{latestNotif.message}</p>
      {/if}
    </div>
  </BentoCard>
{/if}

<style>
  :global(.notification-glance-card) {
    gap: 8px;
    padding: 10px 14px;
    cursor: pointer;
    text-align: left;
    align-items: flex-start;
  }

  :global(.notification-glance-card:active) {
    transform: scale(0.992);
  }

  .glance-top-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    gap: 8px;
  }

  .glance-brand-group {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;
  }

  .glance-avatar-container {
    width: 20px;
    height: 20px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    flex-shrink: 0;
    background: rgba(255, 255, 255, 0.08);
  }

  .glance-app-icon {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .glance-fallback-icon {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--vinta-accent-primary, #60a5fa);
  }

  .glance-app-name {
    font-size: 0.78rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.65));
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-transform: capitalize;
  }

  .glance-count-badge {
    font-size: 0.68rem;
    font-weight: 700;
    color: var(--vinta-text-on-accent, #ffffff);
    background: var(--vinta-accent-primary, #38bdf8);
    border: none;
    border-radius: 999px;
    padding: 1px 6px;
    line-height: 1.2;
    flex-shrink: 0;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.35);
  }

  .glance-actions-group {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  .glance-action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    cursor: pointer;
    border-radius: 6px;
    color: var(--vinta-text-muted, rgba(255, 255, 255, 0.45));
    transition: all 0.15s ease;
    padding: 3px 6px;
    font-size: 0.72rem;
    font-weight: 500;
    gap: 2px;
  }

  .glance-action-btn.dismiss {
    width: 20px;
    height: 20px;
    padding: 0;
    border-radius: 50%;
  }

  .glance-action-btn.dismiss:hover {
    background: rgba(255, 255, 255, 0.12);
    color: var(--vinta-text-primary, #fff);
  }

  .glance-content-preview {
    display: flex;
    flex-direction: column;
    gap: 2px;
    width: 100%;
  }

  .glance-title {
    margin: 0;
    font-size: 0.84rem;
    font-weight: 600;
    letter-spacing: -0.015em;
    color: var(--vinta-text-primary, #ffffff);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: left;
  }

  .glance-message {
    margin: 0;
    font-size: 0.76rem;
    line-height: 1.35;
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.65));
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    word-break: break-word;
    text-align: left;
  }
</style>
